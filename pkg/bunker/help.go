package bunker

import (
	"fmt"

	"axiom/pkg/ui/styles"
)

// Help muestra los comandos disponibles del orquestador bunker.
func (m *Manager) Help() error {
	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderBunkerCard(
		"AXIOM Help",
		"Comandos disponibles del búnker.",
		[]styles.BunkerDetail{
			{Label: "axiom build", Value: "Construye la imagen base con GPU y herramientas IA."},
			{Label: "axiom list", Value: "Abre un selector con búsqueda y muestra la ficha del búnker."},
			{Label: "axiom create <nombre>", Value: "Crea un nuevo búnker o entra en uno existente."},
			{Label: "axiom stop", Value: "Abre el selector de búnkeres activos y detiene el que elijas."},
			{Label: "axiom delete", Value: "Abre el selector de búnkeres y elimina el que elijas."},
			{Label: "axiom delete-image", Value: "Elimina la imagen base activa."},
			{Label: "axiom info", Value: "Muestra la ficha detallada de un búnker."},
			{Label: "axiom prune", Value: "Limpia entornos huérfanos sin contenedor."},
			{Label: "axiom rebuild", Value: "Reconstruye la imagen base."},
			{Label: "axiom reset", Value: "Elimina TODOS los búnkeres e imágenes."},
		},
		[]string{
			"Si no pasas nombre en axiom delete, se abre un selector con flechas.",
			"Al borrar un búnker puedes decidir si también se elimina el código del proyecto.",
			"axiom list ya muestra estado, imagen base, GPU y rutas clave de cada búnker.",
		},
		"",
	))
	return nil
}
