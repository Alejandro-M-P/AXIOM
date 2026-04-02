// Package languages contains programming language items for the DEV slot.
package languages

// Go represents the Go programming language.
// Installation is defined in toml/go.toml
type Go struct{}

func (s *Go) ID() string             { return "go" }
func (s *Go) Name() string           { return "Go" }
func (s *Go) Description() string    { return "Lenguaje de programación Go" }
func (s *Go) Category() string       { return "dev" }
func (s *Go) SubCategory() string    { return "languages" }
func (s *Go) Dependencies() []string { return []string{} }
