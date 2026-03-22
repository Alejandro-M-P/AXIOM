#!/bin/bash
DIR="${AXIOM_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"

[ -f "$DIR/.env" ] && source "$DIR/.env"

export BASE_DEV="${AXIOM_BASE_DIR}"
export BASE_ENV="$BASE_DEV/.entorno"
export AI_CONFIG="$BASE_DEV/ai_config"
export TUTOR_PATH="$AI_CONFIG/teams/tutor.md"

for modulo in gpu core bunkers git bunker; do
    [ -f "$DIR/lib/$modulo.sh" ] && source "$DIR/lib/$modulo.sh"
done

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [ $# -gt 0 ]; then
        COMMAND="$1"
        shift
        if declare -f "$COMMAND" > /dev/null; then
            "$COMMAND" "$@"
        else
            echo "❌ Comando '$COMMAND' no reconocido."
            help
        fi
    else
        help
    fi
fi