package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"loom/internal/core/project" // Import the project package

	"gopkg.in/yaml.v3"
)

// normalizeDir ensures directory paths are consistent for loom.yaml keys.
// Returns "./" for empty or "." paths, otherwise ensures forward slashes and a trailing slash.
func normalizeDir(dirPath string) string {
	if dirPath == "" || dirPath == "." {
		return "./"
	}
	slashed := filepath.ToSlash(dirPath)
	if !strings.HasSuffix(slashed, "/") {
		return slashed + "/"
	}
	return slashed
}

// promptUserForOverwriteInWeave prompts the user with a message and expects a yes/no/skip response.
// Duplicated from add.go for now, consider refactoring to a shared utility if more widely needed.
func promptUserForOverwriteInWeave(message string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [Y]es/[N]o/[S]kip [Yes]: ", message)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.ToLower(strings.TrimSpace(input))
		// Corrected condition to handle Enter key (empty input) as "yes"
		switch input {
		case "", "yes", "y":
			return "yes", nil
		case "no", "n":
			return "no", nil
		case "skip", "s":
			return "skip", nil
		}
		// Corrected error message
		fmt.Println("Invalid input. Please enter 'yes', 'no', 'skip', or press Enter for 'yes'.")
	}
}

// Weave re-applies threads to the project.
// If threadNameToWeave is empty, all threads are woven.
// Otherwise, only the specified thread is woven.
func Weave(threadNameToWeave string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	loomConfig, loomConfigPath, err := loadProjectLoomConfig(projectRoot)
	if err != nil {
		return err // Error already contains context
	}

	foundSpecificThread := false
	for i := range loomConfig.Threads {
		currentThread := &loomConfig.Threads[i] // Use pointer to allow modification by helpers

		// If a specific thread is requested, and this isn't it, we might skip.
		// However, processWeavingForThread handles its own skipping logic based on threadNameToWeave.
		// We set foundSpecificThread if the target thread is encountered.
		if threadNameToWeave != "" && currentThread.Name == threadNameToWeave {
			foundSpecificThread = true
		}

		err := processWeavingForThread(currentThread, loomConfig, projectRoot, threadNameToWeave)
		if err != nil {
			// An error from processWeavingForThread is considered significant enough to stop.
			// It would typically be a file system error or critical prompt failure.
			// Minor issues like a single file not found in source are handled within processWeavingForThread by logging.
			return fmt.Errorf("error weaving thread '%s': %w", currentThread.Name, err)
		}

		// If we were weaving a specific thread and we just processed it, we can break the loop.
		if threadNameToWeave != "" && currentThread.Name == threadNameToWeave {
			break
		}
	}

	if threadNameToWeave != "" && !foundSpecificThread {
		return fmt.Errorf("thread '%s' not found in %s", threadNameToWeave, project.YamlFileName)
	}

	if err := saveProjectLoomConfig(loomConfigPath, loomConfig); err != nil {
		return err // Error already contains context
	}

	fmt.Println("Weave operation completed.")
	return nil
}

// loadProjectLoomConfig reads and parses the loom.yaml file from the project root.
func loadProjectLoomConfig(projectRoot string) (*project.LoomConfig, string, error) {
	loomConfigPath := filepath.Join(projectRoot, project.YamlFileName)
	configData, err := os.ReadFile(loomConfigPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read %s: %w", loomConfigPath, err)
	}

	var loomConfig project.LoomConfig
	if err := yaml.Unmarshal(configData, &loomConfig); err != nil {
		// Ensure Files map is initialized for all threads if it's nil
		// This is good practice though yaml unmarshal of an empty map should be fine.
		for i := range loomConfig.Threads {
			if loomConfig.Threads[i].Files == nil {
				loomConfig.Threads[i].Files = make(map[string][]string)
			}
		}
		return nil, "", fmt.Errorf("failed to parse %s: %w", loomConfigPath, err)
	}
	// Ensure Files map is initialized post-unmarshal
	for i := range loomConfig.Threads {
		if loomConfig.Threads[i].Files == nil {
			loomConfig.Threads[i].Files = make(map[string][]string)
		}
	}
	return &loomConfig, loomConfigPath, nil
}

