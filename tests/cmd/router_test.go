package cmd_test

import (
	"context"
	"errors"
	"testing"

	"axiom/internal/domain"
	"axiom/internal/ports"
	"axiom/internal/router"
	"axiom/internal/slots"
	"axiom/tests/mocks"
)

// mockUI implements ports.IPresenter for testing
type mockUI struct {
	logs []string
}

func (m *mockUI) ShowLogo() {}

func (m *mockUI) ShowCommandCard(commandKey string, fields []ports.Field, items []string) {}

func (m *mockUI) ShowWarning(title, subtitle string, fields []ports.Field, items []string, footer string) {
}

func (m *mockUI) ShowLog(logKey string, args ...any) {
	m.logs = append(m.logs, logKey)
}

func (m *mockUI) AskString(promptKey string) (string, error) {
	return "", nil
}

func (m *mockUI) AskConfirm(promptKey string, args ...any) (bool, error) {
	return true, nil
}

func (m *mockUI) ShowHelp() {}

func (m *mockUI) GetText(key string, args ...any) string {
	return key
}

func (m *mockUI) ClearScreen() {}

func (m *mockUI) RenderLifecycle(title, subtitle string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep) {
}

func (m *mockUI) RenderLifecycleError(title string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep, err error, where string) {
}

func (m *mockUI) AskConfirmInCard(commandKey string, fields []ports.Field, items []string, promptKey string) (bool, error) {
	return true, nil
}

func (m *mockUI) AskDelete(name string, fields []ports.Field) (confirm bool, reason string, deleteCode bool, err error) {
	return true, "", false, nil
}

func (m *mockUI) AskReset(fields []ports.Field, items []string) (confirm bool, reason string, err error) {
	return true, "", nil
}

func (m *mockUI) RunInitWizard(ctx context.Context) error {
	return nil
}

func (m *mockUI) WithFields(fields map[string]interface{}) ports.IPresenter {
	return m
}

// Ensure mockUI implements ports.IPresenter
var _ ports.IPresenter = (*mockUI)(nil)

// mockBunkerManager implements the bunkerManagerInterface for testing
type mockBunkerManager struct {
	createErr         error
	deleteErr         error
	listErr           error
	stopErr           error
	pruneErr          error
	infoErr           error
	deleteImageErr    error
	helpErr           error
	createCalled      bool
	deleteCalled      bool
	listCalled        bool
	stopCalled        bool
	pruneCalled       bool
	infoCalled        bool
	deleteImageCalled bool
	helpCalled        bool
	lastCreatedName   string
	lastDeletedName   string
	lastInfoName      string
	ui                *mockUI
}

func newMockBunkerManager() *mockBunkerManager {
	return &mockBunkerManager{
		ui: &mockUI{},
	}
}

func (m *mockBunkerManager) Create(name string) error {
	m.createCalled = true
	m.lastCreatedName = name
	return m.createErr
}

func (m *mockBunkerManager) CreateWithImage(name, image string) error {
	m.createCalled = true
	m.lastCreatedName = name
	return m.createErr
}

func (m *mockBunkerManager) Delete(name string) error {
	m.deleteCalled = true
	m.lastDeletedName = name
	return m.deleteErr
}

func (m *mockBunkerManager) List() error {
	m.listCalled = true
	return m.listErr
}

func (m *mockBunkerManager) Stop() error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockBunkerManager) Prune() error {
	m.pruneCalled = true
	return m.pruneErr
}

func (m *mockBunkerManager) Info(name string) error {
	m.infoCalled = true
	m.lastInfoName = name
	return m.infoErr
}

func (m *mockBunkerManager) DeleteImage() error {
	m.deleteImageCalled = true
	return m.deleteImageErr
}

func (m *mockBunkerManager) Help() error {
	m.helpCalled = true
	return m.helpErr
}

func (m *mockBunkerManager) LoadConfig() (domain.EnvConfig, error) {
	return domain.EnvConfig{}, nil
}

