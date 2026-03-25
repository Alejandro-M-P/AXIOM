[🇬🇧 English](README-en.md) | [🇪🇸 Español](README.md)

 # AXIOM Bunker System 🛡️ (Refactor Go Edition)

  █████╗ ██╗  ██╗██╗ ██████╗ ███╗   ███╗
 ██╔══██╗╚██╗██╔╝██║██╔═══██╗████╗ ████║
 ███████║ ╚███╔╝ ██║██║   ██║██╔████╔██║
 ██╔══██║ ██╔██╗ ██║██║   ██║██║╚██╔╝██║
 ██║  ██║██╔╝ ██╗██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝

> **Cero suciedad. Entornos inmutables. Desarrollo de alto rendimiento.**



**AXIOM** es un orquestador de entornos de desarrollo aislados (Bunkers) diseñado para mantener tu sistema operativo host impecable. Esta rama contiene la refactorización integral a Go, buscando eliminar la fragilidad de los scripts de Shell y pasar a un binario estático, seguro y con gestión de concurrencia nativa.

Ideal para sistemas atómicos (Bazzite, Fedora Silverblue) o cualquier entorno donde la limpieza del sistema operativo base sea la prioridad absoluta.

---

## 🏗️ Estado Actual del Refactor (WIP)

Esta rama es un laboratorio activo. La lógica se está moviendo de Bash a Go para ganar en estabilidad y velocidad de ejecución.

### Mapa de Funcionalidades

| Comando | Estado | Observaciones |
| :--- | :--- | :--- |
| axiom list | Terminado | Lista búnkeres con tamaño, estado y rama git actual. |
| axiom info | Terminado | Ficha técnica detallada del contenedor y hardware. |
| axiom build | bugs | Implementando la inyección dinámica de drivers GPU. |
| axiom create | Bugs | Problemas conocidos con la persistencia del $HOME. |
| axiom delete |  bugs | Funciona el borrado básico; falta el selector visual interactivo. |
| axiom purge | bugs  | Lógica de limpieza profunda aún no portada de Bash. |



## 🔒 Seguridad y Arquitectura (Go Core)

El núcleo en pkg/bunker abandona la ejecución procedimental por un modelo de Manager de Estados:
- Seguridad de Tipos: Validación estricta de rutas y configuraciones antes de tocar el socket de Podman.
- Sanitización de Ejecución: Los comandos externos se lanzan mediante slices de argumentos tipados, eliminando riesgos de inyección.
- Aislamiento de Secretos: Los tokens se inyectan en /run/axiom/env como volúmenes de solo lectura (Vault).

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
