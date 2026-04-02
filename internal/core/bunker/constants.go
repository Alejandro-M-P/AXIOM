package bunker

import "time"

const (
	DefaultBuildContainerName = "axiom-build"
	BunkerReadyTimeout        = 30 * time.Second
	BunkerStartSleep          = 2 * time.Second
	DirPermission             = 0700
	FilePermission            = 0644
	DateFormat                = "2006-01-02"
)
