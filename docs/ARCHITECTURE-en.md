# 🏛️ AXIOM Architecture

AXIOM uses a **Clean Architecture** based on the Ports and Adapters pattern.
The main goal of this structure is to **completely decouple the business logic (Core) from the user interface and the underlying system (Podman, OS)**.

This allows AXIOM to be highly testable, scalable, and operable by both humans (via an interactive Terminal) and Artificial Intelligences (via a JSON Agent API).

---

## 📂 Directory Structure

All business code is in `internal/`. Entry point is in `cmd/`.

```
cmd/axiom/                  # Entry point
├── main.go                 # Bootstrap + DI
└── router_commands.go      # Router re-exports

internal/                   # 🔒 Private AXIOM code
├── domain/                 # 🧠 ENTITIES (pure, no deps)
│   ├── bunker.go          # Bunker, BunkerConfig
│   ├── system.go          # GPUInfo, EnvConfig
│   └── slot.go            # SlotItem, SlotSelection
│
├── ports/                  # 🔌 CONTRACTS (interfaces)
│   ├── runtime.go          # IBunkerRuntime
│   ├── filesystem.go      # IFileSystem
│   ├── system.go          # ISystem
│   └── presenter.go       # IPresenter
│
├── bunker/                 # 🎯 DOMAIN: Bunker lifecycle
│   ├── manager.go          # BunkerManager (coordinator)
│   ├── create.go          # CreateBunker
│   ├── delete.go          # DeleteBunker + DeleteImage
│   ├── list.go            # ListBunkers + Info
│   ├── stop.go            # StopBunker
│   ├── prune.go           # PruneBunkers
│   └── helpers.go         # sanitize, formatBytes
│
├── build/                  # 🎯 DOMAIN: Image build
│   ├── manager.go          # BuildManager
│   ├── image.go           # BuildImage + RebuildImage
│   ├── steps.go           # Installation steps
│   ├── progress.go       # Progress rendering
│   └── gpu.go            # GPU detection & resolution
│
├── slots/                  # 🎯 DOMAIN: Slot system
│   ├── manager.go         # SlotManager
│   ├── registry.go       # Item discoverer
│   ├── engine.go         # Installation engine
│   ├── domain.go         # Domain models
│   ├── loader.go         # TOML loading
│   ├── base/             # Base system tools
│   ├── dev/              # DEV slots (1 file = 1 item)
│   │   ├── ia/           # Ollama, Opencode, Engram, Gentle
│   │   ├── languages/    # Go, Node.js, Python
│   │   └── tools/        # Starship
│   ├── data/             # DATA slots (DBs)
│   │   ├── postgres.go, mysql.go, mongodb.go, redis.go, sqlite.go
│   └── sandbox/          # SANDBOX slot (empty)
│
├── router/                 # CLI command router
│   └── router.go          # Handle() with 14 commands
│
└── adapters/              # 🔧 INFRASTRUCTURE
    ├── runtime/           # Podman/Distrobox
    │   ├── commands.go   # "podman", "distrobox" commands
    │   └── podman.go     # IBunkerRuntime adapter
    ├── filesystem/       # Filesystem
    │   └── local.go     # IFileSystem adapter
    ├── system/           # System, GPU, Config
    │   ├── install.go   # Installation
    │   ├── config.go    # TOML config
    │   └── gpu/gpu.go   # GPU detection
    └── ui/              # User interface
        ├── views/        # Presenter, form, confirm
        ├── styles/       # TUI styles
        ├── theme/        # Themes
        └── slots/        # Slot selector (Bubbletea)
```

---

## 🗑️ Legacy Scripts (Deprecated)

> **Scripts in `lib/` and `scripts/` are NO LONGER USED.** They remain only as historical reference for the refactor.

| Script | Status | Notes |
|--------|--------|-------|
| `lib/bunker_lifecycle.sh` | ❌ Deprecated | Functionality migrated to `internal/bunker/` |
| `lib/git.sh` | ❌ Deprecated | Git tools planned for future migration |
| `lib/env.sh` | ❌ Deprecated | Migrated to `adapters/system/config.go` |
| `lib/gpu.sh` | ❌ Deprecated | Migrated to `adapters/system/gpu/gpu.go` |
| `scripts/install.sh` | ❌ Deprecated | Migrated to `axiom init` (TUI) |

---

## ⚠️ The Golden Rules (To Avoid Breaking Anything)

