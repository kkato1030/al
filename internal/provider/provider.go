package provider

// SearchResult represents a search result from a provider
type SearchResult struct {
	ID          string
	Name        string
	Description string
}

// Provider represents a package manager provider
type Provider interface {
	// Name returns the name of the provider
	Name() string

	// CheckInstalled checks if the package manager is installed
	CheckInstalled() (bool, error)

	// Install installs the package manager
	Install() error

	// SetupConfig sets up the configuration for the provider
	SetupConfig() error

	// InstallPackage installs a package using the provider
	InstallPackage(packageID string) error

	// UninstallPackage uninstalls a package using the provider
	UninstallPackage(packageID string) error

	// SearchPackage searches for packages matching the query
	SearchPackage(query string) ([]SearchResult, error)
}
