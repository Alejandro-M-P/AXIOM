//go:build integration
// +build integration

package ui

import (
	"testing"
	"time"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// TestModel es un modelo de prueba que embebe BaseModel para testing
type TestModel struct {
	BaseModel
	Width  int
	Height int
}

func (m TestModel) Init() tea.Cmd {
	return m.BaseModel.Init()
}

func (m TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.BaseModel.Update(msg)
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m TestModel) View() string {
	container := components.NewCenteredContainer(m.Width, m.Height)
	return container.Render("Test Content")
}

func TestBubbleteaModelWithAltScreen(t *testing.T) {
	// Skip if not in integration test mode
	if !teatest.IsCI() {
		t.Skip("Skipping integration test in non-CI environment")
	}

	m := TestModel{}

	// Create a test model with AltScreen
	tm := teatest.NewTestModel(t, m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Send window size message to initialize dimensions
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Wait for the model to update
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		return m.(TestModel).Width == 80 && m.(TestModel).Height == 24
	})

	// Get final model
	final := tm.FinalModel(t).(TestModel)

	if final.Width != 80 {
		t.Errorf("expected Width=80, got %d", final.Width)
	}
	if final.Height != 24 {
		t.Errorf("expected Height=24, got %d", final.Height)
	}

	// Cleanup
	tm.WaitFinished(t, teatest.WithDuration(time.Second))
}

func TestBubbleteaResizeEvents(t *testing.T) {
	// Skip if not in integration test mode
	if !teatest.IsCI() {
		t.Skip("Skipping integration test in non-CI environment")
	}

	m := TestModel{}

	// Create test model
	tm := teatest.NewTestModel(t, m)

	// Initial size
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Wait for initial size
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		return m.(TestModel).Width == 80 && m.(TestModel).Height == 24
	})

	// Resize to larger
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Wait for resize
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		return m.(TestModel).Width == 120 && m.(TestModel).Height == 40
	})

	// Resize to smaller
	tm.Send(tea.WindowSizeMsg{Width: 60, Height: 20})

	// Wait for resize
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		return m.(TestModel).Width == 60 && m.(TestModel).Height == 20
	})

	// Verify final dimensions
	final := tm.FinalModel(t).(TestModel)

	if final.Width != 60 {
		t.Errorf("expected final Width=60, got %d", final.Width)
	}
	if final.Height != 20 {
		t.Errorf("expected final Height=20, got %d", final.Height)
	}

	// Cleanup
	tm.WaitFinished(t, teatest.WithDuration(time.Second))
}

func TestBubbleteaKeyNavigation(t *testing.T) {
	// Skip if not in integration test mode
	if !teatest.IsCI() {
		t.Skip("Skipping integration test in non-CI environment")
	}

	type NavigableModel struct {
		BaseModel
		Cursor  int
		Options []string
		Width   int
		Height  int
	}

	newNavigableModel := func() NavigableModel {
		return NavigableModel{
			Options: []string{"Option 1", "Option 2", "Option 3"},
		}
	}

	tm := teatest.NewTestModel(t, newNavigableModel())

	// Initialize with window size
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		nm := m.(NavigableModel)
		return nm.Width == 80
	})

	// Send down arrow
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		return m.(NavigableModel).Cursor == 1
	})

	// Send down again
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		return m.(NavigableModel).Cursor == 2
	})

	// Send up arrow
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	tm.WaitForModel(t, 500*time.Millisecond, func(m tea.Model) bool {
		return m.(NavigableModel).Cursor == 1
	})

	// Verify cursor position
	final := tm.FinalModel(t).(NavigableModel)
	if final.Cursor != 1 {
		t.Errorf("expected Cursor=1, got %d", final.Cursor)
	}

	// Cleanup
	tm.WaitFinished(t, teatest.WithDuration(time.Second))
}
