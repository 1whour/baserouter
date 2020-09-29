package baserouter

import (
	"bytes"
	"fmt"
)

type handle struct {
	handle    handleFunc
	path      string
	paramName string
}

type datrie struct {
	base  []int
	check []int
	tail  []byte

	// 为了支持变量加的结构
	baseHandler []*handle
	head        []int
	handler     []*handle
	pos         int
}

// 初始化函数
func newDatrie() *datrie {
	d := &datrie{
		base:        make([]int, 1024),
		check:       make([]int, 1024),
		tail:        make([]byte, 1024),
		head:        make([]int, 1024),
		baseHandler: make([]*handle, 1024),
		handler:     make([]*handle, 1024),
	}

	d.base[0] = 1
	d.pos = 1
	return d
}

// 拷贝handle
// pos 是相对于insertPath的偏移量
func (d *datrie) copyHandler(pos int, p *path) {
	// 为了tail, head, handler末端对齐， i < len(p.insertPath) - pos ，这里不是等于，原因看下图
	// tail test/word/:
	// tail byte [0 116 101 115 116 47 119 111 114 100 47 58 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	// head      [0 11 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	// handle    [<nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> 0xc0000132f0]
	//p.debug(8)
	for i := 0; i < len(p.insertPath)-pos; i++ {
		d.handler[d.pos+i] = p.paramAndHandle[pos+i]
	}
}

// 没有冲突
func (d *datrie) noConflict(pos int, prevBase int, base int, p *path) {
	// pos位置的字符已经放到base里面，所以跳过这个字符，也是这里pos+1的由来
	path := p.insertPath[pos+1:]

	d.expansionTailAndHandler(path)

	copy(d.tail[d.pos:], path)
	d.head[d.pos] = len(path)
	d.check[base] = prevBase
	d.base[base] = -d.pos

	d.copyHandler(pos+1, p)
	d.pos += len(path)
}

func (d *datrie) debug(max int, insertWord string, index, offset, base int) {
	fmt.Printf("\n#word(%s) index(%d) offset(%d) base(%d)\n", insertWord, index, offset, base)
	fmt.Printf("base %9s %v\n", "", d.base[:max])
	fmt.Printf("check %8s %v\n", "", d.check[:max])
	fmt.Printf("tail %9s %s\n", "", d.tail[:max])
	fmt.Printf("tail byte %4s %v\n", "", d.tail[:max])
	fmt.Printf("head %9s %v\n", "", d.head[:max])
	fmt.Printf("handle %7s %v\n", "", d.handler[:max])
	fmt.Printf("base handle %2s %v\n", "", d.baseHandler[:max])
}

func (d *datrie) findParamOrWildcard(start, k int, path []byte, p *Params) (h *handle, p2 Params) {
	start = -start
	l := d.head[start]

	foundParam := false
	wildcard := false
	prevIndex := 0

	var i int
	var c byte
	for i = k; i < len(path); i++ {

		if !foundParam || !wildcard {
			h = d.handler[start+i]
			c = d.tail[start+i]
		}

		if !wildcard && c == '*' && h != nil && h.paramName != "" {
			p.appendKey(h.paramName)
			prevIndex = i
			wildcard = true
		}

		if !foundParam && c == ':' && h != nil && h.paramName != "" {
			p.appendKey(h.paramName)
			prevIndex = i
			foundParam = true
		}

		if wildcard {
			continue
		}

		if foundParam {
			if k+1+i < len(path) && path[k+1+i] == '/' {
				//TODO 类型转化优化
				p.setVal(string(path[k+1+prevIndex : k+1+i]))
				prevIndex = 0
				foundParam = false
			}

			continue
		}

		if k+1+i < l {
			if path[k+1+i] != d.tail[start+i] {
				return nil, nil
			}
		}

	}

	if foundParam || wildcard {
		// i是相对于path的偏移量，所以不需要+k
		p.setVal(string(path[k+1+prevIndex : i])) //TODO 类型转换优化
	}

	return d.handler[start+l-1], *p
}

func (d *datrie) findBaseHandler(index, prevBase2, base2 *int, path []byte, p *Params) (*handle, Params) {
	foundParam := false
	wildcard := false
	maybe := false
	prevIndex := 0
	prevBase := *prevBase2
	base := *base2

	i := *index
	for ; i < len(path); i++ {
		if !foundParam && !wildcard {
			if path[i] == '/' {
				maybe = true
				continue
			}
		}

		c := path[i]

		if foundParam {

			if c == '/' {
				p.setVal(string(path[prevIndex:i]))
				prevIndex = 0
				foundParam = false
				break
			}
			continue
		}

		if wildcard {
			continue
		}

		if maybe {

			prevBase = d.base[prevBase] + getCodeOffset('/')

			base = d.base[prevBase] + getCodeOffset(':')
			h := d.baseHandler[base]
			if h != nil && h.paramName != "" && d.check[base] == prevBase { //找到普通变量
				prevBase = base
				p.appendKey(h.paramName)
				prevIndex = i
				foundParam = true
				maybe = false
				continue
			}

			base = d.base[prevBase] + getCodeOffset('*')
			h = d.baseHandler[base]
			if h != nil && h.paramName != "" && d.check[base] == prevBase { //找到贪婪匹配
				prevBase = base
				p.appendKey(h.paramName)
				prevIndex = i
				wildcard = true
				maybe = false
				continue
			}

		}

		if c != '/' {
			return nil, *p
		}
	}

	if foundParam || wildcard {
		// i是相对于path的偏移量，所以不需要+k
		p.setVal(string(path[prevIndex:i])) //TODO 类型转换优化
	}

	*index = i
	*prevBase2 = prevBase
	*base2 = base
	return d.baseHandler[base], *p
}

