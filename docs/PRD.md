# Loom Scaffolding Tool - Product Requirements Document

## 1. Introduction

Loom (CLI command: `loom`) is a command-line interface (CLI) tool, built in Go, designed for rapid project scaffolding. It enables developers to quickly assemble new projects by combining pre-defined templates called "threads." These threads, which are essentially directories of code and files, can be sourced from local storage, project-specific directories, or remote GitHub repositories. Loom prioritizes speed and simplicity in project setup, offering a straightforward way to manage and "weave" these threads together, with interactive conflict resolution for overlapping files.

## 2. Core Features

- **Thread Management:**
    - Add "threads" (pre-built templates/directories) to a project.
    - Source threads from local file paths or GitHub repositories.
- **Conflict Resolution:**
    - Interactively prompt the user to choose which thread "owns" a file when file collisions occur between threads being added.
- **Project Manifest:**
    - Maintain a `loom.yaml` file within the project to track all included threads, their sources, and conflict resolution decisions.
- **Thread Removal:**
    - Remove specific threads from a project.
    - Option to remove all threads from a project.
- **Flexible Thread Stores:**
    - **Local PC Store:** Store and access threads from a user-defined global directory on the local machine.
    - **Project Store:** Store and access threads from a `.loom/` directory within the project itself.
    - **GitHub Store:** Access threads directly from GitHub repositories.
    - Opinionated default locations for local PC stores will be suggested/supported.
- **Simplicity and Speed:**
    - Focus on extremely fast project setup.
    - Updates are brute-force: Loom replaces files based on the current state of the thread source. It does not manage versions with hashes or complex dependency trees.
    - Intentionally simpler and with fewer guardrails than a full package management system.
- **Future Considerations:**
    - Templating language support for dynamic folder names and file content based on PC/user variables (see also Section 4.2 for `template_variables` in `config.yml`).

## 3. Project Folder Structure (Using Loom)

A typical project utilizing Loom might have the following structure:

```
- loom.yaml              # Loom manifest file (tracks threads, sources, conflicts)
- .loom/                 # (Optional) Project-specific thread store
- src/                   # Example project source code, partially or fully populated by threads
- docs/                  # Example project documentation, potentially from threads
- scripts/               # Example project scripts, potentially from threads
- README.md              # Project README, potentially initiated by a thread
- *(other project files and directories populated by threads)*
```

## 4. File & Directory Descriptions

### 4.1. `loom.yaml` File

The `loom.yaml` file is a project-specific manifest that Loom uses to keep track of the threads incorporated into the project. All paths defined within this file are relative to the project root.

**Potential Structure (YAML):**

```yaml
# loom.yaml
version: 1
threads:
  - name: "core-backend-structure"
    source: "github:my-org/loom-threads/core-backend"
    files: # Files this thread won, grouped by directory (paths relative to project root)
      "src/config/":
        - "base.go"
        - "database.go"
      "src/core/":
        - "server.go"
      "./": # Represents the project root
        - ".gitignore"
        - "README.md"
  - name: "logging-module"
    source: "local:/Users/username/.loom-store/common-threads/logging"
    files:
      "src/utils/logging/":
        - "logger.go"
        - "formatter.go"
  - name: "project-specific-api"
    source: "project:.loom/custom-api-thread"
    files:
      "api/":
        - "routes.go"
        - "handlers.go"
# Further details to be refined; `last_weave_timestamp` is a potential future improvement.
```

- **version (integer):** Version of the `loom.yaml` file format.
- **threads (list):** A list of thread objects.
    - **name (string):** A unique name for the thread within the project.
    - **source (string):** The URI or path indicating the thread's origin (e.g., `github:user/repo/path/to/thread`, `local:/path/to/thread`, `project:.loom/path/to/thread`).
    - **files (map, optional):** A map where keys are directory paths (strings, relative to the project root, ending with a `/`) and values are lists of filenames (strings) within that directory that this thread "owns" as a result of conflict resolution. A key of `"./"` indicates files in the project root.

### 4.2. Thread `config.yml`

Located within a thread's directory (e.g., `thread_name/config.yml`).

