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

---

## 🧩 Installation

```
bash
go install github.com/yourusername/go-walker@latest
Or build manually:
```

```
bash
git clone https://github.com/yourusername/go-walker.git
cd go-walker
go build -ldflags="-s -w" -o walker
```

Then run:

```
bash
./walker
```

### ⌨️ Key Bindings

#### Key    Action

- TAB    Switch between panels

- ↑ / ↓    Move cursor

- -→ / ENTER    Enter directory / open file

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


#### 📸 Preview
(screenshot or GIF can go here later)

#### 📄 License
MIT License © 2025 Your Name

🛠️ Built With
Go

tcell

yaml