// 查找
func (d *datrie) lookup(path []byte) (h *handle, p Params) {

	prevBase := 1
	var base int

	for k := 0; k < len(path); k++ {

		c := path[k]
		if c == '/' {
			_, p = d.findBaseHandler(&k, &prevBase, &base, path, &p)
		}

		c = path[k]

		base = d.base[prevBase] + getCodeOffset(c)

		if start := d.base[base]; start < 0 && d.base[prevBase] == prevBase {
			fmt.Printf("prevBase:%d lookup %p, index:%d ############ hahahaha:(%s)(%s) base %d, check(%d), %c, %c\n",
				prevBase, &path, k, path, path[k:], base, d.check[base], c, path[k])
			return d.findParamOrWildcard(start, k, path, &p)
		}

		fmt.Printf("prevBase:%d lookup %p, index:%d ############ hahahaha:(%s)(%s) base %d, check(%d), %c, %c\n",
			prevBase, &path, k, path, path[k:], base, d.check[base], c, path[k])

		if d.check[base] <= 0 {
			return nil, nil
		}

		prevBase = base

	}

	return nil, nil
}

// case3 step 8 or 10
func (d *datrie) baseAndCheck(base int, c byte, tail int) {
	q := d.xCheck(c)
	d.base[base] = q
	newBase := d.base[base] + getCodeOffset(c)
	fmt.Printf("%d, d.base[base] = %d,  d.check[base] = %d\n", base, d.base[base], d.check[base])
	fmt.Printf(":::::::::::d.base[newBase] =%d: c = (%c): d.check[newBase] = %d, tail:%d\n", d.base[newBase], c, d.check[newBase], tail)
	d.base[newBase] = -tail
	d.check[newBase] = base //指向它的爸爸索引
}

// step 9
func (d *datrie) moveTailAndHandler(temp int, tailPath []byte) {
	copy(d.tail[temp:], tailPath) //移动字符串
	// 总长度(temp+d.head[temp])-实际长度(len(tailPath)) = 新的需要插入的位置
	copy(d.handler[temp:], d.handler[temp+d.head[temp]-len(tailPath):temp+d.head[temp]])

	for i := len(tailPath); i < d.head[temp]; i++ {
		d.tail[temp+i] = '?'
		d.handler[temp+i] = nil
	}

	d.head[temp] = len(tailPath)
}

// 共同前缀冲突
// TODO test 1短 2长
//           1长 2短
func (d *datrie) samePrefix(path []byte, pos, start int, base int, h handleFunc, p *path) {
	start = -start
	l := d.head[start]
	temp := start //step 4

	if bytes.Equal(path[pos:], d.tail[start:start+l]) {
		// 重复数据插入, 前缀一样
		// TODO, 选择策略 替换，还是panic
		return
	}

	pos++

	insertPath := path[pos:]
	tailPath := d.tail[start : start+l]

	i := 0
	// 处理相同前缀, step 5
	for ; i < len(insertPath) && i < len(tailPath) && insertPath[i] == tailPath[i]; i++ {
		q := d.xCheck(insertPath[i]) //找出可以跳转的位置 , case3 step 5.
		d.base[base] = q             //修改老的跳转位置, case3 step 6.

		newBase := d.base[base] + getCodeOffset(insertPath[i])
		d.check[newBase] = base
		base = newBase

		if d.tail[start+i] == ':' || d.tail[start+i] == '*' {
			if d.handler[start+i] != nil {
				d.baseHandler[base] = d.handler[start+i]
			}
		}

	}

	if i < len(insertPath) && i < len(tailPath) {
		// 处理不同前缀, step 7
		// 找一个没有冲突的parent node的位置
		q := d.xCheckTwo(insertPath[i], tailPath[i])
		d.base[base] = q
	}

	// 开始处理tail 中没有共同前缀的字符串
	if i < len(tailPath) {
		// step 8
		d.baseAndCheck(base, tailPath[i], temp)

		tailPath = tailPath[i+1:]
	} else {
		tailPath = tailPath[i:]
		// 设置为0，在moveTailAndHandler函数里面可以给无效的tail变为?号
	}

	// step 9
	d.moveTailAndHandler(temp, tailPath)

	// 开始处理insertPath 中没有共同前缀的字符串
	if len(tailPath) == 0 {
		//i--
	}

	d.expansionTailAndHandler(insertPath[i+1:])
	// step 10
	d.baseAndCheck(base, insertPath[i], d.pos)

	copy(d.tail[d.pos:], insertPath[i+1:])
	d.head[d.pos] = len(insertPath[i+1:])

	// case3 step 11
	d.copyHandler(len(insertPath)-len(insertPath[i+1+pos:]), p)
	d.pos += len(insertPath[i+1:])
}

