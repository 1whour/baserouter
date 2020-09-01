package baserouter

import (
	"bytes"
	"fmt"
)

type handle struct {
	handle handleFunc
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

	copy(d.tail[d.pos:], path)
	d.head[d.pos] = len(path)
	d.check[base] = prevIndex
	d.base[base] = -d.pos
	if need != 0 {
		newHandler := make([]*handle, need)
		copy(newHandler, d.handler)
		d.handler = newHandler
	}

	d.handler[d.pos+len(path)] = &handle{handle: h}
	//fmt.Printf("insert index:%d\n", d.pos+len(path))
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
			//fmt.Printf("(%s), (%s) start:%d head:%d\n", path[k:], d.tail[start:start+l], start, l)
			if bytes.Equal(path[k:], d.tail[start:start+l]) {
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
	l := d.head[start]

	if bytes.Equal(path[pos:], d.tail[start:start+l]) {
		//TODO, 选择策略 替换，还是panic
		return
	}

	temp := start
	insertPath := path[pos:]
	savePath := d.tail[start : start+l]

	i := 0
	// 处理相同前缀
	for ; insertPath[i] == savePath[i]; i++ {
		q := d.xCheck(insertPath[i])
		d.base[base] = q

		fmt.Printf("d.base[base]->%d: base->%d\n", d.base[base], base)

		d.check[d.base[base]+getCodeOffset(insertPath[i])] = base
		//start = d.base[q]
		break

		fmt.Printf("start:%d\n", d.base[q])
	}

	// 处理不同前缀
	_ = temp
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

func (d *datrie) xCheck(c byte) (q int) {
	for d.check[q+getCodeOffset(c)] != 0 {
		q++
	}

	return q

}
