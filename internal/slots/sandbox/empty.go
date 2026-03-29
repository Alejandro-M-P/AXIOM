package sandbox

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// Empty represents a SANDBOX slot - minimal image without additional installations.
type Empty struct{}

func (s *Empty) ID() string                    { return "empty" }
func (s *Empty) Name() string                  { return "Empty" }
func (s *Empty) Description() string           { return "Imagen mínima sin instalaciones adicionales" }
func (s *Empty) Category() domain.SlotCategory { return domain.SlotSANDBOX }
func (s *Empty) SubCategory() string           { return "" }
func (s *Empty) Dependencies() []string        { return []string{} }

func (s *Empty) Install(ctx context.Context, exec domain.Executor) error {
	// SANDBOX installs nothing - uses base image only
	return nil
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "empty",
		Name:        "Empty",
		Description: "Imagen mínima sin instalaciones adicionales",
		Category:    slots.SlotCategory(domain.SlotSANDBOX),
		SubCategory: "",
		Deps:        []string{},
	})
}
