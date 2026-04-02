// Package data contains database installation items.
package data

// SQLite implements the Slot interface for SQLite.
// Installation is defined in toml/sqlite.toml
type SQLite struct{}

func (s *SQLite) ID() string             { return "sqlite" }
func (s *SQLite) Name() string           { return "SQLite" }
func (s *SQLite) Description() string    { return "Base de datos embebida, sin servidor" }
func (s *SQLite) Category() string       { return "data" }
func (s *SQLite) SubCategory() string    { return "databases" }
func (s *SQLite) Dependencies() []string { return []string{} }
