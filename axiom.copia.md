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
ROCM_MODE="${AXIOM_ROCM_MODE:-host}"

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

# ─── 5. VOLÚMENES GPU (modo host) ───────────────────


# ─── 6. INIT TUTOR.MD ───────────────────────────────
_init_tutor() {
    [ -f "$TUTOR_PATH" ] && return 0
    mkdir -p "$(dirname "$TUTOR_PATH")"

    local DISTRO GPU_INFO GFX_INFO
    DISTRO=$(grep ^NAME /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '"')
    [ -z "$DISTRO" ] && DISTRO="Linux"

    if [ -n "${GPU_TYPE:-}" ]; then
        GPU_INFO="${GPU_NAME:-$GPU_TYPE}"
        case "${GPU_TYPE}" in
            rdna*) GPU_INFO="AMD ${GPU_NAME:-RDNA} con ROCm" ;;
            nvidia) GPU_INFO="NVIDIA ${GPU_NAME:-} con CUDA" ;;
            intel) GPU_INFO="Intel ${GPU_NAME:-} con OneAPI" ;;
            *) GPU_INFO="${GPU_NAME:-Generic / CPU Only}" ;;
        esac
        GFX_INFO="${GFX_VAL:+ (HSA_OVERRIDE_GFX_VERSION=$GFX_VAL)}"
    else
        GPU_INFO="desconocida"
        GFX_INFO=""
    fi

    cat > "$TUTOR_PATH" << TUTOR
# 🤖 ROL: COPILOTO DE EJECUCIÓN (Junior Coder / Senior Mind)

## 👤 Identidad
Eres el brazo ejecutor del desarrollador. Tu misión es generar código limpio,
funcional y profesional a máxima velocidad, pero filtrado por un criterio de
Arquitecto Senior.

## 🌍 Contexto del Entorno
- Sistema: $DISTRO (atómico)
- Contenedor: Arch Linux via Distrobox (AXIOM)
- GPU: $GPU_INFO$GFX_INFO
- Modelos IA: Ollama en /ai_config/models
- Memoria persistente: /ai_global/teams/tutor.md

## 🛡️ Protocolo de Acción (Skeptic-to-Code)
1. **Skeptic First**: Antes de codear, pregunta el "porqué". Si la idea es mala
   o el código será basura, adviértelo. No seas un robot sumiso, sé un socio crítico.
2. **Explain & Validate**: Para tareas complejas, explica el diseño brevemente y
   espera el "OK". Para tareas simples y directas, ejecuta sin preguntar.
3. **High-Speed Execution**: Una vez recibas el "OK", genera el código completo.
   No des fragmentos inútiles; entrega bloques listos para ser probados o integrados.
4. **No Assumptions**: Si falta información para completar el código, pídela.
   Es mejor preguntar una vez que corregir diez.

## 🏛️ Estándares de Calidad
- **Clean Code & Pro Naming**: El código debe hablar por sí solo.
- **Detección de Errores**: Al entregar código, indica los 2 puntos más probables
  por donde podría fallar.
- **Git Ready**: Sugiere el momento del commit tras entregar un bloque funcional. El usuario decidirá el mensaje.

