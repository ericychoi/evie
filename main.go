package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gopkg.in/fsnotify.v0"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				watchFor := fsnotify.Chmod | fsnotify.Rename
				if event.Op&watchFor == watchFor {
					log.Println("modified file:", event.Name)

					filename := strings.SplitN(event.Name, ":", 1)
					process(filename)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/Users/ericchoi/tmp/watch")
	if err != nil {
		log.Fatal(err)
	}

	// block. wait for sigterm or sigint
	wait(syscall.SIGTERM, syscall.SIGINT)
}

func wait(signals ...os.Signal) error {
	ch := make(chan os.Signal)
	signal.Notify(ch, signals...)

	<-ch
	return nil
}

func process(filename string) {
	log.Printf("got %f\n", filename)

}
