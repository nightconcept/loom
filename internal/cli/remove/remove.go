// filepath: c:\Users\solivand\git\loom\internal\cli\remove\remove.go
// Title: Remove Command Implementation
// Purpose: Implements the `loom remove <thread_name>` command to remove a thread and its files from the project.

package remove

import (
	"fmt"
	"os"
	"path/filepath"

	"loom/internal/core/project" // Import the project package

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// Remove local LoomConfig and Thread structs, use project package versions

// Command returns the cli.Command for the "remove" command.
func Command() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a thread from the project",
		ArgsUsage: "<thread_name>",
		Action: func(c *cli.Context) error {
			threadName := c.Args().First()
			if threadName == "" {
				return fmt.Errorf("thread name is required")
			}
			if threadName == "*" {
				return removeAllThreadsAction()
			}
			return removeThreadAction(threadName)
		},
	}
}

// readLoomConfig reads and parses the loom.yaml file from the project root.
func readLoomConfig(projectRoot string) (*project.LoomConfig, error) {
	loomConfigPath := filepath.Join(projectRoot, project.YamlFileName)
	data, err := os.ReadFile(loomConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", project.YamlFileName, err)
	}

	var config project.LoomConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", project.YamlFileName, err)
	}
	return &config, nil
}

// findThreadInConfig searches for a thread by name in the LoomConfig.
// It returns the thread and a new slice of threads with the found thread removed.
// If the thread is not found, it returns an error.
func findThreadInConfig(config *project.LoomConfig, threadName string) (project.Thread, []project.Thread, error) {
	var threadToRemove project.Thread
	var updatedThreads []project.Thread
	threadFound := false

	for _, thread := range config.Threads {
		if thread.Name == threadName {
			threadFound = true
			threadToRemove = thread
		} else {
			updatedThreads = append(updatedThreads, thread)
		}
	}

	if !threadFound {
		return project.Thread{}, nil, fmt.Errorf("thread '%s' not found in %s", threadName, project.YamlFileName)
	}
	return threadToRemove, updatedThreads, nil
}

// removeThreadFiles removes files associated with a given thread and attempts to clean up empty directories.
func removeThreadFiles(thread project.Thread, projectRoot string, threadName string) {
	if thread.Files == nil {
		return
	}
	for dir, files := range thread.Files {
		for _, file := range files {
			filePath := filepath.Join(projectRoot, dir, file)
			err := os.Remove(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("Warning: File %s listed in %s for thread '%s' not found, skipping.\n", filePath, project.YamlFileName, threadName)
				} else {
					fmt.Printf("Warning: Failed to remove file %s: %v\n", filePath, err)
				}
			} else {
				fmt.Printf("Removed file: %s\n", filePath)
			}
		}
		// Attempt to remove the directory if it's empty
		dirPath := filepath.Join(projectRoot, dir)
		if dirPath != projectRoot { // Don't try to remove the project root
			entries, readDirErr := os.ReadDir(dirPath)
			if readDirErr == nil && len(entries) == 0 {
				err := os.Remove(dirPath)
				if err != nil {
					// Ignore error if directory is not empty or other issues
					// fmt.Printf("Warning: Failed to remove directory %s: %v\n", dirPath, err)
				} else {
					fmt.Printf("Removed empty directory: %s\n", dirPath)
				}
			}
		}
	}
}

// updateLoomConfig marshals the updated configuration and writes it back to loom.yaml.
func updateLoomConfig(projectRoot string, config *project.LoomConfig) error {
	loomConfigPath := filepath.Join(projectRoot, project.YamlFileName)
	updatedData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", project.YamlFileName, err)
	}

	err = os.WriteFile(loomConfigPath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated %s: %w", project.YamlFileName, err)
	}
	return nil
}

// removeThreadAction handles the logic for removing a thread.
func removeThreadAction(threadName string) error {
	projectRoot, err := os.Getwd() // Assuming loom commands run from project root
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	config, err := readLoomConfig(projectRoot)
	if err != nil {
		return err // Error already contains context
	}

	threadToRemove, updatedThreads, err := findThreadInConfig(config, threadName)
	if err != nil {
		return err // Error already contains context
	}

	removeThreadFiles(threadToRemove, projectRoot, threadName)

	config.Threads = updatedThreads
	if err := updateLoomConfig(projectRoot, config); err != nil {
		return err // Error already contains context
	}

	fmt.Printf("Thread '%s' removed successfully.\n", threadName)
	return nil
}

// removeThreadFilesAndCollectDirs processes a single thread's files for removal
// and collects directories that might become empty.
func removeThreadFilesAndCollectDirs(thread project.Thread, projectRoot string, directoriesToRemove map[string]bool) {
	fmt.Printf("Processing thread: %s\n", thread.Name)
	if thread.Files != nil {
		for dir, files := range thread.Files {
			actualDir := filepath.Join(projectRoot, dir)
			directoriesToRemove[actualDir] = true // Mark directory for potential removal
			for _, file := range files {
				filePath := filepath.Join(actualDir, file)
				err := os.Remove(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						fmt.Printf("Warning: File %s listed for thread '%s' not found, skipping.\n", filePath, thread.Name)
					} else {
						fmt.Printf("Warning: Failed to remove file %s: %v\n", filePath, err)
					}
				} else {
					fmt.Printf("Removed file: %s\n", filePath)
				}
			}
		}
	}
}

// removeEmptyDirectories attempts to remove directories that are now empty.
func removeEmptyDirectories(projectRoot string, directoriesToRemove map[string]bool) {
	for dirPath := range directoriesToRemove {
		if dirPath != projectRoot { // Don't try to remove the project root
			entries, readDirErr := os.ReadDir(dirPath)
			if readDirErr == nil && len(entries) == 0 {
				err := os.Remove(dirPath)
				if err != nil {
					// fmt.Printf("Warning: Failed to remove directory %s: %v\n", dirPath, err)
				} else {
					fmt.Printf("Removed empty directory: %s\n", dirPath)
				}
			}
		}
	}
}

// removeAllThreadsAction handles the logic for removing all threads.
func removeAllThreadsAction() error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	loomConfigPath := filepath.Join(projectRoot, project.YamlFileName)

	data, err := os.ReadFile(loomConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s not found. No threads to remove.\n", project.YamlFileName)
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", project.YamlFileName, err)
	}

	var config project.LoomConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", project.YamlFileName, err)
	}

	if len(config.Threads) == 0 {
		fmt.Println("No threads found in loom.yaml to remove.")
		return nil
	}

	fmt.Println("Removing all threads and their files...")

	directoriesToRemove := make(map[string]bool)

	for _, thread := range config.Threads {
		removeThreadFilesAndCollectDirs(thread, projectRoot, directoriesToRemove)
	}

	removeEmptyDirectories(projectRoot, directoriesToRemove)

	// Clear threads from config
	config.Threads = []project.Thread{}
	updatedData, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", project.YamlFileName, err)
	}

	err = os.WriteFile(loomConfigPath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated %s: %w", project.YamlFileName, err)
	}

	fmt.Printf("All threads removed and %s cleared successfully.\n", project.YamlFileName)
	return nil
}
