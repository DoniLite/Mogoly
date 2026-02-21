package core

import (
	"path/filepath"
	"time"

	"github.com/DoniLite/Mogoly/core/events"
	"github.com/DoniLite/Mogoly/core/router"
	"github.com/fsnotify/fsnotify"
)

func WatchConfig(path string, onReload func(*router.Config)) error {
	events.Logf(events.LOG_INFO, "[WATCHER]: Initializing watcher for config path: %s", path)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[WATCHER]: Error while initializing the watcher: %v", err.Error())
		return err
	}
	// Watch the directory; filter for the target filename.
	dir, file := filepath.Split(path)
	if dir == "" {
		dir = "."
	}
	if err := w.Add(dir); err != nil {
		events.Logf(events.LOG_ERROR, "[WATCHER]: Error while watching the dir: %s \nerror: %s", dir, err.Error())
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
				for n := range 5 {
					events.Logf(events.LOG_INFO, "[WATCHER]: Trying to load config file: %s tick: %d", path, n)
					if content, err = LoadConfigFile(path); err == nil {
						events.Logf(events.LOG_INFO, "[WATCHER]: File loaded successfully for the path: %s in tick: %d", path, n)
						break
					}
					time.Sleep(80 * time.Millisecond)
				}
				if err != nil {
					events.Logf(events.LOG_ERROR, "[WATCHER]: error loading config file: %v", err)
					continue
				}
				format, err := DiscoverConfigFormat(path)
				if err != nil {
					events.Logf(events.LOG_ERROR, "[WATCHER]: error discovering config format: %v", err)
					continue
				}
				cfg, err := ParseConfig(content, format)
				if err != nil {
					events.Logf(events.LOG_ERROR, "[WATCHER]: config reload error: %v", err)
					continue
				}
				onReload(cfg)
			case err := <-w.Errors:
				events.Logf(events.LOG_ERROR, "[WATCHER]: watch error: %v", err)
			}
		}
	}()
	return nil
}
