AXIOM_BUILD_CONTAINER="axiom-build"

_init_tutor() {
    if [ ! -f "$TUTOR_PATH" ]; then
        mkdir -p "$(dirname "$TUTOR_PATH")"
        touch "$TUTOR_PATH"
    fi
}
_escribir_bashrc() {
    local NOMBRE="$1" R_ENTORNO="$2"
    cat > "$R_ENTORNO/.bashrc" << BASH_VARS
    export AXIOM_GIT_USER="$AXIOM_GIT_USER"
    export AXIOM_GIT_EMAIL="$AXIOM_GIT_EMAIL"
    export AXIOM_GIT_TOKEN="$AXIOM_GIT_TOKEN"
    export AXIOM_AUTH_MODE="${AXIOM_AUTH_MODE:-https}"
    export SSH_AUTH_SOCK="${SSH_AUTH_SOCK:-}"
BASH_VARS

    if [[ -n "${GFX_VAL:-AXIOM_GFX_VAL:-}" ]]; then
        echo "export HSA_OVERRIDE_GFX_VERSION=${GFX_VAL:-AXIOM_GFX_VAL:-}" >> "$R_ENTORNO/.bashrc"
    fi

    cat >> "$R_ENTORNO/.bashrc" << 'BASH_RC'
    source $AXIOM_PATH/lib/core.sh
    source $AXIOM_PATH/lib/git.sh
    eval "$(starship init bash)"
BASH_RC

    echo "cd /$NOMBRE" >> "$R_ENTORNO/.bashrc"

    cat >> "$R_ENTORNO/.bashrc" << 'BASH_RC'
    # Validar si gentle-ai esta instalado o no porque si no esta instalado opencode no va a funcionar
    Archive="$HOME/.axiom_done"

    if [ ! -f "$Archive" ]; then
    gentle-ai
    echo "done" > "$Archive"
    fi
BASH_RC
}


