_init_tutor() {
    [ -f "$TUTOR_PATH" ] && return 0
    mkdir -p "$(dirname "$TUTOR_PATH")"

    local DISTRO GPU_INFO GFX_INFO
    DISTRO=$(grep ^NAME /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '"')
    [ -z "$DISTRO" ] && DISTRO="Linux"

    if [ -n "${GPU_TYPE:-}" ]; then
        GPU_INFO="${GPU_NAME:-$GPU_TYPE}"
        case "${GPU_TYPE}" in
            rdna*) GPU_INFO="AMD ${GPU_NAME:-RDNA} con ROCm" ;;
            nvidia) GPU_INFO="NVIDIA ${GPU_NAME:-} con CUDA" ;;
            intel) GPU_INFO="Intel ${GPU_NAME:-} con OneAPI" ;;
            *) GPU_INFO="${GPU_NAME:-Generic / CPU Only}" ;;
        esac
        GFX_INFO="${GFX_VAL:+ (HSA_OVERRIDE_GFX_VERSION=$GFX_VAL)}"
    else
        GPU_INFO="desconocida"
        GFX_INFO=""
    fi

    cat > "$TUTOR_PATH" << TUTOR
    # 🤖 ROL: COPILOTO DE EJECUCIÓN (Junior Coder / Senior Mind)

    ## 👤 Identidad
    Eres el brazo ejecutor del desarrollador. Tu misión es generar código limpio,
    funcional y profesional a máxima velocidad, pero filtrado por un criterio de
    Arquitecto Senior.

    ## 🌍 Contexto del Entorno
    - Sistema: $DISTRO (atómico)
    - Contenedor: Arch Linux via Distrobox (AXIOM)
    - GPU: $GPU_INFO$GFX_INFO
    - Modelos IA: Ollama en /ai_config/models
    - Memoria persistente: /ai_global/teams/tutor.md

    ## 🛡️ Protocolo de Acción (Skeptic-to-Code)
    1. **Skeptic First**: Antes de codear, pregunta el "porqué". Si la idea es mala
    o el código será basura, adviértelo. No seas un robot sumiso, sé un socio crítico.
    2. **Explain & Validate**: Para tareas complejas, explica el diseño brevemente y
    espera el "OK". Para tareas simples y directas, ejecuta sin preguntar.
    3. **High-Speed Execution**: Una vez recibas el "OK", genera el código completo.
    No des fragmentos inútiles; entrega bloques listos para ser probados o integrados.
    4. **No Assumptions**: Si falta información para completar el código, pídela.
    Es mejor preguntar una vez que corregir diez.

    ## 🏛️ Estándares de Calidad
    - **Clean Code & Pro Naming**: El código debe hablar por sí solo.
    - **Detección de Errores**: Al entregar código, indica los 2 puntos más probables
    por donde podría fallar.
    - **Git Ready**: Sugiere el momento del commit tras entregar un bloque funcional. El usuario decidirá el mensaje.

    ## 💾 Gestión del Entorno
    - **Engram**: Registra archivos creados y decisiones técnicas para mantener contexto.
    - **Save-Rule**: Si detectas una preferencia de código del desarrollador, sugiere
    grabarla con \`save-rule\` para que persista en todos los búnkeres.

    ## 📋 Cuándo guardar una regla
    - Cuando se toma una decisión de arquitectura no obvia
    - Cuando se resuelve un bug con una solución no trivial
    - Cuando se establece un patrón que debe repetirse
    - Cuando se descarta una tecnología con razón clara

    ## Reglas
    - Protocolo de razón técnica activo.
    TUTOR
        echo "✅ tutor.md inicializado con datos del sistema."
}

sync-agents() {
    _init_tutor
    local SYNCED=0
    while IFS= read -r CAJA; do
        [ -d "$BASE_ENV/$CAJA" ] || continue
        local DEST="$BASE_ENV/$CAJA/.config/opencode/AGENTS.md"
        mkdir -p "$(dirname "$DEST")"
        cp "$TUTOR_PATH" "$DEST"
        SYNCED=$((SYNCED+1))
    done < <(podman ps --format '{{.Names}}' 2>/dev/null)
    [ $SYNCED -gt 0 ] && echo "✅ Ley Global sincronizada en $SYNCED búnker(es)."
    return 0
}


save-rule() {
    read -rp "📝 Razón técnica: " REASON
    [ -z "$REASON" ] && echo "❌ Se requiere razón técnica." && return 1
    local R="${1:-}"
    [ -z "$R" ] && read -rp "📝 Regla: " R
    echo "- $R (Razón: $REASON)" >> /ai_global/teams/tutor.md
    sync-agents
}

diagnostico() {
    echo "🔍 [DIAGNÓSTICO AXIOM]"
    echo "──────────────────────"
    echo "1️⃣  GPU:"
    if command -v nvidia-smi &>/dev/null; then
        nvidia-smi | grep "Driver Version" || echo "❌ nvidia-smi falló"
    elif command -v rocminfo &>/dev/null; then
        rocminfo | grep "Agent 1" -A 2 || echo "❌ rocminfo falló"
    else
        echo "⚠️ Sin herramientas de GPU visibles."
    fi
    echo ""
    echo "2️⃣  Git Token:"
    [ -n "$AXIOM_GIT_TOKEN" ] && echo "✅ Token presente." || echo "❌ AXIOM_GIT_TOKEN vacío."
    echo ""
    echo "3️⃣  Ollama:"
    pgrep -x ollama > /dev/null && echo "✅ En ejecución." || echo "⚠️ No está corriendo."
    echo ""
    echo "4️⃣  Herramientas IA:"
    for BIN in opencode engram gentle-ai ollama; do
        command -v $BIN &>/dev/null && echo "  ✅ $BIN" || echo "  ❌ $BIN no encontrado"
    done
    echo ""
    echo "5️⃣  AGENTS.md:"
    [ -f ~/.config/opencode/AGENTS.md ] && echo "✅ Presente." || echo "⚠️ No existe — ejecuta sync-agents"
}


open() {
    sync-agents
    # Función de ayuda para asegurar que Ollama corre antes de abrir opencode
    if command -v ollama &>/dev/null; then
        ollama list &>/dev/null || (ollama serve > /tmp/ollama.log 2>&1 & sleep 2)
    fi
    opencode
}
