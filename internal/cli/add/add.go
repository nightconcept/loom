package add

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"loom/internal/core/globalconfig" // Import the globalconfig package
	"loom/internal/core/project"      // Import the project package

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// Remove local LoomConfig and Thread structs, use project package versions

// parseAddArgs parses the command line arguments for the add command.
// It returns the target store name, thread name, and an error if parsing fails.
func parseAddArgs(fullThreadArg string) (string, string, error) {
	if fullThreadArg == "" {
		return "", "", fmt.Errorf("thread name or store/thread is required")
	}

	var targetStoreName string
	var threadName string
	parts := strings.SplitN(fullThreadArg, "/", 2)
	if len(parts) == 2 {
		targetStoreName = parts[0]
		threadName = parts[1]
		if targetStoreName == "" || threadName == "" {
			return "", "", fmt.Errorf("invalid format for store/thread: '%s'. Both store name and thread name must be specified", fullThreadArg)
		}
	} else {
		threadName = fullThreadArg
	}
	return targetStoreName, threadName, nil
}

// loadProjectLoomConfig loads the loom.yaml configuration from the project root.
// If the file doesn't exist, it initializes an empty configuration.
// It returns the loaded config, the path to the config file, and an error if any.
func loadProjectLoomConfig(projectRoot string) (project.LoomConfig, string, error) {
	loomConfigPath := filepath.Join(projectRoot, project.YamlFileName)
	var loomConfig project.LoomConfig
	configData, err := os.ReadFile(loomConfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return loomConfig, loomConfigPath, fmt.Errorf("failed to read %s: %w", project.YamlFileName, err)
		}
		// Initialize empty config if loom.yaml doesn't exist
		loomConfig = project.LoomConfig{Version: "1", Threads: []project.Thread{}}
	} else {
		err = yaml.Unmarshal(configData, &loomConfig)
		if err != nil {
			return loomConfig, loomConfigPath, fmt.Errorf("failed to parse %s: %w", project.YamlFileName, err)
		}
	}
	return loomConfig, loomConfigPath, nil
}

// findThreadInProjectStore searches for a thread in the project's .loom directory.
// It returns the thread path, thread source, a boolean indicating if found, and an error.
func findThreadInProjectStore(projectRoot, threadName string) (string, string, bool, error) {
	projectThreadPath := filepath.Join(projectRoot, ".loom", threadName, "_thread")
	_, err := os.Stat(projectThreadPath)
	if err == nil {
		threadSource := fmt.Sprintf("project:.loom/%s", threadName)
		return projectThreadPath, threadSource, true, nil
	}
	if os.IsNotExist(err) {
		return "", "", false, nil
	}
	return "", "", false, err
}

// findThreadInLocalStores searches for a thread in the configured local PC stores.
// It returns the thread path, thread source, a boolean indicating if found, and an error.
func findThreadInLocalStores(targetStoreName, threadName string, gConf *globalconfig.GlobalLoomConfig) (string, string, bool, error) {
	for _, store := range gConf.Stores {
		if targetStoreName != "" && store.Name != targetStoreName {
			continue
		}
		if store.Type == "local" {
			potentialThreadPath := filepath.Join(store.Path, threadName, "_thread")
			fileInfo, err := os.Stat(potentialThreadPath)
			if err == nil {
				if fileInfo.IsDir() {
					return potentialThreadPath, store.Name, true, nil
				} else {
					// If the path exists but is not a directory, it's a malformed thread.
					return "", "", false, fmt.Errorf("thread path '%s' in store '%s' is a file, not a directory", potentialThreadPath, store.Name)
				}
			} else if !os.IsNotExist(err) {
				return "", "", false, fmt.Errorf("error accessing thread '%s' in store '%s' (%s): %w", threadName, store.Name, potentialThreadPath, err)
			}
		}
	}
	return "", "", false, nil
}

