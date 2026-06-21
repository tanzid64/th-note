package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Palette (Catppuccin Mocha)
var (
	mauve  = lipgloss.Color("#cba6f7")
	pink   = lipgloss.Color("#f5c2e7")
	blue   = lipgloss.Color("#89b4fa")
	teal   = lipgloss.Color("#94e2d5")
	green  = lipgloss.Color("#a6e3a1")
	yellow = lipgloss.Color("#f9e2af")
	subtle = lipgloss.Color("#6c7086")
	text   = lipgloss.Color("#cdd6f4")
	base   = lipgloss.Color("#1e1e2e")
)

// gradient colors used to paint the logo line by line.
var logoColors = []color.Color{mauve, pink, blue, teal, green}

var logo = []string{
	" ████████╗██╗  ██╗      ███╗   ██╗ ██████╗ ████████╗███████╗",
	" ╚══██╔══╝██║  ██║      ████╗  ██║██╔═══██╗╚══██╔══╝██╔════╝",
	"    ██║   ███████║█████╗██╔██╗ ██║██║   ██║   ██║   █████╗  ",
	"    ██║   ██╔══██║╚════╝██║╚██╗██║██║   ██║   ██║   ██╔══╝  ",
	"    ██║   ██║  ██║      ██║ ╚████║╚██████╔╝   ██║   ███████╗",
	"    ╚═╝   ╚═╝  ╚═╝      ╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚══════╝",
}

var (
	taglineStyle = lipgloss.NewStyle().Foreground(text).Italic(true)

	keyStyle = lipgloss.NewStyle().
			Foreground(base).Background(mauve).Bold(true).Padding(0, 1)

	keyDescStyle = lipgloss.NewStyle().Foreground(subtle)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).BorderForeground(mauve).Padding(1, 4)

	// Editor chrome
	editorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).BorderForeground(blue).Padding(0, 1)

	fieldLabelStyle = lipgloss.NewStyle().Foreground(subtle).Bold(true)

	fieldLabelActiveStyle = lipgloss.NewStyle().Foreground(mauve).Bold(true)

	dividerStyle = lipgloss.NewStyle().Foreground(subtle)

	statusBarStyle = lipgloss.NewStyle().Foreground(subtle)

	savedStyle = lipgloss.NewStyle().Foreground(green)

	dirtyStyle = lipgloss.NewStyle().Foreground(yellow)

	helpStyle = lipgloss.NewStyle().Foreground(subtle)

	helpKeyStyle = lipgloss.NewStyle().Foreground(teal)

	// In-note search match highlight
	matchStyle        = lipgloss.NewStyle().Foreground(base).Background(yellow)
	currentMatchStyle = lipgloss.NewStyle().Foreground(base).Background(pink).Bold(true)

	listTitleStyle = lipgloss.NewStyle().Foreground(mauve).Bold(true)

	red             = lipgloss.Color("#f38ba8")
	confirmStyle    = lipgloss.NewStyle().Foreground(red).Bold(true)
	confirmKeyStyle = lipgloss.NewStyle().Foreground(base).Background(red).Bold(true).Padding(0, 1)

	// Markdown preview styles
	mdH1     = lipgloss.NewStyle().Foreground(mauve).Bold(true).Underline(true)
	mdH2     = lipgloss.NewStyle().Foreground(pink).Bold(true)
	mdH3     = lipgloss.NewStyle().Foreground(blue).Bold(true)
	mdBullet = lipgloss.NewStyle().Foreground(teal)
	mdQuote  = lipgloss.NewStyle().Foreground(subtle).Italic(true)
	mdCode   = lipgloss.NewStyle().Foreground(green).Background(base)
	mdBold   = lipgloss.NewStyle().Foreground(text).Bold(true)
	mdItalic = lipgloss.NewStyle().Foreground(text).Italic(true)
	mdText   = lipgloss.NewStyle().Foreground(text)
)

var (
	mdBoldRe   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdItalicRe = regexp.MustCompile(`\*([^*]+)\*`)
	mdCodeRe   = regexp.MustCompile("`([^`]+)`")
)

