package mocks

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// Compile-time check that MockFileSystem implements ports.IFileSystem
var _ ports.IFileSystem = (*MockFileSystem)(nil)

// MockFileSystem implements ports.IFileSystem for testing.
type MockFileSystem struct {
	mu sync.Mutex

	// Storage
	Dirs  map[string]bool
	Files map[string][]byte

	// Track calls
	MkdirAllCalls     []MkdirAllCall
	RemoveAllCalls    []string
	ReadDirCalls      []string
	StatCalls         []string
	ReadFileCalls     []string
	WriteFileCalls    []WriteFileCall
	OpenFileCalls     []OpenFileCall
	CreateFileCalls   []CreateFileCall
	WalkDirCalls      []string
	ChmodCalls        []ChmodCall
	UserHomeDirCalled bool

	// Configuration
	ShouldMkdirFail bool
	ShouldStatFail  bool
}

type MkdirAllCall struct {
	Path string
	Perm os.FileMode
}

type WriteFileCall struct {
	Path string
	Data []byte
	Perm os.FileMode
}

type OpenFileCall struct {
	Path string
	Flag int
	Perm os.FileMode
}

type ChmodCall struct {
	Path string
	Mode os.FileMode
}

type CreateFileCall struct {
	Path string
	Perm os.FileMode
}

// NewMockFileSystem creates a new MockFileSystem with default values.
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Dirs:           make(map[string]bool),
		Files:          make(map[string][]byte),
		MkdirAllCalls:  []MkdirAllCall{},
		RemoveAllCalls: []string{},
	}
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.MkdirAllCalls = append(m.MkdirAllCalls, MkdirAllCall{Path: path, Perm: perm})

	if m.ShouldMkdirFail {
		return os.ErrPermission
	}
	m.Dirs[path] = true
	return nil
}

func (m *MockFileSystem) RemoveAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RemoveAllCalls = append(m.RemoveAllCalls, path)
	delete(m.Dirs, path)
	delete(m.Files, path)
	return nil
}

func (m *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ReadDirCalls = append(m.ReadDirCalls, path)
	return nil, nil
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StatCalls = append(m.StatCalls, path)

	if m.ShouldStatFail {
		return nil, os.ErrNotExist
	}
	if m.Dirs[path] {
		return &mockFileInfo{name: filepath.Base(path), isDir: true}, nil
	}
	if data, ok := m.Files[path]; ok {
		return &mockFileInfo{name: filepath.Base(path), isDir: false, size: int64(len(data))}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Exists(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.Dirs[path] || m.Files[path] != nil
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ReadFileCalls = append(m.ReadFileCalls, path)

	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.WriteFileCalls = append(m.WriteFileCalls, WriteFileCall{Path: path, Data: data, Perm: perm})
	m.Files[path] = data
	return nil
}

func (m *MockFileSystem) OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OpenFileCalls = append(m.OpenFileCalls, OpenFileCall{Path: path, Flag: flag, Perm: perm})

	// Handle os.O_CREATE flag
	if flag&os.O_CREATE != 0 {
		if _, exists := m.Files[path]; !exists {
			m.Files[path] = []byte{}
		}
	}

	return nil, nil
}

func (m *MockFileSystem) WalkDir(root string, walkFn func(path string, d ports.DirEntry, err error) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.WalkDirCalls = append(m.WalkDirCalls, root)
	return nil
}

func (m *MockFileSystem) Chmod(path string, mode os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ChmodCalls = append(m.ChmodCalls, ChmodCall{Path: path, Mode: mode})
	return nil
}

func (m *MockFileSystem) UserHomeDir() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UserHomeDirCalled = true
	return "/home/test", nil
}

func (m *MockFileSystem) CreateFile(path string, perm os.FileMode) (*os.File, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateFileCalls = append(m.CreateFileCalls, CreateFileCall{Path: path, Perm: perm})
	// Simula creación de archivo en memoria
	if _, exists := m.Files[path]; !exists {
		m.Files[path] = []byte{}
	}
	return nil, nil
}

// mockFileInfo implements os.FileInfo for testing.
type mockFileInfo struct {
	name  string
	isDir bool
	size  int64
}

func (m *mockFileInfo) Name() string { return m.name }
func (m *mockFileInfo) Size() int64  { return m.size }
func (m *mockFileInfo) Mode() os.FileMode {
	if m.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// MockDirEntry implements ports.DirEntry for testing.
type MockDirEntry struct {
	isDir bool
	info  *mockFileInfo
}

func (m *MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m *MockDirEntry) Info() (os.FileInfo, error) {
	if m.info != nil {
		return m.info, nil
	}
	return &mockFileInfo{name: "mock", isDir: m.isDir}, nil
}
