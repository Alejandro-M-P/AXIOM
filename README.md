[🇬🇧 English](README-en.md) | [🇪🇸 Español](README.md)

# AXIOM Bunker System 🛡️

```text
  █████╗ ██╗  ██╗██╗ ██████╗ ███╗   ███╗
 ██╔══██╗╚██╗██╔╝██║██╔═══██╗████╗ ████║
 ███████║ ╚███╔╝ ██║██║   ██║██╔████╔██║
 ██╔══██║ ██╔██╗ ██║██║   ██║██║╚██╔╝██║
 ██║  ██║██╔╝ ██╗██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝
```

> *Cero suciedad. Treinta segundos. Todo listo.*

**AXIOM** es un sistema de desarrollo aislado y modular construido sobre **Distrobox** y **Podman**. Cada búnker es un contenedor Arch Linux independiente con acceso directo a GPU, stack de IA local completo y Starship customizado — sin tocar un solo archivo crítico del host.

Ideal para sistemas atómicos (Bazzite, Fedora Silverblue) o cualquier entorno donde la limpieza del sistema operativo base sea prioridad.

---

## 🔒 Seguridad (AXIOM Vault)

AXIOM implementa medidas estrictas para proteger tus credenciales y el host:
- **Tokens en volúmenes Read-Only:** El token de GitHub nunca se exporta como variable de entorno. Se lee de forma *on-demand* desde un volumen montado como solo lectura (`/run/axiom/env`), evitando que procesos maliciosos o extensiones del búnker puedan capturarlo mediante `printenv`.
- **Prevención de TOCTOU:** Se utiliza `mktemp` en operaciones críticas para bloquear ataques de condición de carrera.
- **Mitigación de Inyecciones:** Los comandos interactúan con las variables sensibles a través de arrays en bash en lugar de evaluar *strings* planos.

---

## 🚀 Instalación Rápida

**Requisitos:**
* Distrobox ≥ 1.7
* Podman
* fzf (Para los menús interactivos)
* Host compatible (Bazzite, Fedora Silverblue, Nobara, CachyOS, cualquier distro con Podman)

1. **Clonar el repositorio:**
```bash
git clone [https://github.com/Alejandro-M-P/AXIOM.git](https://github.com/Alejandro-M-P/AXIOM.git) ~/AXIOM
cd ~/AXIOM
```

2. **Ejecutar el instalador:**
```bash
chmod +x install.sh && ./install.sh
```
*(El asistente configurará credenciales, el directorio base y la detección de hardware gráfico).*

3. **Configurar el Shell y Construir:**
```bash
echo "source ~/AXIOM/axiom.sh" >> ~/.bashrc
source ~/.bashrc

axiom build
```
*(Detecta tu GPU e instala el stack. Tarda 15-30 min dependiendo del modo de driver elegido).*

---

## 💻 Uso Básico: Comandos del Host

Ejecutas `axiom create mi-proyecto` y en 30 segundos tienes un entorno completo listo para usar. Cuando terminas, el host está exactamente igual que al principio.

| Comando | Descripción |
| :--- | :--- |
| `axiom help` | Muestra la ayuda de los comandos del orquestador Go actual. |
| `axiom build` | Construye la imagen base con GPU y herramientas IA. |
| `axiom list` | Lista los búnkeres detectados con estado, tamaño, última entrada y rama git. |
| `axiom create <nombre>` | Crea un nuevo búnker desde la imagen base o entra en uno existente. |
| `axiom delete [nombre]` | Elimina un búnker. Si no pasas nombre, abre un selector con flechas. |
| `axiom delete-image` | Elimina la imagen base activa y muestra las imágenes AXIOM detectadas. |
| `axiom stop` | Detiene la ejecución de un búnker activo. |
| `axiom info [nombre]` | Muestra la ficha detallada de un búnker. |
| `axiom prune` | Limpia entornos huérfanos sin contenedor. |
| `axiom rebuild` | Reconstruye la imagen base. |
| `axiom reset` | Elimina TODOS los búnkeres e imágenes (Reset total). |

### Estado actual de la migración a Go
La lógica de host que ya está portada vive en `pkg/bunker` y se organiza así:

| Ruta | Responsabilidad |
| :--- | :--- |
| `pkg/bunker/bunker.go` | Orquestador, `Manager`, router de comandos y carga de `.env`. |
| `pkg/bunker/lifecycle.go` | Flujo de `axiom build` y ciclo de vida de la imagen base. |
| `pkg/bunker/instance.go` | `create`, `delete`, `list` y borrado de imagen base. |
| `pkg/bunker/select.go` | Selector interactivo para elegir búnkeres con flechas. |
| `pkg/bunker/templates.go` | Inyección de `starship`, `opencode` y archivos de arranque. |
| `pkg/ui/styles/` | Renderizado visual del ciclo de vida y vistas de bunker. |

