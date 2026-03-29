// Package languages contains programming language items for the DEV slot.
package languages

import (
	"axiom/internal/domain"
	"axiom/internal/slots"
	"context"
)

// Go represents the Go programming language.
type Go struct{}

func (s *Go) ID() string                    { return "go" }
func (s *Go) Name() string                  { return "Go" }
func (s *Go) Description() string           { return "Lenguaje de programación Go" }
func (s *Go) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *Go) SubCategory() string           { return "languages" }
func (s *Go) Dependencies() []string        { return []string{} }

func (s *Go) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing Go...", "sudo", "pacman", "-S", "--noconfirm", "go")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "go",
		Name:        "Go",
		Description: "Lenguaje de programación Go",
		Category:    slots.SlotDEV,
		SubCategory: "languages",
		Deps:        []string{},
	})
}
