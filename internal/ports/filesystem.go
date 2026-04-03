package ports

import "os"

// IFileSystem define el contrato para operaciones del sistema de archivos.
// Las implementaciones pueden ser locales o remotas (S3, etc.).
type IFileSystem interface {
	// MkdirAll crea todos los directorios en la ruta especificada.
	MkdirAll(path string, perm os.FileMode) error

	// RemoveAll elimina directorios y su contenido.
	RemoveAll(path string) error

	// ReadDir lista las entradas de un directorio.
	ReadDir(path string) ([]os.DirEntry, error)

	// Stat retorna información del archivo o directorio.
	Stat(path string) (os.FileInfo, error)

	// Exists verifica si existe un path.
	Exists(path string) bool

	// ReadFile lee todo el contenido de un archivo.
	ReadFile(path string) ([]byte, error)

	// WriteFile escribe datos a un archivo.
	WriteFile(path string, data []byte, perm os.FileMode) error

	// OpenFile abre un archivo con las banderas especificadas.
	OpenFile(path string, flag int, perm os.FileMode) (*os.File, error)

	// WalkDir recorre el árbol de directorios.
	WalkDir(root string, walkFn func(path string, d os.DirEntry, err error) error) error

	// Chmod cambia los permisos de un archivo.
	Chmod(path string, mode os.FileMode) error

	// UserHomeDir retorna el directorio home del usuario.
	UserHomeDir() (string, error)

	// CreateFile crea un archivo vacío con los permisos dados (equivale a abrir con O_CREATE | O_WRONLY | O_TRUNC).
	CreateFile(path string, perm os.FileMode) (*os.File, error)
}
