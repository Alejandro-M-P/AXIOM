// Package data contains database installation items.
package data

// MongoDB implements the Slot interface for MongoDB.
// Installation is defined in toml/mongodb.toml
type MongoDB struct{}

func (s *MongoDB) ID() string             { return "mongodb" }
func (s *MongoDB) Name() string           { return "MongoDB" }
func (s *MongoDB) Description() string    { return "Base de datos NoSQL orientada a documentos" }
func (s *MongoDB) Category() string       { return "data" }
func (s *MongoDB) SubCategory() string    { return "databases" }
func (s *MongoDB) Dependencies() []string { return []string{} }
