package baserouter

import (
	"fmt"
	"strings"
	"sync"
)

type handle struct {
	handle    HandleFunc
	path      string
	paramName string
	wildcard  bool
}

type base struct {
	q           int    //基地址
	tailPath    string //保存的末部字符串
	tailHandler []*handle
	handle      *handle
}

func (b *base) String() string {
	if b == nil {
		return "<nil>"
	}

	var o strings.Builder
	fmt.Fprintf(&o, "address = %p ", b)
	fmt.Fprintf(&o, "q = %d ", b.q)
	fmt.Fprintf(&o, "tailPath = %s ", b.tailPath)
	fmt.Fprintf(&o, "tailHandler = %v ", b.tailHandler)
	fmt.Fprintf(&o, "handle = %v ", b.handle)
	return o.String()
}

type datrie struct {
	base  []*base
	check []int //保存爸爸的索引

	path      int //存放保存path个数
	maxParam  int //最大参数个数
	paramPool sync.Pool
}

// 初始化函数
func newDatrie() *datrie {
	d := &datrie{
		base:  make([]*base, 2),
		check: make([]int, 2),
	}

	d.base[0] = &base{q: 1}
	d.base[1] = &base{}
	return d
}

func (d *datrie) expansionBase(index int) {
	if index >= len(d.base) {
		newBase := make([]*base, 2*index)
		copy(newBase, d.base)
		d.base = newBase
	}
}

func (d *datrie) setBase(index int, b *base) {
	d.expansionBase(index)
	d.base[index] = b
}

func (d *datrie) getBase(index int) *base {
	d.expansionBase(index)
	if d.base[index] == nil {
		d.base[index] = &base{}
	}
	return d.base[index]
}

