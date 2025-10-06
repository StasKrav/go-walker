# go-walker

> A lightweight terminal file manager written in Go.

`go-walker` is a fast and minimalistic **TUI file manager** built with [tcell](https://github.com/gdamore/tcell).  
It allows you to browse directories, copy/move/delete files, and manage bookmarks — all from your terminal.

---

## ✨ Features

- Dual-panel navigation (bookmarks + file list)
- Create and delete bookmarks
- Copy, move, and delete files or directories
- Show/hide hidden files
- Open files with default system apps (`xdg-open`)
- Lightweight and dependency-free
- Configurable through JSON config file
- Comprehensive logging for debugging
- Built-in help system (press `?` for key bindings)

---

## 🧩 Installation

```bash
go install github.com/yourusername/go-walker@latest
```

Or build manually:

```bash
bash
git clone https://github.com/yourusername/go-walker.git
cd go-walker
go build -ldflags="-s -w" -o walker cmd/walker/main.go
```

Then run:

```bash
bash
./walker
```

### ⌨️ Key Bindings

#### Key    Action

- TAB    Switch between panels

- ↑ / ↓    Move cursor

- → / ENTER    Enter directory / open file

- ←    Go to parent directory

- a    Add bookmark

- d    Delete bookmark

- m    Mark file/folder for move

- c    Mark file/folder for copy

- p    Paste (move/copy)

- .    Toggle hidden files

- r    Refresh directory

- DELETE    Delete file/folder

- ESC    Exit
- ?      Show help (key bindings)


#### 📸 Preview
(screenshot or GIF can go here later)

#### 🏗️ Architecture

The project is structured into several packages for better maintainability:

- `cmd/walker` - Main entry point
- `internal/app` - Core application logic
- `internal/bookmarks` - Bookmarks management
- `internal/config` - Configuration handling
- `internal/fs` - File system operations
- `internal/logging` - Logging utilities
- `internal/ui` - User interface components

#### 🧪 Testing

The project includes unit tests for all major components:

```bash
go test ./internal/...
```

#### 📄 License
MIT License © 2025 Your Name

#### 🛠️ Built With
- Go
- tcell
