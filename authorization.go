package main

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func parseBasicAuth(r *http.Request) (string, string, bool) {
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		return "", "", false
	}

	payload, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		return "", "", false
	}

	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 || len(pair[0]) < 1 || len(pair[1]) < 1 {
		return "", "", false
	}
	return pair[0], pair[1], true
}

func parseToken(r *http.Request) (string, bool) {
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Bearer" {
		return "", false
	}

	if auth[1] == "" {
		return "", false
	}

	return auth[1], true
}

func requestToken(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", "Bearer realm=\"Token\"")
	http.Error(w, "", 401)
}

func requestCredentials(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", "Basic realm=\"SQL Credentials\"")
	http.Error(w, "", 401)
}
