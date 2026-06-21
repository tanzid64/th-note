# th-note ✨

A fast, colorful note-taking app that lives in your terminal — built with [Go](https://go.dev) and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Write notes in Markdown, browse them in a searchable paginated list, preview the rendered Markdown, and never leave the keyboard. Notes are plain `.md` files on disk, so they're yours forever.

```
╭──────────────────────────────────────────────────────────────────╮
│                                                                    │
│     ████████╗██╗  ██╗      ███╗   ██╗ ██████╗ ████████╗███████╗    │
│     ╚══██╔══╝██║  ██║      ████╗  ██║██╔═══██╗╚══██╔══╝██╔════╝    │
│        ██║   ███████║█████╗██╔██╗ ██║██║   ██║   ██║   █████╗      │
│        ██║   ██╔══██║╚════╝██║╚██╗██║██║   ██║   ██║   ██╔══╝      │
│        ██║   ██║  ██║      ██║ ╚████║╚██████╔╝   ██║   ███████╗    │
│        ╚═╝   ╚═╝  ╚═╝      ╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚══════╝    │
│                                                                    │
│                ✨ your notes, right in the terminal                │
│                                                                    │
│                        n  new note    l  notes    q  quit          │
│                                                                    │
╰──────────────────────────────────────────────────────────────────╯
```

## Features

- 📝 **Smart editor** — separate **Title** and **Note** fields, `Tab` to switch between them.
- 💾 **Autosave** — changes are written to disk automatically (debounced), so you never lose work. The note title becomes the filename.
- 👀 **Live Markdown preview** — toggle a styled preview of headings, **bold**, *italic*, `code`, lists, and quotes.
- 📊 **Live stats** — word count, character count, and cursor line/column as you type.
- 📚 **Paginated notes list** — browse all your notes, newest first.
- 🔍 **Search everywhere** — filter the list by title *and* content, and search for words *inside* a note with match highlighting and jump-to-next.
- 🗑️ **Delete with confirmation** — remove notes safely from the list.
- ⚡ **Shortcut commands** — insert timestamps, bold markers, save, and more.
- 🎨 **Modern look** — a Catppuccin-inspired color theme and rounded borders.

## Where your notes live

Notes are saved as Markdown files in a `.th-note` folder in your home directory:

| OS      | Location                          |
| ------- | --------------------------------- |
| Linux   | `~/.th-note/`                     |
| macOS   | `~/.th-note/`                     |
| Windows | `C:\Users\<you>\.th-note\`        |

Each note is a normal `.md` file (`# Title` followed by the body), so you can read, back up, sync, or edit them with any other tool.

## Requirements

- A terminal that supports **true color** and **Unicode** for the best experience.
  - **Linux/macOS:** most modern terminals (GNOME Terminal, Kitty, Alacritty, iTerm2, the macOS Terminal) work out of the box.
  - **Windows:** use **[Windows Terminal](https://aka.ms/terminal)** — the legacy `cmd.exe`/PowerShell console may not render the logo and emoji correctly.
- To build from source or use `go install`, you need **[Go 1.26+](https://go.dev/dl/)**.

---

## Installation

th-note runs on **Linux, macOS, and Windows**. The easiest options need **no Go and no build step** — just download a pre-built binary from [**Releases**](https://github.com/tanzid64/th-note/releases).

### Option 1 — One-line install (Linux & macOS) ⭐

```sh
curl -fsSL https://raw.githubusercontent.com/tanzid64/th-note/main/install.sh | sh
```

This detects your OS/architecture, downloads the latest release binary, and installs it onto your `PATH`. Then run `th-note`.

### Option 2 — Download a binary manually (all operating systems)

Grab the right archive from the [latest release](https://github.com/tanzid64/th-note/releases/latest):

| OS                    | Download                        | Then…                                                        |
| --------------------- | ------------------------------- | ------------------------------------------------------------ |
| **Linux** (x86-64)    | `th-note_linux_amd64.tar.gz`    | `tar -xzf` it, then `sudo mv th-note /usr/local/bin/`        |
| **Linux** (ARM64)     | `th-note_linux_arm64.tar.gz`    | same as above                                                |
| **macOS** (Intel)     | `th-note_darwin_amd64.tar.gz`   | `tar -xzf` it, then `sudo mv th-note /usr/local/bin/`        |
| **macOS** (Apple Silicon) | `th-note_darwin_arm64.tar.gz` | same as above                                              |
| **Windows** (x86-64)  | `th-note_windows_amd64.zip`     | unzip, then run `th-note.exe` (place it in a folder on your `PATH`) |

You can also download directly with `curl`, e.g. for macOS Apple Silicon:

```sh
curl -fsSL -o th-note.tar.gz \
  https://github.com/tanzid64/th-note/releases/latest/download/th-note_darwin_arm64.tar.gz
tar -xzf th-note.tar.gz
sudo mv th-note /usr/local/bin/
th-note
```

Every release also ships a `checksums.txt` so you can verify your download with `sha256sum -c`.

> **macOS Gatekeeper:** the binary isn't notarized, so the first launch may be blocked. Allow it with `xattr -d com.apple.quarantine /usr/local/bin/th-note`, or via *System Settings → Privacy & Security → Open Anyway*.

> **Windows:** [Windows Terminal](https://aka.ms/terminal) is strongly recommended so the colors, logo, and emoji render correctly.

### Option 3 — Install with Go (all operating systems)

If you have [Go 1.26+](https://go.dev/dl/) installed, this is the quickest way:

```bash
go install github.com/tanzid64/th-note@latest
```

This builds and installs the `th-note` binary into your Go bin directory. Make sure it's on your `PATH`:

- **Linux / macOS** — add to `~/.bashrc`, `~/.zshrc`, etc.:
  ```bash
  export PATH="$PATH:$(go env GOPATH)/bin"
  ```
- **Windows (PowerShell)** — the directory is usually `%USERPROFILE%\go\bin`. Add it via *Settings → Environment Variables*, or:
  ```powershell
  $env:Path += ";$(go env GOPATH)\bin"
  ```

Then just run:

```bash
th-note
```

### Option 4 — Build from source

```bash
# 1. Get the code
git clone https://github.com/tanzid64/th-note.git
cd th-note

# 2a. Build with the Makefile (Linux/macOS)
make build      # produces ./th-note
make run        # build and run

# 2b. …or build directly with Go (any OS)
go build -o th-note .       # Linux/macOS  → ./th-note
go build -o th-note.exe .   # Windows      → .\th-note.exe
```

Run it:

```bash
./th-note          # Linux/macOS
.\th-note.exe      # Windows (PowerShell)
```

### Option 5 — Build a standalone binary yourself

Go can cross-compile a single self-contained executable for any platform — no runtime or dependencies required on the target machine. From the project folder:

```bash
# Linux (x86-64)
GOOS=linux   GOARCH=amd64 go build -o th-note-linux .

# macOS (Intel)
GOOS=darwin  GOARCH=amd64 go build -o th-note-macos-intel .

# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o th-note-macos-arm .

# Windows (x86-64)
GOOS=windows GOARCH=amd64 go build -o th-note.exe .
```

On **Windows PowerShell**, set the variables like this instead:

```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o th-note.exe .
```

Copy the resulting binary anywhere on the target machine and run it directly.

> **Note:** Pre-built binaries for every tagged release are published automatically to [Releases](https://github.com/tanzid64/th-note/releases) by GitHub Actions (see `.github/workflows/release.yml`). Options 1 and 2 use those, so most people won't need to build anything.

---

## Usage

Launch the app and you'll land on the welcome screen. Everything is keyboard-driven.

### Welcome screen

| Key | Action          |
| --- | --------------- |
| `n` | New note        |
| `l` | Browse notes    |
| `a` | About           |
| `q` | Quit            |

### Editor (`n` / `e`)

| Key            | Action                                  |
| -------------- | --------------------------------------- |
| `Tab`          | Switch between Title and Note fields    |
| `ctrl+s`       | Save now                                |
| `ctrl+p`       | Toggle Markdown preview                 |
| `ctrl+t`       | Insert current timestamp                |
| `ctrl+b`       | Insert **bold** markers                 |
| `esc`          | Save and go back                        |
| `ctrl+c`       | Save and quit                           |

Changes also **autosave** a moment after you stop typing.

### Notes list (`l`)

| Key             | Action                                   |
| --------------- | ---------------------------------------- |
| `↑`/`↓` or `j`/`k` | Move selection                        |
| `←`/`→`         | Change page                              |
| `enter`         | Open the selected note                   |
| `/`             | Filter notes by title and content        |
| `e`             | Edit the selected note                   |
| `n`             | New note                                 |
| `d`             | Delete the selected note (asks `y`/`n`)  |
| `esc` / `q`     | Back to welcome                          |

### Note detail (read view)

| Key       | Action                                       |
| --------- | -------------------------------------------- |
| `/`       | Search for a word in this note               |
| `n` / `N` | Jump to next / previous match                |
| `↑`/`↓`   | Scroll                                       |
| `e`       | Edit this note                               |
| `esc`     | Clear the search, then go back to the list   |

---

## Tech stack

- **[Go](https://go.dev)** 1.26+
- **[Bubble Tea v2](https://github.com/charmbracelet/bubbletea)** — the TUI framework
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — list, viewport, textarea, and text input components
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — styling and layout

## Development

```bash
git clone https://github.com/tanzid64/th-note.git
cd th-note
go mod download   # fetch dependencies
go run .          # run without building a binary
go build ./...    # compile
go vet ./...      # static checks
```

The codebase is small:

- `main.go` — UI, screens (welcome, editor, list, detail, about), and persistence.
- `notes.go` — loading notes from disk and adapting them to the list.

## Author

**Md Tanzid Haque** — Software Developer (Laravel · Go · Next.js · real-time systems), Dhaka, Bangladesh.

- 🌐 Website — [tanzid.dev](https://tanzid.dev)
- 🐙 GitHub — [@tanzid64](https://github.com/tanzid64)
- 💼 LinkedIn — [in/tanzid64](https://linkedin.com/in/tanzid64)
- ✍️ Medium — [@tanzid64](https://medium.com/@tanzid64)

You can also see this inside the app on the **About** screen (press `a`).

## License

No license has been chosen yet. Until one is added, all rights are reserved by the author.

---

Made with ☕ and [Charm](https://charm.sh).
