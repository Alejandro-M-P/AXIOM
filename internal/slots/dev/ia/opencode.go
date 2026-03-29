package ia

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// Opencode representa el slot de instalación de opencode-ai via npm.
type Opencode struct{}

func (s *Opencode) ID() string                    { return "opencode" }
func (s *Opencode) Name() string                  { return "opencode-ai" }
func (s *Opencode) Description() string           { return "Asistente de código AI basado en LLMs locales" }
func (s *Opencode) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *Opencode) SubCategory() string           { return "ia" }
func (s *Opencode) Dependencies() []string        { return []string{} }

func (s *Opencode) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing opencode-ai...", "npm", "install", "-g", "opencode-ai")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "opencode",
		Name:        "opencode-ai",
		Description: "Asistente de código AI basado en LLMs locales",
		Category:    slots.SlotCategory(domain.SlotDEV),
		SubCategory: "ia",
		Deps:        []string{},
	})
}
