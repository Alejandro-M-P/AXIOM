#!/bin/bash
# ─── MÓDULO GIT: OPERACIONES SEGURAS E INTERACTIVAS ────────────────

_git_check() {
    git rev-parse --git-dir &>/dev/null || { echo "❌ No estás en un repositorio git."; return 1; }
}

_git_configure() {
    git config user.name  "$AXIOM_GIT_USER"
    git config user.email "$AXIOM_GIT_EMAIL"
}

_git_run() {
    if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
        git "$@"
    else
        [ -z "$AXIOM_GIT_TOKEN" ] && { echo "❌ Token vacío."; return 1; }
        local CRED_FILE
        CRED_FILE=$(mktemp)
        chmod 600 "$CRED_FILE"
        printf 'username=%s\npassword=%s\n' "$AXIOM_GIT_USER" "$AXIOM_GIT_TOKEN" > "$CRED_FILE"
        git -c "credential.helper=store --file $CRED_FILE" "$@"
        local EXIT_CODE=$?
        rm -f "$CRED_FILE"
        return $EXIT_CODE
    fi
}

_fzf_check() {
    command -v fzf &>/dev/null || { echo "❌ fzf no encontrado. Ejecuta: sudo pacman -S fzf"; return 1; }
}

_listar_ramas_locales() {
    git branch --format='%(refname:short)'
}

_listar_ramas_todas() {
    {
        git branch --format='%(refname:short)'
        git branch -r --format='%(refname:short)' | sed 's|origin/||' | grep -v HEAD
    } | sort -u
}

_listar_remotes() {
    git remote
}

clone() {
    local REPO="${1:-}"
    if [ -z "$REPO" ]; then
        read -rp "📦 Repositorio (usuario/repo): " REPO
        [ -z "$REPO" ] && echo "❌ Cancelado." && return 1
    fi
    local DIR_DEFAULT
    DIR_DEFAULT=$(basename "$REPO")
    read -rp "📂 Carpeta destino (Enter para '$DIR_DEFAULT'): " DIR
    DIR="${DIR:-$DIR_DEFAULT}"
    echo "⬇️  Clonando $REPO → $DIR ..."
    if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
        git clone "git@github.com:${REPO}.git" "$DIR"
    else
        _git_run clone "https://github.com/${REPO}.git" "$DIR"
    fi
    echo "✅ Repo clonado en ./$DIR"
}

status() {
    _git_check || return 1
    _fzf_check || return 1
    local CAMBIOS
    CAMBIOS=$(git status --short)
    if [ -z "$CAMBIOS" ]; then
        echo "✅ Árbol limpio, nada que mostrar."
        return 0
    fi
    echo "$CAMBIOS" | fzf \
        --prompt="Archivo > " \
        --preview='git diff --color=always {2}' \
        --preview-window=right:60%:wrap \
        --height=80% --border \
        --header="Enter = diff completo en pager" \
        --bind "enter:execute(git diff {2} | less -R)" \
        --ansi
}

branch() {
    _git_check || return 1
    echo ""
    echo "🌿 Ramas disponibles:"
    _listar_ramas_todas | sed 's/^/  • /'
    echo ""
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    read -rp "📌 Rama base (Enter para '$RAMA_ACTUAL'): " BASE
    BASE="${BASE:-$RAMA_ACTUAL}"
    if ! _listar_ramas_todas | grep -qx "$BASE"; then
        echo "❌ La rama '$BASE' no existe."
        return 1
    fi
    read -rp "📝 Nombre de la nueva rama: " NOMBRE
    [ -z "$NOMBRE" ] && echo "❌ Se requiere un nombre." && return 1
    git checkout "$BASE" 2>/dev/null || { echo "❌ No se pudo cambiar a '$BASE'."; return 1; }
    git checkout -b "$NOMBRE"
    echo "✅ Rama '$NOMBRE' creada desde '$BASE'."
}

switch() {
    _git_check || return 1
    _fzf_check || return 1
    echo "🔄 Actualizando ramas remotas..."
    _git_run fetch --all --prune 2>/dev/null || true
    local RAMA
    RAMA=$(_listar_ramas_todas | fzf --prompt="Cambiar a > " --height=40% --border)
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
    if git branch --format='%(refname:short)' | grep -qx "$RAMA"; then
        git checkout "$RAMA"
    else
        git checkout -b "$RAMA" --track "origin/$RAMA"
    fi
}

