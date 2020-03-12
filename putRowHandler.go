package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func putRowHandler(w http.ResponseWriter, r *http.Request) {
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

	set, err := parseJSONRowRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	where := make(map[string]string, 0)
	if key, ok := url["key"]; ok {
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

	err = setData(CONNECTION, toUpsert(DATABASE, schema, table, set, where))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
