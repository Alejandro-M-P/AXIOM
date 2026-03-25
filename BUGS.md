# 🐛 Bug Tracker & MVP Status

[🇪🇸 Español](#-español) | [🇬🇧 English](#-english)

---

## 🇪🇸 Español
Este documento detalla los fallos conocidos y tareas pendientes durante la migración de Bash a **Go Nativo**.

## 🔴 Críticos y 
- [x] **Seguridad de Volúmenes**: Las funciones nativas de `build` y `create` necesitan auditoría en el manejo de permisos de contenedores rootless.Seguridad
- [x] **Inyección de GPU**: Fallos esporádicos al detectar y montar drivers NVIDIA/AMD desde el binario de Go en `build`.
- [x] **Persistencia**: El contenido de `~/.entorno/` a veces no se monta correctamente en el primer inicio.
- [ ] **[ID-BUG-001] Path Traversal**: Vulnerabilidad crítica en `Create`, `Delete` e `Info` al concatenar el nombre del búnker directamente en las rutas sin usar `filepath.Clean()`.
- [x] **[ID-BUG-005] Permisos Laxos**: `os.MkdirAll` crea carpetas en `.entorno/` con `0755` en lugar del restrictivo `0700`, exponiendo configuraciones a otros usuarios del host.

## 🟡 Pendientes de Migración (Bash ➔ Go)
- [ ] Traducir herramientas de Git interactivo (`lib/git.sh`) a código nativo en `pkg/`.
- [ ] Migrar lógica de limpieza profunda (`axiom purge`).
- [ ] **[ID-BUG-002] Dependencia de utilidades del sistema**: Reemplazar comandos heredados de bash como `du -sh` (en `bunkerEnvSize`) usando `filepath.WalkDir` nativo.
- [ ] **[ID-BUG-003] Parseo frágil**: Dejar de separar texto con `strings.Split` en `distroboxExists` y usar `--format json` decodificando con `encoding/json`.
- [ ] **[ID-BUG-004] Faltan Goroutines**: Tareas masivas como `Prune` borran los entornos secuencialmente. Deben refactorizarse con `sync.WaitGroup` para borrado en paralelo.

## 🟠 Arquitectura y Separación de Responsabilidades
- [ ] **[ID-BUG-006] Acoplamiento Lógica-UI**: Las funciones del `Manager` (`Create`, `Delete`, `Info`) imprimen tarjetas y logs directamente a la consola, mezclando la lógica de negocio con la presentación.
- [ ] **[ID-BUG-008] Interacción con Usuario en el Core**: Funciones como `Delete` y `Prune` leen desde `stdin` para pedir confirmación, bloqueando su uso en scripts y violando la separación de capas.
- [ ] **[ID-BUG-009] Responsabilidades de UI en el Manager**: El `Manager` tiene un método `Help()` cuya única función es imprimir una tarjeta de ayuda, una tarea que corresponde a la capa de UI.
- [ ] **[ID-BUG-007] Dependencia de Git en PATH**: La función `bunkerGitBranch` ejecuta el binario `git` en lugar de usar una librería nativa como `go-git`, creando una dependencia externa frágil.
- [ ] **[ID-BUG-010] Secuencias ANSI Hardcodeadas**: Limpiar la pantalla de la terminal con `fmt.Print("\033[H\033[2J")` dentro de las funciones Core destruye la portabilidad de las llamadas.
- [ ] **[ID-BUG-014] Split Brain Arquitectónico (main.go)**: Bypass del modelo BubbleTea (`tea.Program`) para inyectar lógica interactiva manual (`bufio`) en comandos directos, impidiendo escalar el UI.
- [ ] **[ID-BUG-016] Violación DRY en Ruteo**: Duplicidad de lista de comandos (`Run` vs `KnownCommand`) en `bunkerController.go`. Debe migrarse a mapas de funciones o Cobra/CLI.

## 🟡 Calidad de Código y Estándares Go
- [ ] **[ID-BUG-011] Supresión de Errores (Silencing)**: Múltiples ejecuciones de `exec.Command` descartan silenciosamente el error o el `Stderr` asignándolo a `_`, ocultando fallos críticos (ej. en `prepareSSHAgent`).
- [ ] **[ID-BUG-012] Errores Desnudos (Sin Wrapping)**: Devolver `return err` sin contexto (ej. tras `os.MkdirAll`) dificulta el rastro. Se debe usar `fmt.Errorf("... %w", err)`.
- [ ] **[ID-BUG-013] Parseo Manual del `.env`**: El método en `LoadEnvFile` que utiliza `strings.Cut(line, "=")` truncará secretos o tokens que contengan un igual (`=`) en su valor.
- [ ] **[ID-BUG-017] Resolución Mágica de Variables**: `prepareSSHAgent` confía en `os.Getenv("HOME")` en lugar de usar el robusto y seguro `os.UserHomeDir()`.

## 🟢 Concurrencia y Control de Ejecución
- [ ] **[ID-BUG-015] Procesos Zombis sin Contexto**: Llamadas a binarios con `exec.Command` (como `pacman`) no están limitadas. Usar `exec.CommandContext` para evitar cuelgues permanentes.
- [ ] **[ID-BUG-018] Cuelgues por Prompts Sudo**: `runCommandQuiet("sudo", "-v")` puede dejar la app colgada si pide contraseña sin tener TTY correctamente bindeado.

## 🔵 Próximos Pasos
- [ ] **Configuración**: Mover el sistema de variables de `.env` a un archivo `config.toml` profesional.
- [ ] **Manejo de Errores**: Crear tipos de error personalizados (ej. `ErrBunkerNotFound`, `ErrImageMissing`) en lugar de usar strings genéricos con `fmt.Errorf`.
- [ ] **Integración Podman API**: Investigar la conexión directa al socket REST de Podman en lugar de ejecutar subprocesos de shell.

---

## 🇬🇧 English
This document details the known issues and pending tasks during the migration from Bash to **Native Go**.

### 🔴 Critical & Security
- [x] **Volume Security**: Native `build` and `create` functions need auditing regarding rootless container permission handling.
- [x] **GPU Injection**: Sporadic failures when detecting and mounting NVIDIA/AMD drivers from the Go binary in `build`.
- [x] **Persistence**: `~/.entorno/` contents sometimes fail to mount correctly on the first startup.
- [ ] **[ID-BUG-001] Path Traversal**: Critical vulnerability in `Create`, `Delete`, and `Info` where bunker names are concatenated directly into paths without using `filepath.Clean()`.
- [x] **[ID-BUG-005] Loose Permissions**: `os.MkdirAll` creates folders in `.entorno/` with `0755` instead of the restrictive `0700`, exposing configs to other host users.

### 🟡 Pending Migration (Bash ➔ Go)
- [ ] Translate interactive Git tools (`lib/git.sh`) to native code in `pkg/`.
- [ ] Migrate deep clean logic (`axiom purge`).
- [ ] **[ID-BUG-002] System Utility Dependency**: Replace legacy bash commands like `du -sh` (in `bunkerEnvSize`) using native `filepath.WalkDir`.
- [ ] **[ID-BUG-003] Fragile Parsing**: Stop splitting text with `strings.Split` in `distroboxExists` and use `--format json` decoded with `encoding/json`.
- [ ] **[ID-BUG-004] Missing Goroutines**: Massive tasks like `Prune` delete environments sequentially. They must be refactored with `sync.WaitGroup` for parallel deletion.

### 🟠 Architecture & Separation of Concerns
- [ ] **[ID-BUG-006] Logic-UI Coupling**: `Manager` functions (`Create`, `Delete`, `Info`) print cards and logs directly to the console, mixing business logic with presentation.
- [ ] **[ID-BUG-008] User Interaction in Core**: Functions like `Delete` and `Prune` read from `stdin` to ask for confirmation, blocking their use in scripts and violating layer separation.
- [ ] **[ID-BUG-009] UI Responsibilities in Manager**: The `Manager` has a `Help()` method whose sole function is printing a help card, a task belonging to the UI layer.
- [ ] **[ID-BUG-007] Git PATH Dependency**: The `bunkerGitBranch` function executes the `git` binary instead of using a native library like `go-git`, creating a fragile external dependency.
- [ ] **[ID-BUG-010] Hardcoded ANSI Sequences**: Clearing the terminal screen with `fmt.Print("\033[H\033[2J")` inside Core functions destroys the portability of the calls.
- [ ] **[ID-BUG-014] Architectural Split Brain (main.go)**: Bypassing the BubbleTea model (`tea.Program`) to inject manual interactive logic (`bufio`) in direct commands, preventing UI scaling.
- [ ] **[ID-BUG-016] DRY Violation in Routing**: Duplication of command lists (`Run` vs `KnownCommand`) in `bunkerController.go`. Needs migration to function maps or Cobra/CLI.

### 🟡 Code Quality & Go Standards
- [ ] **[ID-BUG-011] Error Silencing**: Multiple `exec.Command` executions silently discard the error or `Stderr` by assigning it to `_`, hiding critical failures (e.g., in `prepareSSHAgent`).
- [ ] **[ID-BUG-012] Naked Errors (No Wrapping)**: Returning `return err` without context (e.g., after `os.MkdirAll`) makes tracing difficult. `fmt.Errorf("... %w", err)` should be used.
- [ ] **[ID-BUG-013] Manual `.env` Parsing**: The method in `LoadEnvFile` using `strings.Cut(line, "=")` will truncate secrets or tokens containing an equals sign (`=`) in their value.
- [ ] **[ID-BUG-017] Magic Variable Resolution**: `prepareSSHAgent` relies on `os.Getenv("HOME")` instead of using the robust and secure `os.UserHomeDir()`.

### 🟢 Concurrency & Execution Control
- [ ] **[ID-BUG-015] Zombie Processes without Context**: Calls to binaries with `exec.Command` (like `pacman`) are unbounded. Use `exec.CommandContext` to prevent permanent hangs.
- [ ] **[ID-BUG-018] Hangs from Sudo Prompts**: `runCommandQuiet("sudo", "-v")` can leave the app hanging if it asks for a password without a correctly bound TTY.

### 🔵 Next Steps
- [ ] **Configuration**: Move the variables system from `.env` to a professional `config.toml` file.
- [ ] **Error Handling**: Create custom error types (e.g., `ErrBunkerNotFound`, `ErrImageMissing`) instead of using generic strings with `fmt.Errorf`.
- [ ] **Podman API Integration**: Investigate direct connection to the Podman REST socket instead of executing shell subprocesses.
