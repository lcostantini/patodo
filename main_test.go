package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain_Integration(t *testing.T) {
	// This is an integration test that verifies the main components work together
	// We can't easily test the actual main() function since it runs the TUI,
	// but we can test that all the pieces integrate correctly

	tmpDir, err := os.MkdirTemp("", "patodo-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Create a store
	store := &TaskStore{
		filepath: filepath.Join(tmpDir, "tasks.json"),
		tasks:    []Task{},
	}

	// Add some tasks
	err = store.Add("Integration test task 1", "")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	err = store.Add("Integration test task 2", "")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Create initial model
	m := initialModel(store)

	// Verify model is properly initialized
	if m.store == nil {
		t.Error("Model store should not be nil")
	}

	if m.viewMode != ModeList {
		t.Error("Model should start in list mode")
	}

	// Refresh tasks
	m.refreshTasks()

	if len(m.tasks) != 2 {
		t.Errorf("Expected 2 tasks in model, got %d", len(m.tasks))
	}

	// Test that the model can be initialized
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}

	// Test that view can be rendered
	view := m.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestNewTaskStore_Integration(t *testing.T) {
	// Test that NewTaskStore creates the directory structure
	// This will use the actual config directory, so we need to be careful

	store, err := NewTaskStore()
	if err != nil {
		t.Fatalf("NewTaskStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("NewTaskStore returned nil store")
	}

	if store.tasks == nil {
		t.Error("Store tasks should be initialized")
	}

	// Verify the config directory was created
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "patodo")
	info, err := os.Stat(configDir)
	if err != nil {
		t.Errorf("Config directory should exist: %v", err)
	}

	if info != nil && !info.IsDir() {
		t.Error("Config path should be a directory")
	}
}

func TestEndToEnd_TaskLifecycle(t *testing.T) {
	// End-to-end test of a complete task lifecycle
	tmpDir, err := os.MkdirTemp("", "patodo-e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	// Create store
	store := &TaskStore{
		filepath: filepath.Join(tmpDir, "tasks.json"),
		tasks:    []Task{},
	}

	// Create model
	m := initialModel(store)

	// 1. Start in list mode with no tasks
	if len(m.tasks) != 0 {
		t.Error("Should start with no tasks")
	}

	// 2. Add a task via the store
	err = m.store.Add("Buy groceries", "")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// 3. Refresh to see the new task
	m.refreshTasks()
	if len(m.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(m.tasks))
	}

	taskID := m.tasks[0].ID

	// 4. Mark task as in-progress
	err = m.store.UpdateStatus(taskID, StatusInProgress)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	m.refreshTasks()
	if m.tasks[0].Status != StatusInProgress {
		t.Error("Task should be in-progress")
	}

	// 5. Mark task as done
	err = m.store.UpdateStatus(taskID, StatusDone)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	m.refreshTasks()
	if m.tasks[0].Status != StatusDone {
		t.Error("Task should be done")
	}

	// 6. Test filtering
	if err := m.store.Add("Another task", ""); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	status := StatusDone
	m.filterStatus = &status
	m.refreshTasks()

	if len(m.tasks) != 1 {
		t.Errorf("Filter should show only 1 done task, got %d", len(m.tasks))
	}

	// 7. Clear filter
	m.filterStatus = nil
	m.refreshTasks()

	if len(m.tasks) != 2 {
		t.Errorf("Should show all 2 tasks without filter, got %d", len(m.tasks))
	}

	// 8. Delete a task
	err = m.store.Delete(taskID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	m.refreshTasks()
	if len(m.tasks) != 1 {
		t.Errorf("Should have 1 task after deletion, got %d", len(m.tasks))
	}

	// 9. Save and verify persistence
	err = m.store.Save()
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load into new store
	newStore := &TaskStore{
		filepath: m.store.filepath,
		tasks:    []Task{},
	}

	err = newStore.Load()
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(newStore.tasks) != 1 {
		t.Errorf("Loaded store should have 1 task, got %d", len(newStore.tasks))
	}

	if newStore.tasks[0].Description != "Another task" {
		t.Error("Loaded task description doesn't match")
	}
}
