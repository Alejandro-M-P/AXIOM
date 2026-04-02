package ports

// Field representa un par clave-valor para mostrar al usuario.
// Interfaz para permitir diferentes implementaciones de presentación.
type Field interface {
	GetLabel() string
	GetValue() string
}

// NewField crea un Field desde dominio (sin conocer implementación UI)
func NewField(label, value string) Field {
	return &fieldImpl{Label: label, Value: value}
}

// fieldImpl es la implementación por defecto de Field
type fieldImpl struct {
	Label string
	Value string
}

func (f *fieldImpl) GetLabel() string { return f.Label }
func (f *fieldImpl) GetValue() string { return f.Value }
