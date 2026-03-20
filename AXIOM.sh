#!/bin/bash

# ─── 1. CARGA DE ENTORNO ────────────────────────────
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ ! -f "$DIR/.env" ]; then
    echo "❌ Error: No se encontró el archivo .env en $DIR"
    echo "Ejecuta ./install.sh primero."
    return 1 2>/dev/null || exit 1
fi
source "$DIR/.env"

# ─── 2. CÁLCULO DE RUTAS DERIVADAS ──────────────────
BASE_DEV="${AXIOM_BASE_DIR}"
BASE_ENV="$BASE_DEV/.entorno"
AI_GLOBAL="$BASE_DEV/ai_global"
AI_CONFIG="$BASE_DEV/ai_config"
TUTOR_PATH="$AI_GLOBAL/teams/tutor.md"

# ─── 3. LOGO AXIOM ──────────────────────────────────
mostrar_logo() {
    echo -e "\033[1;36m"
    echo "    _   _  _ ___ ___  __  __ "
    echo "   /_\ | \/ |_ _/ _ \|  \/  |"
    echo "  / _ \ >  < | | (_) | |\/| |"
    echo " /_/ \_/_/\_\___\___/|_|  |_|"
    echo -e "\033[0m"
}

# ─── 4. DETECCIÓN DE HARDWARE (GPU) ─────────────────
detect_gpu() {
    if [ -n "$AXIOM_GPU_TYPE" ]; then
        export GPU_TYPE="$AXIOM_GPU_TYPE"
        export GFX_VAL="$AXIOM_GFX_VAL"
        echo "✅ GPU forzada por .env: $GPU_TYPE (GFX: ${GFX_VAL:-N/A})"
        return 0
    fi

    echo "🔍 Detectando hardware gráfico automáticamente..."
    local HAS_RDNA4=0
    local HAS_RDNA3=0
    local HAS_NVIDIA=0
    local HAS_INTEL=0

    local GFX_RDNA4="12.0.1"
    local GFX_RDNA3="11.0.0" 

    while IFS= read -r line; do
        local VENDOR=$(echo "$line" | sed -n 's/.*\[\([0-9a-fA-F]\{4\}\):.*/\1/p')
        local DESC=$(echo "$line" | cut -d ' ' -f 2-)

        case "${VENDOR,,}" in
            10de) HAS_NVIDIA=1 ;;
            1002)
                if echo "$DESC" | grep -iqE '(8[0-9]{3}|9[0-9]{3})'; then
                    HAS_RDNA4=1
                elif echo "$DESC" | grep -iqE '(6[0-9]{3}|7[0-9]{3})'; then
                    HAS_RDNA3=1 
                fi
                ;;
            8086) HAS_INTEL=1 ;;
        esac
    done < <(lspci -nn | grep -iE 'vga|3d|display|accelerator')

    if [ "$HAS_NVIDIA" -eq 1 ]; then
        export GPU_TYPE="nvidia"; export GFX_VAL=""
        echo "✅ GPU detectada: NVIDIA"
    elif [ "$HAS_RDNA4" -eq 1 ]; then
        export GPU_TYPE="rdna4"; export GFX_VAL="$GFX_RDNA4"
        echo "✅ GPU detectada: AMD RDNA 4 -> GFX: $GFX_VAL"
    elif [ "$HAS_RDNA3" -eq 1 ]; then
        export GPU_TYPE="rdna3"; export GFX_VAL="$GFX_RDNA3"
        echo "✅ GPU detectada: AMD RDNA 3/2 -> GFX: $GFX_VAL"
    elif [ "$HAS_INTEL" -eq 1 ]; then
        export GPU_TYPE="intel"; export GFX_VAL=""
        echo "✅ GPU detectada: Intel (Arc/OneAPI)"
    else
        echo "⚠️ Detección automática no concluyente."
        echo "1. RDNA 4 (Serie 8000/9000)"
        echo "2. RDNA 3 | RDNA 2 (Serie 6000/7000)"
        echo "3. NVIDIA"
        echo "4. INTEL"
        echo "5. Generic / CPU Only"
        read -rp "Selecciona una opción [1-5]: " GPU_OPT
        
        case "$GPU_OPT" in
            1) export GPU_TYPE="rdna4"; export GFX_VAL="$GFX_RDNA4" ;;
            2) export GPU_TYPE="rdna3"; export GFX_VAL="$GFX_RDNA3" ;;
            3) export GPU_TYPE="nvidia"; export GFX_VAL="" ;;
            4) export GPU_TYPE="intel"; export GFX_VAL="" ;;
            *) export GPU_TYPE="generic"; export GFX_VAL="" ;;
        esac

        if [[ "$GPU_TYPE" == rdna* ]]; then
            read -rp "📝 ¿Deseas cambiar el GFX Override? (Enter para $GFX_VAL): " MANUAL_GFX
            [ -n "$MANUAL_GFX" ] && export GFX_VAL="$MANUAL_GFX"
        fi
    fi
}

