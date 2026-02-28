---
name: peek
description: Use when the user wants to review, read, or annotate a markdown plan file without leaving the terminal. Triggers on "let me review", "open the plan", "/peek", "let me peek".
---

# Review Markdown Files

Open a markdown file in peek for voice-driven review and annotation.

## Usage

- `/peek` - opens the most recently modified `.md` in `docs/plans/`
- `/peek dark-factory` - fuzzy-matches a file in `docs/plans/` by keyword
- `/peek path/to/file.md` - opens a specific file

## Instructions

When invoked, resolve the target file using the args parameter:

1. **No args**: Run `peek` (opens most recent plan)
2. **Looks like a path** (contains `/` or ends in `.md`): Run `peek <path>`
3. **Keyword**: Run `peek --find <keyword>`

Run the resolved command using the Bash tool.

After the user exits peek, the stdout will contain an annotation summary. Read it and offer to act on the feedback (fix issues, update the plan, etc.).
