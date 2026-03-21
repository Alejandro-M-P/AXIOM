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

ayuda() {
mostrar_logo
    echo ""
    if [ -f "/run/.containerenv" ] || [ -f "/.dockerenv" ]; then
        echo "🤖  BÚNKER — Comandos internos"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  open               Sincronizar leyes y abrir opencode"
        echo "  sync-agents        Copiar tutor.md a AGENTS.md"
        echo "  save-rule [regla]  Guardar regla en tutor.md"
        echo "  diagnostico        Diagnóstico de salud"
        echo "  ayuda               Mostrar esta ayuda"
        echo "  git-clone [u/r]  Clonar repositorio (u: URL, r: repo local)"
        echo "  rama             Crear rama nueva (interactivo)"
        echo "  commit [mensaje] Añadir todo y commitear"
        echo "  push             Push seguro a GitHub"
    else
        echo "🛡️  SISTEMA BÚNKER — Comandos del host"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  build              Construir imagen base"
        echo "  crear [nombre]     Crear o entrar a un búnker"
        echo "  borrar [nombre]    Eliminar búnker y entorno"
        echo "  resetear           Limpieza total del sistema"
        echo "  ayuda               Mostrar esta ayuda"
        
    fi
}