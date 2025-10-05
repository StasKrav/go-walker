package main


import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "time"


"github.com/gdamore/tcell/v2"

)


type Panel struct {
    x, y, w, h int
    items      []string
    cursor     int
    active     bool
    border     bool
    path       string
    offset     int
}


var (
    homeDir    string
    showHidden = false


moveSrc   string
copySrc   string
moveReady = false
copyReady = false

)


// ---------------- bookmarks ----------------
func bookmarksFile() string {
    return filepath.Join(homeDir, ".myfm_bookmarks.json")
}


func saveBookmarks(bookmarks []string) {
    data, _ := json.MarshalIndent(bookmarks, "", "  ")
    _ = os.WriteFile(bookmarksFile(), data, 0644)
}


func loadBookmarks(home string) []string {
    file := bookmarksFile()
    data, err := os.ReadFile(file)
    if err != nil {
        bookmarks := []string{
            "Home",
            filepath.Join(home, "Desktop"),
            filepath.Join(home, "Documents"),
            filepath.Join(home, "Downloads"),
            filepath.Join(home, "Music"),
            filepath.Join(home, "Pictures"),
            filepath.Join(home, "Videos"),
        }
        saveBookmarks(bookmarks)
        return bookmarks
    }
    var bookmarks []string
    if err := json.Unmarshal(data, &bookmarks); err != nil {
        return []string{"Home"}
    }
    return bookmarks
}


// ---------------- drawing ----------------
func visibleCount(p *Panel) int {
    vc := p.h
    if p.border {
        vc -= 2
    }
    if vc < 0 {
        vc = 0
    }
    return vc
}


func drawPanel(s tcell.Screen, p *Panel) {
    // Выбираем стиль границы: активная панель — как было (HotPink), неактивная — серая
    borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
    if p.active {
        borderStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite)
    } else {
        borderStyle = tcell.StyleDefault.Foreground(tcell.ColorGray)
    }


if p.border {
    for cx := p.x; cx < p.x+p.w; cx++ {
        s.SetContent(cx, p.y, '─', nil, borderStyle)
        s.SetContent(cx, p.y+p.h-1, '─', nil, borderStyle)
    }
    for cy := p.y; cy < p.y+p.h; cy++ {
        s.SetContent(p.x, cy, '│', nil, borderStyle)
        s.SetContent(p.x+p.w-1, cy, '│', nil, borderStyle)
    }
    s.SetContent(p.x, p.y, '┌', nil, borderStyle)
    s.SetContent(p.x+p.w-1, p.y, '┐', nil, borderStyle)
    s.SetContent(p.x, p.y+p.h-1, '└', nil, borderStyle)
    s.SetContent(p.x+p.w-1, p.y+p.h-1, '┘', nil, borderStyle)
}

maxItems := visibleCount(p)

for i := 0; i < maxItems && i+p.offset < len(p.items); i++ {
    idx := i + p.offset
    rawName := p.items[idx]

    display := rawName
    if p.path == "" {
        if rawName == "Home" {
            display = "Home"
        } else {
            display = filepath.Base(rawName)
        }
    }

    isDir := false
    var fullPath string
    if p.path == "" {
        if rawName == "Home" {
            fullPath = homeDir
        } else {
            fullPath = rawName
        }
    } else {
        fullPath = filepath.Join(p.path, rawName)
    }
    if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
        isDir = true
    }

    // Если панель активна — используем старую логику (директории белым, файлы серым);
    // если неактивна — делаем всю панель "серой" по цвету текста.
    var styleLine tcell.Style
    if p.active {
        styleLine = tcell.StyleDefault.Foreground(tcell.ColorGray)
        if isDir {
            styleLine = tcell.StyleDefault.Foreground(tcell.ColorWhite)
        }
        // Подсветка текущей строки только на активной панели
        if idx == p.cursor && p.active {
            styleLine = styleLine.Reverse(true)
        }
    } else {
        // неактивная панель — серый текст (все строки)
        styleLine = tcell.StyleDefault.Foreground(tcell.ColorGray)
    }

    yOffset := i
    if p.border {
        yOffset = i + 1
    }

    // текст в пределах панели: оставляем 1 символ отступа слева и справа
    maxChars := p.w - 2
    if maxChars < 0 {
        maxChars = 0
    }
    runes := []rune(display)
    for j := 0; j < maxChars && j < len(runes); j++ {
        s.SetContent(p.x+1+j, p.y+yOffset, runes[j], nil, styleLine)
    }
}

}


