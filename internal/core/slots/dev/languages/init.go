// Package languages contains programming language installation items.
package languages

import (
	"github.com/Alejandro-M-P/AXIOM/internal/core/slots"
)

func init() {
	// Load items from embedded TOMLs and register them
	items, err := slots.LoadSlotsFromEmbeddedTOML("dev/languages/tomls")
	if err != nil {
		// Silently fail - this allows fallback to filesystem loading
		return
	}

	for i := range items {
		slots.RegisterItem(&items[i])
	}
}
