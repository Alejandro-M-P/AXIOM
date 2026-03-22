#!/bin/bash
# ─── MÓDULO GIT: OPERACIONES SEGURAS E INTERACTIVAS ────────────────
# Requiere: fzf (instalado en la imagen base por bunker.sh)

# ── HELPERS INTERNOS ──────────────────────────────────────────────

_git_check() {
    git rev-parse --git-dir &>/dev/null || { echo "❌ No estás en un repositorio git. / Not inside a git repo."; return 1; }
}

_git_configure() {
    git config user.name  "$AXIOM_GIT_USER"
    git config user.email "$AXIOM_GIT_EMAIL"
}

# Ejecuta git con credenciales HTTPS+PAT o SSH según el modo configurado
_git_run() {
    if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
        git "$@"
    else
        [ -z "$AXIOM_GIT_TOKEN" ] && { echo "❌ Token vacío. / Empty token."; return 1; }
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
    command -v fzf &>/dev/null || { echo "❌ fzf no encontrado. Ejecuta 'axiom rebuild'. / fzf not found. Run 'axiom rebuild'."; return 1; }
}

# Selecciona rama local con fzf. Devuelve nombre de rama por stdout.
_fzf_rama_local() {
    local PROMPT="${1:-Selecciona rama / Select branch}"
    git branch --format='%(refname:short)' | fzf --prompt="$PROMPT > " --height=40% --border --ansi
}

# Selecciona rama remota con fzf (sin el prefijo origin/)
_fzf_rama_remota() {
    local PROMPT="${1:-Selecciona rama remota / Select remote branch}"
    git branch -r --format='%(refname:short)' | sed 's|origin/||' | grep -v HEAD | fzf --prompt="$PROMPT > " --height=40% --border --ansi
}

# Selecciona entre local Y remota
_fzf_rama_cualquiera() {
    local PROMPT="${1:-Selecciona rama / Select branch}"
    {
        git branch --format='%(refname:short)'
        git branch -r --format='%(refname:short)' | sed 's|origin/||' | grep -v HEAD
    } | sort -u | fzf --prompt="$PROMPT > " --height=40% --border --ansi
}

# ── COMANDOS PÚBLICOS ──────────────────────────────────────────────

# Clonar repositorio de forma segura
clone() {
    _fzf_check || return 1

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
        echo "🔑 Modo SSH"
        git clone "git@github.com:${REPO}.git" "$DIR"
    else
        echo "🔑 Modo HTTPS + PAT"
        _git_run clone "https://github.com/${REPO}.git" "$DIR"
    fi
    echo "✅ Repo clonado en ./$DIR"
}

# Crear rama nueva
branch() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    local BASE
    BASE=$(_fzf_rama_cualquiera "Rama base / Base branch")
    [ -z "$BASE" ] && echo "❌ Cancelado." && return 1

    read -rp "📝 Nombre de la nueva rama: " NOMBRE
    [ -z "$NOMBRE" ] && echo "❌ Se requiere un nombre." && return 1

    git checkout "$BASE" 2>/dev/null || { echo "❌ No se pudo cambiar a '$BASE'."; return 1; }
    git checkout -b "$NOMBRE"
    echo "✅ Rama '$NOMBRE' creada desde '$BASE'."
}

# Cambiar de rama (checkout interactivo)
switch() {
    _git_check || return 1
    _fzf_check || return 1

    local RAMA
    RAMA=$(_fzf_rama_cualquiera "Cambiar a / Switch to")
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1

    git checkout "$RAMA"
}

# Borrar rama local y/o remota
branch-delete() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    echo "🗑️  ¿Qué tipo de rama quieres borrar? / What type of branch to delete?"
    local TIPO
    TIPO=$(printf "local\nremota\nambas" | fzf --prompt="Tipo / Type > " --height=20% --border)
    [ -z "$TIPO" ] && echo "❌ Cancelado." && return 1

    local RAMA
    RAMA=$(_fzf_rama_local "Rama a borrar / Branch to delete")
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1

    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)
    if [ "$RAMA" = "$RAMA_ACTUAL" ]; then
        echo "❌ No puedes borrar la rama en la que estás ('$RAMA')."
        return 1
    fi

    case "$TIPO" in
        local)
            read -rp "⚠️  ¿Borrar rama local '$RAMA'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git branch -d "$RAMA" || git branch -D "$RAMA"
            echo "✅ Rama local '$RAMA' eliminada."
            ;;
        remota)
            read -rp "⚠️  ¿Borrar rama remota 'origin/$RAMA'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            _git_run push origin --delete "$RAMA"
            echo "✅ Rama remota 'origin/$RAMA' eliminada."
            ;;
        ambas)
            read -rp "⚠️  ¿Borrar rama LOCAL y REMOTA '$RAMA'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
            git branch -d "$RAMA" || git branch -D "$RAMA"
            _git_run push origin --delete "$RAMA" 2>/dev/null || echo "⚠️  No existía en remoto."
            echo "✅ Rama '$RAMA' eliminada (local + remota)."
            ;;
    esac
}

