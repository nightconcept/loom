# Task Checklist

**Purpose:** Tracks all tasks, milestones, and backlog for the Loom scaffolding tool. Each task includes a manual verification step.

**Multiplatform Policy:** All tasks, implementations, and verifications MUST consider cross-platform compatibility (Linux, macOS, and Windows) unless otherwise specified. Contributors (including AI) are required to design, implement, and test with multiplatform support as a baseline expectation. Any platform-specific logic must be clearly documented and justified in both code and task notes.

---

## CLI Tool Name

- The CLI executable is called `loom`.
- All documentation, usage, and examples should refer to the CLI as `loom`.

---

## Milestone 1: Core Manifest and Project Initialization (YYYY-MM-DD)

- [x] **Task 1.1: Design and define `loom.yaml` schema (YYYY-MM-DD)**
  - [x] Define fields: `version`, `threads` (with `name`, `source`, `files`).
  - [x] Manual Verification: Review schema against PRD.md, create a sample `loom.yaml` file, and validate its structure.

- [x] **Task 1.2: Basic CLI structure and entry point (YYYY-MM-DD)**
  - [x] Set up Go project structure (`cmd/loom/main.go`, `internal/`, etc.) as per PRD Section 8.1.
  - [x] Implement CLI argument parsing using `urfave/cli`.
  - [x] Manual Verification: Build and run `loom --version`, `loom --help`.

- [x] **Task 1.3: Implement `loom init` (Placeholder - if needed for `loom.yaml` creation) (YYYY-MM-DD)**
  - [x] If `loom.yaml` is not automatically created by other commands, implement `loom init` to create a basic `loom.yaml`.
  - [x] Manual Verification: Run `loom init`, check for `loom.yaml` creation.

## Milestone 2: Thread Definition and Local Operations (YYYY-MM-DD)

- [x] **Task 2.1: Define Thread Structure and `config.yml` (2025-05-06)**
  - [x] Establish standard thread directory structure: `thread_name/_thread/`, `thread_name/config.yml`, `thread_name/README.md` (optional), `thread_name/LICENSE` (optional).
  - [x] Define `config.yml` schema: `version`, `thread_version`, `metadata` (description, author, license).
  - [x] Manual Verification: Create a sample thread adhering to the structure and `config.yml` schema.

- [x] **Task 2.2: Implement `loom add <name_of_thread>` (YYYY-MM-DD)**
  - [x] Search for thread in the `.loom/` folder of the project.
  - [x] Copy contents of the thread's `_thread/` directory to the project root.
  - [x] Add entry to `loom.yaml` (initially without `files`).
  - [x] Manual Verification: Add a local thread, check files are copied, and `loom.yaml` is updated.

- [x] **Task 2.3: Implement `loom weave` (Initial for all project threads) (YYYY-MM-DD)**
  - [x] Re-apply all threads listed in `loom.yaml` from their sources in sequential (top to bottom) order.
  - [x] For now, this will be a brute-force overwrite for files from local threads.
  - [x] Manual Verification: Modify a file provided by a thread, run `loom weave`, verify file is reverted to thread's version.

- [x] **Task 2.4: Implement `loom list` (Project threads) (2025-05-07)**
  - [x] List threads currently active in the project by reading `loom.yaml`.
  - [x] Manual Verification: Add a thread, run `loom list`, verify output.

- [x] **Task 2.4.5: Update `loom add` to track files per thread (YYYY-MM-DD)**
  - [x] Modify `loom add` to record all files added by each thread in the `loom.yaml` file.
  - [x] Store files under each thread's entry using the directory structure format specified in the PRD (`files` map with directory paths as keys and lists of filenames as values).
  - [x] Manual Verification: Add a thread, verify `loom.yaml` correctly lists all files copied from the thread's `_thread/` directory.

- [x] **Task 2.5: Implement `loom remove <thread_name>` (2025-05-08)**
  - [x] Remove thread entry from `loom.yaml`.
  - [x] Remove files owned by the thread (simple removal for now, conflict/shared files later).
  - [x] Manual Verification: Add a thread, then remove it. Verify files are removed and `loom.yaml` is updated.

## Milestone 3: Conflict Resolution (YYYY-MM-DD)

- [x] **Task 3.1: Implement interactive conflict resolution for `loom add` / `loom weave` (2025-05-08)**
  - [x] When a file collision occurs, prompt the user to choose which thread "owns" the file.
  - [x] Manual Verification: Add two threads with an overlapping file, verify prompt and resolution.

