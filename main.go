package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var (
	destDir     = "/Users/ericchoi/tmp/dest"
	incomingDir = "/Users/ericchoi/tmp/watch"
	copyOnly    = true
	extensions  = []string{"avi", "mp4", "mkv"}
	seen        map[string]bool
)

func main() {
	log.Println("starting...")
	seen = make(map[string]bool)

	go func() {
		ticker := time.Tick(time.Millisecond * 1000)
		for {
			select {
			case <-ticker:
				filename := detectNewFile(incomingDir)
				if filename != "" {
					log.Println("new file:", filename)
					err := process(filename)
					if err != nil {
						log.Println("error from process():", err)
					}
				}
			}

		}
	}()

	// block. wait for sigterm or sigint
	wait(syscall.SIGTERM, syscall.SIGINT)
}

func detectNewFile(dir string) string {
	files, _ := ioutil.ReadDir(fmt.Sprintf("%s/", dir))
	for _, f := range files {
		if !f.IsDir() && isValidExt(f.Name()) && !seen[f.Name()] {
			fmt.Println(f.Name())
			seen[f.Name()] = true
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

func process(filename string) error {
	log.Printf("got %f\n", filename)

	//TODO we will get these from API based on filename
	show := "infinity challenge"
	season := "2014"
	newFile := "infinity challenge - 2014-08-09 - something special.avi"

	dirPath := fmt.Sprintf("%s/%s/%s", destDir, show, season)
	inFullname := fmt.Sprintf("%s/%s", incomingDir, filename)
	outFullname := fmt.Sprintf("%s/%s", dirPath, newFile)

	log.Printf("dirPath: %s\t inFullname: %s\t outFullname: %s\n", dirPath, inFullname, outFullname)

	err := os.MkdirAll(dirPath, os.FileMode(0755))
	if err != nil {
		log.Println("error from MkdirAll(): ", err)
		return err
	}

	if copyOnly {
		return copyFile(inFullname, outFullname)
	}
	return moveFile(inFullname, outFullname)
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
	// open files r and w
	r, err := os.Open(in)
	if err != nil {
		log.Printf("error from CopyFile: opening %s, err: %s\n", in, err)
		return err
	}
	defer r.Close()

	w, err := os.Create(out)
	if err != nil {
		log.Println("error from CopyFile: creating %s, err: %s\n", out, err)
		return err
	}
	defer w.Close()

	// do the actual work
	n, err := io.Copy(w, r)
	if err != nil {
		log.Println("error from CopyFile: copying, err: %s\n", err)
		return err
	}

	log.Println("Successfully copied %s to %s. %v bytes\n", in, out, n)
	return nil
}
