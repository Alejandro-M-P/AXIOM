# AXIOM Bunker System

El sistema de desarrollo definitivo aislado y modular, equipado con herramientas locales de IA y Starship customizado.

## Instalación Rápida

Sigue estos 4 pasos para inicializar tu entorno:

1. **Clonar el repositorio:**
   ```bash
   git clone <URL_DEL_REPO> ~/AXIOM
   cd ~/AXIOM
   ```

2. **Ejecutar el instalador:**
   ```bash
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
| `push` | Envía cambios a GitHub utilizando el token securizado del `.env`. | **Búnker** |
| `diagnostico` | Realiza un diagnóstico de salud del búnker (GPU, Ollama, Token Git). | **Búnker** |
| `ayuda` | Muestra el menú de ayuda en pantalla. | **Host / Búnker** |