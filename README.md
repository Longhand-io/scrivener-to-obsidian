# scrivener-to-obsidian

Move your writing out of Scrivener into an Obsidian vault without losing the things that made Scrivener worth using: the binder, synopses, notes, labels, footnotes, research files, and every snapshot you ever took.

Status: pre-release, under active development. The importer runs; the Obsidian plugin is in design.

## What it does

- Converts a Scrivener 3 project (`.scriv`) into a folder of Markdown files, ordered like the binder.
- Keeps synopses, document notes, labels, statuses, keywords, custom metadata, and bookmarks as frontmatter.
- Converts footnotes and comments, internal links to wikilinks, inline images to attachments.
- Copies research PDFs and images.
- **Replays Scrivener snapshots as git commits** with their original titles and dates, so your version history moves with you.
- Uses its own RTF reader. No pandoc, no third-party packages, one static binary.

## Why not the other converters

Existing tools export text and folders. None carry snapshots across, and most drop notes, footnotes, and links. This one is built around a small written spec (`docs/design/03-spec.md`) so the output is usable by any tool, with or without a plugin.

## Install

```
go install github.com/grassclaw/scrivener-to-obsidian/cmd/scriv2obsidian@latest
```

Requires Go 1.23 or later to build, and `git` on PATH if you want snapshot history.

## Use

```
scriv2obsidian inspect  ~/Dropbox/Apps/Scrivener/Novel.scriv
scriv2obsidian convert  -out ~/Writing -git ~/Dropbox/Apps/Scrivener/Novel.scriv
```

Flags for `convert`:

| Flag | Effect |
|---|---|
| `-out DIR` | Vault directory to write into (required). |
| `-git` | Initialise a repository if needed and replay snapshots as commits. |
| `-poetry` | Treat single paragraph breaks as line breaks. Use for verse. |
| `-no-prefix` | Write `order:` frontmatter instead of numeric file prefixes. |
| `-include-trash` | Export the Trash folder too. |
| `-author "Name <email>"` | Git identity for the commits. |
| `-dry-run` | Print what would be written and stop. |

Your `.scriv` package is never modified. Run it as many times as you like.

## Design

- [User stories](docs/design/01-user-stories.md): who this is for and how Obsidian changes for them.
- [Architecture](docs/design/02-architecture.md): importer, plugin core, modules, security posture.
- [Vault spec](docs/design/03-spec.md): the frontmatter, layout, and commit conventions everything agrees on.

## Limitations

- Scrivener named styles, fonts, colours, and tables are not carried over. Bold, italic, underline, strike, super and subscript, lists, links, and images are.
- Scrivener 1 and 2 projects are not supported. Open and save them in Scrivener 3 first.
- Snapshot replay needs `git`.

## License

MIT. See `LICENSE`.
