# 🗺️ Guía de Ruta: Refactorización a Go Nativo

Este documento es tu brújula. Detalla el orden lógico para implementar la nueva **Arquitectura de Catálogo** y la **Migración Completa a Go**, evitando que te bloquees por dependencias circulares.

---

## 📌 FASE 1: El Cerebro del Catálogo (Prioridad Alta)
*El objetivo de esta fase es que AXIOM entienda qué puede instalar sin tenerlo hardcodeado.*

1. **Crear `catalog.toml` (Raíz del proyecto):**
   - Diseñar el archivo con las categorías `[Categorias.IA]`, `[Categorias.DB]`, etc.
   - Añadir herramientas de prueba (Ollama, Go, PostgreSQL).

2. **Escribir el Parser (`pkg/catalog/parser.go`):**
   - Implementar las estructuras de Go (`Item`, `Category`, `Catalog`).
   - Usar `github.com/BurntSushi/toml` o similar para leer el `catalog.toml`.
   - Crear función `LoadCatalog(path string)` que devuelva los datos estructurados.

3. **Diseñar el Gestor de Estado (`pkg/bunker/state.go`):**
   - Lógica para leer y escribir el `state.json` dentro de cada `.entorno/`.
   - Debe registrar: ¿Qué Slot es? (DEV, DATA, RANDOM), ¿Qué herramientas tiene instaladas?

---

## 📌 FASE 2: Interfaz de Usuario (TUI) Dinámica
*Hacer que la terminal muestre las opciones leídas en la Fase 1.*

1. **Adaptador de Formulario (BubbleTea):**
   - Modificar la TUI de `axiom build` para que reciba las `Categories` del parser.
   - Generar dinámicamente listas de checkboxes basándose en los `Items` del catálogo.

2. **Flujo de Selección:**
   - Pantalla 1: Elegir "Slot" (DEV, DATA, RANDOM).
   - Pantalla 2 (Si es DEV/DATA): Mostrar checkboxes dinámicos del catálogo.
   - Pantalla 3: Resumen final ("Vas a instalar: Ollama, Go").

---

## 📌 FASE 3: El Motor de Aprovisionamiento (`bunker/instance.go`)
*Reemplazar el monstruoso script `build` en Bash por lógica en Go.*

1. **Refactor de la función `Build()`:**
   - Ya no ejecuta el script hardcodeado de `bunker_lifecycle.sh`.
   - Ahora lee el `state.json` generado por la TUI.
   - Itera sobre las herramientas seleccionadas y ejecuta la propiedad `script` de cada `Item` del catálogo.

2. **Limpieza de Bash:**
   - Eliminar por fin la lógica de instalación de `opencode`, `gentle-ai`, etc., del código Bash.

---

## 📌 FASE 4: Módulos Pendientes y Deuda Técnica (El sprint final)

1. **Migración del `config.toml` (Adiós `.env`):**
   - Resolver el bug **[ID-BUG-013]**.
   - Crear `pkg/config/` para manejar configuración en TOML. El `.env` es frágil para strings complejos.

2. **Destruir `lib/git.sh`:**
   - Las funciones de Git en Bash son propensas a fallos interactivos.
   - Crear `pkg/git/` y usar comandos nativos (`exec.Command` bien controlados o la librería `go-git`).
   - Integrar flujos de confirmación (commit, push, branch) en las tarjetas de la TUI.

3. **Seguridad y Procesos (Resolución de Bugs Críticos):**
   - Aplicar `filepath.Clean()` en **TODAS** las rutas **[ID-BUG-001]**.
   - Reemplazar todos los `exec.Command` por `exec.CommandContext` para evitar procesos zombis en el host **[ID-BUG-015]**.
   - Cambiar los permisos de `os.MkdirAll` de `0755` a `0700` **[ID-BUG-005]**.

---

## 💡 Notas para el Desarrollador (Tú)
* **Regla de Oro:** Si tocas lógica de negocio (pkg/bunker), **NO PUEDES** imprimir en pantalla (`fmt.Println`). Solo emitir eventos a la UI.
* **Iteración Corta:** No intentes hacer el motor de provisionamiento entero de golpe. Haz primero que el comando `axiom build` sepa leer el catálogo e imprimir en consola: *"Instalaría esto y esto..."* antes de ejecutar comandos reales en el contenedor.