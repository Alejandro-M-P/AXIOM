package tools

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// Starship is the minimalist prompt for any shell.
type Starship struct{}

func (s *Starship) ID() string                    { return "starship" }
func (s *Starship) Name() string                  { return "Starship" }
func (s *Starship) Description() string           { return "Prompt minimalista para cualquier shell" }
func (s *Starship) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *Starship) SubCategory() string           { return "tools" }
func (s *Starship) Dependencies() []string        { return []string{} }

func (s *Starship) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing Starship...", "/bin/bash", "-c", "curl -sS https://starship.rs/install.sh | sh")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "starship",
		Name:        "Starship",
		Description: "Prompt minimalista para cualquier shell",
		Category:    slots.SlotDEV,
		SubCategory: "tools",
		Deps:        []string{},
	})
}
