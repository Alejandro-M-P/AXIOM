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
    _fzf_check || return 1
    echo "🔄 Actualizando ramas..."
    _git_run fetch --all --prune 2>/dev/null || true
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    local BASE
    BASE=$(_listar_ramas_todas | fzf --prompt="Rama base (actual: $RAMA_ACTUAL) > " --height=40% --border)
    [ -z "$BASE" ] && echo "❌ Cancelado." && return 1
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
        echo "✅ Rama '$RAMA' creada con tracking a origin/$RAMA."
    fi
}

branch-delete() {
    _git_check || return 1
    _fzf_check || return 1
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    local RAMA
    RAMA=$(_listar_ramas_locales | grep -v "^$RAMA_ACTUAL$" | \
        fzf --prompt="Borrar rama > " --height=40% --border)
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
    local TIPO
    TIPO=$(printf "Solo local\nSolo remota\nAmbas" | fzf --prompt="¿Qué borrar? > " --height=20% --border)
    [ -z "$TIPO" ] && echo "❌ Cancelado." && return 1
    case "$TIPO" in
        "Solo local")
            read -rp "⚠️  ¿Borrar rama local '$RAMA'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git branch -d "$RAMA" || git branch -D "$RAMA"
            echo "✅ Rama local '$RAMA' eliminada."
            ;;
        "Solo remota")
            read -rp "⚠️  ¿Borrar rama remota 'origin/$RAMA'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            _git_run push origin --delete "$RAMA"
            echo "✅ Rama remota 'origin/$RAMA' eliminada."
            ;;
        "Ambas")
            read -rp "⚠️  ¿Borrar '$RAMA' local Y remota? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git branch -d "$RAMA" || git branch -D "$RAMA"
            _git_run push origin --delete "$RAMA" 2>/dev/null || echo "⚠️  No existía en remoto."
            echo "✅ Rama '$RAMA' eliminada (local + remota)."
            ;;
    esac
}

commit() {
    _git_check || return 1
    _fzf_check || return 1
    echo ""
    echo "📋 Cambios pendientes:"
    git status --short
    echo ""
    if git diff --cached --quiet && git diff --quiet; then
        echo "ℹ️  No hay cambios para commitear."
        return 0
    fi
    _git_configure
    echo "📂 Selecciona archivos (Tab = seleccionar, Ctrl+A = todos, Enter = confirmar):"
    local ARCHIVOS
    ARCHIVOS=$(git status --short | fzf \
        --multi \
        --prompt="Incluir > " \
        --preview='git diff --color=always {2}' \
        --preview-window=right:55%:wrap \
        --height=80% --border \
        --bind "ctrl-a:select-all" \
        --header="Tab = seleccionar | Ctrl+A = todos | Enter = confirmar" \
        --ansi | awk '{print $2}')
    if [ -z "$ARCHIVOS" ]; then
        git add -A
        echo "✅ Todos los archivos añadidos."
    else
        echo "$ARCHIVOS" | xargs git add
        echo "✅ Archivos seleccionados añadidos."
    fi
    local MENSAJE="${1:-}"
    if [ -z "$MENSAJE" ]; then
        read -rp "📝 Mensaje del commit: " MENSAJE
        [ -z "$MENSAJE" ] && echo "❌ Se requiere un mensaje." && return 1
    fi
    git commit -m "$MENSAJE"
    echo ""
    local DOPUSH
    DOPUSH=$(printf "Sí, hacer push\nNo" | fzf --prompt="¿Hacer push? > " --height=15% --border)
    [[ "$DOPUSH" == "Sí, hacer push" ]] && push
}

push() {
    _git_check || return 1
    _fzf_check || return 1
    _git_configure
    echo "🔄 Actualizando referencias..."
    _git_run fetch --all --prune 2>/dev/null || true
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    local REMOTE
    REMOTE=$(_listar_remotes | fzf --prompt="Remote > " --height=20% --border)
    [ -z "$REMOTE" ] && echo "❌ Cancelado." && return 1
    local RAMA
    RAMA=$(printf "%s\n%s" "$RAMA_ACTUAL" "$(git branch --format='%(refname:short)' | grep -v "^$RAMA_ACTUAL$")" | \
        fzf --prompt="Rama destino > " --height=40% --border)
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
    echo "🚀 Push → $REMOTE/$RAMA ..."
    _git_run push "$REMOTE" "HEAD:$RAMA"
    echo "✅ Push completado."
}

