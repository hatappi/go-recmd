// Package event represent change event
package event

import (
	"time"
)

// Event represents event for recmd
type Event struct {
	Path      string
	CreatedAt time.Time
}
