// Package components proporciona componentes UI reutilizables para la TUI.
package components

import (
	"github.com/charmbracelet/lipgloss"
)

// CenteredContainer es un contenedor que centra su contenido en la terminal.
// Usa lipgloss.Place para posicionar el contenido horizontal y verticalmente.
type CenteredContainer struct {
	width     int
	height    int
	maxWidth  int
	maxHeight int
}

// NewCenteredContainer crea un nuevo CenteredContainer con las dimensiones especificas.
// Si maxWidth o maxHeight es 0, se usara width/height como maximo.
func NewCenteredContainer(width, height int) *CenteredContainer {
	return &CenteredContainer{
		width:     width,
		height:    height,
		maxWidth:  width,
		maxHeight: height,
	}
}

// WithMaxWidth establece el ancho maximo del contenedor.
// Si es 0, se usara el ancho por defecto.
func (c *CenteredContainer) WithMaxWidth(maxWidth int) *CenteredContainer {
	if maxWidth > 0 {
		c.maxWidth = maxWidth
	}
	return c
}

// WithMaxHeight establece la altura maxima del contenedor.
// Si es 0, se usara la altura por defecto.
func (c *CenteredContainer) WithMaxHeight(maxHeight int) *CenteredContainer {
	if maxHeight > 0 {
		c.maxHeight = maxHeight
	}
	return c
}

// SetDimensions actualiza las dimensiones del contenedor.
func (c *CenteredContainer) SetDimensions(width, height int) {
	c.width = width
	c.height = height
	if c.maxWidth == 0 || c.maxWidth > width {
		c.maxWidth = width
	}
	if c.maxHeight == 0 || c.maxHeight > height {
		c.maxHeight = height
	}
}

// Render centra y renderiza el contenido dentro del contenedor.
// El contenido se coloca en el centro usando lipgloss.Place.
func (c *CenteredContainer) Render(content string) string {
	style := lipgloss.NewStyle().
		Width(c.maxWidth).
		Height(c.maxHeight)

	return style.Render(
		lipgloss.Place(
			c.maxWidth,
			c.maxHeight,
			lipgloss.Center,
			lipgloss.Center,
			content,
		),
	)
}

// RenderWithStyle permite renderizar el contenido con un estilo personalizado
// además del centrado.
func (c *CenteredContainer) RenderWithStyle(content string, customStyle lipgloss.Style) string {
	centered := lipgloss.NewStyle().
		Width(c.maxWidth).
		Height(c.maxHeight)

	// Combina el estilo personalizado con el centrado
	styledContent := customStyle.Render(content)

	return centered.Render(
		lipgloss.Place(
			c.maxWidth,
			c.maxHeight,
			lipgloss.Center,
			lipgloss.Center,
			styledContent,
		),
	)
}