// saveProjectLoomConfig marshals and writes the loomConfig back to the loom.yaml file.
func saveProjectLoomConfig(loomConfigPath string, loomConfig *project.LoomConfig) error {
	// Ensure all threads have non-nil Files maps before saving
	for i := range loomConfig.Threads {
		if loomConfig.Threads[i].Files == nil {
			loomConfig.Threads[i].Files = make(map[string][]string)
		}
	}
	updatedData, err := yaml.Marshal(loomConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal updated %s: %w", project.YamlFileName, err)
	}
	err = os.WriteFile(loomConfigPath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated %s: %w", project.YamlFileName, err)
	}
	return nil
}

// removeFileFromThreadManifest removes a file from the specified thread's manifest in the loomConfig.
func removeFileFromThreadManifest(loomConfig *project.LoomConfig, ownerThreadName string, fileRelToProject string) {
	dir, file := filepath.Split(fileRelToProject)
	normalizedDir := normalizeDir(dir)

	for i := range loomConfig.Threads {
		if loomConfig.Threads[i].Name == ownerThreadName {
			if loomConfig.Threads[i].Files == nil { // Should not happen if initialized properly
				loomConfig.Threads[i].Files = make(map[string][]string)
				return // Nothing to remove
			}
			if filesInDir, ok := loomConfig.Threads[i].Files[normalizedDir]; ok {
				var updatedFilesInDir []string
				for _, f := range filesInDir {
					if f != file {
						updatedFilesInDir = append(updatedFilesInDir, f)
					}
				}
				if len(updatedFilesInDir) > 0 {
					loomConfig.Threads[i].Files[normalizedDir] = updatedFilesInDir
				} else {
					delete(loomConfig.Threads[i].Files, normalizedDir)
					if len(loomConfig.Threads[i].Files) == 0 {
						// Ensure it's an empty map, not nil, for consistency, though delete doesn't make map nil.
						loomConfig.Threads[i].Files = make(map[string][]string)
					}
				}
			}
			return // Found the thread and processed
		}
	}
}

// processFileWeavingParams holds parameters for handleFileWeavingOperation.
type processFileWeavingParams struct {
	projectRoot       string
	threadSourcePath  string // Full path to the _thread directory
	relPathFromSource string // Relative path of the file from _thread dir (e.g., "src/button.js" or "main.go")
	currentThreadName string
	threadNameToWeave string              // Specific thread to weave, or "" for all
	loomConfig        *project.LoomConfig // Pointer to the main config for modifications
}

// fileWeavingAction holds the results of the decision logic for a file operation.
type fileWeavingAction struct {
	shouldWrite bool
}

// handleFileConflictOwnedByOther handles logic when a file exists and is owned by another thread.
// It modifies loomConfig if ownership is taken.
// Returns true if the file should be written by the current thread.
func handleFileConflictOwnedByOther(params *processFileWeavingParams, ownerThreadName string, relDestPathForDisplay string) (bool, error) {
	switch params.threadNameToWeave {
	case "": // Weaving all threads, standard conflict prompt
		fmt.Printf("File '%s' is currently owned by thread '%s'.\n", relDestPathForDisplay, ownerThreadName)
		choice, promptErr := promptUserForOverwriteInWeave(fmt.Sprintf("Thread '%s' wants to overwrite it. Take ownership? ", params.currentThreadName))
		if promptErr != nil {
			return false, fmt.Errorf("failed to get user input for '%s': %w", relDestPathForDisplay, promptErr)
		}
		if choice == "yes" {
			fmt.Printf("Thread '%s' is taking ownership of '%s'.\n", params.currentThreadName, relDestPathForDisplay)
			removeFileFromThreadManifest(params.loomConfig, ownerThreadName, relDestPathForDisplay)
			return true, nil
		}
		fmt.Printf("Skipping file '%s'. Thread '%s' retains ownership.\n", relDestPathForDisplay, ownerThreadName)
		return false, nil
	case params.currentThreadName: // Weaving specific thread, and it's this one, taking from another.
		fmt.Printf("File '%s' is currently owned by thread '%s'.\n", relDestPathForDisplay, ownerThreadName)
		fmt.Printf("Thread '%s' (being specifically woven) is taking ownership of '%s'.\n", params.currentThreadName, relDestPathForDisplay)
		removeFileFromThreadManifest(params.loomConfig, ownerThreadName, relDestPathForDisplay)
		return true, nil
	default: // Weaving specific thread, but this file is owned by another (and not the one being woven). Skip.
		fmt.Printf("Skipping file '%s'. It is owned by '%s', and we are weaving '%s' (not '%s').\n", relDestPathForDisplay, ownerThreadName, params.threadNameToWeave, params.currentThreadName)
		return false, nil
	}
}