// handleThreadSearch orchestrates the search for a thread, first in the project store, then in local stores.
func handleThreadSearch(projectRoot, targetStoreName, threadName string) (string, string, error) {
	// Try project store first only if no specific store is targeted
	if targetStoreName == "" {
		threadPath, threadSource, foundInProject, err := findThreadInProjectStore(projectRoot, threadName)
		if err != nil {
			return "", "", fmt.Errorf("error searching in project store: %w", err)
		}
		if foundInProject {
			return threadPath, threadSource, nil
		}
	}

	gConf, err := globalconfig.LoadGlobalConfig()
	if err != nil {
		return "", "", fmt.Errorf("failed to load global loom configuration: %w", err)
	}

	threadPath, threadSource, foundInLocal, err := findThreadInLocalStores(targetStoreName, threadName, gConf)
	if err != nil {
		return "", "", fmt.Errorf("error searching in local stores: %w", err)
	}

	if foundInLocal {
		return threadPath, threadSource, nil
	}

	// Error messages if not found
	if targetStoreName != "" {
		storeExists := false
		for _, store := range gConf.Stores {
			if store.Name == targetStoreName {
				storeExists = true
				break
			}
		}
		if !storeExists {
			return "", "", fmt.Errorf("specified store '%s' not found in global configuration", targetStoreName)
		}
		return "", "", fmt.Errorf("thread '%s' not found in specified store '%s'", threadName, targetStoreName)
	}
	return "", "", fmt.Errorf("thread '%s' not found in project's .loom folder or any configured local PC stores", threadName)
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a thread to the project. Syntax: loom add <thread_name> OR loom add <store_name>/<thread_name>",
		Action: func(c *cli.Context) error {
			fullThreadArg := c.Args().First()
			targetStoreName, threadName, err := parseAddArgs(fullThreadArg)
			if err != nil {
				return err
			}

			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %v", err)
			}

			loomConfig, loomConfigPath, err := loadProjectLoomConfig(projectRoot)
			if err != nil {
				return err // Error already formatted by loadProjectLoomConfig
			}

			threadPath, threadSource, err := handleThreadSearch(projectRoot, targetStoreName, threadName)
			if err != nil {
				return err
			}
			// Safeguard, though handleThreadSearch should error out if not found.
			if threadPath == "" {
				return fmt.Errorf("thread '%s' not found after search (unexpected)", fullThreadArg)
			}

			filesByDir, err := copyDir(threadPath, projectRoot, threadName, threadSource, &loomConfig)
			if err != nil {
				return fmt.Errorf("failed to copy thread files: %v", err)
			}

			err = updateLoomConfig(loomConfigPath, threadName, threadSource, filesByDir, &loomConfig)
			if err != nil {
				return fmt.Errorf("failed to update %s: %v", project.YamlFileName, err)
			}

			fmt.Printf("Thread '%s' added successfully from %s\n", fullThreadArg, threadSource)
			return nil
		},
	}
}

// copyDir recursively copies files from src to dest and tracks the files by their directory structure
// relative to the project root. It returns a map where keys are directory paths (with trailing slash)
// It now includes conflict resolution.
func copyDir(src string, dest string, currentThreadName string, displayCurrentThreadSource string, loomConfig *project.LoomConfig) (map[string][]string, error) {
	// We need to track the original project root to calculate relative paths correctly
	// Ensure the base destination directory exists
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create base destination directory %s: %w", dest, err)
	}
	return copyDirWithBasePath(src, dest, dest, currentThreadName, displayCurrentThreadSource, loomConfig)
}

