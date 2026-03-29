// Package languages contains programming language items for the DEV slot.
package languages

// Python represents the Python programming language.
// Installation is defined in toml/python.toml
type Python struct{}

func (s *Python) ID() string             { return "python" }
func (s *Python) Name() string           { return "Python" }
func (s *Python) Description() string    { return "Lenguaje de programación Python" }
func (s *Python) Category() string       { return "dev" }
func (s *Python) SubCategory() string    { return "languages" }
func (s *Python) Dependencies() []string { return []string{} }
