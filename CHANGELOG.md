# Changelog

User-visible changes, newest first. Each released section is the text published as that version's GitHub release notes.

## Unreleased: v0.1.0

Not yet cut. Scope in `ROADMAP.md`.

### Added

- RTF reader for Scrivener's Cocoa RTF, with hyperlink fields preserved and a poetry mode.
- Binder parser: hierarchy, titles, types, labels, statuses, nested keywords, custom metadata, include-in-compile, dates, bookmarks, trash detection.
- Layout and frontmatter per the Longhand spec: numbered folders and files, folder notes, project note, every document field, notes callout, `-no-prefix`.
- Footnotes to `[^n]`, comments to `%% %%`, internal links to wikilinks with the original text, web links kept, inline images to `_attachments/`.
- Research PDFs, images, and other files copied with their binder position, with a sidecar note when they carry a synopsis or notes.
- Snapshots written to the `_snapshots/` files store, and with `-git` replayed as commits with original dates and `Snapshot-*` trailers, followed by one import commit per project.
- `convert`, `inspect`, `-dry-run`, and a `.scriv2obsidian.json` manifest per project.
- Validated on fourteen real projects; see the README.

### Fixed

- Inline PNG images were written with the picture's `blipuid` hex prepended, which made them unreadable. JPEGs were unaffected.
- A `\\u` Unicode escape emitted its fallback character instead of the character itself.
