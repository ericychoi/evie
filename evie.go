package main

import (
	"flag"
	"log"

	"github.com/go-fsnotify/fsnotify"
)

var (
	destDir     *string
	incomingDir *string
	copyOnly    *bool
	serverUrl   *string

	extensions = []string{"avi", "mp4", "mkv", "m4v"}
)

// usage: ./evie --dest ./dest --incoming ./watch --server evie.rookie1.co:3000/match --copy

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
				/* transmission renaming the torrent after it's done will follow this pattern
				2015/07/23 23:34:41 fsnotify event: "watch/ubuntu-14.04.2-server-amd64.iso.part": RENAME, op: 1000
				2015/07/23 23:34:41 fsnotify event: "watch/ubuntu-14.04.2-server-amd64.iso": CREATE, op: 1 */
				//log.Printf("fsnotify event: %s op: %b\n", event, event.Op)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Printf("file created: %s, calling ktvo\n", event.Name)
					if isValidExt(event.Name) {
						err := ktvo.Do(event.Name)
						if err != nil {
							log.Println("error from ktvo: " + err.Error())
						}
					}
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

func isValidExt(fullName string) bool {
	regex := regexp.MustCompile(strings.Join(extensions, "|"))
	if regex.MatchString(filepath.Ext(fullName)) {
		return true
	}
	return false
}
