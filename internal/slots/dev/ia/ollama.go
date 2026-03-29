package ia

import (
	"context"

	"axiom/internal/domain"
)

// Ollama represents the Ollama installation slot.
type Ollama struct{}

func (s *Ollama) ID() string                    { return "ollama" }
func (s *Ollama) Name() string                  { return "Ollama" }
func (s *Ollama) Description() string           { return "Ejecuta modelos LLM localmente (Llama, Mistral, etc.)" }
func (s *Ollama) Category() domain.SlotCategory { return domain.SlotDEV }
func (s *Ollama) SubCategory() string           { return "ia" }
func (s *Ollama) Dependencies() []string        { return []string{} }

func (s *Ollama) Install(ctx context.Context, exec domain.Executor) error {
	if err := exec(ctx, "Downloading Ollama...", "curl", "-fsSL", "https://ollama.com/download/ollama-linux-amd64.tar.zst", "-o", "/tmp/ollama.tar.zst"); err != nil {
		return err
	}
	if err := exec(ctx, "Extracting Ollama to /usr...", "tar", "-xzf", "/tmp/ollama.tar.zst", "-C", "/usr"); err != nil {
		return err
	}
	return exec(ctx, "Cleaning up...", "rm", "-f", "/tmp/ollama.tar.zst")
}
