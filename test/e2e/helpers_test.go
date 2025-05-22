// Package e2e contains helper functions for end-to-end tests.
package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// CreateTempDir creates a temporary directory for testing and returns its path.
// The directory will be removed when the test completes.
func CreateTempDir() string {
	tempDir, err := os.MkdirTemp("", "loom-e2e-test-")
	Expect(err).NotTo(HaveOccurred())

	// Register cleanup using DeferCleanup
	DeferCleanup(os.RemoveAll, tempDir)

	return tempDir
}

// CreateTempFile creates a temporary file with the given content and returns its path.
// The file will be removed when the test completes.
func CreateTempFile(dir, name, content string) string {
	filePath := filepath.Join(dir, name)
	// Ensure the parent directory exists
	parentDir := filepath.Dir(filePath)
	err := os.MkdirAll(parentDir, 0755) // 0755 is a common permission for directories
	Expect(err).NotTo(HaveOccurred())

	err = os.WriteFile(filePath, []byte(content), 0644)
	Expect(err).NotTo(HaveOccurred())

	return filePath
}

// InitProjectLoomFile creates a basic loom.yaml file in the specified directory.
func InitProjectLoomFile(dir string) string {
	content := `version: "1"
threads: []
`
	return CreateTempFile(dir, "loom.yaml", content)
}

// CreateTestThreadFile creates a dummy thread file for testing.
// It takes the base directory (either project's .loom or a global store path),
// the thread name, and the content for the thread file.
func CreateTestThreadFile(baseDir, threadName, content string) string {
	threadFilePath := filepath.Join(baseDir, threadName+".md")
	err := os.MkdirAll(filepath.Dir(threadFilePath), 0755)
	if err != nil {
		panic(fmt.Sprintf("failed to create directory for test thread file: %v", err))
	}
	err = os.WriteFile(threadFilePath, []byte(content), 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to write test thread file: %v", err))
	}
	return threadFilePath
}

// CreateGlobalStoreDir creates a store directory within the global loom directory.
func CreateGlobalStoreDir(globalLoomDir, storeName string) string {
	storePath := filepath.Join(globalLoomDir, "stores", storeName)
	err := os.MkdirAll(storePath, 0755)
	if err != nil {
		panic(fmt.Sprintf("failed to create global store directory: %v", err))
	}
	return storePath
}