// handleFileConflictUnowned handles logic when a file exists but is not owned by any Loom thread.
// Returns true if the file should be written by the current thread.
func handleFileConflictUnowned(params *processFileWeavingParams, relDestPathForDisplay string) (bool, error) {
	switch params.threadNameToWeave {
	case "": // Weaving all, prompt
		fmt.Printf("File '%s' exists but is not currently owned by any Loom thread.\n", relDestPathForDisplay)
		choice, promptErr := promptUserForOverwriteInWeave(fmt.Sprintf("Thread '%s' wants to overwrite it. Take ownership? ", params.currentThreadName))
		if promptErr != nil {
			return false, fmt.Errorf("failed to get user input for '%s': %w", relDestPathForDisplay, promptErr)
		}
		if choice == "yes" {
			fmt.Printf("Thread '%s' is taking ownership of '%s'.\n", params.currentThreadName, relDestPathForDisplay)
			return true, nil
		}
		fmt.Printf("Skipping file '%s'. It remains an unmanaged file.\n", relDestPathForDisplay)
		return false, nil
	case params.currentThreadName: // Weaving specific thread (this one), file is unowned. Take ownership.
		fmt.Printf("File '%s' exists but is not owned. Thread '%s' (being specifically woven) is taking ownership.\n", relDestPathForDisplay, params.currentThreadName)
		return true, nil
	default: // Weaving specific thread (not this one), file is unowned. Skip.
		fmt.Printf("Skipping unowned file '%s'. We are weaving '%s', not '%s'.\n", relDestPathForDisplay, params.threadNameToWeave, params.currentThreadName)
		return false, nil
	}
}

// decideFileWeavingAction determines if a file should be written and handles ownership changes.
func decideFileWeavingAction(params *processFileWeavingParams, destPathInProject string, relDestPathForDisplay string) (fileWeavingAction, error) {
	action := fileWeavingAction{shouldWrite: true} // Default to write, can be overridden

	_, statErr := os.Stat(destPathInProject)
	fileExists := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return fileWeavingAction{}, fmt.Errorf("error checking destination file %s: %w", destPathInProject, statErr)
	}

	if fileExists {
		ownerThreadName, isOwned := params.loomConfig.IsFileOwned(destPathInProject, params.projectRoot)

		if isOwned && ownerThreadName != params.currentThreadName {
			// Owned by another thread
			var err error
			action.shouldWrite, err = handleFileConflictOwnedByOther(params, ownerThreadName, relDestPathForDisplay)
			if err != nil {
				return fileWeavingAction{}, err
			}
		} else if !isOwned {
			// File exists but not owned by any Loom thread
			var err error
			action.shouldWrite, err = handleFileConflictUnowned(params, relDestPathForDisplay)
			if err != nil {
				return fileWeavingAction{}, err
			}
		} else if isOwned && ownerThreadName == params.currentThreadName {
			// File is owned by the current thread. Re-apply.
			fmt.Printf("Re-applying file '%s' from thread '%s'.\n", relDestPathForDisplay, params.currentThreadName)
			action.shouldWrite = true
		}
	} else { // File does not exist at destination.
		if err := os.MkdirAll(filepath.Dir(destPathInProject), os.ModePerm); err != nil {
			return fileWeavingAction{}, fmt.Errorf("failed to create directory for %s: %w", destPathInProject, err)
		}
		fmt.Printf("Creating new file '%s' from thread '%s'.\n", relDestPathForDisplay, params.currentThreadName)
		action.shouldWrite = true
	}
	return action, nil
}

