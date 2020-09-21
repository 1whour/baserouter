package baserouter

import "net/http"

type handleFunc func(w http.ResponseWriter, r *http.Request, p Params)
