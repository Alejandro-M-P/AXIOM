# Slots — Installable Items

This package contains the **slot system** for installable tools and packages.

## What goes here

- Slot loading from TOML files
- Slot installation logic
- Slot categories (dev, ai, data, sandbox)

## Rules (per GOLDEN_RULES.md)

- Uses `ports/` for I/O abstractions
- **No exec.Command** — uses ports.ICommandRunner
- TOML-based (no code changes for new slots)

## See also

- `docs/GOLDEN_RULES.md` — Architecture rules
- `slots/*.toml` — Slot definitions
- `slots/base/` — Base infrastructure
