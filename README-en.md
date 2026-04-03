[🇬🇧 English](README-en.md) | [🇪🇸 Español](README.md)

# AXIOM Bunker System 🛡️

```text
  █████╗ ██╗  ██╗██╗ ██████╗ ███╗   ███╗
 ██╔══██╗╚██╗██╔╝██║██╔═══██╗████╗ ████║
 ███████║ ╚███╔╝ ██║██║   ██║██╔████╔██║
 ██╔══██║ ██╔██╗ ██║██║   ██║██║╚██╔╝██║
 ██║  ██║██╔╝ ██╗██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝
```

> *Zero dirt. Thirty seconds. Ready to go.*

**AXIOM** is an isolated, modular development system built on **Distrobox** and **Podman**. Every bunker is an independent Arch Linux container with direct GPU access, a fully preconfigured local AI stack, and a customized Starship prompt — without touching a single critical file on your host.

Ideal for atomic OS setups (Bazzite, Fedora Silverblue) or any environment where keeping the host completely clean is a priority.

---

## 🔒 Security (AXIOM Vault)

AXIOM implements strict measures to protect your credentials and host:
- **Tokens in Read-Only Volumes:** Your GitHub token is never exported as an environment variable. It is read *on-demand* from a read-only mounted volume (`/run/axiom/env`), preventing malicious processes or extensions from capturing it via `printenv`.
- **TOCTOU Prevention:** `mktemp` is used for critical operations to block race condition attacks.
- **Injection Mitigation:** Commands interact with sensitive variables via bash arrays rather than evaluating flat strings.

---

## 🚀 Quick Installation

**Requirements:**
* Distrobox ≥ 1.7
* Podman
* fzf (For interactive menus)
* Compatible Host (Bazzite, Fedora Silverblue, Nobara, CachyOS, etc.)

1. **Clone the repo:**
```bash
git clone [https://github.com/Alejandro-M-P/AXIOM.git](https://github.com/Alejandro-M-P/AXIOM.git) ~/AXIOM
cd ~/AXIOM
```

2. **Run the installer:**
```bash
chmod +x install.sh && ./install.sh
```

3. **Configure Shell and Build:**
```bash
echo "source ~/AXIOM/axiom.sh" >> ~/.bashrc
source ~/.bashrc

axiom build
```
*(Takes 15-30 mins depending on the chosen GPU driver mode).*

---

## 💻 Basic Usage: Host Commands

Run `axiom create my-project` and in 30 seconds you have a fully equipped environment. 

| Command | Description |
| :--- | :--- |
| `axiom help` | Shows the commands currently available in the Go orchestrator. |
| `axiom build` | Builds the base image with GPU and AI tools. |
| `axiom list` | Shows detected bunkers with status, size, last entry, and git branch. |
| `axiom create <name>` | Creates a new bunker from the base image or enters an existing one. |
| `axiom delete [name]` | Deletes a bunker. If no name is provided, it opens an arrow-key selector. |
| `axiom delete-image` | Deletes the active base image and shows detected AXIOM images. |
| `axiom stop` | Stops the execution of an active bunker. |
| `axiom info [name]` | Shows the detailed summary card of a bunker. |
| `axiom prune` | Cleans up orphan environments with no container. |
| `axiom enter <name>` | ✅ Enters an existing bunker interactively. |
| `axiom rebuild` | ✅ Rebuilds the base image. |
| `axiom reset` | ✅ Deletes ALL bunkers and images (Total reset). |

### Current Go Migration Layout
AXIOM is currently undergoing a massive refactor to replace Bash scripts (`lib/*.sh`) with a robust Go binary. The current code layout lives in `pkg/` but is actively evolving to decouple logic from presentation:

| Go Port Progress | Status | Architectural Notes |
| :--- | :--- |
| `axiom list` | ✅ Ported | Needs to decode Podman/Distrobox JSONs instead of relying on string parsing. |
| `axiom info` | ✅ Ported | Needs path sanitization (`filepath.Clean`) to mitigate *Path Traversal* risks. |
| `axiom build` | 🚧 WIP | GPU injection in progress. Sudo prompts occasionally hang the execution context. |
| `axiom create` | 🚧 WIP | Adjusting `$HOME` persistence and enforcing strictly secure `0700` permissions. |
| `axiom delete` | 🚧 WIP | Working, but needs decoupling from the `stdin` UI blocks in the core Manager. |
| `axiom prune` | ⏳ Pending | Needs to be rewritten leveraging Goroutines for concurrent background deletion. |
| `axiom enter` | ✅ | Interactive TTY entry into existing bunker via distrobox-enter. |
| `axiom rebuild` | ✅ | Reuses build plan but deletes the image first. |
| `axiom reset` | ✅ | Total cleanup: all bunkers, all images, config reset. |
| **Git Tools** | ⏳ Pending | The interactive bash commands (`lib/git.sh`) will be replaced using `go-git`. |

### 🗺️ Technical Vision & Roadmap
The main goal of the Go refactor is to establish a secure, immutable, and scalable orchestrator:
1. **Native over Shell:** Replacing OS tools (`du`, `grep`) with pure Go libraries (`filepath.WalkDir`, `go-git`).
2. **UI Decoupling:** Moving all console prints, ANSI codes, and `stdin` requests out of the core logic and into a pure presentation layer (BubbleTea).
3. **Podman REST API:** Moving away from executing `podman` as a shell subprocess and talking directly to the Podman API socket.
4. **TOML Configs:** Retiring `.env` for a structured `config.toml` standard.
5. **Safe Concurrency:** Using `sync.WaitGroup` for multi-bunker operations and `context` to prevent zombie processes.

### Security Hotfixes under Audit
- **[ID-BUG-001] Path Traversal:** Strengthening user inputs when concatenating bunker paths.
- **[ID-BUG-005] Loose Permissions:** Enforcing `0700` across all OS directories created by the `Manager`.
- **[ID-BUG-011] Error Silencing:** Restoring error wrapping across external commands to trace execution failures properly.

Check out `BUGS.md` and `CONTRIBUTING.md` if you want to help out!

---

## 🛡️ Internal Tools (Inside the Bunker)

### AI & Agent System
| Command | Description |
| :--- | :--- |
| `open` | Starts and opens `opencode`. |
| `sync-agents` | Copies `tutor.md` to local agent config (`AGENTS.md`). |
| `save-rule <rule>`| Saves a rule in `tutor.md` forcing a technical justification. |
| `diagnostics` | Runs an internal bunker diagnostic. |

### Interactive Git (Powered by fzf)
Inside the bunker, you have custom commands that override Git to streamline visual workflows:

| Command | Description |
| :--- | :--- |
| `status` | Interactive status with real-time visual *diff*. |
| `clone [u/r]` | Clones a repository specifying `user/repo`. |
| `commit [msg]` | Select files with `<Tab>` before committing. |
| `branch` | Interactive branch creation. |
| `switch` | Visual branch switching. |
| `branch-delete`| Secure visual deletion of local and/or remote branches. |
| `push` / `pull`| Sync with interactive remote and branch selection. |
| `merge` / `rebase`| Interactive origin and integration strategy selectors. |
| `log` | Colored history with commit code previews. |
| `stash` | Interactive management (save, apply, drop, view). |
| `remote` | Visual remote management (add, view, delete). |
| `tag` | Visual tag creation and management. |

---

## 🧠 Included AI Stack

Everything runs locally. No data leaves your machine. 

| Tool | Function |
| :--- | :--- |
| `opencode` | Code editor with integrated AI. |
| `engram` | Persistent memory across sessions. |
| `gentle-ai` | AI agent interface. |
| `agent-teams-lite` | Multi-agent coordination (Orchestrator, Apply). |
| `ollama` | Local LLMs running on your GPU. |

---

## 📜 tutor.md — The Law of Your Agents

So the AI doesn't start from scratch every session, it needs context. `tutor.md` is the rule file agents must read on startup.

