package baserouter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	insertPath string
	lookupPath string
	paramKey   []string
	paramValue []string
}

type testCases []testCase

func (tcs *testCases) run(t *testing.T) {
	d := newDatrie()
	done := 0

	for index, tc := range *tcs {
		d.insert(tc.insertPath, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
		d.debug(80, tc.insertPath, index, 0, 0)
	}

	for k, tc := range *tcs {

		h, p := d.lookup(tc.lookupPath)

		cb := func() {
			assert.NotEqual(t, h, (*handle)(nil), fmt.Sprintf("lookup word(%s)", tc.lookupPath))
			if h == nil {
				return
			}

			fmt.Printf("lookup address:%p\n", h)
			h.handle(nil, nil, nil)
			assert.Equal(t, done, k+1)

			for index, needKey := range tc.paramKey {
				if len(needKey) == 0 {
					fmt.Printf("index = %d, needKey = 0\n", k)
					continue
				}

				needVal := tc.paramValue
				b := assert.Equal(t, p[index].Key, needKey, fmt.Sprintf("lookup key(%s)", needKey))
				if !b {
					return
				}

				b = assert.Equal(t, p[index].Value, needVal[index], fmt.Sprintf("lookup key(%s)", needKey))
				if !b {
					return
				}
			}
		}

		b := assert.NotPanics(t, cb, fmt.Sprintf("lookup path is(%s)", tc.lookupPath))
		if !b {
			break
		}
	}
}
