package runtime

// CommandSet define los comandos para un runtime de contenedor
type CommandSet struct {
	// Container lifecycle
	CreateBunker func(name, image, home, flags string) []string
	StartBunker  func(name string) []string
	StopBunker   func(name string) []string
	RemoveBunker func(name string, force bool) []string
	ListBunkers  func() []string
	BunkerExists func(name string) []string

	// Images
	ImageExists    func(image string) []string
	RemoveImage    func(image string, force bool) []string
	CommitImage    func(containerName, imageName, author, message string) []string
	ContainerState func(name string) []string
	StartContainer func(name string) []string

	// Execution
	EnterBunker     func(name, rcPath string) []string
	ExecuteInBunker func(name string, args ...string) []string
}

// Podman es el CommandSet para Podman/Distrobox
var Podman = CommandSet{
	CreateBunker: func(name, image, home, flags string) []string {
		return []string{
			"distrobox-create",
			"--name", name,
			"--image", image,
			"--home", home,
			"--additional-flags", flags,
			"--yes",
		}
	},
	StartBunker: func(name string) []string {
		return []string{"podman", "start", name}
	},
	StopBunker: func(name string) []string {
		return []string{"distrobox-stop", name, "--yes"}
	},
	RemoveBunker: func(name string, force bool) []string {
		if force {
			return []string{"distrobox-rm", name, "--force", "--yes"}
		}
		return []string{"distrobox-rm", name, "--yes"}
	},
	ListBunkers: func() []string {
		return []string{"podman", "ps", "-a", "--format", "json"}
	},
	BunkerExists: func(name string) []string {
		return []string{"podman", "ps", "-a", "--format", "json"}
	},
	ImageExists: func(image string) []string {
		return []string{"podman", "image", "exists", image}
	},
	RemoveImage: func(image string, force bool) []string {
		if force {
			return []string{"podman", "rmi", image, "--force"}
		}
		return []string{"podman", "rmi", image}
	},
	CommitImage: func(containerName, imageName, author, message string) []string {
		return []string{"podman", "commit",
			"--pause=false",
			"-f", "docker",
			"--change", "CMD=/bin/bash",
			"--change", "ENTRYPOINT=",
			"-a", author,
			"-m", message,
			containerName,
			imageName,
		}
	},
	ContainerState: func(name string) []string {
		return []string{"podman", "ps", "-a",
			"--filter", "name=" + name,
			"--format", "{{.State}}"}
	},
	StartContainer: func(name string) []string {
		return []string{"podman", "start", name}
	},
	EnterBunker: func(name, rcPath string) []string {
		if rcPath == "" {
			rcPath = "/dev/null"
		}
		return []string{"distrobox-enter", name, "--", "bash", "--rcfile", rcPath, "-i"}
	},
	ExecuteInBunker: func(name string, args ...string) []string {
		result := []string{"distrobox-enter", "-n", name, "--"}
		return append(result, args...)
	},
}
