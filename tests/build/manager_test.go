package build_test

import (
	"context"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/core/build"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

// mockBuildInstaller implements ports.IBuildInstaller for testing.
type mockBuildInstaller struct {
	executed   bool
	lastItems  []ports.BuildItem
	lastCfg    ports.BuildConfig
	executeErr error
}

func (m *mockBuildInstaller) ExecuteBuild(ctx context.Context, items []ports.BuildItem, containerName string, cfg ports.BuildConfig, progress ports.IBuildProgress, slotManager ports.SlotManagerInterface) error {
	m.executed = true
	m.lastItems = items
	m.lastCfg = cfg
	return m.executeErr
}

func TestNewManager(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	mockSystem := mocks.NewMockSystem()
	buildContainer := "test-container"
	buildInstaller := &mockBuildInstaller{}

	mgr := build.NewManager(runtime, fs, ui, mockSystem, buildContainer, nil, buildInstaller)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestManagerFields(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	mockSystem := mocks.NewMockSystem()
	buildContainer := "test-container"
	buildInstaller := &mockBuildInstaller{}

	mgr := build.NewManager(runtime, fs, ui, mockSystem, buildContainer, nil, buildInstaller)

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestManagerGetUI(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	mockSystem := mocks.NewMockSystem()
	buildContainer := "test-container"
	buildInstaller := &mockBuildInstaller{}

	mgr := build.NewManager(runtime, fs, ui, mockSystem, buildContainer, nil, buildInstaller)

	if mgr.GetUI() != ui {
		t.Fatal("GetUI should return the presenter passed to NewManager")
	}
}

func TestBuildPlanStruct(t *testing.T) {
	plan := &build.BuildPlan{
		Title:    "Test Plan",
		Subtitle: "Test Subtitle",
		Steps: []build.BuildStep{
			{
				Title:  "Step 1",
				Detail: "Detail 1",
				Exec:   func(ctx context.Context) error { return nil },
			},
			{
				Title:  "Step 2",
				Detail: "Detail 2",
				Exec:   func(ctx context.Context) error { return nil },
			},
		},
		Cleanup:   func() {},
		OnSuccess: func() {},
	}

	if plan.Title != "Test Plan" {
		t.Errorf("Title = %q, want %q", plan.Title, "Test Plan")
	}
	if len(plan.Steps) != 2 {
		t.Errorf("Steps length = %d, want 2", len(plan.Steps))
	}
	if plan.Steps[0].Title != "Step 1" {
		t.Errorf("Step 0 Title = %q, want %q", plan.Steps[0].Title, "Step 1")
	}
	if plan.Cleanup == nil {
		t.Error("Cleanup should not be nil")
	}
	if plan.OnSuccess == nil {
		t.Error("OnSuccess should not be nil")
	}
}

func TestBuildPlanExecution(t *testing.T) {
	execCount := 0
	plan := &build.BuildPlan{
		Title:    "Test",
		Subtitle: "Test",
		Steps: []build.BuildStep{
			{
				Title: "Step 1",
				Exec: func(ctx context.Context) error {
					execCount++
					return nil
				},
			},
		},
	}

	// Execute the plan manually (simulating what the router does)
	ctx := context.Background()
	for i, step := range plan.Steps {
		if err := step.Exec(ctx); err != nil {
			t.Fatalf("Step %d failed: %v", i, err)
		}
	}

	if execCount != 1 {
		t.Errorf("Exec count = %d, want 1", execCount)
	}
}

func TestBuildPlanFailure(t *testing.T) {
	cleanupCalled := false
	plan := &build.BuildPlan{
		Title:    "Test",
		Subtitle: "Test",
		Steps: []build.BuildStep{
			{
				Title: "Step 1",
				Exec: func(ctx context.Context) error {
					return nil
				},
			},
			{
				Title: "Step 2",
				Exec: func(ctx context.Context) error {
					return context.DeadlineExceeded
				},
			},
		},
		Cleanup: func() {
			cleanupCalled = true
		},
	}

	// Simulate execution with failure
	ctx := context.Background()
	for i, step := range plan.Steps {
		if err := step.Exec(ctx); err != nil {
			if plan.Cleanup != nil {
				plan.Cleanup()
			}
			// Expected failure at step 1
			if i != 1 {
				t.Errorf("Expected failure at step 1, got step %d", i)
			}
			break
		}
	}

	if !cleanupCalled {
		t.Error("Cleanup should have been called on failure")
	}
}

// TestBuildReturnsPlan verifies that Build() returns a BuildPlan when slot selections exist.
func TestBuildReturnsPlan(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	mockSystem := mocks.NewMockSystem()
	buildContainer := "test-container"
	buildInstaller := &mockBuildInstaller{}

	// Create a mock slot manager that returns selections
	mockSlotManager := &mockSlotManagerWithSelection{
		selections: []ports.SlotSelection{{Slot: "sandbox", Selected: []string{}}},
	}

	mgr := build.NewManager(runtime, fs, ui, mockSystem, buildContainer, mockSlotManager, buildInstaller)

	ctx := context.Background()
	cfg := config.EnvConfig{
		BaseDir:   "/tmp/test",
		AxiomPath: "/tmp/test/axiom",
		ROCMMode:  "",
		GPUType:   "generic",
	}

	plan, err := mgr.Build(ctx, cfg)
	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}
	if plan == nil {
		t.Fatal("Build() should return a BuildPlan, got nil")
	}
	if plan.Title == "" {
		t.Error("BuildPlan Title should not be empty")
	}
	if len(plan.Steps) == 0 {
		t.Error("BuildPlan should have at least one step")
	}
	// Sandbox has only 4 steps: prepare_dirs, recreate_container, install_base, export_image
	if len(plan.Steps) != 4 {
		t.Errorf("Sandbox BuildPlan should have 4 steps, got %d", len(plan.Steps))
	}
}

// TestBuildInstallerIsCalled verifies that the installer's ExecuteBuild is called during build.
func TestBuildInstallerIsCalled(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	mockSystem := mocks.NewMockSystem()
	buildContainer := "test-container"
	buildInstaller := &mockBuildInstaller{}

	mockSlotManager := &mockSlotManagerWithSelection{
		selections: []ports.SlotSelection{{Slot: "dev", Selected: []string{"ollama"}}},
	}

	mgr := build.NewManager(runtime, fs, ui, mockSystem, buildContainer, mockSlotManager, buildInstaller)

	ctx := context.Background()
	cfg := config.EnvConfig{
		BaseDir:   "/tmp/test",
		AxiomPath: "/tmp/test/axiom",
		ROCMMode:  "",
		GPUType:   "generic",
	}

	plan, err := mgr.Build(ctx, cfg)
	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}

	// Execute the install_base step (step index 2) to trigger the installer
	if len(plan.Steps) < 3 {
		t.Fatalf("Expected at least 3 steps, got %d", len(plan.Steps))
	}

	installStep := plan.Steps[2]
	if err := installStep.Exec(ctx); err != nil {
		t.Fatalf("Install step failed: %v", err)
	}

	if !buildInstaller.executed {
		t.Error("BuildInstaller.ExecuteBuild should have been called")
	}
}

// mockSlotManagerWithSelection implements build.SlotManagerInterface for testing.
type mockSlotManagerWithSelection struct {
	selections []ports.SlotSelection
}

func (m *mockSlotManagerWithSelection) HasSelection() bool {
	return len(m.selections) > 0
}

func (m *mockSlotManagerWithSelection) GetSelectedItems(category string) ([]ports.SlotItem, error) {
	return nil, nil
}

func (m *mockSlotManagerWithSelection) RunSlotSelector(category string, items []ports.SlotItem, preselected []string) ([]string, bool, error) {
	return nil, true, nil
}

func (m *mockSlotManagerWithSelection) SaveSelection(selections []ports.SlotSelection) error {
	m.selections = selections
	return nil
}

func (m *mockSlotManagerWithSelection) LoadSelection() ([]ports.SlotSelection, error) {
	return m.selections, nil
}
