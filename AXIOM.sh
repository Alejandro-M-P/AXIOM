

BASE_DEV="$HOME/Documentos/dev"
BASE_ENV="$BASE_DEV/.entorno"
AI_GLOBAL="$HOME/ai_config"
TUTOR_PATH="$AI_GLOBAL/teams/tutor.md"
GFX_VERSION="12.0.1"
AI_CONFIG="$BASE_DEV/ai_config"

# — 🧠 CEREBRO DE HARDWARE (GPU) —

detect_gpu() {
    echo "🔍 Detectando hardware gráfico..."
    local HAS_RDNA4=0
    local HAS_RDNA3=0
    local HAS_NVIDIA=0
    local HAS_INTEL=0

    # Valores por defecto de GFX (ROCm)
    local GFX_RDNA4="12.0.1"
    local GFX_RDNA3="11.0.0" 

    # Iterar línea por línea para soportar sistemas con múltiples GPUs (Laptops Híbridos)
    while IFS= read -r line; do
        # Extraer el Vendor ID usando sed (busca el patrón [XXXX:YYYY])
        local VENDOR
        VENDOR=$(echo "$line" | sed -n 's/.*\[\([0-9a-fA-F]\{4\}\):.*/\1/p')
        
        # Eliminar la dirección del bus (ej. 07:00.0) para evitar falsos positivos
        local DESC
        DESC=$(echo "$line" | cut -d ' ' -f 2-)

        case "${VENDOR,,}" in
            10de) # Vendor ID de NVIDIA
                HAS_NVIDIA=1
                ;;
            1002) # Vendor ID de AMD
                if echo "$DESC" | grep -iqE '(8[0-9]{3}|9[0-9]{3})'; then
                    HAS_RDNA4=1
                elif echo "$DESC" | grep -iqE '(6[0-9]{3}|7[0-9]{3})'; then
                    # Agrupamos RDNA 2 y 3 por compatibilidad base de ROCm
                    HAS_RDNA3=1 
                fi
                ;;
            8086) # Intel (Añadido detección automática)
                if echo "$DESC" | grep -iqE '(Arc|Graphics|Data Center GPU)'; then
                    HAS_INTEL=1
                fi
                ;;
        esac
    done < <(lspci -nn | grep -iE 'vga|3d|display|accelerator')

    # Asignación de Perfil y GFX_VERSION
    if [ "$HAS_NVIDIA" -eq 1 ]; then
        export GPU_TYPE="nvidia"
        export GFX_VAL=""
        echo "✅ GPU detectada: NVIDIA (No requiere GFX Override)"
    elif [ "$HAS_RDNA4" -eq 1 ]; then
        export GPU_TYPE="rdna4"
        export GFX_VAL="$GFX_RDNA4"
        echo "✅ GPU detectada: AMD RDNA 4 -> GFX: $GFX_VAL"
    elif [ "$HAS_RDNA3" -eq 1 ]; then
        export GPU_TYPE="rdna3"
        export GFX_VAL="$GFX_RDNA3"
        echo "✅ GPU detectada: AMD RDNA 3/2 -> GFX: $GFX_VAL"
    elif [ "$HAS_INTEL" -eq 1 ]; then
        export GPU_TYPE="intel"
        export GFX_VAL=""
        echo "✅ GPU detectada: Intel (Arc/OneAPI)"
    else
        echo "⚠️ Detección automática no concluyente."
        echo "Menú de selección manual:"
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

        # — LA CLAVE: El Override Manual —
        if [[ "$GPU_TYPE" == rdna* ]]; then
            read -rp "📝 ¿Deseas cambiar el GFX Override? (Enter para $GFX_VAL): " MANUAL_GFX
            [ -n "$MANUAL_GFX" ] && export GFX_VAL="$MANUAL_GFX"
        fi
        
        echo "✅ Perfil: $GPU_TYPE | GFX_VERSION: ${GFX_VAL:-N/A}"
    fi
}

# — ❓ AYUDA HOST —

ayuda() {
    echo ""
    echo "🛡️  SISTEMA BÚNKER — Comandos del host"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "  crear [nombre]   Crear o entrar a un búnker"
    echo "  borrar [nombre]  Borrar búnker y su entorno"
    echo "  parar  [nombre]  Parar un búnker sin borrarlo"
    echo ""
}