func drawText(s tcell.Screen, x, y int, text string, style tcell.Style, width int) {
    runes := []rune(text)
    for i := 0; i < width; i++ {
        ch := ' '
        if i < len(runes) {
            ch = runes[i]
        }
        s.SetContent(x+i, y, ch, nil, style)
    }
}


// ---------------- fs utils ----------------
func humanSize(size int64) string {
    const unit = 1024
    if size < unit {
        return fmt.Sprintf("%dB", size)
    }
    div, exp := int64(unit), 0
    for n := size / unit; n >= unit && exp < 3; n /= unit {
        div *= unit
        exp++
    }
    suffix := []string{"K", "M", "G"}[exp]
    value := float64(size) / float64(div)
    if value < 10 {
        return fmt.Sprintf("%.1f%s", value, suffix)
    }
    return fmt.Sprintf("%.0f%s", value, suffix)
}


func loadDir(path string) ([]string, error) {
    entries, err := os.ReadDir(path)
    if err != nil {
        return nil, err
    }


var hiddenDirs, dirs, files, hiddenFiles []string

for _, e := range entries {
    name := e.Name()
    isHidden := len(name) > 0 && name[0] == '.'
    if e.IsDir() {
        if isHidden {
            hiddenDirs = append(hiddenDirs, name)
        } else {
            dirs = append(dirs, name)
        }
    } else {
        if isHidden {
            hiddenFiles = append(hiddenFiles, name)
        } else {
            files = append(files, name)
        }
    }
}

sort.Strings(hiddenDirs)
sort.Strings(dirs)
sort.Strings(files)
sort.Strings(hiddenFiles)

if showHidden {
    res := append([]string{}, hiddenDirs...)
    res = append(res, dirs...)
    res = append(res, files...)
    res = append(res, hiddenFiles...)
    return res, nil
}
res := append([]string{}, dirs...)
res = append(res, files...)
return res, nil

}


func fileInfo(path string) string {
    info, err := os.Stat(path)
    if err != nil {
        return "error"
    }


mode := info.Mode().String()
modTime := info.ModTime().Format("2006-01-02 15:04")

if info.IsDir() {
    entries, _ := os.ReadDir(path)
    total := len(entries)
    hidden := 0
    for _, e := range entries {
        if len(e.Name()) > 0 && e.Name()[0] == '.' {
            hidden++
        }
    }
    return fmt.Sprintf("%-12s%6d%6d%10s   %s", mode, total, hidden, "-", modTime)
}

sizeStr := humanSize(info.Size())
return fmt.Sprintf("%-12s%6d%6d%10s   %s", mode, 0, 0, sizeStr, modTime)

}


// ---------------- modal ----------------
func drawModal(s tcell.Screen, text string) {
    sw, sh := s.Size()
    w := len(text) + 6
    if w < 20 {
        w = 20
    }
    h := 5
    x := (sw - w) / 2
    y := (sh - h) / 2


bg := tcell.ColorDefault
borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)
textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)

for cy := y; cy < y+h; cy++ {
    for cx := x; cx < x+w; cx++ {
        s.SetContent(cx, cy, ' ', nil, textStyle)
    }
}
for cx := x; cx < x+w; cx++ {
    s.SetContent(cx, y, '─', nil, borderStyle)
    s.SetContent(cx, y+h-1, '─', nil, borderStyle)
}
for cy := y; cy < y+h; cy++ {
    s.SetContent(x, cy, '│', nil, borderStyle)
    s.SetContent(x+w-1, cy, '│', nil, borderStyle)
}
s.SetContent(x, y, '┌', nil, borderStyle)
s.SetContent(x+w-1, y, '┐', nil, borderStyle)
s.SetContent(x, y+h-1, '└', nil, borderStyle)
s.SetContent(x+w-1, y+h-1, '┘', nil, borderStyle)

drawText(s, x+3, y+2, text, textStyle, w-6)

}


func ensureCursorBounds(p *Panel) {
    if len(p.items) == 0 {
        p.cursor = 0
        p.offset = 0
        return
    }
    if p.cursor < 0 {
        p.cursor = 0
    }
    if p.cursor >= len(p.items) {
        p.cursor = len(p.items) - 1
    }
    // скорректируем offset чтобы курсор был виден
    vc := visibleCount(p)
    if vc == 0 {
        p.offset = 0
        return
    }
    if p.cursor < p.offset {
        p.offset = p.cursor
    }
    if p.cursor >= p.offset+vc {
        p.offset = p.cursor - vc + 1
    }
}


