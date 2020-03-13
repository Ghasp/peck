package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	// settings
	readEnviormentVarialbes()
	readCommandlineParameters()

	var err error
	CONNECTION, err = connectDB(SERVER, DATABASE, USERNAME, PASSWORD)
	if err != nil {
		panic(err)
	}
	defer CONNECTION.Close()

	r := mux.NewRouter()

	// todo: implement build in documentation
	//r.HandleFunc("/", helpHandler).Methods("GET")

	// executes sql commands
	r.HandleFunc("/console", consoleHandler).Methods("POST").Headers("Content-Type", "application/sql")
	r.HandleFunc("/query", queryHandler).Methods("POST").Headers("Content-Type", "application/sql")
	r.HandleFunc("/modify", modifyHandler).Methods("POST").Headers("Content-Type", "application/sql")

	// gets a single row
	r.HandleFunc("/{schema}/{table}", getRowHandler).Methods("GET")
	r.HandleFunc("/{schema}/{table}/{key}", getRowHandler).Methods("GET")

	// creates and updates a single row
	r.HandleFunc("/{schema}/{table}", putRowHandler).Methods("PUT").Headers("Content-Type", "application/json")
	r.HandleFunc("/{schema}/{table}/{key}", putRowHandler).Methods("PUT").Headers("Content-Type", "application/json")

	// deletes a single row
	r.HandleFunc("/{schema}/{table}", deleteRowHandler).Methods("DELETE")
	r.HandleFunc("/{schema}/{table}/{key}", deleteRowHandler).Methods("DELETE")

	// get an array or row using a tables roeign key.
	r.HandleFunc("/{schema}/{table}/join", getRowsHandler).Methods("GET")

	// todo: implement bulk import/update
	//r.HandleFunc("/{schema}/{table}", postRowHandler).Methods("POST").Headers("Content-Type", "application/json")

	// RUN HTTP SERVER
	if TLSCERTPATH != "" && TLSKEYPATH != "" {
		http.ListenAndServeTLS(":443", TLSCERTPATH, TLSKEYPATH, r)
	} else {
		http.ListenAndServe(":80", r)
	}
}
