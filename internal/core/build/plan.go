package build

import "context"

// BuildStep represents a single step in a build plan.
// The Exec function is provided by the caller (adapter layer) to execute it.
type BuildStep struct {
	Title  string
	Detail string
	Exec   func(ctx context.Context) error
}

// BuildPlan represents the complete build plan with title, subtitle, and ordered steps.
// The core defines WHAT to do; the adapter decides HOW to execute and render it.
type BuildPlan struct {
	Title    string
	Subtitle string
	Steps    []BuildStep
	// Cleanup is called on failure if the plan did not complete normally.
	Cleanup func()
	// OnSuccess is called after all steps complete successfully.
	OnSuccess func()
}
