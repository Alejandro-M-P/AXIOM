// Package data contains database installation items.
package data

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// SQLite implements the Slot interface for SQLite.
type SQLite struct{}

func (s *SQLite) ID() string                    { return "sqlite" }
func (s *SQLite) Name() string                  { return "SQLite" }
func (s *SQLite) Description() string           { return "Base de datos embebida, sin servidor" }
func (s *SQLite) Category() domain.SlotCategory { return domain.SlotDATA }
func (s *SQLite) SubCategory() string           { return "databases" }
func (s *SQLite) Dependencies() []string        { return []string{} }

func (s *SQLite) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing SQLite...", "sudo", "pacman", "-S", "--noconfirm", "sqlite")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "sqlite",
		Name:        "SQLite",
		Description: "Base de datos embebida, sin servidor",
		Category:    slots.SlotDATA,
		SubCategory: "databases",
		Deps:        []string{},
	})
}
