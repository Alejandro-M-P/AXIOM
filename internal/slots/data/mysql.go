// Package data contains database installation items.
package data

// MySQL implements the Slot interface for MySQL.
// Installation is defined in toml/mysql.toml
type MySQL struct{}

func (s *MySQL) ID() string             { return "mysql" }
func (s *MySQL) Name() string           { return "MySQL" }
func (s *MySQL) Description() string    { return "Sistema de base de datos relacional open source" }
func (s *MySQL) Category() string       { return "data" }
func (s *MySQL) SubCategory() string    { return "databases" }
func (s *MySQL) Dependencies() []string { return []string{} }
