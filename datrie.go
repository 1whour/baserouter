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

func (d *datrie) expansionBase(index int) {
	if index >= len(d.base) {
		newBase := make([]int, 2*index)
		copy(newBase, d.base)
		d.base = newBase
	}
}

func (d *datrie) setBase(index int, baseValue int) {
	d.expansionBase(index)
	d.base[index] = baseValue
}

func (d *datrie) expansionCheck(index int) {
	if index >= len(d.check) {
		newCheck := make([]int, 2*index)
		copy(newCheck, d.check)
		d.check = newCheck
	}
}

func (d *datrie) setCheck(index int, parentIndex int) {
	d.expansionCheck(index)
	d.check[index] = parentIndex
}

func (d *datrie) expansionTail(pos int, needLen int) {
	if cap(d.tail[pos:]) < needLen {
		newTail := make([]byte, (len(d.tail)-pos+needLen)*2)
		copy(newTail, d.tail)
		d.tail = newTail
	}
}

func (d *datrie) copyTail(pos int, tail string) {
	d.expansionTail(pos, len(tail))
	copy(d.tail[pos:], tail)
}

func (d *datrie) copyTailBytes(pos int, tail []byte) {
	d.expansionTail(pos, len(tail))
	copy(d.tail[pos:], tail)
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
func (d *datrie) noConflict(insertPos int, parentIndex int, index int, p *path) {
	// pos位置的字符已经放到base里面，所以跳过这个字符，也是这里pos+1的由来
	path := p.insertPath[insertPos:]

	d.copyTail(d.pos, path)
	d.head[d.pos] = len(path)
	d.setCheck(index, parentIndex)
	d.setBase(index, -d.pos)

	d.copyHandler(insertPos, p)
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

func (d *datrie) findParamOrWildcard(tailPos, k int, path string, p *Params) (h *handle, p2 *Params) {
	l := d.head[tailPos]

	parentIndex := 0

	var i, j int
	var c byte

	// i 必然是从0开始算的, l里面存放还有多少字符
	// i 指向path 变量余下保存在tail里面第一个字符的位置
	// k + 1 如果这个字符在d.base 和 d.check有记录，余下的才会保存到tail, 所以在path的位置就是k+1
	for i, j = 0, k+1; i < l; i++ {

		h = d.tailHandler[tailPos+i]
		c = d.tail[tailPos+i]

		if c == ':' && h != nil && h.paramName != "" {

			p.appendKey(h.paramName)
			parentIndex = j

			for ; j < len(path) && path[j] != '/'; j++ {
			}

			p.setVal(path[parentIndex:j])

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
			if path[j] != d.tail[tailPos+i] {
				return nil, nil
			}
		}

		j++

	}

	return d.tailHandler[tailPos+l-1], p
}

func (d *datrie) findBaseHandler(k, parentIndex2, index2 *int, path string, p *Params) (*handle, *Params) {
	parentIndex := *parentIndex2
	index := *index2

	var i int

	for i = *k + 1; i < len(path); i++ {

		c := path[i]

		if path[i-1] == '/' {

			parentIndex = d.base[parentIndex] + getCodeOffset('/')
			index = d.base[parentIndex] + getCodeOffset(':')

			h := d.baseHandler[index]
			if h != nil && h.paramName != "" && d.check[index] == parentIndex { //找到普通变量
				parentIndex = index

				p.appendKey(h.paramName)

				var j int
				for j = i + 1; j < len(path) && path[j] != '/'; j++ {
				}

				p.setVal(path[i:j])
				i = j //TODO 这里会不会有bug?有时间再思考下
				break
			}

			index = d.base[parentIndex] + getCodeOffset('*')
			h = d.baseHandler[index]
			if h != nil && h.paramName != "" && d.check[index] == parentIndex { //找到贪婪匹配
				parentIndex = index
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
	*parentIndex2 = parentIndex
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

	parentIndex := 1
	var index int

	for k := 0; k < len(path); k++ {

		c := path[k]
		// 如果只有一个path，baseHandler里面肯定没有数据，就不需要进入findBaseHandler函数
		if c == '/' && d.path > 1 {
			_, p = d.findBaseHandler(&k, &parentIndex, &index, path, p2)
			c = path[k]
		}

		index = d.base[parentIndex] + getCodeOffset(c)

		if index >= len(d.base) {
			return nil, p
		}

		if tailPos := d.base[index]; tailPos < 0 && d.check[index] == parentIndex {
			if d.head[-tailPos] == 0 && k+1 == len(path) {
				break
			}

			return d.findParamOrWildcard(-tailPos, k, path, p2)
		}

		if d.check[index] <= 0 {
			return nil, nil
		}

		parentIndex = index

	}

	return d.baseHandler[index], p
}

// case3 step 8 or 10
func (d *datrie) baseAndCheck(parentIndex int, c byte, tailPos int) {
	q := d.xCheck(c)
	d.setBase(parentIndex, q)
	if d.check[parentIndex] != 0 {
		panic(fmt.Sprintf("baseAndCheck: d.check[parentIndex] %d:parentIndex(%d)", d.check[parentIndex], parentIndex))
	}

	index := d.base[parentIndex] + getCodeOffset(c)

	d.setBase(index, -tailPos)
	d.setCheck(index, parentIndex) //指向它的爸爸索引
}

// step 9
func (d *datrie) moveTailAndHandler(tailPos int, tailPath []byte) {
	//copy(d.tail[tailPos:], tailPath) //移动字符串
	d.copyTailBytes(tailPos, tailPath)
	// 总长度(tailPos+d.head[tailPos])-实际长度(len(tailPath)) = 新的需要插入的位置
	copy(d.tailHandler[tailPos:], d.tailHandler[tailPos+d.head[tailPos]-len(tailPath):tailPos+d.head[tailPos]])

	for i := len(tailPath); i < d.head[tailPos]; i++ {
		d.tail[tailPos+i] = '?'
		d.tailHandler[tailPos+i] = nil
	}

	d.head[tailPos] = len(tailPath)
}

func (d *datrie) setTail(c byte, q int, tailPos int, parentIndex int, tailPath []byte, tailHandler []*handle) {
	// 修改老的跳转基地址, case3 step 6.
	d.setBase(parentIndex, q)
	// 计算c要保存的位置
	index := d.base[parentIndex] + getCodeOffset(c)
	// 记录index的爸爸位置(爸爸都是放到check数组里面的)
	d.setCheck(index, parentIndex)

	// 保存了handle 可能是param 或者就是这个路径的handle
	// TODO check param key是否一样，不一样直接报错
	if d.tailHandler[tailPos] != nil {
		d.baseHandler[index] = d.tailHandler[tailPos]
	}

	// 移动tail的字符往前面移动,无效字符使用?代替
	if len(tailPath) > 0 {
		copy(tailPath, tailPath[1:])
		copy(tailHandler, tailHandler[1:])
		tailPath[len(tailPath)-1] = '?'
		tailHandler[len(tailHandler)-1] = nil
	}

	d.setBase(index, -tailPos)
	d.head[tailPos] = len(tailPath) - 1
}

// 共同前缀冲突
// 有4中情况，
// 1.重复插入, tail里面和insertPath里面是一样的
// 2.tail里面是短的，insertPath里面是长的，tail被包含至insertPath
// 3.tail里面是长的，insertPath里面是短的，tail包含insertPath
// 4.tail和insertPath，有共同前缀，有一个节点分叉出来，引出不同的边长
func (d *datrie) samePrefix(insertPos, tailPos int, parentIndex int, p *path) (next bool) {
	path := p.insertPath
	l := d.head[tailPos]

	if path[insertPos:] == BytesToString(d.tail[tailPos:tailPos+l]) {
		// 重复数据插入, 前缀一样
		// TODO, 选择策略 替换，还是panic,
		// TODO 测试变量不一样的情况 /:name/hello /:name/word 这种直接panic
		return
	}

	insertPos++ //路过一个字符

	insertPath := path[insertPos:]
	tailPath := d.tail[tailPos : tailPos+l]
	tailHandler := d.tailHandler[tailPos : tailPos+l]

	// 处理相同前缀
	if len(insertPath) > 0 && len(tailPath) > 0 && insertPath[0] == tailPath[0] {
		c := tailPath[0]
		// 原有的字符在tail数组里面，现在要拖到d.base
		// 先计算一个没有冲突的位置 case3 step 5.
		q := d.xCheck(c)

		d.setTail(c, q, tailPos, parentIndex, tailPath, tailHandler)
		return true
	}

	// 处理下没有的共同前缀,
	//　主要是情况2, 4, 情况3走不到这里
	list := append([]byte{}, insertPath[0])
	if len(tailPath) > 0 {
		list = append(list, tailPath[0])
	}

	q := d.xCheckArray(list)

	oldTailPos := d.base[parentIndex]
	d.base[parentIndex] = q
	if len(list) > 1 { //tailPath
		d.setTail(list[1], q, -oldTailPos, parentIndex, tailPath, tailHandler)
	}

	index := d.base[parentIndex] + getCodeOffset(list[0])
	d.noConflict(insertPos+1, parentIndex, index, p)
	return false
}

func (d *datrie) findAllChildNode(parentIndex int) (rv []byte) {
	for index, checkParentIndex := range d.check {
		if checkParentIndex == parentIndex {
			// d.base[parentIndex] + offset = index，所以求offset 就是如下
			offset := index - d.base[parentIndex]
			rv = append(rv, getCharFromOffset(offset))
		}
	}
	return
}

func (d *datrie) selectList(parentIndex, index int) (list []byte, lessIndex int) {
	// step 3
	list1 := d.findAllChildNode(parentIndex)
	list2 := d.findAllChildNode(d.check[index])

	list = list1
	lessIndex = parentIndex
	// 取子节点比较少的那个节点
	if len(list1)+1 > len(list2) {
		// 已经有的是list1 这里还要加新节点，所以len(list)+1
		list = list2
		lessIndex = d.check[index]
	}

	return
}

func (d *datrie) resetNode(index int) {
	d.base[index] = 0
	d.check[index] = 0
	d.baseHandler[index] = nil
}

func (d *datrie) insertConflict(insertPos int, parentIndex, index int, p *path) {
	var list []byte

	// 两个爸爸结点parentIndex和d.check[index] 都在抢儿子结点index
	list, lessIndex := d.selectList(parentIndex, index)

	// step 5
	tempBase := d.base[lessIndex]
	d.base[lessIndex] = d.xCheckArray(list)

	for _, c := range list {
		// step 6 or step 9
		oldNode := tempBase + getCodeOffset(c)
		newNode := d.base[lessIndex] + getCodeOffset(c)

		d.setBase(newNode, d.base[oldNode])
		d.setCheck(newNode, d.check[oldNode])

		d.baseHandler[newNode] = d.baseHandler[oldNode]

		// step 7
		if d.base[oldNode] > 0 {
			d.moveToNewParent(oldNode, newNode)
		}

		if parentIndex != lessIndex && oldNode == parentIndex {
			parentIndex = oldNode
		}
		// step 8 or step 10
		d.resetNode(oldNode)
	}

	index = d.base[parentIndex] + getCodeOffset(p.insertPath[insertPos])
	d.noConflict(insertPos+1, parentIndex, index, p)
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
	parentIndex := 1

	p := genPath(path, h)
	d.changePool(p)

	for pos := 0; pos < len(p.insertPath); pos++ {
		c := p.insertPath[pos]
		index := d.base[parentIndex] + getCodeOffset(c)
		if index >= len(d.base) {
			// 扩容
			d.expansion(index)
		}

		if d.check[index] == 0 {
			d.noConflict(pos+1, parentIndex, index, p)
			return
		}

		// 插入的时候冲突，需要修改 父节点或子节点的接续关系
		if d.check[index] != parentIndex {
			d.insertConflict(pos, parentIndex, index, p)
			return
		}

		if tailPos := d.base[index]; tailPos < 0 {
			// tailPos 小于0，说明有共同前缀
			next := d.samePrefix(pos, -tailPos, index, p)
			if !next {
				return
			}
		}

		parentIndex = index

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

func (d *datrie) moveToNewParent(oldParent, newParent int) {
	// check[base[oldParent] + offset] == oldParent
	// check[i] == tempNode1
	// offset = i - base[oldParent]

	found := false
	for i := 0; i < len(d.check); i++ {
		c := d.check[i]
		if c == oldParent {
			found = true
			offset := i - d.base[oldParent]
			d.setCheck(d.base[oldParent]+offset, newParent)
			break
		}
	}

	if !found {
		//panic(fmt.Sprintf("not found oldParent:%d", oldParent))
	}

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
func (d *datrie) xCheck(c byte) (q int) {
	q = 2
	for d.check[q+getCodeOffset(c)] != 0 {
		q++
	}

	return q

}
