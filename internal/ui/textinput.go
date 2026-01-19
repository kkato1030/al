package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TextInputModel represents a UI model for text input
type TextInputModel struct {
	prompt    string
	value     string
	required  bool
	quitting  bool
	submitted bool
}

// Init initializes the model
func (m *TextInputModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if !m.required || m.value != "" {
				m.submitted = true
				m.quitting = true
				return m, tea.Quit
			}
		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
		default:
			// Add character to value
			if len(msg.Runes) > 0 {
				m.value += string(msg.Runes)
			}
		}
	}
	return m, nil
}

// View renders the UI
func (m *TextInputModel) View() string {
	if m.quitting {
		// Show final input
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%s: ", m.prompt))
		if m.value == "" {
			b.WriteString("(none)\n")
		} else {
			b.WriteString(m.value)
			b.WriteString("\n")
		}
		return b.String()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s", m.prompt))
	if !m.required {
		b.WriteString(" (optional, press Enter to skip)")
	}
	b.WriteString(": ")
	b.WriteString(m.value)
	b.WriteString("_\n\n")
	b.WriteString("  Type to enter text  Enter: Confirm  q: Quit\n")

	return b.String()
}

// GetValue returns the input value
func (m *TextInputModel) GetValue() string {
	return m.value
}

// NewTextInputModel creates a new text input model
func NewTextInputModel(prompt string, required bool) *TextInputModel {
	return &TextInputModel{
		prompt:   prompt,
		value:    "",
		required: required,
	}
}