func (d *datrie) getCheck(index int) int {
	d.expansionCheck(index)
	return d.check[index]
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

// 没有冲突
func (d *datrie) noConflict(insertPos int, parentIndex int, index int, p *path) {
	// pos位置的字符已经放到base里面，所以跳过这个字符，也是这里pos+1的由来
	path := p.insertPath[insertPos:]

	last := len(p.paramAndHandle) - 1

	b := &base{q: -1, tailPath: path, tailHandler: p.paramAndHandle[insertPos:], handle: p.paramAndHandle[last]}
	d.setCheck(index, parentIndex)
	d.setBase(index, b)
}

func (d *datrie) debug(max int, insertWord string, index, insertPos, base int) {
	fmt.Printf("\n#word(%s) index(%d) insertPos(%d) base(%d)\n", insertWord, index, insertPos, base)
	/*
		fmt.Printf("base %9s ", "")
		for _, v := range d.base[:max] {
			fmt.Printf("[%v]  ", v)
		}
	*/

	fmt.Printf("\n")
	fmt.Printf("q     ")
	for i := 0; i < max; i++ {
		d := d.base[i]
		q := 0
		if d != nil {
			q = d.q
		}

		fmt.Printf("%02d ", q)
	}

	fmt.Printf("\n\n")

	fmt.Printf("index ")
	for i := 0; i < max; i++ {
		fmt.Printf("%02d ", i)
	}
	fmt.Printf("\n\n")

	fmt.Printf("check ")
	for i := 0; i < max; i++ {
		fmt.Printf("%02d ", d.check[i])
	}
	fmt.Printf("\n\n")
}

func (d *datrie) findParamOrWildcard(b *base, path string, p *Params) (h *handle) {

	parentIndex := 0

	var i, j int

	for i, j = 0, 0; i < len(b.tailPath); i++ {

		h = b.tailHandler[i]

		if h != nil && h.paramName != "" {

			if h.wildcard { //通配符
				p.appendKey(h.paramName)
				p.setVal(path[j:len(path)])
				break
			}

			// 单个变量
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

		if j < len(path) {
			if path[j] != b.tailPath[i] {
				return nil
			}
		}

		j++

	}

	return b.tailHandler[len(b.tailHandler)-1]
}

func (d *datrie) lookupTest(path string) (h *handle, p Params) {
	p = make(Params, 0, d.maxParam)
	return d.lookup(path, &p), p
}

func (d *datrie) getIndex(parentIndex int, c byte) int {
	q := 0
	if d.base[parentIndex] != nil {
		q = d.base[parentIndex].q
	}

	return q + getCodeOffset(c)
}

// 查找
func (d *datrie) lookup(path string, p *Params) (h *handle) {

	parentIndex := 1
	var index int

	var b *base

	for k := 0; k < len(path); {

		c := path[k]

		index = d.getIndex(parentIndex, c)

		if index >= len(d.base) {
			return nil
		}

		if d.check[index] != parentIndex {
			return nil
		}

		b := d.base[index]

		// 如果只有一个path，baseHandler里面肯定没有数据，就不需要进入下面的for循环
		if b != nil && b.q > 0 && b.handle != nil && d.path > 1 {
			h := b.handle
			if h.paramName == "" {

				if len(b.tailPath) == 1 && len(path[k+1:]) == 1 && b.tailPath[0] == path[k+1] {
					//这个path是一个更大的path组装部分, 所以b.q > 0
					//但是它已经是最后一个字符了
					return h
				}

				parentIndex = index
				k++

				continue
			}

			i := k + 1
			p.appendKey(h.paramName)

			if h.wildcard { //通配符号
				p.setVal(path[i:len(path)])
				return b.handle
			}

			var j int
			for j = i; j < len(path) && path[j] != '/'; j++ {
			}

			p.setVal(path[i:j])

			if j == len(path) {
				return h
			}

			k = j
			parentIndex = d.base[index].q + getCodeOffset(':')

			continue

		}

		if b := d.base[index]; b != nil && b.q < 0 {

			return d.findParamOrWildcard(b, path[k+1:], p)
		}

		if d.check[index] <= 0 {
			return nil
		}

		parentIndex = index
		k++
	}

	if b != nil {
		return b.handle
	}

	return nil
}

func (d *datrie) setTail(c byte, q int, parentIndex int, p *path, insertPos int) {
	// 修改老的跳转基地址, case3 step 6.
	oldBase := d.getBase(parentIndex)

	oldBase.q = q

	// 计算c要保存的位置
	index := d.base[parentIndex].q + getCodeOffset(c)
	// 记录index的爸爸位置(爸爸都是放到check数组里面的)
	d.setCheck(index, parentIndex)

	newBase := d.getBase(index)

	// 移动tail的字符往前面移动
	if len(oldBase.tailPath) > 0 {
		// oldBase.tailHandler 和oldBase.tailPath是等长的。
		// 所以这里直接判断len(oldBase.tailPath) > 0
		// 保存了handle 可能是param 或者就是这个路径的handle
		newBase.handle = oldBase.tailHandler[0]

		haveHandle := newBase.handle
		var insertHandle *handle
		if insertPos < len(p.paramAndHandle) {
			insertHandle = p.paramAndHandle[insertPos]
		}

		//fmt.Printf("%p\n", p.paramAndHandle[insertPos])
		newBase.tailPath = oldBase.tailPath[1:]
		newBase.tailHandler = oldBase.tailHandler[1:]
		oldBase.tailPath = string(oldBase.tailPath[0])
		oldBase.tailHandler = oldBase.tailHandler[0:1]
		oldBase.handle = oldBase.tailHandler[0]

		// fmt.Printf("newBase.handle:%p, haveHandle:%p, insertHandle:%p\n", newBase.handle, haveHandle, insertHandle)
		if haveHandle != nil && insertHandle != nil {
			if haveHandle.paramName != insertHandle.paramName && len(haveHandle.paramName) > 0 {
				panic(fmt.Sprintf("There is a problem with the inserted path:%s", p.originalPath))
			}

			newBase.handle = insertHandle
			if oldBase.handle.paramName == "" && insertHandle.paramName != "" {
				oldBase.handle.paramName = insertHandle.paramName
			}

			if oldBase.handle.handle == nil && insertHandle.handle != nil {
				oldBase.handle = insertHandle
			}

		}
	}

	newBase.q = -1
}

// 共同前缀冲突
// 有4中情况，
// 1.重复插入, tail里面和insertPath里面是一样的
// 2.tail里面是短的，insertPath里面是长的，tail被包含至insertPath
// 3.tail里面是长的，insertPath里面是短的，tail包含insertPath
// 4.tail和insertPath，有共同前缀，有一个节点分叉出来，引出不同的边长
func (d *datrie) samePrefix(b *base, insertPos, parentIndex int, p *path) (next bool) {
	path := p.insertPath

	if path[insertPos:] == b.tailPath {
		// 重复数据插入, 前缀一样
		// TODO, 选择策略 替换，还是panic,
		// TODO 测试变量不一样的情况 /:name/hello /:name/word 这种直接panic
		return
	}

	insertPos++ //路过一个字符

	insertPath := path[insertPos:]

	// 处理相同前缀
	if len(insertPath) > 0 && len(b.tailPath) > 0 && insertPath[0] == b.tailPath[0] {
		c := b.tailPath[0]
		// 原有的字符在tail数组里面，现在要拖到d.base
		// 先计算一个没有冲突的位置 case3 step 5.
		q := d.xCheck(c)

		d.setTail(c, q, parentIndex, p, insertPos)
		return true
	}

	d.diff(b, insertPos, insertPath, parentIndex, p)
	return false
}

func (d *datrie) diff(oldBase *base, insertPos int, insertPath string, parentIndex int, p *path) {
	// 处理下没有的共同前缀,
	//　主要是情况2, 3, 4
	var list []byte
	tailPath := oldBase.tailPath
	if len(insertPath) > 0 {
		list = append(list, insertPath[0])
	}

	if len(tailPath) > 0 {
		list = append(list, tailPath[0])
	}

	q := d.xCheckArray(list)

	d.base[parentIndex].q = q
	if len(tailPath) > 1 { //tailPath
		d.setTail(tailPath[0], q, parentIndex, p, insertPos)
	}

	if len(insertPath) > 0 {
		index := d.base[parentIndex].q + getCodeOffset(insertPath[0])
		d.noConflict(insertPos+1, parentIndex, index, p)
	}
}

func (d *datrie) findAllChildNode(parentIndex int) (rv []byte) {
	for index, checkParentIndex := range d.check {
		if checkParentIndex == parentIndex {
			// d.base[parentIndex] + offset = index，所以求offset 就是如下
			offset := index - d.base[parentIndex].q
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
	d.base[index] = nil
	d.check[index] = 0
}

func (d *datrie) insertConflict(insertPos int, parentIndex, index int, p *path) {
	var list []byte

	// 两个爸爸结点parentIndex和d.check[index] 都在抢儿子结点index
	list, lessIndex := d.selectList(parentIndex, index)

	// step 5
	tempBase := d.base[lessIndex].q
	d.base[lessIndex].q = d.xCheckArray(list)

	for _, c := range list {
		// step 6 or step 9
		oldNode := tempBase + getCodeOffset(c)
		newNode := d.base[lessIndex].q + getCodeOffset(c)

		d.setBase(newNode, d.base[oldNode])
		d.setCheck(newNode, d.check[oldNode])

		// step 7
		if d.base[oldNode] != nil && d.base[oldNode].q > 0 {
			d.moveToNewParent(oldNode, newNode)
		}

		if parentIndex != lessIndex && oldNode == parentIndex {
			parentIndex = oldNode
		}
		// step 8 or step 10
		d.resetNode(oldNode)
	}

	index = d.base[parentIndex].q + getCodeOffset(p.insertPath[insertPos])
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
	//p.debug()

	d.changePool(p)

	for pos := 0; pos < len(p.insertPath); pos++ {
		c := p.insertPath[pos]
		index := d.base[parentIndex].q + getCodeOffset(c)

		if d.getCheck(index) == 0 {
			d.noConflict(pos+1, parentIndex, index, p)
			return
		}

		// 插入的时候冲突，需要修改 父节点或子节点的接续关系
		if d.check[index] != parentIndex {
			d.insertConflict(pos, parentIndex, index, p)
			return
		}

		if b := d.base[index]; b.q < 0 {
			// tailPos 小于0，说明有共同前缀
			next := d.samePrefix(b, pos, index, p)
			if !next {
				return
			}
		}

		parentIndex = index

	}
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
			offset := i - d.base[oldParent].q
			d.setCheck(d.base[oldParent].q+offset, newParent)
		}
	}

	if !found {
		panic(fmt.Sprintf("not found oldParent:%d", oldParent))
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
