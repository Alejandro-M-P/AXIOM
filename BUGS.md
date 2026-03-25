# 🐛 Registro de Bugs y Estado del MVP

Este documento detalla los fallos conocidos durante la migración a **Go Nativo**.

## 🔴 Críticos y Seguridad (Lifecycle)
- [ ] **Seguridad de Volúmenes**: Las funciones nativas de `build` y `create` necesitan auditoría en el manejo de permisos de rootless containers.
- [ ] **Inyección de GPU**: Fallos esporádicos al detectar y montar drivers NVIDIA/AMD desde el binario de Go.
- [ ] **Persistencia**: El contenido de `~/.entorno/` a veces no se monta correctamente en el primer inicio.

## 🟡 Pendientes de Migración (Bash ➔ Go)
- [ ] Traducir herramientas de Git interactivo (`lib/git.sh`) a código nativo en `pkg/`.
- [ ] Migrar lógica de limpieza profunda (`axiom purge`).

## 🔵 Próximos Pasos
- [ ] **Configuración**: Mover el sistema de variables de `.env` a un archivo `config.toml` profesional.
