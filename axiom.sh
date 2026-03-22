#!/bin/bash
# Usamos la variable que definimos en el wrapper o detectamos la ruta
DIR="${AXIOM_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"

# 1. Carga de entorno
[ -f "$DIR/.env" ] && source "$DIR/.env"

# 2. Exportar rutas para que los módulos las vean
export BASE_DEV="${AXIOM_BASE_DIR}"
export BASE_ENV="$BASE_DEV/.entorno"
export AI_CONFIG="$BASE_DEV/ai_config"
export TUTOR_PATH="$AI_CONFIG/teams/tutor.md"
# 3. CARGA DE MÓDULOS
for modulo in gpu core git bunker; do
    [ -f "$DIR/lib/$modulo.sh" ] && source "$DIR/lib/$modulo.sh"
done

# 4. LÓGICA DE EJECUCIÓN
# Solo ejecutamos comandos o mostramos ayuda si el script se está EJECUTANDO,
# no cuando le hacemos "source" desde el .bashrc
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [ $# -gt 0 ]; then
        COMMAND="$1"
        shift
        if declare -f "$COMMAND" > /dev/null; then
            "$COMMAND" "$@"
        else
            echo "❌ Comando '$COMMAND' no reconocido. / Command '$COMMAND' not recognized."
            help
        fi
    else
        help
    fi
fi