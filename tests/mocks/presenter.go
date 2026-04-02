package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// MockPresenter implements ports.IPresenter for testing.
type MockPresenter struct {
	mu sync.Mutex

	// Call tracking
	ShowLogoCalls         int
	ShowCommandCardCalls  []ShowCommandCardCall
	ShowWarningCalls      []ShowWarningCall
	ShowLogCalls          []ShowLogCall
	AskStringCalls        []string
	AskConfirmCalls       []AskConfirmCall
	AskConfirmInCardCalls []AskConfirmInCardCall
	AskDeleteCalls        []AskDeleteCall

	// Responses
	AskStringResponse    string
	AskStringErr         error
	AskConfirmResponse   bool
	AskConfirmErr        error
	AskConfirmInCardResp bool
	AskConfirmInCardErr  error
	AskDeleteConfirm     bool
	AskDeleteReason      string
	AskDeleteDeleteCode  bool
	AskDeleteErr         error
}

// ShowCommandCardCall tracks ShowCommandCard calls.
type ShowCommandCardCall struct {
	CommandKey string
	Fields     []ports.Field
	Items      []string
}

// ShowWarningCall tracks ShowWarning calls.
type ShowWarningCall struct {
	Title    string
	Subtitle string
	Fields   []ports.Field
	Items    []string
	Footer   string
}

// ShowLogCall tracks ShowLog calls.
type ShowLogCall struct {
	LogKey string
	Args   []any
}

// AskConfirmCall tracks AskConfirm calls.
type AskConfirmCall struct {
	PromptKey string
	Args      []any
}

// AskConfirmInCardCall tracks AskConfirmInCard calls.
type AskConfirmInCardCall struct {
	CommandKey string
	Fields     []ports.Field
	Items      []string
	PromptKey  string
}

// AskDeleteCall tracks AskDelete calls.
type AskDeleteCall struct {
	Name   string
	Fields []ports.Field
}

// NewMockPresenter creates a MockPresenter with default values.
func NewMockPresenter() *MockPresenter {
	return &MockPresenter{
		ShowCommandCardCalls:  []ShowCommandCardCall{},
		ShowWarningCalls:      []ShowWarningCall{},
		ShowLogCalls:          []ShowLogCall{},
		AskStringCalls:        []string{},
		AskConfirmCalls:       []AskConfirmCall{},
		AskConfirmInCardCalls: []AskConfirmInCardCall{},
		AskDeleteCalls:        []AskDeleteCall{},
	}
}

// ShowLogo implements ports.IPresenter.
func (m *MockPresenter) ShowLogo() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShowLogoCalls++
}

// ShowCommandCard implements ports.IPresenter.
func (m *MockPresenter) ShowCommandCard(commandKey string, fields []ports.Field, items []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShowCommandCardCalls = append(m.ShowCommandCardCalls, ShowCommandCardCall{
		CommandKey: commandKey,
		Fields:     fields,
		Items:      items,
	})
}

// ShowWarning implements ports.IPresenter.
func (m *MockPresenter) ShowWarning(title, subtitle string, fields []ports.Field, items []string, footer string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShowWarningCalls = append(m.ShowWarningCalls, ShowWarningCall{
		Title:    title,
		Subtitle: subtitle,
		Fields:   fields,
		Items:    items,
		Footer:   footer,
	})
}

// ShowLog implements ports.IPresenter.
func (m *MockPresenter) ShowLog(logKey string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShowLogCalls = append(m.ShowLogCalls, ShowLogCall{
		LogKey: logKey,
		Args:   args,
	})
}

// AskString implements ports.IPresenter.
func (m *MockPresenter) AskString(promptKey string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AskStringCalls = append(m.AskStringCalls, promptKey)
	return m.AskStringResponse, m.AskStringErr
}

// AskConfirm implements ports.IPresenter.
func (m *MockPresenter) AskConfirm(promptKey string, args ...any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AskConfirmCalls = append(m.AskConfirmCalls, AskConfirmCall{
		PromptKey: promptKey,
		Args:      args,
	})
	return m.AskConfirmResponse, m.AskConfirmErr
}

// ShowHelp implements ports.IPresenter.
func (m *MockPresenter) ShowHelp() {}

// GetText implements ports.IPresenter.
func (m *MockPresenter) GetText(key string, args ...any) string {
	return key
}

// ClearScreen implements ports.IPresenter.
func (m *MockPresenter) ClearScreen() {}

// RenderLifecycle implements ports.IPresenter.
func (m *MockPresenter) RenderLifecycle(title, subtitle string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep) {
}

// RenderLifecycleError implements ports.IPresenter.
func (m *MockPresenter) RenderLifecycleError(title string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep, err error, where string) {
}

// AskConfirmInCard implements ports.IPresenter.
func (m *MockPresenter) AskConfirmInCard(commandKey string, fields []ports.Field, items []string, promptKey string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AskConfirmInCardCalls = append(m.AskConfirmInCardCalls, AskConfirmInCardCall{
		CommandKey: commandKey,
		Fields:     fields,
		Items:      items,
		PromptKey:  promptKey,
	})
	return m.AskConfirmInCardResp, m.AskConfirmInCardErr
}

// AskDelete implements ports.IPresenter.
func (m *MockPresenter) AskDelete(name string, fields []ports.Field) (confirm bool, reason string, deleteCode bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AskDeleteCalls = append(m.AskDeleteCalls, AskDeleteCall{
		Name:   name,
		Fields: fields,
	})
	return m.AskDeleteConfirm, m.AskDeleteReason, m.AskDeleteDeleteCode, m.AskDeleteErr
}

// AskReset implements ports.IPresenter.
func (m *MockPresenter) AskReset(fields []ports.Field, items []string) (confirm bool, reason string, err error) {
	return true, "", nil
}

// RunInitWizard implements ports.IPresenter.
func (m *MockPresenter) RunInitWizard(ctx context.Context) error {
	return nil
}

// RunInitWizardResult implements ports.IPresenter.
func (m *MockPresenter) RunInitWizardResult(ctx context.Context) (bool, error) {
	return false, nil
}

// RunInitWizardWithParams implements ports.IPresenter.
func (m *MockPresenter) RunInitWizardWithParams(ctx context.Context, axiomPath string, envExists bool, lang string, homeDir string) (bool, error) {
	return false, nil
}

// AskCreateBunker implements ports.IPresenter.
func (m *MockPresenter) AskCreateBunker(images []string) (name string, image string, confirmed bool, err error) {
	return "test-bunker", "axiom-dev", true, nil
}

// AskSelectBunker implements ports.IPresenter.
func (m *MockPresenter) AskSelectBunker(bunkers []string, statuses map[string]string, title, subtitle string) (selected string, confirmed bool, err error) {
	if len(bunkers) > 0 {
		return bunkers[0], true, nil
	}
	return "", false, nil
}

// RunHelpTUI implements ports.IPresenter.
func (m *MockPresenter) RunHelpTUI() error {
	return nil
}

// GetBunkerVolumeFlags implements ports.IPresenter.
func (m *MockPresenter) GetBunkerVolumeFlags(projectDir, name, aiConfigDir, configPath, gpuType, sshSocket string) (map[string]string, error) {
	return map[string]string{
		"volume_project":   fmt.Sprintf("--volume %s:/%s:z", projectDir, name),
		"volume_ai_config": fmt.Sprintf("--volume %s:/ai_config:z", aiConfigDir),
		"volume_config":    fmt.Sprintf("--volume %s:/run/axiom/env:ro,z", configPath),
	}, nil
}
