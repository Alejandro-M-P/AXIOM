package i18n

import (
	_ "embed"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Spanish locale files
//
//go:embed locales/es/commands.toml
var esCommandsTOML []byte

//go:embed locales/es/prompts.toml
var esPromptsTOML []byte

//go:embed locales/es/logs.toml
var esLogsTOML []byte

//go:embed locales/es/lifecycle.toml
var esLifecycleTOML []byte

//go:embed locales/es/errors.toml
var esErrorsTOML []byte

//go:embed locales/es/init.toml
var esInitTOML []byte

//go:embed locales/es/slots.toml
var esSlotsTOML []byte

//go:embed locales/es/wizard.toml
var esWizardTOML []byte

// English locale files
//
//go:embed locales/en/commands.toml
var enCommandsTOML []byte

//go:embed locales/en/prompts.toml
var enPromptsTOML []byte

//go:embed locales/en/logs.toml
var enLogsTOML []byte

//go:embed locales/en/lifecycle.toml
var enLifecycleTOML []byte

//go:embed locales/en/errors.toml
var enErrorsTOML []byte

//go:embed locales/en/init.toml
var enInitTOML []byte

//go:embed locales/en/slots.toml
var enSlotsTOML []byte

//go:embed locales/en/wizard.toml
var enWizardTOML []byte

// Data maps for the loaded locale
var Commands map[string]map[string]string
var Prompts map[string]map[string]string
var Logs map[string]map[string]string
var Lifecycle map[string]map[string]string
var Errors map[string]map[string]string
var Init map[string]map[string]string
var Slots map[string]map[string]string
var Wizard map[string]map[string]string

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

// GetWizardText returns a localized string from wizard.toml.
// section is the TOML section (e.g. "step_language"), key is the field (e.g. "title").
// Returns the key itself as fallback if not found.
func GetWizardText(section, key string) string {
	if Wizard == nil {
		return section + "." + key
	}
	if sectionData, ok := Wizard[section]; ok {
		if val, ok := sectionData[key]; ok {
			return val
		}
	}
	return section + "." + key // fallback: show the key
}

// GetEscapeButtonText returns the localized text for the escape button
func GetEscapeButtonText() string {
	if currentLocale == "en" {
		if text, ok := Prompts["escape_button"]["text"]; ok {
			return text
		}
		return "Exit" // fallback
	}
	// Default to Spanish
	if text, ok := Prompts["escape_button"]["text"]; ok {
		return text
	}
	return "Salir" // fallback
}

// loadLocale loads the TOML files for the specified locale
func loadLocale(locale string) {
	var commandsData, promptsData, logsData, lifecycleData, errorsData, initData, slotsData, wizardData []byte

	if locale == "en" {
		commandsData = enCommandsTOML
		promptsData = enPromptsTOML
		logsData = enLogsTOML
		lifecycleData = enLifecycleTOML
		errorsData = enErrorsTOML
		initData = enInitTOML
		slotsData = enSlotsTOML
		wizardData = enWizardTOML
	} else {
		// Default to Spanish
		commandsData = esCommandsTOML
		promptsData = esPromptsTOML
		logsData = esLogsTOML
		lifecycleData = esLifecycleTOML
		errorsData = esErrorsTOML
		initData = esInitTOML
		slotsData = esSlotsTOML
		wizardData = esWizardTOML
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
	if len(slotsData) > 0 {
		if err := toml.Unmarshal(slotsData, &Slots); err != nil {
			panic("Error en slots.toml: " + err.Error())
		}
	}
	if len(wizardData) > 0 {
		if err := toml.Unmarshal(wizardData, &Wizard); err != nil {
			panic("Error en wizard.toml: " + err.Error())
		}
	}
}
