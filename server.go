package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func startServer(port int, path string) error {
	http.HandleFunc(path, handler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

/*2014/08/22 00:54:37 path: /json/info/New+Text+Document.avi
2014/08/22 00:54:38 path: /json/info/New_Text_Document_2.avi
2014/08/22 00:56:38 path: /json/info/New+Text+Document.avi
2014/08/22 00:56:39 path: /json/info/New_Text_Document_2.avi
2014/08/22 00:56:40 path: /json/info/%EB%B6%80%EB%A6%89%21+%EB%B6%80%EB%A6%89%21+%EB%B8%8C%EB%A3%A8%EB%AF%B8%EC%A6%88+S02E04+%EB%B8%8C%EB%A3%A8%EB%AF%B8%EC%A6%88+%EC%84%9C%EC%BB%A4%EC%8A%A4%EB%8B%A8+120306+HDTV+x264+720p-Ernie.mp4
*/
func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("path: %s\n", r.RequestURI)
	fileName := "TestShowFilename_" + fmt.Sprintf("%d", time.Now().Unix())
	result := EvieResult{
		Show:   "TestShow",
		Season: "TestShowSeason",
		File:   fileName,
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
