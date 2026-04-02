// Package ia contains AI/LLM related installation items.
package ia

// Gentle representa el slot de instalación de gentle-ai.
// Installation is defined in toml/gentle.toml
type Gentle struct{}

func (s *Gentle) ID() string             { return "gentle" }
func (s *Gentle) Name() string           { return "gentle-ai" }
func (s *Gentle) Description() string    { return "Framework de agentes AI con memoria persistente" }
func (s *Gentle) Category() string       { return "dev" }
func (s *Gentle) SubCategory() string    { return "ia" }
func (s *Gentle) Dependencies() []string { return []string{} }
