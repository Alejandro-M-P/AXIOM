# Ports — Interface Definitions

This package defines **abstractions** (interfaces) that the business logic uses to communicate with the outside world.

## What goes here

- `IPresenter` — UI output
- `ISystem` — system commands (sudo, ssh, detection)
- `IBunkerRuntime` — container operations
- `IFileSystem` — file operations
- `ICommandRunner` — shell execution
- `Field`, `LifecycleStep` — UI data types

## Rules

- **No business logic** — only interfaces
- **No imports** from `adapters/` or `bunker/build/slots/`
- Interfaces are implemented by adapters

## See also

- `docs/GOLDEN_RULES.md` — Architecture rules
- `adapters/ui/` — Bubble Tea implementation
- `adapters/runtime/` — Podman implementation
