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

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
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

	title    textinput.Model
	body     textarea.Model
	focus    focusField
	preview  bool   // markdown preview toggled on?
	dirty    bool   // unsaved changes since last write
	savedAt  string // last save time, for the status bar
	savePath string // current file on disk
	status   string // transient message (e.g. errors)
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Untitled note"
	ti.Prompt = ""
	ti.CharLimit = 120

	ta := textarea.New()
	ta.Placeholder = "Start writing… markdown supported. Try ctrl+t for a timestamp."
	ta.ShowLineNumbers = true

	return model{
		screen: screenWelcome,
		title:  ti,
		body:   ta,
		focus:  focusTitle,
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
		if m.screen == screenEditor {
			return m.updateEditor(msg)
		}
		return m.updateWelcome(msg)
	}

	// Forward non-key messages (cursor blink, etc.) to the active field.
	if m.screen == screenEditor {
		return m.routeToField(msg)
	}
	return m, nil
}

func (m model) updateWelcome(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "n":
		m.screen = screenEditor
		m.focus = focusTitle
		m.layout()
		return m, m.title.Focus()
	}
	return m, nil
}

func (m model) updateEditor(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.save() // don't lose work inside the autosave debounce window
		return m, tea.Quit

	case "esc":
		m.save()
		m.screen = screenWelcome
		m.title.Blur()
		m.body.Blur()
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

func (m model) View() tea.View {
	var box string
	if m.screen == screenEditor {
		box = m.editor()
	} else {
		box = m.welcome()
	}

	content := box
	if m.width > 0 && m.height > 0 {
		content = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			box,
		)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