pull() {
    _git_check || return 1
    _fzf_check || return 1
    local REMOTE
    REMOTE=$(_listar_remotes | fzf --prompt="Remote > " --height=20% --border)
    [ -z "$REMOTE" ] && echo "❌ Cancelado." && return 1
    echo "🔄 Actualizando referencias de $REMOTE..."
    _git_run fetch "$REMOTE" --prune 2>/dev/null
    local RAMA
    RAMA=$(git branch -r --format='%(refname:short)' | grep "^$REMOTE/" | \
        sed "s|^$REMOTE/||" | grep -v HEAD | \
        fzf --prompt="Rama a hacer pull > " --height=40% --border)
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
    echo "⬇️  Pull $REMOTE/$RAMA ..."
    _git_run pull "$REMOTE" "$RAMA"
    echo "✅ Pull completado."
}

merge() {
    _git_check || return 1
    _fzf_check || return 1
    echo "🔄 Actualizando ramas..."
    _git_run fetch --all --prune 2>/dev/null || true
    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    local ORIGEN
    ORIGEN=$(_listar_ramas_todas | grep -v "^$RAMA_ACTUAL$" | \
        fzf --prompt="Mergear en '$RAMA_ACTUAL' desde > " --height=40% --border)
    [ -z "$ORIGEN" ] && echo "❌ Cancelado." && return 1
    local ESTRATEGIA
    ESTRATEGIA=$(printf "merge normal (preservar historial)\nsquash (un solo commit)\nno-ff (siempre commit de merge)" | \
        fzf --prompt="Estrategia > " --height=25% --border)
    [ -z "$ESTRATEGIA" ] && echo "❌ Cancelado." && return 1
    _git_configure
    case "$ESTRATEGIA" in
        squash*) git merge --squash "$ORIGEN" && echo "✅ Squash aplicado. Ahora haz commit." ;;
        no-ff*)  git merge --no-ff "$ORIGEN" ;;
        *)       git merge "$ORIGEN" ;;
    esac
}

rebase() {
    _git_check || return 1
    _fzf_check || return 1
    local TIPO
    TIPO=$(printf "Normal (sobre otra rama)\nInteractivo -i (editar commits)" | \
        fzf --prompt="Tipo de rebase > " --height=20% --border)
    [ -z "$TIPO" ] && echo "❌ Cancelado." && return 1
    case "$TIPO" in
        Interactivo*)
            local COMMIT
            COMMIT=$(git log --oneline -20 | \
                fzf --prompt="Rebase desde este commit (no incluido) > " --height=50% --border)
            [ -z "$COMMIT" ] && echo "❌ Cancelado." && return 1
            local HASH
            HASH=$(echo "$COMMIT" | awk '{print $1}')
            git rebase -i "${HASH}^"
            ;;
        Normal*)
            _git_run fetch --all --prune 2>/dev/null || true
            local RAMA
            RAMA=$(_listar_ramas_todas | fzf --prompt="Rebase sobre > " --height=40% --border)
            [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  Esto reescribe el historial. ¿Continuar? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git rebase "$RAMA"
            ;;
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
    _fzf_check || return 1
    local ACCION
    ACCION=$(printf "Guardar stash\nAplicar stash (pop)\nBorrar stash\nVer contenido" | \
        fzf --prompt="Acción > " --height=25% --border)
    [ -z "$ACCION" ] && echo "❌ Cancelado." && return 1
    case "$ACCION" in
        "Guardar stash")
            read -rp "📝 Descripción (Enter para automática): " DESC
            [ -n "$DESC" ] && git stash push -m "$DESC" || git stash push
            echo "✅ Stash guardado."
            ;;
        *)
            local LISTA
            LISTA=$(git stash list)
            [ -z "$LISTA" ] && echo "ℹ️  No hay stashes guardados." && return 0
            local ENTRADA
            ENTRADA=$(echo "$LISTA" | fzf \
                --prompt="$ACCION > " \
                --height=50% --border \
                --preview='git stash show -p --color=always $(echo {} | grep -o "stash@{[0-9]*}")' \
                --preview-window=right:60%:wrap --ansi)
            [ -z "$ENTRADA" ] && echo "❌ Cancelado." && return 1
            local IDX
            IDX=$(echo "$ENTRADA" | grep -o 'stash@{[0-9]*}')
            case "$ACCION" in
                "Aplicar stash (pop)") git stash pop "$IDX" && echo "✅ Stash aplicado." ;;
                "Borrar stash")
                    read -rp "⚠️  ¿Borrar '$IDX'? (s/N): " OK
                    [[ "$OK" =~ ^[sSyY]$ ]] && git stash drop "$IDX" && echo "✅ Stash eliminado."
                    ;;
                "Ver contenido") git stash show -p "$IDX" | less -R ;;
            esac
            ;;
    esac
}

