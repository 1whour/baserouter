package baserouter

import (
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

func Test_lookupAndInsertCase1_param(t *testing.T) {

	d := newDatrie()
	done := 0

	insertPath := []string{"/test/word/:name"}
	for _, word := range insertPath {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
	}

	lookupPath := []string{"/test/word/aaa", "/test/word/bbb", "/test/word/ccc"}
	needVal := []string{"aaa", "bbb", "ccc"}
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
