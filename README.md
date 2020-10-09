# baserouter
baserouter和httprouter定位是比较类似的库，借鉴了httprouter的API设计，底层使用不同的算法和数据结构，它是一种新的尝试，看性能上能否更快


## quick start
```
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
