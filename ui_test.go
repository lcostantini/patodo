package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestViewMode(t *testing.T) {
	modes := []ViewMode{ModeList, ModeCreate, ModeEdit, ModeFilter, ModeFilterCategory}
	if len(modes) != 5 {
		t.Error("Expected 5 view modes")
	}
}

func createTestModel(t *testing.T) (model, string) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "patodo-ui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	store := &TaskStore{
		filepath: filepath.Join(tmpDir, "tasks.json"),
		tasks:    []Task{},
	}

	m := initialModel(store)
	return m, tmpDir
}

func TestInitialModel(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	if m.store == nil {
		t.Error("initialModel should set store")
	}

	if m.tasks == nil {
		t.Error("initialModel should set tasks")
	}

	if m.cursor != 0 {
		t.Errorf("cursor should start at 0, got %d", m.cursor)
	}

	if m.viewMode != ModeList {
		t.Errorf("viewMode should start at ModeList, got %d", m.viewMode)
	}

	if m.quitting {
		t.Error("quitting should be false initially")
	}

	if m.activeInput != 0 {
		t.Errorf("activeInput should start at 0, got %d", m.activeInput)
	}
}

func TestModel_Init(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModel_RefreshTasks(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Add tasks to store
	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 3", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Initially tasks should be empty since we haven't refreshed
	if len(m.tasks) != 0 {
		t.Errorf("Expected 0 tasks before refresh, got %d", len(m.tasks))
	}

	// Refresh tasks
	m.refreshTasks()

	if len(m.tasks) != 3 {
		t.Errorf("Expected 3 tasks after refresh, got %d", len(m.tasks))
	}

	// Test with filter
	if err := m.store.UpdateStatus(m.store.tasks[0].ID, StatusDone); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}
	status := StatusDone
	m.filterStatus = &status
	m.refreshTasks()

	if len(m.tasks) != 1 {
		t.Errorf("Expected 1 done task after filtered refresh, got %d", len(m.tasks))
	}
}

func TestModel_GetStatusIcon(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	tests := []struct {
		status TaskStatus
		want   string
	}{
		{StatusPending, "○"},
		{StatusInProgress, "⟳"},
		{StatusDone, "✓"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			icon := m.getStatusIcon(tt.status)
			if icon != tt.want {
				t.Errorf("getStatusIcon(%v) = %v, want %v", tt.status, icon, tt.want)
			}
		})
	}
}

func TestModel_GetStatusColor(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	tests := []struct {
		status TaskStatus
		want   string
	}{
		{StatusPending, "250"},
		{StatusInProgress, "214"},
		{StatusDone, "34"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			color := m.getStatusColor(tt.status)
			if color != tt.want {
				t.Errorf("getStatusColor(%v) = %v, want %v", tt.status, color, tt.want)
			}
		})
	}
}

func TestModel_UpdateListMode_Navigation(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Add some tasks
	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 3", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	// Test down navigation
	m.cursor = 0
	updatedModel, _ := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(model)
	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after down, got %d", m.cursor)
	}

	// Test up navigation
	updatedModel, _ = m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(model)
	if m.cursor != 0 {
		t.Errorf("cursor should be 0 after up, got %d", m.cursor)
	}

	// Test boundary - up at top
	updatedModel, _ = m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(model)
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 when at top, got %d", m.cursor)
	}

	// Test boundary - down at bottom
	m.cursor = 2
	updatedModel, _ = m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(model)
	if m.cursor != 2 {
		t.Errorf("cursor should stay at 2 when at bottom, got %d", m.cursor)
	}
}

func TestModel_UpdateListMode_CreateTask(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Press 'n' to enter create mode
	updatedModel, _ := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updatedModel.(model)

	if m.viewMode != ModeCreate {
		t.Errorf("viewMode should be ModeCreate, got %d", m.viewMode)
	}
}

