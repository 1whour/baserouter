package baserouter

import (
	"net/http"
)

type router struct {
	method
}

func New() *router {
	r := &router{}
	r.init()
	return r
}

func (r *router) GET(path string, handle HandleFunc) {
	r.Handle(http.MethodGet, path, handle)
}

func (r *router) HEAD(path string, handle HandleFunc) {
	r.Handle(http.MethodHead, path, handle)
}

func (r *router) POST(path string, handle HandleFunc) {
	r.Handle(http.MethodPost, path, handle)
}

func (r *router) PUT(path string, handle HandleFunc) {
	r.Handle(http.MethodPut, path, handle)
}

func (r *router) PATCH(path string, handle HandleFunc) {
	r.Handle(http.MethodPatch, path, handle)
}

func (r *router) DELETE(path string, handle HandleFunc) {
	r.Handle(http.MethodDelete, path, handle)
}

func (r *router) OPTIONS(path string, handle HandleFunc) {
	r.Handle(http.MethodOptions, path, handle)
}

func (r *router) Handle(method, path string, handle HandleFunc) {
	r.save(method, path, handle)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	datrie := r.getDatrie(req.Method)
	if datrie != nil {
		h, p := datrie.lookup(path)
		if h != nil {
			h.handle(w, req, p)
			return
		}
	}

	http.NotFound(w, req)
}
