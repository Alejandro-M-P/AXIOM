# 🐛 Bug Tracker & MVP Status

This document details the known issues and pending tasks during the migration from Bash to **Native Go**.

### 🔴 Critical & Security
- [x] **Volume Security**: Native `build` and `create` functions need auditing regarding rootless container permission handling.
- [x] **GPU Injection**: Sporadic failures when detecting and mounting NVIDIA/AMD drivers from the Go binary in `build`.
- [x] **Persistence**: `~/.entorno/` contents sometimes fail to mount correctly on the first startup.
- [x] **[ID-BUG-001] Path Traversal**: Critical vulnerability in `Create`, `Delete`, and `Info` where bunker names are concatenated directly into paths without using `filepath.Clean()`.
- [x] **[ID-BUG-005] Loose Permissions**: `os.MkdirAll` creates folders in `.entorno/` with `0755` instead of the restrictive `0700`, exposing configs to other host users.
- [ ] **[ID-BUG-019] Deleting image inside cloned bunker**: Deleting an image while inside a cloned bunker deletes the cloned bunker.
- [ ] **[ID-BUG-020] Gentle-ai execution**: When entering the bunker, `gentle-ai` does not execute to trigger the installer (should check installation status via a flag file or native go).

### 🟡 Pending Migration (Bash ➔ Go)
- [ ] Translate interactive Git tools (`lib/git.sh`) to native code in `pkg/`.
- [ ] Migrate deep clean logic (`axiom purge`).
- [ ] **[ID-BUG-002] System Utility Dependency**: Replace legacy bash commands like `du -sh` (in `bunkerEnvSize`) using native `filepath.WalkDir`.
- [x] **[ID-BUG-003] Fragile Parsing**: Stop splitting text with `strings.Split` in `distroboxExists` and use `--format json` decoded with `encoding/json`.
- [x] **[ID-BUG-004] Missing Goroutines**: Massive tasks like `Prune` delete environments sequentially. They must be refactored with `sync.WaitGroup` 
### 🟠 Architecture & Separation of Concerns
- [x] **[ID-BUG-006] Logic-UI Coupling**: `Manager` functions (`Create`, `Delete`, `Info`) print cards and logs directly to the console, mixing business logic with presentation.
- [x] **[ID-BUG-008] User Interaction in Core**: Functions like `Delete` and `Prune` read from `stdin` to ask for confirmation, blocking their use in scripts and violating layer separation.
- [x] **[ID-BUG-009] UI Responsibilities in Manager**: The `Manager` has a `Help()` method whose sole function is printing a help card, a task belonging to the UI layer.
- [ ] **[ID-BUG-007] Git PATH Dependency**: The `bunkerGitBranch` function executes the `git` binary instead of using a native library like `go-git`, creating a fragile external dependency.
- [x] **[ID-BUG-010] Hardcoded ANSI Sequences**: Clearing the terminal screen with `fmt.Print("\033[H\033[2J")` inside Core functions destroys the portability of the calls.
- [x] **[ID-BUG-014] Architectural Split Brain (main.go)**: Bypassing the BubbleTea model (`tea.Program`) to inject manual interactive logic (`bufio`) in direct commands, preventing UI scaling.
- [ ] **[ID-BUG-016] DRY Violation in Routing**: Duplication of command lists (`Run` vs `KnownCommand`) in `bunkerController.go`. Needs migration to function maps or Cobra/CLI.

### 🟡 Code Quality & Go Standards
- [ ] **[ID-BUG-011] Error Silencing**: Multiple `exec.Command` executions silently discard the error or `Stderr` by assigning it to `_`, hiding critical failures (e.g., in `prepareSSHAgent`).
- [ ] **[ID-BUG-012] Naked Errors (No Wrapping)**: Returning `return err` without context (e.g., after `os.MkdirAll`) makes tracing difficult. `fmt.Errorf("... %w", err)` should be used.
- [ ] **[ID-BUG-013] Manual `.env` Parsing**: The method in `LoadEnvFile` using `strings.Cut(line, "=")` will truncate secrets or tokens containing an equals sign (`=`) in their value.
- [ ] **[ID-BUG-017] Magic Variable Resolution**: `prepareSSHAgent` relies on `os.Getenv("HOME")` instead of using the robust and secure `os.UserHomeDir()`.

### 🟢 Concurrency & Execution Control
- [ ] **[ID-BUG-015] Zombie Processes without Context**: Calls to binaries with `exec.Command` (like `pacman`) are unbounded. Use `exec.CommandContext` to prevent permanent hangs.
- [x] **[ID-BUG-015] Zombie Processes without Context**: Calls to binaries with `exec.Command` (like `pacman`) are unbounded. Use `exec.CommandContext` to prevent permanent hangs.
- [x] **[ID-BUG-018] Hangs from Sudo Prompts**: `runCommandQuiet("sudo", "-v")` can leave the app hanging if it asks for a password without a correctly bound TTY.

### 🔵 Next Steps
- [ ] **Configuration**: Move the variables system from `.env` to a professional `config.toml` file.
- [ ] **Error Handling**: Create custom error types (e.g., `ErrBunkerNotFound`, `ErrImageMissing`) instead of using generic strings with `fmt.Errorf`.
- [ ] **Podman API Integration**: Investigate direct connection to the Podman REST socket instead of executing shell subprocesses.
