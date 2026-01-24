package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogBufferEvictsOldest(t *testing.T) {
	buf := NewLogBuffer(2)
	buf.Add(LogEntry{Time: time.Now(), Direction: "tx", Message: "first"})
	buf.Add(LogEntry{Time: time.Now(), Direction: "tx", Message: "second"})
	buf.Add(LogEntry{Time: time.Now(), Direction: "tx", Message: "third"})

	snap := buf.Snapshot()
	require.Len(t, snap, 2)
	require.Equal(t, "second", snap[0].Message)
	require.Equal(t, "third", snap[1].Message)
}

func TestLogBufferMax(t *testing.T) {
	buf := NewLogBuffer(5)
	require.Equal(t, 5, buf.Max())
}
