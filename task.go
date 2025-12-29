package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// TaskStatus represents the state of a task
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in-progress"
	StatusDone       TaskStatus = "done"
)

// TaskCategory represents a task category
type TaskCategory string

// Task represents a single TODO item
type Task struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Category    TaskCategory `json:"category"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// TaskStore handles persistence of tasks
type TaskStore struct {
	filepath string
	tasks    []Task
}

// FilterOptions contains optional filter criteria
type FilterOptions struct {
	Status   *TaskStatus
	Category *TaskCategory
}

// NewTaskStore creates a new task store
func NewTaskStore() (*TaskStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(homeDir, ".config", "patodo")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, "tasks.json")
	store := &TaskStore{
		filepath: filePath,
		tasks:    []Task{},
	}

	// Load existing tasks
	if err := store.Load(); err != nil {
		// If file doesn't exist, that's okay
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return store, nil
}

// Load reads tasks from disk
func (s *TaskStore) Load() error {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.tasks)
}

// Save writes tasks to disk
func (s *TaskStore) Save() error {
	data, err := json.MarshalIndent(s.tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filepath, data, 0644)
}

// GetAll returns all tasks
func (s *TaskStore) GetAll() []Task {
	return s.tasks
}

// GetCategories returns a list of unique categories from all tasks
func (s *TaskStore) GetCategories() []string {
	categorySet := make(map[string]struct{})
	for _, task := range s.tasks {
		if task.Category != "" {
			categorySet[string(task.Category)] = struct{}{}
		}
	}

	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}
	return categories
}

// Add adds a new task
func (s *TaskStore) Add(description string, category TaskCategory) error {
	task := Task{
		ID:          generateID(),
		Description: description,
		Status:      StatusPending,
		Category:    category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.tasks = append(s.tasks, task)
	return s.Save()
}

// findTaskIndex returns the index of a task by ID, or -1 if not found
func (s *TaskStore) findTaskIndex(id string) int {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			return i
		}
	}
	return -1
}

// UpdateStatus updates the status of a task
func (s *TaskStore) UpdateStatus(id string, status TaskStatus) error {
	if idx := s.findTaskIndex(id); idx != -1 {
		s.tasks[idx].Status = status
		s.tasks[idx].UpdatedAt = time.Now()
		return s.Save()
	}
	return nil
}

// UpdateDescription updates the description of a task
func (s *TaskStore) UpdateDescription(id string, description string) error {
	if idx := s.findTaskIndex(id); idx != -1 {
		s.tasks[idx].Description = description
		s.tasks[idx].UpdatedAt = time.Now()
		return s.Save()
	}
	return nil
}

// UpdateCategory updates the category of a task
func (s *TaskStore) UpdateCategory(id string, category TaskCategory) error {
	if idx := s.findTaskIndex(id); idx != -1 {
		s.tasks[idx].Category = category
		s.tasks[idx].UpdatedAt = time.Now()
		return s.Save()
	}
	return nil
}

// Update updates both description and category of a task
func (s *TaskStore) Update(id string, description string, category TaskCategory) error {
	if idx := s.findTaskIndex(id); idx != -1 {
		s.tasks[idx].Description = description
		s.tasks[idx].Category = category
		s.tasks[idx].UpdatedAt = time.Now()
		return s.Save()
	}
	return nil
}

// Delete removes a task
func (s *TaskStore) Delete(id string) error {
	if idx := s.findTaskIndex(id); idx != -1 {
		s.tasks = append(s.tasks[:idx], s.tasks[idx+1:]...)
		return s.Save()
	}
	return nil
}

// Filter returns tasks matching the given criteria
// If a filter option is nil, it's ignored
func (s *TaskStore) Filter(opts FilterOptions) []Task {
	var filtered []Task
	for _, task := range s.tasks {
		// Check status filter
		if opts.Status != nil && task.Status != *opts.Status {
			continue
		}

		// Check category filter
		if opts.Category != nil && task.Category != *opts.Category {
			continue
		}

		filtered = append(filtered, task)
	}
	return filtered
}

// generateID creates a simple unique ID
func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
