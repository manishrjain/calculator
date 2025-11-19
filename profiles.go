package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const profilesDir = ".rentobuy_profiles"

// ensureProfilesDir creates the profiles directory if it doesn't exist
func ensureProfilesDir() error {
	return os.MkdirAll(profilesDir, 0755)
}

// listProfiles returns a sorted list of available profile names
func listProfiles() ([]string, error) {
	if err := ensureProfilesDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			// Remove .json extension
			name := strings.TrimSuffix(entry.Name(), ".json")
			profiles = append(profiles, name)
		}
	}

	sort.Strings(profiles)
	return profiles, nil
}

// loadProfile loads inputs from a named profile
func loadProfile(name string) (map[string]string, error) {
	profilePath := filepath.Join(profilesDir, name+".json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, err
	}

	var inputs map[string]string
	err = json.Unmarshal(data, &inputs)
	if err != nil {
		return nil, err
	}

	return inputs, nil
}

// saveProfile saves inputs to a named profile
func saveProfile(name string, inputs map[string]string) error {
	if err := ensureProfilesDir(); err != nil {
		return err
	}

	profilePath := filepath.Join(profilesDir, name+".json")
	data, err := json.MarshalIndent(inputs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(profilePath, data, 0644)
}
