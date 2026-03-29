// Package ia contains AI/LLM related installation items.
package ia

// Ollama represents the Ollama installation slot.
// Installation is defined in toml/ollama.toml
type Ollama struct{}

func (s *Ollama) ID() string             { return "ollama" }
func (s *Ollama) Name() string           { return "Ollama" }
func (s *Ollama) Description() string    { return "Ejecuta modelos LLM localmente (Llama, Mistral, etc.)" }
func (s *Ollama) Category() string       { return "dev" }
func (s *Ollama) SubCategory() string    { return "ia" }
func (s *Ollama) Dependencies() []string { return []string{} }
