package migrations

import "sort"

type ModelVersion struct {
	Name     string // "user", "item"
	Version  string // "1.0.0"
	Current  interface{}
	Previous interface{}
}

var registry []ModelVersion

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

func GetModelVersions() []ModelVersion {
	return registry
}
