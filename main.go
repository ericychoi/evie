package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	isServer    *bool
	destDir     = "/Users/ericchoi/tmp/dest"
	incomingDir = "/Users/ericchoi/tmp/watch"
	serverHost  = "localhost"
	serverPort  = 55555
	serverPath  = "/json/info"
	copyOnly    = true
	extensions  = []string{"avi", "mp4", "mkv"}
	seen        map[string]bool
	copyCmd     string
	serverUrl   string
)

type EvieResult struct {
	Show   string
	Season string
	File   string
}

func main() {
	serverUrl := fmt.Sprintf("http://%s:%d%s", serverHost, serverPort, serverPath)
	isServer = flag.Bool("server", false, "whether app is start as an server or not")
	flag.Parse()
	if *isServer {
		log.Printf("starting server on %s..\n", serverUrl)
		err := startServer(serverPort, serverPath)
		if err != nil {
			log.Fatalf("couldn't startServer. err: %s", err)
		}
	} else {
		if runtime.GOOS == "windows" {
			copyCmd = "copy"
		} else {
			copyCmd = "cp"
		}

		log.Printf("starting client with copy command: %s..\n", copyCmd)
		seen = make(map[string]bool)

		go func(serverUrl string) {
			ticker := time.Tick(time.Millisecond * 1000)
			for {
				select {
				case <-ticker:
					filename := detectNewFile(incomingDir)
					if filename != "" {
						//TODO check if the file is being written to
						log.Println("new file:", filename)
						err := process(filename, serverUrl)
						if err != nil {
							log.Println("error from process():", err)
						}
					}
				}

			}
		}(serverUrl)
	}

	// block. wait for sigterm or sigint
	wait(syscall.SIGTERM, syscall.SIGINT)
}

func detectNewFile(dir string) string {
	files, _ := ioutil.ReadDir(fmt.Sprintf("%s/", dir))
	for _, f := range files {
		if !f.IsDir() && isValidExt(f.Name()) && !seen[f.Name()] {
			seen[f.Name()] = true
			return f.Name()
		}
	}
	return ""
}

func isValidExt(fullName string) bool {
	regex := regexp.MustCompile(strings.Join(extensions, "|"))
	if regex.MatchString(filepath.Ext(fullName)) {
		return true
	}
	return false
}

func wait(signals ...os.Signal) error {
	ch := make(chan os.Signal)
	signal.Notify(ch, signals...)

	<-ch
	return nil
}

func process(filename, serverUrl string) error {
	log.Printf("got %s\n", filename)

	//TODO we will get these from API based on filename

	log.Printf("serverUrl: %s\n", serverUrl)

	show, season, newFile, err := doEvie(filename, serverUrl)
	if err != nil {
		log.Printf("error from doEvie(): %s\n", err)
		return err
	}

	dirPath := fmt.Sprintf("%s/%s/%s", destDir, show, season)
	inFullname := fmt.Sprintf("%s/%s", incomingDir, filename)
	outFullname := fmt.Sprintf("%s/%s", dirPath, newFile)

	log.Printf("dirPath: %s\t inFullname: %s\t outFullname: %s\n", dirPath, inFullname, outFullname)

	err = os.MkdirAll(dirPath, os.FileMode(0755))
	if err != nil {
		log.Println("error from MkdirAll(): ", err)
		return err
	}

	if copyOnly {
		return copyFile(inFullname, outFullname)
	}
	return moveFile(inFullname, outFullname)
}

func doEvie(filename, serverUrl string) (string, string, string, error) {
	res, err := http.Get(serverUrl)
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
	cpCmd := exec.Command(copyCmd, in, out)
	return cpCmd.Run()
}
