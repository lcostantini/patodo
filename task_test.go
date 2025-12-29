package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTaskStore(t *testing.T) {
	store, err := NewTaskStore()
	if err != nil {
		t.Fatalf("Failed to create task store: %v", err)
	}
	if store == nil {
		t.Fatal("Expected non-nil task store")
	}
	if store.tasks == nil {
		t.Fatal("Expected initialized tasks slice")
	}
}

func TestTaskStore_Add(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	err := store.Add("Test task", "work")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	tasks := store.GetAll()
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Description != "Test task" {
		t.Errorf("Expected description 'Test task', got '%s'", task.Description)
	}
	if task.Category != "work" {
		t.Errorf("Expected category 'work', got '%s'", task.Category)
	}
	if task.Status != StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", task.Status)
	}
	if task.ID == "" {
		t.Error("Expected non-empty task ID")
	}
}

func TestTaskStore_UpdateStatus(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Test task", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	tasks := store.GetAll()
	taskID := tasks[0].ID
	oldUpdatedAt := tasks[0].UpdatedAt

	time.Sleep(10 * time.Millisecond) // Ensure time difference

	err := store.UpdateStatus(taskID, StatusInProgress)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	tasks = store.GetAll()
	if tasks[0].Status != StatusInProgress {
		t.Errorf("Expected status 'in-progress', got '%s'", tasks[0].Status)
	}
	if !tasks[0].UpdatedAt.After(oldUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTaskStore_UpdateCategory(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Test task", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	tasks := store.GetAll()
	taskID := tasks[0].ID

	err := store.UpdateCategory(taskID, "personal")
	if err != nil {
		t.Fatalf("Failed to update category: %v", err)
	}

	tasks = store.GetAll()
	if tasks[0].Category != "personal" {
		t.Errorf("Expected category 'personal', got '%s'", tasks[0].Category)
	}
}

func TestTaskStore_UpdateDescription(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Original description", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	tasks := store.GetAll()
	taskID := tasks[0].ID

	err := store.UpdateDescription(taskID, "Updated description")
	if err != nil {
		t.Fatalf("Failed to update description: %v", err)
	}

	tasks = store.GetAll()
	if tasks[0].Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", tasks[0].Description)
	}
}

func TestTaskStore_Update(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Original description", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	tasks := store.GetAll()
	taskID := tasks[0].ID

	err := store.Update(taskID, "New description", "personal")
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	tasks = store.GetAll()
	if tasks[0].Description != "New description" {
		t.Errorf("Expected description 'New description', got '%s'", tasks[0].Description)
	}
	if tasks[0].Category != "personal" {
		t.Errorf("Expected category 'personal', got '%s'", tasks[0].Category)
	}
}

func TestTaskStore_Delete(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Task 1", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 2", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	tasks := store.GetAll()
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}

	taskID := tasks[0].ID
	err := store.Delete(taskID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	tasks = store.GetAll()
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task after deletion, got %d", len(tasks))
	}
	if tasks[0].ID == taskID {
		t.Error("Deleted task still exists")
	}
}

func TestTaskStore_Filter_ByStatus(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Task 1", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 2", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 3", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	tasks := store.GetAll()
	if err := store.UpdateStatus(tasks[0].ID, StatusInProgress); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}
	if err := store.UpdateStatus(tasks[1].ID, StatusDone); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Filter by status
	status := StatusInProgress
	filtered := store.Filter(FilterOptions{Status: &status})
	if len(filtered) != 1 {
		t.Fatalf("Expected 1 in-progress task, got %d", len(filtered))
	}
	if filtered[0].Status != StatusInProgress {
		t.Errorf("Expected status 'in-progress', got '%s'", filtered[0].Status)
	}
}

func TestTaskStore_Filter_ByCategory(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Task 1", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 2", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 3", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Filter by category
	category := TaskCategory("work")
	filtered := store.Filter(FilterOptions{Category: &category})
	if len(filtered) != 2 {
		t.Fatalf("Expected 2 work tasks, got %d", len(filtered))
	}
	for _, task := range filtered {
		if task.Category != "work" {
			t.Errorf("Expected category 'work', got '%s'", task.Category)
		}
	}
}

func TestTaskStore_Filter_ByCategoryAndStatus(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Task 1", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 2", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 3", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	tasks := store.GetAll()
	if err := store.UpdateStatus(tasks[0].ID, StatusDone); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}
	if err := store.UpdateStatus(tasks[2].ID, StatusDone); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Filter by both category and status
	status := StatusDone
	category := TaskCategory("work")
	filtered := store.Filter(FilterOptions{
		Status:   &status,
		Category: &category,
	})

	if len(filtered) != 1 {
		t.Fatalf("Expected 1 done work task, got %d", len(filtered))
	}
	if filtered[0].Status != StatusDone {
		t.Errorf("Expected status 'done', got '%s'", filtered[0].Status)
	}
	if filtered[0].Category != "work" {
		t.Errorf("Expected category 'work', got '%s'", filtered[0].Category)
	}
}

func TestTaskStore_Filter_NoFilters(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Task 1", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 2", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// No filters should return all tasks
	filtered := store.Filter(FilterOptions{})
	if len(filtered) != 2 {
		t.Fatalf("Expected 2 tasks with no filters, got %d", len(filtered))
	}
}

func TestTaskStore_GetCategories_Empty(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Task 1", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 2", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 3", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 4", TaskCategory("custom")); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	categories := store.GetCategories()
	if len(categories) != 3 {
		t.Fatalf("Expected 3 unique categories, got %d", len(categories))
	}

	categoryMap := make(map[TaskCategory]bool)
	for _, cat := range categories {
		categoryMap[TaskCategory(cat)] = true
	}

	if !categoryMap["work"] {
		t.Error("Expected 'work' category")
	}
	if !categoryMap["personal"] {
		t.Error("Expected 'personal' category")
	}
	if !categoryMap[TaskCategory("custom")] {
		t.Error("Expected 'custom' category")
	}
}

func TestTaskStore_SaveAndLoad(t *testing.T) {
	store := setupTestStore(t)
	defer cleanupTestStore(store)

	if err := store.Add("Task 1", "work"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}
	if err := store.Add("Task 2", "personal"); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Create a new store with the same filepath
	newStore := &TaskStore{
		filepath: store.filepath,
		tasks:    []Task{},
	}

	err := newStore.Load()
	if err != nil {
		t.Fatalf("Failed to load tasks: %v", err)
	}

	tasks := newStore.GetAll()
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 loaded tasks, got %d", len(tasks))
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	time.Sleep(1 * time.Millisecond)
	id2 := generateID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("Expected unique IDs")
	}
}

// Helper functions

func setupTestStore(t *testing.T) *TaskStore {
	t.Helper()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_tasks.json")

	return &TaskStore{
		filepath: testFile,
		tasks:    []Task{},
	}
}

func cleanupTestStore(store *TaskStore) {
	_ = os.Remove(store.filepath)
}