branch-delete() {
    _git_check || return 1
    echo ""
    echo "🌿 Ramas locales disponibles:"
    _listar_ramas_locales | sed 's/^/  • /'
    echo ""
    read -rp "📝 Nombre de la rama a borrar: " RAMA
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    if [ "$RAMA" = "$RAMA_ACTUAL" ]; then
        echo "❌ No puedes borrar la rama en la que estás ('$RAMA')."
        return 1
    fi
    echo ""
    echo "¿Qué quieres borrar?"
    echo "  1) Solo local"
    echo "  2) Solo remota"
    echo "  3) Ambas"
    echo ""
    read -rp "Opción [1/2/3]: " OPT
    case "$OPT" in
        1)
            read -rp "⚠️  ¿Borrar rama local '$RAMA'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git branch -d "$RAMA" || git branch -D "$RAMA"
            echo "✅ Rama local '$RAMA' eliminada."
            ;;
        2)
            read -rp "⚠️  ¿Borrar rama remota 'origin/$RAMA'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            _git_run push origin --delete "$RAMA"
            echo "✅ Rama remota 'origin/$RAMA' eliminada."
            ;;
        3)
            read -rp "⚠️  ¿Borrar '$RAMA' local Y remota? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git branch -d "$RAMA" || git branch -D "$RAMA"
            _git_run push origin --delete "$RAMA" 2>/dev/null || echo "⚠️  No existía en remoto."
            echo "✅ Rama '$RAMA' eliminada (local + remota)."
            ;;
        *)
            echo "❌ Opción no válida."
            ;;
    esac
}

commit() {
    _git_check || return 1
    echo ""
    echo "📋 Cambios pendientes:"
    git status --short
    echo ""
    if git diff --cached --quiet && git diff --quiet; then
        echo "ℹ️  No hay cambios para commitear."
        return 0
    fi
    echo "¿Qué archivos quieres incluir?"
    echo "  1) Todos"
    echo "  2) Elegir uno a uno"
    echo ""
    read -rp "Opción [1/2] (Enter para 1): " OPT
    OPT="${OPT:-1}"
    _git_configure
    if [ "$OPT" = "2" ]; then
        echo ""
        echo "Archivos con cambios:"
        local ARCHIVOS_DISPONIBLES
        ARCHIVOS_DISPONIBLES=$(git status --short | awk '{print $2}')
        echo "$ARCHIVOS_DISPONIBLES" | nl -w2 -s') '
        echo ""
        read -rp "📝 Números separados por espacio (ej: 1 3): " SELECCION
        [ -z "$SELECCION" ] && echo "❌ Cancelado." && return 1
        for N in $SELECCION; do
            local ARCHIVO
            ARCHIVO=$(echo "$ARCHIVOS_DISPONIBLES" | sed -n "${N}p")
            [ -n "$ARCHIVO" ] && git add "$ARCHIVO"
        done
        echo "✅ Archivos seleccionados añadidos."
    else
        git add -A
        echo "✅ Todos los archivos añadidos."
    fi
    local MENSAJE="${1:-}"
    if [ -z "$MENSAJE" ]; then
        read -rp "📝 Mensaje del commit: " MENSAJE
        [ -z "$MENSAJE" ] && echo "❌ Se requiere un mensaje." && return 1
    fi
    git commit -m "$MENSAJE"
    echo ""
    read -rp "🚀 ¿Hacer push ahora? (s/N): " DOPUSH
    [[ "$DOPUSH" =~ ^[sSyY]$ ]] && push
}

push() {
    _git_check || return 1
    _git_configure
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    echo ""
    echo "📡 Remotes disponibles:"
    _listar_remotes | sed 's/^/  • /'
    echo ""
    read -rp "📡 Remote (Enter para 'origin'): " REMOTE
    REMOTE="${REMOTE:-origin}"
    read -rp "🌿 Rama destino (Enter para '$RAMA_ACTUAL'): " RAMA
    RAMA="${RAMA:-$RAMA_ACTUAL}"
    echo "🚀 Push → $REMOTE/$RAMA ..."
    _git_run push "$REMOTE" "HEAD:$RAMA"
    echo "✅ Push completado."
}

