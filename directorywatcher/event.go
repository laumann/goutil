package directorywatcher

import (
	"fmt"
	"os"
	"time"
)

type eventType int

const (
	Added eventType = iota
	Changed
	Deleted
)

// Mapping event types to a string, for implementing Stringer interface
var eventNames = map[eventType]string{
	Added:   "Added",
	Changed: "Changed",
	Deleted: "Deleted",
}

// eventType implements Stringer
func (et eventType) String() string {
	return eventNames[et]
}

// Implement Stringer
func (e Event) String() string {
	return fmt.Sprintf("%s %s", eventNames[e.Type], e.Path)
}

// An event contains its type and the file involved.
type Event struct {
	Type eventType
	Path string
	os.FileInfo
}

// EventsAt contains a list of events (one for each file that changed) and a
// timestamp. 
type EventsAt struct {
	At     time.Time
	Events []Event
}
