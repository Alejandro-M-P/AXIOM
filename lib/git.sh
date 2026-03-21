#!/bin/bash
# ─── MÓDULO GIT: OPERACIONES SEGURAS ────────────────

_git_auth_cmd() {
    echo "!f() { echo \"username=${AXIOM_GIT_USER}\"; echo \"password=${AXIOM_GIT_TOKEN}\"; }; f"
}

push() {
    [ -z "$AXIOM_GIT_TOKEN" ] && echo "❌ AXIOM_GIT_TOKEN no encontrado." && return 1
    git config user.name "$AXIOM_GIT_USER"
    git config user.email "$AXIOM_GIT_EMAIL"

    local RAMA=$(git branch --show-current)
    echo "🚀 Push en progreso ($RAMA)..."
    
    # 🔥 EL FIX: Llamada directa con -c, evitamos el eval
    git -c "credential.helper=$(_git_auth_cmd)" push "$@"
}

commit() {
    local MENSAJE="${1:-}"
    echo -e "\n📋 Cambios pendientes:"
    git status --short
    echo ""
    if [ -z "$MENSAJE" ]; then
        read -rp "📝 Mensaje del commit: " MENSAJE
        [ -z "$MENSAJE" ] && echo "❌ Se requiere un mensaje." && return 1
    fi
    git config user.name "$AXIOM_GIT_USER"
    git config user.email "$AXIOM_GIT_EMAIL"
    git add -A
    git commit -m "$MENSAJE"
    echo ""
    read -rp "🚀 ¿Hacer push ahora? (s/N): " DOPUSH
    [[ "$DOPUSH" =~ ^[sS]$ ]] && push
}

git-clone() {
    if [ -z "${1:-}" ]; then echo "❌ Uso: git-clone [usuario/repo] [carpeta]"; return 1; fi
    [ -z "$AXIOM_GIT_TOKEN" ] && echo "❌ No se encontró AXIOM_GIT_TOKEN" && return 1
    local REPO="$1" DIR="${2:-$(basename "$1")}"

    # También corregimos el clone para que sea seguro
    git -c "credential.helper=$(_git_auth_cmd)" clone "https://github.com/${REPO}.git" "$DIR"
    echo "✅ Repo clonado."
}

rama() {
    echo -e "\n🌿 Ramas disponibles:"
    git branch -a 2>/dev/null | sed 's/^/  /'
    echo ""

    read -rp "📝 Nombre de la nueva rama: " NOMBRE_RAMA
    [ -z "$NOMBRE_RAMA" ] && echo "❌ Se requiere un nombre." && return 1

    local RAMA_BASE=$(git branch --show-current)
    read -rp "Rama base (Enter para heredar de $RAMA_BASE): " R_BASE
    R_BASE="${R_BASE:-$RAMA_BASE}"

    git checkout "$R_BASE" 2>/dev/null || { echo "❌ Rama '$R_BASE' no existe."; return 1; }
    git checkout -b "$NOMBRE_RAMA"
    echo -e "✅ Rama '$NOMBRE_RAMA' creada desde '$R_BASE'.\n"
}