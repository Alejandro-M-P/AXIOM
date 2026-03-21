#!/bin/bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Bienvenido al Instalador de AXIOM"
echo "------------------------------------"

if [ -f "$DIR/.env" ]; then
    echo "⚠️  Ya existe un archivo .env en $DIR."
    echo "Si deseas reconfigurar, bórralo primero e inténtalo de nuevo."
    exit 1
fi

GIT_USER=""
while [ -z "$GIT_USER" ]; do read -rp "1️⃣  Usuario de GitHub: " GIT_USER; done

GIT_EMAIL=""
while [ -z "$GIT_EMAIL" ]; do read -rp "2️⃣  Email de GitHub: " GIT_EMAIL; done

GIT_TOKEN=""
while [ -z "$GIT_TOKEN" ]; do read -rp "3️⃣  Token de GitHub (Classic/Fine-grained): " GIT_TOKEN; done

read -rp "4️⃣  Directorio Base (Enter para $HOME/dev): " BASE_DIR
BASE_DIR=${BASE_DIR:-$HOME/dev}

read -rp "5️⃣  Directorio Modelos Ollama (Enter para $BASE_DIR/ai_config/models): " MODELS_DIR
MODELS_DIR=${MODELS_DIR:-$BASE_DIR/ai_config/models}

read -rp "6️⃣  (Opcional) Forzar GFX_VERSION para AMD (Enter para autodetectar): " GFX_VERSION

echo ""
echo "7️⃣  Modo de drivers GPU:"
echo "   1. Montar desde el host  (recomendado — Bazzite, Fedora, Nobara, CachyOS...)"
echo "      Imagen resultante: ~10-13 GB | Commit: ~3-4 min"
echo "   2. Instalar dentro de la imagen  (máxima portabilidad, cualquier distro)"
echo "      Imagen resultante: ~38 GB    | Commit: ~15 min"
echo ""
read -rp "   Selecciona [1/2] (Enter para 1): " ROCM_OPT
ROCM_OPT=${ROCM_OPT:-1}

case "$ROCM_OPT" in
    2) ROCM_MODE="image" ;;
    *) ROCM_MODE="host"  ;;
esac

cat > "$DIR/.env" << EOF
# ─── IDENTIDAD GIT ───────────────────────────────
AXIOM_GIT_USER="$GIT_USER"
AXIOM_GIT_EMAIL="$GIT_EMAIL"
AXIOM_GIT_TOKEN="$GIT_TOKEN"

# ─── RUTAS BASE ──────────────────────────────────
AXIOM_BASE_DIR="$BASE_DIR"

# ─── IA / MODELOS ────────────────────────────────
AXIOM_OLLAMA_HOST="http://localhost:11434"
AXIOM_MODELS_DIR="$MODELS_DIR"

# ─── GPU ─────────────────────────────────────────
# host  → monta ROCm/CUDA del host (recomendado en Bazzite, Fedora, Nobara...)
# image → instala ROCm/CUDA dentro de la imagen (portabilidad total)
AXIOM_ROCM_MODE="$ROCM_MODE"
AXIOM_GPU_TYPE=""
AXIOM_GFX_VAL="$GFX_VERSION"
EOF

echo ""
echo "✅ ¡Configuración guardada en .env exitosamente!"
echo ""
echo "   Modo GPU seleccionado: $ROCM_MODE"
echo ""
echo "🔥 SIGUIENTES PASOS 🔥"
echo "1. Añade al final de tu ~/.bashrc:"
echo "     source $DIR/axiom.sh"
echo ""
echo "2. Abre una terminal nueva y construye la imagen base (solo una vez):"
echo "     build"
echo ""
echo "3. Crea tu primer búnker:"
echo "     crear mi-primer-proyecto"