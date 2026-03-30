// Package sandbox contains sandbox environment installation items.
package sandbox

import (
	"github.com/Alejandro-M-P/AXIOM/internal/slots"
)

func init() {
	// Load items from embedded TOMLs and register them
	items, err := slots.LoadSlotsFromEmbeddedTOML("sandbox/tomls")
	if err != nil {
		// Silently fail - this allows fallback to filesystem loading
		return
	}

	for i := range items {
		slots.RegisterItem(&items[i])
	}
}
