[🇬🇧 English](README-en.md) | [🇪🇸 Español](README.md)

# AXIOM Bunker System

> *Zero mess. Thirty seconds. Ready to go.*

Isolated and modular development system built on Distrobox + Podman. Each bunker is an independent Arch Linux container with direct GPU access, full local AI stack, and a customized Starship prompt — without touching the host.

---

## The Story

Bazzite is an atomic operating system. That means the host is immutable by design — you can't install packages directly onto it like you would in any other distro. And honestly, even if you could, you wouldn't want to. The host is your home. Your home is not a testing ground.

So when I wanted to start developing seriously, the question was inevitable: where do I install things without breaking anything? The answer was already installed by default in Bazzite: **Distrobox**. Containers that live inside your system, share the kernel, see the GPU, access your files — but are completely separated from the host.

That's where it all started. A small script. Ridiculously small. It just created a box and entered it. That was AXIOM in its first version.

Then came the idea of keeping the development environment in `~/dev` — separated from the system, organized, controlled. Each project in its own folder, each bunker with its own home. The host remained untouched. Zero mess.

And then I enrolled in a free course by Mouredev on learning to code with AI. There I discovered the Gentleman Programming stack: opencode, engram, gentle-ai, agent-teams-lite. A complete ecosystem of AI tools for development — all local, all on your hardware, nothing in the cloud. I wanted to integrate it. And instead of installing it bare-metal on the host or in just any container, I put it inside AXIOM.

One problem led to another, every solution made the system more solid, and here we are. What started as a 10-line script to avoid dirtying Bazzite is now a complete system of isolated development environments with GPU, local AI, and shared memory between agents.

**The goal never changed: zero mess. Everything else came naturally.**

---

## What exactly is it?

You run `axiom create my-project`. In 30 seconds you get a full Arch Linux with direct access to your GPU, Ollama running locally with your models, and the entire Gentleman Programming AI stack ready to use. Code however you want, experiment however you want, break whatever you want.

When you're done, the host is exactly as it was when you started. Zero mess.

If the bunker breaks, you delete it and create another one. 30 seconds. No drama.

---

## Requirements

- Distrobox >= 1.7
- Podman
- Compatible host (Bazzite, Fedora Silverblue, Nobara, CachyOS, any Podman-capable distro)

---

## Quick Install

1. **Clone the repository:**
```bash
git clone https://github.com/Alejandro-M-P/AXIOM.git ~/AXIOM
cd ~/AXIOM
```

2. **Run the installer:**
```bash
chmod +x install.sh && ./install.sh
```
The wizard will ask for your GitHub credentials, token, base directory, and **GPU drivers mode**:

| Mode | Description | Image Size | Commit Time |
| :--- | :--- | :--- | :--- |
| `host` *(recommended)* | Mounts host ROCm/CUDA. For Bazzite, Fedora, Nobara, CachyOS... | ~10-13 GB | ~3-4 min |
| `image` | Installs ROCm/CUDA inside. Works on any distro. | ~38 GB | ~15 min |

3. **Add source to your shell:**
```bash
echo "source ~/AXIOM/axiom.sh" >> ~/.bashrc
source ~/.bashrc
```

4. **Build the base image (only once):**
```bash
axiom build
```
Automatically detects your GPU and installs the entire stack into `localhost/axiom-[gpu]:latest`. It takes ~15-30 min to install plus ~3-15 min for `podman commit` depending on the chosen mode. Upon completion, it automatically cleans all caches before committing.

5. **Create your first bunker:**
```bash
create my-first-project
```
From here on, each new bunker boots in ~30 seconds.

---

## Included AI Stack

Everything runs locally. Nothing goes out to any server. Based on the Gentleman Programming ecosystem, discovered through Mouredev's AI programming course.

| Tool | What it does |
| :--- | :--- |
| `opencode` | Code editor with integrated AI |
| `engram` | Persistent memory between sessions |
| `gentle-ai` | AI agents interface |
| `agent-teams-lite` | Coordination of multiple agents |
| `ollama` | Local language models running on your GPU |