- [x] **Task 3.2: Update `loom.yaml` with `files` (2025-05-08)**
  - [x] Store conflict resolution choices in `loom.yaml` under the winning thread's `files` map.
  - [x] Ensure `loom weave` respects these ownership rules.
  - [x] Manual Verification: Resolve a conflict, inspect `loom.yaml`. Run `loom weave`, verify owned files are correctly applied.

## Milestone 4: Thread Stores (YYYY-MM-DD)

- [x] **Task 4.1: Implement Local PC Store (`loom config add <path>`) (2025-05-09)**
  - [x] Store configuration in a global Loom config file (e.g., `~/.config/loom/config.yml`).
  - [x] Allow adding a local directory as a named thread store.
  - [x] Manual Verification: Add a local PC store, verify global config.

- [x] **Task 4.2: Update `loom add` to use Local PC Store threads (2025-05-09)**
  - [x] Allow `loom add <thread_name>` or similar syntax but prioritize the project's threads first. The next priority should be local PC stores.
  - [x] Manual Verification: Add a thread from a configured local PC store.

- [x] **Task 4.3: Implement Project Store (YYYY-MM-DD)**
  - [x] Support sourcing threads from a `.loom/` directory within the project.
  - [x] Update `loom add` to use project store threads (e.g., `loom add project:<thread_name_in_.loom>`).
  - [x] Manual Verification: Create a thread in `.loom/threads/`, add it to the project.

- [ ] **Task 4.4: Implement GitHub Store (`loom config add github <base_url> [name]`) (YYYY-MM-DD)**
  - [ ] Allow adding a GitHub repository base URL as a named thread store.
  - [ ] Manual Verification: Add a GitHub store, verify global config.

- [ ] **Task 4.5: Update `loom add` to use GitHub Store threads (YYYY-MM-DD)**
  - [ ] Implement logic to fetch thread contents from GitHub (e.g., `loom add github:user/repo/path/to/thread` or `loom add <github_store_name>/<thread_name>`).
  - [ ] Manual Verification: Add a thread from a configured GitHub store or direct GitHub URL.

- [x] **Task 4.6: Update `loom list` to show available store threads (2025-05-13)**
  - [x] Extend `loom list` to show threads available from all configured stores.
  - [x] Manual Verification: Configure stores, run `loom list`, verify output.

- [x] **Task 4.7: Refactor `loom config add` to infer type and name (2025-05-13)**
  - [x] Modify `loom config add` to only require a single `<path_or_url>` argument.
  - [x] Infer store type (local, github) from the argument.
  - [x] Infer store name (basename for local, repo name for GitHub).
  - [x] Prevent adding stores with duplicate paths/URLs (case-insensitive).
  - [x] Prompt for a custom name if the inferred name conflicts with an existing store but the path/URL is unique.
  - [x] Manual Verification:
    - Add a local store using a path, verify type and name are inferred.
    - Attempt to add the same local store path again, verify error.
    - Add a local store with a path that results in a conflicting name (but different path), verify prompt for new name.
    - (Future) Add a GitHub store using a URL, verify type and name are inferred.
    - (Future) Attempt to add the same GitHub URL again, verify error.
    - (Future) Add a GitHub store with a URL that results in a conflicting name (but different URL), verify prompt for new name.

- [x] **Task 4.7: Implement `loom config remove <name_or_path>` (2025-05-14)**
  - [x] Remove a configured thread store from the global Loom config.
  - [x] Manual Verification: Add a store, then remove it. Verify global config.

- [x] **Task 4.8: Implement `loom config list` (2025-05-14)**
  - [x] List all configured thread stores with their names, types, and paths/URLs.
  - [x] Manual Verification: Add a few stores, run `loom config list`, verify output. Remove a store, run `loom config list` again, verify.

## Milestone 5: Advanced Commands & Refinements (YYYY-MM-DD)

- [x] **Task 5.1: Implement `loom weave [thread_name]` (2025-05-14)**
  - [x] Re-apply only the specified thread from its source, respecting its owned files.
  - [x] Manual Verification: Modify a file from a specific thread, run `loom weave <thread_name>`, verify only that thread's files are affected.

- [x] **Task 5.2: Implement `loom remove *` (2025-05-14)**
  - [x] Remove all threads from the project and clear `loom.yaml`.
  - [x] Remove all files that were owned by any thread (careful with shared files not owned by any removed thread - initial simple approach: remove all files listed in any `files`).
  - [x] Manual Verification: Add multiple threads, run `loom remove *`, verify project is cleaned.

