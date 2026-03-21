#!/bin/bash

# ─── 1. CARGA DE ENTORNO ────────────────────────────
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ ! -f "$DIR/.env" ]; then
    echo "❌ Error: No se encontró el archivo .env en $DIR"
    echo "Ejecuta ./install.sh primero."
    return 1 2>/dev/null || exit 1
fi
source "$DIR/.env"

# ─── 2. RUTAS DERIVADAS ─────────────────────────────
BASE_DEV="${AXIOM_BASE_DIR}"
BASE_ENV="$BASE_DEV/.entorno"
AI_GLOBAL="$BASE_DEV/ai_global"
AI_CONFIG="$BASE_DEV/ai_config"
TUTOR_PATH="$AI_GLOBAL/teams/tutor.md"
AXIOM_BUILD_CONTAINER="axiom-build"

# ─── 3. LOGO ────────────────────────────────────────
mostrar_logo() {
    echo -e "\033[1;36m"
    echo "    _   _  _ ___ ___  __  __ "
    echo "   /_\ | \/ |_ _/ _ \|  \/  |"
    echo "  / _ \ >  < | | (_) | |\/| |"
    echo " /_/ \_/_/\_\___\___/|_|  |_|"
    echo -e "\033[0m"
}

# ─── 4. DETECCIÓN DE GPU ────────────────────────────
detect_gpu() {
    if [ -n "${AXIOM_GPU_TYPE:-}" ]; then
        export GPU_TYPE="$AXIOM_GPU_TYPE"
        export GFX_VAL="${AXIOM_GFX_VAL:-}"
        echo "✅ GPU forzada por .env: $GPU_TYPE (GFX: ${GFX_VAL:-N/A})"
        return 0
    fi

    echo "🔍 Detectando hardware gráfico..."
    local HAS_RDNA4=0 HAS_RDNA3=0 HAS_NVIDIA=0 HAS_INTEL=0
    local GFX_RDNA4="12.0.1" GFX_RDNA3="11.0.0"

    while IFS= read -r line; do
        local VENDOR DESC
        VENDOR=$(echo "$line" | sed -n 's/.*\[\([0-9a-fA-F]\{4\}\):.*/\1/p')
        DESC=$(echo "$line" | cut -d ' ' -f 2-)
        case "${VENDOR,,}" in
            10de) HAS_NVIDIA=1 ;;
            1002)
                echo "$DESC" | grep -iqE '(8[0-9]{3}|9[0-9]{3})' && HAS_RDNA4=1
                echo "$DESC" | grep -iqE '(6[0-9]{3}|7[0-9]{3})' && HAS_RDNA3=1
                ;;
            8086) HAS_INTEL=1 ;;
        esac
    done < <(lspci -nn | grep -iE 'vga|3d|display|accelerator')

    if   [ "$HAS_NVIDIA" -eq 1 ]; then export GPU_TYPE="nvidia"; export GFX_VAL=""
    elif [ "$HAS_RDNA4"  -eq 1 ]; then export GPU_TYPE="rdna4";  export GFX_VAL="$GFX_RDNA4"
    elif [ "$HAS_RDNA3"  -eq 1 ]; then export GPU_TYPE="rdna3";  export GFX_VAL="$GFX_RDNA3"
    elif [ "$HAS_INTEL"  -eq 1 ]; then export GPU_TYPE="intel";  export GFX_VAL=""
    else
        echo "⚠️ Detección no concluyente. Selecciona:"
        echo "1. RDNA 4 (8000/9000)  2. RDNA 3/2 (6000/7000)"
        echo "3. NVIDIA              4. INTEL"
        echo "5. Generic / CPU Only"
        read -rp "Opción [1-5]: " GPU_OPT
        case "$GPU_OPT" in
            1) export GPU_TYPE="rdna4";   export GFX_VAL="$GFX_RDNA4" ;;
            2) export GPU_TYPE="rdna3";   export GFX_VAL="$GFX_RDNA3" ;;
            3) export GPU_TYPE="nvidia";  export GFX_VAL="" ;;
            4) export GPU_TYPE="intel";   export GFX_VAL="" ;;
            *) export GPU_TYPE="generic"; export GFX_VAL="" ;;
        esac
        if [[ "$GPU_TYPE" == rdna* ]]; then
            read -rp "📝 GFX Override (Enter para $GFX_VAL): " MANUAL_GFX
            [ -n "$MANUAL_GFX" ] && export GFX_VAL="$MANUAL_GFX"
        fi
    fi
    echo "✅ GPU: $GPU_TYPE ${GFX_VAL:+(GFX: $GFX_VAL)}"
}

