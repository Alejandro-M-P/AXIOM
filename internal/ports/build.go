package ports

// IBuildProgress defines the port for tracking and rendering build progress.
// The UI layer implements this; the core only signals state changes.
type IBuildProgress interface {
	// StartStep marks a step as running and renders.
	StartStep(index int, title string, detail string)
	// FinishStep marks the current step as done and renders.
	FinishStep()
	// FailStep marks the current step as failed with an error.
	FailStep(err error)
	// Render forces a re-render of the current state.
	Render()
}
