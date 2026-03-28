# Skill Registry — AXIOM

> **Auto-generated** — Do not edit manually. Run skill-registry to update.

## Project Information

| Property | Value |
|----------|-------|
| **Project** | AXIOM |
| **Language** | Go |
| **Type** | CLI Tool |
| **Generated** | 2026-03-28 |

## SDD Skills Available

| Skill | Purpose | Triggers |
|-------|---------|----------|
| `sdd-init` | Initialize SDD context in project | `sdd init`, project setup |
| `sdd-explore` | Research/investigate before committing | Exploration phase, unknown territory |
| `sdd-propose` | Create change proposal | Intent, scope, approach definition |
| `sdd-design` | Technical design document | Architecture decisions |
| `sdd-spec` | Write specifications with scenarios | Requirements definition |
| `sdd-tasks` | Break down change into tasks | Implementation planning |
| `sdd-apply` | Implement tasks from change | Code implementation |
| `sdd-verify` | Validate implementation | Verification against specs |
| `sdd-archive` | Archive completed change | Sync and cleanup |

## Stack-Specific Skills

| Skill | Purpose | Triggers |
|-------|---------|----------|
| `go-testing` | Go testing patterns, Bubbletea TUI testing | Writing Go tests, teatest |
| `issue-creation` | Create GitHub issues following issue-first system | Bug reports, feature requests |
| `branch-pr` | PR creation workflow | Opening PRs, preparing changes for review |
| `skill-creator` | Create new AI agent skills | Documenting patterns for AI |

## Project Conventions

### Architecture
- **Pattern**: Clean Architecture (Hexagonal)
- **Structure**:
  - `pkg/core/domain/` — Pure business models
  - `pkg/core/ports/` — Interface contracts
  - `pkg/core/services/` — Business logic
  - `pkg/adapters/` — Concrete implementations

### Technology Stack
- **Language**: Go 1.24.2
- **TUI Framework**: Bubbletea (Charmbracelet)
- **Runtime Dependencies**: Podman, Distrobox
- **Configuration**: TOML, .env files

### Testing Conventions
- **Location**: `unit_tests/` directory (separate from source)
- **Pattern**: Table-driven tests
- **Mocking**: Custom mock implementations
- **Naming**: `*_test.go`

### Code Style
- **Interfaces**: Prefix with `I` (e.g., `IContainerRuntime`)
- **Packages**: Descriptive names (`podman`, `fs`, `system`)
- **Comments**: Spanish language, descriptive
- **DI Pattern**: Constructor injection in services

## SDD Configuration

**Persistence Mode**: `engram`

No `openspec/` directory created. All SDD artifacts persisted to Engram memory.

## Related Documentation

- Main README: `README.md`
- English README: `README-en.md`
- Project docs: `docs/`

---

## How to Use

1. **Start exploration**: `/sdd-explore <topic>`
2. **Create change**: `/sdd-propose <change-name>`
3. **Follow the flow**: propose → design → spec → tasks → apply → verify → archive

For details, see individual skill documentation.
