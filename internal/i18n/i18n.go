package i18n

import (
	_ "embed"
	"fmt"
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
	// Default to Spanish — SetLocale must be called explicitly from main.go
	currentLocale = "es"
	loadLocale(currentLocale)
}

// SetLocale sets the locale and reloads all translation files.
// This should be called from main.go or router after detecting the system locale.
func SetLocale(locale string) {
	currentLocale = normalizeLocale(locale)
	loadLocale(currentLocale)
}

// normalizeLocale converts a LANG value like "es_ES.UTF-8" or "en_US.UTF-8" to "es" or "en".
func normalizeLocale(lang string) string {
	if lang == "" {
		return "es"
	}
	lang = strings.ToLower(lang)
	if strings.HasPrefix(lang, "en") {
		return "en"
	}
	return "es"
}

// detectLocale is kept for backward compatibility but delegates to normalizeLocale.
func detectLocale(lang string) string {
	return normalizeLocale(lang)
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

// GetLifecycleText returns a localized string from lifecycle.toml with optional format args.
// section is the TOML section (e.g. "build.steps"), key is the field (e.g. "ollama_url").
// Returns the key itself as fallback if not found.
func GetLifecycleText(section, key string, args ...interface{}) string {
	if Lifecycle == nil {
		if len(args) > 0 {
			return fmt.Sprintf(section+"."+key, args...)
		}
		return section + "." + key
	}
	if sectionData, ok := Lifecycle[section]; ok {
		if val, ok := sectionData[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(val, args...)
			}
			return val
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(section+"."+key, args...)
	}
	return section + "." + key // fallback: show the key
}

// GetErrorsText returns a localized string from errors.toml with optional format args.
func GetErrorsText(section, key string, args ...interface{}) string {
	if Errors == nil {
		if len(args) > 0 {
			return fmt.Sprintf(section+"."+key, args...)
		}
		return section + "." + key
	}
	if sectionData, ok := Errors[section]; ok {
		if val, ok := sectionData[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(val, args...)
			}
			return val
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(section+"."+key, args...)
	}
	return section + "." + key // fallback: show the key
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
		panic(fmt.Sprintf("Error in commands.toml: %s", err.Error()))
	}
	if err := toml.Unmarshal(promptsData, &Prompts); err != nil {
		panic(fmt.Sprintf("Error in prompts.toml: %s", err.Error()))
	}
	if len(logsData) > 0 {
		if err := toml.Unmarshal(logsData, &Logs); err != nil {
			panic(fmt.Sprintf("Error in logs.toml: %s", err.Error()))
		}
	}
	if len(lifecycleData) > 0 {
		if err := toml.Unmarshal(lifecycleData, &Lifecycle); err != nil {
			panic(fmt.Sprintf("Error in lifecycle.toml: %s", err.Error()))
		}
	}
	if len(errorsData) > 0 {
		if err := toml.Unmarshal(errorsData, &Errors); err != nil {
			panic(fmt.Sprintf("Error in errors.toml: %s", err.Error()))
		}
	}
	if len(initData) > 0 {
		if err := toml.Unmarshal(initData, &Init); err != nil {
			panic(fmt.Sprintf("Error in init.toml: %s", err.Error()))
		}
	}
	if len(slotsData) > 0 {
		if err := toml.Unmarshal(slotsData, &Slots); err != nil {
			panic(fmt.Sprintf("Error in slots.toml: %s", err.Error()))
		}
	}
	if len(wizardData) > 0 {
		if err := toml.Unmarshal(wizardData, &Wizard); err != nil {
			panic(fmt.Sprintf("Error in wizard.toml: %s", err.Error()))
		}
	}
}
