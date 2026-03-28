# 🐛 Bug Tracker & MVP Status

This document details the known issues and pending tasks during the migration from Bash to **Native Go**.

### 🔴 Critical & Security
- [x] **Volume Security**: Native `build` and `create` functions need auditing regarding rootless container permission handling.
- [x] **GPU Injection**: Sporadic failures when detecting and mounting NVIDIA/AMD drivers from the Go binary in `build`.
- [x] **Persistence**: `~/.entorno/` contents sometimes fail to mount correctly on the first startup.
- [x] **[ID-BUG-001] Path Traversal**: Critical vulnerability in `Create`, `Delete`, and `Info` where bunker names are concatenated directly into paths without using `filepath.Clean()`.
- [x] **[ID-BUG-005] Loose Permissions**: `os.MkdirAll` creates folders in `.entorno/` with `0755` instead of the restrictive `0700`, exposing configs to other host users.
- [ ] **[ID-BUG-032] Init creates structure in wrong path**: When the user inputs a relative path like `Documentos/dev` during `axiom install`, the directory structure is created in `$HOME` instead of the resolved absolute path. Likely caused by how `axiomPath` is initialized and passed to `PrepareFS` and `config.Save` — root cause requires `main.go` to confirm.
- [ ] **[ID-BUG-019] Deleting image inside cloned bunker**: Deleting an image while inside a cloned bunker deletes the cloned bunker.
- [ ] **[ID-BUG-020] Gentle-ai execution**: When entering the bunker, `gentle-ai` does not execute to trigger the installer (should check installation status via a flag file or native go).

### 🟡 Pending Migration (Bash ➔ Go)
- [ ] Translate interactive Git tools (`lib/git.sh`) to native code in `pkg/`.
- [ ] Migrate deep clean logic (`axiom purge`).
- [ ] **[ID-BUG-002] System Utility Dependency**: Replace legacy bash commands like `du -sh` (in `bunkerEnvSize`) using native `filepath.WalkDir`.
- [x] **[ID-BUG-003] Fragile Parsing**: Stop splitting text with `strings.Split` in `distroboxExists` and use `--format json` decoded with `encoding/json`.
- [x] **[ID-BUG-004] Missing Goroutines**: Massive tasks like `Prune` delete environments sequentially. They must be refactored with `sync.WaitGroup`.

### 🟠 Architecture & Separation of Concerns
- [x] **[ID-BUG-006] Logic-UI Coupling**: `Manager` functions (`Create`, `Delete`, `Info`) print cards and logs directly to the console, mixing business logic with presentation.
- [x] **[ID-BUG-008] User Interaction in Core**: Functions like `Delete` and `Prune` read from `stdin` to ask for confirmation, blocking their use in scripts and violating layer separation.
- [x] **[ID-BUG-009] UI Responsibilities in Manager**: The `Manager` has a `Help()` method whose sole function is printing a help card, a task belonging to the UI layer.
- [ ] **[ID-BUG-007] Git PATH Dependency**: The `bunkerGitBranch` function executes the `git` binary instead of using a native library like `go-git`, creating a fragile external dependency.
- [x] **[ID-BUG-010] Hardcoded ANSI Sequences**: Clearing the terminal screen with `fmt.Print("\033[H\033[2J")` inside Core functions destroys the portability of the calls.
- [x] **[ID-BUG-014] Architectural Split Brain (main.go)**: Bypassing the BubbleTea model (`tea.Program`) to inject manual interactive logic (`bufio`) in direct commands, preventing UI scaling.
- [ ] **[ID-BUG-016] DRY Violation in Routing**: Duplication of command lists (`Run` vs `KnownCommand`) in `bunkerController.go`. Needs migration to function maps or Cobra/CLI.
- [ ] **[ID-BUG-021] Clean Architecture violation in `select.go`**: `pkg/core/services/select.go` imports `axiom/pkg/adapters/ui/styles` directly, breaking Golden Rule #1. The Core cannot import Adapters.
- [ ] **[ID-BUG-022] Clean Architecture violation in `lifecycle.go`**: `pkg/core/services/lifecycle.go` imports `axiom/pkg/adapters/system/gpu` directly and uses `gpu.GPUInfo` as a concrete type instead of `domain.GPUInfo`, breaking layer isolation.
- [ ] **[ID-BUG-023] System logic inside UI layer (`form.go`)**: `pkg/adapters/ui/views/form.go` imports `axiom/pkg/adapters/system` and calls `install.CheckDeps()`, `install.PrepareFS()`, and `install.CreateWrapper()` directly. A view should never execute system logic.