// renderMarkdown is a lightweight, dependency-free markdown styler for the
// preview pane. It handles headings, lists, blockquotes, fenced code, and
// inline bold/italic/code.
func renderMarkdown(src string, width int) string {
	var out []string
	inFence := false

	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			out = append(out, mdQuote.Render("│ "+trimmed))
			continue
		}
		if inFence {
			out = append(out, mdCode.Render("  "+line))
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "### "):
			out = append(out, mdH3.Render(strings.TrimPrefix(trimmed, "### ")))
		case strings.HasPrefix(trimmed, "## "):
			out = append(out, mdH2.Render(strings.TrimPrefix(trimmed, "## ")))
		case strings.HasPrefix(trimmed, "# "):
			out = append(out, mdH1.Render(strings.TrimPrefix(trimmed, "# ")))
		case strings.HasPrefix(trimmed, "> "):
			out = append(out, mdQuote.Render("┃ "+renderInline(strings.TrimPrefix(trimmed, "> "))))
		case strings.HasPrefix(trimmed, "- "), strings.HasPrefix(trimmed, "* "):
			out = append(out, mdBullet.Render("  • ")+renderInline(trimmed[2:]))
		case trimmed == "":
			out = append(out, "")
		default:
			out = append(out, renderInline(line))
		}
	}
	return strings.Join(out, "\n")
}

// renderInline styles inline markdown spans within a single line.
func renderInline(s string) string {
	s = mdCodeRe.ReplaceAllStringFunc(s, func(m string) string {
		return mdCode.Render(" " + strings.Trim(m, "`") + " ")
	})
	s = mdBoldRe.ReplaceAllStringFunc(s, func(m string) string {
		return mdBold.Render(strings.Trim(m, "*"))
	})
	s = mdItalicRe.ReplaceAllStringFunc(s, func(m string) string {
		return mdItalic.Render(strings.Trim(m, "*"))
	})
	if !strings.ContainsRune(s, '\x1b') {
		return mdText.Render(s)
	}
	return s
}

// screen identifies which view the app is currently showing.
type screen int

const (
	screenWelcome screen = iota
	screenEditor
	screenList
	screenDetail
)

// focusField identifies which editor field is active.
type focusField int

const (
	focusTitle focusField = iota
	focusBody
)

// autosaveMsg fires after a debounce interval to persist pending edits.
type autosaveMsg struct{ seq int }

const autosaveDelay = 700 * time.Millisecond

type model struct {
	screen screen
	width  int
	height int

	editSeq int // bumped on each edit; the autosave tick checks it

	title        textinput.Model
	body         textarea.Model
	focus        focusField
	preview      bool   // markdown preview toggled on?
	dirty        bool   // unsaved changes since last write
	savedAt      string // last save time, for the status bar
	savePath     string // current file on disk
	status       string // transient message (e.g. errors)
	editorReturn screen // screen to return to when leaving the editor

	// List screen
	list          list.Model
	confirmDelete bool // delete-confirmation prompt is showing
	deleteTarget  note // note pending deletion

	// Detail screen
	viewport   viewport.Model
	detail     note            // the note currently open for reading
	search     textinput.Model // in-note word search
	searching  bool            // search box focused/accepting input
	matchLines []int           // line indices in the detail body with matches
	matchIdx   int             // which match is currently focused
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Untitled note"
	ti.Prompt = ""
	ti.CharLimit = 120

	ta := textarea.New()
	ta.Placeholder = "Start writing… markdown supported. Try ctrl+t for a timestamp."
	ta.ShowLineNumbers = true

	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Your notes"
	l.SetShowTitle(true)
	l.Styles.Title = listTitleStyle

	se := textinput.New()
	se.Placeholder = "search word…"
	se.Prompt = "/"

	return model{
		screen:   screenWelcome,
		title:    ti,
		body:     ta,
		focus:    focusTitle,
		list:     l,
		viewport: viewport.New(),
		search:   se,
	}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg { return tea.RequestWindowSize() }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil

	case autosaveMsg:
		// Only save if this is the latest scheduled tick and there are
		// pending changes (debounce: ignore stale ticks).
		if msg.seq == m.editSeq && m.dirty {
			m.save()
		}
		return m, nil

	case tea.KeyPressMsg:
		switch m.screen {
		case screenEditor:
			return m.updateEditor(msg)
		case screenList:
			return m.updateList(msg)
		case screenDetail:
			return m.updateDetail(msg)
		default:
			return m.updateWelcome(msg)
		}
	}

	// Forward non-key messages (cursor blink, scrolling, etc.) to the active
	// component.
	var cmd tea.Cmd
	switch m.screen {
	case screenEditor:
		return m.routeToField(msg)
	case screenList:
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case screenDetail:
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) updateWelcome(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "n":
		return m.openEditor(note{}, screenWelcome)
	case "l":
		return m.openList()
	}
	return m, nil
}

