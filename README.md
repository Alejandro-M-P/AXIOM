[🇬🇧 English](README-en.md) | [🇪🇸 Español](README.md)

# AXIOM Bunker System

> *Cero suciedad. Treinta segundos. Todo listo.*

Sistema de desarrollo aislado y modular construido sobre Distrobox + Podman. Cada búnker es un contenedor Arch Linux independiente con acceso directo a GPU, stack de IA local completo y Starship customizado — sin tocar el host.

---

## La historia

Bazzite es un sistema operativo atómico. Eso significa que el host es inmutable por diseño — no puedes instalar paquetes directamente en él como en cualquier otra distro. Y la verdad es que aunque pudieras, no querrías. El host es tu casa. La casa no es un campo de pruebas.

Así que cuando quise empezar a desarrollar en serio, la pregunta fue inevitable: ¿dónde instalo las cosas sin romper nada? La respuesta estaba ya instalada de serie en Bazzite: **Distrobox**. Contenedores que viven dentro del sistema, comparten el kernel, ven la GPU, acceden a tus archivos — pero están completamente separados del host.

Ahí empezó todo. Un script pequeño. Ridículamente pequeño. Solo creaba una caja y entraba. Eso era AXIOM en su primera versión.

Después llegó la idea de guardar el entorno de desarrollo en `~/dev` — separado del sistema, organizado, controlado. Cada proyecto en su carpeta, cada búnker con su home propio. El host seguía intacto. La suciedad, cero.

