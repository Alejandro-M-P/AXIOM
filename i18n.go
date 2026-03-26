package i18n

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const (
	errReadTemplate   = "i18n: could not read translation file %s: %w"
	errDecodeTemplate = "i18n: could not decode TOML from %s: %w"
)

// Translator maneja la carga y recuperación de cadenas de texto traducidas desde archivos TOML.
type Translator struct {
	translations map[string]string
}

// NewTranslator crea e inicializa un nuevo Translator para el idioma especificado.
// Carga las traducciones desde un archivo .toml ubicado en el directorio `i18nPath`.
// Por ejemplo: NewTranslator("es", "configs/i18n") buscará "configs/i18n/es.toml".
func NewTranslator(lang, i18nPath string) (*Translator, error) {
	filePath := fmt.Sprintf("%s/%s.toml", i18nPath, lang)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(errReadTemplate, filePath, err)
	}

	var translations map[string]string
	if _, err := toml.Decode(string(data), &translations); err != nil {
		return nil, fmt.Errorf(errDecodeTemplate, filePath, err)
	}

	return &Translator{translations: translations}, nil
}

// T recupera una cadena traducida para una clave dada.
// Si la clave no se encuentra, devuelve la propia clave para indicar que falta una traducción.
// También puede formatear la cadena con los argumentos proporcionados usando fmt.Sprintf.
func (t *Translator) T(key string, args ...interface{}) string {
	format, ok := t.translations[key]
	if !ok {
		return key
	}

	// fmt.Sprintf es seguro de usar incluso si 'args' está vacío.
	return fmt.Sprintf(format, args...)
}