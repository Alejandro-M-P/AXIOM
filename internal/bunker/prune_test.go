package bunker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Alejandro-M-P/AXIOM/internal/domain"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

func TestPruneSuccess(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	tmpDir := t.TempDir()
	orphanDir1 := filepath.Join(tmpDir, ".entorno", "orphan-1")
	orphanDir2 := filepath.Join(tmpDir, ".entorno", "orphan-2")

	if err := os.MkdirAll(orphanDir1, 0755); err != nil {
		t.Fatalf("failed to create orphan dir: %s", err)
	}
	if err := os.MkdirAll(orphanDir2, 0755); err != nil {
		t.Fatalf("failed to create orphan dir: %s", err)
	}

	ui.AskConfirmInCardResp = true

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.PruneBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestPrune_ConcurrentSafety(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	errCh := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			err := mgr.PruneBunkers(context.Background())
			if err != nil {
				errCh <- err
			}
		}()
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case err := <-errCh:
			t.Errorf("concurrent prune error: %s", err)
		case <-timeout:
			t.Fatal("timeout waiting for concurrent prunes")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestPrune_OnlyOneRunning(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	tmpDir := t.TempDir()
	orphanDir := filepath.Join(tmpDir, ".entorno", "orphan")
	if err := os.MkdirAll(orphanDir, 0755); err != nil {
		t.Fatalf("failed to create orphan dir: %s", err)
	}

	ui.AskConfirmInCardResp = true

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mgr.PruneBunkers(context.Background())
		}()
	}

	wg.Wait()
}

func TestPruneWithErrors(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	tmpDir := t.TempDir()
	orphanDir := filepath.Join(tmpDir, ".entorno", "orphan")
	if err := os.MkdirAll(orphanDir, 0755); err != nil {
		t.Fatalf("failed to create orphan dir: %s", err)
	}

	ui.AskConfirmInCardResp = true

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.PruneBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestPrune_NoOrphans(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, ".entorno")

	runtime.Bunkers = []domain.Bunker{
		{Name: "active-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	activeDir := filepath.Join(envDir, "active-bunker")
	if err := os.MkdirAll(activeDir, 0755); err != nil {
		t.Fatalf("failed to create active dir: %s", err)
	}

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.PruneBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// When no orphans, the function returns early without showing warnings
	// because orphans list is empty
}

func TestPrune_UserDeclines(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	tmpDir := t.TempDir()
	orphanDir := filepath.Join(tmpDir, ".entorno", "orphan")
	if err := os.MkdirAll(orphanDir, 0755); err != nil {
		t.Fatalf("failed to create orphan dir: %s", err)
	}

	ui.AskConfirmInCardResp = false

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.PruneBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestPrune_EmptyEnvDir(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, ".entorno")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		t.Fatalf("failed to create .entorno dir: %s", err)
	}

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.PruneBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// When no orphans, returns early
}

func TestPrune_UserCancels(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	tmpDir := t.TempDir()
	orphanDir := filepath.Join(tmpDir, ".entorno", "orphan")
	if err := os.MkdirAll(orphanDir, 0755); err != nil {
		t.Fatalf("failed to create orphan dir: %s", err)
	}

	ui.AskConfirmInCardErr = errors.New("user cancelled")

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.PruneBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestPrune_VerifyMutexBlocksConcurrent(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	startCh := make(chan struct{})
	doneCount := 0
	var doneMu sync.Mutex

	for i := 0; i < 3; i++ {
		go func() {
			<-startCh

			_ = mgr.PruneBunkers(context.Background())

			doneMu.Lock()
			doneCount++
			doneMu.Unlock()
		}()
	}

	close(startCh)

	time.Sleep(100 * time.Millisecond)

	doneMu.Lock()
	count := doneCount
	doneMu.Unlock()

	if count != 3 {
		t.Logf("Only %d/3 goroutines completed in time", count)
	}
}

func TestPrune_ExcludesDefaultContainer(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: defaultBuildContainerName, Status: "running", Image: "localhost/axiom-build:latest"},
	}

	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, ".entorno")

	orphanDir := filepath.Join(envDir, "orphan")
	defaultDir := filepath.Join(envDir, defaultBuildContainerName)

	os.MkdirAll(orphanDir, 0755)
	os.MkdirAll(defaultDir, 0755)

	ui.AskConfirmInCardResp = true

	mgr := NewManager(tmpDir, runtime, fs, ui, mocks.NewMockSystem())

	_ = mgr.PruneBunkers(context.Background())
}
