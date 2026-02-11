package config

import (
	"reflect"
	"testing"
	"time"
)

func TestAddOrUpdateProvider_SetsDefaultDependencies(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	providerCfg := ProviderConfig{
		Name:        "mas",
		InstalledAt: time.Now(),
		Version:     "1.0.0",
	}
	if err := AddOrUpdateProvider(providerCfg); err != nil {
		t.Fatal(err)
	}

	got, err := GetProvider("mas")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("GetProvider(mas) returned nil")
	}

	wantDeps := []string{"brew"}
	if !reflect.DeepEqual(got.DependsOn, wantDeps) {
		t.Errorf("mas depends_on = %v, want %v", got.DependsOn, wantDeps)
	}
}

func TestAddOrUpdateProvider_PreservesExistingDependenciesWhenOmitted(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := AddOrUpdateProvider(ProviderConfig{
		Name:        "mas",
		InstalledAt: time.Now(),
		Version:     "1.0.0",
		DependsOn:   []string{"brew", "manual"},
	}); err != nil {
		t.Fatal(err)
	}

	// Simulate version refresh where caller doesn't send depends_on.
	if err := AddOrUpdateProvider(ProviderConfig{
		Name:        "mas",
		InstalledAt: time.Now(),
		Version:     "2.0.0",
	}); err != nil {
		t.Fatal(err)
	}

	got, err := GetProvider("mas")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("GetProvider(mas) returned nil")
	}

	wantDeps := []string{"brew", "manual"}
	if !reflect.DeepEqual(got.DependsOn, wantDeps) {
		t.Errorf("mas depends_on after update = %v, want %v", got.DependsOn, wantDeps)
	}
}

func TestOrderProvidersByDependency_Defaults(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	got, err := OrderProvidersByDependency([]string{"mas", "brew"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"brew", "mas"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("OrderProvidersByDependency() = %v, want %v", got, want)
	}
}

func TestResolveProvidersWithDependencies_Defaults(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	got, err := ResolveProvidersWithDependencies([]string{"mas"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"brew", "mas"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ResolveProvidersWithDependencies() = %v, want %v", got, want)
	}
}

func TestOrderProvidersByDependency_Cycle(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := SaveProvidersConfig(&ProvidersConfig{
		Providers: []ProviderConfig{
			{Name: "a", DependsOn: []string{"b"}},
			{Name: "b", DependsOn: []string{"a"}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := OrderProvidersByDependency([]string{"a", "b"}); err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}
