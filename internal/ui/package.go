package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
)

// ProviderSelectModel represents a UI model for provider selection
type ProviderSelectModel struct {
	items      []config.ProviderConfig
	cursor     int
	title      string
	quitting   bool
	selected   int // -1 means none selected
	defaultIdx int // -1 means no default
}

// Init initializes the model
func (m *ProviderSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *ProviderSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			// Toggle selection
			if m.selected == m.cursor {
				// If deselecting default, keep default selected
				if m.defaultIdx == m.cursor {
					// Keep default selected
				} else {
					m.selected = -1
				}
			} else {
				m.selected = m.cursor
			}
		case "enter":
			// If nothing selected and default exists, use default
			if m.selected == -1 && m.defaultIdx >= 0 {
				m.selected = m.defaultIdx
			}
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m *ProviderSelectModel) View() string {
	if m.quitting {
		// Show final selection
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%s: ", m.title))
		selected := m.GetSelected()
		if selected == "" {
			b.WriteString("(none)\n")
		} else {
			b.WriteString(selected)
			b.WriteString("\n")
		}
		return b.String()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s:\n\n", m.title))

	for i, item := range m.items {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		checkbox := "[ ]"
		if m.selected == i || (m.selected == -1 && m.defaultIdx == i) {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s%s %s", prefix, checkbox, item.Name)
		if i == m.defaultIdx {
			line += " (default)"
		}
		if item.Version != "" {
			line += fmt.Sprintf(" (version: %s)", item.Version)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	if m.defaultIdx >= 0 {
		b.WriteString(fmt.Sprintf("  Default: %s (press Enter to use default)\n", m.items[m.defaultIdx].Name))
	}
	b.WriteString("  ↑/↓: Move  Space: Select/Deselect  Enter: Confirm  q: Quit\n")

	return b.String()
}

// GetSelected returns the selected provider name, or empty string if none
func (m *ProviderSelectModel) GetSelected() string {
	if m.selected >= 0 && m.selected < len(m.items) {
		return m.items[m.selected].Name
	}
	if m.selected == -1 && m.defaultIdx >= 0 {
		return m.items[m.defaultIdx].Name
	}
	return ""
}

// NewProviderSelectModel creates a new provider select model
func NewProviderSelectModel(items []config.ProviderConfig, title string, defaultProvider string) *ProviderSelectModel {
	defaultIdx := -1
	if defaultProvider != "" {
		for i, p := range items {
			if p.Name == defaultProvider {
				defaultIdx = i
				break
			}
		}
	}

	return &ProviderSelectModel{
		items:      items,
		cursor:     0,
		title:      title,
		selected:   -1,
		defaultIdx: defaultIdx,
	}
}

// PackageSelectModel represents a UI model for package selection
type PackageSelectModel struct {
	items    []config.PackageConfig
	cursor   int
	title    string
	quitting bool
	selected int // -1 means none selected
}

// Init initializes the model
func (m *PackageSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *PackageSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			// Toggle selection
			if m.selected == m.cursor {
				m.selected = -1
			} else {
				m.selected = m.cursor
			}
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m *PackageSelectModel) View() string {
	if m.quitting {
		// Show final selection
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%s: ", m.title))
		selected := m.GetSelected()
		if selected == nil {
			b.WriteString("(none)\n")
		} else {
			b.WriteString(fmt.Sprintf("%s (provider: %s, profile: %s)\n", selected.Name, selected.Provider, selected.Profile))
		}
		return b.String()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s:\n\n", m.title))

	for i, item := range m.items {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		checkbox := "[ ]"
		if m.selected == i {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s%s %s (provider: %s, profile: %s", prefix, checkbox, item.Name, item.Provider, item.Profile)
		if item.Version != "" {
			line += fmt.Sprintf(", version: %s", item.Version)
		}
		if item.Description != "" {
			line += fmt.Sprintf(" - %s", item.Description)
		}
		line += ")"
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString("  ↑/↓: Move  Space: Select/Deselect  Enter: Confirm  q: Quit\n")

	return b.String()
}

// GetSelected returns the selected package config, or nil if none
func (m *PackageSelectModel) GetSelected() *config.PackageConfig {
	if m.selected < 0 || m.selected >= len(m.items) {
		return nil
	}
	return &m.items[m.selected]
}

// NewPackageSelectModel creates a new package select model
func NewPackageSelectModel(items []config.PackageConfig, title string) *PackageSelectModel {
	return &PackageSelectModel{
		items:    items,
		cursor:   0,
		title:    title,
		selected: -1,
	}
}

// SearchResultSelectModel represents a UI model for search result selection
type SearchResultSelectModel struct {
	items    []provider.SearchResult
	cursor   int
	title    string
	quitting bool
	selected int // -1 means none selected
}

// Init initializes the model
func (m *SearchResultSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *SearchResultSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			// Toggle selection
			if m.selected == m.cursor {
				m.selected = -1
			} else {
				m.selected = m.cursor
			}
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m *SearchResultSelectModel) View() string {
	if m.quitting {
		// Show final selection
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%s: ", m.title))
		selected := m.GetSelected()
		if selected == nil {
			b.WriteString("(none)\n")
		} else {
			b.WriteString(fmt.Sprintf("%s (ID: %s)\n", selected.Name, selected.ID))
		}
		return b.String()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s:\n\n", m.title))

	for i, item := range m.items {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		checkbox := "[ ]"
		if m.selected == i {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s%s %s", prefix, checkbox, item.Name)
		if item.ID != "" {
			line += fmt.Sprintf(" (ID: %s)", item.ID)
		}
		if item.Description != "" {
			line += fmt.Sprintf(" - %s", item.Description)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString("  ↑/↓: Move  Space: Select/Deselect  Enter: Confirm  q: Quit\n")

	return b.String()
}

// GetSelected returns the selected search result, or nil if none
func (m *SearchResultSelectModel) GetSelected() *provider.SearchResult {
	if m.selected < 0 || m.selected >= len(m.items) {
		return nil
	}
	return &m.items[m.selected]
}

// NewSearchResultSelectModel creates a new search result select model
func NewSearchResultSelectModel(items []provider.SearchResult, title string) *SearchResultSelectModel {
	return &SearchResultSelectModel{
		items:    items,
		cursor:   0,
		title:    title,
		selected: -1,
	}
}