If you are going to modify or add code to AXIOM, **you must strictly respect these rules**:

### 1. Dependencies Flow Inward ⬇️
```
cmd/ → internal/ → domain → ports → bunker/build → adapters
```
- `cmd/` can import anything.
- `internal/` can import from `domain/` and `ports/`.
- **`internal/` CANNOT import ANYTHING from `adapters`**. The Core is blind and does not know what type of interface is using it.

### 2. Using `fmt.Print` in the Core is Forbidden 🚫🖨️
The `internal/` layer **must never** print directly to the screen.
- **How do I show an error or message?** Use the `ports.IPresenter` interface (e.g., `presenter.ShowError(err)`).

### 3. Everything Communicates Through Domain Models 📦
Return domain structs (`[]domain.Bunker`), not formatted strings.
Formatting for the user is the job of `adapters/ui/`.

### 4. Avoid Hardcoded Paths 🗺️
Files like `opencode.json` or scripts should be in `configs/`, never hardcoded.

### 5. Each Slot Item = 1 Separate File 📦
Each installable item (ollama, engram, postgres, etc.) lives in its **own .go file**.
- No monolithic files with multiple installations
- Adding a new item = create a new file, not modifying an existing one

### 6. SlotManager is the Central Orchestrator 🎛️
The `SlotManager` (in `internal/slots/manager.go`) is the only one that knows all available slots.

---

## 🚀 Available Commands

| Command | Alias | Description | Status |
|---------|-------|-------------|--------|
| `create` | - | Create bunker (choose image) | ✅ |
| `delete` | `rm` | Delete bunker | ✅ |
| `list` | `ls` | List bunkers | ✅ |
| `stop` | - | Stop bunker | ✅ |
| `prune` | - | Clean orphans | ✅ |
| `info` | - | Bunker info | ✅ |
| `delete-image` | - | Delete image | ✅ |
| `build` | - | Build image with slots | ✅ |
| `rebuild` | - | Rebuild image | ⚠️ WIP |
| `init` | - | Init wizard | ✅ |
| `slots` | - | Show available slots | ✅ |
| `enter` | - | Enter bunker | ⚠️ Partial |
| `reset` | - | Total reset | ⚠️ WIP |
| `help` | `-h`, `--help` | Help | ✅ |

---

## 🛡️ Implemented Slots

| Category | Slot | Items |
|----------|------|-------|
| **DEV** | AI | Ollama, Opencode, Engram, Gentle-AI |
| **DEV** | Languages | Go, Node.js, Python, Rust |
| **DEV** | Tools | Starship |
| **DATA** | Databases | PostgreSQL, MySQL, MongoDB, Redis, SQLite |
| **SANDBOX** | Empty | Empty minimal image |

---

## 🧪 Tests

```bash
make test              # All tests with race detector
make test-unit         # Tests without race (faster)
make test-coverage     # With coverage report
```

### Coverage

| Package | Coverage |
|---------|----------|
| `adapters/filesystem` | ~89% |
| `adapters/runtime` | ~77% |
| `bunker` | ~70%+ |
| `build` | ~65%+ |

---

## 🛠️ How to Add New Functionality?

Follow this workflow to add a new command (e.g., `axiom snapshot`):

1.  **Domain:** Need a new data model? Add it to `internal/domain/`.
2.  **Ports:** Need Podman to do something new? Add it to `internal/ports/`.
3.  **Bunker/Build:** Create the logic in `internal/bunker/` or `internal/build/`.
4.  **Adapters:** Implement the call in `internal/adapters/runtime/`.
5.  **UI:** Design how it looks in `internal/adapters/ui/views/`.
6.  **Router:** Add the route in `internal/router/router.go`.
7.  **Main:** Ensure DI in `cmd/axiom/main.go`.

---

## 🎯 Switching from Podman to Docker

The goal is that switching from Podman to Docker should be **1 file + 1 variable** change.

### 1. `internal/adapters/runtime/commands.go`
```go
// THIS IS THE ONLY FILE with "podman", "distrobox" strings
var Podman = CommandSet{
    CreateBunker: func(name, image, home, flags string) []string {
        return []string{"distrobox-create", "--name", name, ...}
    },
}
```

### 2. `cmd/axiom/main.go`
```go
// Change this line to use Docker:
runtime = podman.NewPodmanAdapter(commands.Podman)
// runtime = docker.NewDockerAdapter(commands.Docker)
```