Y entonces me apunté a un curso gratuito de Mouredev sobre aprender a programar con IA. Ahí descubrí el stack de [Gentleman Programming](https://github.com/Gentleman-Programming): opencode, engram, gentle-ai, agent-teams-lite. Un ecosistema completo de herramientas de IA para desarrollo — todo local, todo en tu hardware, nada en la nube. Lo quise integrar. Y en vez de instalarlo a pelo en el host o en un contenedor cualquiera, lo metí dentro de AXIOM.

Un problema llevó a otro, cada solución hizo el sistema más sólido, y aquí estamos. Lo que empezó como un script de 10 líneas para no ensuciar Bazzite es ahora un sistema completo de entornos de desarrollo aislados con GPU, IA local y memoria compartida entre agentes.

**El objetivo nunca cambió: cero suciedad. Todo lo demás vino solo.**

---

## ¿Qué es exactamente?

Ejecutas `crear mi-proyecto`. En 30 segundos tienes un Arch Linux completo con acceso directo a tu GPU, Ollama corriendo local con tus modelos, y todo el stack de IA de Gentleman Programming listo para usar. Escribes como quieras, experimentas como quieras, rompes lo que quieras.

Cuando terminas, el host está exactamente igual que cuando empezaste. Cero suciedad.

Si el búnker se rompe, lo borras y creas otro. 30 segundos. Sin drama.

---

## Requisitos

- Distrobox ≥ 1.7
- Podman
- Host compatible (Bazzite, Fedora Silverblue, Nobara, CachyOS, cualquier distro con Podman)

---

## Instalación Rápida

1. **Clonar el repositorio:**
```bash
git clone https://github.com/Alejandro-M-P/AXIOM.git ~/AXIOM
cd ~/AXIOM
```

2. **Ejecutar el instalador:**
```bash
chmod +x install.sh && ./install.sh
```
El asistente te pedirá, en orden:

1. Usuario de GitHub
2. Email de GitHub
3. Token de GitHub (Classic o Fine-grained)
4. Directorio base (por defecto `~/dev`)
5. Directorio de modelos de Ollama (por defecto `~/dev/ai_config/models`)
6. *(Opcional)* GFX version para AMD — Enter para autodetectar
7. **Modo de drivers GPU:**

| Modo | Descripción | Tamaño imagen | Tiempo commit |
| :--- | :--- | :--- | :--- |
| `host` *(recomendado)* | Monta ROCm/CUDA del host. Para Bazzite, Fedora, Nobara, CachyOS... | ~10-13 GB | ~3-4 min |
| `image` | Instala ROCm/CUDA dentro. Funciona en cualquier distro. | ~38 GB | ~15 min |

3. **Añadir source a tu shell:**
```bash
echo "source ~/AXIOM/axiom.sh" >> ~/.bashrc
source ~/.bashrc
```

4. **Construir la imagen base (solo una vez):**
```bash
build
```
Detecta tu GPU automáticamente e instala todo el stack en `localhost/axiom-[gpu]:latest`. Tarda ~15-30 min de instalación más ~3-15 min de `podman commit` según el modo elegido. Al finalizar limpia automáticamente todos los cachés antes de commitear.

5. **Crear tu primer búnker:**
```bash
crear mi-primer-proyecto
```
A partir de aquí cada búnker nuevo arranca en ~30 segundos.

---

## Stack de IA incluido

Todo corre local. Nada sale a ningún servidor. Basado en el ecosistema de [Gentleman Programming](https://github.com/Gentleman-Programming), descubierto a través del curso de programación con IA de Mouredev.

| Herramienta | Qué hace |
| :--- | :--- |
| `opencode` | Editor de código con IA integrada |
| `engram` | Memoria persistente entre sesiones |
| `gentle-ai` | Interfaz de agentes IA |
| `agent-teams-lite` | Coordinación de múltiples agentes |
| `ollama` | Modelos de lenguaje corriendo en tu GPU |

### Configuración de Ollama en opencode

AXIOM escribe automáticamente la conexión a Ollama en `~/.config/opencode/opencode.json` dentro de cada búnker. opencode ya viene con sus propios providers por defecto — este bloque solo añade Ollama local encima de lo que ya trae:

```json
"provider": {
  "ollama": {
    "npm": "@ai-sdk/openai-compatible",
    "options": {
      "baseURL": "http://localhost:11434/v1"
    },
    "models": {
      "TU_MODELO:latest": {
        "reasoning": true
      }
    }
  }
}
```

Sustituye `TU_MODELO` por cualquier modelo que tengas descargado en Ollama. Algunos ejemplos habituales:

```bash
ollama pull qwen2.5:latest        # bueno para código, ligero
ollama pull qwen2.5-coder:latest  # especializado en código
ollama pull deepseek-r1:latest    # razonamiento fuerte
ollama pull llama3.1:latest       # uso general
```

Para ver qué modelos tienes disponibles:
```bash
ollama list
```

---

## tutor.md — La ley de tus agentes

Cuando trabajas con agentes de IA como opencode, el agente necesita contexto. Sin él, cada sesión empieza de cero — no sabe cómo quieres que trabaje, qué convenciones usas, qué decisiones técnicas has tomado, qué errores no quieres que repita.

`tutor.md` resuelve eso. Es un archivo de reglas que tú defines — o que defines junto a la IA — y que los agentes están obligados a leer cada vez que arrancan. No es una sugerencia. Se mete directamente en la carpeta de configuración del agente dentro del búnker, en el sitio exacto donde opencode lo lee al iniciar. El agente no puede ignorarlo.

```
~/dev/ai_global/teams/tutor.md
```

**Lo importante es dónde vive.** `tutor.md` no está dentro del búnker — está en `ai_global`, que es un volumen compartido montado en todos los búnkeres. Eso significa que cuando borras un búnker, las reglas no desaparecen. Están fuera, seguras, y el próximo búnker que crees las heredará automáticamente.

**Cómo añadir reglas:**

Puedes escribirlas tú directamente en el archivo, o usar `save-rule` desde dentro del búnker:

```bash
save-rule "usar siempre TypeScript estricto"
```

Te pedirá una razón técnica — obligatoria, por diseño. Fuerza a pensar antes de añadir una regla:

```
📝 Razón técnica: porque los errores de tipos en runtime cuestan más que escribir los tipos
```

La regla se guarda en `tutor.md` con su razón, y `sync-agents` se ejecuta automáticamente — copia el archivo actualizado a la configuración de todos los búnkeres que estén corriendo en ese momento.

**También puede hacerlo la IA.** Si durante una sesión de trabajo el agente detecta un patrón o una decisión que debería recordar, puede llamar a `save-rule` él mismo. La regla queda guardada para todas las sesiones futuras, en todos los búnkeres.

**`sync-agents`** es la función que mantiene todo sincronizado. Copia `tutor.md` a `~/.config/opencode/AGENTS.md` dentro de cada búnker activo, añadiendo encima un bloque de contexto con el tipo de GPU y la versión GFX detectada. Las notas locales que hayas escrito al principio del archivo se preservan siempre. Se ejecuta automáticamente al crear un búnker, al entrar en uno existente y al guardar una regla nueva. También puedes lanzarla manualmente:

```bash
sync-agents
```

El resultado es que tus agentes siempre trabajan con las mismas reglas, en todos los proyectos, sin importar cuántos búnkeres crees o destruyas. La memoria de cómo quieres trabajar sobrevive a todo.

### tutor.md recomendado

Este es un punto de partida sólido. Cópialo, adáptalo y hazlo tuyo:

```markdown
# 🤖 ROL: COPILOTO DE EJECUCIÓN (Junior Coder / Senior Mind)

## 👤 Identidad
Eres el brazo ejecutor del desarrollador. Tu misión es generar código limpio,
funcional y profesional a máxima velocidad, pero filtrado por un criterio de
Arquitecto Senior.

## 🌍 Contexto del Entorno
- Sistema: [tu distro, ej. Bazzite Linux]
- Contenedor: Arch Linux via Distrobox
- GPU: [tu GPU y drivers, ej. RDNA 4 con ROCm / NVIDIA con CUDA]
- Modelos IA: Ollama en /ai_config/models
- Proyecto activo: montado en /[nombre-del-búnker]
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
- **Git Ready**: Sugiere el momento del commit tras entregar un bloque funcional.

## 💾 Gestión del Entorno
- **Engram**: Registra archivos creados y decisiones técnicas para mantener contexto.
- **Save-Rule**: Si detectas una preferencia de código del desarrollador, sugiere
  grabarla con `save-rule` para que persista en todos los búnkeres.

## 📋 Cuándo guardar una regla
- Cuando se toma una decisión de arquitectura no obvia
- Cuando se resuelve un bug con una solución no trivial
- Cuando se establece un patrón que debe repetirse
- Cuando se descarta una tecnología con razón clara

## Reglas
- Protocolo de razón técnica activo.
```

---

## Comandos Disponibles

| Comando | Descripción | Entorno |
| :--- | :--- | :--- |
| `build` | Construye la imagen base con GPU, herramientas IA y starship. Solo se ejecuta una vez por máquina. | **Host** |
| `rebuild` | Reconstruye la imagen base para actualizar el stack. Los búnkeres existentes no se ven afectados. | **Host** |
| `resetear` | Borra la imagen base. Pregunta si también quieres borrar todos los búnkeres. | **Host** |
| `crear [nombre]` | Crea un nuevo búnker desde la imagen base (~30 seg) o entra en uno existente. | **Host** |
| `borrar [nombre]` | Solicita razón técnica y destruye el búnker y su memoria local por completo. | **Host** |
| `parar [nombre]` | Detiene el contenedor del búnker sin eliminar sus datos. | **Host** |
| `open` | Sincroniza leyes y abre el entorno inteligente `opencode`. | **Búnker** |
| `sync-agents` | Sincroniza `tutor.md` a la configuración local del agente. | **Búnker** |
| `save-rule [regla]` | Guarda una nueva regla técnica y la sincroniza con todos los búnkeres activos. | **Búnker** |
| `git-clone [u/r]` | Clona un repositorio de GitHub con el token pasado en memoria — sin modificar ni quedar grabado en el remote. | **Búnker** |
| `rama` | Crea una rama nueva de forma interactiva — pide nombre y rama base. | **Búnker** |
| `commit [mensaje]` | Añade todos los cambios y commitea. Si no hay mensaje lo pide. | **Búnker** |
| `push` | Hace push a GitHub de forma segura usando el token del `.env`. | **Búnker** |
| `diagnostico` | Diagnóstico de salud: GPU, Ollama y Token Git. | **Búnker** |
| `ayuda` | Muestra el menú de ayuda en pantalla. | **Host / Búnker** |

---

## Estructura de carpetas

Después de instalar y crear tu primer búnker, así queda todo en disco:

```
~/dev/                              ← AXIOM_BASE_DIR (configurable en install)
│
├── mi-proyecto/                    ← carpeta del proyecto, montada en /mi-proyecto dentro del búnker
│
├── ai_global/                      ← compartido entre TODOS los búnkeres
│   ├── teams/
│   │   └── tutor.md                ← reglas globales de los agentes
│   └── models/
│
├── ai_config/                      ← configuración de IA compartida
│   └── models/                     ← modelos de Ollama (un solo directorio para todos)
│
└── .entorno/                       ← homes de cada búnker (separados del proyecto)
    └── mi-proyecto/                ← home del búnker mi-proyecto
        ├── .bashrc                 ← variables, PATH, funciones del búnker
        ├── .config/
        │   ├── starship.toml       ← prompt personalizado Tokyo Night
        │   └── opencode/
        │       ├── opencode.json   ← conexión a Ollama local
        │       └── AGENTS.md       ← copia sincronizada de tutor.md
        └── ...                     ← resto del home aislado del host

~/AXIOM/                            ← el propio sistema AXIOM
├── axiom.sh                        ← script principal
├── install.sh                      ← instalador
└── .env                            ← tus credenciales y configuración (no se sube a git)
```

Lo más importante: el **proyecto** y el **entorno** están separados. El proyecto vive en `~/dev/mi-proyecto` y se monta dentro del búnker. El home del búnker vive en `.entorno/mi-proyecto`. Si borras el búnker, el proyecto no desaparece — solo el entorno.

---

## FAQ

**¿Puedo tener varios búnkeres a la vez?**
Sí, tantos como quieras. Cada uno tiene su propio home aislado, su propio `.bashrc`, su propia configuración. Todos comparten `ai_global` y `ai_config`, así que los modelos de Ollama y las reglas de `tutor.md` son los mismos en todos.

**¿Qué pasa si borro un búnker?**
Solo desaparece el entorno — el contenedor y su home en `.entorno/`. El código de tu proyecto en `~/dev/mi-proyecto` no se toca. Las reglas de `tutor.md` tampoco. Puedes recrear el búnker en 30 segundos y seguir exactamente donde lo dejaste.

**¿Los modelos de Ollama se comparten entre búnkeres?**
Sí. `ai_config/models` se monta en todos los búnkeres como directorio de modelos de Ollama. Descargas un modelo una vez y está disponible en todos.

**¿Puedo usar AXIOM sin GPU?**
Sí. Durante el `build` puedes seleccionar `Generic / CPU Only` y no se instalarán drivers de GPU. Ollama funciona en CPU, más lento pero funciona.

**¿Necesito cuenta en ningún servicio de IA?**
No. Todo el stack corre local en tu máquina. Ollama descarga los modelos directamente, sin cuentas ni APIs externas.

**¿Puedo cambiar el modo GPU después de instalar?**
Sí. Edita `AXIOM_ROCM_MODE` en el `.env` y ejecuta `rebuild`. Los búnkeres existentes no se ven afectados — solo los nuevos usarán el modo actualizado.

**¿Qué pasa si `build` falla a medias?**
Ejecuta esto para limpiar y volver a intentarlo:
```bash
distrobox-rm axiom-build --force
rm -rf ~/dev/.entorno/axiom-build
build
```

---

## Troubleshooting

**`opencode` / `engram` / `gentle-ai` no encontrados dentro del búnker**

Los binarios se instalan en `/usr/local/bin` durante el build. Comprueba:
```bash
podman run --rm localhost/axiom-rdna4:latest ls /usr/local/bin
```
Si no están, haz `rebuild`.

**`rocminfo: command not found` dentro del búnker**

En modo `host`, ROCm se monta desde el host. Comprueba que existe en tu sistema:
```bash
which rocminfo
ls /usr/lib/rocm
```
Si no está, cambia a modo `image` en el `.env` y haz `rebuild`.

**El `podman commit` tarda demasiado o parece congelado**

Normal con imágenes grandes. No lo interrumpas. Puede tardar hasta 15 min en modo `image`. Comprueba que sigue vivo:
```bash
podman ps
```

**`sync-agents` no actualiza los búnkeres existentes**

Solo sincroniza búnkeres que estén corriendo en ese momento. Si el búnker está parado, entra en él y ejecuta `sync-agents` manualmente.

**opencode no conecta con Ollama**

Comprueba que Ollama está corriendo dentro del búnker:
```bash
_ollama_ensure
ollama list
```
Si no responde, revisa los logs:
```bash
cat /tmp/ollama.log
```

**Starship no carga o da error**

Comprueba que está en `/usr/local/bin`:
```bash
which starship
starship --version
```
Si no está, haz `rebuild`.

---

## Contribuir

Las contribuciones son bienvenidas. Si encuentras un bug, tienes una idea o quieres mejorar algo, abre un issue o manda un PR.

**Para contribuir:**

1. Haz fork del repositorio
2. Crea una rama descriptiva:
```bash
git checkout -b fix/nombre-del-fix
# o
git checkout -b feat/nombre-de-la-feature
```
3. Haz tus cambios y commitea con un mensaje claro
4. Abre un Pull Request explicando qué cambia y por qué

**Lo que más ayuda:**
- Bugs reproducibles con pasos claros
- Soporte para distros o GPUs no contempladas
- Mejoras al proceso de build o al tamaño de la imagen
- Ejemplos de `tutor.md` para distintos lenguajes y stacks

Si tienes dudas antes de ponerte a trabajar en algo, abre un issue primero para alinearnos.

---

## Créditos y proyectos relacionados

AXIOM no existiría sin estos proyectos. Si algo falla o quieres profundizar, aquí están las fuentes:

| Proyecto | Repositorio | Para qué se usa |
| :--- | :--- | :--- |
| Distrobox | [89luca89/distrobox](https://github.com/89luca89/distrobox) | Motor de contenedores del sistema |
| Podman | [containers/podman](https://github.com/containers/podman) | Runtime de contenedores |
| opencode | [sst/opencode](https://github.com/sst/opencode) | Editor de código con IA |
| Ollama | [ollama/ollama](https://github.com/ollama/ollama) | Modelos de lenguaje local |
| engram | [Gentleman-Programming/engram](https://github.com/Gentleman-Programming/engram) | Memoria persistente entre sesiones |
| gentle-ai | [Gentleman-Programming/gentle-ai](https://github.com/Gentleman-Programming/gentle-ai) | Interfaz de agentes IA |
| agent-teams-lite | [Gentleman-Programming/agent-teams-lite](https://github.com/Gentleman-Programming/agent-teams-lite) | Coordinación de múltiples agentes |
| Starship | [starship/starship](https://github.com/starship/starship) | Prompt del terminal |
| Gentleman Programming | [Gentleman-Programming](https://github.com/Gentleman-Programming) | Stack de IA y curso de referencia |