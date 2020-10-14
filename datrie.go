package baserouter

import (
	"fmt"
	"sync"
)

type handle struct {
	handle    HandleFunc
	path      string
	paramName string
}

type datrie struct {
	base  []int
	check []int
	tail  []byte

	// 为了支持变量加的元数据
	baseHandler []*handle
	head        []int
	tailHandler []*handle
	pos         int
	path        int //存放保存path个数
	maxParam    int //最大参数个数
	paramPool   sync.Pool
}

// 初始化函数
func newDatrie() *datrie {
	d := &datrie{
		base:        make([]int, 1024),
		check:       make([]int, 1024),
		tail:        make([]byte, 1024),
		head:        make([]int, 1024),
		baseHandler: make([]*handle, 1024),
		tailHandler: make([]*handle, 1024),
	}

	d.base[0] = 1
	d.pos = 1
	return d
}

// 拷贝handle
// pos 是相对于insertPath的偏移量
func (d *datrie) copyHandler(pos int, p *path) {
	// 为了tail, head, tailHandler末端对齐， i < len(p.insertPath) - pos ，这里不是等于，原因看下图
	// tail test/word/:
	// tail byte      [0 116 101 115 116 47 119 111 114 100 47 58 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	// head           [0 11 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	// tail handle    [<nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> 0xc0000132f0]
	//p.debug()

	for i := 0; i < len(p.insertPath)-pos; i++ {
		d.tailHandler[d.pos+i] = p.paramAndHandle[pos+i]
	}
}

// 没有冲突
func (d *datrie) noConflict(pos int, prevIndex int, base int, p *path) {
	// pos位置的字符已经放到base里面，所以跳过这个字符，也是这里pos+1的由来
	path := p.insertPath[pos+1:]

	d.expansionTailAndHandler(path)

	copy(d.tail[d.pos:], path)
	d.head[d.pos] = len(path)
	d.check[base] = prevIndex
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
	fmt.Printf("tail handle %2s %v\n", "", d.tailHandler[:max])
	fmt.Printf("base handle %2s %v\n", "", d.baseHandler[:max])
	fmt.Printf("pos  %9s %d\n", "", d.pos)
}

func (d *datrie) findParamOrWildcard(start, k int, path string, p *Params) (h *handle, p2 *Params) {
	start = -start
	l := d.head[start]

	prevIndex := 0

	var i, j int
	var c byte

	// i 必然是从0开始算的, l里面存放还有多少字符
	// i 指向path 变量余下保存在tail里面第一个字符的位置
	// k + 1 如果这个字符在d.base 和 d.check有记录，余下的才会保存到tail, 所以在path的位置就是k+1
	for i, j = 0, k+1; i < l; i++ {

		h = d.tailHandler[start+i]
		c = d.tail[start+i]

		if c == ':' && h != nil && h.paramName != "" {

			p.appendKey(h.paramName)
			prevIndex = j

			for ; j < len(path) && path[j] != '/'; j++ {
			}

			p.setVal(path[prevIndex:j])

			if j == len(path) { // 这是路径里面最后一个变量
				break
			}

			continue //该path可能还有变量

		}

		if c == '*' && h != nil && h.paramName != "" {

			p.appendKey(h.paramName)
			p.setVal(path[j:len(path)])
			break
		}

		if j < len(path) {
			if path[j] != d.tail[start+i] {
				return nil, nil
			}
		}

		j++

	}

	return d.tailHandler[start+l-1], p
}

func (d *datrie) findBaseHandler(k, prevIndex2, index2 *int, path string, p *Params) (*handle, *Params) {
	prevIndex := *prevIndex2
	index := *index2

	var i int

	for i = *k + 1; i < len(path); i++ {

		c := path[i]

		if path[i-1] == '/' {

			prevIndex = d.base[prevIndex] + getCodeOffset('/')
			index = d.base[prevIndex] + getCodeOffset(':')

			h := d.baseHandler[index]
			if h != nil && h.paramName != "" && d.check[index] == prevIndex { //找到普通变量
				prevIndex = index

				p.appendKey(h.paramName)

				var j int
				for j = i + 1; j < len(path) && path[j] != '/'; j++ {
				}

				p.setVal(path[i:j])
				i = j //TODO 这里会不会有bug?有时间再思考下
				break
			}

			index = d.base[prevIndex] + getCodeOffset('*')
			h = d.baseHandler[index]
			if h != nil && h.paramName != "" && d.check[index] == prevIndex { //找到贪婪匹配
				prevIndex = index
				p.appendKey(h.paramName)
				p.setVal(path[i:len(path)])
				i = len(path)
				break
			}

		}

		if c != '/' {
			return nil, p
		}
	}

	*k = i
	*prevIndex2 = prevIndex
	*index2 = index
	return d.baseHandler[index], p
}