func TestModel_UpdateListMode_ToggleDone(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Add a task
	if err := m.store.Add("Test task", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	// Toggle to done
	updatedModel, _ := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updatedModel.(model)

	m.refreshTasks()
	if m.tasks[0].Status != StatusDone {
		t.Errorf("Task should be done, got %v", m.tasks[0].Status)
	}

	// Toggle back to pending
	updatedModel, _ = m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updatedModel.(model)

	m.refreshTasks()
	if m.tasks[0].Status != StatusPending {
		t.Errorf("Task should be pending, got %v", m.tasks[0].Status)
	}
}

func TestModel_UpdateListMode_SetInProgress(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Add a task
	if err := m.store.Add("Test task", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	// Set to in-progress
	updatedModel, _ := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updatedModel.(model)

	m.refreshTasks()
	if m.tasks[0].Status != StatusInProgress {
		t.Errorf("Task should be in-progress, got %v", m.tasks[0].Status)
	}
}

func TestModel_UpdateListMode_DeleteTask(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Add tasks
	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	// Delete first task
	updatedModel, _ := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updatedModel.(model)

	m.refreshTasks()
	if len(m.tasks) != 1 {
		t.Errorf("Should have 1 task after delete, got %d", len(m.tasks))
	}
}

func TestModel_UpdateListMode_Filter(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Press 'f' to enter filter mode
	updatedModel, _ := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updatedModel.(model)

	if m.viewMode != ModeFilter {
		t.Errorf("viewMode should be ModeFilter, got %d", m.viewMode)
	}
}

func TestModel_UpdateListMode_Quit(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Press 'q' to quit
	updatedModel, cmd := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updatedModel.(model)

	if !m.quitting {
		t.Error("quitting should be true after pressing q")
	}

	if cmd == nil {
		t.Error("quit command should not be nil")
	}
}

func TestModel_UpdateCreateMode(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Enter create mode
	m.viewMode = ModeCreate
	m.textInput = textinput.New()
	m.textInput.SetValue("New task")
	m.categoryInput = textinput.New()
	m.categoryInput.SetValue("work")
	m.activeInput = 0

	// Press Enter to create task with category
	updatedModel, _ := m.updateCreateMode(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after creating task")
	}

	m.refreshTasks()
	if len(m.tasks) != 1 {
		t.Errorf("Should have 1 task after creation, got %d", len(m.tasks))
	}

	if m.tasks[0].Description != "New task" {
		t.Errorf("Task description should be 'New task', got %v", m.tasks[0].Description)
	}

	if m.tasks[0].Category != TaskCategory("work") {
		t.Errorf("Task category should be 'work', got %v", m.tasks[0].Category)
	}
}

func TestModel_UpdateCreateMode_TabSwitching(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Enter create mode
	m.viewMode = ModeCreate
	m.activeInput = 0

	// Press Tab to switch to category input
	updatedModel, _ := m.updateCreateMode(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)

	if m.activeInput != 1 {
		t.Errorf("activeInput should be 1 after Tab, got %d", m.activeInput)
	}

	// Press Tab again to switch back to description
	updatedModel, _ = m.updateCreateMode(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)

	if m.activeInput != 0 {
		t.Errorf("activeInput should be 0 after second Tab, got %d", m.activeInput)
	}
}

func TestModel_UpdateCreateMode_Cancel(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Enter create mode
	m.viewMode = ModeCreate

	// Press Esc to cancel
	updatedModel, _ := m.updateCreateMode(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after canceling")
	}

	if len(m.tasks) != 0 {
		t.Error("Should not create task when canceling")
	}
}

func TestModel_UpdateFilterMode(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add tasks with different statuses
	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 3", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.UpdateStatus(m.store.tasks[0].ID, StatusDone); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}
	if err := m.store.UpdateStatus(m.store.tasks[1].ID, StatusInProgress); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Enter filter mode
	m.viewMode = ModeFilter

	// Filter by done
	updatedModel, _ := m.updateFilterMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after filtering")
	}

	if m.filterStatus == nil {
		t.Error("filterStatus should be set")
	} else if *m.filterStatus != StatusDone {
		t.Errorf("filterStatus should be StatusDone, got %v", *m.filterStatus)
	}

	if len(m.tasks) != 1 {
		t.Errorf("Should show 1 done task, got %d", len(m.tasks))
	}
}

func TestModel_UpdateFilterMode_ShowAll(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add tasks
	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Set a filter
	status := StatusDone
	m.filterStatus = &status
	m.viewMode = ModeFilter

	// Press 'a' to show all
	updatedModel, _ := m.updateFilterMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedModel.(model)

	if m.filterStatus != nil {
		t.Error("filterStatus should be nil when showing all")
	}

	if len(m.tasks) != 2 {
		t.Errorf("Should show all 2 tasks, got %d", len(m.tasks))
	}
}

func TestModel_View(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test list view
	view := m.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}

	if !contains(view, "patodo") {
		t.Error("View should contain title")
	}

	// Test quitting view
	m.quitting = true
	view = m.View()
	if !contains(view, "Goodbye") {
		t.Error("Quitting view should contain goodbye message")
	}
}

