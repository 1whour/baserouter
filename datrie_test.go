package baserouter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO test
// fail
// /:name
// /aa
func Test_lookupAndInsertCase1(t *testing.T) {

	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor"}
	for _, word := range insertWord {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	for k, word := range insertWord {
		h, _ := d.lookup(word)

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
	}
}

func Test_lookupAndInsertCase1_Param(t *testing.T) {

	d := newDatrie()
	done := 0

	insertPath := []string{"/test/word/:name", "/get/word/*name"}
	//insertPath := []string{"/get/word/*name"}
	for _, word := range insertPath {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		//d.debug(20, word, 0, 0, 0)
	}

	lookupPath := []string{
		"/test/word/aaa",
		"/test/word/bbb",
		"/test/word/ccc",
		"/get/word/action1",
		"/get/word/action2",
		"/get/word/action3",
		"/get/word/ccc/ddd",
	}

	needVal := []string{
		"aaa",
		"bbb",
		"ccc",
		"action1",
		"action2",
		"action3",
		"ccc/ddd",
	}

	needKey := "name"
	for k, word := range lookupPath {
		h, p := d.lookup(word)

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}

		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1, fmt.Sprintf("search word(%s)", word))
		assert.Equal(t, needKey, p[0].Key, fmt.Sprintf("search word(%s)", word))
		assert.Equal(t, needVal[k], p[0].Value, fmt.Sprintf("search word(%s)", word))
	}
}

func Test_lookupAndInsertCase2(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor", "jar"}
	for _, word := range insertWord {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	for k, word := range insertWord {
		h, _ := d.lookup(word)

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
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
		//d.debug(20, word, 0, 0, 0)
	}

	for k, word := range insertWord {
		h, _ := d.lookup(word)

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

	insertWord := []string{"/some/:name/12", "/some/:name/12/cc", "/a/:name", "/j/:name"}
	for _, word := range insertWord {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		//d.debug(90, word, 0, 0, 0)
		//fmt.Printf("=================\n")
	}

	lookupPath := []string{
		"/some/4567/12",
		"/some/1234/12/cc",
		"/a/aaa",
		"/a/bbb",
		"/j/ccc",
		"/j/ddd",
	}

	needKey := "name"

	needVal := []string{
		"4567",
		"1234",
		"aaa",
		"bbb",
		"ccc",
		"ddd",
	}

	for k, word := range lookupPath {

		h, p := d.lookup(word)

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

		b = assert.Equal(t, p[0].Value, needVal[k], fmt.Sprintf("lookup key(%s)", needKey))
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
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
		//d.debug(70, word, 0, 0, 0)
	}

	for k, word := range insertWord {
		h, _ := d.lookup(word)

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
	}

}

func Test_lookupAndInsertCase4_Param(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"/authorizations", "/authorizations/:id", "/repos/:owner/:repo/events"}
	for _, word := range insertWord {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		d.debug(64, word, 0, 0, 0)
		//fmt.Printf("=================\n")
	}

	lookupPath := []string{
		"/authorizations",
		"/authorizations/123",
		"/repos/antlabs/baserouter/events",
	}

	needKeyArr := [][]string{
		[]string{},
		[]string{"id"},
		[]string{"owner", "repo"},
	}

	needValArr := [][]string{
		[]string{},
		[]string{"123"},
		[]string{"antlabs", "baserouter"},
	}

	for k, word := range lookupPath {

		h, p := d.lookup(word)

		assert.NotEqual(t, h, (*handle)(nil), fmt.Sprintf("lookup word(%s)", word))
		if h == nil {
			return
		}

		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)

		for index, needKey := range needKeyArr[k] {
			needVal := needValArr[k]
			b := assert.Equal(t, p[index].Key, needKey, fmt.Sprintf("lookup key(%s)", needKey))
			if !b {
				break
			}

			b = assert.Equal(t, p[index].Value, needVal[index], fmt.Sprintf("lookup key(%s)", needKey))
			if !b {
				break
			}
		}
	}
}