func (d *datrie) lookup(path string) (h *handle, p Params) {
	p = make(Params, 0, d.maxParam)
	h, p2 := d.lookup2(path, &p)
	if p2 == nil {
		return nil, p
	}

	return h, *p2
}

// 查找
func (d *datrie) lookup2(path string, p2 *Params) (h *handle, p *Params) {

	prevIndex := 1
	var index int

	for k := 0; k < len(path); k++ {

		c := path[k]
		// 如果只有一个path，baseHandler里面肯定没有数据，就不需要进入findBaseHandler函数
		if c == '/' && d.path > 1 {
			_, p = d.findBaseHandler(&k, &prevIndex, &index, path, p2)
			c = path[k]
		}

		index = d.base[prevIndex] + getCodeOffset(c)

		if start := d.base[index]; start < 0 && d.check[index] == prevIndex {
			return d.findParamOrWildcard(start, k, path, p2)
		}

		//fmt.Printf("index = %d, d.check[index] = %d, d.base[index] = %d, %c\n", index, d.check[index], d.base[index], c)
		if d.check[index] <= 0 {
			return nil, nil
		}

		prevIndex = index

	}

	return d.baseHandler[index], p
}

// case3 step 8 or 10
func (d *datrie) baseAndCheck(base int, c byte, tail int) {
	q := d.xCheck(c)
	d.base[base] = q
	nextBase := d.base[base] + getCodeOffset(c)

	//fmt.Printf("baseAndCheck->base %d, %c, check:%d\n", base, c, d.check[base])
	d.base[nextBase] = -tail
	d.check[nextBase] = base //指向它的爸爸索引
}

// step 9
func (d *datrie) moveTailAndHandler(temp int, tailPath []byte) {
	copy(d.tail[temp:], tailPath) //移动字符串
	// 总长度(temp+d.head[temp])-实际长度(len(tailPath)) = 新的需要插入的位置
	copy(d.tailHandler[temp:], d.tailHandler[temp+d.head[temp]-len(tailPath):temp+d.head[temp]])

	for i := len(tailPath); i < d.head[temp]; i++ {
		d.tail[temp+i] = '?'
		d.tailHandler[temp+i] = nil
	}

	d.head[temp] = len(tailPath)
}

