#!/bin/bash
# ─── MÓDULO BUNKER: CICLO DE VIDA (BUILD, CREATE, DELETE, RESET, REBUILD, PRUNE) ──────────

AXIOM_BUILD_CONTAINER="axiom-build"

build() {
    mostrar_logo
    detect_gpu

    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo ""
    echo "🏗️  Construyendo imagen base / Building base image: $IMAGEN"
    echo "    Modo GPU / GPU Mode: $AXIOM_ROCM_MODE"
    if [ "$AXIOM_ROCM_MODE" = "host" ]; then
        echo "    ROCm se montará desde el host / ROCm will be mounted from host → imagen ~10-13 GB"
    else
        echo "    ROCm se instalará dentro / ROCm will be installed inside → imagen ~38 GB"
    fi
    echo ""

    mkdir -p "$AI_CONFIG/models" "$AI_CONFIG/teams"
    sudo chown -R "$USER:$USER" "$AI_CONFIG"
    _init_tutor

    echo "🧹 Limpiando búnker de construcción anterior... / Cleaning previous build bunker..."
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force --yes 2>/dev/null || true

    if [ -d "$BASE_ENV/$AXIOM_BUILD_CONTAINER" ]; then
        echo "🔓 Desbloqueando caché de Go para eliminación... / Unlocking Go cache for deletion..."
        chmod -R +w "$BASE_ENV/$AXIOM_BUILD_CONTAINER" 2>/dev/null
        rm -rf "$BASE_ENV/$AXIOM_BUILD_CONTAINER"
    fi

    echo "📦 Creando contenedor de build... / Creating build container..."
    # Card 2: se añade el volumen .env read-only para que el build script / Add read-only .env volume so the build script
    # pueda leer el token on-demand sin recibirlo como variable de entorno. / can read token on-demand without env var exposure.
    distrobox-create --name "$AXIOM_BUILD_CONTAINER" \
        --image archlinux:latest \
        --home "$BASE_ENV/$AXIOM_BUILD_CONTAINER" \
        --additional-flags "--volume $AI_CONFIG:/ai_config \
        --volume \"$DIR/.env\":/run/axiom/env:ro \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render" \
        --yes

    # Card 4: usar mktemp en lugar de ruta fija para evitar TOCTOU. / use mktemp instead of fixed path to prevent TOCTOU.
    local BUILD_SCRIPT
    BUILD_SCRIPT=$(mktemp /tmp/axiom-build-XXXXXX.sh)
    chmod 700 "$BUILD_SCRIPT"
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
set -euo pipefail

export PATH="\$HOME/.local/bin:\$HOME/go/bin:/usr/local/bin:\$PATH"
GPU_PKGS="${GPU_PKGS:-}"

echo "⚡ [1/4] Sistema base..."
sudo pacman -Sy --needed --noconfirm base-devel git curl jq wget nodejs npm go fzf

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
    # Card 3: array de argumentos en lugar de eval para evitar inyección. / Argument arrays instead of eval to prevent injection.
    TOKEN=\$(grep -oP '(?<=AXIOM_GIT_TOKEN=")[^"]+' /run/axiom/env 2>/dev/null || echo "")

    CURL_ARGS=(-fsSL --fail)
    [ -n "\$TOKEN" ] && CURL_ARGS+=(-H "Authorization: Bearer \$TOKEN")

    GA_LATEST=\$(curl "\${CURL_ARGS[@]}" \
        "https://api.github.com/repos/Gentleman-Programming/gentle-ai/releases/latest" \
        | grep -o '"tag_name": *"[^"]*"' | grep -o '[0-9][^"]*' || echo "")

    [ -z "\$GA_LATEST" ] && GA_LATEST="0.1.0"

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

wait \$PID_OC || { echo "❌ opencode falló"; exit 1; }
echo "✅ opencode"
wait \$PID_EN || { echo "❌ engram falló"; exit 1; }
echo "✅ engram"
wait \$PID_GA || { echo "❌ gentle-ai falló"; exit 1; }
echo "✅ gentle-ai"
wait \$PID_OL || { echo "❌ ollama falló"; exit 1; }
echo "✅ ollama"

echo "⚡ [4/4] agent-teams-lite..."
ollama serve > /tmp/ollama-build.log 2>&1 &
OLLAMA_PID=\$!

ELAPSED=0
until curl -s http://localhost:11434/ > /dev/null; do
    sleep 1
    ((ELAPSED++))
    [ \$ELAPSED -ge 60 ] && { echo "❌ Ollama no arrancó en 60s"; exit 1; }
done
echo "✅ Ollama arrancado"

git clone https://github.com/Gentleman-Programming/agent-teams-lite.git /tmp/agent-teams
cd /tmp/agent-teams && ./scripts/setup.sh --all && echo "✅ agent-teams-lite"
kill \$OLLAMA_PID 2>/dev/null || true

echo "⚡ Copiando binarios a /usr/local/bin..."
[ -f "\$HOME/go/bin/engram" ] && sudo cp -f "\$HOME/go/bin/engram" /usr/local/bin/

echo "🧹 Limpiando caché interna..."
sudo pacman -Scc --noconfirm
chmod -R +w ~/.cache/go ~/.cache 2>/dev/null || true
sudo rm -rf /tmp/* ~/.cache/go /var/cache/pacman/pkg 2>/dev/null || true

echo "✅ Build completo."
SCRIPT

    # Card 2: sin prefijo AXIOM_GIT_TOKEN= — el token lo lee el script desde /run/axiom/env. / No AXIOM_GIT_TOKEN= prefix — script reads token from /run/axiom/env.
    distrobox-enter -n "$AXIOM_BUILD_CONTAINER" -- bash "$BUILD_SCRIPT"

    # Card 4: limpiar el script temporal tras ejecutar. / clean up temp script after execution.
    rm -f "$BUILD_SCRIPT"

    echo "📦 Exportando imagen $IMAGEN (esto puede tardar)... / Exporting image $IMAGEN (this may take a while)..."
    podman commit "$AXIOM_BUILD_CONTAINER" "$IMAGEN"

    echo "🧹 Limpieza final... / Final cleanup..."
    distrobox-rm "$AXIOM_BUILD_CONTAINER" --force --yes
    chmod -R +w "$BASE_ENV/$AXIOM_BUILD_CONTAINER" 2>/dev/null
    rm -rf "$BASE_ENV/$AXIOM_BUILD_CONTAINER"

    echo ""
    echo "✅ Imagen $IMAGEN lista. Usa: axiom create [nombre]"
}

create() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: axiom create [nombre]"; return 1; fi
    local NOMBRE="$1"
    local R_PROYECTO="$BASE_DEV/$NOMBRE"
    local R_ENTORNO="$BASE_ENV/$NOMBRE"

    echo "🛡️  Acceso al Búnker '$NOMBRE'..."
    if ! sudo -v; then echo "❌ Acceso denegado."; return 1; fi

    # Si ya existe, entrar directamente
    if distrobox-list --no-color | grep -qw "$NOMBRE"; then
        if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
            ssh-add -l &>/dev/null || ssh-add ~/.ssh/id_ed25519 2>/dev/null
        fi
        distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
        return 0
    fi

    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)

    if ! podman image exists "$IMAGEN"; then
        echo ""
        echo "⚠️  No se encontró la imagen base $IMAGEN."
        echo "    Ejecuta: axiom build"
        return 1
    fi

    echo "⚡ Creando búnker '$NOMBRE' desde $IMAGEN..."
    mkdir -p "$R_PROYECTO" "$R_ENTORNO" "$AI_CONFIG/models"
    _init_tutor

    local GPU_VOLS=""
    [ "$AXIOM_ROCM_MODE" = "host" ] && GPU_VOLS=$(_gpu_volumes_host)

    local SSH_VOL=""
    if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ] && [ -n "${SSH_AUTH_SOCK:-}" ] && [ -S "$SSH_AUTH_SOCK" ]; then
        SSH_VOL="--volume $SSH_AUTH_SOCK:$SSH_AUTH_SOCK"
        echo "🔑 Socket SSH detectado y montado."
    fi

    # Card 1: .env montado read-only para lectura on-demand del token en _git_run.
    distrobox-create --name "$NOMBRE" \
        --image "$IMAGEN" \
        --home "$R_ENTORNO" \
        --additional-flags "--volume $R_PROYECTO:/$NOMBRE \
        --volume $AI_CONFIG:/ai_config \
        --volume \"$DIR/.env\":/run/axiom/env:ro \
        --device /dev/kfd --device /dev/dri \
        --security-opt label=disable --group-add video --group-add render \
        $GPU_VOLS $SSH_VOL" \
        --yes

    distrobox-enter -n "$NOMBRE" -- bash -c "
        sudo pacman -Syu --noconfirm --needed 2>/dev/null | tail -3
        echo '✅ Sistema actualizado.'
    "

    _escribir_bashrc "$NOMBRE" "$R_ENTORNO"
    _escribir_starship "$R_ENTORNO"

    mkdir -p "$R_ENTORNO/.config/opencode"
    [ -f "$TUTOR_PATH" ] && cp "$TUTOR_PATH" "$R_ENTORNO/.config/opencode/AGENTS.md"

    _escribir_opencode_config "$NOMBRE" "$R_ENTORNO"

    if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
        ssh-add -l &>/dev/null || ssh-add ~/.ssh/id_ed25519 2>/dev/null
    fi
    distrobox-enter "$NOMBRE" -- bash --rcfile "$R_ENTORNO/.bashrc" -i
}

delete() {
    mostrar_logo
    if [ -z "${1:-}" ]; then echo "❌ Uso: axiom delete [nombre]"; return 1; fi

    read -rp "📝 Razón técnica obligatoria: " REASON
    [ -z "$REASON" ] && echo "❌ Cancelado: se requiere justificación." && return 1
    echo "- Borrado búnker $1 (Razón: $REASON)" >> "$TUTOR_PATH"

    read -rp "❗ ¿Borrar búnker '$1'? (s/N): " CONFIRM
    if [[ "$CONFIRM" =~ ^[sSyY]$ ]]; then
        distrobox-rm "$1" --force --yes
        if [ -d "$BASE_ENV/$1" ]; then
            chmod -R +w "$BASE_ENV/$1"
            rm -rf "$BASE_ENV/$1"
        fi
        echo "🔥 Búnker '$1' eliminado."
    else
        echo "❌ Cancelado."
    fi
}

reset() {
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

    echo "🔎 Escaneando búnkeres en el sistema..."
    local LISTA_BUNKERES
    if distrobox list --format json &>/dev/null; then
        LISTA_BUNKERES=$(distrobox list --format json | jq -r '.[].name')
    else
        LISTA_BUNKERES=$(distrobox-list --no-color | awk -F'|' 'NR>1 {gsub(/[[:space:]]/, "", $2); if($2!="") print $2}')
    fi
    echo ""

    if [ -n "$LISTA_BUNKERES" ]; then
        echo "📂 Búnkeres encontrados:"
        echo "------------------------------------------------"
        echo "$LISTA_BUNKERES" | sed 's/^/  • /'
        echo "------------------------------------------------"
        echo ""
        read -rp "⚠️  ¿Borrar TODOS estos búnkeres y sus entornos? (s/N): " BORRAR_TODO
    else
        echo "ℹ️  No se detectaron búnkeres creados."
        BORRAR_TODO="n"
    fi
    echo ""

    if [[ "$BORRAR_TODO" =~ ^[sSyY]$ ]]; then
        read -rp "📝 Razón técnica para el reset total: " REASON
        if [ -z "$REASON" ]; then
            echo "❌ Operación cancelada: se requiere justificación técnica."
            return 1
        fi

        echo "- Reset global ejecutado (Razón: $REASON)" >> "$TUTOR_PATH"

        echo "🔥 Iniciando limpieza profunda..."
        for CAJA in $LISTA_BUNKERES; do
            echo "  🗑️  Eliminando $CAJA..."
            distrobox-rm "$CAJA" --force --yes 2>/dev/null
            if [ -d "$BASE_ENV/$CAJA" ]; then
                chmod -R +w "$BASE_ENV/$CAJA" 2>/dev/null
                rm -rf "$BASE_ENV/$CAJA"
            fi
        done
        echo "✅ Todos los búnkeres eliminados."
    fi

    echo ""
    echo "🗑️  Eliminando imagen base de Podman..."
    if podman rmi "$IMAGEN" --force 2>/dev/null; then
        echo "✅ Imagen $IMAGEN eliminada."
    else
        echo "⚠️  No se encontró la imagen o ya fue eliminada."
    fi

    echo ""
    echo "✨ Sistema limpio. Usa 'axiom build' para generar una base nueva."
}

rebuild() {
    mostrar_logo
    detect_gpu
    local IMAGEN
    IMAGEN=$(_imagen_base)

    echo "🔄 Reconstruyendo imagen base $IMAGEN..."
    echo "    Los búnkeres existentes NO se ven afectados."
    echo ""

    if podman image exists "$IMAGEN"; then
        echo "    Tamaño actual: $(podman images --format '{{.Size}}' $IMAGEN)"
    else
        echo "    (La imagen no existe actualmente)"
    fi
    echo ""

    read -rp "¿Continuar? (s/N): " CONFIRM
    [[ "$CONFIRM" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }

    echo "🗑️  Eliminando imagen anterior..."
    podman rmi "$IMAGEN" --force 2>/dev/null || true

    build
}

prune() {
    mostrar_logo
    if [ ! -d "$BASE_ENV" ]; then
        echo "ℹ️  No existe el directorio de entornos: $BASE_ENV"
        return 0
    fi

    local NOMBRES_DISTROBOX
    NOMBRES_DISTROBOX=$(_bunker_lista_nombres)

    local HUERFANOS=()
    while IFS= read -r -d '' DIR; do
        local NOMBRE
        NOMBRE=$(basename "$DIR")
        [ "$NOMBRE" = "axiom-build" ] && continue
        if ! echo "$NOMBRES_DISTROBOX" | grep -qx "$NOMBRE"; then
            HUERFANOS+=("$NOMBRE")
        fi
    done < <(find "$BASE_ENV" -mindepth 1 -maxdepth 1 -type d -print0)

    if [ ${#HUERFANOS[@]} -eq 0 ]; then
        echo "✅ No hay entornos huérfanos. Todo limpio."
        return 0
    fi

    echo ""
    echo "🔍 Entornos huérfanos encontrados:"
    for H in "${HUERFANOS[@]}"; do
        local TAM
        TAM=$(_bunker_tamanio "$H")
        printf "  • %-25s %s\n" "$H" "$TAM"
    done
    echo ""

    read -rp "🗑️  ¿Eliminar todos estos entornos huérfanos? (s/N): " OK
    [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }

    read -rp "📝 Razón técnica: " REASON
    [ -z "$REASON" ] && echo "❌ Se requiere una justificación." && return 1

    for H in "${HUERFANOS[@]}"; do
        echo "  🗑️  Eliminando: $H"
        chmod -R +w "$BASE_ENV/$H" 2>/dev/null
        rm -rf "$BASE_ENV/$H"
    done
    echo "✅ Entornos huérfanos eliminados."
}