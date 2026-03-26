# 🤝 Contribute to AXIOM

Hi! I'm Alejandro, the sole developer of AXIOM. I'm looking for allies to help me retire the old Bash code and build a robust orchestrator in Go.

## 📜 The Golden Rule: Native Go
We don't want to rely on Bash scripts. The goal is for **AXIOM** to be a pure Go binary that manages your bunkers safely and efficiently. 

## 🚧 Current Situation
I have already ported the **lifecycle** functions to Go. However, as an MVP, the code currently has:
1. **Security flaws** in system interactions.
2. **Logic bugs** that need polishing.

## 🤝 How to Help
If you know Go and System Security:
* Help me audit and fix the functions in `pkg/bunker/lifecycle.go`.
* Collaborate on the upcoming migration from `.env` to **TOML**.
* Help ensure the selective cleanup of `~/.entorno/` is indestructible.

Let's make immutable development clean and safe!
