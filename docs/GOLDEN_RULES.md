



# ⚡ Reglas de Oro de AXIOM

Estas reglas son **innegociables**. Antes de hacer un PR o mergear cualquier
cosa, verificá que las cumplís todas. Si rompés una regla, el compilador no
te va a avisar — pero el proyecto se va a pudrir lentamente.

---

## El mapa mental antes de escribir una línea

```
¿Qué tipo de código estoy escribiendo?

  ¿Toma decisiones de negocio?        → va en internal/bunker/, build/, slots/
  ¿Define un contrato entre capas?    → va en internal/ports/
  ¿Habla con el sistema operativo?    → va en internal/adapters/system/
  ¿Habla con Podman o Distrobox?      → va en internal/adapters/runtime/
  ¿Muestra algo al usuario?           → va en internal/adapters/ui/
  ¿Es una entidad del dominio?        → va en internal/domain/
  ¿Conecta comandos con handlers?     → va en internal/router/
```

Si no sabés dónde va, preguntate: **¿este código cambiaría si mañana
reemplazamos Podman por Docker?** Si sí → es infraestructura, va en
`adapters/`. Si no → es lógica de negocio, va en el core.

---

## Regla 1 — El core es ciego al sistema

Los paquetes `internal/bunker/`, `internal/build/` e `internal/slots/`
**no saben que existe Podman, Distrobox, sudo, ssh-add, curl, pacman,
ni ningún comando del sistema operativo**.

Si el core necesita ejecutar algo en el sistema, usa un port.

```go
// ✅ CORRECTO — el core habla con una abstracción
m.runtime.EnterBunker(ctx, name)
m.runtime.CommitImage(ctx, containerName, imageName)
m.system.RefreshSudo(ctx)
m.system.PrepareSSHAgent(ctx, keyPath)
m.runner.RunShell(ctx, "curl -fsSL https://...")

// ❌ INCORRECTO — el core sabe que existe Podman
exec.Command("podman", "commit", ...)
exec.Command("distrobox-enter", name, ...)
exec.Command("sudo", "-v")
exec.CommandContext(ctx, "ssh-add", "-l")
exec.CommandContext(ctx, "sh", "-c", cmd)
```

**¿Dónde sí puede vivir exec.Command?**
Solo en `internal/adapters/runtime/`, `internal/adapters/system/`,
y `internal/slots/base/` (infraestructura de slots).
En ningún otro lugar.

---

## Regla 2 — El core es mudo

Los paquetes `internal/bunker/`, `internal/build/` e `internal/slots/`
**nunca escriben en pantalla directamente**. Ni logs, ni errores, ni mensajes.
Ni siquiera warnings.

La razón es simple: el core no sabe si el usuario está en una terminal,
en un agente de IA, en una API JSON, o en una interfaz web. Eso lo decide
el adapter de UI.

```go
// ✅ CORRECTO — el core delega la presentación
m.ui.ShowLog("logs.creating_bunker", name)
m.ui.ShowError(err)
m.ui.ShowWarning("warnings.bunker_exists.title", ...)

// ❌ INCORRECTO — el core habla directamente
fmt.Println("Creando bunker...")
fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
log.Printf("error: %v", err)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
```

**¿Dónde sí puede vivir fmt.Print y os.Stdout?**
Solo en `internal/adapters/ui/` y en `cmd/axiom/main.go`.
En ningún otro lugar.

---

## Regla 3 — El core es sordo al entorno

El core **no lee variables de entorno directamente**. Las variables del
sistema entran por `IFileSystem.UserHomeDir()` o por `ISystem`.

La razón: `os.Getenv` hace el core dependiente del entorno de ejecución,
lo que rompe los tests y hace el comportamiento impredecible.

```go
// ✅ CORRECTO — el core pregunta al adapter
homeDir, err := m.fs.UserHomeDir()
axiomPath := m.system.GetEnv("AXIOM_PATH")

// ❌ INCORRECTO — el core habla con el sistema directamente
os.Getenv("HOME")
os.Getenv("USER")
os.Getenv("AXIOM_PATH")
os.Getenv("LANG")
```

**¿Dónde sí puede vivir os.Getenv?**
Solo en `internal/adapters/`, `internal/router/` (para config inicial)
y en `cmd/axiom/main.go`.

---

## Regla 4 — Siempre hay contexto

**Nunca usar `exec.Command()`**. Siempre `exec.CommandContext(ctx, ...)`.
Sin contexto, un proceso puede colgar para siempre y AXIOM se congela.

```go
// ✅ CORRECTO — el proceso puede cancelarse
cmd := exec.CommandContext(ctx, "podman", "ps")

// ❌ INCORRECTO — si el proceso se cuelga, AXIOM se cuelga
cmd := exec.Command("podman", "ps")
```

