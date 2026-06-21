package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// note is a single saved note loaded from disk.
type note struct {
	path  string
	title string
	body  string
	mod   time.Time
}

// parseNote splits the on-disk markdown ("# Title\n\n<body>") into its title
// and body. Files that don't start with a heading are treated as bodies with
// an "Untitled" title.
func parseNote(s string) (title, body string) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
		title = strings.TrimPrefix(lines[0], "# ")
		rest := lines[1:]
		for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
			rest = rest[1:]
		}
		body = strings.Join(rest, "\n")
	} else {
		title = "Untitled"
		body = s
	}
	return strings.TrimSpace(title), strings.TrimRight(body, "\n")
}

// loadNotes reads every markdown note from the notes directory, newest first.
func loadNotes() ([]note, error) {
	dir := notesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var notes []note
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		title, body := parseNote(string(data))
		notes = append(notes, note{path: p, title: title, body: body, mod: info.ModTime()})
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].mod.After(notes[j].mod)
	})
	return notes, nil
}

// noteItem adapts a note to the bubbles list.DefaultItem interface so the list
// can render and filter it. FilterValue spans title + body so search matches
// note contents, not just titles.
type noteItem struct{ n note }

func (i noteItem) Title() string { return i.n.title }

func (i noteItem) Description() string {
	snippet := strings.Join(strings.Fields(i.n.body), " ")
	if snippet == "" {
		return "(empty) · " + i.n.mod.Format("2006-01-02 15:04")
	}
	const max = 56
	if len([]rune(snippet)) > max {
		snippet = string([]rune(snippet)[:max]) + "…"
	}
	return snippet
}

func (i noteItem) FilterValue() string { return i.n.title + " " + i.n.body }
