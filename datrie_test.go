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

	insertWord := []string{"bachelor"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	for k, word := range insertWord {
		h, _ := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
	}
}

// TODO 打卡注释的测试
func Test_lookupAndInsertCase1_Param(t *testing.T) {

	d := newDatrie()
	done := 0

	insertPath := []string{"/test/word/:name"}
	//insertPath := []string{"/test/word/:name", "/get/word/*name"}
	for _, word := range insertPath {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	lookupPath := []string{"/test/word/aaa",
		"/test/word/bbb",
		"/test/word/ccc",
		/*
			"/get/word/action1",
			"/get/word/action2",
			"/get/word/action3",
			"/get/word/ccc/ddd",
		*/
	}

	needVal := []string{"aaa",
		"bbb",
		"ccc",
		/*
			"action1",
			"action2",
			"action3",
			"ccc/ddd",
		*/
	}

	needKey := "name"
	for k, word := range lookupPath {
		h, p := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}

		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
		assert.Equal(t, p[0].Key, needKey)
		assert.Equal(t, p[0].Value, needVal[k])
	}
}

func Test_lookupAndInsertCase2(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor", "jar"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	for k, word := range insertWord {
		h, _ := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
	}
}

func Test_lookupAndInsertCase3(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor", "jar", "badge"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	for k, word := range insertWord {
		h, _ := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
	}

}

func Test_lookupAndInsertCase3_Param(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"/a/:name", "/j/:name"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		//d.debug(64, word, 0, 0, 0)
		//fmt.Printf("=================\n")
	}

	lookupPath := []string{
		"/a/aaa",
		"/a/bbb",
		"/j/ccc",
		"/j/ddd",
	}

	needKey := "name"

	needVal := []string{
		"aaa",
		"bbb",
		"ccc",
		"ddd",
	}

	for k, word := range lookupPath {
		h, p := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil), fmt.Sprintf("lookup word(%s)", word))
		if h == nil {
			return
		}

		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
		b := assert.Equal(t, p[0].Key, needKey, fmt.Sprintf("lookup key(%s)", needKey))
		if !b {
			break
		}

		b = assert.Equal(t, p[0].Value, needVal[k])
		if !b {
			break
		}
	}
}

func Test_lookupAndInsertCase4(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor", "jar", "badge", "baby"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	for k, word := range insertWord {
		h, _ := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
	}

}
