// Package ia contains AI/LLM related installation items.
package ia

// Engram representa el slot de instalación de Engram.
// Installation is defined in toml/engram.toml
type Engram struct{}

func (s *Engram) ID() string             { return "engram" }
func (s *Engram) Name() string           { return "Engram" }
func (s *Engram) Description() string    { return "Sistema de memoria persistente para agentes de IA" }
func (s *Engram) Category() string       { return "dev" }
func (s *Engram) SubCategory() string    { return "ia" }
func (s *Engram) Dependencies() []string { return []string{} }
