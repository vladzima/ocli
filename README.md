# OCLI - Terminal Outliner

![OCLI Screenshot](https://cdn.arbatov.dev/762J4rdLGoGAJEECGJFYngGPq.png)


A terminal-based outliner and task manager, built with Go. Run locally or as an SSH app!

![OCLI Demo](https://img.shields.io/badge/Status-Ready-green)

## Features

- 📝 **Hierarchical bullet points** with unlimited nesting
- ✅ **Task management** with checkboxes and completion tracking
- 🎨 **Color coding** for bullets (5 color options)
- 🔍 **Zoom functionality** to focus on specific branches
- ⚙️ **Settings system** with visual hierarchy lines toggle
- 💾 **Persistent storage** with JSON-based save/load
- 🎯 **Vim-style navigation** with keyboard shortcuts
- 🌳 **Visual hierarchy** with optional tree-style connectors

## Installation

### Method 1: Install from source (recommended)

Requirements: Go 1.19+ installed on your system

```bash
go install github.com/vladzima/ocli@latest
```

Make sure your `$GOPATH/bin` is in your `$PATH`. Add this to your shell profile if needed:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Method 2: Download pre-built binaries

Download the latest release for your platform from the [releases page](https://github.com/vladzima/ocli/releases).

### Method 3: Build from source

```bash
git clone https://github.com/vladzima/ocli.git
cd ocli
make build
# Or simply: go build -o ocli .
```

### Verify installation

```bash
ocli --version
```

## Usage

Simply run the command in your terminal:

```bash
ocli
```

On first run, OCLI creates a config directory at `~/.config/ocli/` with default tutorial content.

## Use as remote SSH app

Use OCLI remotely with persistent cloud storage:

```bash
# Install remote client (one-time setup)
curl -fsSL https://raw.githubusercontent.com/vladzima/ocli/main/cmd/ocli-ssh/install-client.sh | bash

# Connect remotely
ocli
```

Your data is automatically saved on the server and synced across sessions. Perfect for accessing your notes from anywhere!

> **Alternative**: Use `ssh ocli` after running the [SSH config setup](cmd/ocli-ssh#quick-start).

## Keyboard Shortcuts

### Navigation
- `↑↓` or `j/k` - Navigate up/down
- `←` - Zoom out
- `→` - Zoom in

### Editing
- `Enter` - Create new bullet
- `e` - Edit selected bullet
- `d` - Delete selected bullet

### Organization
- `Tab` - Indent (move right)
- `Shift+Tab` - Outdent (move left)
- `Shift+↑↓` - Move bullet up/down
- `Space` - Collapse/expand

### Formatting
- `c` - Cycle bullet color
- `t` - Toggle task mode
- `x` - Mark task complete/incomplete

### Other
- `h` - Show help screen
- `s` - Open settings
- `q` - Quit (auto-saves)

## Data Storage

OCLI automatically saves your data to `~/.config/ocli/data.json`. All changes are auto-saved when you:
- Add/edit/delete bullets
- Change settings
- Quit the application

**Update Safety**: Your data is always preserved when updating OCLI. Tutorial content only appears for new installations - existing users keep all their data intact.

## Configuration

Settings are stored in the same JSON file and include:
- Hierarchy lines display toggle
- Future customization options

## Technical Details

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling and layout
- [Bubbles](https://github.com/charmbracelet/bubbles) - Common UI components

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Author

Created by [Vlad Arbatov](https://github.com/vladzima)