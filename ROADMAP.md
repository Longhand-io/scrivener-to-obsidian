# scriv2obsidian roadmap

> Hand-written public roadmap. `docs/PHASES.md` is the engineering ledger and the source of truth for delivery state. Edit by hand.

scriv2obsidian turns a Scrivener 3 project into a vault that follows the [Longhand spec](https://github.com/grassclaw/longhand/blob/main/docs/spec.md), and replays every Scrivener snapshot as history. It is a single Go binary with no dependencies beyond the standard library, and it never modifies the source package.

## Shipped

- **RTF reader.** Tokenizer, parser, and Markdown renderer for Scrivener's Cocoa RTF: paragraphs, line breaks, bold, italic, underline, strike, super and subscript, lists, tabs, hyperlink fields including `scrivcmt://` comment anchors and `scrivlnk://` internal links, cp1252 and Unicode escapes, inline images. Built because pandoc's RTF reader drops hyperlink fields, text included. Not yet covered by tests.

## Next: v0.1.0

- **Binder parsing.** The `.scrivx` XML: hierarchy, titles, types, labels, statuses, keywords, custom metadata, include-in-compile, created and modified dates, bookmarks.
- **Layout and frontmatter.** Numbered folders and files, folder notes, `_Project.md`, every spec field, trash skipped unless asked.
- **Comments, links, images.** Footnotes to `[^n]`, comments to `%% %%`, internal links to wikilinks, web links kept, inline images to `_attachments/`.
- **Snapshots.** Scrivener's `Snapshots/` replayed in date order into the files store, and with `-git` into commits with the spec's message shape and original dates.
- **Research.** PDFs and images copied with their binder position; sidecar notes when they carry synopses or notes.
- **CLI.** `convert` with the documented flags, `inspect` for a dry inventory, a manifest, exit codes.
- **Tests.** A synthetic `.scriv` fixture with every feature above; golden files for the output; the RTF reader covered.
- **Validation.** Run against real projects; document counts, word counts within 1%, and snapshot counts reconciled against the package.
- **Release.** Binaries for macOS arm64 and amd64, Linux, and Windows from CI, with checksums.

## Future

- **Poetry heuristics.** Detect verse per document instead of the `-poetry` flag for the whole run.
- **Scrivener styles.** Map named paragraph styles to headings where a project used them consistently.
- **Tables.** Scrivener tables to Markdown tables.
- **Windows Scrivener packages.** Same format, different folder casing; verify rather than assume.
- **Re-import.** Import a project again into an existing vault and only add what is new, keyed on `id`.
- **Other sources.** Ulysses and Storyist export the same shape of thing; the layout writer is source-agnostic.

### Non-goals

- Round-tripping back to Scrivener.
- Scrivener 1 and 2 packages. Open and save them in Scrivener 3 first.
- Any dependency outside the Go standard library, and any network access.
