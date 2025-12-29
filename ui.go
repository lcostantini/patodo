package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents different views in the TUI
type ViewMode int

const (
	ModeList ViewMode = iota
	ModeCreate
	ModeEdit
	ModeFilter
	ModeFilterCategory
)

// Color constants
const (
	colorTitle      = "86"
	colorEmpty      = "240"
	colorMessage    = "241"
	colorHelp       = "240"
	colorCategory   = "63"
	colorPending    = "250"
	colorInProgress = "214"
	colorDone       = "34"
)

// Model holds the application state
type model struct {
	store          *TaskStore
	tasks          []Task
	cursor         int
	viewMode       ViewMode
	textInput      textinput.Model
	categoryInput  textinput.Model
	filterStatus   *TaskStatus
	filterCategory *TaskCategory
	message        string
	quitting       bool
	activeInput    int    // 0 for description, 1 for category
	editingTaskID  string // ID of task being edited
	viewAsTable    bool   // true for table view, false for list view
}

// initialModel creates the initial model
func initialModel(store *TaskStore) model {
	ti := textinput.New()
	ti.Placeholder = "Enter task description..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	ci := textinput.New()
	ci.Placeholder = "Enter category (work, personal, etc.)..."
	ci.CharLimit = 50
	ci.Width = 50

	return model{
		store:         store,
		tasks:         store.GetAll(),
		cursor:        0,
		viewMode:      ModeList,
		textInput:     ti,
		categoryInput: ci,
		activeInput:   0,
		viewAsTable:   true,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.viewMode {
		case ModeCreate:
			return m.updateCreateMode(msg)
		case ModeEdit:
			return m.updateEditMode(msg)
		case ModeFilter:
			return m.updateFilterMode(msg)
		case ModeFilterCategory:
			return m.updateFilterCategoryMode(msg)
		default:
			return m.updateListMode(msg)
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "n":
		m.viewMode = ModeCreate
		m.textInput.Reset()
		m.categoryInput.Reset()
		m.textInput.Focus()
		m.categoryInput.Blur()
		m.activeInput = 0
		m.editingTaskID = ""
		m.message = "Enter task details (Tab to switch fields, Enter to save, ESC to cancel)"
		return m, textinput.Blink

	case "e":
		if m.hasCurrentTask() {
			task := m.getCurrentTask()
			m.viewMode = ModeEdit
			m.editingTaskID = task.ID
			m.textInput.SetValue(task.Description)
			m.categoryInput.SetValue(string(task.Category))
			m.textInput.Focus()
			m.categoryInput.Blur()
			m.activeInput = 0
			m.message = "Edit task (Tab to switch fields, Enter to save, ESC to cancel)"
			return m, textinput.Blink
		}

	case "f":
		m.viewMode = ModeFilter
		m.message = "Filter: (a)ll, (p)ending, (i)n-progress, (d)one, (c)ategory, ESC to cancel"
		return m, nil

	case "v":
		m.viewAsTable = !m.viewAsTable
		if m.viewAsTable {
			m.message = "Switched to table view"
		} else {
			m.message = "Switched to list view"
		}
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}

	case "d":
		if m.hasCurrentTask() {
			task := m.getCurrentTask()
			if task.Status == StatusDone {
				m.updateTaskStatus(StatusPending)
				m.message = "Task marked as pending"
			} else {
				m.updateTaskStatus(StatusDone)
				m.message = "Task marked as done!"
			}
		}

	case "i":
		if m.hasCurrentTask() {
			m.updateTaskStatus(StatusInProgress)
			m.message = "Task marked as in-progress"
		}

	case "p":
		if m.hasCurrentTask() {
			m.updateTaskStatus(StatusPending)
			m.message = "Task marked as pending"
		}

	case "x":
		if m.hasCurrentTask() {
			task := m.getCurrentTask()
			if err := m.store.Delete(task.ID); err != nil {
				m.message = fmt.Sprintf("Error deleting task: %v", err)
			} else {
				m.message = "Task deleted"
			}
			m.refreshTasks()
			if m.cursor >= len(m.tasks) && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m model) updateCreateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.viewMode = ModeList
		m.message = "Task creation cancelled"
		return m, nil

	case tea.KeyTab:
		// Switch between description and category inputs
		if m.activeInput == 0 {
			m.activeInput = 1
			m.textInput.Blur()
			m.categoryInput.Focus()
		} else {
			m.activeInput = 0
			m.categoryInput.Blur()
			m.textInput.Focus()
		}
		return m, textinput.Blink

	case tea.KeyEnter:
		description := strings.TrimSpace(m.textInput.Value())
		if description == "" {
			m.viewMode = ModeList
			m.message = "Task creation cancelled - description is required"
			return m, nil
		}

		categoryStr := strings.TrimSpace(m.categoryInput.Value())
		if categoryStr == "" {
			m.viewMode = ModeList
			m.message = "Task creation cancelled - category is required"
			return m, nil
		}
		category := TaskCategory(categoryStr)
		if err := m.store.Add(description, category); err != nil {
			m.message = fmt.Sprintf("Error creating task: %v", err)
		} else {
			m.message = fmt.Sprintf("Task created: %s [%s]", description, categoryStr)
		}
		m.refreshTasks()
		m.viewMode = ModeList
		return m, nil
	}

	var cmd tea.Cmd
	if m.activeInput == 0 {
		m.textInput, cmd = m.textInput.Update(msg)
	} else {
		m.categoryInput, cmd = m.categoryInput.Update(msg)
	}
	return m, cmd
}

