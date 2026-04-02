package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
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
			Margin(1, 2)

	BunkerDangerCardStyle = BunkerCardStyle.Copy().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(Red)

	BunkerTitleStyle = lipgloss.NewStyle().
				Foreground(Cyan).
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
	return BunkerCardStyle.Render(strings.Join(BuildCardLines(title, subtitle, details, items, footer), "\n"))
}

func BuildCardLines(title, subtitle string, details []BunkerDetail, items []string, footer string) []string {
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
		lines = append(lines, BunkerLabelStyle.Render(i18n.GetSlotsText("bunker.list", "elements_detected")))
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
	return lines
}

func RenderBunkerWarning(title, subtitle string, details []BunkerDetail, items []string, warning string) string {
	body := RenderBunkerCard(title, subtitle, details, items, "")
	return body + "\n" + BunkerWarningStyle.Render(warning)
}

type BunkerRow struct {
	Name        string
	Status      string
	Size        string
	LastEntry   string
	GitBranch   string
	Image       string
	GPU         string
	ProjectPath string
	EnvPath     string
}

func RenderBunkerList(title, subtitle string, rows []BunkerRow, footer string) string {
	const (
		nameWidth   = 18
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
		BunkerLabelStyle.Render(padRight(i18n.GetSlotsText("bunker.list", "header_bunker"), nameWidth)),
		BunkerLabelStyle.Render(padRight(i18n.GetSlotsText("bunker.list", "header_status"), statusWidth)),
		BunkerLabelStyle.Render(padRight(i18n.GetSlotsText("bunker.list", "header_size"), sizeWidth)),
		BunkerLabelStyle.Render(padRight(i18n.GetSlotsText("bunker.list", "header_last"), dateWidth)),
		BunkerLabelStyle.Render(padRight(i18n.GetSlotsText("bunker.list", "header_branch"), branchWidth)),
	))

	for _, row := range rows {
		statusText := i18n.GetSlotsText("bunker.list", "status_stopped")
		statusStyle := lipgloss.NewStyle().Foreground(Gray)
		if strings.TrimSpace(row.Status) == "running" {
			statusText = i18n.GetSlotsText("bunker.list", "status_running")
			statusStyle = lipgloss.NewStyle().Foreground(Green).Bold(true)
		}

		branch := fallbackValue(row.GitBranch, "-")
		lastEntry := fallbackValue(row.LastEntry, "-")
		size := fallbackValue(row.Size, "-")

		lines = append(lines, renderBunkerListRow(
			lipgloss.NewStyle().Foreground(White).Bold(true).Render(padRight(row.Name, nameWidth)),
			statusStyle.Render(padRight(statusText, statusWidth)),
			BunkerValueStyle.Render(padRight(size, sizeWidth)),
			BunkerValueStyle.Render(padRight(lastEntry, dateWidth)),
			BunkerValueStyle.Render(padRight(branch, branchWidth)),
		))

		imgLabel := i18n.GetSlotsText("bunker.list", "label_image")
		gpuLabel := i18n.GetSlotsText("bunker.list", "label_gpu")
		envLabel := i18n.GetSlotsText("bunker.list", "label_env")
		projectLabel := i18n.GetSlotsText("bunker.list", "label_project")
		meta := []string{
			fmt.Sprintf("%s %s", imgLabel, fallbackValue(row.Image, "-")),
			fmt.Sprintf("%s %s", gpuLabel, fallbackValue(row.GPU, "-")),
			fmt.Sprintf("%s %s", envLabel, truncateText(fallbackValue(row.EnvPath, "-"), 34)),
			fmt.Sprintf("%s %s", projectLabel, truncateText(fallbackValue(row.ProjectPath, "-"), 34)),
		}
		lines = append(lines, BunkerSubtitleStyle.Render("  "+strings.Join(meta, "  |  ")))
		lines = append(lines, "")
	}

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
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

func truncateText(value string, width int) string {
	trimmed := strings.TrimSpace(value)
	if lipgloss.Width(trimmed) <= width {
		return trimmed
	}
	if width <= 1 {
		return lipgloss.NewStyle().MaxWidth(width).Render(trimmed)
	}
	return lipgloss.NewStyle().MaxWidth(width-1).Render(trimmed) + "..."
}

func fallbackValue(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
