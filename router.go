package baserouter

import "net/http"

type router struct {
	method
}

func New() *router {
	r := &router{}
	r.init()
	return r
}

func (r *router) GET(path string, handle handleFunc) {
	r.handleCore(http.MethodGet, path, handle)
}

func (r *router) HEAD(path string, handle handleFunc) {
	r.handleCore(http.MethodHead, path, handle)
}

func (r *router) POST(path string, handle handleFunc) {
	r.handleCore(http.MethodPost, path, handle)
}

func (r *router) PUT(path string, handle handleFunc) {
	r.handleCore(http.MethodPut, path, handle)
}

func (r *router) PATCH(path string, handle handleFunc) {
	r.handleCore(http.MethodPatch, path, handle)
}

func (r *router) DELETE(path string, handle handleFunc) {
	r.handleCore(http.MethodDelete, path, handle)
}

func (r *router) OPTIONS(path string, handle handleFunc) {
	r.handleCore(http.MethodOptions, path, handle)
}

func (r *router) handleCore(method, path string, handle handleFunc) {
	r.save(method, path, handle)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	datrie := r.getDatrie(req.Method)
	if datrie != nil {
		h := datrie.lookup([]byte(path))
		if h != nil {
			h.handle(w, req)
			return
		}
	}

	http.NotFound(w, req)
}
