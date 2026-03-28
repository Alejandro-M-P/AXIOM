# 🏛️ Arquitectura de AXIOM 2.0

AXIOM utiliza una **Arquitectura Limpia (Clean Architecture)** basada en el patrón de Puertos y Adaptadores.
El objetivo principal de esta estructura es **desacoplar completamente la lógica de negocio (Core) de la interfaz de usuario y del sistema subyacente (Podman, OS)**.

Esto permite que AXIOM sea altamente testeable, escalable y que pueda ser operado tanto por humanos (vía Terminal interactiva) como por Inteligencias Artificiales (vía Agent API JSON).

---

## 📂 Estructura de Directorios

Todo el código de negocio está en `internal/`. El punto de entrada está en `cmd/`.

```text
.
├── cmd/axiom/                    # Punto de entrada
│   ├── main.go                  # Bootstrap + DI
│   └── router_commands.go       # Routing de comandos → handlers
│
├── internal/                     # 🔒 Código privado de AXIOM
│   ├── domain/                  # 🧠 ENTIDADES (puro, sin deps)
│   │   ├── bunker.go            # Bunker, BunkerConfig
│   │   ├── image.go             # Image, BuildConfig
│   │   └── system.go            # GPUInfo, EnvConfig
│   │
│   ├── ports/                   # 🔌 CONTRATOS (interfaces)
│   │   ├── runtime.go           # IBunkerRuntime
│   │   ├── filesystem.go        # IFileSystem
│   │   ├── system.go            # ISystem
│   │   └── presenter.go         # IPresenter
│   │
│   ├── bunker/                  # 🎯 DOMINIO: Bunker lifecycle
│   │   ├── manager.go           # BunkerManager (coordinador)
│   │   ├── create.go            # CreateBunker
│   │   ├── delete.go            # DeleteBunker + DeleteImage
│   │   ├── list.go              # ListBunkers + Info
│   │   ├── stop.go              # StopBunker
│   │   ├── prune.go             # PruneBunkers (mutex-protected)
│   │   └── helpers.go           # sanitize, formatBytes, etc.
│   │
│   ├── build/                   # 🎯 DOMINIO: Image build
│   │   ├── manager.go           # BuildManager (coordinador)
│   │   ├── image.go             # BuildImage + RebuildImage
│   │   ├── steps.go             # installSystemBase, installDevTools...
│   │   ├── progress.go          # Progress rendering
│   │   └── gpu.go               # resolveBuildGPU, normalizeGPUType
│   │
│   └── adapters/                 # 🔧 INFRAESTRUCTURA
│       ├── runtime/             # Container runtime
│       │   ├── commands.go      # ⚠️ SOLO AQUÍ: "podman", "distrobox"
│       │   └── podman.go        # Adapter implements IBunkerRuntime
│       ├── filesystem/
│       │   └── local.go         # Adapter implements IFileSystem
│       ├── system/
│       │   ├── system.go        # Adapter implements ISystem
│       │   ├── config.go        # Config TOML
│       │   └── gpu.go           # GPU detection
│       └── ui/
│           ├── presenter.go      # Console presenter
│           ├── form.go          # Bubbletea forms
│           ├── styles/           # UI styling
│           └── i18n/            # Translations (es/, en/)
│
├── tests/                       # 🧪 Tests centralizados
│   ├── bunker/                 # Tests bunker
│   ├── build/                  # Tests build
│   ├── adapters/               # Tests adapters
│   ├── cmd/                    # Tests router
│   └── mocks/                  # Mocks compartidos
│
├── configs/                     # 📄 Templates y assets
│   ├── templates.go            # .bashrc, starship.toml
│   └── assets/                  # opencode.json
│
└── docs/                       # 📚 Documentación
```

---

## ⚠️ Las 4 Reglas de Oro (Para no romper nada)

Si vas a modificar o agregar código a AXIOM, **debes respetar estrictamente estas reglas**:

### 1. La Dependencia fluye hacia adentro ⬇️
```
cmd/ → internal/ → domain → ports → bunker/build → adapters
```
- `cmd/` puede importar cualquier cosa.
- `internal/` puede importar de `domain/` y `ports/`.
- **`internal/` NO PUEDE importar NADA de `adapters`**. El Core es ciego y no sabe qué tipo de interfaz lo está usando.

