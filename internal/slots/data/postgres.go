// Package data contains database installation items.
package data

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// Postgres implements the Slot interface for PostgreSQL.
type Postgres struct{}

func (s *Postgres) ID() string                    { return "postgres" }
func (s *Postgres) Name() string                  { return "PostgreSQL" }
func (s *Postgres) Description() string           { return "Sistema de base de datos relacional" }
func (s *Postgres) Category() domain.SlotCategory { return domain.SlotDATA }
func (s *Postgres) SubCategory() string           { return "databases" }
func (s *Postgres) Dependencies() []string        { return []string{} }

func (s *Postgres) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing PostgreSQL...", "sudo", "pacman", "-S", "--noconfirm", "postgresql")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "postgres",
		Name:        "PostgreSQL",
		Description: "Sistema de base de datos relacional",
		Category:    slots.SlotCategory(domain.SlotDATA),
		SubCategory: "databases",
		Deps:        []string{},
	})
}