# — 🔄 SINCRONIZACIÓN DE LEYES (HOST) —

sync-agents() {
    [ ! -f "$TUTOR_PATH" ] && return 0
    local CAJAS
    mapfile -t CAJAS < <(distrobox-list --no-color | awk -F'|' 'NR>1 && $2!~/^\s*(ID)?\s*$/ {gsub(/[[:space:]]/, "", $2); if($2!="") print $2}')
    for CAJA in "${CAJAS[@]}"; do
        local DEST="$BASE_ENV/$CAJA/.config/opencode/AGENTS.md"
        [ ! -d "$BASE_ENV/$CAJA" ] && continue
        mkdir -p "$(dirname "$DEST")"
        cat "$TUTOR_PATH" >> "$DEST"
    done
    echo "✅ Ley Global sincronizada en búnkeres."
}

# — 🏗️ COMANDO CREAR / ENTRAR —

crear() {
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

    echo "⚡ Construyendo infraestructura…"
    mkdir -p "$R_PROYECTO" "$R_ENTORNO" "$AI_GLOBAL/models" "$AI_GLOBAL/teams"
    mkdir -p "$AI_CONFIG/models"
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

    cat > "$R_ENTORNO/setup.sh" << 'SCRIPT'
#!/bin/bash
set -euo pipefail

echo "⚡ Actualizando sistema base..."
sudo pacman -Syu --noconfirm base-devel git curl jq wget nodejs npm go

echo "⚡ Instalando paru..."
git clone https://aur.archlinux.org/paru.git /tmp/paru
cd /tmp/paru && makepkg -si --noconfirm
cd ~ && rm -rf /tmp/paru

echo "⚡ Instalando paquetes con paru..."
SCRIPT

    local PKGS="starship"

    case "${GPU_TYPE:-}" in
        nvidia) PKGS="$PKGS nvidia-utils cuda" ;;
        rdna*)  PKGS="$PKGS rocm-hip-sdk mesa-vdpau" ;;
        intel)  PKGS="$PKGS intel-compute-runtime onevpl-intel-gpu" ;;
    esac

    echo "echo '⚡ Instalando paquetes específicos para ${GPU_TYPE:-generic}...'" >> "$R_ENTORNO/setup.sh"
    echo "paru -S --noconfirm $PKGS" >> "$R_ENTORNO/setup.sh"

    cat >> "$R_ENTORNO/setup.sh" << 'SCRIPT'

# 🎨 ESTÉTICA VISUAL — Tokyo Night Extended

mkdir -p ~/.config && cat << 'EOF' > ~/.config/starship.toml

# Configuración "Professional Developer" - Tokyo Night Extended

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

Ubuntu = " "

Fedora = " "

Debian = " "

Linux = " "

Macos = " "

Windows = "󰍲 "

[directory]

style = "bg:#24283b fg:#e0af68"

format = "[ $path ]($style)"

truncation_length = 3

fish_style_pwd_dir_length = 1

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

# --- EXTRAS PROFESIONALES ---

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

[status]

disabled = false

format = '[\[$symbol $common_meaning$exit_code\]]($style) '

symbol = "✖"

style = "fg:#f7768e"

[jobs]

symbol = " "

style = "fg:#bb9af7"

format = "[$symbol$number]($style) "

[character]

success_symbol = "[󰁔](bold #9ece6a) "

error_symbol = "[󰁔](bold #f7768e) "

[custom.distrobox]

description = "Distrobox"

when = 'test -f /run/.containerenv'

command = 'grep "name=" /run/.containerenv | cut -d"\"" -f2'

symbol = "📦"

style = "bg:#1a1b26 fg:#bb9af7"

format = '[$symbol $output ]($style)'

# --- LENGUAJES ---

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

EOF
# 🤖 INSTALACIÓN IA EN PARALELO

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

echo "⚡ Instalando agent-teams-lite..."
git clone https://github.com/Gentleman-Programming/agent-teams-lite.git ~/agent-teams
cd ~/agent-teams && ./scripts/setup.sh --all && echo "✅ agent-teams listo" || echo "❌ agent-teams falló"
cd ~

