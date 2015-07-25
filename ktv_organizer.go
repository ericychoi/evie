package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
)

type KTVOrganizer struct {
	dst       string
	src       string
	copyOnly  bool
	serverUrl string
}

type EvieResult struct {
	Show   string
	Season string
	File   string
}

// expects full filepath
func (k *KTVOrganizer) Do(filename string) error {
	log.Printf("ktvo: got %s\n", filename)

	show, season, newFile, err := k.getShowInfo(filename)
	if err != nil {
		log.Printf("error from getShowInfo(): %s\n", err)
		return err
	}

	dirPath := fmt.Sprintf("%s/%s/%s", k.dst, show, season)
	outFullname := fmt.Sprintf("%s/%s", dirPath, newFile)
	inFullname := fmt.Sprintf("%s/%s", k.src, filename)

	log.Printf("dirPath: %s\t inFullname: %s\t outFullname: %s\n", dirPath, inFullname, outFullname)

	err = os.MkdirAll(dirPath, os.FileMode(0755))
	if err != nil {
		log.Println("error from MkdirAll(): ", err)
		return err
	}

	if k.copyOnly {
		return copyFile(inFullname, outFullname)
	}

	return moveFile(inFullname, outFullname)
}

func (k *KTVOrganizer) getShowInfo(filename string) (string, string, string, error) {
	res, err := http.Get(k.serverUrl + fmt.Sprintf("/%s", url.QueryEscape(filename)))
	if err != nil {
		log.Printf("couldn't get from %s: err: %s", serverUrl, err)
		return "", "", "", err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("couldn't read from res.Body(): err: %s", err)
		return "", "", "", err
	}
	fmt.Printf("data: %s\n", data)

	var result EvieResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		fmt.Printf("json unmarshall data: %s error: %s\n", data, err)
	}
	fmt.Printf("json: %+v\n", result)

	return result.Show, result.Season, result.File, nil
}

func moveFile(in, out string) error {
	err := copyFile(in, out)
	if err != nil {
		log.Printf("error from moveFile: copyFile returns %s", err)
	} else {
		err := os.Remove(in)
		if err != nil {
			log.Printf("error from os.Remove() while attempting to remove %s. err: %s\n", in, err)
		}
	}
	return err
}

func copyFile(in, out string) error {
	log.Printf("in copyFile in: %s out: %s\n", in, out)

	var copyCmd *exec.Cmd

	if runtime.GOOS == "windows" {
		//TODO make the directory path separator \ for windows, also wrap the path in quotes for windows
		copyCmd = exec.Command("cmd", `/C`, "copy", in, out)
	} else {
		copyCmd = exec.Command("cp", in, out)
	}

	return copyCmd.Run()
}