### 2. Prohibido usar `fmt.Print` en el Core 🚫🖨️
La capa `internal/` **jamás** debe imprimir en pantalla directamente.
- **¿Cómo muestro un error o mensaje?** Usa la interfaz `ports.IPresenter` (ej. `presenter.ShowError(err)`).

### 3. Todo se comunica mediante Modelos de Dominio 📦
Devuelve structs de dominio (`[]domain.Bunker`), no strings formateados.
El trabajo de formatear para el usuario es de `adapters/ui/`.

### 4. Evitar rutas *Hardcodeadas* 🗺️
Archivos como `opencode.json` o scripts deben estar en `configs/`, nunca inyectados rígidamente.

---

## 🎯 Cambiar de Podman a Docker

El objetivo es que cambiar de Podman a Docker sea modificar **1 archivo + 1 variable**.

### 1. `internal/adapters/runtime/commands.go`
```go
// ESTE ES EL ÚNICO ARCHIVO con strings "podman", "distrobox"
var Podman = CommandSet{
    CreateBunker: func(name, image, home, flags string) []string {
        return []string{"distrobox-create", "--name", name, ...}
    },
    // ...
}
```

### 2. `cmd/axiom/main.go`
```go
// Cambiar esta línea para usar Docker:
runtime = podman.NewPodmanAdapter(commands.Podman)
// runtime = docker.NewDockerAdapter(commands.Docker)
```

---

## 📁 Commands Runtime: El corazón del cambio

```
commands.go ──────→ podman.go ──────→ ports/runtime.go
   │                    │                    │
   │                    │                    │
strings         implementation        IBunkerRuntime
"podman",              │              (interface)
"distrobox"           ↓
              exec.CommandContext()
```

**Regla: Los strings "podman" y "distrobox" SOLO viven en `commands.go`**

---

## 🔄 Estado de Refactorización

### ✅ AXIOM 2.0 COMPLETO (2026-03-28)

| Fecha | Cambio | Descripción |
|-------|--------|-------------|
| 2026-03-28 | Estructura | Creada estructura `internal/` con domain/ports/bunker/build/adapters |
| 2026-03-28 | Runtime Abstraction | `IBunkerRuntime` con métodos semánticos |
| 2026-03-28 | Commands | `commands.go` único archivo con podman/distrobox |
| 2026-03-28 | Bunker Domain | Dividido en create, delete, list, stop, prune, helpers |
| 2026-03-28 | Build Domain | Dividido en image, steps, progress, gpu |
| 2026-03-28 | Router | `router_commands.go` con 12 comandos |
| 2026-03-28 | Tests | ~200+ tests con mocks, coverage 77-89% |
| 2026-03-28 | Cleanup | Eliminado `pkg/` legacy, `unit_tests/` |

---

## 🧪 Tests

```bash
make test              # Todos los tests con race detector
make test-unit         # Tests sin race (más rápido)
make test-coverage     # Con coverage report
```

### Coverage
| Paquete | Coverage |
|---------|----------|
| `adapters/filesystem` | 89.3% |
| `adapters/runtime` | 77.6% |

---

## 🚀 Comandos Disponibles

| Comando | Alias | Descripción |
|---------|-------|-------------|
| `create` | - | Crear un bunker |
| `delete` | `rm` | Eliminar bunker |
| `list` | `ls` | Listar bunkers |
| `stop` | - | Detener bunker |
| `prune` | - | Limpiar bunkers huérfanos |
| `build` | - | Construir imagen |
| `rebuild` | - | Reconstruir imagen |
| `info` | - | Info de bunker |
| `enter` | - | Entrar a bunker |
| `init` | - | Inicializar AXIOM |
| `help` | `-h`, `--help` | Mostrar ayuda |

---

## 🛡️ Slots System (AXIOM 2.0+)

| Slot | Propósito | Características |
|------|-----------|-----------------|
| **DEV** | Oficina de Programación | IA, lenguajes, agentes |
| **DATA** | Laboratorio de Persistencia | MySQL, Postgres, Mongo |
| **BOX** | Zona de Pruebas | Aislamiento, carpetas mapeadas |

---

## 📚 Documentación

- `docs/ARCHITECTURE.md` — Esta arquitectura
- `docs/GO_REFACTOR_GUIDE.md` — Guía para contribuyen
