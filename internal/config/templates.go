package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProfileTemplate represents a profile template
type ProfileTemplate struct {
	Name     string          `json:"name"`
	Profiles []ProfileConfig `json:"profiles"`
}

// TemplatesConfig represents the collection of template configurations
type TemplatesConfig struct {
	Templates []ProfileTemplate `json:"templates"`
}

// GetDefaultTemplates returns the default templates embedded in the code
func GetDefaultTemplates() []ProfileTemplate {
	return []ProfileTemplate{
		{
			Name: "stable-only",
			Profiles: []ProfileConfig{
				{
					Name:  "<profile_name>",
					Stage: "stable",
				},
			},
		},
		{
			Name: "stable-trial",
			Profiles: []ProfileConfig{
				{
					Name:  "<profile_name>",
					Stage: "stable",
				},
				{
					Name:      "<profile_name>.trial",
					Stage:     "trial",
					Extends:   []string{"<profile_name>"},
					PromoteTo: "<profile_name>",
				},
			},
		},
	}
}

// GetTemplatesConfigPath returns the path to the templates.json file
func GetTemplatesConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "templates.json"), nil
}

// LoadTemplatesConfig loads the templates configuration from JSON file
func LoadTemplatesConfig() (*TemplatesConfig, error) {
	configPath, err := GetTemplatesConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &TemplatesConfig{Templates: []ProfileTemplate{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config TemplatesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveTemplatesConfig saves the templates configuration to JSON file
func SaveTemplatesConfig(config *TemplatesConfig) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetTemplatesConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetAllTemplates returns all templates (default + user-defined)
func GetAllTemplates() ([]ProfileTemplate, error) {
	// Get default templates
	defaultTemplates := GetDefaultTemplates()

	// Get user-defined templates
	userConfig, err := LoadTemplatesConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading templates config: %w", err)
	}

	// Create a map to track template names (default templates take precedence)
	templateMap := make(map[string]ProfileTemplate)
	for _, tmpl := range defaultTemplates {
		templateMap[tmpl.Name] = tmpl
	}

	// Add user-defined templates (only if not overridden by default)
	for _, tmpl := range userConfig.Templates {
		if _, exists := templateMap[tmpl.Name]; !exists {
			templateMap[tmpl.Name] = tmpl
		}
	}

	// Convert map to slice
	templates := make([]ProfileTemplate, 0, len(templateMap))
	for _, tmpl := range templateMap {
		templates = append(templates, tmpl)
	}

	return templates, nil
}

// GetTemplate returns a template by name
func GetTemplate(name string) (*ProfileTemplate, error) {
	templates, err := GetAllTemplates()
	if err != nil {
		return nil, err
	}

	for _, tmpl := range templates {
		if tmpl.Name == name {
			return &tmpl, nil
		}
	}

	return nil, fmt.Errorf("template '%s' not found", name)
}

// AddOrUpdateTemplate adds or updates a user-defined template
func AddOrUpdateTemplate(template ProfileTemplate) error {
	config, err := LoadTemplatesConfig()
	if err != nil {
		return err
	}

	// Check if it's a default template
	defaultTemplates := GetDefaultTemplates()
	for _, dt := range defaultTemplates {
		if dt.Name == template.Name {
			return fmt.Errorf("cannot override default template '%s'", template.Name)
		}
	}

	// Check if template already exists
	found := false
	for i, t := range config.Templates {
		if t.Name == template.Name {
			config.Templates[i] = template
			found = true
			break
		}
	}

	// If not found, add it
	if !found {
		config.Templates = append(config.Templates, template)
	}

	return SaveTemplatesConfig(config)
}

// RemoveTemplate removes a user-defined template
func RemoveTemplate(name string) error {
	// Check if it's a default template
	defaultTemplates := GetDefaultTemplates()
	for _, dt := range defaultTemplates {
		if dt.Name == name {
			return fmt.Errorf("cannot remove default template '%s'", name)
		}
	}

	config, err := LoadTemplatesConfig()
	if err != nil {
		return err
	}

	// Find and remove the template
	found := false
	for i, t := range config.Templates {
		if t.Name == name {
			config.Templates = append(config.Templates[:i], config.Templates[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("template '%s' not found", name)
	}

	return SaveTemplatesConfig(config)
}

// ApplyTemplate applies a template to create profiles with the given profile name
func ApplyTemplate(template *ProfileTemplate, profileName string) ([]ProfileConfig, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	if profileName == "" {
		return nil, fmt.Errorf("profile name is required")
	}

	profiles := make([]ProfileConfig, 0, len(template.Profiles))
	for _, profile := range template.Profiles {
		// Replace <profile_name> with actual profile name
		newProfile := profile
		newProfile.Name = strings.ReplaceAll(profile.Name, "<profile_name>", profileName)

		// Replace <profile_name> in Extends
		if len(profile.Extends) > 0 {
			newExtends := make([]string, len(profile.Extends))
			for i, ext := range profile.Extends {
				newExtends[i] = strings.ReplaceAll(ext, "<profile_name>", profileName)
			}
			newProfile.Extends = newExtends
		}

		// Replace <profile_name> in PromoteTo
		if profile.PromoteTo != "" {
			newProfile.PromoteTo = strings.ReplaceAll(profile.PromoteTo, "<profile_name>", profileName)
		}

		profiles = append(profiles, newProfile)
	}

	return profiles, nil
}
