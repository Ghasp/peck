package main

import (
	"fmt"
	"net/http"
)

func modifyHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := parseToken(r)
	if !ok {
		http.Error(w, `token in the authorization header was not found.`, http.StatusBadRequest)
		return
	}
	if token != TOKEN {
		fmt.Println(token)
		requestToken(w)
		return
	}

	stmt, ok := getRequestBody(r)
	if !ok || len(stmt) == 0 {
		http.Error(w, `No sql statement provided. Please provide a sql statment in the http request body`, http.StatusBadRequest)
		return
	}

	err := setData(CONNECTION, string(stmt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