func (m model) updateEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.viewMode = ModeList
		m.message = "Edit cancelled"
		m.editingTaskID = ""
		return m, nil

	case tea.KeyTab:
		// Switch between description and category inputs
		if m.activeInput == 0 {
			m.activeInput = 1
			m.textInput.Blur()
			m.categoryInput.Focus()
		} else {
			m.activeInput = 0
			m.categoryInput.Blur()
			m.textInput.Focus()
		}
		return m, textinput.Blink

	case tea.KeyEnter:
		description := strings.TrimSpace(m.textInput.Value())
		if description == "" {
			m.viewMode = ModeList
			m.message = "Edit cancelled - description is required"
			m.editingTaskID = ""
			return m, nil
		}

		categoryStr := strings.TrimSpace(m.categoryInput.Value())
		category := TaskCategory(categoryStr)
		if err := m.store.Update(m.editingTaskID, description, category); err != nil {
			m.message = fmt.Sprintf("Error updating task: %v", err)
		} else {
			m.message = "Task updated successfully"
		}
		m.refreshTasks()
		m.editingTaskID = ""
		m.viewMode = ModeList
		return m, nil
	}

	var cmd tea.Cmd
	if m.activeInput == 0 {
		m.textInput, cmd = m.textInput.Update(msg)
	} else {
		m.categoryInput, cmd = m.categoryInput.Update(msg)
	}
	return m, cmd
}

func (m model) updateFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeList
		m.message = "Filter cancelled"
		return m, nil

	case "a":
		m.filterStatus = nil
		m.filterCategory = nil
		m.refreshTasks()
		m.viewMode = ModeList
		m.message = "Showing all tasks"
		m.cursor = 0

	case "p":
		m.applyStatusFilter(StatusPending, "Showing pending tasks")

	case "i":
		m.applyStatusFilter(StatusInProgress, "Showing in-progress tasks")

	case "d":
		m.applyStatusFilter(StatusDone, "Showing done tasks")

	case "c":
		m.viewMode = ModeFilterCategory
		m.message = "Select category to filter by"
	}

	return m, nil
}

func (m model) updateFilterCategoryMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ModeList
		m.message = "Filter cancelled"
		return m, nil

	case "a":
		m.filterCategory = nil
		m.refreshTasks()
		m.viewMode = ModeList
		m.message = "Showing all categories"
		m.cursor = 0
		return m, nil
	}

	// Check if user pressed a number key for category selection
	categories := m.store.GetCategories()
	if len(msg.String()) == 1 && msg.String()[0] >= '1' && msg.String()[0] <= '9' {
		idx := int(msg.String()[0] - '1')
		if idx < len(categories) {
			categoryStr := categories[idx]
			category := TaskCategory(categoryStr)
			m.filterCategory = &category
			m.refreshTasks()
			m.viewMode = ModeList
			m.message = fmt.Sprintf("Showing tasks in category: %s", categoryStr)
			m.cursor = 0
		}
	}

	return m, nil
}

func (m *model) refreshTasks() {
	opts := FilterOptions{
		Status:   m.filterStatus,
		Category: m.filterCategory,
	}
	m.tasks = m.store.Filter(opts)
}

// hasCurrentTask checks if there's a valid task at the cursor position
func (m model) hasCurrentTask() bool {
	return len(m.tasks) > 0 && m.cursor < len(m.tasks)
}

// getCurrentTask returns the task at the cursor position
func (m model) getCurrentTask() Task {
	return m.tasks[m.cursor]
}

// updateTaskStatus updates the status of the current task and refreshes
func (m *model) updateTaskStatus(status TaskStatus) {
	if m.hasCurrentTask() {
		task := m.getCurrentTask()
		if err := m.store.UpdateStatus(task.ID, status); err != nil {
			m.message = fmt.Sprintf("Error updating task: %v", err)
		}
		m.refreshTasks()
	}
}

