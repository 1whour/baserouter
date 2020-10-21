package baserouter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_github_lookupAndInsertCase3_Param0(t *testing.T) {
	d := newDatrie()
	done := 0

	//insertWord := []string{"/authorizations/:id", "/applications/:client_id/tokens"}
	insertWord := []string{
		"/teams/:id/repos",
		"/teams/:id/repos/:owner/:repo",
		//"/repos/:owner/:repo/pulls/:number/files",
		//"/repos/:owner/:repo/pulls/:number/merge",
		//"/repos/:owner/:repo/pulls/:number/comments",
	}
	for _, word := range insertWord {
		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})

		d.debug(90, word, 0, 0, 0)
		//fmt.Printf("=================\n")
	}

	lookupPath := []string{
		"/teams/antlabs/repos",
		"/teams/antlabs-aaa/repos/guonaihong/baserouter-aaa",
		//"/repos/guonaihong/baserouter/pulls/1/files",
		//"/repos/NaihongGuo/deepcopy/pulls/2/merge",
		//"/repos/guonh/timer/pulls/3/comments",
	}

	needKeyArr := [][]string{
		[]string{"id"},
		[]string{"id", "owner", "repo"},
		//[]string{"owner", "repo", "number"},
		//[]string{"owner", "repo", "number"},
		//[]string{"owner", "repo", "number"},
	}

	needValArr := [][]string{
		[]string{"antlabs"},
		[]string{"antlabs-aaa", "guonaihong", "baserouter-aaa"},
		[]string{"guonaihong", "baserouter", "1"},
		[]string{"NaihongGuo", "deepcopy", "2"},
		//[]string{"guonh", "timer", "3"},
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

func Test_github_lookupAndInsertCase3_Param1(t *testing.T) {
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

func Test_github_lookupAndInsertCase4_Param(t *testing.T) {
	d := newDatrie()
	done := 0

	insertWord := []string{
		"/authorizations",
		"/authorizations/:id",
		"/applications/:client_id/tokens/:access_token",
		"/repos/:owner/:repo/events",
		"/orgs/:org/events",
	}

	for _, word := range insertWord {
		/*
			d.debug(90, word, 0, 0, 0)
			fmt.Printf("insert start=================\n")
		*/

		d.insert(word, func(w http.ResponseWriter, r *http.Request, p Params) {
			done++
		})
		/*
			d.debug(90, word, 0, 0, 0)

			fmt.Printf("insert end=================\n")
		*/

	}

	lookupPath := []string{
		"/authorizations",
		"/authorizations/12",
		"/applications/client_id-bbb/tokens/access_token-aaa",
		"/repos/guonaihong/baserouter/events",
		"/orgs/antlabs/events",
	}

	needKeyArr := [][]string{
		[]string{""},
		[]string{"id"},
		[]string{"client_id", "access_token"},
		[]string{"owner", "repo"},
		[]string{"org"},
	}

	needValArr := [][]string{
		[]string{""},
		[]string{"12"},
		[]string{"client_id-bbb", "access_token-aaa"},
		[]string{"guonaihong", "baserouter"},
		[]string{"antlabs"},
	}

	for k, word := range lookupPath {

		h, p := d.lookup(word)

		assert.NotEqual(t, h, (*handle)(nil), fmt.Sprintf("lookup word(%s)", word))
		if h == nil {
			return
		}

		h.handle(nil, nil, nil)
		assert.Equal(t, done, k+1)

		if len(p) == 0 {
			fmt.Printf("search word(%s),k = %d\n", word, k)
		}

		for index, needKey := range needKeyArr[k] {
			if len(needKey) == 0 {
				fmt.Printf("index = %d, needKey = 0\n", k)
				continue
			}

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
