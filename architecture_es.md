# 🏛️ Arquitectura de AXIOM

AXIOM utiliza una **Arquitectura Limpia** basada en el patrón Puertos y Adaptadores.
El objetivo principal de esta estructura es **desacoplar completamente la lógica de negocio (Core) de la interfaz de usuario y el sistema subyacente (Podman, OS)**.

Esto permite que AXIOM sea altamente testeable, escalable y operable tanto por humanos (a través de una Terminal interactiva) como por Inteligencias Artificiales (a través de una API JSON de Agentes).

---

## 📂 Estructura de Directorios

Todo el código de negocio está en `internal/`. El punto de entrada está en `cmd/`.

```
cmd/axiom/                  # Punto de entrada
├── main.go                 # Bootstrap + DI
└── router_commands.go      # Re-exportaciones del router

internal/                   # 🔒 Código privado de AXIOM
├── domain/                 # 🧠 ENTIDADES (puras, sin dependencias)
│   ├── bunker.go          # Bunker, BunkerConfig
│   ├── system.go          # GPUInfo, EnvConfig
│   └── slot.go            # SlotItem, SlotSelection
│
├── ports/                  # 🔌 CONTRATOS (interfaces)
│   ├── runtime.go          # IBunkerRuntime
│   ├── filesystem.go      # IFileSystem
│   ├── system.go          # ISystem
│   └── presenter.go       # IPresenter
│
├── bunker/                 # 🎯 DOMINIO: Ciclo de vida del Bunker
│   ├── manager.go          # BunkerManager (coordinador)
│   ├── create.go          # CreateBunker
│   ├── delete.go          # DeleteBunker + DeleteImage
│   ├── list.go            # ListBunkers + Info
│   ├── stop.go            # StopBunker
│   ├── prune.go           # PruneBunkers
│   └── helpers.go         # sanitize, formatBytes
│
├── build/                  # 🎯 DOMINIO: Construcción de imágenes
│   ├── manager.go          # BuildManager
│   ├── image.go           # BuildImage + RebuildImage
│   ├── steps.go           # Pasos de instalación
│   ├── progress.go       # Renderizado de progreso
│   └── gpu.go            # Detección y resolución de GPU
│
├── slots/                  # 🎯 DOMINIO: Sistema de Slots
│   ├── manager.go         # SlotManager
│   ├── registry.go       # Descubridor de ítems
│   ├── engine.go         # Motor de instalación
│   ├── domain.go         # Modelos de dominio
│   ├── loader.go         # Carga TOML
│   ├── base/             # Herramientas del sistema base
│   ├── dev/              # Slots DEV (1 archivo = 1 ítem)
│   │   ├── ia/           # Ollama, Opencode, Engram, Gentle
│   │   ├── languages/    # Go, Node.js, Python
│   │   └── tools/        # Starship
│   ├── data/             # Slots DATA (BDs)
│   │   ├── postgres.go, mysql.go, mongodb.go, redis.go, sqlite.go
│   └── sandbox/          # Slot SANDBOX (vacío)
│
├── router/                 # Enrutador de comandos CLI
│   └── router.go          # Handle() con 14 comandos
│
└── adapters/              # 🔧 INFRAESTRUCTURA
    ├── runtime/           # Podman/Distrobox
    │   ├── commands.go   # comandos "podman", "distrobox"
    │   └── podman.go     # adaptador IBunkerRuntime
    ├── filesystem/       # Sistema de archivos
    │   └── local.go     # adaptador IFileSystem
    ├── system/           # Sistema, GPU, Config
    │   ├── install.go   # Instalación
    │   ├── config.go    # Configuración TOML
    │   └── gpu/gpu.go   # Detección de GPU
    └── ui/              # Interfaz de usuario
        ├── views/        # Presentador, formulario, confirmación
        ├── styles/       # Estilos TUI
        ├── theme/        # Temas
        └── slots/        # Selector de slots (Bubbletea)
```

---

## 🗑️ Scripts Legados (Obsoletos)

> **Los scripts en `lib/` y `scripts/` YA NO SE UTILIZAN.** Permanecen únicamente como referencia histórica para el refactor.

| Script | Estado | Notas |
|--------|--------|-------|
| `lib/bunker_lifecycle.sh` | ❌ Obsoleto | Funcionalidad migrada a `internal/bunker/` |
| `lib/git.sh` | ❌ Obsoleto | Herramientas git planeadas para futura migración |
| `lib/env.sh` | ❌ Obsoleto | Migrado a `adapters/system/config.go` |
| `lib/gpu.sh` | ❌ Obsoleto | Migrado a `adapters/system/gpu/gpu.go` |
| `scripts/install.sh` | ❌ Obsoleto | Migrado a `axiom init` (TUI) |

---

## ⚠️ Las Reglas Doradas (Para Evitar Romper Algo)

Si vas a modificar o agregar código a AXIOM, **debes respetar estrictamente estas reglas**:

