 # 🏛️ Arquitectura de AXIOM

AXIOM utiliza una **Arquitectura Limpia (Clean Architecture)** basada en el patrón de Puertos y Adaptadores. 
El objetivo principal de esta estructura es **desacoplar completamente la lógica de negocio (Core) de la interfaz de usuario y del sistema subyacente (Podman, OS)**.

Esto permite que AXIOM sea altamente testeable, escalable y que pueda ser operado tanto por humanos (vía Terminal interactiva) como por Inteligencias Artificiales (vía Agent API JSON).

---

## 📂 Estructura de Directorios

Todo el código fuente público y reutilizable se encuentra en `pkg/`. El punto de entrada está en `cmd/`.

```text
.
├── cmd/axiom/                  # Punto de entrada. Inicializa e inyecta dependencias.
├── pkg/
│   ├── core/                   # 🧠 CAPA 1: Lógica de Negocio Pura
│   │   ├── domain/             # Entidades (Structs puros: Bunker, Config).
│   │   ├── ports/              # Interfaces que dictan cómo interactuar con el exterior.
│   │   └── services/           # Reglas de negocio (Crear, Validar, Borrar).
│   ├── controller/             # 🕹️ CAPA 2: El Orquestador
│   │   └── orchestrator.go     # Recibe comandos del main, consulta y delega al Core.
│   └── adapters/               # 🔌 CAPA 3: Adaptadores (Mundo Exterior)
│       ├── system/             # Interacciones con OS, GPUs e instalaciones.
│       ├── podman/             # Ejecución de comandos de contenedores.
│       ├── ui/                 # Interfaz humana (Terminal gráfica, colores, i18n).
│       └── api/                # Interfaz de máquina (Salida JSON para agentes IA).
├── configs/assets/             # Archivos estáticos de configuración (opencode, starship).
└── scripts/legacy_bash/        # Scripts heredados en proceso de migración.
```

---

## ⚠️ Las 4 Reglas de Oro (Para no romper nada)

Si vas a modificar o agregar código a AXIOM, **debes respetar estrictamente estas reglas**:

### 1. La Dependencia fluye hacia adentro ⬇️
- `cmd/` puede importar cualquier cosa.
- `pkg/adapters/` puede importar de `pkg/core/` y `pkg/controller/`.
- `pkg/controller/` puede importar de `pkg/core/`.
- **`pkg/core/` NO PUEDE importar NADA de `adapters`, `controller` ni `cmd`.** El Core es ciego y no sabe qué tipo de interfaz lo está usando.

### 2. Prohibido usar `fmt.Print` en el Core 🚫🖨️
La capa `pkg/core/` o `pkg/controller/` **jamás** debe imprimir en pantalla directamente (nada de `fmt.Println` ni `log.Fatal`). 
- **¿Cómo muestro un error o mensaje?** Debes usar la interfaz `ports.IPresenter` (ej. `presenter.ShowError(err)`). Así, si estamos en modo consola, se dibujará en rojo; pero si estamos en modo Agente IA, saldrá en un JSON limpio.

### 3. Todo se comunica mediante Modelos de Dominio 📦
Si creas un servicio que lista ranuras (Slots), no devuelvas un bloque de texto o strings formateados. Devuelve un `[]domain.Slot`. Será trabajo de la capa `adapters/ui/` convertir ese array en una tabla colorida.

### 4. Evitar rutas *Hardcodeadas* 🗺️
Archivos como `opencode.json` o scripts deben leerse desde el directorio del sistema que corresponda o desde los directorios configurados (ej. `configs/assets/`), nunca inyectados rígidamente en la lógica del Core.

---

## 🛠️ ¿Cómo agregar una nueva funcionalidad?

Sigue este flujo de trabajo para añadir un nuevo comando (Ej: `axiom snapshot`):

1. **Core (Domain):** ¿Necesitamos un nuevo modelo de datos? (Ej. `type Snapshot struct`). Añádelo a `pkg/core/domain/`.
2. **Core (Ports):** ¿Necesitamos que Podman haga algo nuevo? Añade la función a la interfaz `IPodman` en `pkg/core/ports/`.
3. **Core (Services):** Crea la lógica pura en `pkg/core/services/snapshot.go` que reciba los puertos e implemente las validaciones.
4. **Adapters:** Implementa la llamada real a Podman en `pkg/adapters/podman/` y diseña cómo se verá en pantalla en `pkg/adapters/ui/views/`.
5. **Controller:** Agrega la ruta en el Orquestador (`pkg/controller/`) para que cuando el usuario escriba `axiom snapshot`, este conecte el Servicio con el Adaptador.
6. **Main:** Asegúrate de que las dependencias estén correctamente inyectadas en `cmd/axiom/main.go`.

---

## 🔄 Estado de Refactorización (Clean Architecture)

Esta sección documenta el progreso de la migración a Clean Architecture.

### ✅ Completado

| Fecha | Cambio | Descripción |
|-------|--------|-------------|
| 2026-03-27 | Phase 1 | Creado `pkg/core/domain/` con modelos puros |
| 2026-03-27 | Phase 1 | Creado `pkg/core/ports/` con interfaces |
| 2026-03-27 | Phase 2 | Creado `pkg/adapters/podman/adapter.go` |
| 2026-03-27 | Phase 2 | Creado `pkg/adapters/fs/adapter.go` |
| 2026-03-27 | Phase 2 | Reorganizado `pkg/adapters/system/` (gpu en subdir) |
| 2026-03-27 | Phase 3.1 | Creado `pkg/core/services/manager.go` con inyección |
| 2026-03-27 | Phase 3.2 | Migrado `runCommandQuiet` → `m.Runtime.RunCommand()` |
| 2026-03-27 | Phase 3.2 | Migrado `runCommandWithInput` → `m.Runtime.RunCommandWithInput()` |
| 2026-03-27 | Phase 3.2 | Migrado `runCommandOutput` → `m.Runtime.RunCommandOutput()` |
| 2026-03-27 | Phase 3.2 | Migrado `distroboxExists` → `m.Runtime.ContainerExists()` |
| 2026-03-27 | Phase 3.2 | Migrado `podmanImageExists` → `m.Runtime.ImageExists()` |
| 2026-03-27 | Phase 3.2 | Migrado `bunkerStatus`, `listBunkerNames`, `listAxiomImages` |

### 🔄 En Progreso

- Phase 3: Migración Grupo 2 (funciones FS: removePathWritable, ensureTutorFile, etc.)

### 📋 Pendiente

- Phase 3: Migrar funciones FS (removePathWritable, ensureTutorFile, appendTutorLog, etc.)
- Phase 3: Migrar funciones System (hostGPUVolumeFlags, sshVolumeFlag, prepareSSHAgent, resolveBuildGPU)
- Phase 4: Integration (main.go)
- Phase 5: Verification (testear comandos)
- Phase 6: Cleanup

### 📁 Estructura Actual

```
pkg/
├── core/
│   ├── domain/models.go          ✅ Modelos puros
│   └── ports/                    ✅ Interfaces
│       ├── podman.go            ✅ IContainerRuntime
│       ├── filesystem.go        ✅ IFileSystem
│       ├── system.go            ✅ ISystem
│       └── presenter.go          ✅ IPresenter
├── adapters/
│   ├── podman/adapter.go        ✅ PodmanAdapter
│   ├── fs/adapter.go           ✅ FSAdapter
│   └── system/
│       ├── system.go            ✅ SystemAdapter
│       └── gpu/gpu.go          ✅ DetectGPU (movido)
└── services/
    ├── manager.go               ✅ Manager con inyección
    └── instance.go              🔄 Migrando...
```