# Ver estado del repositorio con preview de diff por archivo
status() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    git status --short | fzf \
        --prompt="Archivo / File > " \
        --preview='git diff --color=always {2}' \
        --preview-window=right:60%:wrap \
        --height=80% --border \
        --header="Enter para ver diff completo / Enter to see full diff" \
        --bind "enter:execute(git diff {2} | less -R)" \
        --ansi
}

# Commit interactivo con preview del diff completo
commit() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    echo "📋 Cambios pendientes / Pending changes:"
    git status --short
    echo ""

    if git diff --cached --quiet && git diff --quiet; then
        echo "ℹ️  No hay cambios para commitear. / Nothing to commit."
        return 0
    fi

    echo "📂 Selecciona archivos a incluir (Tab para multi-selección, Enter para confirmar):"
    echo "   (ESC o vacío = incluir todos / ESC or empty = include all)"
    local ARCHIVOS
    ARCHIVOS=$(git status --short | fzf \
        --multi \
        --prompt="Incluir / Include > " \
        --preview='git diff --color=always {2}' \
        --preview-window=right:55%:wrap \
        --height=80% --border \
        --bind "ctrl-a:select-all" \
        --ansi | awk '{print $2}')

    _git_configure

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
    read -rp "🚀 ¿Hacer push ahora? / Push now? (s/N): " DOPUSH
    [[ "$DOPUSH" =~ ^[sSyY]$ ]] && push
}

# Push seguro con selección de remote y rama
push() {
    _git_check || return 1
    _fzf_check || return 1

    _git_configure

    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)

    # Selección de remote (si hay más de uno)
    local REMOTES
    REMOTES=$(git remote)
    local REMOTE_COUNT
    REMOTE_COUNT=$(echo "$REMOTES" | wc -l)

    local REMOTE
    if [ "$REMOTE_COUNT" -gt 1 ]; then
        REMOTE=$(echo "$REMOTES" | fzf --prompt="Remote > " --height=20% --border)
        [ -z "$REMOTE" ] && echo "❌ Cancelado." && return 1
    else
        REMOTE="${REMOTES:-origin}"
    fi

    # Confirmación de rama
    local RAMA
    RAMA=$(printf "%s\n(otra / other)" "$RAMA_ACTUAL" | fzf --prompt="Rama destino / Target branch > " --height=20% --border)
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1

    if [ "$RAMA" = "(otra / other)" ]; then
        read -rp "📝 Nombre de la rama destino: " RAMA
        [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
    fi

    echo "🚀 Push → $REMOTE/$RAMA ..."
    _git_run push "$REMOTE" "HEAD:$RAMA"
    echo "✅ Push completado."
}

# Pull con selección de remote y rama
pull() {
    _git_check || return 1
    _fzf_check || return 1

    local REMOTE
    REMOTE=$(git remote | fzf --prompt="Remote > " --height=20% --border)
    [ -z "$REMOTE" ] && echo "❌ Cancelado." && return 1

    # Fetch primero para ver ramas remotas actualizadas
    echo "🔄 Actualizando referencias remotas... / Fetching remote refs..."
    _git_run fetch "$REMOTE" --prune 2>/dev/null

    local RAMA
    RAMA=$(git branch -r --format='%(refname:short)' | grep "^$REMOTE/" | sed "s|^$REMOTE/||" | grep -v HEAD | \
        fzf --prompt="Rama a hacer pull / Branch to pull > " --height=40% --border)
    [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1

    echo "⬇️  Pull $REMOTE/$RAMA ..."
    _git_run pull "$REMOTE" "$RAMA"
    echo "✅ Pull completado."
}

# Merge interactivo
merge() {
    _git_check || return 1
    _fzf_check || return 1

    local RAMA_ACTUAL
    RAMA_ACTUAL=$(git branch --show-current)

    echo ""
    echo "🔀 Merge hacia '$RAMA_ACTUAL' desde:"
    local ORIGEN
    ORIGEN=$(_fzf_rama_cualquiera "Merge desde / Merge from")
    [ -z "$ORIGEN" ] && echo "❌ Cancelado." && return 1

    echo ""
    echo "🔀 Estrategia / Strategy:"
    local ESTRATEGIA
    ESTRATEGIA=$(printf "merge (preservar historial)\nmerge --squash (un solo commit)\nmerge --no-ff (siempre commit de merge)" | \
        fzf --prompt="Estrategia / Strategy > " --height=25% --border)
    [ -z "$ESTRATEGIA" ] && echo "❌ Cancelado." && return 1

    _git_configure

    case "$ESTRATEGIA" in
        *squash*)
            git merge --squash "$ORIGEN"
            echo "✅ Squash aplicado. Ahora haz commit con los cambios."
            ;;
        *no-ff*)
            git merge --no-ff "$ORIGEN"
            ;;
        *)
            git merge "$ORIGEN"
            ;;
    esac
}

