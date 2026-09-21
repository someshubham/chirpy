package main

import (
	"encoding/json"
	"net/http"
)

type errorRes struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, msg string) {
	errResp := errorRes{
		Error: msg,
	}
	dat, err := json.Marshal(errResp)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(400)
	w.Write(dat)
}