pull() {
    _git_check || return 1
    echo ""
    echo "📡 Remotes disponibles:"
    _listar_remotes | sed 's/^/  • /'
    echo ""
    read -rp "📡 Remote (Enter para 'origin'): " REMOTE
    REMOTE="${REMOTE:-origin}"
    echo "🔄 Actualizando referencias de $REMOTE..."
    _git_run fetch "$REMOTE" --prune 2>/dev/null
    echo ""
    echo "🌿 Ramas remotas disponibles:"
    git branch -r --format='%(refname:short)' | grep "^$REMOTE/" | sed "s|^$REMOTE/||" | grep -v HEAD | sed 's/^/  • /'
    echo ""
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    read -rp "🌿 Rama a hacer pull (Enter para '$RAMA_ACTUAL'): " RAMA
    RAMA="${RAMA:-$RAMA_ACTUAL}"
    echo "⬇️  Pull $REMOTE/$RAMA ..."
    _git_run pull "$REMOTE" "$RAMA"
    echo "✅ Pull completado."
}

merge() {
    _git_check || return 1
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    echo ""
    echo "🌿 Ramas disponibles:"
    _listar_ramas_todas | grep -v "^$RAMA_ACTUAL$" | sed 's/^/  • /'
    echo ""
    read -rp "🔀 Rama a mergear en '$RAMA_ACTUAL': " ORIGEN
    [ -z "$ORIGEN" ] && echo "❌ Cancelado." && return 1
    echo ""
    echo "Estrategia:"
    echo "  1) merge normal (preservar historial)"
    echo "  2) squash (un solo commit)"
    echo "  3) no-ff (siempre commit de merge)"
    echo ""
    read -rp "Opción [1/2/3] (Enter para 1): " OPT
    OPT="${OPT:-1}"
    _git_configure
    case "$OPT" in
        2) git merge --squash "$ORIGEN" && echo "✅ Squash aplicado. Ahora haz commit." ;;
        3) git merge --no-ff "$ORIGEN" ;;
        *) git merge "$ORIGEN" ;;
    esac
}

rebase() {
    _git_check || return 1
    echo ""
    echo "Tipo de rebase:"
    echo "  1) Normal (sobre otra rama)"
    echo "  2) Interactivo -i (editar commits)"
    echo ""
    read -rp "Opción [1/2]: " OPT
    case "$OPT" in
        2)
            echo ""
            echo "Últimos commits:"
            git log --oneline -15 | nl -w2 -s') '
            echo ""
            read -rp "📝 ¿Desde cuántos commits atrás? (ej: 3): " N
            [ -z "$N" ] && echo "❌ Cancelado." && return 1
            git rebase -i "HEAD~$N"
            ;;
        1)
            echo ""
            echo "🌿 Ramas disponibles:"
            _listar_ramas_todas | sed 's/^/  • /'
            echo ""
            read -rp "📌 Rama sobre la que hacer rebase: " RAMA
            [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  Esto reescribe el historial. ¿Continuar? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git rebase "$RAMA"
            ;;
        *) echo "❌ Opción no válida." ;;
    esac
}

log() {
    _git_check || return 1
    _fzf_check || return 1
    git log --oneline --color=always --decorate | \
        fzf --ansi \
            --prompt="Commit > " \
            --preview='git show --color=always --stat {1}' \
            --preview-window=right:60%:wrap \
            --height=90% --border \
            --header="Enter = diff completo" \
            --bind "enter:execute(git show --color=always {1} | less -R)"
}

stash() {
    _git_check || return 1
    echo ""
    echo "Acción:"
    echo "  1) Guardar stash"
    echo "  2) Aplicar stash (pop)"
    echo "  3) Borrar stash"
    echo "  4) Ver contenido"
    echo ""
    read -rp "Opción [1/2/3/4]: " OPT
    case "$OPT" in
        1)
            read -rp "📝 Descripción (Enter para automática): " DESC
            [ -n "$DESC" ] && git stash push -m "$DESC" || git stash push
            echo "✅ Stash guardado."
            ;;
        2|3|4)
            _fzf_check || return 1
            local LISTA
            LISTA=$(git stash list)
            [ -z "$LISTA" ] && echo "ℹ️  No hay stashes guardados." && return 0
            local LABEL
            case "$OPT" in
                2) LABEL="Aplicar stash" ;;
                3) LABEL="Borrar stash" ;;
                4) LABEL="Ver stash" ;;
            esac
            local ENTRADA
            ENTRADA=$(echo "$LISTA" | fzf \
                --prompt="$LABEL > " \
                --height=50% --border \
                --preview='git stash show -p --color=always $(echo {} | grep -o "stash@{[0-9]*}")' \
                --preview-window=right:60%:wrap --ansi)
            [ -z "$ENTRADA" ] && echo "❌ Cancelado." && return 1
            local IDX
            IDX=$(echo "$ENTRADA" | grep -o 'stash@{[0-9]*}')
            case "$OPT" in
                2) git stash pop "$IDX" && echo "✅ Stash aplicado." ;;
                3)
                    read -rp "⚠️  ¿Borrar '$IDX'? (s/N): " OK
                    [[ "$OK" =~ ^[sSyY]$ ]] && git stash drop "$IDX" && echo "✅ Stash eliminado."
                    ;;
                4) git stash show -p "$IDX" | less -R ;;
            esac
            ;;
        *) echo "❌ Opción no válida." ;;
    esac
}

