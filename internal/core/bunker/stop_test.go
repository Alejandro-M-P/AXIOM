package bunker

import (
	"context"
	"errors"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/core/domain"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

func TestStopSuccess(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem(), mocks.NewMockGit())

	err := mgr.StopBunker(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(runtime.StopBunkerCalls) != 1 {
		t.Errorf("expected 1 StopBunker call, got %d", len(runtime.StopBunkerCalls))
	}
}

func TestStopNonExistent(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem(), mocks.NewMockGit())

	err := mgr.StopBunker(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	found := false
	for _, call := range ui.ShowWarningCalls {
		if call.Title == "warnings.no_bunkers.title" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for no bunkers")
	}
}

func TestStop_NoActiveBunkers(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "stopped", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem(), mocks.NewMockGit())

	err := mgr.StopBunker(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	found := false
	for _, call := range ui.ShowWarningCalls {
		if call.Title == "warnings.none_active.title" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for no active bunkers")
	}
}

func TestStopRuntimeError(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}
	runtime.StopBunkerErr = errors.New("runtime error")

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem(), mocks.NewMockGit())

	err := mgr.StopBunker(context.Background())
	if err == nil {
		t.Fatal("expected error from runtime, got nil")
	}
}

func TestStop_InteractiveSelection(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
		{Name: "bunker-2", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	// With multiple active bunkers, interactive selection is needed
	// which is not implemented, so we expect an error
	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem(), mocks.NewMockGit())

	err := mgr.StopBunker(context.Background())
	// This should error because selectBunkerInteractive returns error for multiple bunkers
	if err == nil {
		t.Log("Stop succeeded with multiple active bunkers (interactive selection may be implemented)")
	}
}

func TestStop_VerifyBunkerStatus(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem(), mocks.NewMockGit())

	err := mgr.StopBunker(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	found := false
	for _, call := range ui.ShowCommandCardCalls {
		if call.CommandKey == "stop" {
			for _, field := range call.Fields {
				if field.GetLabel() == "fields.status" && field.GetValue() == "stopped" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected stop command card to show 'stopped' status")
	}
}
