#!/bin/bash
# ─── MÓDULO BUNKER: LECTURA Y ESTADO ──────────

_bunker_estado() {
    local NOMBRE="$1"
    if podman ps --format '{{.Names}}' 2>/dev/null | grep -qx "$NOMBRE"; then
        echo "running"
    else
        echo "stopped"
    fi
}

_bunker_tamanio() {
    local NOMBRE="$1"
    local RUTA="$BASE_ENV/$NOMBRE"
    if [ -d "$RUTA" ]; then
        du -sh "$RUTA" 2>/dev/null | awk '{print $1}'
    else
        echo "—"
    fi
}

_bunker_rama_git() {
    local NOMBRE="$1"
    local RUTA="$BASE_DEV/$NOMBRE"
    if [ -d "$RUTA/.git" ]; then
        git -C "$RUTA" branch --show-current 2>/dev/null || echo "—"
    else
        echo "—"
    fi
}

_bunker_ultima_entrada() {
    local NOMBRE="$1"
    if [ -d "$BASE_ENV/$NOMBRE" ]; then
        stat -c '%y' "$BASE_ENV/$NOMBRE" 2>/dev/null | cut -d'.' -f1 | cut -d' ' -f1
    else
        echo "—"
    fi
}

_bunker_lista_nombres() {
    distrobox-list --no-color 2>/dev/null \
        | awk -F'|' 'NR>1 {gsub(/[[:space:]]/, "", $2); if($2!="") print $2}'
}

list() {
    mostrar_logo
    local NOMBRES
    NOMBRES=$(_bunker_lista_nombres)

    if [ -z "$NOMBRES" ]; then
        echo "ℹ️  No hay búnkeres creados. Usa: axiom create [nombre]"
        return 0
    fi

    echo ""
    printf "  \033[1;34m%-22s  %-10s  %-8s  %-12s  %-12s\033[0m\n" \
        "BÚNKER" "ESTADO" "TAMAÑO" "ÚLTIMA ENTRADA" "RAMA GIT"
    echo "  ──────────────────────────────────────────────────────────────────"

    while IFS= read -r NOMBRE; do
        local ESTADO TAMANIO RAMA FECHA COLOR_ESTADO
        ESTADO=$(_bunker_estado "$NOMBRE")
        TAMANIO=$(_bunker_tamanio "$NOMBRE")
        RAMA=$(_bunker_rama_git "$NOMBRE")
        FECHA=$(_bunker_ultima_entrada "$NOMBRE")

        if [ "$ESTADO" = "running" ]; then
            COLOR_ESTADO="\033[1;32m● running \033[0m"
        else
            COLOR_ESTADO="\033[0;90m○ stopped\033[0m"
        fi

        printf "  \033[1m%-22s\033[0m  %b  %-8s  %-14s  \033[0;36m%-12s\033[0m\n" \
            "$NOMBRE" "$COLOR_ESTADO" "$TAMANIO" "$FECHA" "$RAMA"
    done <<< "$NOMBRES"

    echo ""
    echo "  Total: $(echo "$NOMBRES" | wc -l) búnker(es)"
    echo ""
}

stop() {
    if [ -z "${1:-}" ]; then
        local CORRIENDO
        CORRIENDO=$(podman ps --format '{{.Names}}' 2>/dev/null)
        if [ -z "$CORRIENDO" ]; then
            echo "ℹ️  No hay búnkeres corriendo."
            return 0
        fi
        echo ""
        echo "🟢 Búnkeres activos:"
        echo "$CORRIENDO" | sed 's/^/  • /'
        echo ""
        read -rp "📝 Nombre del búnker a parar (Enter para cancelar): " NOMBRE
        [ -z "$NOMBRE" ] && echo "❌ Cancelado." && return 0
    else
        NOMBRE="$1"
    fi

    if ! _bunker_lista_nombres | grep -qx "$NOMBRE"; then
        echo "❌ El búnker '$NOMBRE' no existe."
        return 1
    fi

    if [ "$(_bunker_estado "$NOMBRE")" = "stopped" ]; then
        echo "ℹ️  El búnker '$NOMBRE' ya está parado."
        return 0
    fi

    echo "⏹️  Parando '$NOMBRE'..."
    distrobox-stop "$NOMBRE" --yes
    echo "✅ Búnker '$NOMBRE' parado."
}

info() {
    if [ -z "${1:-}" ]; then
        local NOMBRES
        NOMBRES=$(_bunker_lista_nombres)
        if [ -z "$NOMBRES" ]; then
            echo "ℹ️  No hay búnkeres creados."
            return 0
        fi
        echo ""
        echo "🛡️  Búnkeres disponibles:"
        echo "$NOMBRES" | sed 's/^/  • /'
        echo ""
        read -rp "📝 Nombre del búnker: " NOMBRE
        [ -z "$NOMBRE" ] && echo "❌ Cancelado." && return 0
    else
        NOMBRE="$1"
    fi

    if ! _bunker_lista_nombres | grep -qx "$NOMBRE"; then
        echo "❌ El búnker '$NOMBRE' no existe."
        return 1
    fi

    local ESTADO TAMANIO RAMA FECHA
    ESTADO=$(_bunker_estado "$NOMBRE")
    TAMANIO=$(_bunker_tamanio "$NOMBRE")
    RAMA=$(_bunker_rama_git "$NOMBRE")
    FECHA=$(_bunker_ultima_entrada "$NOMBRE")

    local R_PROYECTO="$BASE_DEV/$NOMBRE"
    local R_ENTORNO="$BASE_ENV/$NOMBRE"

    echo ""
    echo "  ┌─────────────────────────────────────────┐"
    printf "  │  \033[1m%-39s\033[0m│\n" "$NOMBRE"
    echo "  ├─────────────────────────────────────────┤"
    if [ "$ESTADO" = "running" ]; then
        printf "  │  Estado          \033[1;32m%-23s\033[0m│\n" "● running"
    else
        printf "  │  Estado          \033[0;90m%-23s\033[0m│\n" "○ stopped"
    fi
    printf "  │  Tamaño entorno  %-23s│\n" "$TAMANIO"
    printf "  │  Última entrada  %-23s│\n" "$FECHA"
    printf "  │  Rama git        \033[0;36m%-23s\033[0m│\n" "$RAMA"
    echo "  ├─────────────────────────────────────────┤"
    printf "  │  Proyecto        %-23s│\n" "${R_PROYECTO/#$HOME/\~}"
    printf "  │  Entorno         %-23s│\n" "${R_ENTORNO/#$HOME/\~}"
    local IMAGEN
    IMAGEN=$(podman inspect "$NOMBRE" --format '{{.ImageName}}' 2>/dev/null || echo "—")
    printf "  │  Imagen base     %-23s│\n" "$IMAGEN"
    if [ -d "$R_PROYECTO" ]; then
        local TAM_PROYECTO
        TAM_PROYECTO=$(du -sh "$R_PROYECTO" 2>/dev/null | awk '{print $1}')
        printf "  │  Tamaño proyecto %-23s│\n" "$TAM_PROYECTO"
    fi
    echo "  └─────────────────────────────────────────┘"
    echo ""
}