package bunker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"axiom/internal/domain"
	"axiom/tests/mocks"
)

func TestDeleteSuccess(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	ui.AskDeleteConfirm = true
	ui.AskDeleteDeleteCode = false

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunker(context.Background(), "test-bunker", false, false)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(runtime.RemoveBunkerCalls) != 1 {
		t.Errorf("expected 1 RemoveBunker call, got %d", len(runtime.RemoveBunkerCalls))
	}
	if runtime.RemoveBunkerCalls[0].Name != "test-bunker" {
		t.Errorf("expected RemoveBunker called with 'test-bunker', got '%s'", runtime.RemoveBunkerCalls[0].Name)
	}
}

func TestDeleteWithForce(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	ui.AskDeleteConfirm = true
	ui.AskDeleteDeleteCode = false

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunker(context.Background(), "test-bunker", true, false)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !runtime.RemoveBunkerCalls[0].Force {
		t.Error("expected force flag to be set")
	}
}

func TestDeleteWithImage(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	ui.AskDeleteConfirm = true
	ui.AskDeleteDeleteCode = false

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunker(context.Background(), "test-bunker", false, true)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunker(context.Background(), "nonexistent", false, false)
	if err != nil && err.Error() != "invalid_name" {
		t.Errorf("expected 'invalid_name' error for non-existent bunker, got: %s", err)
	}
}

func TestDeleteUserDeclines(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	ui.AskDeleteConfirm = false

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunker(context.Background(), "test-bunker", false, false)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(runtime.RemoveBunkerCalls) != 0 {
		t.Errorf("expected 0 RemoveBunker calls (user declined), got %d", len(runtime.RemoveBunkerCalls))
	}
}

func TestDeleteRuntimeError(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}
	runtime.RemoveBunkerErr = errors.New("runtime error")

	ui.AskDeleteConfirm = true
	ui.AskDeleteDeleteCode = false

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunker(context.Background(), "test-bunker", false, false)
	if err == nil {
		t.Fatal("expected error from runtime, got nil")
	}
}

func TestDeleteImage_Success(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Images = []string{"localhost/axiom-generic:latest"}

	ui.AskConfirmInCardResp = true

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunkerImage(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestDeleteImage_UserDeclines(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Images = []string{"localhost/axiom-generic:latest"}

	ui.AskConfirmInCardResp = false

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.DeleteBunkerImage(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestListBunkerNames_FromRuntime(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
		{Name: "bunker-2", Status: "stopped", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	names, err := mgr.listBunkerNames(context.Background(), EnvConfig{BaseDir: "/root"})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestListBunkerNames_ExcludesDefault(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
		{Name: defaultBuildContainerName, Status: "running", Image: "localhost/axiom-build:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	names, err := mgr.listBunkerNames(context.Background(), EnvConfig{BaseDir: "/root"})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	for _, name := range names {
		if name == defaultBuildContainerName {
			t.Error("defaultBuildContainerName should be excluded from list")
		}
	}
}

func TestSelectBunkerInteractive_EmptyList(t *testing.T) {
	_, err := selectBunkerInteractive("title", "action", []string{})
	if err == nil {
		t.Error("expected error for empty list")
	}
}

func TestSelectBunkerInteractive_SingleBunker(t *testing.T) {
	name, err := selectBunkerInteractive("title", "action", []string{"only-bunker"})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if name != "only-bunker" {
		t.Errorf("expected 'only-bunker', got '%s'", name)
	}
}

func TestSelectBunkerInteractive_MultipleBunkers(t *testing.T) {
	_, err := selectBunkerInteractive("title", "action", []string{"bunker-1", "bunker-2"})
	if err == nil {
		t.Error("expected error for multiple bunkers (interactive not implemented)")
	}
}

func TestRemoveProjectPath_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}

	err := removeProjectPath(projectDir)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Error("expected directory to be removed")
	}
}

func TestRemoveProjectPath_NotExist(t *testing.T) {
	err := removeProjectPath("/nonexistent/path")
	if err != nil {
		t.Errorf("unexpected error for non-existent path: %s", err)
	}
}

func TestRemoveProjectPath_NotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}
	f.Close()

	err = removeProjectPath(filePath)
	if err == nil {
		t.Error("expected error for non-directory path")
	}
	if err.Error() != "not_dir" {
		t.Errorf("expected 'not_dir' error, got '%s'", err.Error())
	}
}

func TestRemovePathWritable_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir")

	if err := os.MkdirAll(path, 0500); err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}

	err := removePathWritable(path)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected directory to be removed")
	}
}

func TestRemovePathWritable_NotExist(t *testing.T) {
	err := removePathWritable("/nonexistent/path")
	if err != nil {
		t.Errorf("unexpected error for non-existent path: %s", err)
	}
}

func TestAppendTutorLog_Placeholder(t *testing.T) {
	err := appendTutorLog("test log entry")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestListAxiomImages_Function(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
		{Name: "bunker-2", Status: "running", Image: "localhost/axiom-rdna4:latest"},
		{Name: "bunker-3", Status: "running", Image: "docker.io/library/alpine:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	images, err := mgr.listAxiomImages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(images) != 2 {
		t.Errorf("expected 2 images, got %d", len(images))
	}
}

func TestListAxiomPodmanImages_Function(t *testing.T) {
	runtime := mocks.NewMockRuntime()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
		{Name: "bunker-2", Status: "running", Image: "localhost/axiom-rdna4:latest"},
	}

	images, err := listAxiomPodmanImages(context.Background(), runtime)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(images) != 2 {
		t.Errorf("expected 2 images, got %d", len(images))
	}
}
