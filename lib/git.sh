

#!/bin/bash
# ─── MÓDULO GIT: OPERACIONES SEGURAS ────────────────


_git_auth_cmd() {
    echo "!printf 'username=$AXIOM_GIT_USER\npassword=$AXIOM_GIT_TOKEN\n'"
}

push() {
    # Aseguramos que el remoto tenga el token inyectado
    git remote set-url origin "https://${AXIOM_GIT_USER}:${AXIOM_GIT_TOKEN}@github.com/Alejandro-M-P/AXIOM.git"
    
    echo "🚀 Push en progreso..."
    git push origin "$(git branch --show-current)"
}



commit() {
    local MENSAJE="${1:-}"
    echo -e "\n📋 Cambios pendientes / Pending changes:"
    git status --short
    echo ""
    if [ -z "$MENSAJE" ]; then
        read -rp "📝 Mensaje del commit / Commit message: " MENSAJE
        [ -z "$MENSAJE" ] && echo "❌ Se requiere un mensaje. / A message is required." && return 1
    fi
    git config user.name "$AXIOM_GIT_USER"
    git config user.email "$AXIOM_GIT_EMAIL"
    git add -A
    git commit -m "$MENSAJE"
    echo ""
    read -rp "🚀 ¿Hacer push ahora? / Push now? (s/N/y/N): " DOPUSH
    [[ "$DOPUSH" =~ ^[sSyY]$ ]] && push
}

git-clone() {
    if [ -z "${1:-}" ]; then echo "❌ Uso/Usage: git-clone [usuario/repo] [carpeta/folder]"; return 1; fi
    [ -z "$AXIOM_GIT_TOKEN" ] && echo "❌ No se encontró / Not found AXIOM_GIT_TOKEN" && return 1
    local REPO="$1" DIR="${2:-$(basename "$1")}"

    # También corregimos el clone para que sea seguro
    git clone "https://${AXIOM_GIT_USER}:${AXIOM_GIT_TOKEN}@github.com/${REPO}.git" "$DIR"
}

branch() {
    echo -e "\n🌿 Ramas disponibles / Available branches:"
    git branch -a 2>/dev/null | sed 's/^/  /'
    echo ""

    read -rp "📝 Nombre de la nueva rama / New branch name: " NOMBRE_RAMA
    [ -z "$NOMBRE_RAMA" ] && echo "❌ Se requiere un nombre. / A name is required." && return 1

    local RAMA_BASE=$(git branch --show-current)
    read -rp "Rama base / Base branch (Enter para heredar de / to inherit from $RAMA_BASE): " R_BASE
    R_BASE="${R_BASE:-$RAMA_BASE}"

    git checkout "$R_BASE" 2>/dev/null || { echo "❌ Rama '$R_BASE' no existe. / Branch '$R_BASE' does not exist."; return 1; }
    git checkout -b "$NOMBRE_RAMA"
    echo -e "✅ Rama '$NOMBRE_RAMA' creada desde / created from '$R_BASE'.\n"
}