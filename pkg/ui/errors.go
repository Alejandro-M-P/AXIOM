package ui

import (
	_ "embed"
	"fmt"

	"axiom/pkg/ui/styles"
	"github.com/pelletier/go-toml/v2"
)

//go:embed locales/es/errors.toml
var errorsTOML []byte

type ErrorDef struct {
	Title       string `toml:"title"`
	Description string `toml:"description"`
	Action      string `toml:"action"`
}

type ErrorCatalog map[string]map[string]ErrorDef

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
	cmdErrors, hasCmd := catalog[command]
	if !hasCmd {
		cmdErrors = catalog["global"]
	}

	def, hasErr := cmdErrors[errCode]
	if !hasErr {
		def = catalog["global"]["unknown"]
		techDetail := "[Detalle técnico: %v]"
		if t, ok := Logs["cli"]["technical_detail"]; ok {
			techDetail = t
		}
		def.Description = fmt.Sprintf("%s\n\n%s", def.Description, fmt.Sprintf(techDetail, err))
	}

	return styles.RenderErrorCard(command, def.Title, def.Description, def.Action)
}