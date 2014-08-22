package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	destDir     *string
	incomingDir *string
	serverHost  *string
	serverPort  *int
	serverPath  *string
	copyOnly    *bool

	extensions = []string{"avi", "mp4", "mkv", "m4v"}
	seen       map[string]bool
	serverUrl  string
)

type EvieResult struct {
	Show   string
	Season string
	File   string
}

func main() {
	serverHost = flag.String("host", "localhost", "host for Evie server")
	serverPort = flag.Int("port", 55555, "port for Evie server")
	serverPath = flag.String("path", "/json/info/", "path for Evie server")
	destDir = flag.String("dest", "", "destination dir path")
	incomingDir = flag.String("incoming", "", "incoming dir path")
	copyOnly = flag.Bool("copy-only", false, "whether app is to copy only (no delete)")
	isServer = flag.Bool("server", false, "whether app is to start as an server or not")
	flag.Parse()
	if !*isServer && *destDir == "" {
		log.Fatalln("dest option required")
	}
	if !*isServer && *incomingDir == "" {
		log.Fatalln("incoming option required")
	}

	serverUrl := fmt.Sprintf("http://%s:%d%s", *serverHost, *serverPort, *serverPath)

	if *isServer {
		log.Printf("starting server on %s..\n", serverUrl)
		err := startServer(*serverPort, *serverPath)
		if err != nil {
			log.Fatalf("couldn't startServer. err: %s", err)
		}
	} else {
		log.Printf("starting client..\n")
		seen = make(map[string]bool)

		go func(serverUrl string) {
			ticker := time.Tick(time.Millisecond * 1000)
			for {
				select {
				case <-ticker:
					filename := detectNewFile(*incomingDir)
					fullPath := fmt.Sprintf("%s/%s", *incomingDir, filename)
					if filename != "" && !isOpen(fullPath) {
						//TODO check if the file is being written to
						log.Printf("new file: %s\n", fullPath)
						err := process(filename, serverUrl)
						if err != nil {
							log.Println("error from process():", err)
						} else {
							log.Printf("successfully processed %s\n", filename)
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

func isOpen(file string) bool {
	//TODO test this in windows
	f, err := os.OpenFile(file, os.O_RDWR|os.O_SYNC, 0755)
	defer f.Close()
	if err == nil {
		return false
	}
	log.Printf("can't open %s for writing. err %s. this is good.\n", file, err)
	return true
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

	dirPath := fmt.Sprintf("%s/%s/%s", *destDir, show, season)
	inFullname := fmt.Sprintf("%s/%s", *incomingDir, filename)
	outFullname := fmt.Sprintf("%s/%s", dirPath, newFile)

	log.Printf("dirPath: %s\t inFullname: %s\t outFullname: %s\n", dirPath, inFullname, outFullname)

	err = os.MkdirAll(dirPath, os.FileMode(0755))
	if err != nil {
		log.Println("error from MkdirAll(): ", err)
		return err
	}

	if *copyOnly {
		return copyFile(inFullname, outFullname)
	}
	return moveFile(inFullname, outFullname)
}

func doEvie(filename, serverUrl string) (string, string, string, error) {
	res, err := http.Get(serverUrl + fmt.Sprintf("/%s", urlEncode(filename)))
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

func urlEncode(in string) string {
	/*
		if utf8.RuneCountInString(in) > 0 {

					buf := make([]byte, 100)
					n := utf8.EncodeRune(buf, in)
			if n <= 0 {
				log.Printf("couldn't encode %s into utf-8\n", in)
				return ""
			}
			return url.QueryEscape(string(buf))
		}
	*/
	return url.QueryEscape(in)
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
