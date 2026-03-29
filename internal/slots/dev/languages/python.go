// Package languages contains programming language items for the DEV slot.
package languages

import (
	"axiom/internal/domain"
	"axiom/internal/slots"
	"context"
)

// Python represents the Python programming language.
type Python struct{}

func (s *Python) ID() string                    { return "python" }
func (s *Python) Name() string                  { return "Python" }
func (s *Python) Description() string           { return "Lenguaje de programación Python" }
func (s *Python) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *Python) SubCategory() string           { return "languages" }
func (s *Python) Dependencies() []string        { return []string{} }

func (s *Python) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing Python...", "sudo", "pacman", "-S", "--noconfirm", "python")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "python",
		Name:        "Python",
		Description: "Lenguaje de programación Python",
		Category:    slots.SlotCategory(domain.SlotDEV),
		SubCategory: "languages",
		Deps:        []string{},
	})
}