// handleExistingFileConflict checks if a file at destPath conflicts with the thread being added.
// It prompts the user if necessary and returns true if the file should be overwritten,
// false if it should be skipped, and an error if a critical issue occurs (e.g., stat fails unexpectedly, prompt fails).
func handleExistingFileConflict(destPath, baseProjectPath, displayCurrentThreadSource string, loomConfig *project.LoomConfig) (bool, error) {
	// Check if the file already exists in the destination
	_, statErr := os.Stat(destPath)
	if statErr == nil { // File exists
		ownerThreadNameFromConfig, isOwned := loomConfig.IsFileOwned(destPath, baseProjectPath)
		relDestPath, err := filepath.Rel(baseProjectPath, destPath)
		if err != nil {
			// Treat failure to determine relative path as a fatal error.
			// This makes the error handling stricter for path resolution issues.
			return false, fmt.Errorf("failed to determine relative path for '%s' from base '%s': %w", destPath, baseProjectPath, err)
		}

		if isOwned {
			var ownerThreadSourceFromConfig string
			for _, t := range loomConfig.Threads {
				if t.Name == ownerThreadNameFromConfig {
					ownerThreadSourceFromConfig = t.Source
					break
				}
			}
			if ownerThreadSourceFromConfig == "" {
				ownerThreadSourceFromConfig = ownerThreadNameFromConfig
			}

			if ownerThreadSourceFromConfig == displayCurrentThreadSource {
				return true, nil
			}
			fmt.Printf("File '%s' is currently owned by thread '%s'.\n", relDestPath, ownerThreadSourceFromConfig)
			choice, promptErr := promptUserForOverwrite(fmt.Sprintf("Do you want thread '%s' to take ownership of '%s' and overwrite it?", displayCurrentThreadSource, relDestPath))
			if promptErr != nil {
				return false, fmt.Errorf("failed to get user input for %s: %w", relDestPath, promptErr)
			}

			if choice == "yes" {
				fmt.Printf("Thread '%s' is taking ownership of '%s'.\n", displayCurrentThreadSource, relDestPath)
				return true, nil
			}
			fmt.Printf("Skipping file '%s'. Thread '%s' retains ownership.\n", relDestPath, ownerThreadSourceFromConfig)
			return false, nil
		}
		fmt.Printf("File '%s' exists but is not currently owned by any Loom thread.\n", relDestPath)
		choice, promptErr := promptUserForOverwrite(fmt.Sprintf("Do you want thread '%s' to take ownership of '%s' and overwrite it?", displayCurrentThreadSource, relDestPath))
		if promptErr != nil {
			return false, fmt.Errorf("failed to get user input for %s: %w", relDestPath, promptErr)
		}
		if choice == "yes" {
			fmt.Printf("Thread '%s' is taking ownership of '%s'.\n", displayCurrentThreadSource, relDestPath)
			return true, nil
		}
		fmt.Printf("Skipping file '%s'. It remains an unmanaged file or user version.\n", relDestPath)
		return false, nil
	} else if os.IsNotExist(statErr) {
		return true, nil
	}
	return false, fmt.Errorf("failed to stat destination path %s: %w", destPath, statErr)
}

// _processFileCopy handles the logic for copying a single file, including conflict resolution.
// It returns the relative directory path (e.g., "./", "subdir/") and the file name if the file was successfully copied,
// or empty strings and potentially an error if skipped or an error occurred.
func _processFileCopy(srcPath, destPath, baseProjectPath, currentThreadName, displayCurrentThreadSource string, srcFileInfo os.FileInfo, loomConfig *project.LoomConfig) (string, string, error) {
	destFileDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destFileDir, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("failed to create parent directory for destination file %s: %w", destPath, err)
	}

	shouldOverwrite, conflictErr := handleExistingFileConflict(destPath, baseProjectPath, displayCurrentThreadSource, loomConfig)
	if conflictErr != nil {
		return "", "", conflictErr
	}

	if !shouldOverwrite {
		return "", "", nil // Skipped
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read source file %s: %w", srcPath, err)
	}
	err = os.WriteFile(destPath, data, srcFileInfo.Mode())
	if err != nil {
		return "", "", fmt.Errorf("failed to write destination file %s: %w", destPath, err)
	}

	relDir := "./"
	if destFileDir != baseProjectPath {
		relPathCurrent, err := filepath.Rel(baseProjectPath, destFileDir)
		if err != nil {
			return "", "", fmt.Errorf("failed to get relative path for %s from %s: %w", destFileDir, baseProjectPath, err)
		}
		if relPathCurrent == "." {
			relDir = "./"
		} else {
			relDir = filepath.ToSlash(relPathCurrent) + "/"
		}
	}
	return relDir, srcFileInfo.Name(), nil
}