El `ctx` siempre viene del router — cada comando tiene un contexto
con timeout. Los adapters reciben ese contexto y lo pasan hacia abajo.

---

## Regla 5 — Los strings del sistema viven en un solo lugar

Los strings `"podman"`, `"distrobox"`, `"distrobox-create"`,
`"distrobox-enter"` solo existen en **un archivo**:
`internal/adapters/runtime/commands.go`.

Si ves uno de esos strings en cualquier otro archivo es una violación.
Eso es lo que hace posible cambiar de Podman a Docker tocando
un solo lugar.

```go
// ✅ CORRECTO — el string vive en commands.go
var Podman = CommandSet{
    CreateBunker: func(name, image, home, flags string) []string {
        return []string{"distrobox-create", "--name", name, ...}
    },
}

// ❌ INCORRECTO — el string está disperso
exec.Command("distrobox-create", "--name", name, ...)  // en bunker/create.go
exec.Command("distrobox-enter", name, ...)              // en bunker/create.go
```

---

## Regla 6 — El core devuelve dominio, no strings

El core devuelve structs de dominio. Nunca strings formateados,
nunca tablas de texto, nunca ANSI. Formatear para el usuario
es responsabilidad exclusiva de `adapters/ui/`.

```go
// ✅ CORRECTO — el core devuelve datos
func (m *Manager) ListBunkers(ctx context.Context) ([]domain.Bunker, error)

// ❌ INCORRECTO — el core formatea para el usuario
func (m *Manager) ListBunkers(ctx context.Context) (string, error)
// devolviendo "bunker1 (running)\nbunker2 (stopped)"
```

---

## Regla 7 — 1 slot = 1 TOML

Añadir un nuevo slot es crear **un archivo `.toml`**.
No tocar Go. No registrar nada manualmente. No recompilar.

```toml
# internal/slots/dev/tools/tomls/neovim.toml
id = "neovim"
name = "Neovim"
description = "Editor de texto modal"
category = "dev"
subcategory = "tools"
dependencies = []

[install]
cmd = "sudo pacman -S --noconfirm neovim"
```

Si para añadir un slot hay que tocar un archivo `.go` →
la arquitectura está rota.

---

## Regla 8 — Los textos visibles no existen en Go

Ningún string que el usuario pueda ver vive en el código Go.
Ni mensajes de error, ni labels, ni confirmaciones, ni warnings.
Todo pasa por el sistema de i18n vía claves.

```go
// ✅ CORRECTO — clave i18n, el texto vive en el TOML
m.ui.ShowLog("logs.bunker_created", name)
m.ui.ShowWarning("warnings.image_missing.title", ...)
return fmt.Errorf("missing_image")

// ❌ INCORRECTO — string visible hardcodeado en Go
m.ui.ShowLog("Bunker creado: " + name)
return fmt.Errorf("La imagen %s no existe", image)
```

Los textos viven en:
`internal/adapters/ui/views/i18n/locales/es/` y `/en/`

---

## Regla 9 — El core no ejecuta, no lee, no escribe al sistema

El core (`bunker/`, `slots/`, `build/`) **nunca** llama directamente a:
- `exec.Command`, `exec.CommandContext`, `exec.LookPath`
- `os.Getenv`, `os.Stat`, `os.ReadFile`, `os.WriteFile`
- `fmt.Print`, `fmt.Fprintf(os.Stderr, ...)`, `log.Printf`

Si necesita ejecutar un comando → usa `ICommandRunner.RunShell(ctx, cmd)`.
Si necesita buscar un binario → usa `ISystem.GetCommandPath(name)`.
Si necesita una variable de entorno → usa `ISystem.GetEnv(key)`.
Si necesita mostrar algo → usa `IPresenter.ShowLog(key, args...)` o `GetText(key, args...)`.
Si necesita reportar un error → devuelve `fmt.Errorf("clave_i18n: %w", err)`.

```go
// ✅ CORRECTO — el core usa abstracciones
output, err := m.runner.RunShell(ctx, "curl -fsSL https://...")
path, err := m.system.GetCommandPath("pacman")
m.ui.ShowLog("logs.installing", packageName)
return fmt.Errorf("errors.slots.command_failed: %w", err)

// ❌ INCORRECTO — el core habla con el sistema directamente
exec.CommandContext(ctx, "sh", "-c", cmd)
exec.LookPath("pacman")
os.Getenv("AXIOM_PATH")
fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
```

**¿Dónde sí puede vivir exec.Command?**
Solo en `internal/adapters/` y `internal/slots/base/` (infraestructura).
En ningún otro lugar.

---

## Regla 10 — Cero texto hardcodeado en Go

