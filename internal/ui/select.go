package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SelectModel represents a UI model for selecting from a list of options
type SelectModel struct {
	prompt    string
	options   []string
	cursor    int
	selected  string
	quitting  bool
	submitted bool
}

// Init initializes the model
func (m *SelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.options[m.cursor]
			m.submitted = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m *SelectModel) View() string {
	if m.quitting {
		// Show final selection
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%s: %s\n", m.prompt, m.selected))
		return b.String()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s\n\n", m.prompt))

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", cursor, option))
	}

	b.WriteString("\n  ↑/k: Up  ↓/j: Down  Enter: Select  q: Quit\n")

	return b.String()
}

// GetSelected returns the selected option
func (m *SelectModel) GetSelected() string {
	return m.selected
}

// NewSelectModel creates a new select model
func NewSelectModel(prompt string, options []string) *SelectModel {
	return &SelectModel{
		prompt:  prompt,
		options: options,
		cursor:  0,
	}
}
