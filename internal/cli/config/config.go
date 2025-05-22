// Package config implements the subcommands for the 'loom config' command.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"loom/internal/core/globalconfig"

	"github.com/urfave/cli/v2"
)

// Command returns the cli.Command for the "config" group.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage Loom's configuration for thread stores.",
		Subcommands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "Add a new thread store. Usage: loom config add <path_or_url>",
				ArgsUsage: "<path_or_url>",
				Action:    addStoreAction,
			},
			{
				Name:      "remove",
				Usage:     "Remove a configured thread store. Usage: loom config remove <name_or_path>",
				ArgsUsage: "<name_or_path>",
				Action:    removeStoreAction,
			},
			{
				Name:   "list",
				Usage:  "List all configured thread stores. Usage: loom config list",
				Action: listStoresAction,
			},
			// Remove subcommand will be added in Task 4.7
		},
	}
}

// inferStoreDetails infers the store type, name, and normalized path from the input.
// For now, it primarily handles local paths. GitHub URL handling is a placeholder.
func inferStoreDetails(pathOrURL string) (storeType string, storeName string, normalizedPathOrURL string, err error) {
	// Basic check for what might be a URL (very simplistic for now)
	if strings.HasPrefix(strings.ToLower(pathOrURL), "http:") || strings.HasPrefix(strings.ToLower(pathOrURL), "https:") || strings.Contains(strings.ToLower(pathOrURL), "github.com") {
		// Placeholder for GitHub URL handling
		// For now, assume it's a local path if it's not obviously a URL starting with http/https
		// This will be expanded in Task 4.4
		// For the purpose of this task, we will treat non-http/https prefixed paths as local.
		// return "github", "gh-" + filepath.Base(pathOrURL), pathOrURL, nil // Simplified for now
		return "", "", "", fmt.Errorf("github URL store type not yet fully implemented, path was: %s", pathOrURL)
	}

	// Assume local path
	storeType = "local"
	absPath, err := filepath.Abs(pathOrURL)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get absolute path for \"%s\": %w", pathOrURL, err)
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", "", fmt.Errorf("path \"%s\" does not exist", absPath)
		}
		return "", "", "", fmt.Errorf("failed to stat path \"%s\": %w", absPath, err)
	}
	if !fileInfo.IsDir() {
		return "", "", "", fmt.Errorf("path \"%s\" is not a directory", absPath)
	}

	storeName = filepath.Base(absPath)
	normalizedPathOrURL = absPath
	return
}

// addStoreAction implements the logic for "loom config add <path_or_url>".
func addStoreAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("incorrect number of arguments. Expected <path_or_url>")
	}

	userInputPathOrURL := c.Args().Get(0)

	storeType, inferredStoreName, normalizedPathOrURL, err := inferStoreDetails(userInputPathOrURL)
	if err != nil {
		// If inferStoreDetails specifically said GitHub isn't implemented, pass that through.
		if strings.Contains(err.Error(), "github URL store type not yet fully implemented") {
			// For now, we treat this as a "not yet supported" rather than a hard error for CLI flow.
			// This allows local paths to work.
			// A more robust solution would be to have inferStoreDetails return a specific error type.
			fmt.Printf("Attempted to add a store that looks like a GitHub URL (%s). This functionality is planned but not yet implemented.\n", userInputPathOrURL)
			fmt.Println("Please provide a local directory path for now.")
			return nil // Or a specific error if preferred, but nil to allow local to proceed.
		}
		return err // Other errors from inferStoreDetails (e.g., path not found, not a dir)
	}

	// This check is now more specific after inferStoreDetails might return an error for GitHub paths.
	// If storeType is empty, it means inferStoreDetails couldn't determine it (e.g. GitHub not implemented path taken).
	if storeType == "" {
		// This case should ideally be handled by the error from inferStoreDetails already.
		// If we reach here, it implies a logic flaw or that inferStoreDetails allowed an empty type.
		return fmt.Errorf("could not determine store type for input: %s", userInputPathOrURL)
	}

	config, err := globalconfig.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global Loom configuration: %w", err)
	}

	finalStoreName := inferredStoreName
	nameConflictExists := false

	for _, existingStore := range config.Stores {
		// Path/URL conflict check (case-insensitive for paths, should be for URLs too)
		// For local paths, ensure OS-specific path comparison if necessary, though Abs should normalize.
		// For URLs, direct string comparison after normalization (e.g., lowercase, remove trailing slash)
		if strings.EqualFold(existingStore.Path, normalizedPathOrURL) {
			return fmt.Errorf("the path/url \"%s\" is already registered as store \"%s\" (type: %s)", normalizedPathOrURL, existingStore.Name, existingStore.Type)
		}
		if strings.EqualFold(existingStore.Name, inferredStoreName) {
			nameConflictExists = true
		}
	}

	if nameConflictExists {
		fmt.Printf("A store named \"%s\" already exists. The path \"%s\" is unique.\n", inferredStoreName, normalizedPathOrURL)
		fmt.Print("Please enter a new name for this store, or press Enter to cancel: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		customName := strings.TrimSpace(input)
		if customName == "" {
			fmt.Println("Store addition cancelled.")
			return nil
		}
		finalStoreName = customName

		// Re-check if the custom name also conflicts
		for _, existingStore := range config.Stores {
			if strings.EqualFold(existingStore.Name, finalStoreName) {
				return fmt.Errorf("the custom name \"%s\" also conflicts with an existing store. Please try again", finalStoreName)
			}
		}
	}

	newStore := globalconfig.Store{
		Name: finalStoreName,
		Type: storeType,
		Path: normalizedPathOrURL, // Store the normalized path/URL
	}

	config.Stores = append(config.Stores, newStore)

	if err := globalconfig.SaveGlobalConfig(config); err != nil {
		return fmt.Errorf("failed to save global Loom configuration: %w", err)
	}

	fmt.Printf("Successfully added %s store \"%s\" with path/url \"%s\"\n", storeType, finalStoreName, normalizedPathOrURL)
	configPath, _ := globalconfig.GetGlobalConfigPath()
	fmt.Printf("Configuration saved to: %s\n", configPath)
	return nil
}

