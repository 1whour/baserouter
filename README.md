# baserouter
baserouter和httprouter定位是比较类似的库，借鉴了httprouter的API设计，底层使用不同的算法和数据结构，它是一种新的尝试，看性能上能否更快

[![Go](https://github.com/antlabs/baserouter/workflows/Go/badge.svg)](https://github.com/antlabs/baserouter/actions)
[![codecov](https://codecov.io/gh/antlabs/baserouter/branch/master/graph/badge.svg)](https://codecov.io/gh/antlabs/baserouter)

## feature
**近似零拷贝** 只有在需要分配参数，才有可能从堆上分配内存。

**高性能** 现在某些指标目前比httprouter慢10ns

## quick start
```go
package main

import (
    "fmt"
    "net/http"
    "log"

    "github.com/antlabs/baserouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
    router := baserouter.New()
    router.GET("/", Index)
    router.GET("/hello/:name", Hello)

    log.Fatal(http.ListenAndServe(":8080", router))
}
```
