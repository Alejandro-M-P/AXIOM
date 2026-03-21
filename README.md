# AXIOM Bunker System

El sistema de desarrollo definitivo aislado y modular, construido sobre Distrobox + Podman. Cada búnker es un contenedor Arch Linux independiente con acceso a GPU, herramientas locales de IA y Starship customizado.

---

## ¿Por qué existe AXIOM?

**AXIOM** nació de una pregunta simple: ¿por qué cada vez que empiezas un proyecto nuevo tienes que perder una tarde configurando el entorno?

La respuesta habitual es "usa Docker" o "usa un devcontainer". Pero eso te aleja del sistema, te mete en capas de abstracción que no controlas, y en Linux gaming — donde el host ya es Bazzite, ya tiene ROCm, ya tiene todo — se siente como construir una jaula dentro de tu propia casa.

AXIOM toma otra dirección. En vez de abstraer, **aisla sin desconectar**. Cada búnker es un Arch Linux completo con acceso directo a tu GPU, tus modelos de Ollama, tu token de GitHub — pero sin tocar el host. Puedes romper el búnker, quemarlo, borrarlo con una razón técnica obligatoria, y el host sigue intacto. El proyecto sigue intacto. Solo desaparece el entorno.

La palabra **búnker** no es casual. Un búnker no es una prisión — es un lugar desde el que operas cuando el exterior es hostil. Dependencias que rompen, versiones incompatibles, experimentos que salen mal. El búnker absorbe el daño.

La imagen base es el corazón del sistema. Se construye una vez, con todo: drivers de GPU, el stack de IA completo, starship configurado. A partir de ahí, crear un búnker nuevo es clonar esa imagen y entrar. Treinta segundos. Sin esperas, sin instalar lo mismo por décima vez.

**¿Por qué Arch dentro de distrobox?**
Porque el AUR existe. Porque `paru` resuelve en una línea lo que en otras distros son tres horas de compilación manual. Porque si necesitas ROCm bleeding edge, está ahí. Arch dentro de distrobox es la combinación más pragmática que existe para desarrollo en Linux — tienes el ecosistema más amplio de paquetes sin arriesgar la estabilidad del host.

**¿Por qué las AI tools integradas desde el build?**
Porque el flujo de trabajo cambió. `opencode` no es un accesorio — es donde ocurre el trabajo. Engram recuerda. Gentle-AI conecta. Agent Teams coordina. Ollama corre local, en tu hardware, sin mandar nada a ningún servidor. Integrar todo esto en la imagen base no es comodidad — es una declaración de que este stack es parte del entorno de desarrollo, no un añadido opcional.

**¿Por qué la razón técnica obligatoria para borrar?**
Porque borrar sin pensar es fácil. La fricción es intencional. Si no puedes articular por qué estás destruyendo un entorno, quizás no deberías hacerlo todavía.

AXIOM no es para todo el mundo. Es para el desarrollador que ya vive en la terminal, que ya juega en Linux, que ya confía más en `pacman` que en cualquier gestor de paquetes gráfico, y que quiere que su entorno de IA local sea tan serio como el resto de su setup.

---

## Requisitos

- Distrobox ≥ 1.7
- Podman
- Host compatible (Bazzite, Fedora Silverblue, cualquier distro con Podman)

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
Sigue el asistente para proveer tus datos de GitHub, tu token y tu directorio de desarrollo base preferido.

3. **Añadir source a tu shell:**
```bash
echo "source ~/AXIOM/axiom.sh" >> ~/.bashrc
source ~/.bashrc
```

4. **Construir la imagen base (solo una vez):**
```bash
build
```
Detecta tu GPU automáticamente e instala todo el stack (drivers, herramientas IA, starship) en una imagen `localhost/axiom-[gpu]:latest`. Tarda ~15-30 min pero **solo se ejecuta una vez**. Al finalizar, el build limpia automáticamente todos los cachés (pacman, paru, Go, `/tmp`) antes de commitear la imagen, reduciendo su tamaño de ~40 GB a ~15 GB sin perder nada funcional.

5. **Crear tu primer búnker:**
```bash
crear mi-primer-proyecto
```
A partir de aquí cada búnker nuevo arranca en ~30 segundos.

---

## Comandos Disponibles

| Comando | Descripción | Entorno |
| :--- | :--- | :--- |
| `build` | Construye la imagen base con GPU, herramientas IA y starship. Solo se ejecuta una vez por máquina. | **Host** |
| `rebuild` | Reconstruye la imagen base para actualizar el stack. Los búnkeres existentes no se ven afectados. | **Host** |
| `crear [nombre]` | Crea un nuevo búnker desde la imagen base (~30 seg) o entra en uno existente. | **Host** |
| `borrar [nombre]` | Solicita razón técnica y destruye el búnker y su memoria local por completo. | **Host** |
| `parar [nombre]` | Detiene el contenedor del búnker sin eliminar sus datos. | **Host** |
| `open` | Sincroniza leyes y abre el entorno inteligente `opencode`. | **Búnker** |
| `sync-agents` | Sincroniza `tutor.md` a la configuración local del agente. | **Búnker** |
| `save-rule [regla]` | Guarda una nueva regla técnica y la sincroniza con todos los búnkeres activos. | **Búnker** |
| `git-clone [u/r]` | Clona un repositorio de GitHub con token y limpia las credenciales del remote. | **Búnker** |
| `push` | Hace push a GitHub de forma segura usando el token del `.env`. | **Búnker** |
| `diagnostico` | Diagnóstico de salud: GPU, Ollama y Token Git. | **Búnker** |
| `ayuda` | Muestra el menú de ayuda en pantalla. | **Host / Búnker** |