// removeStoreAction implements the logic for "loom config remove <name_or_path>".
func removeStoreAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("incorrect number of arguments. Expected <name_or_path>")
	}

	nameOrPathToRemove := c.Args().Get(0)

	config, err := globalconfig.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global Loom configuration: %w", err)
	}

	found := false
	var updatedStores []globalconfig.Store
	removedStoreDetails := ""

	// Attempt to match by name first (case-insensitive)
	for _, store := range config.Stores {
		if strings.EqualFold(store.Name, nameOrPathToRemove) {
			found = true
			removedStoreDetails = fmt.Sprintf("store \"%s\" (type: %s, path/url: %s)", store.Name, store.Type, store.Path)
			// Skip adding this store to updatedStores
			continue
		}
		updatedStores = append(updatedStores, store)
	}

	// If not found by name, attempt to match by path/URL
	if !found {
		updatedStores = nil // Reset for path matching pass
		normalizedInputPath := nameOrPathToRemove
		// Attempt to normalize if it looks like a local path (not a URL)
		if !strings.HasPrefix(strings.ToLower(normalizedInputPath), "http:") && !strings.HasPrefix(strings.ToLower(normalizedInputPath), "https:") && !strings.Contains(strings.ToLower(normalizedInputPath), "github.com") {
			absPath, err := filepath.Abs(nameOrPathToRemove)
			if err == nil { // If Abs path resolution is successful
				normalizedInputPath = absPath
			}
			// If Abs fails, we proceed with the original input for comparison, it might be a non-existent path or a URL fragment that Abs can't handle.
		}

		for _, store := range config.Stores {
			// Compare normalized input with stored path (which is already normalized for local stores)
			if strings.EqualFold(store.Path, normalizedInputPath) {
				found = true
				removedStoreDetails = fmt.Sprintf("store \"%s\" (type: %s, path/url: %s)", store.Name, store.Type, store.Path)
				// Skip adding this store to updatedStores
				continue
			}
			updatedStores = append(updatedStores, store)
		}
	}

	if !found {
		return fmt.Errorf("store with name or path/url \"%s\" not found", nameOrPathToRemove)
	}

	config.Stores = updatedStores

	if err := globalconfig.SaveGlobalConfig(config); err != nil {
		return fmt.Errorf("failed to save global Loom configuration: %w", err)
	}

	fmt.Printf("Successfully removed %s\n", removedStoreDetails)
	configPath, _ := globalconfig.GetGlobalConfigPath()
	fmt.Printf("Configuration saved to: %s\n", configPath)
	return nil
}

// listStoresAction implements the logic for "loom config list".
func listStoresAction(c *cli.Context) error {
	config, err := globalconfig.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global Loom configuration: %w", err)
	}

	hasPrintedStore := false
	if len(config.Stores) > 0 {
		fmt.Println("Configured Thread Stores:")
		for i, store := range config.Stores {
			fmt.Printf("  Name:     %s\n", store.Name)
			fmt.Printf("  Type:     %s\n", store.Type)
			fmt.Printf("  Path/URL: %s\n", store.Path)
			if i < len(config.Stores)-1 {
				fmt.Println() // Add a blank line between store entries
			}
			hasPrintedStore = true
		}
	}

	// Check for project-specific store
	currentDir, err := os.Getwd()
	if err != nil {
		// If we can't get the current directory, we can't check for a project store.
		// This is unlikely, but we should handle it gracefully.
		// We might not want to error out the whole command for this.
		fmt.Fprintf(os.Stderr, "Warning: Could not determine current directory to check for project store: %v\n", err)
	} else {
		projectStorePath := filepath.Join(currentDir, ".loom")
		if _, err := os.Stat(projectStorePath); err == nil {
			if hasPrintedStore {
				fmt.Println() // Add a blank line if global stores were printed
			}
			fmt.Println("Project Store:")
			fmt.Printf("  Name:     (Project)\n") // Project store doesn't have a configurable name
			fmt.Printf("  Type:     project\n")
			fmt.Printf("  Path/URL: %s\n", projectStorePath)
			hasPrintedStore = true
		}
	}

	if !hasPrintedStore {
		fmt.Println("No configured global stores or project-specific store found.")
	}

	return nil
}