It lives outside the bunker (`~/dev/ai_config/teams/tutor.md`). If you delete a bunker, your rules don't disappear. The next bunker inherits them.

### Recommended tutor.md Template
Copy this into your `tutor.md` to enforce strict, professional AI behavior:

```markdown
# 🤖 ROLE: EXECUTION COPILOT (Junior Coder / Senior Mind)

## 👤 Identity
You are the developer's execution arm. Your mission is to generate clean, functional, and professional code at maximum speed, filtered through a Senior Architect's judgment.

## 🛡️ Action Protocol (Skeptic-to-Code)
1. **Skeptic First**: Before coding, ask "why". If the idea is bad, warn the user. Be a critical partner, not a submissive robot.
2. **Explain & Validate**: For complex tasks, briefly explain the design and wait for the "OK".
3. **High-Speed Execution**: Deliver complete, testable blocks. No useless snippets.
4. **No Assumptions**: If information is missing, ask for it. Better to ask once than fix ten times.

## 🏛️ Quality Standards
- **Clean Code & Pro Naming**: Code must speak for itself.
- **Error Detection**: When delivering code, point out the 2 most likely points of failure.
```

---

## 🔬 Deep Dive: Internal Architecture & Technical Decisions

For developers who need to understand what's happening under the hood:

### 1. Why is the GPU image 12GB vs 38GB?
During `install.sh`, you are asked for a **GPU Driver Mode**:
* **`host` Mode (~12GB):** AXIOM does not install heavy graphics SDKs inside the bunker. Instead, it uses bind mounts to inject `/usr/lib/rocm` or `/usr/local/cuda` directly from your host OS into the container. Most efficient, but requires the host to have drivers.
* **`image` Mode (~38GB):** AXIOM downloads and installs full ROCm/CUDA packages *inside* the Arch Linux image. This balloons the size but makes the bunker 100% host-independent.

### 2. The `build` -> `commit` -> `create` Lifecycle
AXIOM does not install dependencies every time you create a project.
* `axiom build` creates a temporary container, installs everything (Go, Node, Ollama, Opencode), and executes a `podman commit`. This "freezes" the state into `localhost/axiom-[gpu]:latest`.
* `axiom create my-project` simply tells Distrobox to clone that frozen image in 30 seconds, injecting an isolated `--home` at `~/.entorno/my-project`.

### 3. Physical Isolation
* **The Code:** Lives at `~/dev/my-project` (on the host) and is mounted at `/my-project` in the bunker.
* **The System (Home):** Lives at `~/dev/.entorno/my-project`. Contains config, bashes, caches, etc.
* Running `axiom delete` destroys the Podman container and the `.entorno/` directory. **Your project code is only removed if you explicitly confirm that option during deletion.**

---

## 🛠️ FAQ & Troubleshooting

**Can I use AXIOM without a GPU?**
Yes. During `build`, select `Generic / CPU Only`. Ollama will run on the CPU.

**opencode isn't connecting to Ollama**
Ensure Ollama is running inside: `ollama list`. If unresponsive, check `/tmp/ollama.log`.

**`rocminfo: command not found` inside the bunker**
In `host` mode, ROCm is mounted from the host. If your host lacks it, switch to `image` mode in `.env` and `rebuild`.

**`podman commit` takes too long**
Normal for large images (e.g., 38GB). It can take 15 mins. Verify it's active with `podman ps` in another terminal.

---

## 🤝 Contributing
Fork, create a descriptive branch, commit clearly, and open a PR explaining why. High-value contributions: Support for unlisted distros/GPUs and build optimizations.

---

## 📖 History & Philosophy

Bazzite is an atomic OS; the host is immutable. Your home is not a testing ground. 

To code seriously without breaking the OS, Distrobox was the initial answer. What started as a tiny 10-line script evolved into a strict organizational system in `~/dev`. Integrating local AI (via the *Gentleman Programming* ecosystem) turned this container into a high-performance bunker. 

The goal never changed: **zero dirt on your machine. Everything else just followed.**
