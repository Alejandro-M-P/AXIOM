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
│   ├── slots/                    # 🎯 DOMINIO: Sistema de slots instalables
│   │   ├── manager.go           # SlotManager (coordinador central)
│   │   ├── registry.go          # Descubre items disponibles
│   │   ├── engine.go            # Ejecuta instalaciones ordenadas
│   │   ├── domain.go            # Modelos: SlotItem, SlotSelection
│   │   ├── dev/                 # Items DEV (1 archivo = 1 item)
│   │   │   ├── ia/              # Herramientas IA
│   │   │   │   ├── ollama.go    # Ollama LLM runtime
│   │   │   │   ├── opencode.go  # opencode-ai
│   │   │   │   ├── engram.go    # engram memoria IA
│   │   │   │   └── gentle.go    # gentle-ai
│   │   │   ├── languages/       # Programming languages
│   │   │   │   ├── go.go        # Go
│   │   │   │   ├── nodejs.go    # Node.js + npm
│   │   │   │   └── python.go    # Python
│   │   │   └── tools/           # Herramientas varias
│   │   │       └── starship.go  # starship prompt
│   │   ├── data/                # Items DATA (1 archivo = 1 item)
│   │   │   ├── postgres.go      # PostgreSQL
│   │   │   ├── mysql.go         # MySQL
│   │   │   ├── mongodb.go       # MongoDB
│   │   │   ├── redis.go         # Redis
│   │   │   └── sqlite.go        # SQLite
│   │   └── sandbox/             # Items SANDBOX
│   │       └── empty.go         # Imagen mínima sin instalaciones
│   │
│   └── adapters/ui/slots/       # 🎯 TUI: Slot selector (Bubbletea)
│       ├── slot_selector.go     # Componente TUI multi-select
│       └── slot_adapter.go      # Bridge domain → UI
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
│   ├── assets/                  # opencode.json
│   │   └── available.toml     # Comandos de instalación de slots
│   └── slots/                  # (deprecated - usar i18n/locales)
│
├── internal/adapters/ui/views/i18n/  # 🌐 Traducciones
│   └── locales/
│       ├── en/
│       │   ├── available.toml  # Nombres y descripciones EN
│       │   ├── commands.toml
│       │   ├── errors.toml
│       │   ├── logs.toml
│       │   └── prompts.toml
│       └── es/
│           ├── available.toml  # Nombres y descripciones ES
│           ├── commands.toml
│           ├── errors.toml
│           ├── logs.toml
│           └── prompts.toml
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

### 5. Cada Slot Item = 1 archivo separado 📦
Cada item instalable (ollama, engram, postgres, etc.) vive en su **propio archivo .go**.
- No hay archivos monolíticos con múltiples instalaciones
- Agregar un nuevo item = crear un archivo nuevo, no modificar uno existente
- El `SlotManager` coordina pero cada item se instala a sí mismo

### 6. El SlotManager es el orquestador central 🎛️
El `SlotManager` (en `internal/slots/manager.go`) es el único que conoce todos los slots disponibles.
- `BuildManager` delega en `SlotManager`
- `SlotManager` usa `Registry` para descubrir items
- `SlotManager` usa `Engine` para ejecutar en orden
- **Ningún item sabe del manager ni de otros items**

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

## 🗑️ Scripts Legacy (Deprecated)

> **Los scripts en `lib/` y `scripts/` ya NO se usan.** Quedan únicamente como referencia histórica para la refactorización.

| Script | Estado | Notas |
|--------|--------|-------|
| `lib/bunker_lifecycle.sh` | ❌ Deprecated | Funcionalidad migrada a `internal/bunker/` |
| `lib/git.sh` | ❌ Deprecated | Git tools planeados para migración futura |
| `lib/env.sh` | ❌ Deprecated | Migrado a `adapters/system/config.go` |
| `lib/gpu.sh` | ❌ Deprecated | Migrado a `adapters/system/gpu/gpu.go` |
| `scripts/install.sh` | ❌ Deprecated | Migrado a `axiom init` (TUI) |

---

## 🔄 Estado de Implementación

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

### 🔄 AXIOM Slots System (Completado)

