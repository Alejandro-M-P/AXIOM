#!/bin/bash
set -e # Abortar si hay errores

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Bienvenido al Instalador de AXIOM / Welcome to the AXIOM Installer"
echo "------------------------------------"

if [ -f "$DIR/.env" ]; then
    echo "⚠️  Ya existe un archivo .env en $DIR. / A .env file already exists in $DIR."
    echo "Si deseas reconfigurar, bórralo primero e inténtalo de nuevo. / If you want to reconfigure, delete it first and try again."
    exit 1
fi

# ─── 1. CAPTURA DE DATOS ───────────────────────────
GIT_USER=""
while [ -z "$GIT_USER" ]; do read -rp "1️⃣  Usuario de GitHub / GitHub Username: " GIT_USER; done

GIT_EMAIL=""
while [ -z "$GIT_EMAIL" ]; do read -rp "2️⃣  Email de GitHub / GitHub Email: " GIT_EMAIL; done

GIT_TOKEN=""
while [ -z "$GIT_TOKEN" ]; do read -rp "3️⃣  Token de GitHub / GitHub Token (Classic/Fine-grained): " GIT_TOKEN; done

read -rp "4️⃣  Directorio Base / Base Directory (Enter para/for $HOME/dev): " BASE_DIR
BASE_DIR=${BASE_DIR:-$HOME/dev}

read -rp "5️⃣  Directorio Modelos Ollama / Ollama Models Directory (Enter para/for $BASE_DIR/ai_config/models): " MODELS_DIR
MODELS_DIR=${MODELS_DIR:-$BASE_DIR/ai_config/models}

read -rp "6️⃣  (Opcional/Optional) Forzar/Force GFX_VERSION para/for AMD (Enter para autodetectar / for autodetection): " GFX_VERSION

echo ""
echo "7️⃣  Modo de drivers GPU / GPU drivers mode:"
echo "   1. Montar desde el host / Mount from host (recomendado/recommended — Bazzite, Fedora, Nobara...)"
echo "   2. Instalar dentro de la imagen / Install inside the image (máxima portabilidad / max portability)"
echo ""
read -rp "   Selecciona/Select [1/2] (Enter para/for 1): " ROCM_OPT
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
echo "📂 Preparando estructura de archivos... / Preparing file structure..."
mkdir -p "$DIR/lib"
# Creamos la jerarquía de búnkeres
mkdir -p "$BASE_DIR"/{ai_global/teams,ai_config/models,.entorno}

echo "🔐 Asegurando permisos de configuración... / Securing config permissions..."
# Permisos 600: Solo tu usuario puede leer o escribir tus tokens
chmod 600 "$DIR/.env"
# Aseguramos que el script principal (ahora en minúsculas) sea ejecutable
chmod +x "$DIR/axiom.sh"
[ -d "$DIR/lib" ] && chmod +x "$DIR/lib/"*.sh 2>/dev/null || true

# ─── 4. CREACIÓN DEL COMANDO GLOBAL ───────────────
BIN_PATH="$HOME/.local/bin"
mkdir -p "$BIN_PATH"

echo "🛠️  Creando acceso directo 'axiom' en $BIN_PATH... / Creating 'axiom' shortcut in $BIN_PATH..."
cat > "$BIN_PATH/axiom" << EOF
#!/bin/bash
# Wrapper profesional para AXIOM
export AXIOM_PATH="$DIR"
bash "\$AXIOM_PATH/axiom.sh" "\$@"
EOF

chmod +x "$BIN_PATH/axiom"
# ─── 5. FINALIZACIÓN ──────────────────────────────
echo ""
echo "✅ ¡Instalación completada con éxito! / Installation completed successfully!"
echo "------------------------------------"
echo "🚀 Ahora puedes usar el comando 'axiom' en cualquier terminal. / You can now use the 'axiom' command in any terminal."
echo ""
echo "⚠️  NOTA/NOTE: Si 'axiom' no se reconoce / If 'axiom' is not recognized, añade esto a tu / add this to your ~/.bashrc:"
echo "   export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "🔥 PRIMER PASO / FIRST STEP: axiom build"