# ✅ INICIALIZAR AGENTS.MD
echo "⚡ Inicializando AGENTS.md..."
mkdir -p ~/.config/opencode
[ -f /ai_global/teams/tutor.md ] && cat /ai_global/teams/tutor.md >> ~/.config/opencode/AGENTS.md
echo "✅ AGENTS.md listo."

SCRIPT

    # Nombre del proyecto se expande desde el host
    cat >> "$R_ENTORNO/setup.sh" << SCRIPT

if [[ -n "$GFX_VAL" ]]; then
    echo "export HSA_OVERRIDE_GFX_VERSION=$GFX_VAL" >> ~/.bashrc
fi

cat >> ~/.bashrc << 'BASH'
export OLLAMA_MODELS="/ai_config/models"
export PATH="\$HOME/.local/bin:\$HOME/go/bin:/usr/local/bin:\$PATH"
eval "\$(starship init bash)"
cd /$NOMBRE

_ollama_ensure() {
    local i=0
    pgrep -x ollama > /dev/null || (OLLAMA_MODELS="/ai_config/models" ollama serve > /tmp/ollama.log 2>&1 &)
    until ollama list &>/dev/null 2>&1; do
        sleep 1; i=\$((i+1))
        [ \$i -ge 30 ] && echo "❌ ollama no respondió en 30s" && return 1
    done
}

sync-agents() {
    [ ! -f /ai_global/teams/tutor.md ] && return 0
    mkdir -p ~/.config/opencode
    cat /ai_global/teams/tutor.md >> ~/.config/opencode/AGENTS.md
    echo "✅ AGENTS.md sincronizado."
}

save-rule() {
    read -rp "📝 Razón técnica: " REASON
    [ -z "\$REASON" ] && echo "❌ Error: Se requiere razón técnica." && return 1
    local R="\${1:-}"
    [ -z "\$R" ] && read -rp "📝 Regla: " R
    echo "- \$R (Razón: \$REASON)" >> /ai_global/teams/tutor.md
    sync-agents
    echo "✅ Regla actualizada en Host y Búnker."
}

git-clone() {
    if [ -z "\${1:-}" ]; then echo "❌ Uso: git-clone [usuario/repo] [carpeta]"; return 1; fi
    local TOKEN
    TOKEN=\$(cat /run/host/home/\$(whoami)/.git_token 2>/dev/null || cat ~/.git_token 2>/dev/null)
    if [ -z "\$TOKEN" ]; then echo "❌ No se encontró ~/.git_token en el host"; return 1; fi
    local REPO="\$1"
    local USUARIO=\$(echo "\$REPO" | cut -d'/' -f1)
    git clone "https://\${USUARIO}:\${TOKEN}@github.com/\${REPO}.git" \${2:-.}
    echo "✅ Repo clonado."
}

open() {
    sync-agents
    _ollama_ensure && opencode
}

ayuda() {
    echo ""
    echo "🤖  BÚNKER — Comandos disponibles"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "  open               Sincronizar leyes y abrir opencode"
    echo "  sync-agents        Copiar tutor.md a AGENTS.md"
    echo "  save-rule [regla]  Guardar regla técnica en tutor.md"
    echo "  git-clone [u/r]    Clonar repo de GitHub con token"
    echo ""
}
BASH
rm -- "\$0"
SCRIPT

    distrobox-enter -n "$NOMBRE" -- bash "$R_ENTORNO/setup.sh"
    sync-agents
    distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
}

# — 🗑️ COMANDO BORRAR —

axiom:borrar() {
    if [ -z "${1:-}" ]; then echo "❌ Uso: axiom:borrar [nombre]"; return 1; fi
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

# — ⏹️ COMANDO PARAR —

axiom:parar() {
    if [ -z "${1:-}" ]; then echo "❌ Uso: axiom:parar [nombre]"; return 1; fi
    if ! distrobox-list --no-color | grep -qw "$1"; then
        echo "❌ Búnker '$1' no existe."
        return 1
    fi
    podman stop "$1" && echo "⏹️ Búnker '$1' parado."
}
