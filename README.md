[🇬🇧 English](README-en.md) | [🇪🇸 Español](README.md)

# AXIOM Bunker System 🛡️

```text
  █████╗ ██╗  ██╗██╗ ██████╗ ███╗   ███╗
 ██╔══██╗╚██╗██╔╝██║██╔═══██╗████╗ ████║
 ███████║ ╚███╔╝ ██║██║   ██║██╔████╔██║
 ██╔══██║ ██╔██╗ ██║██║   ██║██║╚██╔╝██║
 ██║  ██║██╔╝ ██╗██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝
```

> **Cero suciedad. Entornos inmutables. Desarrollo de alto rendimiento.**

**AXIOM** es un orquestador de entornos de desarrollo aislados (Bunkers) construido sobre **Distrobox** y **Podman**. Cada búnker es un contenedor Arch Linux independiente con acceso directo a GPU, un stack de IA local preconfigurado, y un prompt de Starship personalizado — sin tocar un solo archivo crítico de tu host.

Ideal para sistemas atómicos (Bazzite, Fedora Silverblue) o cualquier entorno donde mantener el host completamente limpio sea prioridad.

---

## 🏗️ Arquitectura

AXIOM usa **Clean Architecture**. Código en `internal/`.

### Estructura

```
cmd/axiom/                  # Entry point
├── main.go
└── router_commands.go

internal/
├── domain/                 # Entidades
├── ports/                  # Interfaces
├── bunker/                 # Ciclo de vida
├── build/                  # Construcción imágenes
├── slots/                  # Slots (DEV/DATA/SANDBOX)
├── router/                 # CLI router
└── adapters/              # Infra (Podman, FS, UI)
```

---

## 🚀 Comandos

| Comando | Estado | Notas |
|---------|--------|-------|
| `axiom create` | ✅ | Pide imagen y nombre |
| `axiom delete` | ✅ | |
| `axiom list` | ✅ | |
| `axiom stop` | ✅ | |
| `axiom prune` | ✅ | |
| `axiom info` | ✅ | |
| `axiom delete-image` | ✅ | |
| `axiom build` | 🚧 | Wizard para seleccionar slots, construcción de imágenes |
| `axiom rebuild` | ❌ | No implementado |
| `axiom reset` | ❌ | No implementado |
| `axiom enter` | ❌ | No implementado |
| `axiom init` | ✅ | Wizard TUI |
| `axiom slots` | ✅ | |
| `axiom help` | ✅ | |

---

## 🛡️ Slots Disponibles

| Categoría | Items |
|-----------|-------|
| **DEV IA** | Ollama, Opencode, Engram, Gentle-AI |
| **DEV Languages** | Go, Python, Node.js, Rust |
| **DEV Tools** | Starship |
| **DATA** | PostgreSQL, MySQL, MongoDB, Redis, SQLite |
| **SANDBOX** | Empty |

---

## 🛠️ Instalación

```bash
git clone https://github.com/Alejandro-M-P/AXIOM.git ~/AXIOM
cd ~/AXIOM

go build -o axiom ./cmd/axiom
./axiom init
./axiom build
./axiom create mi-proyecto
```

---

## 🧪 Tests

```bash
make test
```

- `internal/bunker`: ✅ Pasan
- `adapters/filesystem`: ✅ Pasan
- `adapters/runtime`: ✅ Pasan
- `build`: ⚠️ 1 test fallando

---

## 🔒 Seguridad

- Tokens en volúmenes de solo lectura
- Path sanitization (`filepath.Clean()`)
- Permisos `0700`
- Arrays para comandos

---

## ⚠️ Bugs Conocidos

- `axiom create` no muestra las opciones de imagen antes de pedir nombre
- `axiom enter` no implementado
- `axiom rebuild` / `axiom reset` no implementados
- Init wizard con bugs en rutas relativas

---

## 📖 Filosofía

Tu host es sagrado. AXIOM es el lab donde instalás, rompé y probá sin consecuencias.
