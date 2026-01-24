package core

import (
	"sync"
	"time"
)

type LogEntry struct {
	Time      time.Time `json:"time"`
	Direction string    `json:"direction"`
	Message   string    `json:"message"`
}

type LogBuffer struct {
	mu      sync.Mutex
	entries []LogEntry
	max     int
}

func NewLogBuffer(max int) *LogBuffer {
	return &LogBuffer{
		entries: make([]LogEntry, 0, max),
		max:     max,
	}
}

func (lb *LogBuffer) Add(entry LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.max <= 0 {
		return
	}
	if len(lb.entries) >= lb.max {
		copy(lb.entries, lb.entries[1:])
		lb.entries[len(lb.entries)-1] = entry
		return
	}
	lb.entries = append(lb.entries, entry)
}

func (lb *LogBuffer) Snapshot() []LogEntry {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	out := make([]LogEntry, len(lb.entries))
	copy(out, lb.entries)
	return out
}

func (lb *LogBuffer) Max() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.max
}
