// Package tools contains developer tools installation items.
package tools

// Starship is the minimalist prompt for any shell.
// Installation is defined in toml/starship.toml
type Starship struct{}

func (s *Starship) ID() string             { return "starship" }
func (s *Starship) Name() string           { return "Starship" }
func (s *Starship) Description() string    { return "Prompt minimalista para cualquier shell" }
func (s *Starship) Category() string       { return "dev" }
func (s *Starship) SubCategory() string    { return "tools" }
func (s *Starship) Dependencies() []string { return []string{} }
