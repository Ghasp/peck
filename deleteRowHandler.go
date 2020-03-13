package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func deleteRowHandler(w http.ResponseWriter, r *http.Request) {
	url := mux.Vars(r)
	schema := url["schema"]
	table := url["table"]

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

	where := make(map[string]string, 0)
	if key, ok := url["key"]; ok && PRIMARYKEY != "" {
		where[PRIMARYKEY] = key
	} else {
		for k, v := range r.URL.Query() {
			where[k] = v[0]
		}
	}
	if len(where) < 1 {
		http.Error(w, `row index not specified`, http.StatusBadRequest)
		return
	}

	err := remData(CONNECTION, toDelete(DATABASE, schema, table, where))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
