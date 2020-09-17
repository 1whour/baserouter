package baserouter

import (
	"bytes"
	"fmt"
)

type path struct {
	originalPath []byte    //原始路径
	insertPath   []byte    //修改后的路径，单个变量变为: 所有变量变为*
	paramPath    []*handle //存放param
}

func genPath(p []byte) *path {
	p2 := &path{}
	p2.originalPath = p

	var paramName bytes.Buffer
	var insertPath bytes.Buffer

	foundParam := false

	defer func() {
		if insertPath.Len() > 0 {
			p2.insertPath = insertPath.Bytes()
		}
	}()

	for i := 0; i < len(p); i++ {
		if p[i] == '/' && !foundParam {
			if i+1 >= len(p) {
				insertPath.WriteByte('/')
				continue
			}

			if p[i+1] == ':' {
				foundParam = true
				i++
				insertPath.WriteString("/:")
				continue
			}

			if p[i+1] == '*' {
				foundParam = true
				i++
				insertPath.WriteString("/*")
				continue
			}
		}

		if foundParam {
			if p[i] == '/' {
				foundParam = false
				if paramName.Len() == 0 {
					panic(fmt.Sprintf("wildcards must be named with a non-empty name in path:%s",
						p))
				}

				if p2.paramPath == nil {
					p2.paramPath = make([]*handle, len(p))
				}

				p2.paramPath[insertPath.Len()-1] = &handle{paramName: paramName.String()}
				insertPath.WriteByte('/')

				continue
			}

			paramName.WriteByte(p[i])
			continue
		}

		insertPath.WriteByte(p[i])

	}

	return p2
}
