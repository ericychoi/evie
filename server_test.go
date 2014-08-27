package main

import (
	"testing"
)

func TestParseFilename(t *testing.T) {
	uri := `/json/info/%EB%B6%80%EB%A6%89%21+%EB%B6%80%EB%A6%89%21+%EB%B8%8C%EB%A3%A8%EB%AF%B8%EC%A6%88+S02E04+%EB%B8%8C%EB%A3%A8%EB%AF%B8%EC%A6%88+%EC%84%9C%EC%BB%A4%EC%8A%A4%EB%8B%A8+120306+HDTV+x264+720p-Ernie.mp4`
	actual, err := parseFilename(uri)
	if err != nil {
		t.Errorf("parseFilname() returned error with input %s", uri)
	}
	expected := `부릉! 부릉! 브루미즈 S02E04 브루미즈 서커스단 120306 HDTV x264 720p-Ernie.mp4`

	if actual != expected {
		t.Errorf("got %s, expected %s", actual, expected)
	}
}
