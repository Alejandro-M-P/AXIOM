# Build — Business Logic

This package contains **business logic** for image building operations.

## What goes here

- Build context preparation
- GPU resolution and normalization
- Container creation and image export
- Build orchestration

## Rules (per GOLDEN_RULES.md)

- Uses `ports/` for I/O abstractions
- **No exec.Command** — uses ports.IBunkerRuntime
- **No fmt.Print** — uses ports.IPresenter
- **No os.Getenv** — uses ports.ISystem

## See also

- `docs/GOLDEN_RULES.md` — Architecture rules
- `adapters/runtime/` — Podman implementation
- `ports/runtime.go` — IBunkerRuntime interface
