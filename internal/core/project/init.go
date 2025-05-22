// Package project provides project initialization and management functionality
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings" // Added missing import
)

// YamlFileName is the name of the loom configuration file
const YamlFileName = "loom.yaml"

// LoomConfig represents the structure of loom.yaml
// Note: Renamed from Config to LoomConfig and Version type changed to string
type LoomConfig struct {
	Version string   `yaml:"version"`
	Threads []Thread `yaml:"threads"`
}

// Thread represents a thread entry in loom.yaml
type Thread struct {
	Name   string              `yaml:"name"`
	Source string              `yaml:"source"`
	Files  map[string][]string `yaml:"files,omitempty"`
}

// IsFileOwned checks if a given file path is owned by any thread in the config.
// It returns the name of the owning thread and true if owned, otherwise an empty string and false.
func (lc *LoomConfig) IsFileOwned(filePath string, projectRoot string) (string, bool) {
	relPath, err := filepath.Rel(projectRoot, filePath)
	if err != nil {
		// If we can't make it relative, assume it's not owned or handle error appropriately
		return "", false
	}
	relPath = filepath.ToSlash(relPath) // Ensure consistent path separators

	for _, thread := range lc.Threads {
		if thread.Files == nil {
			continue
		}
		for dir, files := range thread.Files {
			// Normalize dir to ensure it ends with a slash if it's not "./"
			normalizedDir := dir
			if normalizedDir != "./" && !strings.HasSuffix(normalizedDir, "/") {
				normalizedDir += "/"
			}

			for _, ownedFile := range files {
				var fullOwnedPath string
				if normalizedDir == "./" {
					fullOwnedPath = ownedFile
				} else {
					fullOwnedPath = filepath.ToSlash(filepath.Join(normalizedDir, ownedFile))
				}

				if fullOwnedPath == relPath {
					return thread.Name, true
				}
			}
		}
	}
	return "", false
}

// InitProject initializes a new loom.yaml file in the current directory
func InitProject() error {
	// Check if loom.yaml already exists
	if _, err := os.Stat(YamlFileName); err == nil { // Changed fileInfo to _
		// File exists, check if it's empty or only comments/whitespace
		content, err := os.ReadFile(YamlFileName)
		if err != nil {
			return fmt.Errorf("failed to read existing %s: %w", YamlFileName, err)
		}

		trimmedContent := strings.TrimSpace(string(content))
		if trimmedContent != "" {
			// Check if the content is only comments
			lines := strings.Split(string(content), "\n")
			isEmptyOrComments := true
			for _, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
					isEmptyOrComments = false
					break
				}
			}
			if !isEmptyOrComments {
				return fmt.Errorf("%s already exists and is not empty", YamlFileName)
			}
		}
		// If we are here, the file exists but is empty or comments-only, so we can overwrite.
	} else if !os.IsNotExist(err) {
		// Some other error occurred when stating the file
		return fmt.Errorf("failed to check for %s: %w", YamlFileName, err)
	}

	// Create a minimal loom.yaml content
	// Note: Changed version to "1" (string)
	contentString := `# loom.yaml - Loom project configuration file
version: "1"
threads: []
` // Renamed content to contentString to avoid conflict

	// Write the content to loom.yaml
	errWrite := os.WriteFile(YamlFileName, []byte(contentString), 0644) // Used contentString and new err var
	if errWrite != nil {
		return fmt.Errorf("failed to create %s: %w", YamlFileName, errWrite)
	}

	return nil
}

// GetProjectRoot attempts to find the root of the project by locating loom.yaml
// If not found, returns the current directory
func GetProjectRoot() (string, error) {
	// Start at the current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if loom.yaml exists in the current directory
	if _, err := os.Stat(filepath.Join(dir, YamlFileName)); err == nil {
		return dir, nil
	}

	// For simplicity, just return the current directory if loom.yaml doesn't exist
	// In the future, we might want to search up the directory tree for loom.yaml
	return dir, nil
}
