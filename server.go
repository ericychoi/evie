package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func startServer(port int, path string) error {
	http.HandleFunc(path, handler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("path: %s\n", r.RequestURI)
	filename, err := parseFilename(r.RequestURI)
	if err != nil {
		log.Printf("couldn't get filename from uri: %s", r.RequestURI)
	}

	//TODO
	// save the extension
	// fragment the filename proper, and search against the known keywords to find the show name
	// for now, let's just deal with date-based shows
	// get the date, which is probably YYMMDD, and convert it to the right format YYYY-MM-DD
	// season name will be just YYYY (year)

	filename = "TestShowFilename_" + fmt.Sprintf("%d", time.Now().Unix())
	result := EvieResult{
		Show:   "TestShow",
		Season: "TestShowSeason",
		File:   filename,
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

func parseFilename(uri string) (string, error) {
	path, err := url.QueryUnescape(uri)
	if err != nil {
		return "", err
	}
	segments := strings.Split(path, `/`)
	return segments[len(segments)-1], nil
}
