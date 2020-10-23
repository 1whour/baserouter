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

	for _, word := range insertPath {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		d.debug(20, word, 0, 0, 0)
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