// applyStatusFilter applies a status filter and returns to list mode
func (m *model) applyStatusFilter(status TaskStatus, message string) {
	m.filterStatus = &status
	m.refreshTasks()
	m.viewMode = ModeList
	m.message = message
	m.cursor = 0
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var s strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorTitle)).
		MarginBottom(1)
	s.WriteString(titleStyle.Render("📝 patodo"))
	s.WriteString("\n\n")

	// Message bar (above content)
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMessage)).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			MarginBottom(1).
			Italic(true).
			Faint(true)
		s.WriteString(messageStyle.Render(m.message))
		s.WriteString("\n\n")
	}

	switch m.viewMode {
	case ModeCreate:
		s.WriteString("Description:\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString("Category:\n")
		s.WriteString(m.categoryInput.View())
		s.WriteString("\n\n")
	case ModeEdit:
		s.WriteString("Description:\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString("Category:\n")
		s.WriteString(m.categoryInput.View())
		s.WriteString("\n\n")
	case ModeFilterCategory:
		// Show available categories
		categories := m.store.GetCategories()
		if len(categories) > 0 {
			s.WriteString("Select category:\n")
			for i, cat := range categories {
				s.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, cat))
			}
			s.WriteString("  [a] All categories\n")
		} else {
			s.WriteString("No categories yet.\n")
		}
		s.WriteString("\n")
	case ModeFilter:
		// Filter view is just showing the message
	default:
		// List view
		if len(m.tasks) == 0 {
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorEmpty)).
				Italic(true)
			s.WriteString(emptyStyle.Render("No tasks yet. Press 'n' to create one!"))
			s.WriteString("\n\n")
		} else {
			if m.viewAsTable {
				// Table view
				headerStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color(colorTitle)).
					BorderStyle(lipgloss.NormalBorder()).
					BorderBottom(true).
					BorderForeground(lipgloss.Color(colorHelp))

				s.WriteString(headerStyle.Render(fmt.Sprintf("%-3s %-50s %-20s", "Status", "Description", "Category")))
				s.WriteString("\n")

				// Render tasks as table rows
				for i, task := range m.tasks {
					cursor := " "
					if i == m.cursor {
						cursor = ">"
					}

					statusIcon := m.getStatusIcon(task.Status)
					statusColor := m.getStatusColor(task.Status)

					// Truncate description if too long
					description := task.Description
					if len(description) > 48 {
						description = description[:45] + "..."
					}

					// Format category
					category := string(task.Category)
					if len(category) > 18 {
						category = category[:15] + "..."
					}

					categoryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorCategory)).Italic(true)
					categoryText := ""
					if category != "" {
						categoryText = categoryStyle.Render(category)
					}

					// Build row
					row := fmt.Sprintf("%-3s ", cursor)
					statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
					row += statusStyle.Render(fmt.Sprintf("%-3s", statusIcon))
					row += " "

					if i == m.cursor {
						descStyle := lipgloss.NewStyle().
							Bold(true).
							Foreground(lipgloss.Color(colorTitle))
						row += descStyle.Render(fmt.Sprintf("%-50s", description))
					} else {
						taskStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
						row += taskStyle.Render(fmt.Sprintf("%-50s", description))
					}

					row += " " + fmt.Sprintf("%-20s", categoryText)

					s.WriteString(row)
					s.WriteString("\n")
				}
			} else {
				// List view
				for i, task := range m.tasks {
					cursor := " "
					if i == m.cursor {
						cursor = ">"
					}

					statusIcon := m.getStatusIcon(task.Status)
					statusColor := m.getStatusColor(task.Status)

					taskStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))

					line := fmt.Sprintf("%s %s %s", cursor, statusIcon, task.Description)
					if task.Category != "" {
						categoryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorCategory)).Italic(true)
						line += " " + categoryStyle.Render(fmt.Sprintf("[%s]", string(task.Category)))
					}

					if i == m.cursor {
						line = lipgloss.NewStyle().
							Bold(true).
							Foreground(lipgloss.Color(colorTitle)).
							Render(line)
					} else {
						line = taskStyle.Render(line)
					}
					s.WriteString(line)
					s.WriteString("\n")
				}
			}
			s.WriteString("\n")
		}
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorHelp)).
		Faint(true)

	if m.viewMode == ModeList {
		filterInfo := "all"
		if m.filterStatus != nil && m.filterCategory != nil {
			filterInfo = fmt.Sprintf("%s + %s", string(*m.filterStatus), string(*m.filterCategory))
		} else if m.filterStatus != nil {
			filterInfo = string(*m.filterStatus)
		} else if m.filterCategory != nil {
			filterInfo = string(*m.filterCategory)
		}
		viewStyle := "table"
		if !m.viewAsTable {
			viewStyle = "list"
		}
		help := fmt.Sprintf("[n] new task\n[e] edit task\n[v] toggle view (%s)\n[d] done/undone\n[i] in-progress\n[p] pending\n[x] delete\n[f] filter (%s)\n[q] quit", viewStyle, filterInfo)
		s.WriteString(helpStyle.Render(help))
	}

	return s.String()
}

func (m model) getStatusIcon(status TaskStatus) string {
	switch status {
	case StatusDone:
		return "✓"
	case StatusInProgress:
		return "⟳"
	default:
		return "○"
	}
}

func (m model) getStatusColor(status TaskStatus) string {
	switch status {
	case StatusDone:
		return colorDone
	case StatusInProgress:
		return colorInProgress
	default:
		return colorPending
	}
}
