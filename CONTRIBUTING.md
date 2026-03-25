# 🤝 Contribuye a AXIOM

¡Hola! Soy Alejandro, el único dev detrás de AXIOM. Busco aliados para terminar de jubilar el código en Bash y construir un orquestador robusto en Go.

## 📜 Reglas de Oro
1. **Go Nativo o nada**: No aceptamos llamadas a scripts de Bash (`os/exec` sobre archivos .sh). Queremos lógica pura en Go.
2. **Los .sh son "Planos"**: Los archivos en `lib/` están ahí solo para que entiendas la lógica antigua y me ayudes a portarla a `pkg/`.
3. **Seguridad primero**: Si encuentras un fallo en el manejo de contenedores de `pkg/bunker/lifecycle.go`, tu PR tiene prioridad máxima.

## 🏗️ Cómo empezar
El comando más estable ahora mismo es `axiom list`. Puedes empezar revisando cómo lee el sistema de archivos y ayudarme a estabilizar `axiom info` o `axiom build`.