# ─── 5. SINCRONIZACIÓN DE AGENTES (HOST) ────────────
sync-agents() {
    [ ! -f "$TUTOR_PATH" ] && return 0
    local CAJAS
    mapfile -t CAJAS < <(distrobox-list --no-color | awk -F'|' 'NR>1 && $2!~/^\s*(ID)?\s*$/ {gsub(/[[:space:]]/, "", $2); if($2!="") print $2}')
    
    for CAJA in "${CAJAS[@]}"; do
        if podman ps --format "{{.Names}}" | grep -qx "$CAJA"; then
            [ ! -d "$BASE_ENV/$CAJA" ] && continue
            local DEST="$BASE_ENV/$CAJA/.config/opencode/AGENTS.md"
            mkdir -p "$(dirname "$DEST")"
            
            if [ ! -f "$DEST" ]; then
                cat "$TUTOR_PATH" > "$DEST"
            else
                while IFS= read -r line; do
                    grep -qF "$line" "$DEST" || echo "$line" >> "$DEST"
                done < "$TUTOR_PATH"
            fi
        fi
    done
    echo "✅ Ley Global sincronizada en búnkeres activos."
}

# ─── 6. COMANDO CREAR ───────────────────────────────
crear() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: crear [nombre]"; return 1; fi
    local NOMBRE="$1"
    local R_PROYECTO="$BASE_DEV/$NOMBRE"
    local R_ENTORNO="$BASE_ENV/$NOMBRE"

    echo "🛡️ Acceso al Búnker '$NOMBRE':"
    if ! sudo -v; then echo "❌ Acceso denegado."; return 1; fi

    if distrobox-list --no-color | grep -qw "$NOMBRE"; then
        sync-agents
        distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
        return 0
    fi

    detect_gpu
    echo "⚡ Construyendo infraestructura…"
    mkdir -p "$R_PROYECTO" "$R_ENTORNO" "$AI_GLOBAL/models" "$AI_GLOBAL/teams" "$AI_CONFIG/models"
    [ ! -f "$TUTOR_PATH" ] && echo "- Protocolo de razón técnica activo." > "$TUTOR_PATH"

    distrobox-create --name "$NOMBRE" \
        --image archlinux:latest \
        --home "$R_ENTORNO" \
        --additional-flags "--volume $R_PROYECTO:/$NOMBRE \
        --volume $AI_GLOBAL:/ai_global \
        --volume $AI_CONFIG:/ai_config \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render" \
        --yes 2>/dev/null

    # 7. SCRIPT DE INSTALACIÓN DENTRO DEL CONTENEDOR
    cat > "$R_ENTORNO/setup.sh" << 'SCRIPT'
#!/bin/bash
set -euo pipefail

echo "⚡ Actualizando sistema base e instalando utilidades..."
sudo pacman -Syu --noconfirm base-devel git curl jq wget nodejs npm go

