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

type BunkerRow struct {
	Name      string
	Status    string
	Size      string
	LastEntry string
	GitBranch string
}

func RenderBunkerList(title, subtitle string, rows []BunkerRow, footer string) string {
	const (
		nameWidth   = 20
		statusWidth = 10
		sizeWidth   = 8
		dateWidth   = 12
		branchWidth = 14
	)

	var lines []string
	lines = append(lines, BunkerTitleStyle.Render(title))
	if strings.TrimSpace(subtitle) != "" {
		lines = append(lines, BunkerSubtitleStyle.Render(subtitle))
	}
	lines = append(lines, "")
	lines = append(lines, renderBunkerListRow(
		BunkerLabelStyle.Render(padRight("BUNKER", nameWidth)),
		BunkerLabelStyle.Render(padRight("ESTADO", statusWidth)),
		BunkerLabelStyle.Render(padRight("TAM", sizeWidth)),
		BunkerLabelStyle.Render(padRight("ULTIMA", dateWidth)),
		BunkerLabelStyle.Render(padRight("RAMA", branchWidth)),
	))

	for _, row := range rows {
		statusText := "stopped"
		statusStyle := lipgloss.NewStyle().Foreground(Gray)
		if strings.TrimSpace(row.Status) == "running" {
			statusText = "running"
			statusStyle = lipgloss.NewStyle().Foreground(Green).Bold(true)
		}

		branch := strings.TrimSpace(row.GitBranch)
		if branch == "" {
			branch = "-"
		}
		lastEntry := strings.TrimSpace(row.LastEntry)
		if lastEntry == "" {
			lastEntry = "-"
		}
		size := strings.TrimSpace(row.Size)
		if size == "" {
			size = "-"
		}

		lines = append(lines, renderBunkerListRow(
			lipgloss.NewStyle().Foreground(White).Bold(true).Render(padRight(row.Name, nameWidth)),
			statusStyle.Render(padRight(statusText, statusWidth)),
			BunkerValueStyle.Render(padRight(size, sizeWidth)),
			BunkerValueStyle.Render(padRight(lastEntry, dateWidth)),
			BunkerValueStyle.Render(padRight(branch, branchWidth)),
		))
	}

	if strings.TrimSpace(footer) != "" {
		lines = append(lines, "")
		lines = append(lines, BunkerSubtitleStyle.Render(footer))
	}
	return BunkerCardStyle.Render(strings.Join(lines, "\n"))
}

func renderBunkerListRow(columns ...string) string {
	return strings.Join(columns, "  ")
}

func padRight(value string, width int) string {
	trimmed := strings.TrimSpace(value)
	if lipgloss.Width(trimmed) >= width {
		return lipgloss.NewStyle().MaxWidth(width).Render(trimmed)
	}
	return trimmed + strings.Repeat(" ", width-lipgloss.Width(trimmed))
}
