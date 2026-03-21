#!/bin/bash
set -e # Abortar si hay errores

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Bienvenido al Instalador de AXIOM"
echo "------------------------------------"

if [ -f "$DIR/.env" ]; then
    echo "⚠️  Ya existe un archivo .env en $DIR."
    echo "Si deseas reconfigurar, bórralo primero e inténtalo de nuevo."
    exit 1
fi

# ─── 1. CAPTURA DE DATOS ───────────────────────────
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
echo "   1. Montar desde el host  (recomendado — Bazzite, Fedora, Nobara...)"
echo "   2. Instalar dentro de la imagen  (máxima portabilidad)"
echo ""
read -rp "   Selecciona [1/2] (Enter para 1): " ROCM_OPT
ROCM_OPT=${ROCM_OPT:-1}

case "$ROCM_OPT" in
    2) ROCM_MODE="image" ;;
    *) ROCM_MODE="host"  ;;
esac

# ─── 2. GENERACIÓN DE .ENV ─────────────────────────
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
AXIOM_ROCM_MODE="$ROCM_MODE"
AXIOM_GPU_TYPE=""
AXIOM_GFX_VAL="$GFX_VERSION"
EOF

# ─── 3. PREPARACIÓN DE ESTRUCTURA ──────────────────
echo ""
echo "📂 Preparando estructura de archivos..."
mkdir -p "$DIR/lib"
# Creamos los búnkeres y búnker de entorno
mkdir -p "$BASE_DIR"/{ai_global/teams,ai_config/models,.entorno}

echo "🔐 Asegurando permisos de ejecución..."
chmod +x "$DIR/AXIOM.sh"
[ -d "$DIR/lib" ] && chmod +x "$DIR/lib/"*.sh 2>/dev/null || true

# ─── 4. CREACIÓN DEL COMANDO GLOBAL ───────────────
BIN_PATH="$HOME/.local/bin"
mkdir -p "$BIN_PATH"

echo "🛠️  Creando acceso directo 'axiom' en $BIN_PATH..."
cat > "$BIN_PATH/axiom" << EOF
#!/bin/bash
# Wrapper para ejecutar AXIOM desde cualquier lugar
export AXIOM_PATH="$DIR"
bash "\$AXIOM_PATH/axiom.sh" "\$@"
EOF

chmod +x "$BIN_PATH/axiom"

# ─── 5. FINALIZACIÓN ──────────────────────────────
echo ""
echo "✅ ¡Instalación completada con éxito!"
echo "------------------------------------"
echo "🚀 Ahora puedes usar el comando 'axiom' en cualquier terminal."
echo ""
echo "⚠️  NOTA: Si 'axiom' no se reconoce, añade esto a tu ~/.bashrc:"
echo "   export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "🔥 PRIMER PASO: axiom build"