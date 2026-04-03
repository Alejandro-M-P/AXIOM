// Package domain contiene los modelos puros del negocio.
// No tiene dependencias externas - es la capa más interna de la Clean Architecture.
package domain

import "context"

// SlotCategory representa la categoría de un slot de instalación.
type SlotCategory string

// Slot define el contrato para cualquier item de slot de instalación.
type Slot interface {
	// ID retorna el identificador único del slot.
	ID() string
	// Name retorna el nombre descriptivo.
	Name() string
	// Description retorna la descripción del slot.
	Description() string
	// Category retorna la categoría del slot.
	Category() SlotCategory
	// SubCategory retorna la subcategoría.
	SubCategory() string
	// Dependencies retorna las dependencias del slot.
	Dependencies() []string
	// Install ejecuta la instalación del slot.
	Install(ctx context.Context) error
}
