package baserouter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lookupAndInsertMult_Param(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{"/webaudio/:sid/:createTime", "/gettext/:sid/:createTime"}
	for _, word := range insertWord {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		//d.debug(64, word, 0, 0, 0)
		//fmt.Printf("=================\n")
	}

	lookupPath := []string{
		"/webaudio/web111/web222",
		"/gettext/text111/text222",
	}

	needSidKey := "sid"
	needCreateTimeKey := "createTime"

	needSidValue := []string{
		"web111",
		"text111",
	}

	needCreateTimeValue := []string{
		"web222",
		"text222",
	}

	for k, word := range lookupPath {
		h, p := d.lookupTest(word)

		assert.NotEqual(t, h, (*handle)(nil), fmt.Sprintf("lookup word(%s)", word))
		if h == nil {
			return
		}

		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)
		b := assert.Equal(t, needSidKey, p[0].Key, fmt.Sprintf("lookup key(%s)", needSidKey))
		if !b {
			break
		}

		b = assert.Equal(t, needSidValue[k], p[0].Value)
		if !b {
			break
		}

		b = assert.Equal(t, needCreateTimeKey, p[1].Key, fmt.Sprintf("lookup key(%s)", needSidKey))
		if !b {
			break
		}

		b = assert.Equal(t, needCreateTimeValue[k], p[1].Value)
		if !b {
			break
		}
	}
}
