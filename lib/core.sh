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
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  open               Sincronizar leyes y abrir opencode / Sync laws and open opencode"
        echo "  sync-agents        Copiar/Copy tutor.md a/to AGENTS.md"
        echo "  save-rule [regla]  Guardar regla en tutor.md / Save rule in tutor.md"
        echo "  diagnostics        Diagnóstico de salud / Health diagnostics"
        echo "  help               Mostrar esta ayuda / Show this help"
        echo "  git-clone [u/r]    Clonar repositorio / Clone repository"
        echo "  branch             Crear rama nueva / Create new branch (interactivo/interactive)"
        echo "  commit [mensaje]   Añadir todo y commitear / Add all and commit"
        echo "  push               Push seguro a GitHub / Secure push to GitHub"
    else
        echo "🛡️  SISTEMA BÚNKER — Comandos del host / Host commands"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  axiom build              Construir imagen base / Build base image"
        echo "  axiom rebuild            Reconstruir imagen base / Rebuild base image"
        echo "  axiom create [nombre]    Crear o entrar a un búnker / Create or enter a bunker"
        echo "  axiom delete [nombre]    Eliminar búnker y entorno / Delete bunker and environment"
        echo "  axiom stop [nombre]      Detener búnker sin borrarlo / Stop bunker without deleting it"
        echo "  axiom reset              Limpieza total del sistema / Total system cleanup"
        echo "  axiom rebuild            Borrar la imagen base / Delete the base image"
        echo "  axiom help               Mostrar esta ayuda / Show this help"
        
    fi
}

sync-agents() {
    echo "🔄 Sincronizando Tutor con búnkeres... / Syncing Tutor with bunkers..."
    if [ ! -f "$TUTOR_PATH" ]; then
        echo "⚠️  Tutor no encontrado en $TUTOR_PATH"
        return 1
    fi

    # Buscamos todos los .bashrc en la carpeta de entornos
    for bashrc in "$BASE_ENV"/*/.bashrc; do
        if [ -f "$bashrc" ]; then
            local DESTINO="$(dirname "$bashrc")/.config/opencode/AGENTS.md"
            mkdir -p "$(dirname "$DESTINO")"
            cp "$TUTOR_PATH" "$DESTINO"
        fi
    done
    echo "✅ Sincronización completada. / Sync complete."
}