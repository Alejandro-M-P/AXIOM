package ui

import (
	_ "embed"
	"fmt"

	"axiom/pkg/adapters/ui/styles"
	"github.com/pelletier/go-toml/v2"
)

//go:embed i18n/locales/es/errors.toml
var errorsTOML []byte

type ErrorDef struct {
	Title       string `toml:"title"`
	Description string `toml:"description"`
	Action      string `toml:"action"`
}

// UnmarshalTOML permite decodificar dinámicamente tanto strings simples como tablas completas (objetos) del TOML
func (e *ErrorDef) UnmarshalTOML(data any) error {
	switch v := data.(type) {
	case string:
		e.Title = "Error"
		e.Description = v
		e.Action = "Revisa los logs o ejecuta con modo verbose para más detalles."
	case map[string]any:
		if title, ok := v["title"].(string); ok {
			e.Title = title
		}
		if desc, ok := v["description"].(string); ok {
			e.Description = desc
		}
		if action, ok := v["action"].(string); ok {
			e.Action = action
		}
	default:
		return fmt.Errorf("formato no soportado para ErrorDef: %T", data)
		}
	return nil
	}
// UnmarshalText es la interfaz estándar que go-toml/v2 detecta mágicamente
// cuando se encuentra con un string plano en el TOML en lugar de una tabla.
func (e *ErrorDef) UnmarshalText(text []byte) error {
	e.Title = "Error"
	e.Description = string(text)
	e.Action = "Revisa los logs o ejecuta con modo verbose para más detalles."
	return nil
}

type ErrorCatalog map[string]map[string]*ErrorDef

var catalog ErrorCatalog

func init() {
	if err := toml.Unmarshal(errorsTOML, &catalog); err != nil {
		panic("Error interno del orquestador: no se pudo cargar errors.toml - " + err.Error())
	}
}

// RenderCommandError busca el error en el JSON según el comando y lo dibuja
func RenderCommandError(command string, err error) string {
	if err == nil {
		return ""
	}

	errCode := err.Error()
	
	// Seguridad: Si el catálogo no cargó, mostramos el error bruto
	if catalog == nil {
		return fmt.Sprintf("Error Crítico: %v", err)
	}

	cmdErrors, hasCmd := catalog[command]
	if !hasCmd {
		cmdErrors = catalog["global"]
	}

	// Si ni siquiera existe la sección global, evitamos el panic
	if cmdErrors == nil {
		return fmt.Sprintf("Error en %s: %v", command, err)
	}

	def, hasErr := cmdErrors[errCode]
	if !hasErr {
		// Buscamos el error 'unknown' de forma segura
		global, ok := catalog["global"]
		if ok {
			def = global["unknown"]
		}

		if def == nil {
			return styles.RenderErrorCard(command, "Error Desconocido", errCode, "Revisa los logs del sistema.")
		}
		
		techDetail := "[Detalle técnico: %v]"
		// Acceso seguro a Logs para evitar el panic de la línea 80
		if cliLogs, ok := Logs["cli"]; ok {
			if t, ok := cliLogs["technical_detail"]; ok {
				techDetail = t
			}
		}
		
		// IMPORTANTE: No modifiques el def original o corromperás el catálogo en memoria
		finalDesc := fmt.Sprintf("%s\n\n%s", def.Description, fmt.Sprintf(techDetail, err))
		return styles.RenderErrorCard(command, def.Title, finalDesc, def.Action)
	}

	return styles.RenderErrorCard(command, def.Title, def.Description, def.Action)
}