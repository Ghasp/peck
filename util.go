package main

import (
	"io/ioutil"
	"net/http"
)

func getRequestBody(r *http.Request) ([]byte, bool) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, false
	}
	return b, true
}
