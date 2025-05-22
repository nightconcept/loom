// Title: List Command Implementation
// Purpose: Implements the `loom list` command to display active threads in the project.

package cli

import (
	"fmt"
	"os"
	"path/filepath" // Added for store path operations
	"strings"       // Added for string operations

	"loom/internal/core/globalconfig" // Added for global config access
	"loom/internal/core/project"      // Import the project package

	"gopkg.in/yaml.v3"
)

// Remove local Thread and LoomConfig structs, use project package versions

// listThreads reads the loom.yaml file and lists active threads.
// It also lists available threads from configured local stores.
func listThreads() error {
	if err := printActiveProjectThreads(); err != nil {
		return err
	}

	fmt.Println("\nAvailable store threads:")
	gConf, err := globalconfig.LoadGlobalConfig() // This loads the actual global config struct
	if err != nil {
		return fmt.Errorf("failed to load global Loom configuration: %w", err)
	}

	foundAnyStoreThreads := false
	if len(gConf.Stores) == 0 { // gConf is already a pointer, so no need to check gConf == nil separately if LoadGlobalConfig guarantees non-nil on no error
		fmt.Println("No global thread stores configured. Use 'loom config add local <path_to_store> [name]' to add one.")
	} else {
		// Pass the loaded gConf directly to printGlobalStoreThreads
		foundGlobalStoreThreads, errPrintingGlobalStores := printGlobalStoreThreads(gConf)
		if errPrintingGlobalStores != nil {
			fmt.Fprintf(os.Stderr, "Error processing global stores: %v\n", errPrintingGlobalStores)
		}
		foundAnyStoreThreads = foundAnyStoreThreads || foundGlobalStoreThreads
	}

	foundProjectStoreThreads, errPrintingProjectStore := printProjectStoreThreads()
	if errPrintingProjectStore != nil {
		fmt.Fprintf(os.Stderr, "Error processing project store: %v\n", errPrintingProjectStore)
	}
	foundAnyStoreThreads = foundAnyStoreThreads || foundProjectStoreThreads

	// Simplified conditional logic for final messages
	if !foundAnyStoreThreads {
		if len(gConf.Stores) == 0 { // No global stores configured and no project store threads found
			// Message about no global stores already printed. Potentially add a note if project store was also empty/missing.
			// Or rely on printProjectStoreThreads to have printed its specific message.
		} else { // Global stores are configured, but no threads were found in them or in the project store.
			hasLocalStore := false
			for _, store := range gConf.Stores {
				if store.Type == "local" {
					hasLocalStore = true
					break
				}
			}
			if hasLocalStore {
				fmt.Println("No threads found in any configured local stores or the project store, or stores were inaccessible.")
			} else {
				fmt.Println("No 'local' type global thread stores configured. Other store types are not yet supported for listing.")
			}
		}
	}

	return nil
}

// printGlobalStoreThreads iterates over configured global stores and prints their threads.
// It returns true if any threads were found in global stores, false otherwise.
// The gConf parameter should be the struct type defined in the globalconfig package.
func printGlobalStoreThreads(gConf *globalconfig.GlobalLoomConfig) (bool, error) { // Corrected type to globalconfig.GlobalLoomConfig
	foundAny := false
	for _, store := range gConf.Stores {
		if store.Type == "local" { // For now, only supporting local stores
			fmt.Printf("\nStore: %s (Type: %s, Path: %s)\n", store.Name, store.Type, store.Path)
			threads, err := listThreadsInStore(store.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error listing threads in store '%s': %v\n", store.Name, err)
				continue // Continue to the next store
			}
			if len(threads) == 0 {
				fmt.Println("  No threads found in this store.")
			} else {
				foundAny = true
				for _, threadName := range threads {
					fmt.Printf("  - %s\n", threadName)
				}
			}
		}
	}
	return foundAny, nil
}