## Milestone 6: Documentation, Testing, and Go Standards (Ongoing) (YYYY-MM-DD)

- [ ] **Task 6.1: Create User Documentation (README.md for Loom tool) (YYYY-MM-DD)**
  - [ ] Document all commands, `loom.yaml` structure, thread design, and store configuration.
  - [ ] Manual Verification: Review README for clarity, completeness, and accuracy.

- [x] **Task 6.2: Implement Go Coding Standards (Ongoing) (2025-05-14)**
  - [x] Ensure all Go code is formatted with `gofmt`/`goimports`.
  - [x] Set up and use `golangci-lint` with a standard configuration.
  - [x] All exported identifiers have godoc comments.
  - [x] Packages have package comments.
  - [x] Adhere to error handling and naming conventions from PRD.md Section 11.2.
  - [x] Reduce cyclomatic complexity of `add.Command()`
  - [ ] Manual Verification: Periodically review code quality and linting results.

- [ ] **Task 6.3: Develop Unit Tests (Ongoing) (YYYY-MM-DD)**
  - [ ] Write unit tests for core logic in `internal/core`, `internal/cli` modules.
  - [ ] Use the standard `testing` package. Test files named `_test.go`.
  - [ ] Aim for high test coverage.
  - [ ] Manual Verification: Run `go test ./...`, check coverage reports.

- [ ] **Task 6.4: Develop E2E Tests (Ongoing) (YYYY-MM-DD)**
  - [ ] Create end-to-end tests in the `/test/e2e/` directory.
  - [ ] Tests should cover CLI command workflows (add, remove, weave, list, config).
  - [ ] **`loom add` command E2E Test Scenarios:**
    - [ ] **Argument Parsing:**
      - [ ] Test `loom add` with no arguments (should fail with usage message).
      - [ ] Test `loom add /` (invalid format).
      - [ ] Test `loom add store/` (invalid format - missing thread name).
      - [ ] Test `loom add /thread` (invalid format - missing store name).
    - [ ] **Thread Source and Resolution:**
      - [ ] Test adding a thread that already exists in the project's local `.loom/` directory.
      - [ ] Test adding a thread by explicitly specifying an existing store name (e.g., `loom add myStore/myTestThread`).
      - [ ] Test error: Thread not found in the specified store.
      - [ ] Test error: Specified store name does not exist in global configuration.
      - [ ] Test error: Thread not found in any configured store or project `.loom/` folder.
    - [ ] **File Conflict Handling (requires mechanism for non-interactive test or --force flag):**
      - [ ] Test adding a thread when a file to be copied already exists and is owned by the *same thread* (should overwrite).
      - [ ] Test adding a thread when a file to be copied already exists and is owned by a *different thread* (verify prompt or behavior with non-interactive flag).
      - [ ] Test adding a thread when a file to be copied already exists and is *unowned* (verify prompt or behavior with non-interactive flag).
    - [ ] **Project `loom.yaml` Manipulation:**
      - [ ] Test updating an *existing* thread entry in `loom.yaml` (e.g., re-adding a thread, potentially from a different source).
      - [ ] Test `removeFileFromOtherThreads` logic during conflict resolution (if a file changes ownership).
      - [ ] Test adding a thread when `loom.yaml` is present but malformed (e.g., invalid YAML).
    - [ ] **Extraneous Arguments:**
      - [ ] Test `loom add myThread extraneousArg` (should ideally ignore extraneous args or error gracefully depending on CLI library behavior).
  - [ ] Manual Verification: Run E2E tests, verify they cover primary use cases.

## Milestone 7: Future Considerations (Placeholder) (YYYY-MM-DD)

- [ ] **Task 7.1: Investigate templating language support (YYYY-MM-DD)**
  - [ ] Research options for dynamic folder names and file content based on variables (as per PRD Section 2 and 4.2).
  - [ ] Manual Verification: Document findings and potential approaches.

- [ ] **Task 7.2: Refine Deployed Structure and Installers (YYYY-MM-DD)**
  - [ ] Finalize deployed structure for macOS, Linux, and Windows (PRD Section 8.2).
  - [ ] Develop installer scripts (`install.sh`, `install.ps1`) as per PRD Section 8.1.
  - [ ] Manual Verification: Test installers on all target platforms.

*Last updated: 2025-05-09*