### 🟡 Code Quality & Go Standards
- [ ] **[ID-BUG-011] Error Silencing**: Multiple `exec.Command` executions silently discard the error or `Stderr` by assigning it to `_`, hiding critical failures (e.g., in `prepareSSHAgent`).
- [ ] **[ID-BUG-012] Naked Errors (No Wrapping)**: Returning `return err` without context (e.g., after `os.MkdirAll`) makes tracing difficult. `fmt.Errorf("... %w", err)` should be used.
- [ ] **[ID-BUG-013] Manual `.env` Parsing**: The method in `LoadEnvFile` using `strings.Cut(line, "=")` will truncate secrets or tokens containing an equals sign (`=`) in their value.
- [ ] **[ID-BUG-017] Magic Variable Resolution**: `prepareSSHAgent` in `instance.go` relies on `os.Getenv("HOME")` instead of using the robust and secure `os.UserHomeDir()`. Same pattern also present in `NewModel` (`form.go`) where `BaseDir` is initialized with `os.Getenv("HOME")`.
- [ ] **[ID-BUG-024] Orphan import at end of `adapter_test.go`**: `import "fmt"` appears at the bottom of `unit_tests/podman/adapter_test.go` outside the import block. The file does not compile.
- [ ] **[ID-BUG-025] `runInteractiveInContainer` duplicates arg prefix**: In `lifecycle.go`, `runInteractiveInContainer` manually prepends `-n containerName --` to args and then passes `m.buildContainerName` again as the first parameter of `RunCommandWithInput`, resulting in the container name being passed twice.
- [ ] **[ID-BUG-026] `EnsureFileExists` silently drops `Close` error**: In `pkg/adapters/fs/adapter.go`, `EnsureFileExists` calls `file.Close()` but the error is returned directly without wrapping, and any error from `OpenFile` on an already-existing path is not handled consistently.
- [ ] **[ID-BUG-027] Empty `containerName` passed to `RunCommand` in `instance.go`**: Several calls in `instance.go` pass `""` as `containerName` followed by commands like `"distrobox-create"` or `"podman"`. The adapter wraps these as `distrobox-enter -n  -- distrobox-create ...` which is incorrect. Direct host commands need a separate execution path that does not go through `distrobox-enter`.

### 🟢 Concurrency & Execution Control
- [ ] **[ID-BUG-015] Zombie Processes without Context**: Calls to binaries with `exec.Command` (like `pacman`) are unbounded. Use `exec.CommandContext` to prevent permanent hangs.
- [x] **[ID-BUG-018] Hangs from Sudo Prompts**: `runCommandQuiet("sudo", "-v")` can leave the app hanging if it asks for a password without a correctly bound TTY.
- [ ] **[ID-BUG-028] Race condition in `Prune` goroutines**: In `instance.go`, the `Prune` function calls `m.UI.ShowLog(...)` from multiple concurrent goroutines via `sync.WaitGroup`. `ConsoleUI.ShowLog` calls `fmt.Printf` directly which is not goroutine-safe, causing interleaved or corrupted output.

### 🔵 UI & i18n
- [ ] **[ID-BUG-029] i18n keys rendered raw in UI**: `ShowWarning` and `ShowCommandCard` in `presenter.go` pass translation keys like `"warnings.bunker_exists.title"` or `"fields.name"` directly to lipgloss renderers without resolving them through `GetText`. Users see raw keys instead of translated strings.
- [ ] **[ID-BUG-030] `GetText` only handles 2-level keys**: In `presenter.go`, `GetText` splits the key by `"."` and only processes keys with exactly 2 parts (`len(parts) == 2`). Keys with 3 levels like `"warnings.bunker_exists.title"` silently fall through and return the raw key string.
- [ ] **[ID-BUG-031] Starship not found on `--rcfile` entry**: When entering an existing bunker via `axiom create`, bash is launched with `--rcfile`, which skips `/etc/profile` and `~/.bash_profile`. The system PATH is not initialized before the `.bashrc` is sourced, so `starship` is not found and the shell starts without the custom prompt. Fix: prepend explicit PATH export at the top of the generated `.bashrc`.

### 🔵 Next Steps
- [ ] **Configuration**: Move the variables system from `.env` to a professional `config.toml` file.
- [ ] **Error Handling**: Create custom error types (e.g., `ErrBunkerNotFound`, `ErrImageMissing`) instead of using generic strings with `fmt.Errorf`.
- [ ] **Podman API Integration**: Investigate direct connection to the Podman REST socket instead of executing shell subprocesses.