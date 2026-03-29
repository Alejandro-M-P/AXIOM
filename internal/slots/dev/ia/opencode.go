// Package ia contains AI/LLM related installation items.
package ia

// Opencode representa el slot de instalación de opencode-ai.
// Installation is defined in toml/opencode.toml
type Opencode struct{}

func (s *Opencode) ID() string             { return "opencode" }
func (s *Opencode) Name() string           { return "opencode-ai" }
func (s *Opencode) Description() string    { return "Asistente de código AI basado en LLMs locales" }
func (s *Opencode) Category() string       { return "dev" }
func (s *Opencode) SubCategory() string    { return "ia" }
func (s *Opencode) Dependencies() []string { return []string{} }
