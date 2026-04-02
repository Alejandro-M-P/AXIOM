# Adapters — Implementations

This package contains **concrete implementations** of the interfaces defined in `ports/`.

## Subdirectories

- `ui/` — Bubble Tea UI implementation
- `runtime/` — Podman/Distrobox implementation
- `filesystem/` — Local filesystem implementation
- `system/` — System commands (GPU detection, sudo, etc.)

## Rules (per GOLDEN_RULES.md)

- **Only place** where `exec.Command` is allowed
- **Only place** where `os.Getenv` is allowed (besides config)
- Implements ports interfaces

## See also

- `docs/GOLDEN_RULES.md` — Architecture rules
- `ports/` — Interface definitions
- `adapters/runtime/commands.go` — System commands (podman, distrobox)
