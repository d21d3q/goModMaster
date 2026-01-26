package tui

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func listSerialDevices() []string {
	switch runtime.GOOS {
	case "darwin":
		return globDevices([]string{"/dev/tty.*"})
	case "linux":
		return globDevices([]string{"/dev/serial/by-id/*"})
	case "windows":
		return windowsComPorts(32)
	default:
		return []string{}
	}
}

func globDevices(patterns []string) []string {
	devices := []string{}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		devices = append(devices, matches...)
	}
	return devices
}

func windowsComPorts(max int) []string {
	devices := make([]string, 0, max)
	for i := 1; i <= max; i++ {
		devices = append(devices, fmt.Sprintf("COM%d", i))
	}
	return devices
}
