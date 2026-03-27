package ui

import (
    _ "embed"
    "github.com/pelletier/go-toml/v2"
)

//go:embed i18n/locales/es/commands.toml
var commandsTOML []byte

//go:embed i18n/locales/es/prompts.toml
var promptsTOML []byte

//go:embed i18n/locales/es/logs.toml
var logsTOML []byte

//go:embed i18n/locales/es/lifecycle.toml
var lifecycleTOML []byte

// 2. TIPOS DE DATOS: Volvemos a los mapas para que presenter.go pueda leerlos
var Commands map[string]map[string]string
var Prompts map[string]map[string]string
var Logs map[string]map[string]string
var Lifecycle map[string]map[string]string 

func init() {
    if err := toml.Unmarshal(commandsTOML, &Commands); err != nil {
        panic("Error en commands.toml: " + err.Error())
    }
    if err := toml.Unmarshal(promptsTOML, &Prompts); err != nil {
        panic("Error en prompts.toml: " + err.Error())
    }
    if len(logsTOML) > 0 {
        if err := toml.Unmarshal(logsTOML, &Logs); err != nil {
            panic("Error en logs.toml: " + err.Error())
        }
    }
    if len(lifecycleTOML) > 0 {
        if err := toml.Unmarshal(lifecycleTOML, &Lifecycle); err != nil {
            panic("Error en lifecycle.toml: " + err.Error())
        }
    }
}