# nombre de la imagen base según GPU
_imagen_base() {
    echo "localhost/axiom-${GPU_TYPE:-generic}:latest"
}

# ─── 5. SYNC-AGENTS ─────────────────────────────────
sync-agents() {
    [ ! -f "$TUTOR_PATH" ] && return 0
    while IFS= read -r CAJA; do
        [ -d "$BASE_ENV/$CAJA" ] || continue
        local DEST="$BASE_ENV/$CAJA/.config/opencode/AGENTS.md"
        mkdir -p "$(dirname "$DEST")"
        cp "$TUTOR_PATH" "$DEST"
    done < <(podman ps --format '{{.Names}}')
    echo "✅ Ley Global sincronizada."
}

# ─── 6. BUILD — construye la imagen base completa ───
build() {
    mostrar_logo
    detect_gpu

    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo ""
    echo "🏗️  Construyendo imagen base: $IMAGEN"
    echo "    Esto tarda ~15-30 min. Solo se hace una vez."
    echo ""

    # Directorios necesarios
    mkdir -p "$AI_GLOBAL/models" "$AI_GLOBAL/teams"
    sudo chown -R "$USER:$USER" "$AI_GLOBAL"
    [ ! -f "$TUTOR_PATH" ] && echo "- Protocolo de razón técnica activo." > "$TUTOR_PATH"

    # Eliminar contenedor de build anterior si existe
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force 2>/dev/null || true

    # Crear contenedor de build limpio
    distrobox-create --name "$AXIOM_BUILD_CONTAINER" \
        --image archlinux:latest \
        --home "$BASE_ENV/$AXIOM_BUILD_CONTAINER" \
        --additional-flags "--volume $AI_GLOBAL:/ai_global \
        --volume $AI_CONFIG:/ai_config \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render" \
        --yes

    # Script de build completo dentro del contenedor
    local BUILD_SCRIPT="$BASE_ENV/$AXIOM_BUILD_CONTAINER/axiom-build.sh"
    mkdir -p "$BASE_ENV/$AXIOM_BUILD_CONTAINER"

    # Paquetes GPU según tipo detectado
    local GPU_PKGS=""
    case "${GPU_TYPE}" in
        nvidia)  GPU_PKGS="nvidia-utils cuda" ;;
        rdna*)   GPU_PKGS="rocm-hip-sdk" ;;
        intel)   GPU_PKGS="intel-compute-runtime onevpl-intel-gpu" ;;
    esac

    cat > "$BUILD_SCRIPT" << SCRIPT
#!/bin/bash
set -uo pipefail

export PATH="\$HOME/.local/bin:\$HOME/go/bin:/usr/local/bin:\$PATH"

echo "⚡ [1/4] Sistema base..."
sudo pacman -Sy --needed --noconfirm base-devel git curl jq wget nodejs npm go

echo "⚡ [2/4] Instalando paru..."
if ! command -v paru &>/dev/null; then
    git clone https://aur.archlinux.org/paru.git /tmp/paru
    cd /tmp/paru && makepkg -si --noconfirm
    cd ~ && rm -rf /tmp/paru
fi

echo "⚡ [3/4] Paquetes GPU: ${GPU_PKGS:-ninguno} + starship..."
paru -S --noconfirm --needed starship ${GPU_PKGS}

echo "⚡ [4/4] Instalando herramientas IA en paralelo..."

curl -fsSL https://opencode.ai/install | bash &
PID_OC=\$!