// ---------------- main ----------------
func main() {
    const modalDuration = 750 * time.Millisecond  // время показа уведомлений


hd, _ := os.UserHomeDir()
homeDir = hd

s, err := tcell.NewScreen()
if err != nil {
    log.Fatal(err)
}
if err := s.Init(); err != nil {
    log.Fatal(err)
}
defer s.Fini()
s.Clear()

fileItems, _ := loadDir(homeDir)

screenW, screenH := s.Size()
leftW := 24
rightX := leftW + 1
rightW := screenW - rightX - 1
rightH := screenH - 6

bookmarks := loadBookmarks(homeDir)

sidebar := &Panel{
    x: 0, y: 4, w: leftW, h: screenH - 4,
    items:  bookmarks,
    active: true,
    border: false,
    path:   "",
}
filelist := &Panel{
    x: rightX, y: 3, w: rightW, h: rightH,
    items:  fileItems,
    active: false,
    border: true,
    path:   homeDir,
}

ensureCursorBounds(sidebar)
ensureCursorBounds(filelist)

panels := []*Panel{sidebar, filelist}
current := 0

modalActive := false
modalText := ""
modalTimer := time.Time{}

deleteIndex := -1
deleteFileIndex := -1

// канал для событий от tcell — читаем PollEvent в горутине и шлем событие сюда
events := make(chan tcell.Event, 16)
go func() {
    for {
        ev := s.PollEvent()
        // PollEvent обычно не возвращает nil, но защищаемся
        if ev == nil {
            close(events)
            return
        }
        events <- ev
    }
}()

quit := false
for !quit {
    // подготовка канала таймера (nil если таймер не нужен)
    var timerChan <-chan time.Time
    if modalActive && !modalTimer.IsZero() {
        d := time.Until(modalTimer)
        if d <= 0 {
            // таймер уже истёк — сразу закрываем модалку
            modalActive = false
            modalTimer = time.Time{}
        } else {
            timerChan = time.After(d)
        }
    }

    // --- отрисовка ---
    s.Clear()

    addrStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
    drawText(s, rightX+1, 1, fmt.Sprintf(" %s ", filelist.path), addrStyle, rightW-2)

    drawPanel(s, sidebar)
    drawPanel(s, filelist)

    status := " Ready "
    if current == 1 && len(filelist.items) > 0 && filelist.cursor >= 0 && filelist.cursor < len(filelist.items) {
        name := filelist.items[filelist.cursor]
        path := filepath.Join(filelist.path, name)
        status = fileInfo(path)
    }
    statusStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
    drawText(s, rightX+1, filelist.y+filelist.h+1, status, statusStyle, rightW-2)

    if modalActive {
        drawModal(s, modalText)
    }

    s.Show()
    // --- конец отрисовки ---

    // ждём либо событие, либо таймер
    select {
    case ev, ok := <-events:
        if !ok {
            // канал закрыт — безопасно выйти
            quit = true
            continue
        }
        switch ev := ev.(type) {
        case *tcell.EventKey:
            // если модалка активна и у неё ZERO-таймер — это подтверждение (ожидаем y/n/esc)
            if modalActive && modalTimer.IsZero() {
                switch ev.Key() {
                case tcell.KeyEscape:
                    modalActive = false
                    deleteIndex = -1
                    deleteFileIndex = -1
                case tcell.KeyRune:
                    if ev.Rune() == 'y' {
                        if deleteIndex >= 0 && deleteIndex < len(sidebar.items) {
                            // удаление закладки
                            sidebar.items = append(sidebar.items[:deleteIndex], sidebar.items[deleteIndex+1:]...)
                            if sidebar.cursor >= len(sidebar.items) {
                                sidebar.cursor = len(sidebar.items) - 1
                            }
                            saveBookmarks(sidebar.items)
                            deleteIndex = -1
                            modalActive = true
                            modalText = "Bookmark deleted"
                            modalTimer = time.Now().Add(modalDuration)
                            ensureCursorBounds(sidebar)
                        } else if deleteFileIndex >= 0 && deleteFileIndex < len(filelist.items) {
                            // удаление файла/директории
                            fullPath := filepath.Join(filelist.path, filelist.items[deleteFileIndex])
                            err := os.RemoveAll(fullPath)
                            if err != nil {
                                modalText = fmt.Sprintf("Delete error: %v", err)
                            } else {
                                modalText = fmt.Sprintf("Deleted: %s", filelist.items[deleteFileIndex])
                                if items, err := loadDir(filelist.path); err == nil {
                                    filelist.items = items
                                    filelist.cursor = 0
                                    filelist.offset = 0
                                    ensureCursorBounds(filelist)
                                }
                            }
                            deleteFileIndex = -1
                            modalActive = true
                            modalTimer = time.Now().Add(modalDuration)
                        }
                    }
                    if ev.Rune() == 'n' {
                        modalActive = false
                        deleteIndex = -1
                        deleteFileIndex = -1
                    }
                }
                // после обработки подтверждения возвращаемся к верхнему циклу
                continue
            }

            // обычная обработка клавиш
            switch ev.Key() {
            case tcell.KeyEscape:
                if modalActive {
                    // закрываем модалку, не выходим
                    modalActive = false
                    deleteIndex = -1
                    deleteFileIndex = -1
                    modalTimer = time.Time{}
                } else {
                    quit = true
                }

            case tcell.KeyCtrlC:
                quit = true

            case tcell.KeyTAB:
                panels[current].active = false
                current = (current + 1) % len(panels)
                panels[current].active = true

            case tcell.KeyUp:
                if panels[current].cursor > 0 {
                    panels[current].cursor--
                }
                if panels[current].cursor < panels[current].offset {
                    panels[current].offset--
                }
                ensureCursorBounds(panels[current])

            case tcell.KeyDown:
                if panels[current].cursor < len(panels[current].items)-1 {
                    panels[current].cursor++
                }
                if panels[current].cursor >= panels[current].offset+visibleCount(panels[current]) {
                    panels[current].offset++
                }
                ensureCursorBounds(panels[current])

            case tcell.KeyRight, tcell.KeyEnter:
                if current == 1 && len(filelist.items) > 0 && filelist.cursor >= 0 && filelist.cursor < len(filelist.items) {
                    name := filelist.items[filelist.cursor]
                    fullPath := filepath.Join(filelist.path, name)
                    if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
                        if items, err := loadDir(fullPath); err == nil {
                            filelist.path = fullPath
                            filelist.items = items
                            filelist.cursor = 0
                            filelist.offset = 0
                            ensureCursorBounds(filelist)
                        }
                    } else {
                        exec.Command("xdg-open", fullPath).Start()
                    }
                } else if current == 0 && len(sidebar.items) > 0 {
                    bookmark := sidebar.items[sidebar.cursor]
                    var dir string
                    if bookmark == "Home" {
                        dir = homeDir
                    } else {
                        dir = bookmark
                    }
                    if items, err := loadDir(dir); err == nil {
                        filelist.path = dir
                        filelist.items = items
                        filelist.cursor = 0
                        filelist.offset = 0
                        ensureCursorBounds(filelist)
                    }
                }

            case tcell.KeyLeft:
                if current == 1 && filelist.path != "/" {
                    parent := filepath.Dir(filelist.path)
                    if items, err := loadDir(parent); err == nil {
                        filelist.path = parent
                        filelist.items = items
                        filelist.cursor = 0
                        filelist.offset = 0
                        ensureCursorBounds(filelist)
                    }
                }

            case tcell.KeyDelete: // удаление файла/директории (требует подтверждения)
                if current == 1 && len(filelist.items) > 0 {
                    idx := filelist.cursor
                    if idx >= 0 && idx < len(filelist.items) {
                        name := filelist.items[idx]
                        modalText = fmt.Sprintf("Delete \"%s\"? (y/n)", name)
                        modalActive = true
                        modalTimer = time.Time{} // подтверждение — без таймера
                        deleteFileIndex = idx
                    }
                }

            case tcell.KeyRune:
                switch ev.Rune() {
                case 'a':
                    path := filelist.path
                    exists := false
                    for _, bm := range sidebar.items {
                        if bm == path {
                            exists = true
                            break
                        }
                    }
                    if !exists {
                        sidebar.items = append(sidebar.items, path)
                        saveBookmarks(sidebar.items)
                        modalText = "Bookmark added"
                        modalActive = true
                        modalTimer = time.Now().Add(modalDuration)
                        ensureCursorBounds(sidebar)
                    }

                case 'd':
                    if current == 0 && len(sidebar.items) > 0 {
                        idx := sidebar.cursor
                        if idx >= 0 && sidebar.items[idx] != "Home" {
                            modalText = fmt.Sprintf("Delete bookmark \"%s\"? (y/n)", filepath.Base(sidebar.items[idx]))
                            modalActive = true
                            modalTimer = time.Time{} // подтверждение — без таймера
                            deleteIndex = idx
                        }
                    }

                case '.':
                    showHidden = !showHidden
                    if items, err := loadDir(filelist.path); err == nil {
                        filelist.items = items
                        filelist.cursor = 0
                        filelist.offset = 0
                        ensureCursorBounds(filelist)
                    }
                    modalText = "Toggled hidden files"
                    modalActive = true
                    modalTimer = time.Now().Add(modalDuration)

                case 'm':
                    if current == 1 && len(filelist.items) > 0 {
                        name := filelist.items[filelist.cursor]
                        moveSrc = filepath.Join(filelist.path, name)
                        moveReady = true
                        copyReady = false
                        modalText = fmt.Sprintf("Marked for move: %s", name)
                        modalActive = true
                        modalTimer = time.Now().Add(modalDuration)
                    }

                case 'c':
                    if current == 1 && len(filelist.items) > 0 {
                        name := filelist.items[filelist.cursor]
                        copySrc = filepath.Join(filelist.path, name)
                        copyReady = true
                        moveReady = false
                        modalText = fmt.Sprintf("Marked for copy: %s", name)
                        modalActive = true
                        modalTimer = time.Now().Add(modalDuration)
                    }

                case 'p':
                    if moveReady {
                        dst := filepath.Join(filelist.path, filepath.Base(moveSrc))
                        err := os.Rename(moveSrc, dst)
                        if err != nil {
                            modalText = fmt.Sprintf("Move error: %v", err)
                        } else {
                            modalText = fmt.Sprintf("Moved to: %s", filelist.path)
                            moveReady = false
                            moveSrc = ""
                            if items, err := loadDir(filelist.path); err == nil {
                                filelist.items = items
                                filelist.cursor = 0
                                filelist.offset = 0
                                ensureCursorBounds(filelist)
                            }
                        }
                        modalActive = true
                        modalTimer = time.Now().Add(modalDuration)

                    } else if copyReady {
                        dst := filepath.Join(filelist.path, filepath.Base(copySrc))
                        err := copyRecursive(copySrc, dst)
                        if err != nil {
                            modalText = fmt.Sprintf("Copy error: %v", err)
                        } else {
                            modalText = fmt.Sprintf("Copied to: %s", filelist.path)
                            copyReady = false
                            copySrc = ""
                            if items, err := loadDir(filelist.path); err == nil {
                                filelist.items = items
                                filelist.cursor = 0
                                filelist.offset = 0
                                ensureCursorBounds(filelist)
                            }
                        }
                        modalActive = true
                        modalTimer = time.Now().Add(modalDuration)
                    }

                case 'r':
                    // refresh текущей директории
                    if items, err := loadDir(filelist.path); err == nil {
                        filelist.items = items
                        filelist.cursor = 0
                        filelist.offset = 0
                        ensureCursorBounds(filelist)
                        modalText = "Refreshed"
                    } else {
                        modalText = fmt.Sprintf("Refresh error: %v", err)
                    }
                    modalActive = true
                    modalTimer = time.Now().Add(modalDuration)
                }
            }
        }
    case <-timerChan:
        // таймер сработал — закрываем модалку
        modalActive = false
        modalTimer = time.Time{}
    }
}

// (опционально) закрываем канал, чтобы горутина завершилась корректно

}


// ---------------- copy util ----------------
func copyRecursive(src, dst string) error {
    info, err := os.Stat(src)
    if err != nil {
        return err
    }


if info.IsDir() {
    if err := os.MkdirAll(dst, info.Mode()); err != nil {
        return err
    }
    entries, err := os.ReadDir(src)
    if err != nil {
        return err
    }
    for _, e := range entries {
        err = copyRecursive(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name()))
        if err != nil {
            return err
        }
    }
} else {
    data, err := os.ReadFile(src)
    if err != nil {
        return err
    }
    err = os.WriteFile(dst, data, info.Mode())
    if err != nil {
        return err
    }
}
return nil

}