remote() {
    _git_check || return 1
    echo ""
    echo "Acción:"
    echo "  1) Ver remotes"
    echo "  2) Añadir remote"
    echo "  3) Eliminar remote"
    echo ""
    read -rp "Opción [1/2/3]: " OPT
    case "$OPT" in
        1) git remote -v ;;
        2)
            read -rp "📝 Nombre del remote (ej: upstream): " NOMBRE
            [ -z "$NOMBRE" ] && echo "❌ Cancelado." && return 1
            read -rp "🔗 URL o usuario/repo de GitHub: " URL
            [ -z "$URL" ] && echo "❌ Cancelado." && return 1
            if [[ "$URL" != http* ]] && [[ "$URL" != git@* ]]; then
                [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ] \
                    && URL="git@github.com:${URL}.git" \
                    || URL="https://github.com/${URL}.git"
            fi
            git remote add "$NOMBRE" "$URL"
            echo "✅ Remote '$NOMBRE' → $URL añadido."
            ;;
        3)
            echo ""
            echo "📡 Remotes disponibles:"
            _listar_remotes | sed 's/^/  • /'
            echo ""
            read -rp "📝 Nombre del remote a eliminar: " REM
            [ -z "$REM" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Eliminar remote '$REM'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && git remote remove "$REM" && echo "✅ Remote '$REM' eliminado."
            ;;
        *) echo "❌ Opción no válida." ;;
    esac
}

tag() {
    _git_check || return 1
    echo ""
    echo "Acción:"
    echo "  1) Ver tags"
    echo "  2) Crear tag"
    echo "  3) Borrar tag local"
    echo "  4) Borrar tag remoto"
    echo "  5) Pushear todos los tags"
    echo ""
    read -rp "Opción [1/2/3/4/5]: " OPT
    case "$OPT" in
        1)
            local TAGS
            TAGS=$(git tag --sort=-version:refname)
            [ -z "$TAGS" ] && echo "ℹ️  No hay tags." && return 0
            echo "$TAGS"
            ;;
        2)
            read -rp "🏷️  Nombre del tag (ej: v1.0.0): " NOMBRE
            [ -z "$NOMBRE" ] && echo "❌ Cancelado." && return 1
            read -rp "📝 Mensaje (Enter para tag ligero): " MSG
            [ -n "$MSG" ] && git tag -a "$NOMBRE" -m "$MSG" || git tag "$NOMBRE"
            echo "✅ Tag '$NOMBRE' creado."
            read -rp "🚀 ¿Pushear tag al remoto? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && _git_run push origin "$NOMBRE" && echo "✅ Tag pusheado."
            ;;
        3)
            local TAGS
            TAGS=$(git tag --sort=-version:refname)
            [ -z "$TAGS" ] && echo "ℹ️  No hay tags." && return 0
            echo "$TAGS" | nl -w2 -s') '
            echo ""
            read -rp "📝 Nombre del tag a borrar: " T
            [ -z "$T" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Borrar tag local '$T'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && git tag -d "$T" && echo "✅ Tag local '$T' eliminado."
            ;;
        4)
            local TAGS
            TAGS=$(git tag --sort=-version:refname)
            [ -z "$TAGS" ] && echo "ℹ️  No hay tags." && return 0
            echo "$TAGS" | nl -w2 -s') '
            echo ""
            read -rp "📝 Nombre del tag remoto a borrar: " T
            [ -z "$T" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Borrar tag remoto '$T'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && _git_run push origin --delete "$T" && echo "✅ Tag remoto '$T' eliminado."
            ;;
        5)
            echo "🚀 Pusheando todos los tags..."
            _git_run push origin --tags
            echo "✅ Tags pusheados."
            ;;
        *) echo "❌ Opción no válida." ;;
    esac
}

git-clone() { clone "$@"; }
