// Package config handles persistent local configuration for lazytailscale.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "lazytailscale"), nil
}

// LoadUsernames reads the saved hostname→username map from disk.
// Returns an empty map on any error (missing file, parse error, etc.).
func LoadUsernames() map[string]string {
	return loadStringMap("usernames.json")
}

// SaveUsernames writes the hostname→username map to disk.
// Errors are silently ignored — persistence is best-effort.
func SaveUsernames(m map[string]string) {
	saveStringMap("usernames.json", m)
}

// LoadPorts reads the saved hostname→port map from disk.
func LoadPorts() map[string]string {
	return loadStringMap("ports.json")
}

// SavePorts writes the hostname→port map to disk.
func SavePorts(m map[string]string) {
	saveStringMap("ports.json", m)
}

func loadStringMap(filename string) map[string]string {
	dir, err := configDir()
	if err != nil {
		return make(map[string]string)
	}
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return make(map[string]string)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return make(map[string]string)
	}
	return m
}

func saveStringMap(filename string, m map[string]string) {
	dir, err := configDir()
	if err != nil {
		return
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, filename), data, 0o600)
}
