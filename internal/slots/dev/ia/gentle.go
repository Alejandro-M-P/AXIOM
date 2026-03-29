package ia

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// Gentle representa el slot de instalación de gentle-ai via GitHub releases.
type Gentle struct{}

func (s *Gentle) ID() string                    { return "gentle" }
func (s *Gentle) Name() string                  { return "gentle-ai" }
func (s *Gentle) Description() string           { return "Framework de agentes AI con memoria persistente" }
func (s *Gentle) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *Gentle) SubCategory() string           { return "ia" }
func (s *Gentle) Dependencies() []string        { return []string{} }

func (s *Gentle) Install(ctx context.Context, exec domain.Executor) error {
	if err := exec(ctx, "Downloading gentle-ai...", "curl", "-fsSL", "https://github.com/gentle-ai/gentle/releases/latest/download/gentle-linux-amd64", "-o", "/tmp/gentle"); err != nil {
		return err
	}
	return exec(ctx, "Installing gentle-ai...", "/bin/bash", "-c", "mkdir -p \"$HOME/.local/bin\" && mv /tmp/gentle \"$HOME/.local/bin/gentle\" && chmod +x \"$HOME/.local/bin/gentle\"")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "gentle",
		Name:        "gentle-ai",
		Description: "Framework de agentes AI con memoria persistente",
		Category:    slots.SlotCategory(domain.SlotDEV),
		SubCategory: "ia",
		Deps:        []string{},
	})
}
