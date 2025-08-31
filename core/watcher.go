package core

import (
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchConfig(path string, onReload func(*Config)) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	// Watch the directory; filter for the target filename.
	dir, file := filepath.Split(path)
	if dir == "" {
		dir = "."
	}
	if err := w.Add(dir); err != nil {
		return err
	}
	go func() {
		debounce := time.NewTimer(0)
		if !debounce.Stop() {
			<-debounce.C
		}
		for {
			select {
			case e := <-w.Events:
				if filepath.Base(e.Name) != file {
					break
				}
				if e.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					// debounce 300ms
					if !debounce.Stop() {
						select {
						case <-debounce.C:
						default:
						}
					}
					debounce.Reset(300 * time.Millisecond)
				}
			case <-debounce.C:
				// retry a few times in case we catch the file mid-write/rename
				var content []byte
				var err error
				for range 5 {
					if content, err = LoadConfigFile(path); err == nil {
						break
					}
					time.Sleep(80 * time.Millisecond)
				}
				if err != nil {
					log.Printf("error loading config file: %v", err)
					continue
				}
				format, err := DiscoverConfigFormat(path)
				if err != nil {
					log.Printf("error discovering config format: %v", err)
					continue
				}
				cfg, err := ParseConfig(content, format)
				if err != nil {
					log.Printf("config reload error: %v", err)
					continue
				}
				onReload(cfg)
			case err := <-w.Errors:
				log.Printf("watch error: %v", err)
			}
		}
	}()
	return nil
}
