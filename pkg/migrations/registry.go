package migrations

import "sort"

// ModelVersion represents a version of a model
type ModelVersion struct {
	Name     string // "user", "item"
	Version  string // "1.0.0"
	Current  interface{}
	Previous interface{}
}

var registry []ModelVersion

// RegisterModel registers a model version
func RegisterModel(name, version string, current, previous interface{}) {
	registry = append(registry, ModelVersion{
		Name:     name,
		Version:  version,
		Current:  current,
		Previous: previous,
	})

	// Keep versions sorted
	sort.Slice(registry, func(i, j int) bool {
		return registry[i].Name < registry[j].Name ||
			(registry[i].Name == registry[j].Name && registry[i].Version < registry[j].Version)
	})
}

// GetModelVersions returns all registered model versions
func GetModelVersions() []ModelVersion {
	return registry
}
