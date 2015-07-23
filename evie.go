package main

import (
	"flag"
	"log"
	"syscall"

	"github.com/go-fsnotify/fsnotify"
)

var (
	destDir     *string
	incomingDir *string
	copyOnly    *bool
	serverUrl   *string

	extensions = []string{"avi", "mp4", "mkv", "m4v"}
)

func main() {
	destDir = flag.String("dest", "", "destination dir path")
	incomingDir = flag.String("incoming", "", "incoming dir path")
	serverUrl = flag.String("server", "", "evie server path")
	copyOnly = flag.Bool("copy", false, "whether app is to copy only (no delete)")

	flag.Parse()

	if *destDir == "" {
		log.Fatalln("dest option required")
	}
	if *incomingDir == "" {
		log.Fatalln("incoming option required")
	}

	log.Printf("watching %s..\n", *incomingDir)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	ktvo := &KTVOrganizer{dst: *destDir, src: *incomingDir, copyOnly: *copyOnly, serverUrl: *serverUrl}
	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("fsnotify event:", event)
				if event.Op&fsnotify.Chmod == fsnotify.Chmod &&
					event.Op&syscall.IN_CLOSE_WRITE == syscall.IN_CLOSE_WRITE {
					log.Println("writing to file:", event.Name)
				}
				if event.Op&fsnotify.Rename == fsnotify.Write {
					log.Println("renaming file:", event.Name)
					log.Println("calling ktvo")
					err := ktvo.Do(event.Name)
					if err != nil {
						log.Println("error from ktvo: " + err.Error())
					}
				}
			case err := <-watcher.Errors:
				log.Println("fsnotify error:", err)
			}
		}
	}()

	err = watcher.Add(*incomingDir)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}
