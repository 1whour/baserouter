package baserouter

import (
	"fmt"
	"strings"
)

type path struct {
	originalPath   string    //原始路径
	insertPath     string    //修改后的路径，单个变量变为: 所有变量变为*
	paramAndHandle []*handle //存放param
}

func (p *path) debug(max int) {
	fmt.Printf("paramAndHandle: %5s %v\n", "", p.paramAndHandle[:max])
	fmt.Printf("originalPath bytes: %1s %v\n", "", p.originalPath[:max])
	fmt.Printf("originalPath string: %0s %s\n", "", p.originalPath[:max])
	fmt.Printf("insertPath bytes: %3s %v\n", "", p.insertPath[:max])
	fmt.Printf("insertPath string: %2s %s\n", "", p.insertPath[:max])
}

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
					continue
				}

				if p[i] == '*' {
					wildcard = true
					insertPath.WriteString("*")
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

				p2.addParamPath(insertPath, paramName)

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

		p2.addParamPath(insertPath, paramName)
	}

	if foundParam {
		p2.checkParam(paramName)
		p2.addParamPath(insertPath, paramName)
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

func (p *path) addParamPath(insertPath, paramName strings.Builder) {

	p.paramAndHandle[insertPath.Len()-1] = &handle{paramName: paramName.String()}
}
