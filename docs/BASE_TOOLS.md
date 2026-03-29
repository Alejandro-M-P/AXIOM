# Sistema de Base Tools

## Descripción

El sistema de **Base Tools** desacopla la instalación de herramientas base del sistema operativo (brew, npm, pacman, etc.) del wizard de slots. Estas herramientas se instalan automáticamente según el entorno, sin intervención del usuario y sin aparecer en el wizard de selección.

## Arquitectura

### Componentes Principales

```
configs/os_preferences.toml      # Configuración de preferencias por SO
internal/slots/base/
├── os_detector.go               # Detecta el SO (Arch, Ubuntu, macOS)
├── installer.go                 # Instala herramientas base automáticamente
└── wrapper.go                   # Integración con el engine
internal/slots/engine.go         # Integración en el flujo de instalación
```

### Flujo de Trabajo

```
1. Detección de SO
   └── OSDetector.Detect() → Arch/Ubuntu/macOS

2. Preparación del Entorno
   └── InstallerEngine.PrepareEnvironment()
       └── BaseInstaller.InstallBaseTools()

3. Instalación de Slots
   └── InstallerEngine.executeInstall()
       └── SlotCommandAnalyzer.AnalyzeAndInstallRequirements()
           ├── Detectar herramientas requeridas del comando
           └── Instalar automáticamente si faltan
```

## Configuración (os_preferences.toml)

```toml
[arch]
name = "Arch Linux"
package_manager = "pacman"
base_tools = ["base-devel", "git", "curl"]
install_commands = [
    "sudo pacman -Syu --noconfirm",
    "sudo pacman -S --noconfirm base-devel git curl"
]

[ubuntu]
name = "Ubuntu/Debian"
package_manager = "apt"
base_tools = ["build-essential", "git", "curl"]
install_commands = [
    "sudo apt-get update",
    "sudo apt-get install -y build-essential git curl"
]

[macos]
name = "macOS"
package_manager = "brew"
base_tools = ["git", "curl"]
install_commands = []

# Mapeo de herramientas a comandos de instalación por SO
[tools.npm]
arch = "sudo pacman -S --noconfirm nodejs npm"
ubuntu = "sudo apt-get install -y nodejs npm"
macos = "brew install node"
```

## Uso

### 1. Crear Engine con Soporte de Base Tools

```go
// Crear engine con base tools automáticos
engine, err := slots.NewInstallerEngineWithBase(registry, "configs/os_preferences.toml")
if err != nil {
    // Fallback a engine sin base tools
    engine = slots.NewInstallerEngine(registry)
}

// Preparar entorno (instala herramientas base)
ctx := context.Background()
if err := engine.PrepareEnvironment(ctx); err != nil {
    log.Printf("Warning: %v", err)
}
```

### 2. Instalación Automática durante Slots

```go
// El engine detecta e instala herramientas base automáticamente
// antes de ejecutar cada comando del slot
items := []slots.SlotItem{
    {ID: "opencode", InstallCmd: "npm install -g opencode-ai"},
    {ID: "engram", InstallCmd: "brew tap gentle-ai/engram && brew install engram"},
}

// npm y brew se instalarán automáticamente si faltan
engine.Execute(items, progressCallback)
```

### 3. Filtrar en el Wizard

```go
// Los base tools NO aparecen en el wizard
// Usar FilterBaseTools o DiscoverUserSelectable
items := registry.DiscoverUserSelectable() // Excluye base tools
// o
items := slots.FilterBaseTools(allItems)
```

### 4. Detectar SO Manualmente

```go
osType, osName, err := base.DetectOSInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("OS: %s, Package Manager: %s\n", osName, base.GetPackageManager(osType))
```

## Características

### Detección Automática de Herramientas

El sistema analiza los comandos de instalación y detecta automáticamente qué herramientas base se necesitan:

```go
installer, _ := base.NewBaseInstaller(base.DefaultPreferencesPath())
required := installer.DetectRequiredTools("npm install -g opencode-ai")
// Result: ["npm"]
```

### Herramientas Soportadas

- **npm**: Node Package Manager
- **brew**: Homebrew (macOS/Linux)
- **pip/pip3**: Python Package Installer
- **pipx**: Python Application Installer
- **cargo**: Rust Package Manager
- **pacman**: Arch Package Manager (sistema)
- **apt**: Debian Package Manager (sistema)

### Filtrado en UI

Los base tools están marcados con `IsBaseTool: true` en el dominio y se filtran automáticamente del wizard:

```go
type SlotItem struct {
    ID         string
    Name       string
    IsBaseTool bool  // Si true, no aparece en el wizard
}
```

## Reglas de Oro

1. **Usuario NO elige**: Las herramientas base se instalan automáticamente, no aparecen en el wizard
2. **Sin hardcode**: Los comandos de instalación están en TOML, no en código
3. **Lógica en domain**: La lógica de base tools está en `internal/slots/base`, no en UI
4. **Transparente**: El usuario no necesita saber que npm/brew se instalan automáticamente

## Ejemplos

### Ejemplo 1: Slot que usa npm

```toml
# opencode.toml
id = "opencode"
category = "dev"
subcategory = "ia"
dependencies = []

[install]
cmd = "npm install -g opencode-ai"
```

Cuando se instala este slot:
1. El engine detecta que el comando usa `npm`
2. Verifica si npm está instalado
3. Si no, lo instala según el SO (pacman en Arch, apt en Ubuntu)
4. Luego ejecuta el comando de instalación del slot

### Ejemplo 2: Slot que usa brew

```toml
# engram.toml
id = "engram"
category = "dev"
subcategory = "ia"

[install]
steps = [
  "brew tap gentle-ai/engram",
  "brew install engram"
]
```

Cuando se instala este slot:
1. El engine detecta que los comandos usan `brew`
2. Verifica si brew está instalado (especial handling en macOS)
3. Si no, lo instala automáticamente
4. Luego ejecuta los pasos de instalación del slot

## Tests

```bash
# Ejecutar tests de base tools
go test ./tests/slots/base/... -v

# Ejecutar demo
go run ./examples/base_tools_demo/main.go
```

## Integración con el Engine

La integración es transparente. El `InstallerEngine` modificado:

1. Acepta un `BaseInstaller` opcional
2. Antes de cada comando, analiza y asegura las herramientas necesarias
3. Las herramientas del sistema (pacman, apt) se saltan (ya deben estar instaladas)

```go
func (e *InstallerEngine) executeInstall(ctx context.Context, item *SlotItem) error {
    // Asegurar herramientas base antes de ejecutar
    if e.analyzer != nil {
        if err := e.analyzer.AnalyzeAndInstallRequirements(ctx, item.InstallCmd); err != nil {
            return fmt.Errorf("failed to install base tools for %s: %w", item.ID, err)
        }
    }
    
    // Ejecutar comando normalmente
    cmd := exec.CommandContext(ctx, "sh", "-c", item.InstallCmd)
    return cmd.Run()
}
```

## Extensión

Para agregar soporte para un nuevo SO:

1. Agregar configuración en `configs/os_preferences.toml`
2. Agregar constante `OSType` en `os_detector.go`
3. Agregar lógica de detección en `detectLinux()`
4. Agregar comandos de instalación en `[tools.*]`

Para agregar una nueva herramienta base:

1. Agregar entrada en `[tools.nueva_herramienta]` en TOML
2. Agregar patrón de detección en `DetectRequiredTools()`
3. Agregar a la lista en `IsBaseTool()` si debe ocultarse del wizard