func (d *datrie) findAllNode(prevBase int) (rv []byte) {
	for index, base := range d.check {
		if base == prevBase {
			rv = append(rv, getCharFromOffset(index-prevBase))
		}
	}
	return
}

func (d *datrie) insertConflict(path []byte, pos int, prevBase, base int, h handleFunc) {
	var list []byte
	tempNode1 := d.base[prevBase] + getCodeOffset(path[pos])
	// step 3
	list1 := d.findAllNode(prevBase)
	list2 := d.findAllNode(base)

	list = list1
	currBase := prevBase
	// 取子节点比较少的那个节点
	if len(list1)+1 > len(list2) {
		list = list2
		currBase = base
	}

	// step 5
	tempBase := d.base[currBase]
	q := d.xCheckArray(list)
	d.base[currBase] = q

	for _, currChar := range list {
		// step 6 or step 9
		tempNode1 = tempBase + getCodeOffset(currChar)
		tempNode2 := d.base[currBase] + getCodeOffset(currChar)
		d.base[tempNode2] = d.base[tempNode1]
		d.check[tempNode2] = d.check[tempNode1]

		// step 7
		if d.base[tempNode1] > 0 {
			w := d.findOffset(tempNode1)
			d.check[d.base[tempNode1]+w] = tempNode2

		}
		// step 8 or step 10
		d.base[tempNode1] = 0
		d.check[tempNode1] = 0
	}

	// step 11
	tempNode := d.base[prevBase] + getCodeOffset(list[0])

	// step 12
	d.base[prevBase] = -d.pos
	d.base[tempNode] = -d.pos

	d.check[tempNode] = prevBase

	// step 13
	copy(d.tail[d.pos:], path[pos:])

	// step 14
	d.pos += len(path[pos:])
	d.handler[d.pos] = &handle{handle: h, path: string(path) /*TODO*/}
}

// 插入
func (d *datrie) insert(path []byte, h handleFunc) {
	prevBase := 1

	p := genPath(path, h)

	for pos, c := range p.insertPath {
		base := d.base[prevBase] + getCodeOffset(c)
		if base >= len(d.base) {
			// 扩容
			d.expansion(base)
		}

		if d.check[base] == 0 {
			d.noConflict(pos, prevBase, base, p)
			return
		}

		if d.check[base] != prevBase {
			d.insertConflict(path, pos, prevBase, d.check[base], h)
			return
		}

		if start := d.base[base]; start < 0 {
			// start 小于0，说明有共同前缀
			d.samePrefix(p.insertPath, pos, start, base, h, p)
			return
		}

		prevBase = base

	}
}

// 扩容tail 和 handler
func (d *datrie) expansionTailAndHandler(path []byte) {
	need := 0
	if len(d.tail[d.pos:]) < len(path) {
		need = len(d.tail) + len(path)
		if need < len(d.tail)*2 {
			need = len(d.tail) * 2
		}
		newTail := make([]byte, need)
		copy(newTail, d.tail)
		d.tail = newTail
	}

	if need != 0 {
		newHandler := make([]*handle, need)
		copy(newHandler, d.handler)
		d.handler = newHandler
	}
}

func expansion(array *[]int, need int) {
	a := make([]int, need)
	copy(a, *array)
	*array = a
}

// 扩容
func (d *datrie) expansion(max int) {
	need := max
	if need < len(d.base)*2 {
		need = len(d.base) * 2
	}

	expansion(&d.base, need)
	expansion(&d.check, need)

	head := make([]int, need)
	copy(head, d.head)
	d.head = head
}

func (d *datrie) findOffset(tempNode1 int) (w int) {
	found := false
	for i := 0; i < len(d.check); i++ {
		c := d.check[i]
		if c == tempNode1 {
			found = true
			break
		}
	}

	if !found {
		panic("not found offset")
	}

	return tempNode1 - d.base[tempNode1]
}

// 找空位置
func (d *datrie) xCheckArray(arr []byte) (q int) {
	q = 2
	for i := 0; i < len(arr); i++ {
		c := arr[i]
		if d.check[q+getCodeOffset(c)] != 0 {
			q++
			i = 0
		}
	}

	return q
}

// 找空位置
func (d *datrie) xCheckTwo(c1, c2 byte) (q int) {
	q = 2
	for d.check[q+getCodeOffset(c1)] != 0 || d.check[q+getCodeOffset(c2)] != 0 {
		q++
	}

	return q
}

// 找空位置
func (d *datrie) xCheck(c byte) (q int) {
	q = 2
	for d.check[q+getCodeOffset(c)] != 0 {
		q++
	}

	return q

}