// copyDirWithBasePath is an internal helper that maintains the base project path during recursion
// It now includes conflict resolution.
func copyDirWithBasePath(src string, dest string, baseProjectPath string, currentThreadName string, displayCurrentThreadSource string, loomConfig *project.LoomConfig) (map[string][]string, error) {
	filesByDir := make(map[string][]string)
	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		srcFileInfo, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get FileInfo for source %s: %w", srcPath, err)
		}

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, srcFileInfo.Mode()); err != nil {
				return nil, fmt.Errorf("failed to create destination directory %s: %w", destPath, err)
			}

			subFilesByDir, err := copyDirWithBasePath(srcPath, destPath, baseProjectPath, currentThreadName, displayCurrentThreadSource, loomConfig)
			if err != nil {
				return nil, err // Propagate error from recursive call
			}
			for dir, files := range subFilesByDir {
				filesByDir[dir] = append(filesByDir[dir], files...)
			}
		} else {
			// Process file using the new helper function
			relDir, fileName, err := _processFileCopy(srcPath, destPath, baseProjectPath, currentThreadName, displayCurrentThreadSource, srcFileInfo, loomConfig)
			if err != nil {
				return nil, err // Propagate error from file processing
			}
			if fileName != "" { // If fileName is not empty, it means the file was copied
				filesByDir[relDir] = append(filesByDir[relDir], fileName)
			}
		}
	}
	return filesByDir, nil
}

// promptUserForOverwrite prompts the user with a message and expects a yes/no/skip response.
func promptUserForOverwrite(message string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [Y]es/[N]o/[S]kip [Yes]: ", message)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.ToLower(strings.TrimSpace(input))
		switch input {
		case "", "yes", "y":
			return "yes", nil
		case "no", "n":
			return "no", nil
		case "skip", "s":
			return "skip", nil
		}
		fmt.Println("Invalid input. Please enter 'yes', 'no', 'skip', or press Enter for 'yes'.")
	}
}

// removeFileFromOtherThreads removes a specific file from all threads except the currentThreadName.
// It modifies the config.Threads in place.
func removeFileFromOtherThreads(config *project.LoomConfig, currentThreadName, dirToRemove, fileToRemove string) {
	for i, otherThread := range config.Threads {
		if otherThread.Name == currentThreadName {
			continue
		}
		if otherThread.Files == nil {
			continue
		}

		if filesInDir, dirExists := otherThread.Files[dirToRemove]; dirExists {
			var updatedFilesInDir []string
			fileWasRemoved := false
			for _, existingFile := range filesInDir {
				if existingFile == fileToRemove {
					fileWasRemoved = true
				} else {
					updatedFilesInDir = append(updatedFilesInDir, existingFile)
				}
			}

			if fileWasRemoved {
				if len(updatedFilesInDir) == 0 {
					delete(config.Threads[i].Files, dirToRemove)
					// If the Files map itself becomes empty, nil it out for cleaner YAML
					if len(config.Threads[i].Files) == 0 {
						config.Threads[i].Files = nil
					}
				} else {
					config.Threads[i].Files[dirToRemove] = updatedFilesInDir
				}
			}
		}
	}
}

// updateLoomConfig updates the loom.yaml configuration by removing added files from other threads
// and then adding or updating the current thread's information.
func updateLoomConfig(configPath string, threadName string, source string, filesByDir map[string][]string, config *project.LoomConfig) error {
	// Remove the files being added from any other threads
	for dir, files := range filesByDir {
		for _, file := range files {
			removeFileFromOtherThreads(config, threadName, dir, file)
		}
	}

	// Ensure Threads slice is initialized
	if config.Threads == nil {
		config.Threads = []project.Thread{}
	}

	// Find if the thread already exists to update it, otherwise add a new one
	foundThreadIndex := -1
	for i, th := range config.Threads {
		if th.Name == threadName {
			foundThreadIndex = i
			break
		}
	}

	if foundThreadIndex != -1 {
		// Update existing thread
		config.Threads[foundThreadIndex].Source = source
		if config.Threads[foundThreadIndex].Files == nil {
			config.Threads[foundThreadIndex].Files = make(map[string][]string)
		}
		// Update the files for the specified directories in the current thread
		for dir, files := range filesByDir {
			config.Threads[foundThreadIndex].Files[dir] = files
		}
	} else {
		// Add new thread
		newThread := project.Thread{
			Name:   threadName,
			Source: source,
			Files:  filesByDir,
		}
		config.Threads = append(config.Threads, newThread)
	}

	// Marshal and write the updated configuration
	updatedData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, updatedData, os.ModePerm)
}
