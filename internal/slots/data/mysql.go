// Package data contains database installation items.
package data

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// MySQL implements the Slot interface for MySQL.
type MySQL struct{}

func (s *MySQL) ID() string                    { return "mysql" }
func (s *MySQL) Name() string                  { return "MySQL" }
func (s *MySQL) Description() string           { return "Sistema de base de datos relacional open source" }
func (s *MySQL) Category() domain.SlotCategory { return domain.SlotDATA }
func (s *MySQL) SubCategory() string           { return "databases" }
func (s *MySQL) Dependencies() []string        { return []string{} }

func (s *MySQL) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing MySQL...", "sudo", "pacman", "-S", "--noconfirm", "mysql")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "mysql",
		Name:        "MySQL",
		Description: "Sistema de base de datos relacional open source",
		Category:    slots.SlotDATA,
		SubCategory: "databases",
		Deps:        []string{},
	})
}
