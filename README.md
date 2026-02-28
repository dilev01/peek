# peek

Voice-driven markdown review tool for the terminal.

Read, navigate, and annotate markdown files with your voice, right inside your terminal.

## Install

**Prerequisites:** [PortAudio](http://www.portaudio.com/) for microphone capture.

```bash
# macOS
brew install portaudio

# Build and install
git clone https://github.com/dilev01/peek.git
cd peek
make install
```

Or build from source:

```bash
go install github.com/dilev01/peek/cmd/peek@latest
```

## Usage

```bash
peek file.md                    # Open a specific file
peek                            # Open most recent .md in docs/plans/
peek --find dark-factory        # Fuzzy match by keyword
peek --version                  # Print version
```

## Keybindings

| Key | Action |
|-----|--------|
| `v` | Toggle voice recording |
| `a` | Text annotation on current line |
| `Enter` | Show annotation on current line |
| `j` / `k` | Scroll down / up |
| `d` / `u` | Half-page down / up |
| `g` / `G` | Jump to top / bottom |
| `/` | Search (type query, Enter to find) |
| `n` / `N` | Next / previous search match |
| `[` / `]` | Previous / next heading |
| `{` / `}` | Previous / next annotation |
| `?` | Toggle help overlay |
| `q` | Quit and output annotations |

## Voice Commands

Press `v` to start recording, speak a command, press `v` again to stop. Navigation commands are executed immediately:

- "go to line 50" / "line 50"
- "next" / "next page"
- "back" / "previous page"
- "top" / "bottom"
- "find error handling"
- "next heading" / "previous heading"

Anything that doesn't match a navigation pattern is saved as a voice annotation pinned to the current line.

## Annotations

Annotations are saved in two formats:

**Sidecar file** (`<filename>.peek.json`): JSON with all annotations, line numbers, timestamps, and audio file paths.

**Stdout summary** (on quit): Structured text printed to stdout, designed for Claude Code integration.

```
── peek review: docs/plans/my-plan.md ──
   Duration: 5m12s | Annotations: 2 (1 voice, 1 text)

   L9  [voice] this needs error handling for the timeout case
   L12 [text]  missing the fourth component

   Full review: docs/plans/my-plan.peek.json
────────────────────────────────────────────
```

Audio clips are saved as individual `.wav` files in a `.peek/` directory next to the reviewed file.

## Configuration

Set `OPENAI_API_KEY` environment variable, or create `~/.config/peek/config.yaml`:

```yaml
api_key: sk-...
language: en
```

Without an API key, peek starts in keyboard-only mode (no voice features).

## Claude Code Plugin

peek ships as a Claude Code plugin. Install the plugin from the `plugin/` directory for `/peek` command integration in Claude Code sessions.

## Tech Stack

- [Go](https://go.dev/) + [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) (TUI framework)
- [Glamour](https://github.com/charmbracelet/glamour) (markdown rendering)
- [Lip Gloss v2](https://github.com/charmbracelet/lipgloss) (terminal styling)
- [PortAudio](https://github.com/gordonklaus/portaudio) (mic capture)
- [OpenAI Whisper API](https://platform.openai.com/docs/guides/speech-to-text) (transcription)

## License

MIT
