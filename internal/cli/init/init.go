// Package init provides the CLI command for initializing a new loom project
package init

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"loom/internal/core/project"
)

// Command returns the init command for the CLI
func Command() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new loom.yaml file in the current directory",
		Action: func(c *cli.Context) error {
			return handleInit(c)
		},
	}
}

// handleInit handles the init command
func handleInit(c *cli.Context) error {
	// Initialize the project
	err := project.InitProject()
	if err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	fmt.Println("Initialized empty Loom project with loom.yaml")
	return nil
}