### 1. El flujo de dependencias va hacia adentro ⬇️
```
cmd/ → internal/ → domain → ports → bunker/build → adapters
```
- `cmd/` puede importar cualquier cosa.
- `internal/` puede importar de `domain/` y `ports/`.
- **`internal/` NO PUEDE importar NADA de `adapters`**. El núcleo está ciego y no sabe qué tipo de interfaz lo está utilizando.

### 2. Usar `fmt.Print` en el Núcleo está Prohibido 🚫🖨️
La capa `internal/` **nunca debe** imprimir directamente en la pantalla.
- **¿Cómo muestro un error o mensaje?** Usa la interfaz `ports.IPresenter` (p.ej., `presenter.ShowError(err)`).

### 3. Todo se Comunica a Través de Modelos de Dominio 📦
Devuelve structs de dominio (`[]domain.Bunker`), no strings formateados.
El formateo para el usuario es responsabilidad de `adapters/ui/`.

### 4. Evita Rutas Hardcodeadas 🗺️
Archivos como `opencode.json` o scripts deben estar en `configs/`, nunca hardcodeados.

### 5. Cada Ítem de Slot = 1 Archivo Separado 📦
Cada elemento instalable (ollama, engram, postgres, etc.) vive en su **propio archivo .go**.
- No hay archivos monolíticos con múltiples instalaciones
- Agregar un nuevo elemento = crear un nuevo archivo, no modificar uno existente

### 6. SlotManager es el Orquestrador Central 🎛️
El `SlotManager` (en `internal/slots/manager.go`) es el único que conoce todos los slots disponibles.

---

## 🚀 Comandos Disponibles

| Comando | Alias | Descripción | Estado |
|---------|-------|-------------|--------|
| `create` | - | Crear bunker (elige imagen) | ✅ |
| `delete` | `rm` | Eliminar bunker | ✅ |
| `list` | `ls` | Listar bunkers | ✅ |
| `stop` | - | Detener bunker | ✅ |
| `prune` | - | Limpiar huérfanos | ✅ |
| `info` | - | Información del bunker | ✅ |
| `delete-image` | - | Eliminar imagen | ✅ |
| `build` | - | Construir imagen con slots | ✅ |
| `rebuild` | - | Reconstruir imagen | ⚠️ WIP |
| `init` | - | Asistente de inicialización | ✅ |
| `slots` | - | Mostrar slots disponibles | ✅ |
| `enter` | - | Entrar al bunker | ⚠️ Parcial |
| `reset` | - | Reinicio total | ⚠️ WIP |
| `help` | `-h`, `--help` | Ayuda | ✅ |

---

## 🛡️ Slots Implementados

| Categoría | Slot | Ítems |
|-----------|------|-------|
| **DEV** | IA | Ollama, Opencode, Engram, Gentle-AI |
| **DEV** | Lenguajes | Go, Node.js, Python, Rust |
| **DEV** | Herramientas | Starship |
| **DATA** | Bases de datos | PostgreSQL, MySQL, MongoDB, Redis, SQLite |
| **SANDBOX** | Vacío | Imagen mínima vacía |

---

## 🧪 Pruebas

```bash
make test              # Todas las pruebas con detector de carrera
make test-unit         # Pruebas sin carrera (más rápido)
make test-coverage     # Con reporte de cobertura
```

### Cobertura

| Paquete | Cobertura |
|---------|-----------|
| `adapters/filesystem` | ~89% |
| `adapters/runtime` | ~77% |
| `bunker` | ~70%+ |
| `build` | ~65%+ |

---

## 🛠️ Cómo Agregar Nueva Funcionalidad?

Sigue este flujo para agregar un nuevo comando (p.ej., `axiom snapshot`):

1.  **Dominio:** ¿Necesitas un nuevo modelo de datos? Agregalo a `internal/domain/`.
2.  **Puertos:** ¿Necesitas que Podman haga algo nuevo? Agrégalo a `internal/ports/`.
3.  **Bunker/Construcción:** Crea la lógica en `internal/bunker/` o `internal/build/`.
4.  **Adaptadores:** Implementa la llamada en `internal/adapters/runtime/`.
5.  **Interfaz:** Diseña cómo se ve en `internal/adapters/ui/views/`.
6.  **Router:** Agrega la ruta en `internal/router/router.go`.
7.  **Principal:** Asegura la inyección de dependencias en `cmd/axiom/main.go`.

---

## 🎯 Cambiando de Podman a Docker

El objetivo es que cambiar de Podman a Docker debería ser **un archivo + una variable** de cambio.

### 1. `internal/adapters/runtime/commands.go`
```go
// ESTE ES EL ÚNICO ARCHIVO con strings "podman", "distrobox"
var Podman = CommandSet{
    CreateBunker: func(name, image, home, flags string) []string {
        return []string{"distrobox-create", "--name", name, ...}
    },
}
```

### 2. `cmd/axiom/main.go`
```go
// Cambia esta línea para usar Docker:
runtime = podman.NewPodmanAdapter(commands.Podman)
// runtime = docker.NewDockerAdapter(commands.Docker)
``` 