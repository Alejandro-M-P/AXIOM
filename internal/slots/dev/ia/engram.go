package ia

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// Engram representa el slot de instalación de Engram via brew.
type Engram struct{}

func (s *Engram) ID() string                    { return "engram" }
func (s *Engram) Name() string                  { return "Engram" }
func (s *Engram) Description() string           { return "Sistema de memoria persistente para agentes de IA" }
func (s *Engram) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *Engram) SubCategory() string           { return "ia" }
func (s *Engram) Dependencies() []string        { return []string{} }

func (s *Engram) Install(ctx context.Context, exec domain.Executor) error {
	if err := exec(ctx, "Tapping Engram repository...", "brew", "tap", "gentle-ai/engram"); err != nil {
		return err
	}
	return exec(ctx, "Installing Engram...", "brew", "install", "engram")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "engram",
		Name:        "Engram",
		Description: "Sistema de memoria persistente para agentes de IA",
		Category:    slots.SlotDEV,
		SubCategory: "ia",
		Deps:        []string{},
	})
}
