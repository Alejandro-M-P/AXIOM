package ui

import (
	"context"

	"github.com/Alejandro-M-P/AXIOM/internal/core/build"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// Progress tracks and renders build progress for the UI.
// This is the UI adapter implementation of ports.IBuildProgress.
type Progress struct {
	ui          ports.IPresenter
	title       string
	subtitle    string
	steps       []ports.LifecycleStep
	taskTitle   string
	taskSteps   []ports.LifecycleStep
	totalSteps  int
	currentStep int
}

// NewProgress creates a new Progress tracker.
func NewProgress(ui ports.IPresenter, title, subtitle string, steps []ports.LifecycleStep) *Progress {
	return &Progress{
		ui:         ui,
		title:      title,
		subtitle:   subtitle,
		steps:      steps,
		totalSteps: len(steps),
	}
}

// CurrentStep returns the index of the currently running step.
func (p *Progress) CurrentStep() int {
	return p.currentStep
}

// TotalSteps returns the total number of steps.
func (p *Progress) TotalSteps() int {
	return p.totalSteps
}

// StartStep implements ports.IBuildProgress.
// It marks a step as running and updates internal state.
func (p *Progress) StartStep(index int, title string, detail string) {
	p.currentStep = index
	for i := range p.steps {
		if i < index && p.steps[i].Status != ports.LifecycleDone {
			p.steps[i].Status = ports.LifecycleDone
		}
		if i == index {
			p.steps[i].Status = ports.LifecycleRunning
			if title != "" {
				p.steps[i].Title = title
			}
			if detail != "" {
				p.steps[i].Detail = detail
			}
		}
	}
	p.taskTitle = ""
	p.taskSteps = nil
	p.render()
}

// FinishStep implements ports.IBuildProgress.
// It marks the current step as done and renders.
func (p *Progress) FinishStep() {
	if p.currentStep >= 0 && p.currentStep < len(p.steps) {
		p.steps[p.currentStep].Status = ports.LifecycleDone
	}
	p.taskTitle = ""
	p.taskSteps = nil
	p.render()
}

// FailStep implements ports.IBuildProgress.
// It marks the current step as failed with an error.
func (p *Progress) FailStep(err error) {
	if p.currentStep >= 0 && p.currentStep < len(p.steps) {
		p.steps[p.currentStep].Status = ports.LifecycleError
	}
	p.renderErrorWithContext(err, "")
}

// Render implements ports.IBuildProgress.
// It forces a re-render of the current state.
func (p *Progress) Render() {
	p.render()
}

// SetTitle updates the build title and re-renders.
func (p *Progress) SetTitle(title string) {
	p.title = title
	p.render()
}

// SetSubtitle updates the build subtitle and re-renders.
func (p *Progress) SetSubtitle(subtitle string) {
	p.subtitle = subtitle
	p.render()
}

// StartTaskGroup starts a group of sub-tasks within the current step.
func (p *Progress) StartTaskGroup(title string, tasks []ports.LifecycleStep) {
	p.taskTitle = title
	p.taskSteps = tasks
	p.render()
}

// RunTask runs a sub-task within the current task group.
func (p *Progress) RunTask(index int, fn func() error) error {
	for i := range p.taskSteps {
		if i < index && p.taskSteps[i].Status != ports.LifecycleDone {
			p.taskSteps[i].Status = ports.LifecycleDone
		}
		if i == index {
			p.taskSteps[i].Status = ports.LifecycleRunning
		}
	}
	p.render()

	if err := fn(); err != nil {
		if index >= 0 && index < len(p.taskSteps) {
			p.taskSteps[index].Status = ports.LifecycleError
		}
		p.render()
		return err
	}

	if index >= 0 && index < len(p.taskSteps) {
		p.taskSteps[index].Status = ports.LifecycleDone
	}
	p.render()
	return nil
}

// RunStep executes a main build step with progress tracking.
func (p *Progress) RunStep(index int, fn func() error) error {
	p.StartStep(index, "", "")
	if err := fn(); err != nil {
		p.FailStep(err)
		return err
	}
	p.FinishStep()
	return nil
}

// RunStepWithDetail executes a main build step with title and detail.
func (p *Progress) RunStepWithDetail(index int, title, detail string, fn func() error) error {
	p.StartStep(index, title, detail)
	if err := fn(); err != nil {
		p.FailStep(err)
		return err
	}
	p.FinishStep()
	return nil
}

// AppendOutput appends output to the current step's detail (for logging).
func (p *Progress) AppendOutput(line string) {
	// Could be used for logging/debugging in future
}

// render draws the current progress state to the UI.
func (p *Progress) render() {
	p.ui.ClearScreen()
	p.ui.ShowLogo()
	p.ui.RenderLifecycle(p.title, p.subtitle, p.steps, p.taskTitle, p.taskSteps)
}

// RenderError draws the error state to the UI.
func (p *Progress) RenderError(err error) {
	p.renderErrorWithContext(err, "")
}

// renderErrorWithContext draws the error state with context information.
func (p *Progress) renderErrorWithContext(err error, where string) {
	p.ui.ClearScreen()
	p.ui.ShowLogo()
	p.ui.RenderLifecycleError(p.title, p.steps, p.taskTitle, p.taskSteps, err, where)
}

// RunBuildPlan executes a BuildPlan with full progress tracking.
// This is the main execution loop that the adapter calls after receiving a plan from the core.
func RunBuildPlan(ctx context.Context, plan *build.BuildPlan, progress *Progress) error {
	for i, step := range plan.Steps {
		if err := progress.RunStepWithDetail(i, step.Title, step.Detail, func() error {
			return step.Exec(ctx)
		}); err != nil {
			return err
		}
	}
	return nil
}