- **Purpose:** Stores metadata about the thread and definitions for future templating capabilities.
    - **Current uses:**
        - Storing metadata about the thread (description, author, version, license).
    - **Future considerations:**
        - Defining variables for templating features.
        - Specifying dependencies on other threads (though current philosophy is against this).

**Structure (YAML):**

```yaml
# thread_name/config.yml
version: 1 # Version of the config.yml schema itself
thread_version: "0.1.0" # Version of the thread's content, ideally following Conventional Commits
metadata: # All metadata fields are optional
  description: "A brief description of what this thread provides."
  author: "Author Name <author@example.com>"
  license: "MIT" # SPDX license identifier
# Future Improvement:
# template_variables:
#   description: "Variables for templating file content or names."
#   values:
#     service_name: "default_service"
#     default_port: 8080
```

### 4.3. Thread `_thread/` Directory

Located within a thread's directory (e.g., `thread_name/_thread/`).

- **Purpose:** This directory contains all the files and subdirectories that will be copied into the target project's root. The structure within `_thread/` is mirrored in the project.

## 5. Commands

Loom commands will be organized into modules, likely residing in a `modules/` subdirectory of the Go project. The main entry point (`main.go`) will parse arguments and dispatch to the appropriate module.

- **`loom add <thread_source>`**
    - Adds a new thread to the project.
    - `<thread_source>`: URL or path to the thread (e.g., GitHub URL, local path).
    - The thread's contents (from its `_thread` subfolder) are placed into the project root.
    - Prompts for conflict resolution if files collide.
    - Updates the `loom.yaml` file.

- **`loom remove <thread_name_or_source> [*]`**
    - Removes a thread from the project.
    - `<thread_name_or_source>`: The name or source identifier of the thread to remove, as listed in `loom.yaml`.
    - `*`: A special argument to remove all threads from the project.
    - Removes files associated with the thread (respecting ownership if other threads also provided the file initially â€“ complex cases might require careful handling or simply remove files owned by this thread).
    - Updates the `loom.yaml` file.

- **`loom config <subcommand>`**
    - Manages Loom's configuration for thread stores.
    - **`loom config add <path_or_url>`**
        - Adds a new thread store.
        - `<path_or_url>`: Path for local store, base URL for GitHub store (e.g., `github:my-org/loom-threads`).
    - **`loom config remove <name_or_path>`**
        - Removes a configured thread store.

- **`loom list`**
    - Lists all threads available from configured stores.
    - May also list threads currently active in the project (read from `loom.yaml`).

- **`loom weave [thread_name]` (alias: `install`)**
    - "Installs" or "weaves" threads into the project.
    - If `[thread_name]` is provided, re-applies only that specific thread from its source to the project, overwriting existing files it owns.
    - If no argument is provided, re-applies all threads listed in the `loom.yaml` file from their respective sources. This is the brute-force update mechanism.
    - File conflicts resolved previously and recorded in `loom.yaml` will be respected. Re-prompting the user is a potential future improvement.

## 6. Thread Design

A "thread" is a self-contained template with a specific directory structure.

- `thread_name/` (Root directory of the thread)
    - `_thread/`
        - All files and subdirectories within `_thread/` are intended to be copied directly into the root of the target project. The internal structure of `_thread/` is preserved.
    - `README.md` (optional)
        - Documentation for the thread itself: what it provides, how to use it, any configuration options.
    - `LICENSE` (optional)
        - License under which the thread's contents are provided.
    - `config.yml` (see Section 4.2)
        - Configuration and metadata for the thread.

## 7. Stores

Stores are locations where collections of threads are kept, using the "Thread Design" format described above.

-   **GitHub Stores:**
    -   A GitHub repository where each thread is a subdirectory within the repo (or the repo itself is a single thread).
    -   The store URL would be the base path to these threads (e.g., `https://github.com/user/my-threads-repo`). Loom would then list/fetch `my-threads-repo/thread_A`, `my-threads-repo/thread_B`, etc.
-   **Project Stores:**
    -   Located within the user's project, typically in a `.loom/threads/` directory.
    -   Allows for project-specific, reusable components that aren't meant to be global.
-   **Local PC Stores:**
    -   A user-configured directory on their local filesystem (e.g., `~/.config/loom/stores/my-local-threads` or `~/loom-threads`).
    -   Loom will have an opinionated default location if not explicitly configured.
