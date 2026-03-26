# 🏛️ AXIOM Architecture

AXIOM uses a **Clean Architecture** based on the Ports and Adapters pattern.
The main goal of this structure is to **completely decouple the business logic (Core) from the user interface and the underlying system (Podman, OS)**.

This allows AXIOM to be highly testable, scalable, and operable by both humans (via an interactive Terminal) and Artificial Intelligences (via a JSON Agent API).

---

## 📂 Directory Structure

All public and reusable source code is located in `pkg/`. The entry point is in `cmd/`.

```text
.
├── cmd/axiom/                  # Entry point. Initializes and injects dependencies.
├── pkg/
│   ├── core/                   # 🧠 LAYER 1: Pure Business Logic
│   │   ├── domain/             # Entities (Pure Structs: Bunker, Config).
│   │   ├── ports/              # Interfaces that dictate how to interact with the outside.
│   │   └── services/           # Business rules (Create, Validate, Delete).
│   ├── controller/             # 🕹️ LAYER 2: The Orchestrator
│   │   └── orchestrator.go     # Receives commands from main, queries, and delegates to the Core.
│   └── adapters/               # 🔌 LAYER 3: Adapters (Outside World)
│       ├── system/             # Interactions with OS, GPUs, and installations.
│       ├── podman/             # Execution of container commands.
│       ├── ui/                 # Human interface (Graphical terminal, colors, i18n).
│       └── api/                # Machine interface (JSON output for AI agents).
├── configs/assets/             # Static configuration files (opencode, starship).
└── scripts/legacy_bash/        # Legacy scripts undergoing migration.
```

---

## ⚠️ The 4 Golden Rules (To Avoid Breaking Anything)

If you are going to modify or add code to AXIOM, **you must strictly respect these rules**:

### 1. Dependencies Flow Inward ⬇️
- `cmd/` can import anything.
- `pkg/adapters/` can import from `pkg/core/` and `pkg/controller/`.
- `pkg/controller/` can import from `pkg/core/`.
- **`pkg/core/` CANNOT import ANYTHING from `adapters`, `controller`, or `cmd`.** The Core is blind and does not know what type of interface is using it.

### 2. Using `fmt.Print` in the Core is Forbidden 🚫🖨️
The `pkg/core/` or `pkg/controller/` layer **must never** print directly to the screen (no `fmt.Println` or `log.Fatal`).
- **How do I show an error or message?** You must use the `ports.IPresenter` interface (e.g., `presenter.ShowError(err)`). This way, if we are in console mode, it will be drawn in red; but if we are in AI Agent mode, it will be output as clean JSON.

### 3. Everything Communicates Through Domain Models 📦
If you create a service that lists slots, do not return a block of text or formatted strings. Return a `[]domain.Slot`. It will be the job of the `adapters/ui/` layer to convert that array into a colorful table.

### 4. Avoid Hardcoded Paths 🗺️
Files like `opencode.json` or scripts should be read from the corresponding system directory or from configured directories (e.g., `configs/assets/`), never rigidly injected into the Core's logic.

---

## 🛠️ How to Add New Functionality?

Follow this workflow to add a new command (e.g., `axiom snapshot`):

1.  **Core (Domain):** Do we need a new data model? (e.g., `type Snapshot struct`). Add it to `pkg/core/domain/`.
2.  **Core (Ports):** Do we need Podman to do something new? Add the function to the `IPodman` interface in `pkg/core/ports/`.
3.  **Core (Services):** Create the pure logic in `pkg/core/services/snapshot.go` that receives the ports and implements the validations.
4.  **Adapters:** Implement the actual call to Podman in `pkg/adapters/podman/` and design how it will look on screen in `pkg/adapters/ui/views/`.
5.  **Controller:** Add the route in the Orchestrator (`pkg/controller/`) so that when the user types `axiom snapshot`, it connects the Service with the Adapter.
6.  **Main:** Make sure the dependencies are correctly injected in `cmd/axiom/main.go`.