// openList loads notes from disk into the list and shows the list screen.
func (m model) openList() (tea.Model, tea.Cmd) {
	notes, err := loadNotes()
	if err != nil {
		m.status = "could not read notes: " + err.Error()
	}
	items := make([]list.Item, len(notes))
	for i, n := range notes {
		items[i] = noteItem{n: n}
	}
	m.list.SetItems(items)
	m.confirmDelete = false
	m.deleteTarget = note{}
	m.screen = screenList
	m.layout()
	return m, nil
}

// openEditor switches to the editor, optionally pre-filled with an existing
// note (zero note = blank), remembering which screen to return to.
func (m model) openEditor(n note, ret screen) (tea.Model, tea.Cmd) {
	m.title.SetValue(n.title)
	m.body.SetValue(n.body)
	m.savePath = n.path
	m.dirty = false
	m.savedAt = ""
	m.status = ""
	m.preview = false
	m.focus = focusTitle
	m.body.Blur()
	m.editorReturn = ret
	m.screen = screenEditor
	m.layout()
	return m, m.title.Focus()
}

func (m model) updateEditor(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.save() // don't lose work inside the autosave debounce window
		return m, tea.Quit

	case "esc":
		m.save()
		m.title.Blur()
		m.body.Blur()
		if m.editorReturn == screenList {
			return m.openList()
		}
		m.screen = screenWelcome
		return m, nil

	case "tab", "shift+tab":
		m.toggleFocus()
		return m, m.focusCmd()

	case "ctrl+s":
		m.save()
		return m, nil

	case "ctrl+p":
		m.preview = !m.preview
		return m, nil

	case "ctrl+t":
		// Insert a timestamp into the body.
		if m.focus == focusBody {
			stamp := time.Now().Format("Mon 2006-01-02 15:04")
			m.body.InsertString(stamp)
			return m, m.markDirty()
		}
		return m, nil

	case "ctrl+b":
		// Insert markdown bold markers, leaving the cursor between them.
		if m.focus == focusBody {
			li := m.body.LineInfo()
			col := li.StartColumn + li.ColumnOffset
			m.body.InsertString("****")
			m.body.SetCursorColumn(col + 2)
			return m, m.markDirty()
		}
		return m, nil
	}

	// Default: route the key to the focused field, then track changes.
	prevTitle, prevBody := m.title.Value(), m.body.Value()
	updated, cmd := m.routeToField(msg)
	m = updated.(model)
	if m.title.Value() != prevTitle || m.body.Value() != prevBody {
		return m, tea.Batch(cmd, m.markDirty())
	}
	return m, cmd
}

func (m model) routeToField(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.focus == focusTitle {
		m.title, cmd = m.title.Update(msg)
	} else {
		m.body, cmd = m.body.Update(msg)
	}
	return m, cmd
}

// --- List screen -----------------------------------------------------------

