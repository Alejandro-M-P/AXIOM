# AXIOM Bunker System

El sistema de desarrollo definitivo aislado y modular, equipado con herramientas locales de IA y Starship customizado.

## Instalación Rápida

Sigue estos 4 pasos para inicializar tu entorno:

1. **Clonar el repositorio:**
   ```bash
   git clone https://github.com/Alejandro-M-P/AXIOM ~/AXIOM
   cd ~/AXIOM
   ```

2. **Ejecutar el instalador:**
   ```bash
   chmod +x install.sh
./install.sh
   ```
   Sigue el asistente para proveer tus datos de GitHub, tu token y tu directorio de desarrollo base preferido.

3. **Añadir source a tu shell:**
   ```bash
   echo "source ~/AXIOM/axiom.sh" >> ~/.bashrc
   source ~/.bashrc
   ```

4. **Crear tu primer búnker:**
   ```bash
   crear mi-primer-proyecto
   ```

---

## Comandos Disponibles

| Comando | Descripción | Entorno |
| :--- | :--- | :--- |
| `crear [nombre]` | Crea un nuevo búnker o entra en uno existente si ya está configurado. | **Host** |
| `borrar [nombre]` | Solicita razón técnica y destruye el búnker y sus dependencias de memoria local por completo. | **Host** |
| `parar [nombre]` | Detiene la ejecución del contenedor del búnker sin eliminar sus datos. | **Host** |
| `open` | Sincroniza leyes y abre el entorno inteligente `opencode`. | **Búnker** |
| `sync-agents` | Sincroniza la ley global de `tutor.md` a la configuración local del agente. | **Búnker** |
| `save-rule [regla]` | Guarda una nueva regla/razón técnica y la sincroniza con todos los búnkeres. | **Búnker** |
| `git-clone [u/r]` | Clona un repositorio de GitHub utilizando el token securizado en memoria y limpia los origin. | **Búnker** |
| `push` | Envía cambios a GitHub utilizando el token securizado del `.env`. | **Búnker** |
| `diagnostico` | Realiza un diagnóstico de salud del búnker (GPU, Ollama, Token Git). | **Búnker** |
| `ayuda` | Muestra el menú de ayuda en pantalla. | **Host / Búnker** |
