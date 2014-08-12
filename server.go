package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func startServer(port int, path string) error {
	http.HandleFunc(path, handler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	result := EvieResult{
		Show:   "TestShow",
		Season: "TestShowSeason",
		File:   "TestShowFilename",
	}

	json, err := json.Marshal(result)
	if err != nil {
		log.Printf("couldn't json.Marshal: %v\n", result)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