## 💾 Gestión del Entorno
- **Engram**: Registra archivos creados y decisiones técnicas para mantener contexto.
- **Save-Rule**: Si detectas una preferencia de código del desarrollador, sugiere
  grabarla con \`save-rule\` para que persista en todos los búnkeres.

## 📋 Cuándo guardar una regla
- Cuando se toma una decisión de arquitectura no obvia
- Cuando se resuelve un bug con una solución no trivial
- Cuando se establece un patrón que debe repetirse
- Cuando se descarta una tecnología con razón clara

## Reglas
- Protocolo de razón técnica activo.
TUTOR
    echo "✅ tutor.md inicializado con datos del sistema."
}

# ─── 7. SYNC-AGENTS ─────────────────────────────────
sync-agents() {
    _init_tutor
    local SYNCED=0
    while IFS= read -r CAJA; do
        [ -d "$BASE_ENV/$CAJA" ] || continue
        local DEST="$BASE_ENV/$CAJA/.config/opencode/AGENTS.md"
        mkdir -p "$(dirname "$DEST")"
        cp "$TUTOR_PATH" "$DEST"
        SYNCED=$((SYNCED+1))
    done < <(podman ps --format '{{.Names}}' 2>/dev/null)
    [ $SYNCED -gt 0 ] && echo "✅ Ley Global sincronizada en $SYNCED búnker(es)."
    return 0
}

# ─── 8. BUILD ───────────────────────────────────────
build() {
    mostrar_logo
    detect_gpu

    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo ""
    echo "🏗️  Construyendo imagen base: $IMAGEN"
    echo "    Modo GPU: $ROCM_MODE"
    if [ "$ROCM_MODE" = "host" ]; then
        echo "    ROCm se montará desde el host → imagen ~10-13 GB"
    else
        echo "    ROCm se instalará dentro → imagen ~38 GB"
    fi
    echo ""

    # Preparación de directorios globales
    mkdir -p "$AI_GLOBAL/models" "$AI_GLOBAL/teams" "$AI_CONFIG/models"
    sudo chown -R "$USER:$USER" "$AI_GLOBAL" "$AI_CONFIG"
    _init_tutor

    # ─── LIMPIEZA SEGURA DE PERMISOS (Fix Go Mod) ────────
    echo "🧹 Limpiando búnker de construcción anterior..."
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force 2>/dev/null || true

    if [ -d "$BASE_ENV/$AXIOM_BUILD_CONTAINER" ]; then
        echo "🔓 Desbloqueando caché de Go para eliminación..."
        # El comando clave: forzamos escritura recursiva
        chmod -R +w "$BASE_ENV/$AXIOM_BUILD_CONTAINER" 2>/dev/null
        rm -rf "$BASE_ENV/$AXIOM_BUILD_CONTAINER"
    fi
    # ─────────────────────────────────────────────────────

    echo "📦 Creando contenedor de build..."
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
    if [ "$ROCM_MODE" = "image" ]; then
        case "${GPU_TYPE}" in
            nvidia) GPU_PKGS="nvidia-utils cuda" ;;
            rdna*)  GPU_PKGS="rocm-hip-sdk" ;;
            intel)  GPU_PKGS="intel-compute-runtime onevpl-intel-gpu" ;;
        esac
    fi

    # Generamos el script interno
    cat > "$BUILD_SCRIPT" << SCRIPT
#!/bin/bash
set -uo pipefail

export PATH="\$HOME/.local/bin:\$HOME/go/bin:/usr/local/bin:\$PATH"
GPU_PKGS="${GPU_PKGS:-}"

echo "⚡ [1/4] Sistema base..."
sudo pacman -Sy --needed --noconfirm base-devel git curl jq wget nodejs npm go

echo "⚡ [2/4] Starship + GPU..."
curl -fsSL https://starship.rs/install.sh | sh -s -- --yes --bin-dir /usr/local/bin
if [ -n "\$GPU_PKGS" ]; then
    echo "⚡ Instalando GPU: \$GPU_PKGS"
    sudo pacman -S --noconfirm --needed \$GPU_PKGS
fi

echo "⚡ [3/4] Herramientas IA en paralelo..."
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

echo "⚡ Copiando binarios a /usr/local/bin..."
[ -f "\$HOME/go/bin/engram" ] && sudo cp -f "\$HOME/go/bin/engram" /usr/local/bin/

echo "🧹 Limpiando caché interna..."
sudo pacman -Scc --noconfirm
chmod -R +w ~/.cache/go ~/.cache 2>/dev/null || true
sudo rm -rf /tmp/* ~/.cache/go /var/cache/pacman/pkg 2>/dev/null || true

echo "✅ Build completo dentro del contenedor."
rm -- "\$0"
SCRIPT

    chmod +x "$BUILD_SCRIPT"
    distrobox-enter -n "$AXIOM_BUILD_CONTAINER" -- bash "$BUILD_SCRIPT"

    echo "📦 Exportando imagen $IMAGEN (esto puede tardar)..."
    podman commit "$AXIOM_BUILD_CONTAINER" "$IMAGEN"

    echo "🧹 Limpieza final..."
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force
    chmod -R +w "$BASE_ENV/$AXIOM_BUILD_CONTAINER" 2>/dev/null
    rm -rf "$BASE_ENV/$AXIOM_BUILD_CONTAINER"

    echo ""
    echo "✅ Imagen $IMAGEN lista. Ya puedes usar: crear [nombre]"
}

# ─── 9. CREAR ───────────────────────────────────────
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
    local IMAGEN
    IMAGEN=$(_imagen_base)

    if ! podman image exists "$IMAGEN"; then
        echo ""
        echo "⚠️  No se encontró la imagen base $IMAGEN"
        echo "    Ejecuta: build"
        return 1
    fi

    echo "⚡ Creando búnker '$NOMBRE' desde $IMAGEN..."
    mkdir -p "$R_PROYECTO" "$R_ENTORNO" "$AI_CONFIG/models"
    mkdir -p "$AI_GLOBAL/models" "$AI_GLOBAL/teams"
    sudo chown -R "$USER:$USER" "$AI_GLOBAL" "$AI_CONFIG"
    _init_tutor

    local GPU_VOLS=""
    [ "$ROCM_MODE" = "host" ] && GPU_VOLS=$(_gpu_volumes_host)

    distrobox-create --name "$NOMBRE" \
        --image "$IMAGEN" \
        --home "$R_ENTORNO" \
        --additional-flags "--volume $R_PROYECTO:/$NOMBRE \
        --volume $AI_GLOBAL:/ai_global \
        --volume $AI_CONFIG:/ai_config \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render \
        $GPU_VOLS" \
        --yes

    distrobox-enter -n "$NOMBRE" -- bash -c "
        sudo pacman -Syu --noconfirm --needed 2>/dev/null | tail -3
        echo '✅ Sistema actualizado.'
    "

    _escribir_bashrc "$NOMBRE" "$R_ENTORNO"
    _escribir_starship "$R_ENTORNO"

    mkdir -p "$R_ENTORNO/.config/opencode"
    [ -f "$TUTOR_PATH" ] && cp "$TUTOR_PATH" "$R_ENTORNO/.config/opencode/AGENTS.md"
    sync-agents

    distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
}

# ─── 10. BASHRC ─────────────────────────────────────
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



diagnostico() {
    echo "🔍 [DIAGNÓSTICO AXIOM]"
    echo "──────────────────────"
    echo "1️⃣  GPU:"
    if command -v nvidia-smi &>/dev/null; then
        nvidia-smi | grep "Driver Version" || echo "❌ nvidia-smi falló"
    elif command -v rocminfo &>/dev/null; then
        rocminfo | grep "Agent 1" -A 2 || echo "❌ rocminfo falló"
    else
        echo "⚠️ Sin herramientas de GPU visibles."
    fi
    echo ""
    echo "2️⃣  Git Token:"
    [ -n "$AXIOM_GIT_TOKEN" ] && echo "✅ Token presente." || echo "❌ AXIOM_GIT_TOKEN vacío."
    echo ""
    echo "3️⃣  Ollama:"
    pgrep -x ollama > /dev/null && echo "✅ En ejecución." || echo "⚠️ No está corriendo."
    echo ""
    echo "4️⃣  Herramientas IA:"
    for BIN in opencode engram gentle-ai ollama; do
        command -v $BIN &>/dev/null && echo "  ✅ $BIN" || echo "  ❌ $BIN no encontrado"
    done
    echo ""
    echo "5️⃣  AGENTS.md:"
    [ -f ~/.config/opencode/AGENTS.md ] && echo "✅ Presente." || echo "⚠️ No existe — ejecuta sync-agents"
}

open() {
    sync-agents
    _ollama_ensure && opencode
}

ayuda() {
    echo ""
    echo "🤖  BÚNKER — Comandos disponibles"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  open               Sincronizar leyes y abrir opencode"
    echo "  sync-agents        Copiar tutor.md a AGENTS.md"
    echo "  save-rule [regla]  Guardar regla en tutor.md"
    echo "  diagnostico        Diagnóstico de salud"
    echo ""
    echo "  git-clone [u/r]    Clonar repo de forma segura"
    echo "  rama               Crear rama nueva (interactivo)"
    echo "  commit [mensaje]   Añadir todo y commitear"
    echo "  push               Push seguro a GitHub"
    echo ""
}
BASH_RC

    if [[ -n "${GFX_VAL:-}" ]]; then
        echo "export HSA_OVERRIDE_GFX_VERSION=$GFX_VAL" >> "$R_ENTORNO/.bashrc"
    fi
    echo "cd /$NOMBRE" >> "$R_ENTORNO/.bashrc"

    mkdir -p "$R_ENTORNO/.config/opencode"
    cat > "$R_ENTORNO/.config/opencode/opencode.json" << 'OPENCODE_CONFIG'
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
}
# ─── 11. STARSHIP ───────────────────────────────────
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
# ─── 12. BORRAR ─────────────────────────────────────
borrar() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: borrar [nombre]"; return 1; fi
    read -rp "📝 Razón técnica obligatoria: " REASON
    [ -z "$REASON" ] && echo "❌ Cancelado: se requiere justificación." && return 1
    # Registro la razón en el global para que no sea código muerto
    echo "- Borrado búnker $1 (Razón: $REASON)" >> "$TUTOR_PATH"

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

# ─── 13. PARAR ──────────────────────────────────────
parar() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: parar [nombre]"; return 1; fi
    distrobox-list --no-color | grep -qw "$1" || { echo "❌ Búnker '$1' no existe."; return 1; }
    podman stop "$1" && echo "⏹️ Búnker '$1' parado."
}

# ─── 14. RESETEAR ───────────────────────────────────
resetear() {
    mostrar_logo
    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo "🗑️  Estado de la imagen base: $IMAGEN"
    if podman image exists "$IMAGEN"; then
        echo "    Tamaño: $(podman images --format '{{.Size}}' $IMAGEN)"
    else
        echo "    (La imagen no existe actualmente)"
    fi
    echo ""

    # 🔍 Escaneo de búnkeres existentes
    echo "🔎 Escaneando búnkeres en el sistema..."
    local LISTA_BUNKERES
    LISTA_BUNKERES=$(distrobox-list --no-color | awk -F'|' 'NR>1 {gsub(/[[:space:]]/, "", $2); if($2!="") print $2}')

    if [ -n "$LISTA_BUNKERES" ]; then
        echo "📂 Se han encontrado los siguientes búnkeres:"
        echo "------------------------------------------------"
        echo "$LISTA_BUNKERES" | sed 's/^/  • /'
        echo "------------------------------------------------"
        echo ""
        read -rp "⚠️  ¿Deseas borrar TODOS estos búnkeres y sus entornos? (s/N): " BORRAR_TODO
    else
        echo "ℹ️  No se han detectado búnkeres creados."
        BORRAR_TODO="n"
    fi
    echo ""

    # Ejecución del borrado de búnkeres
    if [[ "$BORRAR_TODO" =~ ^[sS]$ ]]; then
        read -rp "📝 Razón técnica para el reset total: " REASON
        if [ -z "$REASON" ]; then
            echo "❌ Operación cancelada: Se requiere una justificación técnica."
            return 1
        fi

        echo "- Reset global ejecutado (Razón: $REASON)" >> "$TUTOR_PATH"

        echo "🔥 Iniciando limpieza profunda..."
        for CAJA in $LISTA_BUNKERES; do
            echo "  🗑️  Eliminando $CAJA..."
            distrobox-rm "$CAJA" --force 2>/dev/null
            # El parche de permisos para evitar el error anterior:
            if [ -d "$BASE_ENV/$CAJA" ]; then
                chmod -R +w "$BASE_ENV/$CAJA" 2>/dev/null
                rm -rf "$BASE_ENV/$CAJA"
            fi
        done
        echo "✅ Todos los búnkeres han sido eliminados."
    fi

    # Borrado de la imagen base
    echo ""
    echo "🗑️  Eliminando imagen base de Podman..."
    if podman rmi "$IMAGEN" --force 2>/dev/null; then
        echo "✅ Imagen $IMAGEN eliminada con éxito."
    else
        echo "⚠️  No se encontró la imagen o ya fue eliminada."
    fi

    echo ""
    echo "✨ Sistema limpio. Usa 'build' para generar una base nueva desde cero."
}

# ─── 15. REBUILD ────────────────────────────────────
rebuild() {
    mostrar_logo
    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)
    echo "🔄 Reconstruyendo imagen base $IMAGEN..."
    echo "    Los búnkeres existentes NO se ven afectados."
    echo ""
    read -rp "¿Continuar? (s/N): " CONFIRM
    [[ "$CONFIRM" =~ ^[sS]$ ]] || return 0
    podman rmi "$IMAGEN" --force 2>/dev/null || true
    build
}

# ─── 16. AYUDA HOST ─────────────────────────────────
ayuda() {
    mostrar_logo
    echo ""
    echo "🛡️  SISTEMA BÚNKER — Comandos del host"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  build              Construir imagen base (primera vez)"
    echo "  rebuild            Reconstruir imagen base"
    echo "  resetear           Borrar imagen base y opcionalmente búnkeres"
    echo "  crear [nombre]     Crear búnker desde imagen base (~30 seg)"
    echo "  borrar [nombre]    Borrar búnker y su entorno"
    echo "  parar  [nombre]    Parar búnker sin borrarlo"
    echo "  ayuda              Mostrar esta ayuda"
    echo ""
    echo "  Modo GPU: $ROCM_MODE"
    echo "  Imágenes disponibles:"
    podman images --format "    {{.Repository}}:{{.Tag}}  ({{.Size}})" | grep axiom || \
        echo "    (ninguna — ejecuta: build)"
    echo ""
}
