package profile

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
)

// OrderedMultiSelectModel represents a UI model for ordered multiple selection
type OrderedMultiSelectModel struct {
	items        []config.ProfileConfig
	selected     map[int]int // map[itemIndex]orderNumber (0 = not selected)
	nextOrder    int
	cursor       int
	title        string
	excludeName  string
	quitting     bool
	selectedKeys []int // ordered list of selected item indices
}

// Init initializes the model
func (m *OrderedMultiSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *OrderedMultiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if _, isSelected := m.selected[m.cursor]; isSelected {
				// Deselect: remove from selection and reorder
				delete(m.selected, m.cursor)
				// Remove from selectedKeys
				for i, key := range m.selectedKeys {
					if key == m.cursor {
						m.selectedKeys = append(m.selectedKeys[:i], m.selectedKeys[i+1:]...)
						break
					}
				}
				// Reorder remaining items
				for i, key := range m.selectedKeys {
					m.selected[key] = i + 1
				}
				// Update nextOrder
				if len(m.selectedKeys) > 0 {
					m.nextOrder = len(m.selectedKeys) + 1
				} else {
					m.nextOrder = 1
				}
			} else {
				// Select: add to selection with next order
				m.selected[m.cursor] = m.nextOrder
				m.selectedKeys = append(m.selectedKeys, m.cursor)
				m.nextOrder++
			}
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m *OrderedMultiSelectModel) View() string {
	if m.quitting {
		// Show final selection
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%s: ", m.title))
		selected := m.GetSelected()
		if len(selected) == 0 {
			b.WriteString("(none)\n")
		} else {
			b.WriteString(strings.Join(selected, ", "))
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
		if order, isSelected := m.selected[i]; isSelected {
			checkbox = fmt.Sprintf("[%d]", order)
		}

		line := fmt.Sprintf("%s%s %s", prefix, checkbox, item.Name)
		if item.Description != "" {
			line += fmt.Sprintf(" - %s", item.Description)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString("  ↑/↓: Move  Space: Select/Deselect  Enter: Confirm  q: Quit\n")

	return b.String()
}

// GetSelected returns the selected items in order
func (m *OrderedMultiSelectModel) GetSelected() []string {
	result := make([]string, len(m.selectedKeys))
	for i, key := range m.selectedKeys {
		result[i] = m.items[key].Name
	}
	return result
}

// NewOrderedMultiSelectModel creates a new ordered multi-select model
func NewOrderedMultiSelectModel(items []config.ProfileConfig, title, excludeName string) *OrderedMultiSelectModel {
	// Filter out excluded profile
	filtered := []config.ProfileConfig{}
	for _, item := range items {
		if item.Name != excludeName {
			filtered = append(filtered, item)
		}
	}

	return &OrderedMultiSelectModel{
		items:       filtered,
		selected:    make(map[int]int),
		nextOrder:   1,
		cursor:      0,
		title:       title,
		excludeName: excludeName,
		selectedKeys: []int{},
	}
}

// SingleSelectModel represents a UI model for single selection
type SingleSelectModel struct {
	items       []config.ProfileConfig
	cursor      int
	title       string
	excludeName string
	quitting    bool
	selected    int // -1 means none selected
}

// Init initializes the model
func (m *SingleSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *SingleSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *SingleSelectModel) View() string {
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
		if m.selected == i {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s%s %s", prefix, checkbox, item.Name)
		if item.Description != "" {
			line += fmt.Sprintf(" - %s", item.Description)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString("  ↑/↓: Move  Space: Select/Deselect  Enter: Confirm  q: Quit\n")

	return b.String()
}

// GetSelected returns the selected item name, or empty string if none
func (m *SingleSelectModel) GetSelected() string {
	if m.selected < 0 || m.selected >= len(m.items) {
		return ""
	}
	return m.items[m.selected].Name
}

// NewSingleSelectModel creates a new single select model
func NewSingleSelectModel(items []config.ProfileConfig, title, excludeName string) *SingleSelectModel {
	// Filter out excluded profile
	filtered := []config.ProfileConfig{}
	for _, item := range items {
		if item.Name != excludeName {
			filtered = append(filtered, item)
		}
	}

	return &SingleSelectModel{
		items:       filtered,
		cursor:      0,
		title:       title,
		excludeName: excludeName,
		selected:    -1,
	}
}

// PackageDuplicationSelectModel represents a UI model for package duplication selection
type PackageDuplicationSelectModel struct {
	options  []string
	descriptions []string
	cursor   int
	selected int // -1 means default (warn)
	quitting bool
}

// Init initializes the model
func (m *PackageDuplicationSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *PackageDuplicationSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case " ":
			// Toggle selection
			if m.selected == m.cursor {
				m.selected = -1 // Default to warn
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
func (m *PackageDuplicationSelectModel) View() string {
	if m.quitting {
		// Show final selection
		var b strings.Builder
		b.WriteString("Package duplication: ")
		selected := m.GetSelected()
		b.WriteString(selected)
		b.WriteString("\n")
		return b.String()
	}

	var b strings.Builder
	b.WriteString("\nPackage duplication:\n\n")

	for i, option := range m.options {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		checkbox := "[ ]"
		if m.selected == i || (m.selected == -1 && i == 2) {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s%s %s", prefix, checkbox, option)
		if i < len(m.descriptions) {
			line += fmt.Sprintf(" - %s", m.descriptions[i])
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString("  ↑/↓: Move  Space: Select/Deselect  Enter: Confirm  q: Quit\n")

	return b.String()
}

// GetSelected returns the selected option, or "warn" as default
func (m *PackageDuplicationSelectModel) GetSelected() string {
	if m.selected < 0 || m.selected >= len(m.options) {
		return "warn" // Default
	}
	return m.options[m.selected]
}

// NewPackageDuplicationSelectModel creates a new package duplication select model
func NewPackageDuplicationSelectModel() *PackageDuplicationSelectModel {
	return &PackageDuplicationSelectModel{
		options: []string{"forbid", "allow", "warn"},
		descriptions: []string{
			"Packages in this profile cannot be installed in other profiles",
			"Packages can be installed in other profiles",
			"Warn when installing packages in other profiles (default)",
		},
		cursor:   0,
		selected: 2, // Default to warn
	}
}

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
