package main

import (
	"log"
	"os"

	addCmd "loom/internal/cli/add"
	configCmd "loom/internal/cli/config" // Added for config command
	initCmd "loom/internal/cli/init"
	listCmd "loom/internal/cli/list"
	removeCmd "loom/internal/cli/remove"
	weaveCmd "loom/internal/cli/weave"

	"github.com/urfave/cli/v2"
)

// VERSION is the current version of the Loom CLI
const VERSION = "0.1.0"

func main() {
	app := &cli.App{
		Name:    "loom",
		Version: VERSION,
		Usage:   "A command-line interface (CLI) tool for rapid project scaffolding",
		Authors: []*cli.Author{
			{
				Name: "Loom Team",
			},
		},
		Commands: []*cli.Command{
			initCmd.Command(),
			addCmd.Command(),
			removeCmd.Command(),
			{
				Name:  "list",
				Usage: "List threads in the project",
				Action: func(c *cli.Context) error {
					listCmd.ExecuteListCommand()
					return nil
				},
			},
			{
				Name:    "weave",
				Aliases: []string{"install"},
				Usage:   "Install or re-apply threads to the project. Optionally specify a thread name to weave only that thread.",
				Action: func(c *cli.Context) error {
					threadName := "" // Default to empty, meaning all threads
					if c.Args().Len() > 0 {
						threadName = c.Args().First()
					}
					if err := weaveCmd.Weave(threadName); err != nil {
						log.Printf("Error during weave: %v", err)
						return err
					}
					return nil
				},
			},
			configCmd.Command(), // Added the config command
			{
				Name:  "version",
				Usage: "Print the version number of Loom CLI",
				Action: func(c *cli.Context) error {
					println(VERSION)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
