package ui

import (
	_ "embed"
	"github.com/pelletier/go-toml/v2"
)

//go:embed locales/es/commands.toml
var commandsTOML []byte

//go:embed locales/es/prompts.toml
var promptsTOML []byte

//go:embed locales/es/logs.toml
var logsTOML []byte

//go:embed locales/es/lifecycle.toml
var lifecycleTOML []byte

var Commands map[string]map[string]string
var Prompts map[string]map[string]string
var Logs map[string]map[string]string
var Lifecycle map[string]map[string]string

func init() {
	if err := toml.Unmarshal(commandsTOML, &Commands); err != nil {
		panic("Error interno: no se pudo cargar commands.toml - " + err.Error())
	}
	if err := toml.Unmarshal(promptsTOML, &Prompts); err != nil {
		panic("Error interno: no se pudo cargar prompts.toml - " + err.Error())
	}
	if len(logsTOML) > 0 {
		if err := toml.Unmarshal(logsTOML, &Logs); err != nil {
			panic("Error interno: no se pudo cargar logs.toml - " + err.Error())
		}
	}
	if len(lifecycleTOML) > 0 {
		if err := toml.Unmarshal(lifecycleTOML, &Lifecycle); err != nil {
			panic("Error interno: no se pudo cargar lifecycle.toml - " + err.Error())
		}
	}
}