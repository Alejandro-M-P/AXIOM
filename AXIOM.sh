#!/bin/bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. Carga de entorno
[ -f "$DIR/.env" ] && source "$DIR/.env"

# 2. Rutas globales
export BASE_DEV="${AXIOM_BASE_DIR}"
export BASE_ENV="$BASE_DEV/.entorno"
export AI_GLOBAL="$BASE_DEV/ai_global"
export AI_CONFIG="$BASE_DEV/ai_config"
export TUTOR_PATH="$AI_GLOBAL/teams/tutor.md"

# 3. CARGA DE MÓDULOS

for modulo in "$DIR/lib/"*.sh; do
    [ -f "$modulo" ] && source "$modulo"
done