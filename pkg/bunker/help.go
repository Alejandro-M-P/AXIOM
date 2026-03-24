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
		"Comandos disponibles del búnker en la versión Go actual.",
		[]styles.BunkerDetail{
			{Label: "Build", Value: "Construye la imagen base: axiom build"},
			{Label: "Create", Value: "Crea o abre un búnker: axiom create <nombre>"},
			{Label: "Delete", Value: "Elimina un búnker por nombre o selector: axiom delete [nombre]"},
			{Label: "Eliminar", Value: "Alias en español de delete: axiom eliminar [nombre]"},
			{Label: "Delete Image", Value: "Elimina la imagen base activa: axiom delete-image"},
		},
		[]string{
			"Si no pasas nombre en delete/eliminar, se abre un selector con flechas.",
			"Al borrar un búnker puedes decidir si también se elimina el código del proyecto.",
			"delete-image también muestra las imágenes de AXIOM detectadas antes y después.",
		},
		"Siguiente paso: seguir portando list, info, stop, rebuild, prune y reset.",
	))
	return nil
}
