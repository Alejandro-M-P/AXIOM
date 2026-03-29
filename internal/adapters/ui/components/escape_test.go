package components

import (
	"testing"
)

func TestNewEscapeButton(t *testing.T) {
	btn := NewEscapeButton()

	if btn.text != "Salir" {
		t.Errorf("text = %q, want %q", btn.text, "Salir")
	}

	if btn.theme == nil {
		t.Error("theme should not be nil")
	}
}

func TestEscapeButton_WithText(t *testing.T) {
	tests := []struct {
		name         string
		initialText  string
		newText      string
		expectedText string
		returnsSame  bool
	}{
		{
			name:         "custom text",
			initialText:  "Salir",
			newText:      "Cancelar",
			expectedText: "Cancelar",
			returnsSame:  true,
		},
		{
			name:         "empty text keeps original",
			initialText:  "Salir",
			newText:      "",
			expectedText: "Salir",
			returnsSame:  true,
		},
		{
			name:         "whitespace text is allowed (current behavior)",
			initialText:  "Salir",
			newText:      "   ",
			expectedText: "   ",
			returnsSame:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := &EscapeButton{text: tt.initialText}

			result := btn.WithText(tt.newText)

			if btn.text != tt.expectedText {
				t.Errorf("text = %q, want %q", btn.text, tt.expectedText)
			}

			if tt.returnsSame && result != btn {
				t.Error("WithText should return the same instance for chaining")
			}
		})
	}
}

func TestEscapeButton_Render(t *testing.T) {
	btn := NewEscapeButton()
	output := btn.Render()

	// Should contain [ESC] indicator
	if !contains(output, "[ESC]") {
		t.Errorf("Render() should contain [ESC], got %q", output)
	}

	// Should contain default text "Salir"
	if !contains(output, "Salir") {
		t.Errorf("Render() should contain 'Salir', got %q", output)
	}

	// Should not be empty
	if output == "" {
		t.Error("Render() should not return empty string")
	}
}

func TestEscapeButton_Render_WithCustomText(t *testing.T) {
	btn := NewEscapeButton().WithText("Cancelar")
	output := btn.Render()

	// Should contain [ESC] indicator
	if !contains(output, "[ESC]") {
		t.Errorf("Render() should contain [ESC], got %q", output)
	}

	// Should contain custom text
	if !contains(output, "Cancelar") {
		t.Errorf("Render() should contain 'Cancelar', got %q", output)
	}
}

func TestEscapeButton_RenderAlt(t *testing.T) {
	btn := NewEscapeButton()
	output := btn.RenderAlt()

	// Should contain [q] indicator instead of [ESC]
	if !contains(output, "[q]") {
		t.Errorf("RenderAlt() should contain [q], got %q", output)
	}

	// Should contain default text
	if !contains(output, "Salir") {
		t.Errorf("RenderAlt() should contain 'Salir', got %q", output)
	}
}

func TestEscapeButton_RenderCompact(t *testing.T) {
	btn := NewEscapeButton()
	output := btn.RenderCompact()

	// Should contain ESC (without brackets)
	if !contains(output, "ESC") {
		t.Errorf("RenderCompact() should contain 'ESC', got %q", output)
	}

	// Should contain default text
	if !contains(output, "Salir") {
		t.Errorf("RenderCompact() should contain 'Salir', got %q", output)
	}
}

func TestEscapeButton_RenderAlt_WithCustomText(t *testing.T) {
	btn := NewEscapeButton().WithText("Volver")
	output := btn.RenderAlt()

	// Should contain custom text
	if !contains(output, "Volver") {
		t.Errorf("RenderAlt() should contain 'Volver', got %q", output)
	}

	// Should still use [q] indicator
	if !contains(output, "[q]") {
		t.Errorf("RenderAlt() should contain [q], got %q", output)
	}
}

func TestEscapeButton_RenderCompact_WithCustomText(t *testing.T) {
	btn := NewEscapeButton().WithText("Atrás")
	output := btn.RenderCompact()

	// Should contain custom text
	if !contains(output, "Atrás") {
		t.Errorf("RenderCompact() should contain 'Atrás', got %q", output)
	}

	// Should still use ESC indicator
	if !contains(output, "ESC") {
		t.Errorf("RenderCompact() should contain 'ESC', got %q", output)
	}
}

// Test that styling is applied (basic smoke test)
func TestEscapeButton_StylingApplied(t *testing.T) {
	btn := NewEscapeButton()

	// Test Render returns styled output
	output := btn.Render()

	// The output should be longer than just "[ESC] Salir" because of styling
	// (lipgloss adds ANSI codes)
	if len(output) <= 9 { // len("[ESC] Salir") = 9
		t.Errorf("Render() output seems too short, might not have styling: %q", output)
	}
}
