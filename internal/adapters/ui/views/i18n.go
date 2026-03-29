package ui

import (
	_ "embed"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Spanish locale files
//
//go:embed i18n/locales/es/commands.toml
var esCommandsTOML []byte

//go:embed i18n/locales/es/prompts.toml
var esPromptsTOML []byte

//go:embed i18n/locales/es/logs.toml
var esLogsTOML []byte

//go:embed i18n/locales/es/lifecycle.toml
var esLifecycleTOML []byte

//go:embed i18n/locales/es/errors.toml
var esErrorsTOML []byte

//go:embed i18n/locales/es/init.toml
var esInitTOML []byte

// English locale files
//
//go:embed i18n/locales/en/commands.toml
var enCommandsTOML []byte

//go:embed i18n/locales/en/prompts.toml
var enPromptsTOML []byte

//go:embed i18n/locales/en/logs.toml
var enLogsTOML []byte

//go:embed i18n/locales/en/lifecycle.toml
var enLifecycleTOML []byte

//go:embed i18n/locales/en/errors.toml
var enErrorsTOML []byte

//go:embed i18n/locales/en/init.toml
var enInitTOML []byte

// Data maps for the loaded locale
var Commands map[string]map[string]string
var Prompts map[string]map[string]string
var Logs map[string]map[string]string
var Lifecycle map[string]map[string]string
var Errors map[string]map[string]string
var Init map[string]map[string]string

// currentLocale stores the detected locale
var currentLocale string

func init() {
	currentLocale = detectLocale()
	loadLocale(currentLocale)
}

// detectLocale detects system locale from LANG env var, defaults to "es"
func detectLocale() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		return "es"
	}
	// LANG format: es_ES.UTF-8, en_US.UTF-8, etc.
	lang = strings.ToLower(lang)
	if strings.HasPrefix(lang, "en") {
		return "en"
	}
	// Default to Spanish for any other locale
	return "es"
}

// GetLocale returns the currently loaded locale
func GetLocale() string {
	return currentLocale
}

// loadLocale loads the TOML files for the specified locale
func loadLocale(locale string) {
	var commandsData, promptsData, logsData, lifecycleData, errorsData, initData []byte

	if locale == "en" {
		commandsData = enCommandsTOML
		promptsData = enPromptsTOML
		logsData = enLogsTOML
		lifecycleData = enLifecycleTOML
		errorsData = enErrorsTOML
		initData = enInitTOML
	} else {
		// Default to Spanish
		commandsData = esCommandsTOML
		promptsData = esPromptsTOML
		logsData = esLogsTOML
		lifecycleData = esLifecycleTOML
		errorsData = esErrorsTOML
		initData = esInitTOML
	}

	if err := toml.Unmarshal(commandsData, &Commands); err != nil {
		panic("Error en commands.toml: " + err.Error())
	}
	if err := toml.Unmarshal(promptsData, &Prompts); err != nil {
		panic("Error en prompts.toml: " + err.Error())
	}
	if len(logsData) > 0 {
		if err := toml.Unmarshal(logsData, &Logs); err != nil {
			panic("Error en logs.toml: " + err.Error())
		}
	}
	if len(lifecycleData) > 0 {
		if err := toml.Unmarshal(lifecycleData, &Lifecycle); err != nil {
			panic("Error en lifecycle.toml: " + err.Error())
		}
	}
	if len(errorsData) > 0 {
		if err := toml.Unmarshal(errorsData, &Errors); err != nil {
			panic("Error en errors.toml: " + err.Error())
		}
	}
	if len(initData) > 0 {
		if err := toml.Unmarshal(initData, &Init); err != nil {
			panic("Error en init.toml: " + err.Error())
		}
	}
}