func (m model) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// The delete-confirmation prompt captures all keys until answered.
	if m.confirmDelete {
		switch msg.String() {
		case "y", "Y", "enter":
			return m.deleteTargetNote()
		default: // n, N, esc, anything else cancels
			m.confirmDelete = false
			m.deleteTarget = note{}
			return m, nil
		}
	}

	// While the filter input is active, let the list own every key.
	if m.list.SettingFilter() {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		// If a filter is applied, the first esc clears it (handled by list).
		if m.list.IsFiltered() {
			break
		}
		m.screen = screenWelcome
		return m, nil
	case "n":
		return m.openEditor(note{}, screenList)
	case "enter":
		if it, ok := m.list.SelectedItem().(noteItem); ok {
			return m.openDetail(it.n)
		}
		return m, nil
	case "e":
		if it, ok := m.list.SelectedItem().(noteItem); ok {
			return m.openEditor(it.n, screenList)
		}
		return m, nil
	case "d":
		if it, ok := m.list.SelectedItem().(noteItem); ok {
			m.confirmDelete = true
			m.deleteTarget = it.n
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// deleteTargetNote removes the pending note from disk and reloads the list.
func (m model) deleteTargetNote() (tea.Model, tea.Cmd) {
	target := m.deleteTarget
	m.confirmDelete = false
	m.deleteTarget = note{}
	if target.path != "" {
		if err := os.Remove(target.path); err != nil {
			m.status = "delete failed: " + err.Error()
		}
	}
	return m.openList()
}

// --- Detail screen ---------------------------------------------------------

// openDetail shows a note in the read-only viewport.
func (m model) openDetail(n note) (tea.Model, tea.Cmd) {
	m.detail = n
	m.screen = screenDetail
	m.searching = false
	m.search.SetValue("")
	m.matchLines = nil
	m.matchIdx = 0
	m.layout()
	m.viewport.SetContent(m.renderDetail())
	m.viewport.GotoTop()
	return m, nil
}

func (m model) updateDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		switch msg.String() {
		case "esc":
			m.searching = false
			m.search.SetValue("")
			m.search.Blur()
			m.matchLines = nil
			m.viewport.SetContent(m.renderDetail())
			return m, nil
		case "enter":
			m.searching = false
			m.search.Blur()
			m.applySearch()
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.applySearch()
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		if len(m.matchLines) > 0 || m.search.Value() != "" {
			// Clear an active search first.
			m.search.SetValue("")
			m.matchLines = nil
			m.viewport.SetContent(m.renderDetail())
			return m, nil
		}
		return m.openList()
	case "/":
		m.searching = true
		return m, m.search.Focus()
	case "n":
		m.jumpMatch(1)
		return m, nil
	case "N":
		m.jumpMatch(-1)
		return m, nil
	case "e":
		return m.openEditor(m.detail, screenList)
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// detailLines returns the note as plain text lines (title + body), the basis
// for both rendering and search.
func (m model) detailLines() []string {
	raw := "# " + m.detail.title + "\n\n" + m.detail.body
	return strings.Split(raw, "\n")
}

// renderDetail returns the markdown-styled note for normal reading.
func (m model) renderDetail() string {
	return renderMarkdown("# "+m.detail.title+"\n\n"+m.detail.body, m.viewport.Width())
}

// applySearch recomputes matches for the current query and re-renders the
// viewport as plain text with every match highlighted.
func (m *model) applySearch() {
	q := strings.TrimSpace(m.search.Value())
	m.matchLines = nil
	m.matchIdx = 0
	if q == "" {
		m.viewport.SetContent(m.renderDetail())
		return
	}

	lq := strings.ToLower(q)
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(q))
	lines := m.detailLines()
	out := make([]string, len(lines))
	for i, ln := range lines {
		if strings.Contains(strings.ToLower(ln), lq) {
			m.matchLines = append(m.matchLines, i)
			out[i] = re.ReplaceAllStringFunc(ln, func(s string) string {
				return matchStyle.Render(s)
			})
		} else {
			out[i] = ln
		}
	}
	m.viewport.SetContent(strings.Join(out, "\n"))
	if len(m.matchLines) > 0 {
		m.scrollToMatch()
	}
}

// jumpMatch moves to the next (dir=1) or previous (dir=-1) match and scrolls
// to it, emphasising the focused match.
func (m *model) jumpMatch(dir int) {
	if len(m.matchLines) == 0 {
		return
	}
	m.matchIdx = (m.matchIdx + dir + len(m.matchLines)) % len(m.matchLines)
	m.rerenderWithCurrent()
	m.scrollToMatch()
}

// rerenderWithCurrent re-highlights matches, emphasising the focused one.
func (m *model) rerenderWithCurrent() {
	q := strings.TrimSpace(m.search.Value())
	if q == "" {
		return
	}
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(q))
	lq := strings.ToLower(q)
	lines := m.detailLines()
	current := -1
	if m.matchIdx < len(m.matchLines) {
		current = m.matchLines[m.matchIdx]
	}
	out := make([]string, len(lines))
	for i, ln := range lines {
		if !strings.Contains(strings.ToLower(ln), lq) {
			out[i] = ln
			continue
		}
		style := matchStyle
		if i == current {
			style = currentMatchStyle
		}
		out[i] = re.ReplaceAllStringFunc(ln, func(s string) string {
			return style.Render(s)
		})
	}
	m.viewport.SetContent(strings.Join(out, "\n"))
}

// scrollToMatch positions the viewport so the focused match line is visible.
func (m *model) scrollToMatch() {
	if m.matchIdx >= len(m.matchLines) {
		return
	}
	target := m.matchLines[m.matchIdx] - m.viewport.Height()/2
	if target < 0 {
		target = 0
	}
	m.viewport.SetYOffset(target)
}

func (m *model) toggleFocus() {
	if m.focus == focusTitle {
		m.focus = focusBody
		m.title.Blur()
	} else {
		m.focus = focusTitle
		m.body.Blur()
	}
}

func (m *model) focusCmd() tea.Cmd {
	if m.focus == focusTitle {
		return m.title.Focus()
	}
	return m.body.Focus()
}

// markDirty flags pending changes and returns a command that triggers a
// debounced autosave. Each call bumps editSeq so only the most recent tick
// actually writes to disk.
func (m *model) markDirty() tea.Cmd {
	m.dirty = true
	m.status = ""
	m.editSeq++
	seq := m.editSeq
	return tea.Tick(autosaveDelay, func(time.Time) tea.Msg {
		return autosaveMsg{seq: seq}
	})
}

// layout sizes the input fields to the current terminal dimensions.
func (m *model) layout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	inner := m.width - 8 // border + padding + outer margins
	if inner < 20 {
		inner = 20
	}
	m.title.SetWidth(inner)

	bodyHeight := m.height - 11 // title row, labels, divider, status, help, borders
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	m.body.SetWidth(inner)
	m.body.SetHeight(bodyHeight)

	// List fills most of the screen.
	m.list.SetSize(m.width-4, m.height-2)

	// Detail viewport leaves room for the header, search line, and help.
	vpW := m.width - 6
	if vpW < 20 {
		vpW = 20
	}
	vpH := m.height - 6
	if vpH < 3 {
		vpH = 3
	}
	m.viewport.SetWidth(vpW)
	m.viewport.SetHeight(vpH)
	m.search.SetWidth(vpW - 4)
}

