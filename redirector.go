package main

import (
	"net/http"
)

type redirector struct {
	H http.Handler
	R map[string]string
}

func (r redirector) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(r.R) == 0 {
		r.H.ServeHTTP(rw, req)
		return
	}

	to, ok := r.R[req.Host]
	if !ok {
		r.H.ServeHTTP(rw, req)
		return
	}

	url := req.URL
	url.Host = to
	http.Redirect(rw, req, url.String(), http.StatusFound)
}
