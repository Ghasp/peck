package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	var err error
	CONNECTION, err = connectDB(SERVER, DATABASE, USERNAME, PASSWORD)
	if err != nil {
		panic(err)
	}
	defer CONNECTION.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")

	r.HandleFunc("/console", consoleHandler).Methods("POST").Headers("Content-Type", "application/sql")
	r.HandleFunc("/query", queryHandler).Methods("POST").Headers("Content-Type", "application/sql")
	r.HandleFunc("/modify", modifyHandler).Methods("POST").Headers("Content-Type", "application/sql")

	r.HandleFunc("/{schema}/{table}", getRowHandler).Methods("GET")
	r.HandleFunc("/{schema}/{table}/{key}", getRowHandler).Methods("GET")
	r.HandleFunc("/{schema}/{table}", putRowHandler).Methods("PUT").Headers("Content-Type", "application/json")
	r.HandleFunc("/{schema}/{table}/{key}", putRowHandler).Methods("PUT").Headers("Content-Type", "application/son")
	r.HandleFunc("/{schema}/{table}", postRowHandler).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{schema}/{table}", deleteRowHandler).Methods("DELETE")

	http.ListenAndServe("localhost:8080", r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello")
}
