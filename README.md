# patodo

A simple and elegant terminal user interface for managing TODO tasks.

## Features

- ✅ Create and edit tasks with descriptions and categories
- 📋 List all tasks with filtering
- 🔄 Change task states (pending, in-progress, done)
- 🔍 Filter tasks by status or category
- 🏷️  Organize tasks with custom categories
- 👁️  Toggle between table and list view modes
- 💾 Persistent storage in `~/.config/patodo/tasks.json`
- ⌨️  Keyboard-driven interface

## Installation

### Quick Install (recommended)
```bash
./install.sh
```

### Manual Build
```bash
go build -o patodo
```

### Install with Go
```bash
go install
```

## Usage

```bash
patodo
```

## Keyboard Shortcuts

### Main View
- `n` - Create new task
- `e` - Edit selected task
- `v` - Toggle between table and list view
- `d` - Toggle task done/pending
- `i` - Mark task as in-progress
- `p` - Mark task as pending
- `x` - Delete task
- `f` - Open filter menu
- `↑/↓` or `j/k` - Navigate tasks
- `q` or `Ctrl+C` - Quit

### Filter Menu (press `f`)
- `a` - Show all tasks
- `p` - Show pending tasks only
- `i` - Show in-progress tasks only
- `d` - Show done tasks only
- `c` - Filter by category
- `ESC` - Cancel filter

### Category Filter (press `c` in filter menu)
- `1-9` - Select category by number
- `a` - Show all categories
- `ESC` - Cancel

### Create/Edit Mode
- `Tab` - Switch between description and category fields
- `Enter` - Save task
- `ESC` - Cancel

## Task Categories

When creating or editing a task, you must assign it a category (e.g., "work", "personal", "shopping"). Categories help organize tasks and can be used for filtering.

## Views

patodo supports two view modes:
- **Table view** (default) - Displays tasks in a structured table format with columns for status, description, and category
- **List view** - Shows tasks in a compact list format

Press `v` to toggle between views.