echo "⚡ Instalando paru..."
git clone https://aur.archlinux.org/paru.git /tmp/paru
cd /tmp/paru && makepkg -si --noconfirm
cd ~ && rm -rf /tmp/paru
SCRIPT

    local PKGS="starship"
    case "${GPU_TYPE:-}" in
        nvidia) PKGS="$PKGS nvidia-utils cuda" ;;
        rdna*)  PKGS="$PKGS rocm-hip-sdk mesa-vdpau" ;;
        intel)  PKGS="$PKGS intel-compute-runtime onevpl-intel-gpu" ;;
    esac

    echo "echo '⚡ Instalando paquetes específicos de GPU...'" >> "$R_ENTORNO/setup.sh"
    echo "paru -S --noconfirm $PKGS" >> "$R_ENTORNO/setup.sh"

    cat >> "$R_ENTORNO/setup.sh" << 'SCRIPT'

mkdir -p ~/.config && cat << 'EOF' > ~/.config/starship.toml
format = """
[](fg:#1a1b26)\
$os\
$custom\
[](fg:#1a1b26 bg:#24283b)\
$directory\
[](fg:#24283b bg:#414868)\
$git_branch\
$git_status\
$git_state\
$time\
[](fg:#414868) \
$fill\
$python$nodejs$rust$golang$c$docker_context$memory_usage$battery\
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
Arch = " "

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

[git_branch]
symbol = " "
style = "bg:#414868 fg:#bb9af7"
format = '[[ $symbol$branch ]($style)]($style)'

[git_status]
style = "bg:#414868 fg:#f7768e"
format = '[[($all_status$ahead_behind )]($style)]($style)'

[git_state]
style = "bg:#414868 fg:#f7768e"
format = '[[($state( $progress_current/$progress_total))]($style)]($style)'

[time]
disabled = false
time_format = "%R"
style = "bg:#414868 fg:#7dcfff"
format = '[[  $time ]($style)]($style)'

[python]
symbol = " "
format = 'via [${symbol}${version} ](bold #79c0ff)'
[nodejs]
symbol = "󰎙 "
format = 'via [${symbol}${version} ](bold #79c0ff)'
[rust]
symbol = "🦀 "
format = 'via [${symbol}${version} ](bold #ff7b72)'
[golang]
symbol = " "
format = 'via [${symbol}${version} ](bold #79c0ff)'
[c]
symbol = " "
format = 'via [${symbol}${version} ](bold #79c0ff)'

[docker_context]
symbol = " "
style = "fg:#0db7ed"
format = "[$symbol$context]($style) "

[memory_usage]
symbol = "󰍛 "
threshold = 75
style = "fg:#e0af68"
format = "[$symbol${ram}]($style) "
disabled = false

[battery]
full_symbol = "󰁹 "
charging_symbol = "󰂄 "
discharging_symbol = "󰂃 "
[[battery.display]]
threshold = 20
style = "bold #f7768e"

[cmd_duration]
min_time = 2_000
format = "took [󱎫 $duration]($style) "
style = "fg:#e0af68"

[jobs]
symbol = " "
style = "fg:#bb9af7"
format = "[$symbol$number]($style) "

[status]
disabled = false
format = '[\[$symbol $common_meaning$exit_code\]]($style) '
symbol = "✖"
style = "fg:#f7768e"

[character]
success_symbol = "[󰁔](bold #9ece6a) "
error_symbol = "[󰁔](bold #f7768e) "
EOF

echo "⚡ Instalando herramientas IA en paralelo..."
curl -fsSL https://ollama.com/install.sh | sh &
PID_OLLAMA=$!
curl -fsSL https://opencode.ai/install | bash &
PID_OPENCODE=$!
curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash &
PID_GENTLE=$!
go install github.com/Gentleman-Programming/engram/cmd/engram@latest &
PID_ENGRAM=$!

wait $PID_OLLAMA   && echo "✅ ollama listo"    || echo "❌ ollama falló"
wait $PID_OPENCODE && echo "✅ opencode listo"  || echo "❌ opencode falló"
wait $PID_GENTLE   && echo "✅ gentle-ai listo" || echo "❌ gentle-ai falló"
wait $PID_ENGRAM   && echo "✅ engram listo"    || echo "❌ engram falló"

echo "⚡ Instalando agent-teams-lite en serie..."
git clone https://github.com/Gentleman-Programming/agent-teams-lite.git ~/agent-teams
cd ~/agent-teams && ./scripts/setup.sh --all && echo "✅ agent-teams listo" || echo "❌ agent-teams falló"
cd ~

echo "⚡ Inicializando AGENTS.md..."
mkdir -p ~/.config/opencode
[ -f /ai_global/teams/tutor.md ] && cat /ai_global/teams/tutor.md >> ~/.config/opencode/AGENTS.md

rm -- "$0"
SCRIPT

    # 8. BASHRC DEL BÚNKER (Construido desde el host)
    cat > "$R_ENTORNO/.bashrc" << BASH_VARS
export AXIOM_GIT_USER="$AXIOM_GIT_USER"
export AXIOM_GIT_EMAIL="$AXIOM_GIT_EMAIL"
export AXIOM_GIT_TOKEN="$AXIOM_GIT_TOKEN"
export OLLAMA_HOST="$AXIOM_OLLAMA_HOST"
BASH_VARS
    cat >> "$R_ENTORNO/.bashrc" << 'BASH_RC'
export OLLAMA_MODELS="/ai_config/models"
export PATH="$HOME/.local/bin:$HOME/go/bin:/usr/local/bin:$PATH"
eval "$(starship init bash)"

_ollama_ensure() {
    local i=0
    pgrep -x ollama > /dev/null || (export OLLAMA_MODELS="/ai_config/models"; ollama serve > /tmp/ollama.log 2>&1 &)
    until ollama list &>/dev/null 2>&1; do
        sleep 1; i=$((i+1))
        [ $i -ge 30 ] && echo "❌ ollama no respondió en 30s" && return 1
    done
}

sync-agents() {
    [ ! -f /ai_global/teams/tutor.md ] && return 0
    mkdir -p ~/.config/opencode
    cat /ai_global/teams/tutor.md > ~/.config/opencode/AGENTS.md
    echo "✅ AGENTS.md sincronizado."
}

save-rule() {
    read -rp "📝 Razón técnica: " REASON
    [ -z "$REASON" ] && echo "❌ Error: Se requiere razón técnica." && return 1
    local R="${1:-}"
    [ -z "$R" ] && read -rp "📝 Regla: " R
    echo "- $R (Razón: $REASON)" >> /ai_global/teams/tutor.md
    sync-agents
}

git-clone() {
    if [ -z "${1:-}" ]; then echo "❌ Uso: git-clone [usuario/repo] [carpeta]"; return 1; fi
    if [ -z "$AXIOM_GIT_TOKEN" ]; then echo "❌ No se encontró AXIOM_GIT_TOKEN"; return 1; fi
    local REPO="$1"
    local DIR="${2:-$(basename "$REPO")}"
    git clone "https://${AXIOM_GIT_USER}:${AXIOM_GIT_TOKEN}@github.com/${REPO}.git" "$DIR"
    git -C "$DIR" remote set-url origin "https://github.com/${REPO}.git"
    echo "✅ Repo clonado y remote limpiado de credenciales."
}

push() {
    if [ -z "$AXIOM_GIT_TOKEN" ]; then echo "❌ Error: AXIOM_GIT_TOKEN no encontrado en .env."; return 1; fi
    git config user.name "$AXIOM_GIT_USER"
    git config user.email "$AXIOM_GIT_EMAIL"

    local REMOTE_URL
    REMOTE_URL=$(git config --get remote.origin.url)
    if [ -z "$REMOTE_URL" ]; then echo "❌ Error: No hay 'remote origin' configurado."; return 1; fi

    local REPO_PATH
    REPO_PATH=$(echo "$REMOTE_URL" | sed -E 's#.*github\.com[:/](.+?)(\.git)?$#\1#')
    if [ -z "$REPO_PATH" ]; then echo "❌ Error: No se pudo parsear el repositorio."; return 1; fi

    echo "⚡ Configurando URL temporal para push..."
    git remote set-url origin "https://${AXIOM_GIT_USER}:${AXIOM_GIT_TOKEN}@github.com/${REPO_PATH}.git"

    echo "🚀 Ejecutando git push..."
    git push "$@"
    local RET=$?

    echo "🧹 Limpiando URL remota..."
    git remote set-url origin "https://github.com/${REPO_PATH}.git"
    return $RET
}

diagnostico() {
    echo "🔍 [DIAGNÓSTICO DE SALUD: AXIOM]"
    echo "---------------------------------"
    echo "1️⃣  Visibilidad de GPU:"
    if command -v nvidia-smi &>/dev/null; then 
        nvidia-smi | grep "Driver Version" || echo "❌ Falla en nvidia-smi"
    elif command -v rocminfo &>/dev/null; then 
        rocminfo | grep "Agent 1" -A 2 || echo "❌ Falla en rocminfo"
    else 
        echo "⚠️ No se encontraron herramientas de GPU."
    fi

    echo ""
    echo "2️⃣  Acceso a Git Token:"
    if [ -n "$AXIOM_GIT_TOKEN" ]; then echo "✅ Token inyectado vía .env correctamente."
    else echo "❌ No se pudo acceder a AXIOM_GIT_TOKEN"
    fi

    echo ""
    echo "3️⃣  Estado de Ollama:"
    if pgrep -x ollama > /dev/null; then echo "✅ Ollama está en ejecución."
    else echo "⚠️ Ollama NO está en ejecución."
    fi
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
    echo "  save-rule [regla]  Guardar regla técnica en tutor.md"
    echo "  git-clone [u/r]    Clonar repo de GitHub con token"
    echo "  diagnostico        Ejecutar diagnóstico de salud del búnker"
    echo "  push               Hacer push automático a GitHub de forma segura"
    echo ""
}
BASH_RC

    if [[ -n "$GFX_VAL" ]]; then
        echo "export HSA_OVERRIDE_GFX_VERSION=$GFX_VAL" >> "$R_ENTORNO/.bashrc"
    fi
    echo "cd /$NOMBRE" >> "$R_ENTORNO/.bashrc"

    distrobox-enter -n "$NOMBRE" -- bash "$R_ENTORNO/setup.sh"
    sync-agents
    distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
}

# ─── 12. BORRAR ─────────────────────────────────────
borrar() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: borrar [nombre]"; return 1; fi
    read -rp "📝 Razón técnica obligatoria para borrar: " REASON
    if [ -z "$REASON" ]; then
        echo "❌ Operación cancelada: Se requiere justificar la eliminación."
        return 1
    fi
    read -rp "❗ ¿Borrar búnker '$1'? (s/N): " CONFIRM
    if [[ "$CONFIRM" =~ ^[sS]$ ]]; then
        distrobox-rm "$1" --force
        if [ -d "$BASE_ENV/$1" ]; then
            chmod -R +w "$BASE_ENV/$1"
            rm -rf "$BASE_ENV/$1"
        fi
        echo "🔥 Limpieza total. Memoria local eliminada."
    fi
}

# ─── 13. PARAR ──────────────────────────────────────
parar() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: parar [nombre]"; return 1; fi
    if ! distrobox-list --no-color | grep -qw "$1"; then
        echo "❌ Búnker '$1' no existe."
        return 1
    fi
    podman stop "$1" && echo "⏹️ Búnker '$1' parado."
}

# ─── 14. AYUDA HOST ─────────────────────────────────
ayuda() {
    mostrar_logo
    echo ""
    echo "🛡️  SISTEMA BÚNKER — Comandos del host"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  crear [nombre]   Crear o entrar a un búnker"
    echo "  borrar [nombre]  Borrar búnker y su entorno"
    echo "  parar  [nombre]  Parar un búnker sin borrarlo"
    echo "  ayuda            Muestra esta ayuda"
    echo ""
}