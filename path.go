package baserouter

import (
	"fmt"
	"strings"
)

type path struct {
	originalPath   string    //原始路径
	insertPath     string    //修改后的路径，单个变量变为: 所有变量变为*
	paramAndHandle []*handle //存放param
	maxParam       int
}

func (p *path) debug() {
	max := len(p.insertPath)
	paramMax := max
	if paramMax < len(p.originalPath) {
		paramMax = len(p.originalPath)
	}

	fmt.Printf("(%d) paramAndHandle: %5s %v\n", cap(p.paramAndHandle), "", p.paramAndHandle[:max])
	fmt.Printf("(%d) originalPath bytes: %1s %v\n", len(p.originalPath), "", p.originalPath[:paramMax])
	fmt.Printf("(%d) originalPath string: %0s %s\n", len(p.originalPath), "", p.originalPath[:paramMax])
	fmt.Printf("(%d) insertPath bytes: %3s %v\n", len(p.insertPath), "", p.insertPath[:max])
	fmt.Printf("(%d) insertPath string: %2s %s\n", len(p.insertPath), "", p.insertPath[:max])
}

// 基于插入是比较少的场景，所以genPath函数没有做性能优化
func genPath(p string, h HandleFunc) *path {
	p2 := &path{}
	p2.originalPath = p
	p2.paramAndHandle = make([]*handle, len(p2.originalPath))

	var paramName strings.Builder
	var insertPath strings.Builder

	foundParam := false
	wildcard := false
	maybeVar := false

	for i := 0; i < len(p); i++ {
		if !wildcard && !foundParam {
			if p[i] == '/' && !maybeVar {
				maybeVar = true
				insertPath.WriteByte('/')
				continue
			}
		}

		if maybeVar {
			maybeVar = false
			if !foundParam && !wildcard {

				if p[i] == ':' {
					foundParam = true
					insertPath.WriteString(":")
					p2.maxParam++
					continue
				}

				if p[i] == '*' {
					wildcard = true
					insertPath.WriteString(":")
					p2.maxParam++
					continue
				}
			}
		}

		if wildcard {
			if p[i] == '/' || foundParam {
				panic(fmt.Sprintf("catch-all routes are only allowed at the end of the path in path '%s'", p))
			}

			paramName.WriteByte(p[i])
			continue
		}

		if foundParam {

			if p[i] == '/' {
				foundParam = false
				maybeVar = true

				p2.checkParam(paramName)

				p2.addParamPath(insertPath, paramName, false)

				insertPath.WriteByte('/')

				paramName.Reset()
				continue
			}

			paramName.WriteByte(p[i])
			continue
		}

		insertPath.WriteByte(p[i])

	}

	if wildcard {

		p2.checkParam(paramName)

		p2.addParamPath(insertPath, paramName, wildcard)
	}

	if foundParam {
		p2.checkParam(paramName)
		p2.addParamPath(insertPath, paramName, false)
	}

	if insertPath.Len() > 0 {
		p2.insertPath = insertPath.String()
	}

	p2.addHandle(insertPath, h)
	return p2
}

func (p *path) checkParam(paramName strings.Builder) {
	if paramName.Len() == 0 {
		panic(fmt.Sprintf("wildcards must be named with a non-empty name in path:%s",
			p.originalPath))
	}
}

func (p *path) addHandle(insertPath strings.Builder, h HandleFunc) {
	index := insertPath.Len() - 1
	if p.paramAndHandle[index] == nil {
		p.paramAndHandle[index] = &handle{handle: h, path: string(p.originalPath)}
	} else {
		p.paramAndHandle[index].handle = h
		p.paramAndHandle[index].path = string(p.originalPath)
	}

	p.paramAndHandle = p.paramAndHandle[:insertPath.Len()]
}

func (p *path) addParamPath(insertPath, paramName strings.Builder, wildcard bool) {

	p.paramAndHandle[insertPath.Len()-1] = &handle{paramName: paramName.String(), wildcard: wildcard}
}
