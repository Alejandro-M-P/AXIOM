// Package data contains database and data storage installation items.
package data

import (
	"github.com/Alejandro-M-P/AXIOM/internal/core/slots"
)

func init() {
	// Load items from embedded TOMLs and register them
	items, err := slots.LoadSlotsFromEmbeddedTOML("data/tomls")
	if err != nil {
		// Silently fail - this allows fallback to filesystem loading
		return
	}

	for i := range items {
		slots.RegisterItem(&items[i])
	}
}