func TestModel_View_WithTasks(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add some tasks
	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	view := m.View()
	if !contains(view, "Task 1") {
		t.Error("View should contain task descriptions")
	}

	if !contains(view, "Task 2") {
		t.Error("View should contain task descriptions")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestModel_Update(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with KeyMsg in list mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(model)

	if m.viewMode != ModeCreate {
		t.Errorf("Should be in create mode after pressing 'n', got %d", m.viewMode)
	}

	// Test with non-KeyMsg (should update textInput)
	m.viewMode = ModeCreate
	m.textInput.SetValue("test")
	updatedModel, _ = m.Update(tea.WindowSizeMsg{})
	m = updatedModel.(model)

	// Should still be in create mode
	if m.viewMode != ModeCreate {
		t.Error("Should remain in create mode for non-key messages")
	}
}

func TestModel_UpdateFilterCategoryMode(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add tasks with categories
	if err := m.store.Add("Task 1", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", TaskCategory("personal")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 3", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	// Enter filter category mode
	m.viewMode = ModeFilterCategory

	// Test escape
	updatedModel, _ := m.updateFilterCategoryMode(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after escape")
	}

	// Test showing all categories
	m.viewMode = ModeFilterCategory
	updatedModel, _ = m.updateFilterCategoryMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after showing all")
	}

	if m.filterCategory != nil {
		t.Error("filterCategory should be nil when showing all")
	}

	// Test selecting a category by number
	m.viewMode = ModeFilterCategory
	updatedModel, _ = m.updateFilterCategoryMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after selecting category")
	}

	if m.filterCategory == nil {
		t.Error("filterCategory should be set after selection")
	}

	// Test invalid number (out of range)
	m.viewMode = ModeFilterCategory
	updatedModel, _ = m.updateFilterCategoryMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}})
	m = updatedModel.(model)

	// Should still be in filter category mode since selection was invalid
	if m.viewMode != ModeFilterCategory {
		t.Error("Should remain in filter category mode for invalid selection")
	}
}

func TestModel_UpdateCreateMode_WithCategory(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Set up create mode with both description and category
	m.viewMode = ModeCreate
	m.textInput = textinput.New()
	m.textInput.SetValue("New task with category")
	m.categoryInput = textinput.New()
	m.categoryInput.SetValue("work")

	// Press Enter to create task with category
	updatedModel, _ := m.updateCreateMode(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after creating task")
	}

	m.refreshTasks()
	if len(m.tasks) != 1 {
		t.Errorf("Should have 1 task after creation, got %d", len(m.tasks))
	}

	if m.tasks[0].Category != TaskCategory("work") {
		t.Errorf("Task category should be 'work', got %v", m.tasks[0].Category)
	}

	if !contains(m.message, "work") {
		t.Error("Message should mention the category")
	}
}

func TestModel_UpdateCreateMode_EmptyCategory(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Set up create mode with description but empty category
	m.viewMode = ModeCreate
	m.textInput = textinput.New()
	m.textInput.SetValue("Task without category")
	m.categoryInput = textinput.New()
	m.categoryInput.SetValue("")

	// Press Enter to create task without category
	updatedModel, _ := m.updateCreateMode(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode")
	}

	if !contains(m.message, "category is required") {
		t.Error("Should show category required message")
	}

	m.refreshTasks()
	if len(m.tasks) != 0 {
		t.Error("Should not create task without category")
	}
}

func TestModel_UpdateCreateMode_EmptyDescription(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Enter create mode with empty description
	m.viewMode = ModeCreate
	m.textInput.SetValue("")
	m.categoryInput = textinput.New()
	m.categoryInput.SetValue("work")

	// Press Enter with empty description
	updatedModel, _ := m.updateCreateMode(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode when description is empty")
	}

	if !contains(m.message, "description is required") {
		t.Error("Should show description required message")
	}

	m.refreshTasks()
	if len(m.tasks) != 0 {
		t.Error("Should not create task with empty description")
	}
}

func TestModel_UpdateCreateMode_TextInput(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Enter create mode
	m.viewMode = ModeCreate
	m.textInput.Reset()

	// Simulate typing (regular key that's not Enter or Esc)
	updatedModel, _ := m.updateCreateMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedModel.(model)

	// Should still be in create mode
	if m.viewMode != ModeCreate {
		t.Error("Should remain in create mode while typing")
	}
}

