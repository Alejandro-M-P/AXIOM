// Package tools contains developer tool installation items.
package tools

import (
	"github.com/Alejandro-M-P/AXIOM/internal/core/slots"
)

func init() {
	// Load items from embedded TOMLs and register them
	items, err := slots.LoadSlotsFromEmbeddedTOML("dev/tools/tomls")
	if err != nil {
		// Silently fail - this allows fallback to filesystem loading
		return
	}

	for i := range items {
		slots.RegisterItem(&items[i])
	}
}
