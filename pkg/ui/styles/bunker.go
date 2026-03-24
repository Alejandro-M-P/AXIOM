package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type BunkerDetail struct {
	Label string
	Value string
}

var (
	BunkerCardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Cyan).
		Padding(1, 2).
		Margin(1, 2).
		Width(84)

	BunkerTitleStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(Dark).
		Padding(0, 1).
		Bold(true)

	BunkerSubtitleStyle = lipgloss.NewStyle().
		Foreground(Gray).
		Italic(true)

	BunkerLabelStyle = lipgloss.NewStyle().
		Foreground(White).
		Bold(true)

	BunkerValueStyle = lipgloss.NewStyle().
		Foreground(Cyan)

	BunkerListStyle = lipgloss.NewStyle().
		Foreground(Green)

	BunkerWarningStyle = lipgloss.NewStyle().
		Foreground(Red).
		Bold(true)
)

func RenderBunkerCard(title, subtitle string, details []BunkerDetail, items []string, footer string) string {
	var lines []string
	lines = append(lines, BunkerTitleStyle.Render(title))
	if strings.TrimSpace(subtitle) != "" {
		lines = append(lines, BunkerSubtitleStyle.Render(subtitle))
	}
	if len(details) > 0 {
		lines = append(lines, "")
		for _, detail := range details {
			if strings.TrimSpace(detail.Label) == "" && strings.TrimSpace(detail.Value) == "" {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s %s", BunkerLabelStyle.Render(detail.Label+":"), BunkerValueStyle.Render(detail.Value)))
		}
	}
	if len(items) > 0 {
		lines = append(lines, "")
		lines = append(lines, BunkerLabelStyle.Render("Elementos detectados:"))
		for _, item := range items {
			if strings.TrimSpace(item) == "" {
				continue
			}
			lines = append(lines, BunkerListStyle.Render("• "+item))
		}
	}
	if strings.TrimSpace(footer) != "" {
		lines = append(lines, "")
		lines = append(lines, BunkerSubtitleStyle.Render(footer))
	}
	return BunkerCardStyle.Render(strings.Join(lines, "\n"))
}

func RenderBunkerWarning(title, subtitle string, details []BunkerDetail, items []string, warning string) string {
	body := RenderBunkerCard(title, subtitle, details, items, "")
	return body + "\n" + BunkerWarningStyle.Render(warning)
}
