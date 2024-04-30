package main

import (
	"fmt"
	"folderwatch/config"
	"folderwatch/dbsqlite"
	"folderwatch/fwlog"
	"folderwatch/scanner"
	"log"
	"os"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

func main() {

	//1. Load Config
	cfg := config.LoadConfig()
	if cfg == nil {
		os.Exit(1)
	}

	//2. Create log file
	lfl := fwlog.CreateLogFile(cfg)

	//3. Open database. create fsevents table if necessary, remove old records.
	dbs := dbsqlite.OpenDatabase(cfg.DatabaseLocation, lfl)
	if dbs == nil {
		os.Exit(1)
	}

	//4. create event channel
	chn := make(chan watcher.Event)

	//5. create scanner. One scanner can handle multiple event handling goroutines
	scn := scanner.CreateScanner(cfg, chn, dbs, lfl)

	//6. add initial records for files in target directory
	err := scn.AddInitialRecords()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//7. setup watcher
	w := watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	//w.SetMaxEvents(100)

	// Only notify rename and move events.
	//	w.FilterOps(watcher.Rename, watcher.Move, watcher.)

	// Only files that match the regular expression during file listings
	// will be watched.
	//r := regexp.MustCompile(".*")
	//w.AddFilterHook(watcher.RegexFilterHook(r, false))

	var wg sync.WaitGroup

	//this is the main event handling loop
	//TODO: put this code in a loop where we create cfg.NumThreads number of goroutines to handle the events.
	wg.Add(1)
	go func() {
		defer wg.Done()
		scn.HandleFileSystemChange()
	}()

	wg.Add(1)
	//this loop captures events from watcher, filters them and pass on to event handling goroutine
	go func() {
		defer wg.Done()
		for {
			select {
			case event := <-w.Event:
				if !event.IsDir() { //ignoring dir events. a file change results in a DIRECTORY event AND a FILE event. so we ignore DIRECTORY events.
					//fmt.Println(event) // Print the event's info.
					chn <- event
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch test_folder recursively for changes.
	fmt.Println("Target Dir", cfg.TargetDir)
	if err := w.AddRecursive(cfg.TargetDir); err != nil {
		log.Fatalln(err)
	}

	fmt.Println()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {

		log.Fatalln(err)
	}
	wg.Wait()
}
