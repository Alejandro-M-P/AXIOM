// Package data contains database installation items.
package data

// Postgres implements the Slot interface for PostgreSQL.
// Installation is defined in toml/postgres.toml
type Postgres struct{}

func (s *Postgres) ID() string             { return "postgres" }
func (s *Postgres) Name() string           { return "PostgreSQL" }
func (s *Postgres) Description() string    { return "Sistema de base de datos relacional" }
func (s *Postgres) Category() string       { return "data" }
func (s *Postgres) SubCategory() string    { return "databases" }
func (s *Postgres) Dependencies() []string { return []string{} }
