package testagent

import (
	"sync"
	"time"
)

// Status is the agent run state.
type Status string

const (
	StatusStopped Status = "stopped"
	StatusRunning Status = "running"
	StatusStalled Status = "stalled"
)

// Phase is a journal entry phase.
type Phase string

const (
	PhaseObserve Phase = "observe"
	PhasePlan    Phase = "plan"
	PhaseAct     Phase = "act"
	PhaseBoot    Phase = "boot"
)

// JournalEntry records one agent step.
type JournalEntry struct {
	Time      time.Time `json:"time"`
	Phase     Phase     `json:"phase"`
	URL       string    `json:"url,omitempty"`
	Action    string    `json:"action,omitempty"`
	Status    int       `json:"status,omitempty"`
	Error     string    `json:"error,omitempty"`
	Reasoning string    `json:"reasoning,omitempty"`
}

// Journal is a thread-safe append-only log.
type Journal struct {
	mu      sync.RWMutex
	entries []JournalEntry
}

// NewJournal creates an empty journal.
func NewJournal() *Journal {
	return &Journal{}
}

// Append adds an entry and returns a copy for callers.
func (j *Journal) Append(e JournalEntry) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	j.mu.Lock()
	j.entries = append(j.entries, e)
	j.mu.Unlock()
}

// Entries returns a copy of all entries (newest last).
func (j *Journal) Entries() []JournalEntry {
	j.mu.RLock()
	defer j.mu.RUnlock()
	out := make([]JournalEntry, len(j.entries))
	copy(out, j.entries)
	return out
}

// EntriesNewestFirst returns entries reversed.
func (j *Journal) EntriesNewestFirst() []JournalEntry {
	all := j.Entries()
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	return all
}

// Last returns the most recent entry or nil.
func (j *Journal) Last() *JournalEntry {
	j.mu.RLock()
	defer j.mu.RUnlock()
	if len(j.entries) == 0 {
		return nil
	}
	e := j.entries[len(j.entries)-1]
	return &e
}

// ErrorCount counts entries with non-empty Error.
func (j *Journal) ErrorCount() int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	n := 0
	for _, e := range j.entries {
		if e.Error != "" {
			n++
		}
	}
	return n
}
