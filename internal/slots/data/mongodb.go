// Package data contains database installation items.
package data

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// MongoDB implements the Slot interface for MongoDB.
type MongoDB struct{}

func (s *MongoDB) ID() string                    { return "mongodb" }
func (s *MongoDB) Name() string                  { return "MongoDB" }
func (s *MongoDB) Description() string           { return "Base de datos NoSQL orientada a documentos" }
func (s *MongoDB) Category() domain.SlotCategory { return domain.SlotDATA }
func (s *MongoDB) SubCategory() string           { return "databases" }
func (s *MongoDB) Dependencies() []string        { return []string{} }

func (s *MongoDB) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing MongoDB...", "sudo", "pacman", "-S", "--noconfirm", "mongodb")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "mongodb",
		Name:        "MongoDB",
		Description: "Base de datos NoSQL orientada a documentos",
		Category:    slots.SlotCategory(domain.SlotDATA),
		SubCategory: "databases",
		Deps:        []string{},
	})
}