---

## 🛡️ Herramientas Internas (Dentro del Búnker)

### Sistema de Agentes e IA
| Comando | Descripción |
| :--- | :--- |
| `open` | Inicia y abre `opencode`. |
| `sync-agents` | Copia `tutor.md` a la configuración local de los agentes (`AGENTS.md`). |
| `save-rule <regla>`| Guarda una regla en `tutor.md` obligando a dar una justificación técnica. |
| `diagnostics` | Ejecuta un diagnóstico interno del búnker. |

### Git Interactivo (Basado en fzf)
Dentro del búnker, dispones de comandos propios que sobreescriben Git para agilizar flujos y evitar errores visualmente:

| Comando | Descripción |
| :--- | :--- |
| `status` | Estado interactivo con *diff* visual a tiempo real. |
| `clone [u/r]` | Clona un repositorio indicando directamente `usuario/repo`. |
| `commit [msg]` | Selección de archivos con `<Tab>` antes de confirmar el commit. |
| `branch` | Creación interactiva de nuevas ramas. |
| `switch` | Cambio visual entre ramas existentes. |
| `branch-delete`| Borrado seguro y visual de ramas locales y/o remotas. |
| `push` / `pull`| Sincronización con selección interactiva de *remote* y rama. |
| `merge` / `rebase`| Selectores interactivos para origen y estrategias de integración. |
| `log` | Historial a color con vista previa del código modificado. |
| `stash` | Gestión interactiva (guardar, aplicar, borrar, ver contenido). |
| `remote` | Gestión visual de remotes (añadir, ver, eliminar). |
| `tag` | Creación y gestión interactiva de etiquetas. |

---

## 🧠 Stack de IA incluido

Todo corre local. Nada sale a ningún servidor. Basado en el ecosistema de Gentleman Programming.

| Herramienta | Función |
| :--- | :--- |
| `opencode` | Editor de código con IA integrada. |
| `engram` | Memoria persistente entre sesiones. |
| `gentle-ai` | Interfaz de agentes IA. |
| `agent-teams-lite` | Coordinación de múltiples agentes (Orchestrator, Apply). |
| `ollama` | Modelos de lenguaje corriendo localmente en tu GPU. |

---

## 📜 tutor.md — La ley de tus agentes

Para que la IA no empiece de cero en cada sesión, necesita contexto. `tutor.md` es el archivo de reglas que los agentes están obligados a leer al iniciar. 

Vive fuera del búnker (`~/dev/ai_config/teams/tutor.md`), en un volumen compartido. Si borras un búnker, tus reglas no desaparecen. El próximo las heredará. Puedes añadir reglas con `save-rule "regla"` desde dentro del búnker.

### Plantilla recomendada para tutor.md
Copia este bloque en tu `tutor.md` para establecer un comportamiento estricto y profesional en la IA:

```markdown
# 🤖 ROL: COPILOTO DE EJECUCIÓN (Junior Coder / Senior Mind)

## 👤 Identidad
Eres el brazo ejecutor del desarrollador. Tu misión es generar código limpio, funcional y profesional a máxima velocidad, pero filtrado por un criterio de Arquitecto Senior.

## 🛡️ Protocolo de Acción (Skeptic-to-Code)
1. **Skeptic First**: Antes de codear, pregunta el "porqué". Si la idea es mala o el código será basura, adviértelo. No seas un robot sumiso, sé un socio crítico.
2. **Explain & Validate**: Para tareas complejas, explica el diseño brevemente y espera el "OK". Para tareas simples, ejecuta.
3. **High-Speed Execution**: Entrega bloques completos listos para ser probados. No fragmentos inútiles.
4. **No Assumptions**: Si falta información, pídela. Es mejor preguntar una vez que corregir diez.

## 🏛️ Estándares de Calidad
- **Clean Code & Pro Naming**: El código debe hablar por sí solo.
- **Detección de Errores**: Al entregar código, indica los 2 puntos más probables por donde podría fallar.
```

---

## 🔬 Deep Dive: Arquitectura Interna y Decisiones Técnicas

Para los desarrolladores que necesitan entender qué ocurre bajo el capó:

### 1. ¿Por qué la imagen de GPU pesa 12GB vs 38GB?
Al ejecutar `install.sh`, se te pregunta por el **Modo de drivers GPU**:
* **Modo `host` (~12GB):** AXIOM no instala los pesados SDKs gráficos dentro del búnker. En su lugar, usa *bind mounts* para inyectar `/usr/lib/rocm` o `/usr/local/cuda` directamente desde tu sistema operativo base hacia el contenedor. Es la opción más eficiente, pero requiere que el host ya tenga los drivers instalados (ideal para Bazzite o distros gaming).
* **Modo `image` (~38GB):** AXIOM descarga e instala los paquetes completos de ROCm/CUDA (cuyo tamaño es masivo) *dentro* de la imagen Arch Linux. Esto infla la imagen, pero hace que el búnker sea 100% independiente de los drivers del host.

### 2. El Ciclo `build` -> `commit` -> `create`
AXIOM no instala dependencias cada vez que creas un proyecto.
* Cuando haces `axiom build`, se crea un contenedor temporal oculto, se instalan dependencias, Go, Node, Ollama y Opencode, y luego se ejecuta `podman commit`. Esto "congela" el estado en una imagen llamada `localhost/axiom-[gpu]:latest`.
* Cuando haces `axiom create mi-proyecto`, Distrobox simplemente clona esa imagen congelada en 30 segundos, inyectando un nuevo `--home` aislado en `~/.entorno/mi-proyecto`. 

### 3. Aislamiento Físico
AXIOM separa el código fuente del sistema operativo del contenedor:
* **El Código:** Vive en `~/dev/mi-proyecto` (tu host) y se monta en `/mi-proyecto` dentro del búnker.
* **El Sistema (Home):** Vive en `~/dev/.entorno/mi-proyecto`. Aquí están las configuraciones de Opencode, historiales de Bash, cachés, etc. 
* Si ejecutas `axiom delete`, se destruye el contenedor de Podman y el directorio `.entorno/`. **Tu código solo se borra si lo confirmas explícitamente durante `axiom delete`.**

---

## 🛠️ FAQ y Troubleshooting

**¿Puedo usar AXIOM sin GPU?**
Sí. Durante el `build` selecciona `Generic / CPU Only`. Ollama funcionará en CPU.

**opencode no conecta con Ollama**
Comprueba que Ollama corre dentro del búnker: `ollama list`. Si no responde, revisa `/tmp/ollama.log`.

**`rocminfo: command not found` dentro del búnker**
En modo `host`, ROCm se monta desde el host. Si tu host no lo tiene, usa el modo `image` en tu `.env` y haz `rebuild`.

**El `podman commit` tarda demasiado**
Normal con imágenes de 38GB (modo `image`). Puede tardar 15 minutos. Verifica que sigue activo con `podman ps` en otra terminal.

---

## 🤝 Contribuir
Haz fork, crea una rama descriptiva (`feat/nueva-funcion`), commitea con claridad y abre un Pull Request explicando el porqué. Lo que más ayuda: Soporte para distros/GPUs no contempladas y optimizaciones del `build`.

---

## 📖 La Historia y Filosofía

Bazzite es un sistema operativo atómico. Eso significa que el host es inmutable por diseño — no puedes instalar paquetes directamente en él como en cualquier otra distro. Y la verdad es que, aunque pudieras, no querrías. El host es tu casa, y la casa no es un campo de pruebas para código roto o dependencias huérfanas.

Cuando quise empezar a desarrollar en serio, la pregunta fue inevitable: ¿dónde instalo las cosas sin romper nada? La respuesta estaba instalada de serie: **Distrobox**. Contenedores que comparten el kernel y ven la GPU, pero están completamente separados del host. Ahí empezó todo: un script ridículamente pequeño que solo creaba una caja y entraba. 

Pronto evolucionó hacia la organización. Cada proyecto tendría su carpeta, cada búnker su propio `home` oculto. El host seguía intacto. Cero suciedad. Luego, tras realizar un curso de IA de Mouredev, descubrí el stack de *Gentleman Programming*. Un ecosistema potente, pero complejo de instalar y mantener. En lugar de ensuciar el host, lo integré todo directamente en el proceso de construcción de AXIOM.

Lo que empezó como un script de 10 líneas para proteger Bazzite es ahora una arquitectura completa de despliegue rápido. Entornos efímeros pero con memoria persistente, acceso a hardware dedicado y herramientas interactivas. 

**El objetivo nunca cambió: cero suciedad en tu máquina. Todo lo demás vino solo.**