| Fecha | Cambio | Descripción |
|-------|--------|-------------|
| 2026-03-29 | Slots Core | SlotManager, Registry, Engine implementados |
| 2026-03-29 | DEV Items | IA (ollama, opencode, engram, gentle), Languages (go, nodejs, python), Tools (starship) |
| 2026-03-29 | DATA Items | postgres, mysql, mongodb, redis, sqlite |
| 2026-03-29 | SANDBOX | empty.go — imagen mínima |
| 2026-03-29 | TUI Selector | Bubbletea slot selector con multi-select |
| 2026-03-29 | Create Flow | Selección de imagen (axiom-dev, axiom-data, axiom-sandbox) |
| 2026-03-29 | i18n | Traducciones EN/ES para slots en `i18n/locales/` |

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

| Comando | Alias | Descripción | Estado |
|---------|-------|-------------|--------|
| `create` | - | Crear bunker (elige imagen) | ✅ |
| `delete` | `rm` | Eliminar bunker | ✅ |
| `list` | `ls` | Listar bunkers | ✅ |
| `stop` | - | Detener bunker | ✅ |
| `prune` | - | Limpiar huérfanos | ✅ |
| `info` | - | Info de bunker | ✅ |
| `delete-image` | - | Eliminar imagen | ✅ |
| `build` | - | Construir imagen con slots | ✅ |
| `rebuild` | - | Reconstruir imagen | ⚠️ WIP |
| `init` | - | Wizard de inicialización | ✅ |
| `slots` | - | Mostrar slots disponibles | ✅ |
| `enter` | - | Entrar a bunker | ⚠️ Parcial |
| `reset` | - | Reset total | ⚠️ WIP |
| `help` | `-h`, `--help` | Mostrar ayuda | ✅ |

---

## 🛡️ Slots System (AXIOM 2.0+)

Cada slot es un **grupo de items instalables** seleccionado por el usuario.

### Imágenes generadas

| Slot | Imagen | Descripción |
|------|--------|-------------|
| **DEV** | `axiom-dev` | IA tools + lenguajes + starship |
| **DATA** | `axiom-data` | Bases de datos seleccionadas |
| **SANDBOX** | `axiom-sandbox` | Imagen mínima sin instalaciones |

### Flujo `axiom build`

```
axiom build
    ↓
BuildManager → Check SlotManager.HasSelection()
    ↓
Si no hay selección → Mostrar TUI Slot Selector (Bubbletea)
    ↓
Usuario selecciona items con checkboxes (Space= toggle, Enter= confirmar)
    ↓
SlotManager.SaveSelection() → guarda en memoria
    ↓
Engine.Execute(items) → instala en orden de dependencias
    ↓
runtime.CommitImage() → genera imagen (axiom-dev/data/sandbox)
```

### Flujo `axiom create`

```
axiom create
    ↓
Pregunta: "¿Qué imagen querés usar?"
    ├── axiom-dev   (DEV slot)
    ├── axiom-data  (DATA slot)
    └── axiom-sandbox (SANDBOX)
    ↓
Usuario elige (1, 2, 3 o nombre)
    ↓
Pregunta: "¿Nombre del bunker?"
    ↓
Crea bunker con la imagen seleccionada
```

### Estructura de un Item

```go
// internal/slots/dev/ia/ollama.go
type Ollama struct{}

func (s *Ollama) ID() string          { return "ollama" }
func (s *Ollama) Name() string        { return "Ollama" }
func (s *Ollama) Description() string { return "Ejecuta modelos LLM localmente" }
func (s *Ollama) Category() SlotCategory { return SlotDEV }
func (s *Ollama) SubCategory() string { return "ia" }  // ia, languages, tools
func (s *Ollama) Dependencies() []string { return []string{} }

func (s *Ollama) Install(ctx context.Context, exec Executor) error {
    // curl -fsSL https://ollama.com/download/ollama-linux-amd64.tar.zst
    // tar -xzf -C /usr
    return nil
}
```

### Slots Disponibles

| Slot | Categoría | Items |
|------|-----------|-------|
| **DEV** | IA | opencode, engram, gentle-ai, ollama |
| **DEV** | Languages | go, nodejs, python, rust |
| **DEV** | Tools | starship |
| **DATA** | Databases | postgres, mysql, mongodb, redis, sqlite |
| **SANDBOX** | — | empty — imagen mínima |

---

## 📚 Documentación

- `docs/ARCHITECTURE.md` — Esta arquitectura
- `docs/GO_REFACTOR_GUIDE.md` — Guía para contribuyen
