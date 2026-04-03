// Package bunker contiene la lógica de negocio para búnkeres (contenedores de desarrollo).
// Sigue el patrón de Spec-Driven Development (SDD) con inyección de dependencias.
package bunker

import (
	"context"
	"sync"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// Manager orquesta todas las operaciones de búnkeres.
// Recibe las dependencias (adapters) en su constructor, permitiendo testing y flexibilidad.
type Manager struct {
	rootDir string
	runtime ports.IBunkerRuntime
	fs      ports.IFileSystem
	ui      ports.IPresenter
	system  ports.ISystem
	git     ports.IGit
	mu      sync.Mutex // Protege operaciones que pueden tener race conditions
}

// NewManager crea una nueva instancia del Manager con sus dependencias.
func NewManager(rootDir string, runtime ports.IBunkerRuntime, fs ports.IFileSystem, ui ports.IPresenter, system ports.ISystem, git ports.IGit) *Manager {
	return &Manager{
		rootDir: rootDir,
		runtime: runtime,
		fs:      fs,
		ui:      ui,
		system:  system,
		git:     git,
	}
}

// LoadConfig carga la configuración desde el archivo config.toml
func (m *Manager) LoadConfig() (config.EnvConfig, error) {
	return LoadConfig(m.fs, m.rootDir)
}

// CreateBunker crea o reusa un búnker y entra directamente dentro de él.
func (m *Manager) CreateBunker(ctx context.Context, name string) error {
	return m.create(ctx, name)
}

// DeleteBunker elimina un búnker y permite decidir si también se borra el código del proyecto.
func (m *Manager) DeleteBunker(ctx context.Context, name string, force, deleteImage bool) error {
	return m.delete(ctx, name, force, deleteImage)
}

// DeleteBunkerImage elimina la imagen base correspondiente a la GPU configurada/detectada.
func (m *Manager) DeleteBunkerImage(ctx context.Context) error {
	return m.deleteImage(ctx)
}

// ListBunkers muestra el estado de los búnkeres detectados en el sistema.
func (m *Manager) ListBunkers(ctx context.Context) error {
	return m.list(ctx)
}

// BunkerInfo muestra información detallada de un búnker específico.
func (m *Manager) BunkerInfo(ctx context.Context, name string) error {
	return m.info(ctx, name)
}

// StopBunker detiene un búnker activo sin borrar su entorno ni el proyecto.
func (m *Manager) StopBunker(ctx context.Context) error {
	return m.stop(ctx)
}

// PruneBunkers limpia búnkeres huérfanos (directorios sin contenedor activo).
// Este método es thread-safe gracias al mutex.
func (m *Manager) PruneBunkers(ctx context.Context) error {
	return m.prune(ctx)
}

// BunkerStatus retorna el estado actual de un búnker.
func (m *Manager) BunkerStatus(ctx context.Context, name string) string {
	return m.bunkerStatus(ctx, name)
}

// ImageExists verifica si una imagen existe en el runtime.
func (m *Manager) ImageExists(ctx context.Context, image string) bool {
	exists, _ := m.runtime.ImageExists(ctx, image)
	return exists
}

// ListAxiomImages lista las imágenes de AXIOM disponibles.
func (m *Manager) ListAxiomImages(ctx context.Context) ([]string, error) {
	return m.listAxiomImages(ctx)
}

// Help muestra la ayuda de comandos disponibles.
func (m *Manager) Help() error {
	m.ui.ShowLogo()
	m.ui.ShowHelp()
	return nil
}

// Create es un alias para CreateBunker sin contexto.
func (m *Manager) Create(name string) error {
	return m.CreateBunker(context.Background(), name)
}

// CreateWithImage creates a bunker with a specific image.
func (m *Manager) CreateWithImage(name, image string) error {
	return m.createWithImage(context.Background(), name, image)
}

// Delete es un alias para DeleteBunker sin contexto.
func (m *Manager) Delete(name string) error {
	return m.DeleteBunker(context.Background(), name, false, false)
}

// List es un alias para ListBunkers sin contexto.
func (m *Manager) List() error {
	return m.ListBunkers(context.Background())
}

// Stop es un alias para StopBunker sin contexto.
func (m *Manager) Stop() error {
	return m.StopBunker(context.Background())
}

// Prune es un alias para PruneBunkers sin contexto.
func (m *Manager) Prune() error {
	return m.PruneBunkers(context.Background())
}

// Info es un alias para BunkerInfo sin contexto.
func (m *Manager) Info(name string) error {
	return m.BunkerInfo(context.Background(), name)
}

// DeleteImage es un alias para DeleteBunkerImage sin contexto.
func (m *Manager) DeleteImage() error {
	return m.DeleteBunkerImage(context.Background())
}

// UI exposes the presenter for router access
func (m *Manager) GetUI() ports.IPresenter {
	return m.ui
}

// ConfigureGit configures Git user.name and user.email for a project directory.
func (m *Manager) ConfigureGit(ctx context.Context, cfg EnvConfig, projectDir string) error {
	if cfg.GitUser == "" || cfg.GitEmail == "" {
		return nil
	}
	return m.git.ConfigureUser(ctx, projectDir, cfg.GitUser, cfg.GitEmail)
}
