package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputModel struct {
	input textinput.Model
	label string
	hint  string
}

func newInputModel() inputModel {
	in := textinput.New()
	in.Prompt = ""
	return inputModel{input: in}
}

func (m inputModel) Set(label, value, hint string, limit int) inputModel {
	m.label = label
	m.hint = hint
	m.input.CharLimit = limit
	m.input.SetValue(value)
	m.input.CursorEnd()
	return m
}

func (m *inputModel) Focus() tea.Cmd {
	return m.input.Focus()
}

func (m *inputModel) Blur() {
	m.input.Blur()
}

func (m inputModel) Update(msg tea.Msg) (inputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m inputModel) Value() string {
	return m.input.Value()
}

func (m inputModel) View() string {
	lines := []string{m.label, "", m.input.View()}
	if m.hint != "" {
		lines = append(lines, "", m.hint)
	}
	return strings.Join(lines, "\n")
}