// handleFileWeavingOperation processes a single file for the weave operation.
// Returns true if the file was written, false otherwise, and an error if one occurred.
func handleFileWeavingOperation(params *processFileWeavingParams) (bool, error) {
	pathInThreadSource := filepath.Join(params.threadSourcePath, params.relPathFromSource)
	destPathInProject := filepath.Join(params.projectRoot, params.relPathFromSource)

	sourceInfo, statSourceErr := os.Stat(pathInThreadSource)
	if os.IsNotExist(statSourceErr) {
		fmt.Printf("Warning: Source file %s for thread '%s' not found. Skipping this file.\n", pathInThreadSource, params.currentThreadName)
		return false, nil
	} else if statSourceErr != nil {
		fmt.Printf("Error stating source file %s for thread '%s': %v. Skipping this file.\n", pathInThreadSource, params.currentThreadName, statSourceErr)
		return false, nil // Logged, not a fatal error for the whole weave
	}

	if sourceInfo.IsDir() {
		fmt.Printf("Warning: Source path %s is a directory, expected a file. Skipping.\n", pathInThreadSource)
		return false, nil
	}

	relDestPathForDisplay, _ := filepath.Rel(params.projectRoot, destPathInProject)
	relDestPathForDisplay = filepath.ToSlash(relDestPathForDisplay) // For consistent display and map keys

	action, err := decideFileWeavingAction(params, destPathInProject, relDestPathForDisplay)
	if err != nil {
		return false, err // Propagate errors from decision logic (e.g., prompt failure)
	}

	if action.shouldWrite {
		data, readErr := os.ReadFile(pathInThreadSource)
		if readErr != nil {
			return false, fmt.Errorf("failed to read source file %s: %w", pathInThreadSource, readErr)
		}
		if writeErr := os.WriteFile(destPathInProject, data, sourceInfo.Mode()); writeErr != nil {
			return false, fmt.Errorf("failed to write file %s: %w", destPathInProject, writeErr)
		}
		return true, nil
	}
	return false, nil
}

// determineThreadSourcePath calculates the absolute path to the thread's source directory (_thread).
func determineThreadSourcePath(thread *project.Thread, projectRoot string) string {
	if strings.HasPrefix(thread.Source, "project:") {
		relativePath := strings.TrimPrefix(thread.Source, "project:")
		return filepath.Join(projectRoot, relativePath, "_thread")
	}
	return filepath.Join(projectRoot, ".loom", thread.Name, "_thread")
}

// collectFilesToProcessForWeaving determines the set of files to process for a given thread.
// Returns a map of [normalized directory relative to project] -> [list of filenames].
func collectFilesToProcessForWeaving(
	thread *project.Thread,
	threadSourcePath string,
	projectRoot string, // Not directly used here, but kept for potential future use or consistency
	threadNameToWeave string,
) (map[string][]string, error) {
	filesToProcess := make(map[string][]string)

	// If weaving a specific thread, and it's this thread, use its manifest.
	if threadNameToWeave != "" && threadNameToWeave == thread.Name {
		fmt.Printf("Weaving specific thread '%s'. Will only process files it owns as per %s.\n", thread.Name, project.YamlFileName)
		if len(thread.Files) == 0 {
			fmt.Printf("Thread '%s' does not own any files according to %s. Nothing to weave for this thread.\n", thread.Name, project.YamlFileName)
			return filesToProcess, nil // Empty map, no error
		}
		for dir, filesInDir := range thread.Files {
			normalizedDir := normalizeDir(dir) // Should be normalized already, but ensure.
			filesToProcess[normalizedDir] = append(filesToProcess[normalizedDir], filesInDir...)
		}
	} else if threadNameToWeave == "" { // Weaving all threads - walk the source directory.
		walkErr := filepath.Walk(threadSourcePath, func(path string, info os.FileInfo, walkErrInner error) error {
			if walkErrInner != nil {
				return walkErrInner // Propagate errors from previous WalkFunc calls
			}
			if info.IsDir() {
				return nil // Skip directories
			}
			relPathFromSourceDir, err := filepath.Rel(threadSourcePath, path)
			if err != nil {
				// This error is critical for this file, wrap it with more context.
				return fmt.Errorf("failed to get relative path for %s (base: %s): %w", path, threadSourcePath, err)
			}
			destDirRelToProject, fileName := filepath.Split(relPathFromSourceDir)
			destDirNorm := normalizeDir(destDirRelToProject)
			filesToProcess[destDirNorm] = append(filesToProcess[destDirNorm], fileName)
			return nil
		})
		if walkErr != nil {
			// Error during walk is significant for this thread's processing.
			return nil, fmt.Errorf("error walking source directory for thread '%s' (%s): %w", thread.Name, threadSourcePath, walkErr)
		}
	}
	// If threadNameToWeave is specific but NOT this thread, filesToProcess remains empty, which is correct.
	return filesToProcess, nil
}

