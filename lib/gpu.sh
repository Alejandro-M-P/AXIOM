detect_gpu() {
    if [ -n "${AXIOM_GPU_TYPE:-}" ]; then
        export GPU_TYPE="$AXIOM_GPU_TYPE"
        export GFX_VAL="${AXIOM_GFX_VAL:-}"
        echo "✅ GPU forzada por .env / GPU forced by .env: $GPU_TYPE (GFX: ${GFX_VAL:-N/A})"
        return 0
    fi

    echo "🔍 Detectando hardware gráfico... / Detecting graphics hardware..."
    local HAS_RDNA4=0 HAS_RDNA3=0 HAS_NVIDIA=0 HAS_INTEL=0
    local GFX_RDNA4="12.0.1" GFX_RDNA3="11.0.0"
    export GPU_NAME=""
    
        local ROCM_GFX=""
    command -v rocminfo &>/dev/null && ROCM_GFX=$(rocminfo 2>/dev/null | grep -o 'gfx[0-9]*' | head -1)

    while IFS= read -r line; do
        local VENDOR DESC
        VENDOR=$(echo "$line" | sed -n 's/.*\[\([0-9a-fA-F]\{4\}\):.*/\1/p')
        DESC=$(echo "$line" | cut -d ' ' -f 2-)
        case "${VENDOR,,}" in
            10de) HAS_NVIDIA=1; GPU_NAME="$DESC" ;;
            1002)
                GPU_NAME="$DESC"
                case "$ROCM_GFX" in
                    gfx12*) HAS_RDNA4=1 ;;
                    gfx11*|gfx10*) HAS_RDNA3=1 ;;
                    *)
                        echo "$DESC" | grep -iqE '(8[0-9]{3}|9[0-9]{3})' && HAS_RDNA4=1
                        echo "$DESC" | grep -iqE '(6[0-9]{3}|7[0-9]{3})' && HAS_RDNA3=1
                        ;;
                esac
                ;;
            8086) HAS_INTEL=1; GPU_NAME="$DESC" ;;
        esac
    done < <(lspci -nn | grep -iE 'vga|3d|display|accelerator')

    if   [ "$HAS_NVIDIA" -eq 1 ]; then export GPU_TYPE="nvidia"; export GFX_VAL=""
    elif [ "$HAS_RDNA4"  -eq 1 ]; then export GPU_TYPE="rdna4";  export GFX_VAL="$GFX_RDNA4"
    elif [ "$HAS_RDNA3"  -eq 1 ]; then export GPU_TYPE="rdna3";  export GFX_VAL="$GFX_RDNA3"
    elif [ "$HAS_INTEL"  -eq 1 ]; then export GPU_TYPE="intel";  export GFX_VAL=""
    else
        echo "⚠️ Detección no concluyente. Selecciona: / Inconclusive detection. Select:"
        echo "1. RDNA 4 (8000/9000)  2. RDNA 3/2 (6000/7000)"
        echo "3. NVIDIA              4. INTEL"
        echo "5. Generic / CPU Only"
        read -rp "Opción/Option [1-5]: " GPU_OPT
        case "$GPU_OPT" in
            1) export GPU_TYPE="rdna4";   export GFX_VAL="$GFX_RDNA4" ;;
            2) export GPU_TYPE="rdna3";   export GFX_VAL="$GFX_RDNA3" ;;
            3) export GPU_TYPE="nvidia";  export GFX_VAL="" ;;
            4) export GPU_TYPE="intel";   export GFX_VAL="" ;;
            *) export GPU_TYPE="generic"; export GFX_VAL="" ;;
        esac
        if [[ "$GPU_TYPE" == rdna* ]]; then
            read -rp "📝 GFX Override (Enter para/for $GFX_VAL): " MANUAL_GFX
            [ -n "$MANUAL_GFX" ] && export GFX_VAL="$MANUAL_GFX"
        fi
    fi
    echo "✅ GPU: $GPU_TYPE ${GFX_VAL:+(GFX: $GFX_VAL)}"
}

_imagen_base() {
    echo "localhost/axiom-${GPU_TYPE:-generic}:latest"
}

_gpu_volumes_host() {
    local VOLS=""
    case "${GPU_TYPE:-}" in
        rdna*|generic)
            for P in /usr/lib/rocm /usr/lib64/rocm /opt/rocm; do
                [ -d "$P" ] && VOLS="$VOLS --volume $P:$P:ro"
            done
            for B in rocminfo rocm-smi; do
                local FP
                FP=$(command -v "$B" 2>/dev/null)
                [ -n "$FP" ] && VOLS="$VOLS --volume $FP:$FP:ro"
            done
            ;;
        nvidia)
            for P in /usr/lib/x86_64-linux-gnu/libcuda.so.1 /usr/local/cuda; do
                [ -e "$P" ] && VOLS="$VOLS --volume $P:$P:ro"
            done
            ;;
        intel)
            for P in /usr/lib/intel-opencl /usr/lib/x86_64-linux-gnu/intel-opencl; do
                [ -d "$P" ] && VOLS="$VOLS --volume $P:$P:ro"
            done
            ;;
    esac
    echo "$VOLS"
}