# Rebase interactivo
rebase() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    echo "🔁 Tipo de rebase / Rebase type:"
    local TIPO
    TIPO=$(printf "rebase normal (sobre otra rama)\nrebase -i (editar commits)" | \
        fzf --prompt="Tipo / Type > " --height=20% --border)
    [ -z "$TIPO" ] && echo "❌ Cancelado." && return 1

    if [[ "$TIPO" == *"-i"* ]]; then
        local COMMITS
        COMMITS=$(git log --oneline -20 | fzf --prompt="Hasta este commit (no incluido) / Up to this commit > " --height=50% --border)
        [ -z "$COMMITS" ] && echo "❌ Cancelado." && return 1
        local HASH
        HASH=$(echo "$COMMITS" | awk '{print $1}')
        git rebase -i "${HASH}^"
    else
        local RAMA
        RAMA=$(_fzf_rama_cualquiera "Rebase sobre / Rebase onto")
        [ -z "$RAMA" ] && echo "❌ Cancelado." && return 1
        echo "⚠️  Esto reescribe el historial. ¿Continuar? (s/N): "
        read -r OK
        [[ "$OK" =~ ^[sSyY]$ ]] || { echo "❌ Cancelado."; return 0; }
        git rebase "$RAMA"
    fi
}

# Log visual interactivo con preview del diff de cada commit
log() {
    _git_check || return 1
    _fzf_check || return 1

    git log --oneline --color=always --decorate | \
        fzf --ansi \
            --prompt="Commit > " \
            --preview='git show --color=always --stat {1}' \
            --preview-window=right:60%:wrap \
            --height=90% --border \
            --header="Enter = diff completo / d = diff en pager" \
            --bind "enter:execute(git show --color=always {1} | less -R)" \
            --bind "d:execute(git diff {1}^ {1} | less -R)"
}

# Stash interactivo: guardar, listar, aplicar o borrar
stash() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    local ACCION
    ACCION=$(printf "guardar (push)\nlistar y aplicar (pop)\nlistar y borrar (drop)\nver contenido" | \
        fzf --prompt="Acción stash / Stash action > " --height=25% --border)
    [ -z "$ACCION" ] && echo "❌ Cancelado." && return 1

    case "$ACCION" in
        guardar*)
            read -rp "📝 Descripción del stash (Enter para automática): " DESC
            if [ -n "$DESC" ]; then
                git stash push -m "$DESC"
            else
                git stash push
            fi
            echo "✅ Stash guardado."
            ;;
        *aplicar*)
            local ENTRADA
            ENTRADA=$(git stash list | fzf --prompt="Aplicar stash / Apply stash > " --height=40% --border \
                --preview='git stash show -p --color=always {1}' --preview-window=right:60%:wrap --ansi)
            [ -z "$ENTRADA" ] && echo "❌ Cancelado." && return 1
            local IDX
            IDX=$(echo "$ENTRADA" | grep -o 'stash@{[0-9]*}')
            git stash pop "$IDX"
            echo "✅ Stash aplicado y eliminado."
            ;;
        *borrar*)
            local ENTRADA
            ENTRADA=$(git stash list | fzf --prompt="Borrar stash / Drop stash > " --height=40% --border \
                --preview='git stash show -p --color=always {1}' --preview-window=right:60%:wrap --ansi)
            [ -z "$ENTRADA" ] && echo "❌ Cancelado." && return 1
            local IDX
            IDX=$(echo "$ENTRADA" | grep -o 'stash@{[0-9]*}')
            read -rp "⚠️  ¿Borrar '$IDX'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && git stash drop "$IDX" && echo "✅ Stash eliminado."
            ;;
        *contenido*)
            git stash list | fzf --prompt="Ver stash / View stash > " --height=40% --border \
                --preview='git stash show -p --color=always {1}' --preview-window=right:60%:wrap --ansi \
                --bind "enter:execute(git stash show -p {1} | less -R)"
            ;;
    esac
}

