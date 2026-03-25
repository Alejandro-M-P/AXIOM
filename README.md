[🇬🇧 English](README-en.md) | [🇪🇸 Español](README.md)

# AXIOM Bunker System 🛡️ (Refactor Go Edition)

```text
  █████╗ ██╗  ██╗██╗ ██████╗ ███╗   ███╗
 ██╔══██╗╚██╗██╔╝██║██╔═══██╗████╗ ████║
 ███████║ ╚███╔╝ ██║██║   ██║██╔████╔██║
 ██╔══██║ ██╔██╗ ██║██║   ██║██║╚██╔╝██║
 ██║  ██║██╔╝ ██╗██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝
```

> **Cero suciedad. Entornos inmutables. Desarrollo de alto rendimiento.**



**AXIOM** es un orquestador de entornos de desarrollo aislados (Bunkers) diseñado para mantener tu sistema operativo host impecable. Esta rama contiene la refactorización integral a Go, buscando eliminar la fragilidad de los scripts de Shell y pasar a un binario estático, seguro y con gestión de concurrencia nativa.

Ideal para sistemas atómicos (Bazzite, Fedora Silverblue) o cualquier entorno donde la limpieza del sistema operativo base sea la prioridad absoluta.

---

## 🏗️ Estado de la Migración a Go (WIP)

Actualmente AXIOM se encuentra en una transición profunda de scripts en Bash (`lib/*.sh`) a un binario robusto compilado en Go. El objetivo es eliminar la fragilidad de depender de `os/exec`, manejar estados de forma estricta y proporcionar una base para una futura API o TUI compleja.

### 🗺️ Hoja de Ruta Arquitectónica (Visión)
- **De Shell a Nativo**: Reemplazar llamadas a utilidades CLI (`du`, `grep`, `awk`) por librerías nativas de Go (`filepath.WalkDir`, `encoding/json`, `go-git`).
- **Desacoplamiento UI / Core**: Extraer la lógica interactiva (prints directos, lectura de `stdin` con `bufio`) del Core (`Manager`) hacia una capa de presentación externa (BubbleTea / Cobra).
- **Orquestación Avanzada**: Abandonar la ejecución de comandos `podman` por subprocesos y conectar AXIOM directamente al **socket REST API de Podman**.
- **Gestión de Configuración**: Dejar atrás el frágil parseo manual de `.env` a favor de un estándar profesional usando `config.toml`.
- **Concurrencia Segura**: Implementar `Goroutines` controladas con `sync.WaitGroup` para tareas masivas (como `axiom prune` en paralelo) y usar `context` para evitar procesos zombis en el sistema.

### Mapa de Funcionalidades

| Comando | Estado | Notas de Migración |
| :--- | :--- | :--- |
| `axiom list` | ✅ Portado | Falta refactorizar para usar decodificación JSON en vez de parseo manual de strings. |
| `axiom info` | ✅ Portado | Pendiente blindar contra vulnerabilidades de *Path Traversal* validando rutas. |
| `axiom build` | 🚧 En proceso | Inyección de drivers GPU. Pendiente de auditar bloqueos en prompts de `sudo`. |
| `axiom create` | 🚧 En proceso | Ajustando persistencia del `$HOME` con permisos estrictamente aislados (`0700`). |
| `axiom delete` | 🚧 En proceso | Funciona, pero requiere extraer los bloqueos interactivos (`stdin`) fuera del Core. |
| `axiom prune` | ⏳ Pendiente | Requiere reescribirse usando concurrencia nativa para borrado en paralelo. |
| `axiom purge` | ⏳ Pendiente | Lógica de limpieza profunda aún no portada de Bash. |
| **Git Tools** | ⏳ Pendiente | Reemplazar los flujos interactivos de `lib/git.sh` por llamadas a `go-git`. |



## 🔒 Seguridad y Arquitectura (AXIOM Vault en Go)

El núcleo en pkg/bunker abandona la ejecución procedimental por un modelo de Manager de Estados:
- Aislamiento de Secretos: Los tokens se inyectan en /run/axiom/env como volúmenes de solo lectura (Vault).
- Prevención de Vulnerabilidades: Auditoría activa contra *Path Traversal* (`filepath.Clean`) y permisos laxos al crear directorios (asegurando `os.MkdirAll` con `0700`).
- Sanitización de Ejecución: Los comandos externos (mientras se migran a API nativas) se lanzan mediante arrays tipados en lugar de evaluar strings, bloqueando inyecciones.
- Control de Zombis: Migración hacia `exec.CommandContext` para abortar subprocesos colgados automáticamente.

---

## 🛡️ Herramientas Internas (Dentro del Búnker)

Cada búnker inyecta automáticamente un ecosistema optimizado:

### IA Local & Contexto
- open: Lanza el editor opencode vinculado a la aceleración por hardware.
- sync-agents: Sincroniza las directivas de tutor.md con los agentes locales.
- Contexto Persistente: Las reglas de IA viven en el host y se heredan entre búnkeres mediante volúmenes compartidos.

### Git Interactivo
Comandos visuales  que eliminan la carga cognitiva:
- status: Diff visual en tiempo real.
- commit: Selección visual de archivos con <Tab> antes de confirmar.
- branch / switch: Navegación visual fluida entre ramas.

---

## 🧠 Stack de IA Incluido
- Ollama: Inferencia de modelos de lenguaje en local.
- Opencode: IDE optimizado con integración nativa de agentes.
- Agent Teams: Coordinación multiactente para flujos de trabajo profesionales.
- gentle-ai: tareas 
---

## 🛠️ Instalación y Desarrollo (Rama Refactor)

Requisitos: Go 1.21+, Podman, Distrobox.

1. Clonar la rama de desarrollo:
git clone https://github.com/Alejandro-M-P/AXIOM.git -b refactor/go_refactor
cd AXIOM

2. Compilar el orquestador:
go build -o axiom main.go


3. Verificar estado actual:
./axiom list

---

## 📖 Filosofía AXIOM

Tu host es sagrado. AXIOM es el laboratorio donde instalas, rompes y pruebas sin consecuencias para el sistema base. Si un stack de IA corrompe el entorno, borras el búnker y recreas uno nuevo en 30 segundos. El código fuente es lo único permanente; el entorno es desechable.