go install github.com/Gentleman-Programming/engram/cmd/engram@latest &
PID_EN=\$!

curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash &
PID_GA=\$!

curl -fsSL https://ollama.com/install.sh | sh &
PID_OL=\$!

wait \$PID_OC && echo "✅ opencode" || echo "❌ opencode falló"
wait \$PID_EN && echo "✅ engram"   || echo "❌ engram falló"
wait \$PID_GA && echo "✅ gentle-ai"|| echo "❌ gentle-ai falló"
wait \$PID_OL && echo "✅ ollama"   || echo "❌ ollama falló"

echo "🧹 Limpiando caché para reducir tamaño de imagen..."
sudo pacman -Scc --noconfirm
paru -Scc --noconfirm 2>/dev/null || true
sudo rm -rf /tmp/* ~/.cache/go ~/.cache/paru /var/cache/pacman/pkg

echo "✅ Imagen base lista."
rm -- "\$0"
SCRIPT

    chmod +x "$BUILD_SCRIPT"
    distrobox-enter -n "$AXIOM_BUILD_CONTAINER" -- bash "$BUILD_SCRIPT"

    echo "📦 Exportando imagen $IMAGEN..."
    podman commit "$AXIOM_BUILD_CONTAINER" "$IMAGEN"

    echo "🧹 Limpiando contenedor de build..."
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force
    rm -rf "$BASE_ENV/$AXIOM_BUILD_CONTAINER"

    echo ""
    echo "✅ Imagen $IMAGEN lista. Ya puedes usar: crear [nombre]"
}

# ─── 7. CREAR ────────────────────────────────────────
crear() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: crear [nombre]"; return 1; fi
    local NOMBRE="$1"
    local R_PROYECTO="$BASE_DEV/$NOMBRE"
    local R_ENTORNO="$BASE_ENV/$NOMBRE"

    echo "🛡️ Acceso al Búnker '$NOMBRE':"
    if ! sudo -v; then echo "❌ Acceso denegado."; return 1; fi

    # Si ya existe, entrar directamente
    if distrobox-list --no-color | grep -qw "$NOMBRE"; then
        sync-agents
        distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
        return 0
    fi

    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)

    # Verificar que existe la imagen base
    if ! podman image exists "$IMAGEN"; then
        echo ""
        echo "⚠️  No se encontró la imagen base $IMAGEN"
        echo "    Ejecuta: build"
        echo "    (Solo se hace una vez, tarda ~15-30 min)"
        return 1
    fi

    echo "⚡ Creando búnker '$NOMBRE' desde $IMAGEN..."
    mkdir -p "$R_PROYECTO" "$R_ENTORNO" "$AI_CONFIG/models"
    mkdir -p "$AI_GLOBAL/models" "$AI_GLOBAL/teams" 2>/dev/null || \
        sudo mkdir -p "$AI_GLOBAL/models" "$AI_GLOBAL/teams"
    sudo chown -R "$USER:$USER" "$AI_GLOBAL"
    [ ! -f "$TUTOR_PATH" ] && echo "- Protocolo de razón técnica activo." > "$TUTOR_PATH"

    distrobox-create --name "$NOMBRE" \
        --image "$IMAGEN" \
        --home "$R_ENTORNO" \
        --additional-flags "--volume $R_PROYECTO:/$NOMBRE \
        --volume $AI_GLOBAL:/ai_global \
        --volume $AI_CONFIG:/ai_config \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render" \
        --yes

    # Actualización rápida (solo diffs desde el último build)
    distrobox-enter -n "$NOMBRE" -- bash -c "
        sudo pacman -Syu --noconfirm --needed 2>/dev/null | tail -3
        echo '✅ Sistema actualizado.'
    "

    # Escribir .bashrc con variables del host
    _escribir_bashrc "$NOMBRE" "$R_ENTORNO"

    # Starship config
    _escribir_starship "$R_ENTORNO"

    sync-agents
    distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
}

# ─── 8. BASHRC ──────────────────────────────────────
_escribir_bashrc() {
    local NOMBRE="$1" R_ENTORNO="$2"

    cat > "$R_ENTORNO/.bashrc" << BASH_VARS
export AXIOM_GIT_USER="$AXIOM_GIT_USER"
export AXIOM_GIT_EMAIL="$AXIOM_GIT_EMAIL"
export AXIOM_GIT_TOKEN="$AXIOM_GIT_TOKEN"
export OLLAMA_HOST="${AXIOM_OLLAMA_HOST:-}"
export OLLAMA_MODELS="${AXIOM_MODELS_DIR:-/ai_config/models}"
BASH_VARS

    cat >> "$R_ENTORNO/.bashrc" << 'BASH_RC'
export PATH="$HOME/.local/bin:$HOME/go/bin:/usr/local/bin:$PATH"
eval "$(starship init bash)"

_ollama_ensure() {
    ollama list &>/dev/null && return 0
    ollama serve > /tmp/ollama.log 2>&1 &
    local i=0
    until ollama list &>/dev/null; do
        sleep 1; i=$((i+1))
        [ $i -ge 15 ] && echo "❌ ollama no respondió en 15s" && return 1
    done
}

sync-agents() {
    [ ! -f /ai_global/teams/tutor.md ] && return 0
    mkdir -p ~/.config/opencode
    cp /ai_global/teams/tutor.md ~/.config/opencode/AGENTS.md
    echo "✅ AGENTS.md sincronizado."
}

save-rule() {
    read -rp "📝 Razón técnica: " REASON
    [ -z "$REASON" ] && echo "❌ Se requiere razón técnica." && return 1
    local R="${1:-}"
    [ -z "$R" ] && read -rp "📝 Regla: " R
    echo "- $R (Razón: $REASON)" >> /ai_global/teams/tutor.md
    sync-agents
}

git-clone() {
    if [ -z "${1:-}" ]; then echo "❌ Uso: git-clone [usuario/repo] [carpeta]"; return 1; fi
    [ -z "$AXIOM_GIT_TOKEN" ] && echo "❌ No se encontró AXIOM_GIT_TOKEN" && return 1
    local REPO="$1" DIR="${2:-$(basename "$1")}"
    git clone "https://${AXIOM_GIT_USER}:${AXIOM_GIT_TOKEN}@github.com/${REPO}.git" "$DIR"
    git -C "$DIR" remote set-url origin "https://github.com/${REPO}.git"
    echo "✅ Repo clonado y remote limpiado."
}

push() {
    [ -z "$AXIOM_GIT_TOKEN" ] && echo "❌ AXIOM_GIT_TOKEN no encontrado." && return 1
    git config user.name "$AXIOM_GIT_USER"
    git config user.email "$AXIOM_GIT_EMAIL"
    local REMOTE_URL REPO_PATH
    REMOTE_URL=$(git config --get remote.origin.url)
    [ -z "$REMOTE_URL" ] && echo "❌ No hay remote origin." && return 1
    REPO_PATH=$(echo "$REMOTE_URL" | sed -E 's#.*github\.com[:/](.+?)(\.git)?$#\1#')
    [ -z "$REPO_PATH" ] && echo "❌ No se pudo parsear el repositorio." && return 1
    git remote set-url origin "https://${AXIOM_GIT_USER}:${AXIOM_GIT_TOKEN}@github.com/${REPO_PATH}.git"
    git push "$@"
    local RET=$?
    git remote set-url origin "https://github.com/${REPO_PATH}.git"
    return $RET
}

diagnostico() {
    echo "🔍 [DIAGNÓSTICO AXIOM]"
    echo "──────────────────────"
    echo "1️⃣  GPU:"
    if command -v nvidia-smi &>/dev/null; then
        nvidia-smi | grep "Driver Version" || echo "❌ nvidia-smi falló"
    elif command -v rocminfo &>/dev/null; then
        rocminfo | grep "Agent 1" -A 2 || echo "❌ rocminfo falló"
    else
        echo "⚠️ Sin herramientas de GPU."
    fi
    echo ""
    echo "2️⃣  Git Token:"
    [ -n "$AXIOM_GIT_TOKEN" ] && echo "✅ Token presente." || echo "❌ AXIOM_GIT_TOKEN no encontrado."
    echo ""
    echo "3️⃣  Ollama:"
    pgrep -x ollama > /dev/null && echo "✅ En ejecución." || echo "⚠️ No está corriendo."
}

open() {
    sync-agents
    _ollama_ensure && opencode
}

mostrar_logo() {
    echo -e "\033[1;36m"
    echo "    _   _  _ ___ ___  __  __ "
    echo "   /_\ | \/ |_ _/ _ \|  \/  |"
    echo "  / _ \ >  < | | (_) | |\/| |"
    echo " /_/ \_/_/\_\___\___/|_|  |_|"
    echo -e "\033[0m"
}

ayuda() {
    mostrar_logo
    echo ""
    echo "🤖  BÚNKER — Comandos disponibles"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  open               Sincronizar leyes y abrir opencode"
    echo "  sync-agents        Copiar tutor.md a AGENTS.md"
    echo "  save-rule [regla]  Guardar regla en tutor.md"
    echo "  git-clone [u/r]    Clonar repo con token"
    echo "  diagnostico        Diagnóstico de salud"
    echo "  push               Push seguro a GitHub"
    echo ""
}
BASH_RC

    if [[ -n "${GFX_VAL:-}" ]]; then
        echo "export HSA_OVERRIDE_GFX_VERSION=$GFX_VAL" >> "$R_ENTORNO/.bashrc"
    fi
    echo "cd /$NOMBRE" >> "$R_ENTORNO/.bashrc"
}

# ─── 9. STARSHIP ────────────────────────────────────
_escribir_starship() {
    local R_ENTORNO="$1"
    mkdir -p "$R_ENTORNO/.config"
    cat > "$R_ENTORNO/.config/starship.toml" << 'STARSHIP'
format = """
[](fg:#1a1b26)\
$os\
$custom\
[](fg:#1a1b26 bg:#24283b)\
$directory\
[](fg:#24283b bg:#414868)\
$git_branch\
$git_status\
$git_state\
$git_metrics\
$time\
[](fg:#414868) \
$python$nodejs$rust$golang$c\
$fill\
$memory_usage\
$cmd_duration\
$jobs\
$status\
$line_break\
$character"""

[fill]
symbol = " "

[os]
disabled = false
style = "bg:#1a1b26 fg:#7aa2f7"
format = "[ $symbol ]($style)"

[os.symbols]
Arch = " "
Ubuntu = " "
Fedora = " "
Debian = " "
Linux = " "
Macos = " "
Windows = "󰍲 "

[custom.distrobox]
description = "Distrobox"
when = 'test -f /run/.containerenv'
command = 'grep "name=" /run/.containerenv | cut -d"\"" -f2'
symbol = "📦"
style = "bg:#1a1b26 fg:#bb9af7"
format = '[$symbol $output ]($style)'

[directory]
style = "bg:#24283b fg:#e0af68"
format = "[ $path ]($style)"
truncation_length = 3
fish_style_pwd_dir_length = 1

[git_branch]
symbol = " "
style = "bg:#414868 fg:#bb9af7"
format = '[[ $symbol$branch ]($style)]($style)'
truncation_length = 20
truncation_symbol = "…"

[git_status]
style = "bg:#414868 fg:#f7768e"
format = '[[( $all_status$ahead_behind )]($style)]($style)'
ahead = "⇡${count}"
behind = "⇣${count}"
diverged = "⇕⇡${ahead_count}⇣${behind_count}"
staged = "[+${count}](bold green)"
modified = "[~${count}](bold yellow)"
untracked = "[?${count}](bold red)"
deleted = "[-${count}](bold red)"
conflicted = "[=${count}](bold red)"
stashed = "[󰏗 ${count}](bold blue)"

[git_state]
style = "bg:#414868 fg:#f7768e"
format = '[[( $state $progress_current/$progress_total)]($style)]($style)'

[git_metrics]
added_style = "bold #9ece6a"
deleted_style = "bold #f7768e"
format = '([+$added]($added_style) )([-$deleted]($deleted_style) )'
disabled = false

[time]
disabled = false
time_format = "%R"
style = "bg:#414868 fg:#7dcfff"
format = '[[  $time ]($style)]($style)'

[cmd_duration]
min_time = 2_000
format = "took [󱎫 $duration]($style) "
style = "fg:#e0af68"

[status]
disabled = false
format = '[\[$symbol $common_meaning$exit_code\]]($style) '
symbol = "✖"
style = "fg:#f7768e"

[jobs]
symbol = " "
style = "fg:#bb9af7"
format = "[$symbol$number]($style) "

[memory_usage]
symbol = "󰍛 "
threshold = 75
style = "fg:#e0af68"
format = "[$symbol${ram}]($style) "
disabled = false

[character]
success_symbol = "[󰁔](bold #9ece6a) "
error_symbol = "[󰁔](bold #f7768e) "

[python]
symbol = " "
format = 'via [${symbol}${version} ](bold #79c0ff)'

[nodejs]
symbol = "󰎙 "
format = 'via [${symbol}${version} ](bold #79c0ff)'

[rust]
symbol = "🦀 "
format = 'via [${symbol}${version} ](bold #ff7b72)'

[golang]
symbol = " "
format = 'via [${symbol}${version} ](bold #79c0ff)'

[c]
symbol = " "
format = 'via [${symbol}${version} ](bold #79c0ff)'
STARSHIP
}

# ─── 10. BORRAR ─────────────────────────────────────
borrar() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: borrar [nombre]"; return 1; fi
    read -rp "📝 Razón técnica obligatoria: " REASON
    [ -z "$REASON" ] && echo "❌ Cancelado: se requiere justificación." && return 1
    read -rp "❗ ¿Borrar búnker '$1'? (s/N): " CONFIRM
    if [[ "$CONFIRM" =~ ^[sS]$ ]]; then
        distrobox-rm "$1" --force
        if [ -d "$BASE_ENV/$1" ]; then
            chmod -R +w "$BASE_ENV/$1"
            rm -rf "$BASE_ENV/$1"
        fi
        echo "🔥 Búnker '$1' eliminado."
    fi
}

# ─── 11. PARAR ──────────────────────────────────────
parar() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: parar [nombre]"; return 1; fi
    distrobox-list --no-color | grep -qw "$1" || { echo "❌ Búnker '$1' no existe."; return 1; }
    podman stop "$1" && echo "⏹️ Búnker '$1' parado."
}

# ─── 12. REBUILD — actualiza la imagen base ─────────
rebuild() {
    mostrar_logo
    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo "🔄 Reconstruyendo imagen base $IMAGEN..."
    echo "    Los búnkeres existentes NO se ven afectados."
    echo "    Los nuevos búnkeres usarán la imagen actualizada."
    echo ""
    read -rp "¿Continuar? (s/N): " CONFIRM
    [[ "$CONFIRM" =~ ^[sS]$ ]] || return 0

    podman rmi "$IMAGEN" --force 2>/dev/null || true
    build
}

# ─── 13. AYUDA HOST ─────────────────────────────────
ayuda() {
    mostrar_logo
    echo ""
    echo "🛡️  SISTEMA BÚNKER — Comandos del host"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  build              Construir imagen base (primera vez, ~20 min)"
    echo "  rebuild            Reconstruir imagen base (actualizar)"
    echo "  crear [nombre]     Crear búnker desde imagen base (~30 seg)"
    echo "  borrar [nombre]    Borrar búnker y su entorno"
    echo "  parar  [nombre]    Parar búnker sin borrarlo"
    echo "  ayuda              Mostrar esta ayuda"
    echo ""
    echo "  Imágenes base disponibles:"
    podman images --format "    {{.Repository}}:{{.Tag}}  ({{.Size}})" | grep axiom || \
        echo "    (ninguna — ejecuta: build)"
    echo ""
}