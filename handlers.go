package main

import "net/http"

func healthHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")

	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte(http.StatusText(http.StatusOK)))
}
