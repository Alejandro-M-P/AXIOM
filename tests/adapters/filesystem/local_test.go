package filesystem_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/filesystem"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

func TestLocalAdapterImplementsIFileSystem(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	var _ ports.IFileSystem = adapter
}

func TestNewFSAdapter(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	if adapter == nil {
		t.Fatal("NewFSAdapter returned nil")
	}
}

func TestReadFile(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")

	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	data, err := adapter.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestWriteFile(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "test.txt")
	content := []byte("test content")

	err := adapter.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestWriteFileCreatesDir(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "a", "b", "c", "nested.txt")
	content := []byte("nested content")

	err := adapter.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(testFile)); os.IsNotExist(err) {
		t.Error("parent directory was not created")
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestWriteFileWithPermissions(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "perms.txt")
	content := []byte("perms test")
	perm := os.FileMode(0755)

	err := adapter.WriteFile(testFile, content, perm)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if info.Mode() != perm {
		t.Errorf("expected permission %v, got %v", perm, info.Mode())
	}
}

func TestMkdirAll(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "a", "b", "c")

	err := adapter.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	if !adapter.Exists(testDir) {
		t.Error("directory was not created")
	}

	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("path is not a directory")
	}
}

func TestMkdirAll_ExistingDir(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()

	err := adapter.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll on existing dir failed: %v", err)
	}
}

func TestExists_File(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "exists.txt")

	if adapter.Exists(testFile) {
		t.Error("file should not exist before creation")
	}

	_, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !adapter.Exists(testFile) {
		t.Error("file should exist after creation")
	}
}

func TestExists_Directory(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "existsdir")

	if adapter.Exists(testDir) {
		t.Error("directory should not exist before creation")
	}

	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if !adapter.Exists(testDir) {
		t.Error("directory should exist after creation")
	}
}

func TestExists_NotFound(t *testing.T) {
	adapter := filesystem.NewFSAdapter()

	if adapter.Exists("/nonexistent/path/that/does/not/exist") {
		t.Error("nonexistent path should return false")
	}
}

func TestExists_AfterDelete(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "todelete.txt")

	_, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !adapter.Exists(testFile) {
		t.Error("file should exist before delete")
	}

	err = os.Remove(testFile)
	if err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	if adapter.Exists(testFile) {
		t.Error("file should not exist after delete")
	}
}

func TestRemove(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "remove.txt")

	_, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	err = os.Remove(testFile)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if adapter.Exists(testFile) {
		t.Error("file should not exist after remove")
	}
}

func TestRemove_Directory(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "removeme")

	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	err = os.Remove(testDir)
	if err != nil {
		t.Fatalf("Remove directory failed: %v", err)
	}

	if adapter.Exists(testDir) {
		t.Error("directory should not exist after remove")
	}
}

func TestReadNonexistent(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	nonexistent := filepath.Join(os.TempDir(), "this_file_does_not_exist_12345.txt")

	_, err := adapter.ReadFile(nonexistent)
	if err == nil {
		t.Error("expected error when reading nonexistent file")
	}
}

func TestReadNonexistent_ErrorType(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	nonexistent := filepath.Join(os.TempDir(), "nonexistent_file_xyz.txt")

	_, err := adapter.ReadFile(nonexistent)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected os.IsNotExist error, got: %v", err)
	}
}

func TestStat(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "stat.txt")
	content := []byte("stat test")

	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	info, err := adapter.Stat(testFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Name() != "stat.txt" {
		t.Errorf("expected name 'stat.txt', got %q", info.Name())
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), info.Size())
	}
	if info.IsDir() {
		t.Error("file should not be a directory")
	}
}

func TestStat_Directory(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()

	info, err := adapter.Stat(tmpDir)
	if err != nil {
		t.Fatalf("Stat on directory failed: %v", err)
	}
	if !info.IsDir() {
		t.Error("path should be a directory")
	}
}

func TestStat_Nonexistent(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	nonexistent := filepath.Join(os.TempDir(), "nonexistent_stat_file_abc.txt")

	_, err := adapter.Stat(nonexistent)
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected os.IsNotExist error, got: %v", err)
	}
}

func TestReadDir(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	dir1 := filepath.Join(tmpDir, "subdir")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)
	os.Mkdir(dir1, 0755)

	entries, err := adapter.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestReadDir_Nonexistent(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	nonexistent := filepath.Join(os.TempDir(), "nonexistent_dir_xyz.txt")

	_, err := adapter.ReadDir(nonexistent)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestReadDir_EmptyDirectory(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()

	entries, err := adapter.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty dir, got %d", len(entries))
	}
}

func TestOpenFile(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "openfile.txt")
	content := []byte("open file test")

	err := adapter.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	file, err := adapter.OpenFile(testFile, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read from opened file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestOpenFile_ReadOnly(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "readonly.txt")

	_, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	file, err := adapter.OpenFile(testFile, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile read-only failed: %v", err)
	}
	defer file.Close()

	var stat os.FileInfo
	stat, err = file.Stat()
	if err != nil {
		t.Fatalf("failed to stat opened file: %v", err)
	}
	if stat.Mode()&os.ModeAppend != 0 {
		t.Error("file should not have append flag")
	}
}

func TestOpenFile_Nonexistent(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	nonexistent := filepath.Join(os.TempDir(), "nonexistent_file_open_123.txt")

	_, err := adapter.OpenFile(nonexistent, os.O_RDONLY, 0644)
	if err == nil {
		t.Error("expected error opening nonexistent file")
	}
}

func TestWalkDir(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "f1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(subDir, "f2.txt"), []byte("2"), 0644)

	var visited []string
	err := adapter.WalkDir(tmpDir, func(path string, d ports.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visited = append(visited, path)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}
	if len(visited) < 3 {
		t.Errorf("expected at least 3 paths visited, got %d", len(visited))
	}
}

func TestWalkDir_Error(t *testing.T) {
	adapter := filesystem.NewFSAdapter()

	err := adapter.WalkDir("/nonexistent/path/for/walking", func(path string, d ports.DirEntry, err error) error {
		return err
	})
	if err == nil {
		t.Error("expected error walking nonexistent path")
	}
}

func TestChmod(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "chmod.txt")

	_, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	newPerm := os.FileMode(0755)
	err = adapter.Chmod(testFile, newPerm)
	if err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat after chmod: %v", err)
	}
	if info.Mode() != newPerm {
		t.Errorf("expected permission %v, got %v", newPerm, info.Mode())
	}
}

func TestChmod_Nonexistent(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	nonexistent := filepath.Join(os.TempDir(), "chmod_nonexistent_xyz.txt")

	err := adapter.Chmod(nonexistent, 0644)
	if err == nil {
		t.Error("expected error chmod nonexistent file")
	}
}

func TestRemoveAll_Nonexistent(t *testing.T) {
	adapter := filesystem.NewFSAdapter()
	nonexistent := filepath.Join(os.TempDir(), "removeall_nonexistent_xyz_123.txt")

	err := adapter.RemoveAll(nonexistent)
	if err != nil {
		t.Fatalf("RemoveAll on nonexistent path failed: %v", err)
	}
}
