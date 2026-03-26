#!/bin/bash
mostrar_logo() {
    echo -e "\033[1;34m"
    echo "  █████╗ ██╗  ██╗██╗ ██████╗ ███╗   ███╗"
    echo " ██╔══██╗╚██╗██╔╝██║██╔═══██╗████╗ ████║"
    echo " ███████║ ╚███╔╝ ██║██║   ██║██╔████╔██║"
    echo " ██╔══██║ ██╔██╗ ██║██║   ██║██║╚██╔╝██║"
    echo " ██║  ██║██╔╝ ██╗██║╚██████╔╝██║ ╚═╝ ██║"
    echo " ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝"
    echo -e "\033[0m"
}

help() {
    mostrar_logo
    echo ""
    if [ -f "/run/.containerenv" ] || [ -f "/.dockerenv" ]; then
        echo "🤖  BÚNKER — Comandos internos / Internal commands"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "  ── Sistema / System ─────────────────────────────────"
        echo "  open               Sincronizar y abrir opencode / Sync and open opencode"
        echo "  sync-agents        Copiar tutor.md → AGENTS.md / Copy tutor.md → AGENTS.md"
        echo "  save-rule [regla]  Guardar regla en tutor.md / Save rule in tutor.md"
        echo "  diagnostics        Diagnóstico del búnker / Bunker diagnostics"
        echo "  help               Mostrar esta ayuda / Show this help"
        echo ""
        echo "  ── Git ──────────────────────────────────────────────"
        echo "  clone [u/r]        Clonar repositorio / Clone repository"
        echo "  status             Estado interactivo con diff / Interactive status + diff"
        echo "  branch             Crear rama nueva / Create new branch"
        echo "  switch             Cambiar de rama / Switch branch"
        echo "  branch-delete      Borrar rama local y/o remota / Delete local and/or remote branch"
        echo "  commit [msg]       Seleccionar archivos y commitear / Stage files and commit"
        echo "  push               Push con selección de remote y rama / Push with remote and branch select"
        echo "  pull               Pull con selección de remote y rama / Pull with remote and branch select"
        echo "  merge              Merge interactivo / Interactive merge"
        echo "  rebase             Rebase interactivo / Interactive rebase"
        echo "  log                Log visual con preview de diff / Visual log with diff preview"
        echo "  stash              Gestión de stashes / Stash management"
        echo "  remote             Gestionar remotes / Manage remotes"
        echo "  tag                Crear, ver o borrar tags / Create, view or delete tags"
        echo ""
    else
        echo "🛡️  SISTEMA BÚNKER — Comandos del host / Host commands"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "  ── Imagen base / Base image ─────────────────────────"
        echo "  axiom build              Construir imagen base / Build base image"
        echo "  axiom rebuild            Reconstruir imagen base / Rebuild base image"
        echo "  axiom reset              Limpieza total del sistema / Full system cleanup"
        echo ""
        echo "  ── Búnkeres / Bunkers ───────────────────────────────"
        echo "  axiom list               Ver todos los búnkeres / List all bunkers"
        echo "  axiom create [nombre]    Crear o entrar a un búnker / Create or enter a bunker"
        echo "  axiom stop [nombre]      Parar búnker sin borrarlo / Stop bunker without deleting"
        echo "  axiom info [nombre]      Detalles de un búnker / Bunker details"
        echo "  axiom delete [nombre]    Eliminar búnker y entorno / Delete bunker and environment"
        echo "  axiom prune              Limpiar entornos huérfanos / Clean orphan environments"
        echo ""
        echo "  axiom help               Mostrar esta ayuda / Show this help"
        echo ""
    fi
}

sync-agents() {
    echo "🔄 Sincronizando Tutor con búnkeres... / Syncing Tutor with bunkers..."
    if [ ! -f "$TUTOR_PATH" ]; then
        echo "⚠️  Tutor no encontrado en / Not found at $TUTOR_PATH"
        return 1
    fi

    for bashrc in "$BASE_ENV"/*/.bashrc; do
        if [ -f "$bashrc" ]; then
            local DESTINO
            DESTINO="$(dirname "$bashrc")/.config/opencode/AGENTS.md"
            mkdir -p "$(dirname "$DESTINO")"
            cp "$TUTOR_PATH" "$DESTINO"
        fi
    done
    echo "✅ Sincronización completada. / Sync complete."
}