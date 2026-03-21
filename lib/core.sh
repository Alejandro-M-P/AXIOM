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
        echo "  build              Construir imagen base / Build base image"
        echo "  rebuild            Reconstruir imagen base / Rebuild base image"
        echo "  create [nombre]    Crear o entrar a un búnker / Create or enter a bunker"
        echo "  delete [nombre]    Eliminar búnker y entorno / Delete bunker and environment"
        echo "  stop [nombre]      Detener búnker sin borrarlo / Stop bunker without deleting it"
        echo "  reset              Limpieza total del sistema / Total system cleanup"
        echo "  reset-base         Borrar la imagen base / Delete the base image"
        echo "  help               Mostrar esta ayuda / Show this help"
        
    fi
}