// --- Persistence -----------------------------------------------------------

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "untitled"
	}
	return s
}

func notesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".th-note"
	}
	return filepath.Join(home, ".th-note")
}

// save autosaves the note to disk as markdown. The title becomes the
// filename; renaming the title renames the file on disk.
func (m *model) save() {
	title := strings.TrimSpace(m.title.Value())
	body := m.body.Value()
	if title == "" && strings.TrimSpace(body) == "" {
		return // nothing worth saving yet
	}

	dir := notesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		m.status = "save failed: " + err.Error()
		return
	}

	target := filepath.Join(dir, slugify(title)+".md")
	// If the title changed, move the old file rather than orphan it.
	if m.savePath != "" && m.savePath != target {
		_ = os.Rename(m.savePath, target)
	}

	heading := title
	if heading == "" {
		heading = "Untitled note"
	}
	content := "# " + heading + "\n\n" + body + "\n"
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		m.status = "save failed: " + err.Error()
		return
	}

	m.savePath = target
	m.savedAt = time.Now().Format("15:04:05")
	m.dirty = false
}

// --- Views -----------------------------------------------------------------

func (m model) welcome() string {
	logoLines := make([]string, len(logo))
	for i, line := range logo {
		c := logoColors[i%len(logoColors)]
		logoLines[i] = lipgloss.NewStyle().Foreground(c).Bold(true).Render(line)
	}
	art := strings.Join(logoLines, "\n")

	tagline := taglineStyle.Render("✨ your notes, right in the terminal")

	hint := func(key, desc string) string {
		return keyStyle.Render(key) + " " + keyDescStyle.Render(desc)
	}
	help := lipgloss.JoinHorizontal(
		lipgloss.Center,
		hint("n", "new note"),
		"   ",
		hint("l", "notes"),
		"   ",
		hint("q", "quit"),
	)

	content := lipgloss.JoinVertical(lipgloss.Center, art, "", tagline, "", help)
	return boxStyle.Render(content)
}

func (m model) statusBar() string {
	body := m.body.Value()
	words := len(strings.Fields(body))
	chars := utf8.RuneCountInString(body)
	line := m.body.Line() + 1
	col := m.body.LineInfo().ColumnOffset + 1

	stats := statusBarStyle.Render(
		fmt.Sprintf("%d words · %d chars · ln %d, col %d", words, chars, line, col),
	)

	var state string
	switch {
	case m.status != "":
		state = dirtyStyle.Render(m.status)
	case m.dirty:
		state = dirtyStyle.Render("● unsaved")
	case m.savedAt != "":
		rel := m.savePath
		if home, err := os.UserHomeDir(); err == nil {
			if r, err := filepath.Rel(home, m.savePath); err == nil {
				rel = "~/" + r
			}
		}
		state = savedStyle.Render("✓ saved " + m.savedAt + " · " + rel)
	default:
		state = statusBarStyle.Render("ready")
	}

	gap := m.width - 6 - lipgloss.Width(stats) - lipgloss.Width(state)
	if gap < 1 {
		gap = 1
	}
	return stats + strings.Repeat(" ", gap) + state
}

