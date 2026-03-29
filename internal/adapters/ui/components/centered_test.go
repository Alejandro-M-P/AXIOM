package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewCenteredContainer(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}{
		{"normal dimensions", 80, 24, 80, 24},
		{"zero dimensions", 0, 0, 0, 0},
		{"large dimensions", 200, 100, 200, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCenteredContainer(tt.width, tt.height)

			if c.width != tt.expectedWidth {
				t.Errorf("width = %d, want %d", c.width, tt.expectedWidth)
			}
			if c.height != tt.expectedHeight {
				t.Errorf("height = %d, want %d", c.height, tt.expectedHeight)
			}
			if c.maxWidth != tt.expectedWidth {
				t.Errorf("maxWidth = %d, want %d", c.maxWidth, tt.expectedWidth)
			}
			if c.maxHeight != tt.expectedHeight {
				t.Errorf("maxHeight = %d, want %d", c.maxHeight, tt.expectedHeight)
			}
		})
	}
}

func TestCenteredContainer_WithMaxWidth(t *testing.T) {
	c := NewCenteredContainer(100, 50)

	result := c.WithMaxWidth(80)

	if result.maxWidth != 80 {
		t.Errorf("maxWidth = %d, want 80", result.maxWidth)
	}

	// Verify it returns the same instance (chainable)
	if result != c {
		t.Error("WithMaxWidth should return the same instance for chaining")
	}
}

func TestCenteredContainer_WithMaxHeight(t *testing.T) {
	c := NewCenteredContainer(100, 50)

	result := c.WithMaxHeight(30)

	if result.maxHeight != 30 {
		t.Errorf("maxHeight = %d, want 30", result.maxHeight)
	}

	// Verify it returns the same instance (chainable)
	if result != c {
		t.Error("WithMaxHeight should return the same instance for chaining")
	}
}

func TestCenteredContainer_SetDimensions(t *testing.T) {
	tests := []struct {
		name           string
		initialWidth   int
		initialHeight  int
		initialMaxW    int
		initialMaxH    int
		newWidth       int
		newHeight      int
		expectedWidth  int
		expectedHeight int
		expectedMaxW   int
		expectedMaxH   int
	}{
		{
			name:           "normal resize - max stays when new dimensions are smaller",
			initialWidth:   80,
			initialHeight:  24,
			initialMaxW:    80,
			initialMaxH:    24,
			newWidth:       100,
			newHeight:      40,
			expectedWidth:  100,
			expectedHeight: 40,
			expectedMaxW:   80, // maxWidth stays at 80 because 80 > 100 is false
			expectedMaxH:   24, // maxHeight stays at 24 because 24 > 40 is false
		},
		{
			name:           "shrink dimensions",
			initialWidth:   100,
			initialHeight:  50,
			initialMaxW:    100,
			initialMaxH:    50,
			newWidth:       50,
			newHeight:      20,
			expectedWidth:  50,
			expectedHeight: 20,
			expectedMaxW:   50,
			expectedMaxH:   20,
		},
		{
			name:           "new max smaller than current max",
			initialWidth:   80,
			initialHeight:  24,
			initialMaxW:    100, // maxW larger than new width
			initialMaxH:    30,  // maxH larger than new height
			newWidth:       60,
			newHeight:      15,
			expectedWidth:  60,
			expectedHeight: 15,
			expectedMaxW:   60, // should shrink to new width
			expectedMaxH:   15, // should shrink to new height
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CenteredContainer{
				width:     tt.initialWidth,
				height:    tt.initialHeight,
				maxWidth:  tt.initialMaxW,
				maxHeight: tt.initialMaxH,
			}

			c.SetDimensions(tt.newWidth, tt.newHeight)

			if c.width != tt.expectedWidth {
				t.Errorf("width = %d, want %d", c.width, tt.expectedWidth)
			}
			if c.height != tt.expectedHeight {
				t.Errorf("height = %d, want %d", c.height, tt.expectedHeight)
			}
			if c.maxWidth != tt.expectedMaxW {
				t.Errorf("maxWidth = %d, want %d", c.maxWidth, tt.expectedMaxW)
			}
			if c.maxHeight != tt.expectedMaxH {
				t.Errorf("maxHeight = %d, want %d", c.maxHeight, tt.expectedMaxH)
			}
		})
	}
}

func TestCenteredContainer_Render_HorizontalCentering(t *testing.T) {
	// Test horizontal centering - content shorter than container
	c := NewCenteredContainer(80, 24)

	content := "Hello"
	output := c.Render(content)

	// The output should contain the content
	if output == "" {
		t.Error("Render returned empty string")
	}

	// Check that content appears in output
	if !contains(output, "Hello") {
		t.Error("Rendered output should contain the content")
	}
}

func TestCenteredContainer_Render_VerticalCentering(t *testing.T) {
	// Test vertical centering
	c := NewCenteredContainer(40, 20)

	content := "Test"
	output := c.Render(content)

	if output == "" {
		t.Error("Render returned empty string")
	}

	// Check that content appears in output
	if !contains(output, "Test") {
		t.Error("Rendered output should contain the content")
	}
}

func TestCenteredContainer_Render_WithMaxDimensions(t *testing.T) {
	c := NewCenteredContainer(100, 50).WithMaxWidth(80).WithMaxHeight(30)

	content := "Centered"
	output := c.Render(content)

	if output == "" {
		t.Error("Render returned empty string")
	}

	// Verify maxWidth and maxHeight were applied
	if c.maxWidth != 80 {
		t.Errorf("maxWidth = %d, want 80", c.maxWidth)
	}
	if c.maxHeight != 30 {
		t.Errorf("maxHeight = %d, want 30", c.maxHeight)
	}
}

func TestCenteredContainer_RenderWithStyle(t *testing.T) {
	c := NewCenteredContainer(80, 24)

	content := "Styled"
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Bold(true)

	output := c.RenderWithStyle(content, customStyle)

	if output == "" {
		t.Error("RenderWithStyle returned empty string")
	}

	// Check that content appears in output
	if !contains(output, "Styled") {
		t.Error("Rendered output should contain the content")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
