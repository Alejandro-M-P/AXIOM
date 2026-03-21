#!/bin/bash
# Usamos la variable que definimos en el wrapper o detectamos la ruta
DIR="${AXIOM_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"

# 1. Carga de entorno
[ -f "$DIR/.env" ] && source "$DIR/.env"

# 2. Exportar rutas para que los módulos las vean
export BASE_DEV="${AXIOM_BASE_DIR}"
export BASE_ENV="$BASE_DEV/.entorno"
export AI_GLOBAL="$BASE_DEV/ai_global"
export AI_CONFIG="$BASE_DEV/ai_config"
export TUTOR_PATH="$AI_GLOBAL/teams/tutor.md"

# 3. CARGA DE MÓDULOS
for modulo in "$DIR/lib/"*.sh; do
    [ -f "$modulo" ] && source "$modulo"
done

# 4. LÓGICA DE EJECUCIÓN
# Si se pasa un argumento (ej: axiom build), lo ejecuta.
# Si no, muestra la ayuda.
if [ $# -gt 0 ]; then
    COMMAND="$1"
    shift
    if declare -f "$COMMAND" > /dev/null; then
        "$COMMAND" "$@"
    else
        echo "❌ Comando '$COMMAND' no reconocido."
        ayuda
    fi
else
    ayuda
fi