### Ollama configuration in opencode

AXIOM automatically writes the Ollama connection to `~/.config/opencode/config.json` inside each bunker. opencode comes with its default providers — this block simply adds local Ollama on top of what it already has:

```json
 } ,
  "provider": {
    "ollama": {
      "npm": "@ai-sdk/openai-compatible",
      "options": {
        "baseURL": "http://localhost:11434/v1"
      },
      "models": {
        "YOUR_MODEL:latest": {
          "reasoning": true
        }
      }
    }
  }
}
```

Replace `YOUR_MODEL` with any model you have downloaded in Ollama. Some typical examples:

```bash
ollama pull qwen2.5:latest        # good for code, lightweight
ollama pull qwen2.5-coder:latest  # specialized in coding
ollama pull deepseek-r1:latest    # strong reasoning
ollama pull llama3.1:latest       # general purpose
```

To see available models:
```bash
ollama list
```

---

## tutor.md — The Law of your Agents

When working with AI agents like opencode, the agent needs context. Without it, every session starts from scratch — it doesn't know how you want it to work, what conventions you use, what technical decisions you made, what mistakes it shouldn't repeat.

`tutor.md` solves that. It's a rules file that you define — or that you define alongside AI — and that agents are forced to read every time they start. It's not a suggestion. It gets directly injected into the agent configuration folder inside the bunker, in the exact spot opencode reads upon boot. The agent cannot ignore it.

```
~/dev/ai_global/teams/tutor.md
```

**The important part is where it lives.** `tutor.md` is not inside the bunker — it's in `ai_global`, which is a shared volume mounted across all bunkers. That means when you delete a bunker, the rules do not vanish. They are outside, secure, and the next bunker you create will automatically inherit them.

**How to add rules:**

You can write them directly to the file, or use `save-rule` from within the bunker:

```bash
save-rule "always use strict TypeScript"
```

It will ask for a technical reason — mandatory by design. It forces you to think before adding a rule:

```
📝 Technical reason: because runtime type errors cost more than typing out the code
```

The rule is saved to `tutor.md` along with its reason, and `sync-agents` runs automatically — it copies the updated file to the local configuration of all currently running bunkers.

**AI can do it too.** If during a work session the agent detects a pattern or decision it should remember, it can call `save-rule` itself. The rule remains saved for all future sessions, across all bunkers.

**`sync-agents`** is the function keeping everything synchronized. It copies `tutor.md` to `~/.config/opencode/AGENTS.md` inside every active bunker. It runs automatically upon creating a bunker, entering an existing one, and when a new rule is saved. You can also trigger it manually:

```bash
sync-agents
```

The result is that your agents always work with the same rules across all projects, regardless of how many bunkers you create or destroy. The memory of your workflow outlives them all.

---

## Available Commands

| Command | Description | Environment |
| :--- | :--- | :--- |
| `build` | Builds base image with GPU, AI tools, and starship. Runs only once per machine. | **Host** |
| `rebuild` | Rebuilds base image to update the stack. Existing bunkers are not affected. | **Host** |
| `resetear` | Deletes base image. Asks if you also want to delete all bunkers. | **Host** |
| `reset-base` | Deletes base image without touching existing bunkers. Useful for clearing space. | **Host** |
| `crear [name]` | Creates a new bunker from base image (~30s) or enters an existing one. | **Host** |
| `borrar [name]` | Demands a technical reason and completely destroys the bunker and its local memory. | **Host** |
| `parar [name]` | Stops the bunker container without deleting its data. | **Host** |
| `open` | Syncs laws and opens the intelligent `opencode` environment. | **Bunker** |
| `sync-agents` | Synchronizes `tutor.md` to local agent config. | **Bunker** |
| `save-rule [rule]`| Saves a new technical rule and synchronizes it across all active bunkers. | **Bunker** |
| `git-clone [u/r]` | Clones a GitHub repository securely using token credentials. | **Bunker** |
| `rama` | Interactively creates a new branch — prompts for name and base branch. | **Bunker** |
| `commit [msg]` | Stages all changes and commits. Asks for a message if none is provided. | **Bunker** |
| `push` | Securely pushes to GitHub using the token from `.env`. | **Bunker** |
| `diagnostico` | Health diagnostics: GPU, Ollama, and Git Token. | **Bunker** |
| `ayuda` | Shows the help menu on screen. | **Host / Bunker** |