-   **Store Configuration:**
    -   The configuration for defined stores (paths, URLs, names) will be stored in a global Loom configuration file, likely located alongside the Loom CLI executable or in a standard user configuration directory (e.g., `~/.config/loom/config.yml`).

## 8. Folder Structure (Loom Tool - Development and Deployed)

### 8.1. Development Structure

**Note:** The following structure is based on the `golang-standards/project-layout`. This means Go source code is organized into `cmd/`, `internal/`, and `pkg/` directories rather than a single top-level `src/` directory.

```
loom/
â”œâ”€â”€ cmd/                      # Main applications for the project
â”‚   â””â”€â”€ loom/                 # The Loom CLI application
â”‚       â””â”€â”€ main.go           # Main entry point, CLI argument parsing, command dispatch
â”œâ”€â”€ internal/                 # Private application and library code
â”‚   â”œâ”€â”€ cli/                  # CLI command logic (previously src/modules)
â”‚   â”‚   â”œâ”€â”€ add/
â”‚   â”‚   â”‚   â””â”€â”€ add.go
â”‚   â”‚   â”œâ”€â”€ init/             # Command for initializing a new loom project
â”‚   â”‚   â”‚   â””â”€â”€ init.go
â”‚   â”‚   â”œâ”€â”€ list/             # Command for listing threads
â”‚   â”‚   â”‚   â””â”€â”€ list.go
â”‚   â”‚   â”œâ”€â”€ remove/
â”‚   â”‚   â”‚   â””â”€â”€ remove.go
â”‚   â”‚   â”œâ”€â”€ weave/            # Command for weaving threads
â”‚   â”‚   â”‚   â””â”€â”€ weave.go
â”‚   â”‚   â”œâ”€â”€ config/           # Command for managing Loom configuration
â”‚   â”‚   â”‚   â””â”€â”€ config.go
â”‚   â”‚   â””â”€â”€ ...               # Other command modules (if any)
â”‚   â”œâ”€â”€ core/                 # Core application logic (file ops, loom.yaml parsing, store mgmt - previously src/core)
â”‚   â”‚   â””â”€â”€ ...
â”‚   â”œâ”€â”€ util/                 # Utility functions (previously src/util)
â”‚   â”‚   â””â”€â”€ ...
â”œâ”€â”€ pkg/                      # Public library code reusable by other projects (if any - initially empty)
â”‚   â””â”€â”€ ...
â”œâ”€â”€ scripts/                  # Scripts for build, install, analysis, etc. (previously install/)
â”‚   â”œâ”€â”€ install.sh            # Installer script for macOS/Linux
â”‚   â””â”€â”€ install.ps1           # Installer script for Windows
â”œâ”€â”€ configs/                  # Configuration files (e.g., for different environments - placeholder)
â”œâ”€â”€ docs/                     # Loom tool's own documentation (user guides, design docs, etc.)
â”œâ”€â”€ test/                     # Test data, E2E tests, or other non-unit tests
â”‚   â””â”€â”€ e2e/                  # End-to-end tests
â”‚       â””â”€â”€ ...
â”œâ”€â”€ .github/                  # GitHub-specific files (workflows, issue templates)
â”œâ”€â”€ go.mod                    # Go modules file
â”œâ”€â”€ go.sum                    # Go checksum file
â””â”€â”€ README.md                 # Project README for Loom development
```

### 8.2. Deployed Structure (Preliminary - Details for future finalization)

The exact deployed structure is to be determined as a future improvement, but might look like:

**macOS/Linux:**
```
/usr/local/bin/loom or ~/.local/bin/loom (the main executable wrapper)

~/.config/loom/
    config.yml (global Loom configuration, stores, etc.)
    default_store/ (optional default local PC store)

Potentially an install/ directory if Loom manages its own updates:
~/.local/share/loom/install/
    next/ (for staging updates)
    update_pending/ (flag or data for pending updates)
```

**Windows:**
```
C:\Program Files\Loom\loom.exe or %LOCALAPPDATA%\Loom\loom.exe

%APPDATA%\Loom\ or %LOCALAPPDATA%\Loom\
    config.yml
    default_store/

Update mechanism TBD.
```

