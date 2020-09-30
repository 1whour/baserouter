package baserouter

import "net/http"

type HandleFunc func(w http.ResponseWriter, r *http.Request, p Params)
