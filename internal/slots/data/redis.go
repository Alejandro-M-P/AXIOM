// Package data contains database installation items.
package data

import (
	"context"

	"axiom/internal/domain"
	"axiom/internal/slots"
)

// Redis implements the Slot interface for Redis.
type Redis struct{}

func (s *Redis) ID() string                    { return "redis" }
func (s *Redis) Name() string                  { return "Redis" }
func (s *Redis) Description() string           { return "Base de datos en memoria, cache y message broker" }
func (s *Redis) Category() domain.SlotCategory { return domain.SlotDATA }
func (s *Redis) SubCategory() string           { return "databases" }
func (s *Redis) Dependencies() []string        { return []string{} }

func (s *Redis) Install(ctx context.Context, exec domain.Executor) error {
	return exec(ctx, "Installing Redis...", "sudo", "pacman", "-S", "--noconfirm", "redis")
}

func init() {
	slots.RegisterItem(&slots.SlotItem{
		ID:          "redis",
		Name:        "Redis",
		Description: "Base de datos en memoria, cache y message broker",
		Category:    slots.SlotDATA,
		SubCategory: "databases",
		Deps:        []string{},
	})
}