// processWeavingForThread handles the weaving logic for a single thread.
func processWeavingForThread(
	thread *project.Thread, // Pointer to the thread in loomConfig
	loomConfig *project.LoomConfig,
	projectRoot string,
	threadNameToWeave string,
) error {
	// If weaving a specific thread, only proceed if this IS the thread.
	if threadNameToWeave != "" && thread.Name != threadNameToWeave {
		return nil // Not the target thread for a specific weave.
	}

	threadSourcePath := determineThreadSourcePath(thread, projectRoot)
	if _, statErr := os.Stat(threadSourcePath); os.IsNotExist(statErr) {
		fmt.Printf("Thread source directory not found for thread '%s': %s. Skipping this thread.\n", thread.Name, threadSourcePath)
		return nil // Skip this thread, not a fatal error for the whole weave operation.
	}

	// If we are here, either weaving all, or (weaving specific AND this is the target thread).
	fmt.Printf("Weaving thread '%s' from %s...\n", thread.Name, threadSourcePath)

	filesToProcess, err := collectFilesToProcessForWeaving(thread, threadSourcePath, projectRoot, threadNameToWeave)
	if err != nil {
		// Error already has context from collectFilesToProcessForWeaving.
		fmt.Printf("Failed to collect files for thread '%s': %v. Skipping this thread.\n", thread.Name, err)
		return nil // Skip this thread.
	}

	// collectFilesToProcessForWeaving already prints a message if thread.Files is empty for a specific weave.
	// if threadNameToWeave != "" && thread.Name == threadNameToWeave && len(filesToProcess) == 0 {
	// 	// Message already printed by collectFilesToProcessForWeaving if thread.Files was empty.
	// 	// If filesToProcess is empty for other reasons (e.g. manifest points to non-existent files),
	// 	// the loop below will simply not run.
	// 	// No explicit message needed here if collectFilesToProcess already informed.
	// }

	filesActuallyWrittenByThisThread := make(map[string][]string)

	for dirToProcess, filesInDirToProcess := range filesToProcess { // dirToProcess is normalized
		for _, fileToProcess := range filesInDirToProcess { // fileToProcess is just filename
			relPathFromFileSource := filepath.Join(dirToProcess, fileToProcess) // Reconstruct relative path

			params := processFileWeavingParams{
				projectRoot:       projectRoot,
				threadSourcePath:  threadSourcePath,
				relPathFromSource: relPathFromFileSource,
				currentThreadName: thread.Name,
				threadNameToWeave: threadNameToWeave,
				loomConfig:        loomConfig,
			}

			fileWasWritten, opErr := handleFileWeavingOperation(&params)
			if opErr != nil {
				// Propagate error if file operation failed critically
				return fmt.Errorf("processing file '%s' for thread '%s': %w", relPathFromFileSource, thread.Name, opErr)
			}

			if fileWasWritten {
				// dirToProcess is already normalized (e.g., "./" or "src/components/")
				filesActuallyWrittenByThisThread[dirToProcess] = append(filesActuallyWrittenByThisThread[dirToProcess], fileToProcess)
			}
		}
	}

	// Update the thread's manifest in loomConfig with files it actually wrote/owns.
	// This is critical: thread is a pointer, so loomConfig is directly updated.
	thread.Files = filesActuallyWrittenByThisThread
	if thread.Files == nil { // Should be handled by make(), but defensive.
		thread.Files = make(map[string][]string)
	}

	return nil
}