// printProjectStoreThreads lists threads from the project-specific .loom store.
// It returns true if any threads were found in the project store, false otherwise.
func printProjectStoreThreads() (bool, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine current directory to check for project store: %v\n", err)
		return false, nil // Not a fatal error for listing, just can't check project store
	}

	projectStorePath := filepath.Join(projectRoot, ".loom")
	if _, statErr := os.Stat(projectStorePath); statErr == nil {
		fmt.Printf("\nProject Store (.loom):\n")
		threads, listErr := listThreadsInStore(projectStorePath)
		if listErr != nil {
			fmt.Fprintf(os.Stderr, "  Error listing threads in project store: %v\n", listErr)
			return false, nil // Error occurred, but treat as no threads found for the purpose of the caller
		}
		if len(threads) == 0 {
			fmt.Println("  No threads found in this store.")
			return false, nil
		}
		for _, threadName := range threads {
			fmt.Printf("  - %s\n", threadName)
		}
		return true, nil // Threads found
	} else if !os.IsNotExist(statErr) {
		// Report error if .loom exists but cannot be stated, unless it's simply not found
		fmt.Fprintf(os.Stderr, "Warning: Could not stat project store at '%s': %v\n", projectStorePath, statErr)
	}
	return false, nil // Project store does not exist or error stating it
}

// printActiveProjectThreads handles reading loom.yaml and printing active project threads.
func printActiveProjectThreads() error {
	file, err := os.Open(project.YamlFileName) // Use project.YamlFileName
	if err != nil {
		// If loom.yaml doesn't exist, it's not an error for listing, just means no project threads
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to open %s: %w", project.YamlFileName, err)
		}
		fmt.Println("No active project configuration (loom.yaml) found.")
		return nil // Not an error in this context
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file %s: %v\n", project.YamlFileName, err)
		}
	}()

	var projectConfig project.LoomConfig // Use project.LoomConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&projectConfig); err != nil {
		return fmt.Errorf("failed to parse %s: %w", project.YamlFileName, err)
	}

	gConfForActive, _ := globalconfig.LoadGlobalConfig() // Load global config to check store names

	if len(projectConfig.Threads) == 0 {
		fmt.Println("No threads are currently active in the project.")
	} else {
		fmt.Println("Active project threads:")
		for _, thread := range projectConfig.Threads { // Iterate over Thread structs
			displaySource := thread.Source
			// Check if the source matches a known local store and format accordingly
			if gConfForActive != nil {
				for _, store := range gConfForActive.Stores {
					if store.Type == "local" && strings.HasPrefix(thread.Source, store.Name) {
						// Ensure it's not a project store source that happens to start with a store name
						if !strings.HasPrefix(thread.Source, "project:") {
							displaySource = fmt.Sprintf("local:%s", thread.Source)
							break
						}
					}
				}
			}
			fmt.Printf("- %s (Source: %s)\n", thread.Name, displaySource) // Print thread name and source
		}
	}
	return nil
}

// listThreadsInStore lists subdirectories in a given store path that appear to be valid Loom threads.
// A directory is considered a thread if it contains a 'config.yml' file or a '_thread/' subdirectory.
func listThreadsInStore(storePath string) ([]string, error) {
	entries, err := os.ReadDir(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read store directory '%s': %w", storePath, err)
	}

	var threadNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			threadName := entry.Name()
			// Check for config.yml or _thread/ directory to qualify as a thread
			configFilePath := filepath.Join(storePath, threadName, "config.yml")
			threadDirPath := filepath.Join(storePath, threadName, "_thread")

			_, errConfig := os.Stat(configFilePath)
			_, errDir := os.Stat(threadDirPath)

			if errConfig == nil || errDir == nil { // If either exists, it's a thread
				threadNames = append(threadNames, threadName)
			}
		}
	}
	return threadNames, nil
}

// ExecuteListCommand is the entry point for the `loom list` command.
func ExecuteListCommand() {
	if err := listThreads(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
