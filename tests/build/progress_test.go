package build_test

import (
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/core/build"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

func TestProgressStartStep(t *testing.T) {
	ui := mocks.NewMockPresenter()
	steps := []ports.LifecycleStep{
		{Title: "Step 1", Status: ports.LifecyclePending},
		{Title: "Step 2", Status: ports.LifecyclePending},
		{Title: "Step 3", Status: ports.LifecyclePending},
	}

	progress := build.NewProgress(ui, "Test Title", "Test Subtitle", steps)

	// Start step 1
	progress.StartStep(1)

	if progress.CurrentStep() != 1 {
		t.Errorf("CurrentStep = %d, want 1", progress.CurrentStep())
	}

	if progress.TotalSteps() != 3 {
		t.Errorf("TotalSteps = %d, want 3", progress.TotalSteps())
	}
}

func TestProgressFinishStep(t *testing.T) {
	ui := mocks.NewMockPresenter()
	steps := []ports.LifecycleStep{
		{Title: "Step 1", Status: ports.LifecyclePending},
		{Title: "Step 2", Status: ports.LifecyclePending},
	}

	progress := build.NewProgress(ui, "Test Title", "Test Subtitle", steps)

	// Start and finish step 0
	progress.StartStep(0)
	progress.FinishStep()

	if progress.CurrentStep() != 0 {
		t.Errorf("CurrentStep = %d after FinishStep, want 0", progress.CurrentStep())
	}
}

func TestProgressAppendOutput(t *testing.T) {
	ui := mocks.NewMockPresenter()
	steps := []ports.LifecycleStep{
		{Title: "Step 1", Status: ports.LifecyclePending},
	}

	progress := build.NewProgress(ui, "Test Title", "Test Subtitle", steps)

	// AppendOutput should not panic
	progress.AppendOutput("test output line")

	// The method exists and does nothing in the current implementation
	// This test verifies it doesn't panic
}

func TestProgressRender(t *testing.T) {
	ui := mocks.NewMockPresenter()
	steps := []ports.LifecycleStep{
		{Title: "Step 1", Status: ports.LifecyclePending},
		{Title: "Step 2", Status: ports.LifecyclePending},
	}

	progress := build.NewProgress(ui, "Test Title", "Test Subtitle", steps)

	// Test that NewProgress doesn't panic and returns valid progress
	if progress == nil {
		t.Fatal("NewProgress returned nil")
	}

	// Test exported methods work correctly
	if progress.TotalSteps() != 2 {
		t.Errorf("TotalSteps = %d, want 2", progress.TotalSteps())
	}

	if progress.CurrentStep() != 0 {
		t.Errorf("CurrentStep = %d, want 0", progress.CurrentStep())
	}
}

func TestProgressMultipleSteps(t *testing.T) {
	ui := mocks.NewMockPresenter()
	steps := []ports.LifecycleStep{
		{Title: "Step 1", Status: ports.LifecyclePending},
		{Title: "Step 2", Status: ports.LifecyclePending},
		{Title: "Step 3", Status: ports.LifecyclePending},
	}

	progress := build.NewProgress(ui, "Test Title", "Test Subtitle", steps)

	// Run through multiple steps
	for i := 0; i < 3; i++ {
		progress.StartStep(i)
		progress.FinishStep()
	}

	if progress.CurrentStep() != 2 {
		t.Errorf("CurrentStep = %d after all steps, want 2", progress.CurrentStep())
	}

	if progress.TotalSteps() != 3 {
		t.Errorf("TotalSteps = %d, want 3", progress.TotalSteps())
	}
}
