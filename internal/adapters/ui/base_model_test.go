package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBaseModel_Init(t *testing.T) {
	m := &BaseModel{}

	cmd := m.Init()

	// Init should return a command that produces WindowSizeMsg
	if cmd == nil {
		t.Error("Init() should return a non-nil command")
	}

	// Execute the command and verify it returns a WindowSizeMsg
	msg := cmd()
	if _, ok := msg.(tea.WindowSizeMsg); !ok {
		t.Errorf("Init() command should return tea.WindowSizeMsg, got %T", msg)
	}
}

func TestBaseModel_Update_WindowSizeMsg(t *testing.T) {
	tests := []struct {
		name           string
		initialWidth   int
		initialHeight  int
		msgWidth       int
		msgHeight      int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "update with new dimensions",
			initialWidth:   0,
			initialHeight:  0,
			msgWidth:       80,
			msgHeight:      24,
			expectedWidth:  80,
			expectedHeight: 24,
		},
		{
			name:           "resize to larger dimensions",
			initialWidth:   40,
			initialHeight:  12,
			msgWidth:       120,
			msgHeight:      40,
			expectedWidth:  120,
			expectedHeight: 40,
		},
		{
			name:           "resize to smaller dimensions",
			initialWidth:   100,
			initialHeight:  50,
			msgWidth:       60,
			msgHeight:      20,
			expectedWidth:  60,
			expectedHeight: 20,
		},
		{
			name:           "zero dimensions",
			initialWidth:   0,
			initialHeight:  0,
			msgWidth:       0,
			msgHeight:      0,
			expectedWidth:  0,
			expectedHeight: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BaseModel{
				Width:  tt.initialWidth,
				Height: tt.initialHeight,
			}

			msg := tea.WindowSizeMsg{
				Width:  tt.msgWidth,
				Height: tt.msgHeight,
			}

			newModel, _ := m.Update(msg)

			// Check that dimensions were updated
			if m.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", m.Width, tt.expectedWidth)
			}
			if m.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d", m.Height, tt.expectedHeight)
			}

			// Verify newModel is nil (as per implementation)
			if newModel != nil {
				t.Errorf("Update should return nil model, got %v", newModel)
			}
		})
	}
}

func TestBaseModel_Update_OtherMessages(t *testing.T) {
	// Test that other messages don't change dimensions
	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "key message",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
		},
		{
			name: "type message",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
		},
		{
			name: "mouse message",
			msg:  tea.MouseMsg{X: 5, Y: 10},
		},
		{
			name: "nil message",
			msg:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BaseModel{
				Width:  80,
				Height: 24,
			}

			m.Update(tt.msg)

			// Dimensions should remain unchanged
			if m.Width != 80 {
				t.Errorf("Width changed to %d, expected 80", m.Width)
			}
			if m.Height != 24 {
				t.Errorf("Height changed to %d, expected 24", m.Height)
			}
		})
	}
}

func TestBaseModel_Initialization(t *testing.T) {
	// Test that new BaseModel starts with zero dimensions
	m := NewBaseModel()

	if m.Width != 0 {
		t.Errorf("new BaseModel should have Width=0, got %d", m.Width)
	}

	if m.Height != 0 {
		t.Errorf("new BaseModel should have Height=0, got %d", m.Height)
	}
}

// NewBaseModel is a helper to create BaseModel instances
func NewBaseModel() *BaseModel {
	return &BaseModel{}
}

func TestBaseModel_Embed(t *testing.T) {
	// Test that BaseModel can be embedded in other structs
	type TestModel struct {
		BaseModel
		ExtraField string
	}

	m := &TestModel{
		ExtraField: "test",
	}

	// Should be able to set dimensions via embedded BaseModel
	m.Width = 80
	m.Height = 24

	if m.Width != 80 {
		t.Errorf("embedded Width = %d, want 80", m.Width)
	}
	if m.Height != 24 {
		t.Errorf("embedded Height = %d, want 24", m.Height)
	}

	// Test Update works via embedding
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	m.Update(msg)

	if m.Width != 100 {
		t.Errorf("after Update, Width = %d, want 100", m.Width)
	}
	if m.Height != 40 {
		t.Errorf("after Update, Height = %d, want 40", m.Height)
	}
}
