package baserouter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_github_lookupAndInsertCase3_Param(t *testing.T) {
	d := newDatrie()
	done := 0

	//insertWord := []string{"/authorizations/:id", "/applications/:client_id/tokens"}
	insertWord := []string{"/authorizations/:id", "/applications/:client_id/tokens", "/applications/:client_id/tokens/:access_token"}
	for _, word := range insertWord {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		//d.debug(90, word, 0, 0, 0)
		//fmt.Printf("=================\n")
	}

	lookupPath := []string{
		"/authorizations/12",
		"/applications/client_id-aaa/tokens",
		"/applications/client_id-bbb/tokens/access_token-aaa",
	}

	needKeyArr := [][]string{
		[]string{"id"},
		[]string{"client_id"},
		[]string{"client_id", "access_token"},
	}

	needValArr := [][]string{
		[]string{"12"},
		[]string{"client_id-aaa"},
		[]string{"client_id-bbb", "access_token-aaa"},
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
