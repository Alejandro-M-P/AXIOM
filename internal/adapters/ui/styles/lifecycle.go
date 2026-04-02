package styles

import (
	"fmt"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
	bubbleprogress "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

const (
	LifecyclePending = "pending"
	LifecycleRunning = "running"
	LifecycleDone    = "done"
	LifecycleError   = "error"
)

type LifecycleStep struct {
	Title  string
	Detail string
	Status string
}

var (
	LifecycleShellStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Cyan).
				Padding(1, 2).
				Margin(1, 2).
				Width(84)

	LifecycleTitleStyle = lipgloss.NewStyle().
				Foreground(White).
				Background(Dark).
				Padding(0, 1).
				Bold(true)

	LifecycleSubtitleStyle = lipgloss.NewStyle().
				Foreground(Gray).
				Italic(true)

	LifecycleSectionStyle = lipgloss.NewStyle().
				Foreground(White).
				Bold(true)

	LifecycleStepPendingStyle = lipgloss.NewStyle().Foreground(Gray)
	LifecycleStepRunningStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	LifecycleStepDoneStyle    = lipgloss.NewStyle().Foreground(Green).Bold(true)
	LifecycleStepErrorStyle   = lipgloss.NewStyle().Foreground(Red).Bold(true)
	LifecycleMetaStyle        = lipgloss.NewStyle().Foreground(White)
	LifecyclePathStyle        = lipgloss.NewStyle().Foreground(Cyan)
)

func RenderLifecycle(title, subtitle string, steps []LifecycleStep) string {
	return RenderLifecycleWithTasks(title, subtitle, steps, "", nil)
}

func RenderLifecycleWithTasks(title, subtitle string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep) string {
	completed := countCompleted(steps)

	var lines []string
	lines = append(lines, LifecycleTitleStyle.Render(title))
	if strings.TrimSpace(subtitle) != "" {
		lines = append(lines, LifecycleSubtitleStyle.Render(subtitle))
	}
	lines = append(lines, "")
	lines = append(lines, renderProgressBar(completed, len(steps)))
	lines = append(lines, LifecycleMetaStyle.Render(i18n.GetLifecycleText("progress", "counter", completed, len(steps))))
	lines = append(lines, "")

	for i, step := range steps {
		label := fmt.Sprintf(i18n.Lifecycle["format"]["numbered_step"], i+1, step.Title)
		if strings.TrimSpace(step.Detail) != "" {
			label += "  " + step.Detail
		}
		lines = append(lines, renderLifecycleStep(label, step.Status))
	}

	if strings.TrimSpace(taskTitle) != "" && len(taskSteps) > 0 {
		taskDone := countCompleted(taskSteps)
		lines = append(lines, "")
		lines = append(lines, LifecycleSectionStyle.Render(taskTitle))
		lines = append(lines, renderProgressBar(taskDone, len(taskSteps)))
		lines = append(lines, LifecycleMetaStyle.Render(i18n.GetLifecycleText("progress", "installation_counter", taskDone, len(taskSteps))))
		for i, step := range taskSteps {
			label := fmt.Sprintf(i18n.Lifecycle["format"]["numbered_step"], i+1, step.Title)
			if strings.TrimSpace(step.Detail) != "" {
				label += "  " + step.Detail
			}
			lines = append(lines, renderLifecycleStep(label, step.Status))
		}
	}

	return LifecycleShellStyle.Render(strings.Join(lines, "\n"))
}

func RenderLifecycleError(title string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep, err error, where string) string {
	var lines []string
	lines = append(lines, RenderLifecycleWithTasks(title, i18n.GetLifecycleText("progress", "error_title"), steps, taskTitle, taskSteps))
	lines = append(lines, lipgloss.NewStyle().Foreground(Red).Bold(true).Render(i18n.GetLifecycleText("progress", "error_label")), err.Error())
	if strings.TrimSpace(where) != "" {
		lines = append(lines, LifecycleMetaStyle.Render(i18n.GetLifecycleText("progress", "location_label")), LifecyclePathStyle.Render(where))
	}
	return strings.Join(lines, "\n")
}

func countCompleted(steps []LifecycleStep) int {
	completed := 0
	for _, step := range steps {
		if step.Status == LifecycleDone {
			completed++
		}
	}
	return completed
}

func renderProgressBar(done, total int) string {
	if total <= 0 {
		total = 1
	}
	percent := float64(done) / float64(total)
	bar := bubbleprogress.New(
		bubbleprogress.WithWidth(28),
		bubbleprogress.WithDefaultGradient(),
	)
	return bar.ViewAs(percent)
}

func renderLifecycleStep(label, status string) string {
	prefix := "[ ]"
	style := LifecycleStepPendingStyle

	switch status {
	case LifecycleRunning:
		prefix = "[>]"
		style = LifecycleStepRunningStyle
	case LifecycleDone:
		prefix = "[x]"
		style = LifecycleStepDoneStyle
	case LifecycleError:
		prefix = "[!]"
		style = LifecycleStepErrorStyle
	}

	return style.Render(prefix + " " + label)
}