---

## Folder Structure

After installing and creating your first bunker, your disk looks like this:

```
~/dev/                              ← AXIOM_BASE_DIR (configurable during install)
│
├── my-project/                     ← project folder, mounted as /my-project inside bunker
│
├── ai_global/                      ← shared between ALL bunkers
│   ├── teams/
│   │   └── tutor.md                ← global agent rules
│   └── models/
│
├── ai_config/                      ← shared AI configuration
│   └── models/                     ← Ollama models (one directory for all)
│
└── .entorno/                       ← bunker homes (separated from the project code)
    └── my-project/                 ← home of bunker my-project
        ├── .bashrc                 ← bunker variables, PATH, functions
        ├── .config/
        │   ├── starship.toml       ← custom Tokyo Night prompt
        │   └── opencode/
        │       ├── config.json     ← local Ollama connection
        │       └── AGENTS.md       ← synced copy of tutor.md
        └── ...                     ← rest of home, isolated from the host

~/AXIOM/                            ← AXIOM system itself
├── axiom.sh                        ← main script
├── install.sh                      ← installer
└── .env                            ← your credentials and config (git-ignored)
```

Most importantly: the **project** and the **environment** are separated. The project lives in `~/dev/my-project` and is mounted inside the bunker. The bunker's home lives in `.entorno/my-project`. If you delete the bunker, your project won't disappear — only the environment.

---

## FAQ

**Can I have multiple bunkers running at the same time?**
Yes, as many as you want. Each has its own isolated home, `.bashrc`, and configuration. They all share `ai_global` and `ai_config`, so Ollama models and `tutor.md` rules remain the same everywhere.

**What happens if I delete a bunker?**
Only the environment disappears — the container and its `.entorno/` home. Your code at `~/dev/my-project` remains untouched. So do `tutor.md` rules. You can recreate the bunker in 30 seconds and continue exactly where you left off.

**Are Ollama models shared between bunkers?**
Yes. `ai_config/models` mounts as the Ollama model directory in every bunker. You pull a model once, and it is available everywhere.

**Can I use AXIOM without a GPU?**
Yes. During `build` you can choose `Generic / CPU Only` and no GPU drivers will be installed. Ollama works on CPU, it's just slower.

**Can I change GPU mode after installing?**
Yes. Edit `AXIOM_ROCM_MODE` inside `.env` and run `rebuild`. Existing bunkers aren't affected — only new ones will use the updated mode.

**What happens if `build` fails halfway?**
Run this to clean up and retry:
```bash
distrobox-rm axiom-build --force
rm -rf ~/dev/.entorno/axiom-build
build
```

---

## Troubleshooting

**`opencode` / `engram` / `gentle-ai` not found inside bunker**
Binaries are installed to `/usr/local/bin` during build. If they are missing, run `rebuild`.

**`rocminfo: command not found` inside bunker**
In `host` mode, ROCm mounts from the host. Ensure it exists on your system. If not, switch to `image` mode in `.env` and `rebuild`.

**`paru` fails during build with permission errors**
Happens occasionally with `makepkg` under distrobox. Clean up with the commands from the FAQ and `build` again.

**`podman commit` takes too long or freezes**
Normal with heavy images. Don't interrupt it. It can take up to 15 min in `image` mode.

**`sync-agents` doesn't update existing bunkers**
It only syncs to bunkers currently running. If the bunker is stopped, enter it and run `sync-agents` manually.

---

## Contributing

Contributions are highly welcome. Fork, create a feature branch (`feat/...` or `fix/...`), commit, and open a PR. Explain clearly what changes and why.

---

## Credits and related projects

AXIOM wouldn't exist without these:
- Distrobox | Podman
- opencode | Ollama
- Gentleman Programming (Engram, gentle-ai, agent-teams-lite)