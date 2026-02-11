package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ProviderConfig represents a provider configuration
type ProviderConfig struct {
	Name        string    `json:"name"`
	InstalledAt time.Time `json:"installed_at"`
	Version     string    `json:"version,omitempty"`
	DependsOn   []string  `json:"depends_on,omitempty"`
}

// ProvidersConfig represents the collection of provider configurations
type ProvidersConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

// defaultProviderDependencies defines default dependencies between providers.
var defaultProviderDependencies = map[string][]string{
	"mas": {"brew"},
}

// GetDefaultProviderDependencies returns default dependencies for a provider.
func GetDefaultProviderDependencies(providerName string) []string {
	deps, ok := defaultProviderDependencies[providerName]
	if !ok {
		return nil
	}
	return copyStringSlice(deps)
}

// LoadProvidersConfig loads the providers configuration from JSON file
func LoadProvidersConfig() (*ProvidersConfig, error) {
	configPath, err := GetProvidersConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ProvidersConfig{Providers: []ProviderConfig{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ProvidersConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveProvidersConfig saves the providers configuration to JSON file
func SaveProvidersConfig(config *ProvidersConfig) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetProvidersConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddOrUpdateProvider adds or updates a provider in the configuration
func AddOrUpdateProvider(provider ProviderConfig) error {
	config, err := LoadProvidersConfig()
	if err != nil {
		return err
	}

	provider.Name = strings.TrimSpace(provider.Name)
	provider.DependsOn = normalizeProviderDependencies(provider.Name, provider.DependsOn)

	// Check if provider already exists
	found := false
	for i, p := range config.Providers {
		if p.Name == provider.Name {
			if provider.DependsOn == nil {
				// Keep existing dependency settings when callers only update version/timestamps.
				provider.DependsOn = resolveConfiguredProviderDependencies(provider.Name, p.DependsOn)
			}
			config.Providers[i] = provider
			found = true
			break
		}
	}

	// If not found, add it
	if !found {
		if provider.DependsOn == nil {
			provider.DependsOn = GetDefaultProviderDependencies(provider.Name)
		}
		config.Providers = append(config.Providers, provider)
	}

	return SaveProvidersConfig(config)
}

// GetProvider returns a provider configuration by name
func GetProvider(name string) (*ProviderConfig, error) {
	config, err := LoadProvidersConfig()
	if err != nil {
		return nil, err
	}

	for _, p := range config.Providers {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, nil // Provider not found
}

// OrderProvidersByDependency sorts providers so dependencies come first.
// Dependencies outside of providerNames are ignored.
func OrderProvidersByDependency(providerNames []string) ([]string, error) {
	uniqueNames := uniqueProviderNames(providerNames)
	if len(uniqueNames) == 0 {
		return []string{}, nil
	}

	providerDeps, err := loadProviderDependencyMap()
	if err != nil {
		return nil, err
	}

	return topologicalSortProviders(uniqueNames, providerDeps)
}

// ResolveProvidersWithDependencies expands providerNames with dependencies and sorts them.
func ResolveProvidersWithDependencies(providerNames []string) ([]string, error) {
	uniqueNames := uniqueProviderNames(providerNames)
	if len(uniqueNames) == 0 {
		return []string{}, nil
	}

	providerDeps, err := loadProviderDependencyMap()
	if err != nil {
		return nil, err
	}

	expandedSet := make(map[string]bool)
	queue := make([]string, 0, len(uniqueNames))
	for _, name := range uniqueNames {
		if !expandedSet[name] {
			expandedSet[name] = true
			queue = append(queue, name)
		}
	}

	for i := 0; i < len(queue); i++ {
		name := queue[i]
		for _, dep := range resolveConfiguredProviderDependencies(name, providerDeps[name]) {
			if dep == "" || dep == name {
				continue
			}
			if !expandedSet[dep] {
				expandedSet[dep] = true
				queue = append(queue, dep)
			}
		}
	}

	expandedNames := make([]string, 0, len(expandedSet))
	for name := range expandedSet {
		expandedNames = append(expandedNames, name)
	}

	return topologicalSortProviders(expandedNames, providerDeps)
}

func loadProviderDependencyMap() (map[string][]string, error) {
	cfg, err := LoadProvidersConfig()
	if err != nil {
		return nil, err
	}

	depsByProvider := make(map[string][]string, len(cfg.Providers))
	for _, p := range cfg.Providers {
		depsByProvider[p.Name] = normalizeProviderDependencies(p.Name, p.DependsOn)
	}
	return depsByProvider, nil
}

func topologicalSortProviders(providerNames []string, depsByProvider map[string][]string) ([]string, error) {
	set := make(map[string]bool, len(providerNames))
	for _, name := range providerNames {
		set[name] = true
	}

	inDegree := make(map[string]int, len(providerNames))
	dependents := make(map[string][]string, len(providerNames))
	for _, name := range providerNames {
		inDegree[name] = 0
	}

	for _, name := range providerNames {
		depSeen := make(map[string]bool)
		for _, dep := range resolveConfiguredProviderDependencies(name, depsByProvider[name]) {
			if dep == "" || dep == name || !set[dep] || depSeen[dep] {
				continue
			}
			depSeen[dep] = true
			inDegree[name]++
			dependents[dep] = append(dependents[dep], name)
		}
	}

	var queue []string
	for _, name := range providerNames {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		next := append([]string(nil), dependents[current]...)
		sort.Strings(next)
		for _, name := range next {
			inDegree[name]--
			if inDegree[name] == 0 {
				queue = append(queue, name)
			}
		}
		sort.Strings(queue)
	}

	if len(order) != len(providerNames) {
		cycleNames := append([]string(nil), providerNames...)
		sort.Strings(cycleNames)
		return nil, fmt.Errorf("provider: cycle in depends_on dependency (%s)", strings.Join(cycleNames, ", "))
	}

	return order, nil
}

func uniqueProviderNames(providerNames []string) []string {
	seen := make(map[string]bool, len(providerNames))
	var unique []string
	for _, name := range providerNames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		unique = append(unique, trimmed)
	}
	return unique
}

func normalizeProviderDependencies(providerName string, dependsOn []string) []string {
	if dependsOn == nil {
		return nil
	}

	seen := make(map[string]bool, len(dependsOn))
	var normalized []string
	for _, dep := range dependsOn {
		trimmed := strings.TrimSpace(dep)
		if trimmed == "" || trimmed == providerName || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		normalized = append(normalized, trimmed)
	}

	if len(normalized) == 0 {
		return []string{}
	}
	sort.Strings(normalized)
	return normalized
}

func resolveConfiguredProviderDependencies(providerName string, configured []string) []string {
	if configured != nil {
		return copyStringSlice(configured)
	}
	return GetDefaultProviderDependencies(providerName)
}

func copyStringSlice(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
