# 🚀 Resumen del Refactor Arquitectónico (Go)

[🇪🇸 Español](#-español) | [🇬🇧 English](#-english)

---

## 🇪🇸 Español

Este hito representa la transición de AXIOM desde un prototipo funcional a una aplicación con una **arquitectura de software profesional, limpia y escalable**. El objetivo ha sido desacoplar por completo la lógica de negocio (el "Core") de la capa de presentación (la "UI"), preparándola para una fácil traducción y futuras interfaces (web, API, etc.).

### Logros Clave

1.  **Arquitectura Hexagonal (Puertos y Adaptadores)**:
    *   Se ha definido una interfaz `bunker.UI` que actúa como un "puerto" de comunicación.
    *   El Core (`pkg/bunker`) ahora es 100% agnóstico a la UI. No contiene `fmt.Println`, ni colores, ni sabe qué es una terminal. Solo emite datos puros y eventos a través de la interfaz.
    *   Se ha creado un "adaptador" (`pkg/ui/presenter.go`) que implementa dicha interfaz y se encarga de "traducir" los eventos del Core en componentes visuales para la terminal usando `bubbletea`.

2.  **Política de "Cero Strings" (Internacionalización)**:
    *   Se han erradicado **todos** los textos, frases, emojis y mensajes de error del código fuente de Go.
    *   Se ha creado una estructura de localización en `pkg/ui/locales/es/` que contiene archivos `.toml` descriptivos (`commands.toml`, `errors.toml`, `prompts.toml`, `lifecycle.toml`, `logs.toml`).
    *   El sistema ahora carga dinámicamente todos los textos desde estos archivos, permitiendo una futura traducción a otros idiomas simplemente creando una nueva carpeta (ej. `locales/en/`).

3.  **UI Interactiva y Segura**:
    *   Se han eliminado todas las lecturas de `stdin` (`bufio`) del Core.
    *   Las confirmaciones críticas (`delete`, `reset`, `rebuild`, `prune`) ya no son un simple texto `(s/N)`. Ahora se presentan en **tarjetas interactivas multipaso** que unifican todo el flujo de preguntas (confirmación, razón técnica, opciones adicionales) en un único componente visual.
    *   Las operaciones destructivas utilizan un estilo de "peligro" (borde rojo) para alertar visualmente al usuario.

4.  **Limpieza de Código y Dependencias**:
    *   Se ha eliminado el arte ASCII del logo del código Go, que ahora se carga desde un archivo `logo.txt` mediante `//go:embed`.
    *   Se ha migrado la configuración de internacionalización de JSON a TOML para mejorar la legibilidad y permitir comentarios.

5.  **Arquitectura de Catálogo Dinámico y Ecosistema Modular**:
    *   Se introduce un `catalog.toml` que define un catálogo de herramientas instalables (IA, DBs, DevTools), eliminando las instalaciones "hardcoded" del binario.
    *   Se crea el concepto de "Slots" (`DEV`, `DATA`, `RANDOM`) con comportamientos de instalación diferenciados según el propósito del búnker.
    *   El comando `build` evoluciona de un script monolítico a un motor de aprovisionamiento dinámico que lee las selecciones del usuario (guardadas en un `state.json`) y ejecuta los instaladores definidos en el catálogo.
    *   La TUI ahora genera las opciones de instalación dinámicamente desde el catálogo, permitiendo que el ecosistema de AXIOM crezca y se modifique sin necesidad de recompilar el programa principal.
    *   Esta arquitectura resuelve el problema de fondo del acoplamiento lógico, haciendo que AXIOM sea verdaderamente modular y extensible.



---

## 🇬🇧 English

This milestone represents AXIOM's transition from a functional prototype to an application with a **professional, clean, and scalable software architecture**. The main goal was to completely decouple the business logic (the "Core") from the presentation layer (the "UI"), preparing it for easy translation and future interfaces (web, API, etc.).

### Key Achievements

1.  **Hexagonal Architecture (Ports and Adapters)**:
    *   A `bunker.UI` interface has been defined, acting as a communication "port".
    *   The Core (`pkg/bunker`) is now 100% UI-agnostic. It contains no `fmt.Println`, no colors, and has no knowledge of what a terminal is. It only emits pure data and events through the interface.
    *   A "presenter" adapter (`pkg/ui/presenter.go`) has been created, which implements this interface and is responsible for "translating" Core events into visual components for the terminal using `bubbletea`.

2.  **"Zero Strings" Policy (Internationalization)**:
    *   **All** texts, phrases, emojis, and error messages have been eradicated from the Go source code.
    *   A localization structure has been created in `pkg/ui/locales/es/` containing descriptive `.toml` files (`commands.toml`, `errors.toml`, `prompts.toml`, `lifecycle.toml`, `logs.toml`).
    *   The system now dynamically loads all texts from these files, allowing for future translation into other languages simply by creating a new folder (e.g., `locales/en/`).

3.  **Interactive and Secure UI**:
    *   All `stdin` reads (`bufio`) have been removed from the Core.
    *   Critical confirmations (`delete`, `reset`, `rebuild`, `prune`) are no longer simple `(y/N)` text prompts. They are now presented in **interactive multi-step cards** that unify the entire question flow (confirmation, technical reason, additional options) into a single visual component.
    *   Destructive operations use a "danger" style (red border) to visually alert the user.

4.  **Code and Dependency Cleanup**:
    *   The ASCII logo art has been removed from the Go code and is now loaded from a `logo.txt` file using `//go:embed`.
    *   The internationalization configuration has been migrated from JSON to TOML to improve readability and allow for comments.

5.  **Dynamic Catalog Architecture and Modular Ecosystem**:
    *   A `catalog.toml` file is introduced to define a catalog of installable tools (AI, DBs, DevTools), eliminating hardcoded installations from the binary.
    *   The concept of "Slots" (`DEV`, `DATA`, `RANDOM`) is created, featuring distinct installation behaviors based on the bunker's purpose.
    *   The `build` command evolves from a monolithic script into a dynamic provisioning engine that reads user selections (stored in a `state.json` file) and executes the installers defined in the catalog.
    *   The TUI now dynamically generates installation options from the catalog, allowing the AXIOM ecosystem to grow and be modified without recompiling the main program.
    *   This architecture resolves the underlying problem of logical coupling, making AXIOM truly modular and extensible.


5.  **Dynamic Catalog Architecture and Modular Ecosystem**:
    *   A `catalog.toml` file is introduced to define a catalog of installable tools (AI, DBs, DevTools), eliminating hardcoded installations from the binary.
    *   The concept of "Slots" (`DEV`, `DATA`, `RANDOM`) is created, featuring distinct installation behaviors based on the bunker's purpose.
    *   The `build` command evolves from a monolithic script into a dynamic provisioning engine that reads user selections (stored in a `state.json` file) and executes the installers defined in the catalog.
    *   The TUI now dynamically generates installation options from the catalog, allowing the AXIOM ecosystem to grow and be modified without recompiling the main program.
    *   This architecture resolves the underlying problem of logical coupling, making AXIOM truly modular and extensible.