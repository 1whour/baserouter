package baserouter

import (
	"bytes"
)

type handle struct {
	handle handleFunc
	path   string
}

type datrie struct {
	base    []int
	check   []int
	tail    []byte
	head    []int
	handler []*handle
	pos     int
}

// 初始化函数
func newDatrie() *datrie {
	d := &datrie{
		base:    make([]int, 1024),
		check:   make([]int, 1024),
		tail:    make([]byte, 1024),
		head:    make([]int, 1024),
		handler: make([]*handle, 1024),
	}

	d.base[0] = 1
	d.pos = 1
	return d
}

// 没有冲突
func (d *datrie) noConflict(index int, path []byte, prevIndex int, base int, h handleFunc) {
	oldPath := path
	path = path[1:]

	d.expansionTailAndHandler(path)

	copy(d.tail[d.pos:], path)
	d.head[d.pos] = len(path)
	d.check[base] = prevIndex
	d.base[base] = -d.pos

	d.handler[d.pos+len(path)] = &handle{handle: h, path: string(oldPath) /*TODO*/}
	d.pos += len(path)
}

// 查找
func (d *datrie) lookup(path []byte) *handle {

	prevIndex := 1
	for k, c := range path {
		base := d.base[prevIndex] + getCodeOffset(c)
		if d.check[base] <= 0 {
			return nil
		}

		if start := d.base[base]; start < 0 {
			start = -start
			l := d.head[start]

			//fmt.Printf("%s:%s\n", path[k+1:], d.tail[start:start+l])
			if bytes.Equal(path[k+1:], d.tail[start:start+l]) {
				return d.handler[start+l]
			}
			return nil
		}

		prevIndex = base

	}

	return nil
}

// 共同前缀冲突
func (d *datrie) samePrefix(path []byte, pos, start int, base int, h handleFunc) {
	start = -start
	temp := start
	l := d.head[start]

	pos++
	if bytes.Equal(path[pos:], d.tail[start:start+l]) {
		//TODO, 选择策略 替换，还是panic
		return
	}

	insertPath := path[pos:]
	savePath := d.tail[start : start+l]

	i := 0
	// 处理相同前缀
	for ; insertPath[i] == savePath[i]; i++ {
		q := d.xCheck(insertPath[i]) //找出可以跳转的位置 , case3 step 5.
		d.base[base] = q             //修改老的跳转位置, case3 step 6.
		d.check[d.base[base]+getCodeOffset(insertPath[i])] = base

		base = d.base[base] + getCodeOffset(insertPath[i])

	}

	// 处理不同前缀
	q := d.xCheckTwo(insertPath[i], savePath[i])
	d.base[base] = q
	newBase := d.base[base] + getCodeOffset(savePath[i])
	d.base[newBase] = -temp
	d.check[newBase] = q
	savePath = savePath[i+1:]

	// case3 step 9
	copy(d.tail[temp:], savePath)
	copy(d.handler[temp:], d.handler[temp:temp+d.head[temp]])
	d.handler[temp+len(savePath)] = d.handler[temp+d.head[temp]]
	for i := len(savePath); i < d.head[start]; i++ {
		d.tail[temp+i] = '?'
		d.handler[temp+i+1] = nil
	}
	d.head[temp] = len(savePath)

	d.expansionTailAndHandler(insertPath[i+1:])
	// case3 step 10
	newBase = d.base[base] + getCodeOffset(insertPath[i])
	d.base[newBase] = -d.pos
	d.check[newBase] = base
	copy(d.tail[d.pos:], insertPath[i+1:])
	d.head[d.pos] = len(insertPath[i+1:])

	// case3 step 11
	d.pos += len(insertPath[i+1:])
	d.handler[d.pos] = &handle{handle: h, path: string(path)}
}

// 插入
func (d *datrie) insert(path []byte, h handleFunc) {
	prevIndex := 1
	for pos, c := range path {
		base := d.base[prevIndex] + getCodeOffset(c)
		if base >= len(d.base) {
			// 扩容
			d.expansion(base)
		}

		if d.check[base] == 0 {
			d.noConflict(pos, path[pos:], prevIndex, base, h)
			return
		}

		if start := d.base[base]; start < 0 {
			d.samePrefix(path, pos, start, base, h)
			return
		}
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

// 找空位置
func (d *datrie) xCheckTwo(c1, c2 byte) (q int) {
	q = 1
	for d.check[q+getCodeOffset(c1)] != 0 || d.check[q+getCodeOffset(c2)] != 0 {
		q++
	}

	return q
}

// 找空位置
func (d *datrie) xCheck(c byte) (q int) {
	q = 1
	for d.check[q+getCodeOffset(c)] != 0 {
		q++
	}

	return q

}
