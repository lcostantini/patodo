# patodo

A simple and elegant terminal user interface for managing TODO tasks.

## Features

- ✅ Create tasks with descriptions and categories
- 📋 List all tasks with filtering
- 🔄 Change task states (pending, in-progress, done)
- 🔍 Filter tasks by status or category
- 🏷️  Organize tasks with custom categories
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
- `n` - Create new task (prompts for description, then category)
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

## Task Categories

When creating a task, you can optionally assign it a category (e.g., "work", "personal"). Categories help organize tasks and can be used for filtering. Leave blank to create a task without a category.
