# AXIOM Bunker System

El sistema de desarrollo definitivo aislado y modular, equipado con herramientas locales de IA y Starship customizado.

## Instalación Rápida

Sigue estos pasos para inicializar tu entorno:

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
Detecta tu GPU automáticamente e instala todo el stack (drivers, herramientas IA, starship) en una imagen `localhost/axiom-[gpu]:latest`. Tarda ~15-30 min pero **solo se ejecuta una vez**.

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