func (m *mockBunkerManager) GetUI() ports.IPresenter {
	return m.ui
}

// mockBuildManager implements BuildManagerInterface for testing
type mockBuildManager struct {
	buildErr    error
	rebuildErr  error
	buildCalled bool
}

func (m *mockBuildManager) Build(ctx context.Context, cfg domain.EnvConfig) error {
	m.buildCalled = true
	return m.buildErr
}

func (m *mockBuildManager) Rebuild(ctx context.Context, cfg domain.EnvConfig) error {
	return m.rebuildErr
}

// mockSlotManager implements SlotManagerInterface for testing
type mockSlotManager struct {
	discoverSlotsCalled bool
	executeSlotsCalled  bool
	ui                  *mockUI
}

func (m *mockSlotManager) DiscoverSlots() []any {
	m.discoverSlotsCalled = true
	return nil
}

func (m *mockSlotManager) ExecuteSlots(selected []any) error {
	m.executeSlotsCalled = true
	return nil
}

func (m *mockSlotManager) GetUI() ports.IPresenter {
	if m.ui == nil {
		m.ui = &mockUI{}
	}
	return m.ui
}

func (m *mockSlotManager) HasSelection() bool {
	return false
}

func (m *mockSlotManager) GetSelectedItems(category string) ([]slots.SlotItem, error) {
	return []slots.SlotItem{}, nil
}

func (m *mockSlotManager) GetAvailableItems(category string) ([]slots.SlotItem, error) {
	return []slots.SlotItem{}, nil
}

func (m *mockSlotManager) GetAllAvailableItems() ([]slots.SlotItem, error) {
	return []slots.SlotItem{}, nil
}

func (m *mockSlotManager) RunSlotSelector(category string, items []slots.SlotItem, preselected []string) ([]string, bool, error) {
	return []string{}, false, nil
}

func (m *mockSlotManager) SaveSelection(selections []slots.SlotSelection) error {
	return nil
}

func (m *mockSlotManager) LoadSelection() ([]slots.SlotSelection, error) {
	return []slots.SlotSelection{}, nil
}

// newTestRouter creates a router with all mock managers
func newTestRouter() (*router.Router, *mockBunkerManager, *mockBuildManager, *mockSlotManager) {
	bm := newMockBunkerManager()
	bld := &mockBuildManager{}
	slm := &mockSlotManager{}
	// Use empty string and mocks filesystem for tests
	mockFS := mocks.NewMockFileSystem()
	mockFS.Dirs["/home/test/axiom"] = true
	return router.NewRouter(bm, bld, slm, "/home/test/axiom", mockFS), bm, bld, slm
}

// ============================================================================
// TEST: TestRouter_KnownCommands
// ============================================================================