func (m model) editor() string {
	inner := m.width - 6
	if inner < 20 {
		inner = 20
	}

	titleLabel := fieldLabelStyle.Render("TITLE")
	bodyLabel := fieldLabelStyle.Render("NOTE")
	if m.focus == focusTitle {
		titleLabel = fieldLabelActiveStyle.Render("● TITLE")
	} else {
		bodyLabel = fieldLabelActiveStyle.Render("● NOTE")
	}

	divider := dividerStyle.Render(strings.Repeat("─", inner))

	// Body area: markdown preview or the editable textarea.
	var bodyView string
	if m.preview {
		bodyView = m.previewView(inner)
	} else {
		bodyView = m.body.View()
	}

	help := m.helpLine()

	parts := []string{
		titleLabel,
		m.title.View(),
		divider,
		bodyLabel,
		bodyView,
		divider,
		m.statusBar(),
		help,
	}
	body := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return editorBoxStyle.Render(body)
}

func (m model) previewView(width int) string {
	src := "# " + strings.TrimSpace(m.title.Value()) + "\n\n" + m.body.Value()
	rendered := renderMarkdown(src, width)

	// Constrain height so the preview stays inside the box.
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	maxH := m.body.Height()
	if len(lines) > maxH {
		lines = lines[:maxH]
	}
	// Pad to keep the box a stable height.
	for len(lines) < maxH {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (m model) helpLine() string {
	k := helpKeyStyle.Render
	d := helpStyle.Render
	mode := "edit"
	if m.preview {
		mode = "preview"
	}
	return helpStyle.Render(
		k("tab") + d(" switch · ") +
			k("ctrl+s") + d(" save · ") +
			k("ctrl+p") + d(" "+mode+" · ") +
			k("ctrl+t") + d(" time · ") +
			k("ctrl+b") + d(" bold · ") +
			k("esc") + d(" back"),
	)
}

func (m model) listView() string {
	body := m.list.View()

	var footer string
	if m.confirmDelete {
		footer = " " + confirmStyle.Render(
			"Delete \""+m.deleteTarget.title+"\"?  "+
				confirmKeyStyle.Render("y")+" yes  "+
				confirmKeyStyle.Render("n")+" no",
		)
	} else {
		footer = " " + helpStyle.Render(
			helpKeyStyle.Render("enter")+" open · "+
				helpKeyStyle.Render("/")+" search · "+
				helpKeyStyle.Render("e")+" edit · "+
				helpKeyStyle.Render("n")+" new · "+
				helpKeyStyle.Render("d")+" delete · "+
				helpKeyStyle.Render("esc")+" back",
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, "", footer)
}

func (m model) detailView() string {
	header := listTitleStyle.Render("📄 " + m.detail.title)

	// Search status / input line.
	var bar string
	switch {
	case m.searching:
		bar = m.search.View()
	case m.search.Value() != "":
		n := len(m.matchLines)
		pos := 0
		if n > 0 {
			pos = m.matchIdx + 1
		}
		bar = helpStyle.Render(fmt.Sprintf("/%s  — %d/%d matches", m.search.Value(), pos, n))
	default:
		bar = helpStyle.Render(strings.Repeat(" ", 1))
	}

	help := helpStyle.Render(
		helpKeyStyle.Render("/") + " search · " +
			helpKeyStyle.Render("n/N") + " next/prev · " +
			helpKeyStyle.Render("↑/↓") + " scroll · " +
			helpKeyStyle.Render("e") + " edit · " +
			helpKeyStyle.Render("esc") + " back",
	)

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		bar,
		m.viewport.View(),
		help,
	)
	return editorBoxStyle.Render(inner)
}

func (m model) View() tea.View {
	var content string
	switch m.screen {
	case screenEditor:
		content = m.centered(m.editor())
	case screenList:
		content = m.listView()
	case screenDetail:
		content = m.centered(m.detailView())
	default:
		content = m.centered(m.welcome())
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// centered places a box in the middle of the screen when dimensions are known.
func (m model) centered(box string) string {
	if m.width <= 0 || m.height <= 0 {
		return box
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
