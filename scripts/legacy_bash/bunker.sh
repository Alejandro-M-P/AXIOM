#!/bin/bash
# ─── MÓDULO BUNKER: ENTRYPOINT ──────────
# Delegamos la lógica en los submódulos para mayor legibilidad.

_BUNKER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "$_BUNKER_DIR/bunker_lifecycle.sh"
source "$_BUNKER_DIR/bunker_info.sh"
source "$_BUNKER_DIR/bunker_templates.sh"