remote() {
    _git_check || return 1
    _fzf_check || return 1
    local ACCION
    ACCION=$(printf "Ver remotes\nAñadir remote\nEliminar remote" | \
        fzf --prompt="Acción > " --height=20% --border)
    [ -z "$ACCION" ] && echo "❌ Cancelado." && return 1
    case "$ACCION" in
        "Ver remotes") git remote -v ;;
        "Añadir remote")
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
        "Eliminar remote")
            local REM
            REM=$(_listar_remotes | fzf --prompt="Eliminar remote > " --height=20% --border)
            [ -z "$REM" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Eliminar remote '$REM'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && git remote remove "$REM" && echo "✅ Remote '$REM' eliminado."
            ;;
    esac
}

tag() {
    _git_check || return 1
    _fzf_check || return 1
    local ACCION
    ACCION=$(printf "Ver tags\nCrear tag\nBorrar tag local\nBorrar tag remoto\nPushear todos los tags" | \
        fzf --prompt="Acción > " --height=30% --border)
    [ -z "$ACCION" ] && echo "❌ Cancelado." && return 1
    case "$ACCION" in
        "Ver tags")
            local TAGS
            TAGS=$(git tag --sort=-version:refname)
            [ -z "$TAGS" ] && echo "ℹ️  No hay tags." && return 0
            echo "$TAGS" | fzf --prompt="Tags > " --height=50% --border \
                --preview='git show --color=always --stat {}' \
                --preview-window=right:60%:wrap --ansi
            ;;
        "Crear tag")
            read -rp "🏷️  Nombre del tag (ej: v1.0.0): " NOMBRE
            [ -z "$NOMBRE" ] && echo "❌ Cancelado." && return 1
            read -rp "📝 Mensaje (Enter para tag ligero): " MSG
            [ -n "$MSG" ] && git tag -a "$NOMBRE" -m "$MSG" || git tag "$NOMBRE"
            echo "✅ Tag '$NOMBRE' creado."
            local PUSH
            PUSH=$(printf "Sí, pushear tag\nNo" | fzf --prompt="¿Pushear? > " --height=15% --border)
            [[ "$PUSH" == "Sí, pushear tag" ]] && _git_run push origin "$NOMBRE" && echo "✅ Tag pusheado."
            ;;
        "Borrar tag local")
            local T
            T=$(git tag --sort=-version:refname | fzf --prompt="Borrar tag local > " --height=40% --border)
            [ -z "$T" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Borrar tag local '$T'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && git tag -d "$T" && echo "✅ Tag local '$T' eliminado."
            ;;
        "Borrar tag remoto")
            local T
            T=$(git tag --sort=-version:refname | fzf --prompt="Borrar tag remoto > " --height=40% --border)
            [ -z "$T" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Borrar tag remoto '$T'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && _git_run push origin --delete "$T" && echo "✅ Tag remoto '$T' eliminado."
            ;;
        "Pushear todos los tags")
            echo "🚀 Pusheando todos los tags..."
            _git_run push origin --tags
            echo "✅ Tags pusheados."
            ;;
    esac
}

git-clone() { clone "$@"; }
