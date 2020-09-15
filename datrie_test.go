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
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request) {
			done++
		})
	}

	for k, word := range insertWord {
		h := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil)
		assert.Equal(t, done, k+1)
	}
}

func Test_lookupAndInsertCase2(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor", "jar"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request) {
			done++
		})
	}

	for k, word := range insertWord {
		h := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil)
		assert.Equal(t, done, k+1)
	}
}

func Test_lookupAndInsertCase3(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor", "jar", "badge"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request) {
			done++
		})
	}

	for k, word := range insertWord {
		h := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil)
		assert.Equal(t, done, k+1)
	}

}

func Test_lookupAndInsertCase4(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"bachelor", "jar", "badge", "baby"}
	for _, word := range insertWord {
		d.insert([]byte(word), func(w http.ResponseWriter, r *http.Request) {
			done++
		})
	}

	for k, word := range insertWord {
		h := d.lookup([]byte(word))

		assert.NotEqual(t, h, (*handle)(nil))
		if h == nil {
			return
		}
		h.handle(nil, nil)
		assert.Equal(t, done, k+1)
	}

}