## 9. Conclusion

Loom aims to be a developer-friendly tool that significantly accelerates the initial setup phase of projects. By leveraging reusable "threads" and a simple, direct approach to file management, Loom allows developers to focus on building features rather than boilerplate, without the overhead of complex dependency management systems. Its core philosophy is speed and ease of use for scaffolding.

## 10. Tech Stack

-   **Language:** Go (Golang)
    -   **CLI Framework:** `urfave/cli`
-   **Platform:** Cross-platform (macOS, Linux, Windows)

## 11. Project-Specific Coding Rules (Go)

These rules supplement any global AI project guidelines and define standards specific to the Loom Go project.

### 11.1. Language & Environment

-   **Go Version:** Target the latest stable Go version (e.g., Go 1.21+ or as decided by the team). Specify in `go.mod`.
-   **Modules:** Use Go Modules for dependency management.

### 11.2. Go Coding Standards

#### 11.2.1. Style & Formatting

-   **Formatting:** All Go code must be formatted with `gofmt` (or `goimports`). Configure IDEs/editors to format on save.
-   **Linting:** Use `golangci-lint` with a pre-defined configuration to enforce a common set of linters (e.g., `errcheck`, `govet`, `staticcheck`, `unused`, `stylecheck`).
-   **Line Length:** Aim for a reasonable line length (e.g., 100-120 characters), but prioritize clarity over strict adherence if a longer line is more readable.
-   **Naming Conventions:**
    -   Follow standard Go naming conventions (e.g., `camelCase` for local variables and parameters, `PascalCase` for exported identifiers).
    -   Package names should be short, concise, and lowercase.
    -   Strive for descriptive names. Single-letter variables are acceptable for very short scopes or idiomatic loops (e.g., `i`, `k`, `v`).
-   **Error Handling:**
    -   Errors are values. Handle errors explicitly; do not ignore them using `_` unless there's a very deliberate reason (and comment it).
    -   Use `fmt.Errorf` with `%w` to wrap errors for context, or use custom error types where appropriate.
    -   Error messages should be lowercase and not end with punctuation.
-   **Concurrency:**
    -   Use goroutines and channels judiciously.
    -   Be mindful of race conditions; use the race detector (`go test -race`).
    -   Prefer channels for synchronization and communication where idiomatic.

#### 11.2.2. Documentation & Comments

-   **Godoc:** All exported identifiers (variables, functions, types, constants) must have godoc comments.
    -   Comments should start with the name of the thing being described.
    -   Provide clear, concise explanations of purpose, behavior, parameters, and return values.
-   **Package Comments:** Each package should have a package comment (`// package mypackage ...`) explaining its role.
-   **Inline Comments:** Use inline comments (`//`) to explain complex, non-obvious, or important sections of code. Avoid commenting on obvious code.

#### 11.2.3. Testing

-   **Standard Library:** Use the standard `testing` package for unit and integration tests.
-   **Test Files:** Test files should be named `_test.go` (e.g., `myfile_test.go`).
-   **Coverage:** Aim for high test coverage. Use `go test -cover` to check.
-   **Table-Driven Tests:** Use table-driven tests for testing multiple scenarios of a function.
-   **Testable Code:** Write code that is easily testable (e.g., by decoupling dependencies, using interfaces).

#### 11.2.4. Project Structure

-   **Layout:** Adhere to the `golang-standards/project-layout` as described in Section 8.1 of this document.
    -   `cmd/loom/main.go`: Main application entry point.
    -   `internal/`: Contains all private application code, organized into sub-packages like `cli` (for command-specific logic), `core` (for business logic), and `util` (for shared utilities). This code is not importable by other projects.
    -   `pkg/`: (If applicable) For code that is intended to be used as a public library by other projects.
-   **Unit Tests:** Co-locate unit tests with the code they are testing. Test files should be named `_test.go` (e.g., `myfile_test.go` alongside `myfile.go`).
-   **E2E/Integration Tests:** Place end-to-end or integration tests in the top-level `/test` directory (e.g., `/test/e2e/`).
-   Internal packages (not meant to be imported by other projects) should reside within `internal/`.


