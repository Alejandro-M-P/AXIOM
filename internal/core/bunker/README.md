# Bunker — Business Logic

This package contains **business logic** for bunker operations.

## What goes here

- Bunker creation, deletion, listing
- Business rules and validation
- Orchestration of bunker operations

## Rules (per GOLDEN_RULES.md)

- Uses `ports/` for I/O abstractions
- **No exec.Command** — uses ports.IBunkerRuntime
- **No fmt.Print** — uses ports.IPresenter
- **No os.Getenv** — uses ports.ISystem

## See also

- `docs/GOLDEN_RULES.md` — Architecture rules
- `adapters/runtime/` — Podman implementation
- `ports/runtime.go` — IBunkerRuntime interface
