#!/bin/bash
# ─── MÓDULO BUNKER: GENERACIÓN DE CONFIGURACIONES (TEMPLATES) ──────────

_init_tutor() {
    if [ ! -f "$TUTOR_PATH" ]; then
        mkdir -p "$(dirname "$TUTOR_PATH")"
        touch "$TUTOR_PATH"
    fi
}

_escribir_bashrc() {
    local NOMBRE="$1" R_ENTORNO="$2"

    # ── Variables de entorno (expansión deliberada) ──
    # Card 1: AXIOM_GIT_TOKEN eliminado / removed — nunca se exporta al búnker / never exported to the bunker.
    # El token se lee on-demand / The token is read on-demand desde /run/axiom/env (montado read-only en create).
    cat > "$R_ENTORNO/.bashrc" << BASH_VARS
    export AXIOM_GIT_USER="$AXIOM_GIT_USER"
    export AXIOM_GIT_EMAIL="$AXIOM_GIT_EMAIL"
    export AXIOM_AUTH_MODE="${AXIOM_AUTH_MODE:-https}"
    export SSH_AUTH_SOCK="${SSH_AUTH_SOCK:-}"
    export OLLAMA_MODELS="/ai_config/models"
BASH_VARS

    if [[ -n "${GFX_VAL:-$AXIOM_GFX_VAL}" ]]; then
        echo "export HSA_OVERRIDE_GFX_VERSION=${GFX_VAL:-$AXIOM_GFX_VAL}" >> "$R_ENTORNO/.bashrc"
    fi

    # ── Shell setup (sin expansión) ──
    cat >> "$R_ENTORNO/.bashrc" << 'BASH_RC'
    source $AXIOM_PATH/lib/core.sh
    source $AXIOM_PATH/lib/git.sh
    eval "$(starship init bash)"
BASH_RC

    # ── Ir al directorio del proyecto ──
    echo "    cd /$NOMBRE" >> "$R_ENTORNO/.bashrc"

    # ── Arranque de gentle-ai (una sola vez) y sync ──
    cat >> "$R_ENTORNO/.bashrc" << 'BASH_RC'
    Archive="$HOME/.axiom_done"
    if [ ! -f "$Archive" ]; then
        if command -v gentle-ai &>/dev/null; then
            gentle-ai
        else
            echo "⚠️  gentle-ai no encontrado, omitiendo arranque inicial."
        fi
        echo "done" > "$Archive"
    fi
    sync-agents
BASH_RC
}

_escribir_opencode_config() {
    local NOMBRE="$1" R_ENTORNO="$2"

    mkdir -p "$R_ENTORNO/.config/opencode"
    local CONF="$R_ENTORNO/.config/opencode/opencode.json"

    if [ ! -f "$CONF" ]; then
        cat > "$CONF" << 'OPENCODE_CONFIG'
{
  "$schema": "https://opencode.ai/config.json",
  "agent": {
    "gentleman": {
      "description": "Senior Architect mentor - helpful first, challenging when it matters",
      "mode": "primary",
      "prompt": "{file:./AGENTS.md}",
      "tools": {
        "edit": true,
        "write": true
      }
    },
    "sdd-orchestrator": {
      "description": "Gentleman personality + SDD delegate-only orchestrator",
      "mode": "all",
      "prompt": "{file:./AGENTS.md}",
      "tools": {
        "bash": true,
        "edit": true,
        "read": true,
        "write": true
      }
    },
    "sdd-apply": {
      "description": "SDD delegate-only apply sub-agent",
      "mode": "all",
      "prompt": "{file:./AGENTS.md}",
      "tools": {
        "bash": true,
        "edit": true,
        "read": true,
        "write": true
      }
    }
  },
  "mcp": {
    "context7": {
      "enabled": true,
      "type": "remote",
      "url": "https://mcp.context7.com/mcp"
    },
    "engram": {
      "command": ["engram", "mcp"],
      "enabled": true,
      "type": "local"
    }
  },
  "permission": {
    "bash": {
      "*": "allow",
      "git commit *": "ask",
      "git push": "ask",
      "git push *": "ask",
      "git push --force *": "ask",
      "git rebase *": "ask",
      "git reset --hard *": "ask"
    },
    "read": {
      "*": "allow",
      "**/.env": "deny",
      "**/.env.*": "deny",
      "**/credentials.json": "deny",
      "**/secrets/**": "deny",
      "*.env": "deny",
      "*.env.*": "deny"
    }
  },
  "provider": {
    "ollama": {
      "npm": "@ai-sdk/openai-compatible",
      "options": {
        "baseURL": "http://localhost:11434/v1"
      },
      "models": {
        "TU_MODELO:latest": {
          "reasoning": true
        }
      }
    }
  }
}
OPENCODE_CONFIG
    else
        jq '.agent = (.agent // {})
          | .agent["sdd-orchestrator"] = {
              "description": "Gentleman personality + SDD delegate-only orchestrator",
              "mode": "all",
              "prompt": "{file:./AGENTS.md}",
              "tools": { "bash": true, "edit": true, "read": true, "write": true }
            }
          | .agent["sdd-apply"] = {
              "description": "SDD delegate-only apply sub-agent",
              "mode": "all",
              "prompt": "{file:./AGENTS.md}",
              "tools": { "bash": true, "edit": true, "read": true, "write": true }
            }' "$CONF" > "${CONF}.tmp" && mv "${CONF}.tmp" "$CONF"
    fi
}

_escribir_starship() {
local R_ENTORNO="$1"

    mkdir -p "$R_ENTORNO/.config"

    cat > "$R_ENTORNO/.config/starship.toml" << 'STARSHIP'

    add_newline = true



    format = """

    [](fg:#88c0d0)$os[](fg:#88c0d0 bg:#81a1c1)$username[](fg:#81a1c1 bg:#4c566a)$directory[](fg:#4c566a bg:#a3be8c)$git_branch$git_status[](fg:#a3be8c bg:#5e81ac)$time[ ](fg:#5e81ac)

    $character"""



    [os]

    disabled = false

    style = "bg:#88c0d0 fg:#2e3440"



    [os.symbols]

    Fedora = " "

    Arch = "󰣇 "



    [username]

    show_always = true

    style_user = "bg:#81a1c1 fg:#eceff4"

    format = "[ $user ]($style)"



    [directory]

    style = "bg:#4c566a fg:#eceff4"

    format = "[ $path ]($style)"

    home_symbol = "~"

    truncation_length = 3

    fish_style_pwd_dir_length = 1



    [directory.substitutions]

    "/var/home/alejandro" = "~"



    [git_branch]

    symbol = " "

    style = "bg:#a3be8c fg:#2e3440"

    format = "[[ $symbol$branch ]($style)]($style)"



    [git_status]

    style = "bg:#a3be8c fg:#2e3440"

    format = "[[($all_status$ahead_behind )]($style)]($style)"



    [time]

    disabled = false

    time_format = "%R"

    style = "bg:#5e81ac fg:#eceff4"

    format = "[  $time ]($style)"



    [character]

    success_symbol = "[╰─>](bold #a3be8c) "

    error_symbol = "[╰─>](bold #bf616a) "

STARSHIP
}