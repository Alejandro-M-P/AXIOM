# 🐛 Bug Tracker

Este documento rastrea issues conocidos. La migración a Go está **en proceso** (muchas features aún pendientes).

---

## ✅ Resueltos

### 🔴 Critical & Security (Resueltos)
- ✅ **Volume Security**: Auditado y corregido en la nueva implementación Go.
- ✅ **GPU Injection**: Implementado correctamente en `adapters/system/gpu/`.
- ✅ **Persistence**: Corregido con mount automático de `.entorno/`.
- ✅ **[ID-BUG-001] Path Traversal**: Implementado `filepath.Clean()` en todos los paths.
- ✅ **[ID-BUG-005] Loose Permissions**: Ahora se usa `0700` en todos los directorios creados.

### 🟠 Architecture (Resueltos)
- ✅ **[ID-BUG-003] Fragile Parsing**: Ahora usa JSON decode con `encoding/json`.
- ✅ **[ID-BUG-004] Missing Goroutines**: Prune ahora usa `sync.WaitGroup`.
- ✅ **[ID-BUG-006] Logic-UI Coupling**: UI desacoplada via `ports.IPresenter`.
- ✅ **[ID-BUG-008] User Interaction in Core**: Confirmaciones via `adapters/ui/views/confirm.go`.
- ✅ **[ID-BUG-009] UI Responsibilities in Manager**: Help movido a UI layer.
- ✅ **[ID-BUG-010] Hardcoded ANSI Sequences**: Eliminado, ahora usa lipgloss/bubbletea.
- ✅ **[ID-BUG-014] Architectural Split Brain**: Router centralizado en `internal/router/`.

---

## 🔴 Critical & Security (Pendientes)

- **[ID-BUG-032] Init creates structure in wrong path**: When the user inputs a relative path like `Documentos/dev` during `axiom init`, the directory structure is created in `$HOME` instead of the resolved absolute path.
- **[ID-BUG-019] Deleting image inside cloned bunker**: Deleting an image while inside a cloned bunker deletes the cloned bunker.
- **[ID-BUG-020] Gentle-ai execution**: When entering the bunker, `gentle-ai` does not execute to trigger the installer (should check installation status via a flag file or native go).

---

## 🐛 Bugs Actuales (2026-03-29)

- **`axiom create`**: No muestra de qué imagen/slot se va a crear el bunker — debe mostrar las opciones (axiom-dev, axiom-data, axiom-sandbox) antes de pedir el nombre.
- **`axiom create`**: El selector de imagen no muestra descripción de cada slot.
- **`axiom build`**: Necesita verificar que las imágenes existan antes de permitir crear bunkers.

---

## 🟡 Pending Tasks

### Bash → Go Migration (Legacy Scripts)
- Translate interactive Git tools (`lib/git.sh`) to native code — planned for future.
- Migrate deep clean logic (`axiom purge`).
- **[ID-BUG-002] System Utility Dependency**: Replace legacy bash commands like `du -sh` using native `filepath.WalkDir`.

---

## 🟠 Architecture & Separation of Concerns (Pendientes)

- **[ID-BUG-007] Git PATH Dependency**: Use `go-git` library instead of executing `git` binary.
- **[ID-BUG-016] DRY Violation in Routing**: Command lists duplication — needs refactor to function maps or Cobra/CLI.

---

## 🟡 Code Quality & Go Standards (Pendientes)

- **[ID-BUG-011] Error Silencing**: Multiple `exec.Command` executions silently discard errors.
- **[ID-BUG-012] Naked Errors**: Need more error wrapping with `fmt.Errorf("... %w", err)`.
- **[ID-BUG-013] Manual `.env` Parsing**: `strings.Cut(line, "=")` truncates secrets with `=` in value.
- **[ID-BUG-017] Magic Variable Resolution**: Use `os.UserHomeDir()` instead of `os.Getenv("HOME")`.

---

## 🟢 Concurrency & Execution Control (Pendientes)

- **[ID-BUG-015] Zombie Processes**: Use `exec.CommandContext` to prevent permanent hangs.
- **[ID-BUG-028] Race condition in Prune**: `ShowLog` called from multiple goroutines via `sync.WaitGroup` is not goroutine-safe.

---

## 🔵 UI & i18n (Pendientes)

- **[ID-BUG-029] i18n keys rendered raw**: Translation keys passed directly to lipgloss without `GetText()`.
- **[ID-BUG-030] GetText only handles 2-level keys**: Keys with 3+ levels fall through.
- **[ID-BUG-031] Starship not found on --rcfile entry**: PATH not initialized before `.bashrc` is sourced.

---

## 🔵 Next Steps (Planeados)

- **Podman API Integration**: Investigate direct connection to the Podman REST socket instead of shell subprocesses.
- **Custom Error Types**: Create `ErrBunkerNotFound`, `ErrImageMissing` etc.
- **Config TOML**: Ya implementado parcialmente en `adapters/system/config.go`.

---

## 📝 Notas

- Los scripts en `lib/` y `scripts/` son **deprecated** y ya no se usan.
- La estructura cambió de `pkg/` a `internal/`.
- Algunos bugs referencian rutas antiguas (`pkg/`) que ahora son `internal/`.
