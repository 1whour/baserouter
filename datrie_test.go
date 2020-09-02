package baserouter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lookupAndInsertCase1(t *testing.T) {
	d := newDatrie()
	done := 0
	d.insert([]byte("bachelor"), func(w http.ResponseWriter, r *http.Request) {
		done = 1
	})

	/*
		fmt.Printf("base %v\n", d.base[:20])
		fmt.Printf("check %v\n", d.check[:20])
		fmt.Printf("check %s\n", d.tail[:20])
		fmt.Printf("check %v\n", d.head[:20])
	*/

	h := d.lookup([]byte("bachelor"))

	assert.NotEqual(t, h, (*handle)(nil))
	if h == nil {
		return
	}
	h.handle(nil, nil)
	assert.Equal(t, done, 1)
}

func Test_lookupAndInsertCase2(t *testing.T) {
	d := newDatrie()
	done := 0
	d.insert([]byte("bachelor"), func(w http.ResponseWriter, r *http.Request) {
		done++
	})

	d.insert([]byte("jar"), func(w http.ResponseWriter, r *http.Request) {
		done++
	})

	/*
		fmt.Printf("base %v\n", d.base[:20])
		fmt.Printf("check %v\n", d.check[:20])
		fmt.Printf("check %s\n", d.tail[:20])
		fmt.Printf("check %v\n", d.head[:20])
	*/

	h := d.lookup([]byte("bachelor"))

	assert.NotEqual(t, h, (*handle)(nil))
	if h == nil {
		return
	}
	h.handle(nil, nil)
	assert.Equal(t, done, 1)

	h = d.lookup([]byte("jar"))

	assert.NotEqual(t, h, (*handle)(nil))
	if h == nil {
		return
	}
	h.handle(nil, nil)
	assert.Equal(t, done, 2)
}

func Test_lookupAndInsertCase3(t *testing.T) {
	d := newDatrie()
	done := 0
	d.insert([]byte("bachelor"), func(w http.ResponseWriter, r *http.Request) {
		done++
	})

	fmt.Printf("base %v\n", d.base[:20])
	fmt.Printf("check %v\n", d.check[:20])
	fmt.Printf("tail %s\n", d.tail[:20])
	fmt.Printf("head %v\n", d.head[:20])
	fmt.Printf("handler %v\n", d.handler[:20])
	fmt.Printf("==================\n")

	d.insert([]byte("badge"), func(w http.ResponseWriter, r *http.Request) {
		done++
	})

	fmt.Printf("base %v\n", d.base[:20])
	fmt.Printf("check %v\n", d.check[:20])
	fmt.Printf("tail %s\n", d.tail[:20])
	fmt.Printf("head %v\n", d.head[:20])
	fmt.Printf("handler %v\n", d.handler[:20])

	h := d.lookup([]byte("bachelor"))

	assert.NotEqual(t, h, (*handle)(nil))
	if h == nil {
		return
	}
	h.handle(nil, nil)
	assert.Equal(t, done, 1)

	h = d.lookup([]byte("badge"))

	assert.NotEqual(t, h, (*handle)(nil))
	if h == nil {
		return
	}
	h.handle(nil, nil)
	assert.Equal(t, done, 2)
}
