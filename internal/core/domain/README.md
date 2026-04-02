# Domain — Pure Domain Models

This package contains **pure domain types** with no external dependencies.

## What goes here

- Entities: `Bunker`, `Slot`, etc.
- Value objects: simple structs with no behavior
- Domain interfaces (if any)

## Rules

- **No imports** from `ports/`, `adapters/`, `router/`, or `config/`
- **No exec.Command**, **no os.Getenv**, **no fmt.Print**
- Only Go standard library + primitives

## See also

- `docs/GOLDEN_RULES.md` — Architecture rules
- `domain/bunker.go` — Bunker entity
- `domain/slot.go` — Slot interface
