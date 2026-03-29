// Package ia contains AI/LLM related installation items.
package ia

import (
	"axiom/internal/slots"
)

func init() {
	// Load items from embedded TOMLs and register them
	items, err := slots.LoadSlotsFromEmbeddedTOML("dev/ia/tomls")
	if err != nil {
		// Silently fail - this allows fallback to filesystem loading
		return
	}

	for i := range items {
		slots.RegisterItem(&items[i])
	}
}
