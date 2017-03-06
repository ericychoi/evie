package main

import (
	"encoding/json"
	"errors"
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
	fullUrl := k.serverUrl + fmt.Sprintf("?f=%s", url.QueryEscape(filename))
	res, err := http.Get(fullUrl)
	if err != nil {
		log.Printf("couldn't get from %s: err: %s\n", fullUrl, err.Error())
		return "", "", "", err
	}

	if res.StatusCode != 200 {
		log.Printf("status code not 200 from %s, statusCode: %d\n", fullUrl, res.StatusCode)
		return "", "", "", errors.New("error from Evie")
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("couldn't read from res.Body(): err: %s", err.Error())
		return "", "", "", err
	}
	fmt.Printf("data: %s\n", data)

	var result EvieResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		fmt.Printf("json unmarshall data: %s error: %s\n", data, err.Error())
	}
	fmt.Printf("json: %+v\n", result)

	return result.Show, result.Season, result.File, nil
}

func moveFile(in, out string) error {
	err := copyFile(in, out)
	if err != nil {
		log.Printf("error from moveFile: copyFile returns %s", err.Error())
	} else {
		err := os.Remove(in)
		if err != nil {
			log.Printf("error from os.Remove() while attempting to remove %s. err: %s\n", in, err.Error())
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
		copyCmd = exec.Command("cp", `-p`, in, out)
	}

	return copyCmd.Run()
}
