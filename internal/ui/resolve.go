package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
)

// ResolvePackageByName finds packages by display name. If exactly one match, returns it.
// If multiple matches, runs interactive selection and returns the selected package.
func ResolvePackageByName(packageName string) (*config.PackageConfig, error) {
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading packages config: %w", err)
	}
	var matching []config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		if pkg.Name == packageName {
			matching = append(matching, pkg)
		}
	}
	if len(matching) == 0 {
		return nil, fmt.Errorf("package '%s' not found", packageName)
	}
	if len(matching) == 1 {
		return &matching[0], nil
	}
	model := NewPackageSelectModel(matching, fmt.Sprintf("Select package (found %d matching)", len(matching)))
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return nil, fmt.Errorf("error running UI: %w", err)
	}
	selected := model.GetSelected()
	if selected == nil {
		return nil, fmt.Errorf("package selection is required")
	}
	return selected, nil
}
