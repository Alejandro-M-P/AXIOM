// Package languages contains programming language items for the DEV slot.
package languages

import (
	"axiom/internal/domain"
	"axiom/internal/slots"
	"context"
)

// NodeJS represents Node.js with npm.
type NodeJS struct{}

func (s *NodeJS) ID() string                    { return "nodejs" }
func (s *NodeJS) Name() string                  { return "Node.js" }
func (s *NodeJS) Description() string           { return "Runtime de JavaScript con npm" }
func (s *NodeJS) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *NodeJS) SubCategory() string           { return "languages" }
func (s *NodeJS) Dependencies() []string        { return []string{} }

func (s *NodeJS) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing Node.js...", "sudo", "pacman", "-S", "--noconfirm", "nodejs", "npm")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "nodejs",
		Name:        "Node.js",
		Description: "Runtime de JavaScript con npm",
		Category:    slots.SlotDEV,
		SubCategory: "languages",
		Deps:        []string{},
	})
}
