// Package config provides filesystem path construction utilities.
// These are filesystem conventions, not domain concepts.
package config

import "path/filepath"

// BuildWorkspaceDir returns the path to a bunker's workspace directory.
func BuildWorkspaceDir(baseDir, containerName string) string {
	return filepath.Join(baseDir, ".entorno", containerName)
}

// AIConfigDir returns the path to the AI configuration directory.
func AIConfigDir(baseDir string) string {
	return filepath.Join(baseDir, "ai_config")
}

// TutorPath returns the path to the tutor.md file.
func TutorPath(baseDir string) string {
	return filepath.Join(AIConfigDir(baseDir), "teams", "tutor.md")
}
