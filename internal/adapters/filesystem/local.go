// Package filesystem implementa el adapter para operaciones del sistema de archivos.
// Implementa la interfaz IFileSystem definida en internal/ports.
package filesystem

import (
	"os"
	"path/filepath"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// FSAdapter implementa IFileSystem usando operaciones estándar de Go.
type FSAdapter struct{}

// NewFSAdapter crea una nueva instancia del adapter.
func NewFSAdapter() *FSAdapter {
	return &FSAdapter{}
}

// MkdirAll crea todos los directorios en la ruta especificada.
func (f *FSAdapter) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// RemoveAll elimina directorios y su contenido.
func (f *FSAdapter) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// ReadDir lista las entradas de un directorio.
func (f *FSAdapter) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

// Stat retorna información del archivo o directorio.
func (f *FSAdapter) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// Exists verifica si existe un path.
func (f *FSAdapter) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile lee todo el contenido de un archivo.
func (f *FSAdapter) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile escribe datos a un archivo.
func (f *FSAdapter) WriteFile(path string, data []byte, perm os.FileMode) error {
	// Asegurar que el directorio padre exista
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

// OpenFile abre un archivo con las banderas especificadas.
func (f *FSAdapter) OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

// WalkDir recorre el árbol de directorios.
func (f *FSAdapter) WalkDir(root string, walkFn func(path string, d os.DirEntry, err error) error) error {
	return filepath.WalkDir(root, walkFn)
}

// Chmod cambia los permisos de un archivo.
func (f *FSAdapter) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// UserHomeDir retorna el directorio home del usuario.
func (f *FSAdapter) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// EnsureDirExists crea el directorio si no existe.
func (f *FSAdapter) EnsureDirExists(path string, perm os.FileMode) error {
	if !f.Exists(path) {
		return f.MkdirAll(path, perm)
	}
	return nil
}

// EnsureFileExists crea el archivo vacío si no existe.
func (f *FSAdapter) EnsureFileExists(path string, perm os.FileMode) error {
	if f.Exists(path) {
		return nil
	}
	dir := filepath.Dir(path)
	if err := f.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE, perm)
	if err != nil {
		return err
	}
	return file.Close()
}

// CreateFile creates a file with the given permissions, truncating if it exists.
func (f *FSAdapter) CreateFile(path string, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
}

// VerifyInterface verifica que el adapter implementa la interfaz correctamente.
var _ ports.IFileSystem = (*FSAdapter)(nil)