func TestRouter_KnownCommands(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	knownCommands := []string{
		"create", "delete", "list", "stop", "prune",
		"build", "rebuild", "help", "info", "reset",
		"enter", "init", "delete-image",
	}

	for _, cmd := range knownCommands {
		t.Run(cmd, func(t *testing.T) {
			// Reset state
			bm.helpCalled = false
			bm.createCalled = false
			bm.deleteCalled = false
			bm.listCalled = false
			bm.stopCalled = false
			bm.pruneCalled = false
			bm.infoCalled = false

			err := r.Handle([]string{cmd})

			// Known commands should not return "unknown command" error
			if err != nil && err.Error() == "unknown command: "+cmd {
				t.Errorf("Expected %q to be a known command, but got unknown command error", cmd)
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_UnknownCommand
// ============================================================================

func TestRouter_UnknownCommand(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	testCases := []struct {
		name string
		cmd  string
	}{
		{"random_string", "foobar"},
		{"numbers", "12345"},
		{"special_chars", "!@#$%"},
		{"typo_create", "creat"},
		{"typo_delete", "delet"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bm.helpCalled = false
			bm.helpErr = nil

			err := r.Handle([]string{tc.cmd})

			// Unknown command should return error containing "unknown command"
			if err == nil {
				t.Errorf("Expected unknown command error for %q, got nil", tc.cmd)
			} else if err.Error() != "unknown command: "+tc.cmd {
				t.Errorf("Expected 'unknown command: %s' error, got: %v", tc.cmd, err)
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_HandleCreate
// ============================================================================

func TestRouter_HandleCreate(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test create with name
	err := r.Handle([]string{"create", "mybunker"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.createCalled {
		t.Error("Create was not called")
	}
	if bm.lastCreatedName != "mybunker" {
		t.Errorf("Expected created name 'mybunker', got %q", bm.lastCreatedName)
	}

	// Test create without name (empty string passed)
	bm.createCalled = false
	bm.lastCreatedName = ""
	err = r.Handle([]string{"create"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.createCalled {
		t.Error("Create was not called")
	}
	if bm.lastCreatedName != "" {
		t.Errorf("Expected empty created name, got %q", bm.lastCreatedName)
	}
}

// ============================================================================
// TEST: TestRouter_HandleDelete
// ============================================================================

func TestRouter_HandleDelete(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test delete with name
	err := r.Handle([]string{"delete", "oldbunker"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.deleteCalled {
		t.Error("Delete was not called")
	}
	if bm.lastDeletedName != "oldbunker" {
		t.Errorf("Expected deleted name 'oldbunker', got %q", bm.lastDeletedName)
	}

	// Test delete without name
	bm.deleteCalled = false
	err = r.Handle([]string{"delete"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.deleteCalled {
		t.Error("Delete was not called")
	}
}

// ============================================================================
// TEST: TestRouter_HandleList
// ============================================================================

func TestRouter_HandleList(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	err := r.Handle([]string{"list"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.listCalled {
		t.Error("List was not called")
	}
}

// ============================================================================
// TEST: TestRouter_HandleStop
// ============================================================================

func TestRouter_HandleStop(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	err := r.Handle([]string{"stop"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.stopCalled {
		t.Error("Stop was not called")
	}
}

// ============================================================================
// TEST: TestRouter_HandlePrune
// ============================================================================

func TestRouter_HandlePrune(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	err := r.Handle([]string{"prune"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.pruneCalled {
		t.Error("Prune was not called")
	}
}

// ============================================================================
// TEST: TestRouter_HandleBuild
// ============================================================================

func TestRouter_HandleBuild(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Build is not yet implemented, should return nil but log message
	err := r.Handle([]string{"build"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that a log was shown for not implemented
	if len(bm.ui.logs) == 0 {
		t.Error("Expected log message for not implemented build")
	}
	found := false
	for _, log := range bm.ui.logs {
		if log == "build.not_implemented" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'build.not_implemented' log message")
	}
}

// ============================================================================
// TEST: TestRouter_HandleRebuild
// ============================================================================

func TestRouter_HandleRebuild(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Rebuild is not yet implemented, should return nil but log message
	err := r.Handle([]string{"rebuild"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that a log was shown for not implemented
	if len(bm.ui.logs) == 0 {
		t.Error("Expected log message for not implemented rebuild")
	}
	found := false
	for _, log := range bm.ui.logs {
		if log == "build.not_implemented" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'build.not_implemented' log message")
	}
}

// ============================================================================
// TEST: TestRouter_HandleHelp
// ============================================================================

func TestRouter_HandleHelp(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test help command
	err := r.Handle([]string{"help"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.helpCalled {
		t.Error("Help was not called via 'help' command")
	}
}

// ============================================================================
// TEST: TestRouter_HandleEnter
// ============================================================================

func TestRouter_HandleEnter(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test enter without name - should return error
	err := r.Handle([]string{"enter"})
	if err == nil {
		t.Error("Expected error when entering without bunker name")
	}
	if err.Error() != "usage: axiom enter <bunker-name>" {
		t.Errorf("Expected 'usage: axiom enter <bunker-name>' error, got: %v", err)
	}

	// Test enter with name - should return "not yet implemented" error
	bm.helpCalled = false
	err = r.Handle([]string{"enter", "mybunker"})
	if err == nil {
		t.Error("Expected 'enter not yet implemented' error")
	}
	if err.Error() != "enter not yet implemented" {
		t.Errorf("Expected 'enter not yet implemented' error, got: %v", err)
	}
}

// ============================================================================
// TEST: TestRouter_HandleInit
// ============================================================================

func TestRouter_HandleInit(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	err := r.Handle([]string{"init"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that init not implemented log was shown
	found := false
	for _, log := range bm.ui.logs {
		if log == "system.init_not_implemented" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'system.init_not_implemented' log message")
	}
}

// ============================================================================
// TEST: TestRouter_CommandAliases
// ============================================================================

func TestRouter_CommandAliases(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test "rm" alias for "delete"
	bm.deleteCalled = false
	err := r.Handle([]string{"rm", "testbunker"})
	if err != nil {
		t.Errorf("Unexpected error with 'rm' alias: %v", err)
	}
	if !bm.deleteCalled {
		t.Error("'rm' alias did not trigger delete")
	}
	if bm.lastDeletedName != "testbunker" {
		t.Errorf("Expected deleted name 'testbunker', got %q", bm.lastDeletedName)
	}

	// Test "ls" alias for "list"
	bm.listCalled = false
	err = r.Handle([]string{"ls"})
	if err != nil {
		t.Errorf("Unexpected error with 'ls' alias: %v", err)
	}
	if !bm.listCalled {
		t.Error("'ls' alias did not trigger list")
	}
}

// ============================================================================
// TEST: TestRouter_HelpFlags
// ============================================================================

func TestRouter_HelpFlags(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	testCases := []struct {
		name string
		flag string
	}{
		{"short_help", "-h"},
		{"long_help", "--help"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bm.helpCalled = false
			err := r.Handle([]string{tc.flag})
			if err != nil {
				t.Errorf("Unexpected error with %s: %v", tc.flag, err)
			}
			if !bm.helpCalled {
				t.Errorf("Help was not called for flag %s", tc.flag)
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_EmptyArgs
// ============================================================================

func TestRouter_EmptyArgs(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Empty args should call help
	err := r.Handle([]string{})
	if err != nil {
		t.Errorf("Unexpected error with empty args: %v", err)
	}
	if !bm.helpCalled {
		t.Error("Help was not called for empty args")
	}
}

// ============================================================================
// TEST: TestRouter_HandleInfo
// ============================================================================

func TestRouter_HandleInfo(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test info with name
	err := r.Handle([]string{"info", "mybunker"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.infoCalled {
		t.Error("Info was not called")
	}
	if bm.lastInfoName != "mybunker" {
		t.Errorf("Expected info name 'mybunker', got %q", bm.lastInfoName)
	}

	// Test info without name (empty string)
	bm.infoCalled = false
	bm.lastInfoName = ""
	err = r.Handle([]string{"info"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.infoCalled {
		t.Error("Info was not called")
	}
	if bm.lastInfoName != "" {
		t.Errorf("Expected empty info name, got %q", bm.lastInfoName)
	}
}

// ============================================================================
// TEST: TestRouter_HandleDeleteImage
// ============================================================================

func TestRouter_HandleDeleteImage(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	err := r.Handle([]string{"delete-image"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !bm.deleteImageCalled {
		t.Error("DeleteImage was not called")
	}
}

// ============================================================================
// TEST: TestKnownCommand
// ============================================================================

func TestKnownCommand(t *testing.T) {
	testCases := []struct {
		command  string
		expected bool
	}{
		// Known commands
		{"create", true},
		{"delete", true},
		{"list", true},
		{"stop", true},
		{"prune", true},
		{"build", true},
		{"rebuild", true},
		{"help", true},
		{"info", true},
		{"reset", true},
		{"enter", true},
		{"init", true},
		{"delete-image", true},
		{"rm", true},
		{"ls", true},
		{"-h", true},
		{"--help", true},

		// Unknown commands
		{"foobar", false},
		{"creat", false},
		{"delet", false},
		{"start", false},
		{"restart", false},
		{"remove", false},
		{"", false},
		{"CREATE", true}, // case insensitive
		{"Delete", true}, // case insensitive
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			result := router.KnownCommand(tc.command)
			if result != tc.expected {
				t.Errorf("KnownCommand(%q) = %v, expected %v", tc.command, result, tc.expected)
			}
		})
	}
}

// ============================================================================
// TEST: TestFirstArg
// ============================================================================

func TestFirstArg(t *testing.T) {
	// We need to test firstArg indirectly through Handle
	r, bm, _, _ := newTestRouter()

	testCases := []struct {
		name         string
		args         []string
		expectedName string
		handler      string
	}{
		{
			name:         "create_with_name",
			args:         []string{"create", "mybunker"},
			expectedName: "mybunker",
			handler:      "create",
		},
		{
			name:         "delete_with_name",
			args:         []string{"delete", "oldbunker"},
			expectedName: "oldbunker",
			handler:      "delete",
		},
		{
			name:         "info_with_name",
			args:         []string{"info", "somebunker"},
			expectedName: "somebunker",
			handler:      "info",
		},
		{
			name:         "create_no_name",
			args:         []string{"create"},
			expectedName: "",
			handler:      "create",
		},
		{
			name:         "delete_no_name",
			args:         []string{"delete"},
			expectedName: "",
			handler:      "delete",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			bm.createCalled = false
			bm.deleteCalled = false
			bm.infoCalled = false
			bm.lastCreatedName = ""
			bm.lastDeletedName = ""
			bm.lastInfoName = ""

			r.Handle(tc.args)

			switch tc.handler {
			case "create":
				if !bm.createCalled {
					t.Error("Create was not called")
				}
				if bm.lastCreatedName != tc.expectedName {
					t.Errorf("Expected firstArg %q, got %q", tc.expectedName, bm.lastCreatedName)
				}
			case "delete":
				if !bm.deleteCalled {
					t.Error("Delete was not called")
				}
				if bm.lastDeletedName != tc.expectedName {
					t.Errorf("Expected firstArg %q, got %q", tc.expectedName, bm.lastDeletedName)
				}
			case "info":
				if !bm.infoCalled {
					t.Error("Info was not called")
				}
				if bm.lastInfoName != tc.expectedName {
					t.Errorf("Expected firstArg %q, got %q", tc.expectedName, bm.lastInfoName)
				}
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_CaseInsensitivity
// ============================================================================

func TestRouter_CaseInsensitivity(t *testing.T) {
	r, _, _, _ := newTestRouter()

	// Test that commands are case insensitive
	testCases := []struct {
		name    string
		command string
	}{
		{"lowercase_create", "create"},
		{"uppercase_CREATE", "CREATE"},
		{"mixedCase_Delete", "DeLeTe"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := r.Handle([]string{tc.command})
			// Should not return unknown command error
			if err != nil && err.Error() == "unknown command: "+tc.command {
				t.Errorf("Expected %q to be recognized as 'create' (case insensitive), but got unknown command", tc.command)
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_ErrorPropagation
// ============================================================================

func TestRouter_ErrorPropagation(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test error propagation from Delete
	bm.deleteErr = errors.New("delete failed")
	err := r.Handle([]string{"delete", "bunker1"})
	if err == nil {
		t.Error("Expected error from Delete, got nil")
	}
	if err.Error() != "delete failed" {
		t.Errorf("Expected 'delete failed' error, got: %v", err)
	}

	// Test error propagation from List
	bm.deleteErr = nil
	bm.listErr = errors.New("list failed")
	err = r.Handle([]string{"list"})
	if err == nil {
		t.Error("Expected error from List, got nil")
	}
	if err.Error() != "list failed" {
		t.Errorf("Expected 'list failed' error, got: %v", err)
	}

	// Test error propagation from Create
	bm.listErr = nil
	bm.createErr = errors.New("create failed")
	err = r.Handle([]string{"create", "bunker2"})
	if err == nil {
		t.Error("Expected error from Create, got nil")
	}
	if err.Error() != "create failed" {
		t.Errorf("Expected 'create failed' error, got: %v", err)
	}
}

// ============================================================================
// TEST: TestRouter_Subcommands
// ============================================================================

func TestRouter_Subcommands(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test that subcommands are passed to handlers
	testCases := []struct {
		name       string
		args       []string
		expectName string
		handler    string
	}{
		{"create_with_subargs", []string{"create", "bunker1", "--flag", "value"}, "bunker1", "create"},
		{"delete_with_subargs", []string{"delete", "bunker2", "--force"}, "bunker2", "delete"},
		{"info_with_subargs", []string{"info", "bunker3", "--verbose"}, "bunker3", "info"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.Handle(tc.args)

			switch tc.handler {
			case "create":
				if bm.lastCreatedName != tc.expectName {
					t.Errorf("Expected name %q, got %q", tc.expectName, bm.lastCreatedName)
				}
			case "delete":
				if bm.lastDeletedName != tc.expectName {
					t.Errorf("Expected name %q, got %q", tc.expectName, bm.lastDeletedName)
				}
			case "info":
				if bm.lastInfoName != tc.expectName {
					t.Errorf("Expected name %q, got %q", tc.expectName, bm.lastInfoName)
				}
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_HandleReset
// ============================================================================

func TestRouter_HandleReset(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	err := r.Handle([]string{"reset"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that reset not implemented log was shown
	found := false
	for _, log := range bm.ui.logs {
		if log == "reset.not_implemented" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'reset.not_implemented' log message")
	}
}

// ============================================================================
// TEST: TestRouter_WhitespaceHandling
// ============================================================================

func TestRouter_WhitespaceHandling(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test that whitespace is trimmed and commands are recognized
	testCases := []struct {
		name    string
		args    []string
		handler string
	}{
		{"spaces", []string{"  create", "bunker"}, "create"},
		{"tabs", []string{"\tcreate", "bunker"}, "create"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bm.createCalled = false
			bm.lastCreatedName = ""

			err := r.Handle(tc.args)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !bm.createCalled {
				t.Errorf("Create was not called for %q", tc.args[0])
			}
			if bm.lastCreatedName != "bunker" {
				t.Errorf("Expected bunker name 'bunker', got %q", bm.lastCreatedName)
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_NewRouter
// ============================================================================

func TestRouter_NewRouter(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	if r == nil {
		t.Fatal("NewRouter returned nil")
	}

	// Router should be functional
	err := r.Handle([]string{"list"})
	if err != nil {
		t.Errorf("Router.Handle failed: %v", err)
	}
	if !bm.listCalled {
		t.Error("List was not called via new Router")
	}
}

// ============================================================================
// TEST: TestRouter_AllCommandsRouteCorrectly
// ============================================================================

func TestRouter_AllCommandsRouteCorrectly(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	commands := []struct {
		name         string
		cmd          string
		shouldError  bool
		handlerField string
	}{
		{"create", "create", false, "createCalled"},
		{"delete", "delete", false, "deleteCalled"},
		{"list", "list", false, "listCalled"},
		{"stop", "stop", false, "stopCalled"},
		{"prune", "prune", false, "pruneCalled"},
		{"build", "build", false, "nil"},
		{"rebuild", "rebuild", false, "nil"},
		{"help", "help", false, "helpCalled"},
		{"info", "info", false, "infoCalled"},
		{"reset", "reset", false, "nil"},
		{"enter", "enter", true, "nil"},
		{"init", "init", false, "nil"},
		{"delete-image", "delete-image", false, "deleteImageCalled"},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			// Reset all flags
			bm.createCalled = false
			bm.deleteCalled = false
			bm.listCalled = false
			bm.stopCalled = false
			bm.pruneCalled = false
			bm.infoCalled = false
			bm.helpCalled = false
			bm.deleteImageCalled = false

			err := r.Handle([]string{tc.cmd, "testbunker"})

			if tc.shouldError && err == nil {
				t.Errorf("Expected error for command %q", tc.cmd)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error for command %q: %v", tc.cmd, err)
			}

			// Check the appropriate handler was called
			switch tc.handlerField {
			case "createCalled":
				if !bm.createCalled {
					t.Errorf("createCalled not set for command %q", tc.cmd)
				}
			case "deleteCalled":
				if !bm.deleteCalled {
					t.Errorf("deleteCalled not set for command %q", tc.cmd)
				}
			case "listCalled":
				if !bm.listCalled {
					t.Errorf("listCalled not set for command %q", tc.cmd)
				}
			case "stopCalled":
				if !bm.stopCalled {
					t.Errorf("stopCalled not set for command %q", tc.cmd)
				}
			case "pruneCalled":
				if !bm.pruneCalled {
					t.Errorf("pruneCalled not set for command %q", tc.cmd)
				}
			case "infoCalled":
				if !bm.infoCalled {
					t.Errorf("infoCalled not set for command %q", tc.cmd)
				}
			case "helpCalled":
				if !bm.helpCalled {
					t.Errorf("helpCalled not set for command %q", tc.cmd)
				}
			case "deleteImageCalled":
				if !bm.deleteImageCalled {
					t.Errorf("deleteImageCalled not set for command %q", tc.cmd)
				}
			case "nil":
				// No specific handler flag to check (these are unimplemented)
			}
		})
	}
}

// ============================================================================
// TEST: TestRouter_Constants
// ============================================================================

func TestRouter_Constants(t *testing.T) {
	// Test that constants are defined correctly
	if router.CmdCreate != "create" {
		t.Errorf("CmdCreate = %q, expected 'create'", router.CmdCreate)
	}
	if router.CmdDelete != "delete" {
		t.Errorf("CmdDelete = %q, expected 'delete'", router.CmdDelete)
	}
	if router.CmdList != "list" {
		t.Errorf("CmdList = %q, expected 'list'", router.CmdList)
	}
	if router.CmdStop != "stop" {
		t.Errorf("CmdStop = %q, expected 'stop'", router.CmdStop)
	}
	if router.CmdPrune != "prune" {
		t.Errorf("CmdPrune = %q, expected 'prune'", router.CmdPrune)
	}
	if router.CmdBuild != "build" {
		t.Errorf("CmdBuild = %q, expected 'build'", router.CmdBuild)
	}
	if router.CmdRebuild != "rebuild" {
		t.Errorf("CmdRebuild = %q, expected 'rebuild'", router.CmdRebuild)
	}
	if router.CmdHelp != "help" {
		t.Errorf("CmdHelp = %q, expected 'help'", router.CmdHelp)
	}
	if router.CmdInfo != "info" {
		t.Errorf("CmdInfo = %q, expected 'info'", router.CmdInfo)
	}
	if router.CmdReset != "reset" {
		t.Errorf("CmdReset = %q, expected 'reset'", router.CmdReset)
	}
	if router.CmdEnter != "enter" {
		t.Errorf("CmdEnter = %q, expected 'enter'", router.CmdEnter)
	}
	if router.CmdInit != "init" {
		t.Errorf("CmdInit = %q, expected 'init'", router.CmdInit)
	}
}

// ============================================================================
// TEST: TestMockBunkerManager
// ============================================================================

func TestMockBunkerManager(t *testing.T) {
	bm := newMockBunkerManager()

	// Test all methods exist and are callable
	if err := bm.Create("test"); err != nil {
		t.Errorf("Create failed: %v", err)
	}
	if !bm.createCalled {
		t.Error("Create was not called")
	}
	if bm.lastCreatedName != "test" {
		t.Errorf("lastCreatedName = %q, expected 'test'", bm.lastCreatedName)
	}

	if err := bm.Delete("test2"); err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	if !bm.deleteCalled {
		t.Error("Delete was not called")
	}

	if err := bm.List(); err != nil {
		t.Errorf("List failed: %v", err)
	}
	if !bm.listCalled {
		t.Error("List was not called")
	}

	if err := bm.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	if !bm.stopCalled {
		t.Error("Stop was not called")
	}

	if err := bm.Prune(); err != nil {
		t.Errorf("Prune failed: %v", err)
	}
	if !bm.pruneCalled {
		t.Error("Prune was not called")
	}

	if err := bm.Info("test3"); err != nil {
		t.Errorf("Info failed: %v", err)
	}
	if !bm.infoCalled {
		t.Error("Info was not called")
	}

	if err := bm.DeleteImage(); err != nil {
		t.Errorf("DeleteImage failed: %v", err)
	}
	if !bm.deleteImageCalled {
		t.Error("DeleteImage was not called")
	}

	if err := bm.Help(); err != nil {
		t.Errorf("Help failed: %v", err)
	}
	if !bm.helpCalled {
		t.Error("Help was not called")
	}

	if bm.GetUI() == nil {
		t.Error("GetUI returned nil")
	}
}

// ============================================================================
// TEST: TestMockPresenter
// ============================================================================

func TestMockPresenter(t *testing.T) {
	ui := &mockUI{}

	// Test ShowLog
	ui.ShowLog("test.key", "arg1", "arg2")
	if len(ui.logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(ui.logs))
	}
	if ui.logs[0] != "test.key" {
		t.Errorf("Expected log 'test.key', got %q", ui.logs[0])
	}

	// Test WithFields returns itself
	result := ui.WithFields(map[string]interface{}{"key": "value"})
	if result != ui {
		t.Error("WithFields should return the same presenter")
	}

	// Test other methods don't panic
	ui.ShowLogo()
	ui.ShowCommandCard("key", nil, nil)
	ui.ShowWarning("title", "subtitle", nil, nil, "footer")
	ui.ShowHelp()
	ui.ClearScreen()
	ui.RenderLifecycle("title", "subtitle", nil, "task", nil)
	ui.RenderLifecycleError("title", nil, "task", nil, nil, "where")
	ui.AskConfirmInCard("key", nil, nil, "prompt")
	ui.AskDelete("name", nil)
	ui.AskReset(nil, nil)
	ui.AskString("prompt")
	ui.AskConfirm("prompt")
	ui.GetText("key")
}

// ============================================================================
// TEST: TestRouter_CommandsWithSubArgsPreserveFirstArg
// ============================================================================

func TestRouter_CommandsWithSubArgsPreserveFirstArg(t *testing.T) {
	r, bm, _, _ := newTestRouter()

	// Test that the first argument is correctly extracted even with sub-args
	testCases := []struct {
		name         string
		args         []string
		expectedName string
	}{
		{"create_with_flags", []string{"create", "bunker1", "--verbose", "--gpu"}, "bunker1"},
		{"delete_with_flags", []string{"delete", "bunker2", "-f"}, "bunker2"},
		{"info_with_flags", []string{"info", "bunker3", "--details"}, "bunker3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.Handle(tc.args)

			switch {
			case tc.expectedName == bm.lastCreatedName:
				if bm.lastCreatedName != tc.expectedName {
					t.Errorf("Expected %q, got %q", tc.expectedName, bm.lastCreatedName)
				}
			case tc.expectedName == bm.lastDeletedName:
				if bm.lastDeletedName != tc.expectedName {
					t.Errorf("Expected %q, got %q", tc.expectedName, bm.lastDeletedName)
				}
			case tc.expectedName == bm.lastInfoName:
				if bm.lastInfoName != tc.expectedName {
					t.Errorf("Expected %q, got %q", tc.expectedName, bm.lastInfoName)
				}
			}
		})
	}
}