**Ningún string que el usuario pueda ver existe en el código Go.**
Ni mensajes de error, ni labels, ni logs, ni confirmaciones, ni warnings,
ni subtítulos, ni nombres de categorías. **Todo** va a los TOMLs de i18n.

La razón: si un string está en Go, está duplicado, no se traduce, y no
se puede cambiar sin recompilar. Si está en un TOML, se traduce, se
modifica sin tocar código, y vive en un solo lugar.

```go
// ✅ CORRECTO — la clave i18n, el texto vive en el TOML
m.ui.ShowLog("logs.slots.installing", item.Name)
m.ui.GetText("slots.subcategories.tools")
return fmt.Errorf("errors.slots.command_failed: %w", err)

// ❌ INCORRECTO — string hardcodeado en Go
m.ui.ShowLog("info", "Installing:", item.Name)
return "Developer Tools"
return fmt.Errorf("command failed: %w", err)
```

**¿Qué NO es texto hardcodeado?**
- Claves i18n como `"logs.slots.installing"` → son referencias, no texto visible
- Placeholders como `%s`, `%w`, `%d` → son formatos, no texto visible
- IDs técnicos como `"ia"`, `"tools"`, `"data"` → son claves de negocio, no texto visible

**¿Qué SÍ es texto hardcodeado?**
- `"Installing:"`, `"Running:"`, `"Step %d/%d"` → texto visible
- `"AI / LLM Models"`, `"Developer Tools"` → texto visible
- `"command failed"`, `"slot selector failed"` → texto visible
- `"Warning: %v\n"` → texto visible

Los textos viven exclusivamente en:
`internal/i18n/locales/es/` y `internal/i18n/locales/en/`

---

## La tabla de qué puede hacer cada capa

| | `exec` | `os.Stdout` | `fmt.Print` | `os.Getenv` | Bubbletea | Texto visible |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| `domain/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `ports/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `bunker/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `build/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `slots/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `router/` | ❌ | ❌ | ⚠️ solo errores fatales | ⚠️ solo config inicial | ❌ | ❌ |
| `adapters/runtime/` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `adapters/system/` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| `adapters/filesystem/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `adapters/ui/` | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| `slots/base/` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `cmd/axiom/` | ❌ | ⚠️ solo bootstrap | ⚠️ solo bootstrap | ✅ | ❌ | ❌ |

---

## Verificación rápida — ejecutar antes de cada PR

Si alguno de estos comandos devuelve algo, hay una violación:

```bash
# exec fuera de donde debe estar
grep -rn "exec\.Command" internal/ --include="*.go" \
  | grep -v "/adapters/" | grep -v "/slots/base/" | grep -v "_test.go"

# stdout/stderr en el core
grep -rn "os\.Stdout\|os\.Stderr\|os\.Stdin" internal/ --include="*.go" \
  | grep -v "/adapters/" | grep -v "_test.go"

# prints en el core
grep -rn "fmt\.Print\|fmt\.Fprintf\|fmt\.Fprintln" internal/ --include="*.go" \
  | grep -v "/adapters/" | grep -v "_test.go"

# variables de entorno en el core
grep -rn "os\.Getenv" internal/ --include="*.go" \
  | grep -v "/adapters/" | grep -v "/router/" | grep -v "_test.go"

# strings de sistema fuera de commands.go
grep -rn '"podman"\|"distrobox"\|"distrobox-create"\|"distrobox-enter"' \
  internal/ --include="*.go" | grep -v "commands\.go" | grep -v "_test.go"
```

Metelos en el Makefile como `make lint-arch` para no tener que
acordarte de ejecutarlos.

```makefile
lint-arch:
	@echo "🏛️  Verificando arquitectura..."
	@! grep -rn "exec\.Command" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "/slots/base/" | grep -v "_test.go" \
	  | grep . && echo "✅ exec.Command OK" || (echo "❌ exec.Command en el core" && exit 1)
	@! grep -rn "os\.Stdout\|os\.Stderr\|os\.Stdin" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "_test.go" \
	  | grep . && echo "✅ os.Std* OK" || (echo "❌ os.Std* en el core" && exit 1)
	@! grep -rn "fmt\.Print\|fmt\.Fprintf\|fmt\.Fprintln" internal/ --include="*.go" \
	  | grep -v "/adapters/" | grep -v "_test.go" \
	  | grep . && echo "✅ fmt.Print OK" || (echo "❌ fmt.Print en el core" && exit 1)
	@! grep -rn '"podman"\|"distrobox"\|"distrobox-create"\|"distrobox-enter"' \
	  internal/ --include="*.go" | grep -v "commands\.go" | grep -v "_test.go" \
	  | grep . && echo "✅ Strings de sistema OK" || (echo "❌ Strings de sistema fuera de commands.go" && exit 1)
	@echo "✅ Arquitectura limpia"
```
