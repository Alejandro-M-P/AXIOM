// Package sandbox contains the minimal sandbox slot.
package sandbox

// Empty represents a SANDBOX slot - minimal image without additional installations.
// Installation is defined in toml/empty.toml
type Empty struct{}

func (s *Empty) ID() string             { return "empty" }
func (s *Empty) Name() string           { return "Empty" }
func (s *Empty) Description() string    { return "Imagen mínima sin instalaciones adicionales" }
func (s *Empty) Category() string       { return "sandbox" }
func (s *Empty) SubCategory() string    { return "" }
func (s *Empty) Dependencies() []string { return []string{} }
