package baserouter

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_router_easy(t *testing.T) {
	router := New()

	total := int32(0)
	router.GET("/GET", func(w http.ResponseWriter, req *http.Request, _ Params) {
		atomic.AddInt32(&total, 1)
	})

	router.HEAD("/HEAD", func(w http.ResponseWriter, req *http.Request, _ Params) {
		atomic.AddInt32(&total, 1)
	})

	router.POST("/POST", func(w http.ResponseWriter, req *http.Request, _ Params) {
		atomic.AddInt32(&total, 1)
	})

	router.PATCH("/PATCH", func(w http.ResponseWriter, req *http.Request, _ Params) {
		atomic.AddInt32(&total, 1)
	})

	router.DELETE("/DELETE", func(w http.ResponseWriter, req *http.Request, _ Params) {
		atomic.AddInt32(&total, 1)
	})

	router.OPTIONS("/OPTIONS", func(w http.ResponseWriter, req *http.Request, _ Params) {
		atomic.AddInt32(&total, 1)
	})

	ts := httptest.NewServer(http.HandlerFunc(router.ServeHTTP))
	defer ts.Close()

	for _, method := range []string{"GET", "HEAD", "POST", "PATCH", "DELETE", "OPTIONS"} {

		req, err := http.NewRequest(method, ts.URL+"/"+method, nil)

		if err != nil {
			panic(err.Error())
		}

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			panic(err.Error())
		}
	}
	assert.Equal(t, total, int32(6))
}
