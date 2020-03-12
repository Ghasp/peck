package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func consoleHandler(w http.ResponseWriter, r *http.Request) {
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

	conn, err := connectDB(SERVER, DATABASE, USERNAME, PASSWORD)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	results, err := unsafeSQL(conn, string(stmt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, string(data))
}
