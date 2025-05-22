# Loom – Scaffolding Tool 🛠️

![License](https://img.shields.io/github/license/nightconcept/loom)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/nightconcept/loom/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/nightconcept/loom/badge.svg)](https://coveralls.io/github/nightconcept/loom)
![GitHub last commit](https://img.shields.io/github/last-commit/nightconcept/loom)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/nightconcept/loom/badge)](https://scorecard.dev/viewer/?uri=github.com/nightconcept/loom)
[![Go Report Card](https://goreportcard.com/badge/github.com/nightconcept/loom)](https://goreportcard.com/report/github.com/nightconcept/loom)

A modern, cross-platform, developer-friendly scaffolding tool for your projects.
Easily generate project structures, files, and configurations from predefined or custom templates.

## Features

- 🚀 **Rapid Project Bootstrapping**: Quickly initialize new projects with consistent structures using templates.
- 📝 **Customizable Blueprints**: Define your own templates and blueprints for any project type or component.
- ⚙️ **Extensible CLI**: Easily add new commands and scaffolding logic to fit your workflow.
- 🛠️ **Cross-Platform**: Works on Linux, macOS, and Windows.

## Installation

You can install `loom` by running the following commands in your terminal. These scripts will download and run the appropriate installer for your system from the `main` branch of the official repository.

### macOS/Linux Install

```sh
curl -LsSf https://raw.githubusercontent.com/nightconcept/loom/main/install.sh | sh
```

### Windows Install

```powershell
powershell -ExecutionPolicy Bypass -c "irm https://raw.githubusercontent.com/nightconcept/loom/main/install.ps1 | iex"
```

## Usage

```sh
loom init                                           # Initialize a new loom.yaml file in the current directory
loom add <thread_name>                              # Add a thread to the project. Syntax: loom add <thread_name> OR loom add <store_name>/<thread_name>
loom remove <thread_name>                           # Remove a thread from the project
loom list                                           # List threads in the project
loom weave [thread_name]                            # Install or re-apply threads to the project. Optionally specify a thread name to weave only that thread.
loom install [thread_name]                          # Alias for weave
loom config                                         # Manage Loom's configuration for thread stores.
```

## Development Requirements

- Go 1.24+
- [Mise](https://mise.jdx.dev/)
- [pre-commit](https://pre-commit.com/)

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