# Añadir remote de forma interactiva
remote() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    local ACCION
    ACCION=$(printf "añadir remote\nver remotes\neliminar remote" | \
        fzf --prompt="Acción / Action > " --height=20% --border)
    [ -z "$ACCION" ] && echo "❌ Cancelado." && return 1

    case "$ACCION" in
        añadir*)
            read -rp "📝 Nombre del remote (ej: upstream): " NOMBRE
            [ -z "$NOMBRE" ] && echo "❌ Cancelado." && return 1
            read -rp "🔗 URL del repositorio (usuario/repo para GitHub): " URL
            [ -z "$URL" ] && echo "❌ Cancelado." && return 1
            # Aceptar formato corto usuario/repo
            if [[ "$URL" != http* ]] && [[ "$URL" != git@* ]]; then
                if [ "${AXIOM_AUTH_MODE:-https}" = "ssh" ]; then
                    URL="git@github.com:${URL}.git"
                else
                    URL="https://github.com/${URL}.git"
                fi
            fi
            git remote add "$NOMBRE" "$URL"
            echo "✅ Remote '$NOMBRE' → $URL añadido."
            ;;
        ver*)
            git remote -v
            ;;
        eliminar*)
            local REM
            REM=$(git remote | fzf --prompt="Eliminar remote / Remove remote > " --height=20% --border)
            [ -z "$REM" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Eliminar remote '$REM'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && git remote remove "$REM" && echo "✅ Remote '$REM' eliminado."
            ;;
    esac
}

# Crear, listar o borrar tags
tag() {
    _git_check || return 1
    _fzf_check || return 1

    echo ""
    local ACCION
    ACCION=$(printf "crear tag\nver tags\nborrar tag local\nborrar tag remoto\npushear tags al remoto" | \
        fzf --prompt="Acción / Action > " --height=30% --border)
    [ -z "$ACCION" ] && echo "❌ Cancelado." && return 1

    case "$ACCION" in
        crear*)
            read -rp "🏷️  Nombre del tag (ej: v1.0.0): " NOMBRE
            [ -z "$NOMBRE" ] && echo "❌ Cancelado." && return 1
            read -rp "📝 Mensaje (Enter para tag ligero sin mensaje): " MSG
            if [ -n "$MSG" ]; then
                git tag -a "$NOMBRE" -m "$MSG"
            else
                git tag "$NOMBRE"
            fi
            echo "✅ Tag '$NOMBRE' creado."
            read -rp "🚀 ¿Pushear tag al remoto? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && _git_run push origin "$NOMBRE" && echo "✅ Tag pusheado."
            ;;
        ver*)
            git tag --sort=-version:refname | fzf --prompt="Tags > " --height=50% --border \
                --preview='git show --color=always --stat {}' --preview-window=right:60%:wrap --ansi
            ;;
        borrar*local*)
            local T
            T=$(git tag | fzf --prompt="Borrar tag local / Delete local tag > " --height=40% --border)
            [ -z "$T" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Borrar tag local '$T'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && git tag -d "$T" && echo "✅ Tag local '$T' eliminado."
            ;;
        borrar*remoto*)
            local T
            T=$(git tag | fzf --prompt="Borrar tag remoto / Delete remote tag > " --height=40% --border)
            [ -z "$T" ] && echo "❌ Cancelado." && return 1
            read -rp "⚠️  ¿Borrar tag remoto '$T'? (s/N): " OK
            [[ "$OK" =~ ^[sSyY]$ ]] && _git_run push origin --delete "$T" && echo "✅ Tag remoto '$T' eliminado."
            ;;
        pushear*)
            echo "🚀 Pusheando todos los tags..."
            _git_run push origin --tags
            echo "✅ Tags pusheados."
            ;;
    esac
}

# Mantener alias de compatibilidad
git-clone() { clone "$@"; }