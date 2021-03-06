// The directory package contains the main cloned machine detecting
// logic of the server.
//
// It works by keeping track of the last token presented by each machine.
// On each request, the client application running on a machine will
// present the previously valid token and a freshly generated one.
// If the token matches the one recorded by the server, the record is updated.
// If the token does not match, the machine is treated as a clone - a
// new record is created for it, increasing the unique machine count.
//
// Since machine records expire, inactive machines do not effect the
// count.
package directory

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Maximum age at which a machine is no longer considered alive.
const maxAge = time.Minute

// Directory tracks unique machines.
type Directory struct {
	m sync.Mutex

	machines map[string]machineEntry
}

// New returns a new Directory instance.
func New() *Directory {
	return &Directory{
		machines: make(map[string]machineEntry),
	}
}

type machineEntry struct {
	id       string
	token    string
	lastSeen time.Time
}

// UpdateMachine attempts to update the token associated with a machine.
// The method returns the id of the machine. In case a duplicate machine is detected,
// a new id is returned.
func (d *Directory) UpdateMachine(id, oldToken, newToken string) string {
	d.m.Lock()
	defer d.m.Unlock()

	m, ok := d.machines[id]
	// New machine.
	if !ok {
		d.machines[id] = machineEntry{
			id:       id,
			token:    newToken,
			lastSeen: time.Now(),
		}
		return id
	}
	// Clone detected.
	if m.token != oldToken {
		newId := uuid.New().String()
		d.machines[newId] = machineEntry{
			id:       newId,
			token:    newToken,
			lastSeen: time.Now(),
		}
		return newId
	}
	// Existing machine.
	d.machines[id] = machineEntry{
		id:       id,
		token:    newToken,
		lastSeen: time.Now(),
	}
	return id
}

// Count returns the number of unique machines.
func (d *Directory) Count() int {
	d.m.Lock()
	defer d.m.Unlock()
	var cnt int
	for _, m := range d.machines {
		if time.Since(m.lastSeen) <= maxAge {
			cnt++
		}
	}
	return cnt
}