build() {
    mostrar_logo
    detect_gpu

    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo ""
    echo "🏗️  Construyendo imagen base / Building base image: $IMAGEN"
    echo "    Modo GPU / GPU Mode: $AXIOM_ROCM_MODE"
    if [ "$AXIOM_ROCM_MODE" = "host" ]; then
        echo "    ROCm se montará desde el host → imagen ~10-13 GB / ROCm will be mounted from host → image ~10-13 GB"
    else
        echo "    ROCm se instalará dentro → imagen ~38 GB / ROCm will be installed inside → image ~38 GB"
    fi
    echo ""

    mkdir -p "$AI_GLOBAL/models" "$AI_GLOBAL/teams" "$AI_CONFIG/models"
    sudo chown -R "$USER:$USER" "$AI_GLOBAL" "$AI_CONFIG"
    _init_tutor

    echo "🧹 Limpiando búnker de construcción anterior... / Cleaning previous build bunker..."
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force --yes 2>/dev/null || true

    if [ -d "$BASE_ENV/$AXIOM_BUILD_CONTAINER" ]; then
        echo "🔓 Desbloqueando caché de Go para eliminación... / Unlocking Go cache for deletion..."
        chmod -R +w "$BASE_ENV/$AXIOM_BUILD_CONTAINER" 2>/dev/null
        rm -rf "$BASE_ENV/$AXIOM_BUILD_CONTAINER"
    fi

    echo "📦 Creando contenedor de build... / Creating build container..."
    distrobox-create --name "$AXIOM_BUILD_CONTAINER" \
        --image archlinux:latest \
        --home "$BASE_ENV/$AXIOM_BUILD_CONTAINER" \
        --additional-flags "--volume $AI_GLOBAL:/ai_global \
        --volume $AI_CONFIG:/ai_config \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render" \
        --yes

    local BUILD_SCRIPT="$BASE_ENV/$AXIOM_BUILD_CONTAINER/axiom-build.sh"
    mkdir -p "$BASE_ENV/$AXIOM_BUILD_CONTAINER"

    local GPU_PKGS=""
    if [ "$AXIOM_ROCM_MODE" = "image" ]; then
        case "${GPU_TYPE}" in
            nvidia) GPU_PKGS="nvidia-utils cuda" ;;
            rdna*)  GPU_PKGS="rocm-hip-sdk" ;;
            intel)  GPU_PKGS="intel-compute-runtime onevpl-intel-gpu" ;;
        esac
    fi

    cat > "$BUILD_SCRIPT" << SCRIPT
    #!/bin/bash
    set -uo pipefail

    export PATH="\$HOME/.local/bin:\$HOME/go/bin:/usr/local/bin:\$PATH"
    GPU_PKGS="${GPU_PKGS:-}"

    echo "⚡ [1/4] Sistema base... / Base system..."
    sudo pacman -Sy --needed --noconfirm base-devel git curl jq wget nodejs npm go

    echo "⚡ [2/4] Starship + GPU..."
    curl -fsSL https://starship.rs/install.sh | sh -s -- --yes --bin-dir /usr/local/bin
    if [ -n "\$GPU_PKGS" ]; then
        echo "⚡ Instalando GPU / Installing GPU: \$GPU_PKGS"
        sudo pacman -S --noconfirm --needed \$GPU_PKGS
    fi

    echo "⚡ [3/4] Herramientas IA en paralelo... / Parallel AI tools..."
    curl -fsSL https://opencode.ai/install | OPENCODE_INSTALL=/usr/local bash &
    PID_OC=\$!

    go install github.com/Gentleman-Programming/engram/cmd/engram@latest &
    PID_EN=\$!

    (
        AUTH_HEADER=""
        [ -n "${AXIOM_GIT_TOKEN:-}" ] && AUTH_HEADER="-H \"Authorization: Bearer ${AXIOM_GIT_TOKEN:-}\""

        GA_LATEST=\$(eval curl -fsSL \$AUTH_HEADER https://api.github.com/repos/Gentleman-Programming/gentle-ai/releases/latest | grep -o '"tag_name": *"[^"]*"' | grep -o '[0-9][^"]*' || echo "latest")

        if [ -z "\$GA_LATEST" ] || [ "\$GA_LATEST" = "latest" ]; then
            GA_LATEST="0.1.0"
        fi

        GA_URL="https://github.com/Gentleman-Programming/gentle-ai/releases/download/v\${GA_LATEST}/gentle-ai_\${GA_LATEST}_linux_amd64.tar.gz"
        if curl -fsSL "\$GA_URL" -o /tmp/gentle-ai.tar.gz; then
            tar -xzf /tmp/gentle-ai.tar.gz -C /tmp/
            sudo mv /tmp/gentle-ai /usr/local/bin/gentle-ai
            sudo chmod +x /usr/local/bin/gentle-ai
            rm -f /tmp/gentle-ai.tar.gz
        fi
    ) &
    PID_GA=\$!

    curl -fsSL https://ollama.com/install.sh | sh &
    PID_OL=\$!

    wait \$PID_OC && echo "✅ opencode"
    wait \$PID_EN && echo "✅ engram"
    wait \$PID_GA && echo "✅ gentle-ai"
    wait \$PID_OL && echo "✅ ollama"

    echo "⚡ [4/4] agent-teams-lite..."
    ollama serve > /tmp/ollama-build.log 2>&1 &
    OLLAMA_PID=\$!

    until curl -s http://localhost:11434/ > /dev/null; do sleep 1; done

    git clone https://github.com/Gentleman-Programming/agent-teams-lite.git /tmp/agent-teams
    cd /tmp/agent-teams && ./scripts/setup.sh --all && echo "✅ agent-teams-lite"
    kill \$OLLAMA_PID 2>/dev/null || true

    echo "⚡ Copiando binarios a /usr/local/bin... / Copying binaries to /usr/local/bin..."
    [ -f "\$HOME/go/bin/engram" ] && sudo cp -f "\$HOME/go/bin/engram" /usr/local/bin/

    echo "🧹 Limpiando caché interna... / Cleaning internal cache..."
    sudo pacman -Scc --noconfirm
    chmod -R +w ~/.cache/go ~/.cache 2>/dev/null || true
    sudo rm -rf /tmp/* ~/.cache/go /var/cache/pacman/pkg 2>/dev/null || true

    echo "✅ Build completo dentro del contenedor. / Build complete inside the container."
    rm -- "\$0"
SCRIPT

    chmod +x "$BUILD_SCRIPT"
    distrobox-enter -n "$AXIOM_BUILD_CONTAINER" -- bash "$BUILD_SCRIPT"

    echo "📦 Exportando imagen $IMAGEN (esto puede tardar)... / Exporting image $IMAGEN (this may take a while)..."
    podman commit "$AXIOM_BUILD_CONTAINER" "$IMAGEN"

    echo "🧹 Limpieza final... / Final cleanup..."
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force --yes
    chmod -R +w "$BASE_ENV/$AXIOM_BUILD_CONTAINER" 2>/dev/null
    rm -rf "$BASE_ENV/$AXIOM_BUILD_CONTAINER"

    echo ""
    echo "✅ Imagen $IMAGEN lista. Ya puedes usar: axiom create [nombre] / Image $IMAGEN ready. You can now use: axiom create [name]"
}


create() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso/Usage: axiom create [nombre/name]"; return 1; fi
    local NOMBRE="$1"
    local R_PROYECTO="$BASE_DEV/$NOMBRE"
    local R_ENTORNO="$BASE_ENV/$NOMBRE"

    echo "🛡️ Acceso al Búnker '$NOMBRE': / Bunker Access: '$NOMBRE'"
    if ! sudo -v; then echo "❌ Acceso denegado. / Access denied."; return 1; fi

    if distrobox-list --no-color | grep -qw "$NOMBRE"; then
        if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
            ssh-add -l &>/dev/null || ssh-add ~/.ssh/id_ed25519 2>/dev/null
        fi
        sync-agents
        distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
        return 0
    fi

    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)

    if ! podman image exists "$IMAGEN"; then
        echo ""
        echo "⚠️  No se encontró la imagen base $IMAGEN / Base image $IMAGEN not found."
        echo "    Ejecuta / Run: axiom build"
        return 1
    fi

    echo "⚡ Creando búnker '$NOMBRE' desde $IMAGEN... / Creating bunker '$NOMBRE' from $IMAGEN..."
    mkdir -p "$R_PROYECTO" "$R_ENTORNO" "$AI_CONFIG/models"
    mkdir -p "$AI_GLOBAL/models" "$AI_GLOBAL/teams"
    sudo chown -R "$USER:$USER" "$AI_GLOBAL" "$AI_CONFIG"
    _init_tutor

    local GPU_VOLS=""
    [ "$AXIOM_ROCM_MODE" = "host" ] && GPU_VOLS=$(_gpu_volumes_host)

    local SSH_VOL=""
    if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ] && [ -n "${SSH_AUTH_SOCK:-}" ] && [ -S "$SSH_AUTH_SOCK" ]; then
        SSH_VOL="--volume $SSH_AUTH_SOCK:$SSH_AUTH_SOCK"
        echo "🔑 Socket SSH detectado y montado. / SSH socket detected and mounted."
    fi

    distrobox-create --name "$NOMBRE" \
        --image "$IMAGEN" \
        --home "$R_ENTORNO" \
        --additional-flags "--volume $R_PROYECTO:/$NOMBRE \
        --volume $AI_GLOBAL:/ai_global \
        --volume $AI_CONFIG:/ai_config \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render \
        $GPU_VOLS $SSH_VOL" \
        --yes

    distrobox-enter -n "$NOMBRE" -- bash -c "
        sudo pacman -Syu --noconfirm --needed 2>/dev/null | tail -3
        echo '✅ Sistema actualizado. / System updated.'
    "

    _escribir_bashrc "$NOMBRE" "$R_ENTORNO"
    _escribir_starship "$R_ENTORNO"

    mkdir -p "$R_ENTORNO/.config/opencode"
    [ -f "$TUTOR_PATH" ] && cp "$TUTOR_PATH" "$R_ENTORNO/.config/opencode/AGENTS.md"
    sync-agents

    _escribir_opencode_config "$NOMBRE" "$R_ENTORNO"

    if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
        ssh-add -l &>/dev/null || ssh-add ~/.ssh/id_ed25519 2>/dev/null
    fi
    distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
}




delete() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso/Usage: axiom delete [nombre/name]"; return 1; fi
    read -rp "📝 Razón técnica obligatoria / Mandatory technical reason: " REASON
    [ -z "$REASON" ] && echo "❌ Cancelado: se requiere justificación. / Canceled: justification required." && return 1
    echo "- Borrado búnker $1 (Razón: $REASON)" >> "$TUTOR_PATH"

    read -rp "❗ ¿Borrar búnker '$1'? / Delete bunker '$1'? (s/N/y/N): " CONFIRM
    if [[ "$CONFIRM" =~ ^[sSyY]$ ]]; then
        distrobox-rm "$1" --force --yes
        if [ -d "$BASE_ENV/$1" ]; then
            chmod -R +w "$BASE_ENV/$1"
            rm -rf "$BASE_ENV/$1"
        fi
        echo "🔥 Búnker '$1' eliminado. / Bunker '$1' deleted."
    fi
}


reset() {
    mostrar_logo
    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo "🗑️  Estado de la imagen base / Base image status: $IMAGEN"
    if podman image exists "$IMAGEN"; then
        echo "    Tamaño / Size: $(podman images --format '{{.Size}}' $IMAGEN)"
    else
        echo "    (La imagen no existe actualmente / The image currently does not exist)"
    fi
    echo ""

    echo "🔎 Escaneando búnkeres en el sistema... / Scanning bunkers in the system..."
    local LISTA_BUNKERES
    LISTA_BUNKERES=$(distrobox-list --no-color | awk -F'|' 'NR>1 {gsub(/[[:space:]]/, "", $2); if($2!="") print $2}')

    if [ -n "$LISTA_BUNKERES" ]; then
        echo "📂 Se han encontrado los siguientes búnkeres: / Found the following bunkers:"
        echo "------------------------------------------------"
        echo "$LISTA_BUNKERES" | sed 's/^/  • /'
        echo "------------------------------------------------"
        echo ""
        read -rp "⚠️  ¿Deseas borrar TODOS estos búnkeres y sus entornos? / Delete ALL these bunkers and environments? (s/N/y/N): " BORRAR_TODO
    else
        echo "ℹ️  No se han detectado búnkeres creados. / No created bunkers detected."
        BORRAR_TODO="n"
    fi
    echo ""

    if [[ "$BORRAR_TODO" =~ ^[sSyY]$ ]]; then
        read -rp "📝 Razón técnica para el reset total / Technical reason for full reset: " REASON
        if [ -z "$REASON" ]; then
            echo "❌ Operación cancelada: Se requiere una justificación técnica. / Operation canceled: Technical justification required."
            return 1
        fi

        echo "- Reset global ejecutado (Razón: $REASON)" >> "$TUTOR_PATH"

        echo "🔥 Iniciando limpieza profunda... / Starting deep cleanup..."
        for CAJA in $LISTA_BUNKERES; do
            echo "  🗑️  Eliminando / Deleting $CAJA..."
            distrobox-rm "$CAJA" --force --yes 2>/dev/null
            if [ -d "$BASE_ENV/$CAJA" ]; then
                chmod -R +w "$BASE_ENV/$CAJA" 2>/dev/null
                rm -rf "$BASE_ENV/$CAJA"
            fi
        done
        echo "✅ Todos los búnkeres han sido eliminados. / All bunkers have been deleted."
    fi

    echo ""
    echo "🗑️  Eliminando imagen base de Podman... / Deleting Podman base image..."
    if podman rmi "$IMAGEN" --force 2>/dev/null; then
        echo "✅ Imagen $IMAGEN eliminada con éxito. / Image $IMAGEN deleted successfully."
    else
        echo "⚠️  No se encontró la imagen o ya fue eliminada. / Image not found or already deleted."
    fi

    echo ""
    echo "✨ Sistema limpio. Usa 'axiom build' para generar una base nueva desde cero. / System clean. Use 'axiom build' to generate a new base from scratch."
}


rebuild() {
    mostrar_logo
    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)
    echo "🔄 Reconstruyendo imagen base $IMAGEN... / Rebuilding base image $IMAGEN..."
    echo "    Los búnkeres existentes NO se ven afectados. / Existing bunkers are NOT affected."
    echo ""
    read -rp "¿Continuar? / Continue? (s/N/y/N): " CONFIRM
    [[ "$CONFIRM" =~ ^[sSyY]$ ]] || return 0
    podman rmi "$IMAGEN" --force 2>/dev/null || true
    build
}

_escribir_opencode_config(){
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
        jq '.agent = (.agent // {}) | .agent["sdd-orchestrator"] = { "description": "Gentleman personality + SDD delegate-only orchestrator", "mode": "all", "prompt": "{file:./AGENTS.md}", "tools": { "bash": true, "edit": true, "read": true, "write": true } } | .agent["sdd-apply"] = { "description": "SDD delegate-only apply sub-agent", "mode": "all", "prompt": "{file:./AGENTS.md}", "tools": { "bash": true, "edit": true, "read": true, "write": true } }' "$CONF" > "${CONF}.tmp" && mv "${CONF}.tmp" "$CONF"
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