// Package e2e contains end-to-end tests for the Loom CLI tool.
package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestE2E runs the E2E test suite for the Loom CLI tool.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loom E2E Suite")
}