func TestModel_View_CreateMode(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	m.viewMode = ModeCreate
	view := m.View()

	if !contains(view, "patodo") {
		t.Error("View should contain title in create mode")
	}

	if !contains(view, "Description") {
		t.Error("View should contain Description label")
	}

	if !contains(view, "Category") {
		t.Error("View should contain Category label")
	}
}

func TestModel_View_FilterCategoryMode(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add tasks with categories
	if err := m.store.Add("Task 1", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", TaskCategory("personal")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	m.viewMode = ModeFilterCategory
	view := m.View()

	if !contains(view, "Select category") {
		t.Error("View should show category selection prompt")
	}

	if !contains(view, "work") {
		t.Error("View should list available categories")
	}
}

func TestModel_View_FilterCategoryMode_NoCategories(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	m.viewMode = ModeFilterCategory
	view := m.View()

	if !contains(view, "No categories") {
		t.Error("View should show 'No categories' when there are none")
	}
}

func TestModel_View_WithCategories(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add task with category
	if err := m.store.Add("Task with category", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	view := m.View()

	if !contains(view, "work") {
		t.Error("View should display task category")
	}
}

func TestModel_View_FilterInfo(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with status filter
	status := StatusDone
	m.filterStatus = &status
	view := m.View()

	if !contains(view, "done") {
		t.Error("View should show status filter in help text")
	}

	// Test with category filter
	category := TaskCategory("work")
	m.filterStatus = nil
	m.filterCategory = &category
	view = m.View()

	if !contains(view, "work") {
		t.Error("View should show category filter in help text")
	}

	// Test with both filters
	m.filterStatus = &status
	view = m.View()

	if !contains(view, "done") || !contains(view, "work") {
		t.Error("View should show both filters in help text")
	}
}

func TestModel_HasCurrentTask(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// No tasks
	if m.hasCurrentTask() {
		t.Error("Should return false when there are no tasks")
	}

	// Add task
	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	if !m.hasCurrentTask() {
		t.Error("Should return true when cursor is on valid task")
	}

	// Cursor out of range
	m.cursor = 10
	if m.hasCurrentTask() {
		t.Error("Should return false when cursor is out of range")
	}
}

func TestModel_GetCurrentTask(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := m.store.Add("First task", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Second task", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	// Get first task
	m.cursor = 0
	task := m.getCurrentTask()
	if task.Description != "First task" {
		t.Errorf("Expected 'First task', got %s", task.Description)
	}

	// Get second task
	m.cursor = 1
	task = m.getCurrentTask()
	if task.Description != "Second task" {
		t.Errorf("Expected 'Second task', got %s", task.Description)
	}
}

func TestModel_UpdateTaskStatus(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := m.store.Add("Task", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()
	m.cursor = 0

	// Update status
	m.updateTaskStatus(StatusDone)

	if m.tasks[0].Status != StatusDone {
		t.Errorf("Task status should be done, got %v", m.tasks[0].Status)
	}
}

func TestModel_ApplyStatusFilter(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := m.store.Add("Task 1", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.Add("Task 2", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := m.store.UpdateStatus(m.store.tasks[0].ID, StatusDone); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	m.viewMode = ModeFilter
	m.applyStatusFilter(StatusDone, "Showing done")

	if m.viewMode != ModeList {
		t.Error("Should return to list mode")
	}

	if m.filterStatus == nil || *m.filterStatus != StatusDone {
		t.Error("Filter status should be set to done")
	}

	if len(m.tasks) != 1 {
		t.Errorf("Should show 1 done task, got %d", len(m.tasks))
	}

	if m.cursor != 0 {
		t.Error("Cursor should be reset to 0")
	}

	if m.message != "Showing done" {
		t.Errorf("Message should be 'Showing done', got %s", m.message)
	}
}

func TestModel_UpdateListMode_EnterEditMode(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add a task
	if err := m.store.Add("Task to edit", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()

	// Press 'e' to enter edit mode
	updatedModel, _ := m.updateListMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updatedModel.(model)

	if m.viewMode != ModeEdit {
		t.Errorf("viewMode should be ModeEdit, got %d", m.viewMode)
	}

	if m.editingTaskID == "" {
		t.Error("editingTaskID should be set")
	}

	if m.textInput.Value() != "Task to edit" {
		t.Errorf("textInput should be populated with task description, got '%s'", m.textInput.Value())
	}

	if m.categoryInput.Value() != "work" {
		t.Errorf("categoryInput should be populated with task category, got '%s'", m.categoryInput.Value())
	}
}

func TestModel_UpdateEditMode_Save(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add a task
	if err := m.store.Add("Original task", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()
	taskID := m.tasks[0].ID

	// Enter edit mode
	m.viewMode = ModeEdit
	m.editingTaskID = taskID
	m.textInput = textinput.New()
	m.textInput.SetValue("Updated task")
	m.categoryInput = textinput.New()
	m.categoryInput.SetValue("personal")

	// Press Enter to save
	updatedModel, _ := m.updateEditMode(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after saving")
	}

	if m.editingTaskID != "" {
		t.Error("editingTaskID should be cleared after saving")
	}

	m.refreshTasks()
	if len(m.tasks) != 1 {
		t.Fatalf("Should still have 1 task, got %d", len(m.tasks))
	}

	if m.tasks[0].Description != "Updated task" {
		t.Errorf("Task description should be 'Updated task', got '%s'", m.tasks[0].Description)
	}

	if m.tasks[0].Category != TaskCategory("personal") {
		t.Errorf("Task category should be 'personal', got '%s'", m.tasks[0].Category)
	}

	if !contains(m.message, "updated") {
		t.Error("Message should indicate task was updated")
	}
}

func TestModel_UpdateEditMode_Cancel(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add a task
	if err := m.store.Add("Original task", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()
	taskID := m.tasks[0].ID

	// Enter edit mode
	m.viewMode = ModeEdit
	m.editingTaskID = taskID
	m.textInput = textinput.New()
	m.textInput.SetValue("Modified task")

	// Press Esc to cancel
	updatedModel, _ := m.updateEditMode(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode after canceling")
	}

	if m.editingTaskID != "" {
		t.Error("editingTaskID should be cleared after canceling")
	}

	m.refreshTasks()
	// Task should remain unchanged
	if m.tasks[0].Description != "Original task" {
		t.Errorf("Task description should remain 'Original task', got '%s'", m.tasks[0].Description)
	}
}

func TestModel_UpdateEditMode_EmptyDescription(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add a task
	if err := m.store.Add("Original task", TaskCategory("work")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	m.refreshTasks()
	taskID := m.tasks[0].ID

	// Enter edit mode with empty description
	m.viewMode = ModeEdit
	m.editingTaskID = taskID
	m.textInput = textinput.New()
	m.textInput.SetValue("")
	m.categoryInput = textinput.New()

	// Press Enter to save
	updatedModel, _ := m.updateEditMode(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if m.viewMode != ModeList {
		t.Error("Should return to list mode")
	}

	if !contains(m.message, "cancelled") {
		t.Error("Message should indicate edit was cancelled")
	}

	m.refreshTasks()
	// Task should remain unchanged
	if m.tasks[0].Description != "Original task" {
		t.Errorf("Task description should remain unchanged, got '%s'", m.tasks[0].Description)
	}
}

func TestModel_UpdateEditMode_TabSwitching(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Enter edit mode
	m.viewMode = ModeEdit
	m.activeInput = 0

	// Press Tab to switch to category input
	updatedModel, _ := m.updateEditMode(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)

	if m.activeInput != 1 {
		t.Errorf("activeInput should be 1 after Tab, got %d", m.activeInput)
	}

	// Press Tab again to switch back to description
	updatedModel, _ = m.updateEditMode(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)

	if m.activeInput != 0 {
		t.Errorf("activeInput should be 0 after second Tab, got %d", m.activeInput)
	}
}

func TestModel_View_EditMode(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	m.viewMode = ModeEdit
	view := m.View()

	if !contains(view, "patodo") {
		t.Error("View should contain title in edit mode")
	}

	if !contains(view, "Description") {
		t.Error("View should contain Description label")
	}

	if !contains(view, "Category") {
		t.Error("View should contain Category label")
	}
}

func TestModel_UpdateEditMode_TextInput(t *testing.T) {
	m, tmpDir := createTestModel(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Enter edit mode
	m.viewMode = ModeEdit
	m.textInput = textinput.New()
	m.categoryInput = textinput.New()
	m.activeInput = 0

	// Simulate typing in description field
	updatedModel, _ := m.updateEditMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedModel.(model)

	// Should still be in edit mode
	if m.viewMode != ModeEdit {
		t.Error("Should remain in edit mode while typing")
	}

	// Switch to category input
	m.activeInput = 1

	// Simulate typing in category field
	updatedModel, _ = m.updateEditMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updatedModel.(model)

	// Should still be in edit mode
	if m.viewMode != ModeEdit {
		t.Error("Should remain in edit mode while typing")
	}
}
