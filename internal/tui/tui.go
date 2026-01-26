package tui

import (
	"context"
	"fmt"

	"gomodmaster/internal/config"
	"gomodmaster/internal/core"

	tea "github.com/charmbracelet/bubbletea"
)

const logBufferSize = 500

// Run starts the TUI and blocks until it exits.
func Run(cfg config.Config) error {
	service := core.NewServiceWithLogSize(cfg, logBufferSize)

	state := newModel(cfg, service)
	program := tea.NewProgram(state, tea.WithAltScreen())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-service.Events():
				program.Send(eventMsg{event: event})
			}
		}
	}()

	final, err := program.Run()
	_ = service.Disconnect()
	if err != nil {
		return err
	}
	if m, ok := final.(model); ok && m.printInvocation {
		fmt.Println(m.cfg.InvocationTUI())
	}
	return nil
}
