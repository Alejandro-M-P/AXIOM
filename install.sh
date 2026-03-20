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

# ─── GPU (auto-detectado, solo sobreescribir si falla) ───
AXIOM_GPU_TYPE=""
AXIOM_GFX_VAL="$GFX_VERSION"
EOF

echo ""
echo "✅ ¡Configuración guardada en .env exitosamente!"
echo ""
echo "🔥 SIGUIENTE PASO 🔥"
echo "Añade la siguiente línea al final de tu ~/.bashrc:"
echo "  source $DIR/axiom.sh"
echo ""
echo "Luego abre una terminal nueva y usa 'crear mi-primer-proyecto' para empezar."