// 共同前缀冲突
func (d *datrie) samePrefix(pos, start int, base int, p *path) {
	path := p.insertPath
	start = -start
	l := d.head[start]
	temp := start //step 4

	if path[pos:] == BytesToString(d.tail[start:start+l]) {
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
		c := insertPath[i]
		q := d.xCheck(c) //找出可以跳转的位置 , case3 step 5.
		d.base[base] = q //修改老的跳转位置, case3 step 6.

		nextBase := d.base[base] + getCodeOffset(c)
		d.check[nextBase] = base
		//fmt.Printf("c = %c, d.base[base] = %d, q = %d, nextBase:%d, d.check = %d\n", c, d.base[base], q, nextBase, d.check[nextBase])
		base = nextBase

		if d.tailHandler[start+i] != nil {
			d.baseHandler[base] = d.tailHandler[start+i]
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

	/*
		if len(tailPath) == 0 {
			//i--
		}
	*/

	// 开始处理insertPath 中没有共同前缀的字符串
	d.expansionTailAndHandler(insertPath[i+1:])
	// step 10
	d.baseAndCheck(base, insertPath[i], d.pos)

	copy(d.tail[d.pos:], insertPath[i+1:])
	d.head[d.pos] = len(insertPath[i+1:])

	// case3 step 11
	// i是相对于insertPath 加上pos就是相当于对于path, 最后+ 1就是跳过当前字符
	d.copyHandler(i+1+pos, p)
	d.pos += len(insertPath[i+1:])
}

func (d *datrie) findAllNode(prevIndex int) (rv []byte) {
	for index, checkPrevIndex := range d.check {
		if checkPrevIndex == prevIndex {
			// d.base[prevIndex] + offset = index，所以求offset 就是如下
			offset := index - d.base[prevIndex]
			rv = append(rv, getCharFromOffset(offset))
		}
	}
	return
}

func (d *datrie) selectList(prevIndex, index int) (list []byte, lessIndex, moreIndex int) {
	// step 3
	list1 := d.findAllNode(prevIndex)
	list2 := d.findAllNode(d.check[index])

	list = list1
	lessIndex = prevIndex
	moreIndex = d.check[index]
	// 取子节点比较少的那个节点
	if len(list1)+1 > len(list2) {
		// 已经有的是list1 这里还要加新节点，所以len(list)+1
		list = list2
		lessIndex = d.check[index]
		moreIndex = prevIndex
	}

	return
}

func (d *datrie) insertConflict(pos int, prevIndex, index int, p *path) {
	path := p.insertPath
	var list []byte
	tempNode1 := index
	// step 2
	if d.check[tempNode1] == 0 {
		// 如果d.check[tempNode1] 是0，说明这个节点还没有使用过，直接插入
		// 然后直接返回
		// TODO
		// fmt.Printf(":::::::%d\n", d.check[tempNode1])
		//return
	}

	list, lessIndex, moreIndex := d.selectList(prevIndex, index)

	// step 5
	tempBase := d.base[lessIndex]
	q := d.xCheckArray(list)
	d.base[lessIndex] = q

	for _, currChar := range list {
		// step 6 or step 9
		tempNode1 = tempBase + getCodeOffset(currChar)
		tempNode2 := d.base[lessIndex] + getCodeOffset(currChar)
		d.base[tempNode2] = d.base[tempNode1]
		d.check[tempNode2] = d.check[tempNode1]

		d.baseHandler[tempNode2] = d.baseHandler[tempNode1]

		/*
			fmt.Printf("currChar(%c) tempNode1 = %d, tempNode2 = %d, base %d <- %d, check %d <- %d\n",
				currChar, tempNode1, tempNode2, d.base[tempNode2], d.base[tempNode1], d.check[tempNode2], d.check[tempNode1])
		*/
		// step 7
		if d.base[tempNode1] > 0 {
			//fmt.Printf("tempNode1 = %d, path(%s), currChar(%c) check:%d\n", tempNode1, path, currChar, d.check[tempNode1])
			offset := d.findOffset(tempNode1)
			d.check[d.base[tempNode1]+offset] = tempNode2
		}

		// step 8 or step 10
		d.base[tempNode1] = 0
		d.check[tempNode1] = 0
		d.baseHandler[tempNode1] = nil
	}

	// step 11
	tempNode := d.base[moreIndex] + getCodeOffset(list[0])

	// step 12
	d.base[tempNode] = -d.pos

	d.check[tempNode] = moreIndex // step 12.2

	// step 13
	copy(d.tail[d.pos:], path[pos+1:])
	d.head[d.pos] = len(path[pos+1:])
	d.copyHandler(pos+1, p)

	// step 14
	d.pos += len(path[pos+1:])
}

func (d *datrie) changePool(p *path) {
	if d.paramPool.New == nil {
		d.paramPool.New = func() interface{} {
			p := make(Params, 0, 0)
			return &p
		}
	}

	if p.maxParam > d.maxParam {
		d.maxParam = p.maxParam
		d.paramPool.New = func() interface{} {
			p := make(Params, 0, d.maxParam)
			return &p
		}
	}
}

// 插入
func (d *datrie) insert(path string, h HandleFunc) {
	d.path++
	prevIndex := 1

	p := genPath(path, h)
	d.changePool(p)

	for pos := 0; pos < len(p.insertPath); pos++ {
		c := p.insertPath[pos]
		index := d.base[prevIndex] + getCodeOffset(c)
		if index >= len(d.base) {
			// 扩容
			d.expansion(index)
		}

		if d.check[index] == 0 {
			d.noConflict(pos, prevIndex, index, p)
			return
		}

		// 插入的时候冲突，需要修改 父节点或子节点的接续关系
		if d.check[index] != prevIndex {
			d.insertConflict(pos, prevIndex, index, p)
			return
		}

		if start := d.base[index]; start < 0 {
			// start 小于0，说明有共同前缀
			d.samePrefix(pos, start, index, p)
			return
		}

		prevIndex = index

	}
}

// 扩容tail 和 tailHandler
func (d *datrie) expansionTailAndHandler(path string) {
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
		copy(newHandler, d.tailHandler)
		d.tailHandler = newHandler
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

func (d *datrie) findOffset(tempNode1 int) (offset int) {
	// check[base[tempNode1] + offset] == tempNode1
	// check[i] == tempNode1
	// offset = i - base[tempNode1]
	i := 0
	for i = 0; i < len(d.check); i++ {
		c := d.check[i]
		if c == tempNode1 {
			break
		}
	}

	if i == len(d.check) {
		panic("not found offset")
	}

	return i - d.base[tempNode1]
}

// 找空位置
func (d *datrie) xCheckArray(arr []byte) (q int) {
	q = 2
	for i := 0; i < len(arr); {
		c := arr[i]
		if d.check[q+getCodeOffset(c)] != 0 {
			q++
			i = 0
			continue
		}
		i++
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
