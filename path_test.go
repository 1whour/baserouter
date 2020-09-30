package baserouter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPathCase struct {
	got        path
	need       path
	checkParam func(paramPath []*handle) bool
}

type testFailPathCase struct {
	path string
}

// 测试异常情况
func Test_GenPath_Fail(t *testing.T) {
	for _, p := range []testFailPathCase{
		// 1.先catch all 然后单个变量
		{
			path: string("/test/*name/:last"),
		},
		// 1.先catch all , catch all
		{
			path: string("/test/*name/*last"),
		},
	} {

		assert.Panics(t, func() {

			genPath(p.path, nil)
		})
	}

}

func Test_GenPath(t *testing.T) {

	for i, p := range []testPathCase{
		{
			need: path{
				originalPath: "/test/:name/last",
				insertPath:   "/test/:/last",
			},
			checkParam: func(paramPath []*handle) bool {
				param := paramPath[len("/test/")]
				b := assert.NotNil(t, param)
				if !b {
					return b
				}

				return assert.Equal(t, param.paramName, "name")
			},
		},
		{
			need: path{
				originalPath: "/test/*last",
				insertPath:   "/test/*",
			},
			checkParam: func(paramPath []*handle) bool {
				param := paramPath[len("/test/")]
				b := assert.NotNil(t, paramPath[len("/test/")])
				if !b {
					return b
				}

				return assert.Equal(t, param.paramName, "last")
			},
		},
		{
			need: path{
				originalPath: "/test/:name/*last",
				insertPath:   "/test/:/*",
			},
			checkParam: func(paramPath []*handle) bool {
				param := paramPath[len("/test/")]
				b := assert.NotNil(t, paramPath[len("/test/")])

				if !b {
					return b
				}
				assert.Equal(t, param.paramName, "name")

				catchAll := paramPath[len("/test/:/")]
				b = assert.NotNil(t, paramPath[len("/test/:/")])
				if !b {
					return b
				}
				return assert.Equal(t, catchAll.paramName, "last")
			},
		},
	} {

		got := genPath(p.need.originalPath, nil)
		assert.Equal(t, p.need.originalPath, got.originalPath, fmt.Sprintf("-->test index:%d", i))
		assert.Equal(t, p.need.insertPath, got.insertPath, fmt.Sprintf("-->test index:%d", i))

		if !p.checkParam(got.paramAndHandle) {
			t.Logf("test index:%d\n", i)
			break
		}
	}

}
