// import "util/directorywatcher"
//
// TODO Move directorywatcher to its own github repo (along with other util functions?)
package directorywatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// Type of observer function - adding an observer means adding a function of this type
type Observer chan EventsAt

// The directory watcher struct - note that the struct is not exported
// (disallowing manual construct), but certain fields are (so we can set them
// after creation).
type directoryWatcher struct {
	Interval  uint64 // interval in ms
	Recursive bool   // Use filepath.Walk or filepath.Glob?
	Pattern   string // glob pattern

	// Internal details
	path      string                 // the path being watched
	files     map[string]os.FileInfo // Map of files watched
	ticker    *time.Ticker           // The interval timer - if the ticker is != nil, then we assume that it's started
	observers []Observer             // List of observers

	// Extra features
	Preload bool
}

/**
 * Usage:
 * 
 * import DW "util/directorywatcher"
 *
 * func main() {
 * 	dw := DW.New(".")
 *	c := dw.AddNewObserver()
 * 	dw.Start()
 *	for {
 *		select {
 *		case args := <-c:
 *			fmt.Printf("%d files changed at %s!\n", len(args.Events), args.At)
 *		}
 * 	}
 * }
 *
 */
func New(path string) (*directoryWatcher, error) {
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("Provided path is not a directory: %s", path)
	}

	return &directoryWatcher{
		Interval:  2000,
		Pattern:   "*",
		observers: []Observer{},
		path:      path,
		files:     make(map[string]os.FileInfo),
	}, nil
}

func NewOpts(path string, opts map[string]interface{}) (*directoryWatcher, error) {
	dw, err := New(path)
	if err != nil {
		return nil, err
	}

	// Reflect - iterates over (exported) fields of directoryWatcher and
	// sets those where a value has been provided in opts (if they are the
	// same kind.
	dwValue := reflect.ValueOf(dw).Elem()
	dwTyp := dwValue.Type()
	for i := 0; i < dwValue.NumField(); i++ {
		if !dwValue.Field(i).CanSet() {
			continue
		}
		if v, ok := opts[dwTyp.Field(i).Name]; ok {
			field := dwValue.Field(i)
			val := reflect.ValueOf(v)
			if field.Kind() == val.Kind() {
				field.Set(val)
				delete(opts, dwTyp.Field(i).Name)
			}
		}
	}

	// Warn about unused keys
	for k, v := range opts {
		fmt.Printf("Unused option: %s (%v)\n", k, v)
	}

	return dw, nil
}

// The watcher runs in a goroutine, sending notifications back over to the
// attached observers (channels). Notifications are only sent if any files have
// actually changed.
func (dw *directoryWatcher) Start() {
	if dw.ticker != nil {
		return
	}
	go func() {
		now := time.Now()
		if fst := dw.scan1(); !dw.Preload {
			dw.notify(EventsAt{now, fst})
		}
		dw.ticker = time.NewTicker(time.Duration(dw.Interval) * time.Millisecond)
		for now = range dw.ticker.C {
			dw.notify(EventsAt{now, dw.scan1()})
		}
	}()
}

func (dw *directoryWatcher) Stop() {
	dw.ticker.Stop()
	dw.ticker = nil
}

// We use the ticker to decide whether or not we're running.
func (dw *directoryWatcher) Running() bool {
	return dw.ticker != nil
}

func NewObserver() Observer {
	return make(Observer)
}

func (dw *directoryWatcher) AddNewObserver() Observer {
	o := make(Observer)
	dw.observers = append(dw.observers, o)
	return o
}

func (dw *directoryWatcher) AddObserver(obs Observer) {
	dw.observers = append(dw.observers, obs)
}

// Only sends notification if the number of events is greater than zero
func (dw *directoryWatcher) notify(evAt EventsAt) {
	if len(evAt.Events) == 0 {
		return
	}
	for _, ch := range dw.observers {
		ch <- evAt
	}
}

// The actual walking function: Scans and returns a list of events on all the
// files that somehow changed (added, changed or deleted).
func (dw *directoryWatcher) scan1() (changed []Event) {
	touched := make(map[string]bool) // path names of the files seen in a pass
	if dw.Recursive {
		filepath.Walk(dw.path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || !matches(dw.Pattern, info.Name()) {
				return nil
			}
			if ev, ok := dw.hasChange(path, info); ok {
				dw.files[path] = info
				changed = append(changed, ev)
			}
			touched[path] = true
			return err
		})
	} else {
		matches, _ := filepath.Glob(filepath.Join(dw.path, dw.Pattern))
		for _, path := range matches {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}
			if ev, ok := dw.hasChange(path, info); ok {
				dw.files[path] = info
				changed = append(changed, ev)
			}
			touched[path] = true
		}
	}
	for path, info := range dw.files {
		if !touched[path] {
			changed = append(changed, Event{Deleted, path, info})
			delete(dw.files, path)
		}
	}
	return
}

func matches(pattern, name string) bool {
	matched, err := filepath.Match(pattern, name)
	return err == nil && matched
}

/**
 * This tells us if a given file has been changed or added.
 *
 * Uses the comma-ok style to indicate whether or not a given file actually changed.
 */
func (dw *directoryWatcher) hasChange(path string, info os.FileInfo) (Event, bool) {
	if oldInfo, ok := dw.files[path]; ok {
		return Event{Changed, path, info}, info.ModTime().After(oldInfo.ModTime())
	}
	return Event{Added, path, info}, true
}
