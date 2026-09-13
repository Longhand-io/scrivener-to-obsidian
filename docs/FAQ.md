# FAQ

Things writers notice after importing a project, most of them not bugs. If yours is not here, ask in [Discussions](https://github.com/Longhand-io/longhand/discussions).

<!-- This file is the source of the site page https://longhand-io.github.io/longhand-site/faq.html. Render with: go run ./hack/faqgen -md docs/FAQ.md -html ../longhand-site/faq.html. CI fails if the site drifts. -->

## It looks like ordinary Obsidian. Where is the binder, the inspector, the snapshots tab?

Those are the Longhand plugin, which is being built. The importer's only job is to move your writing into plain files with nothing lost. What you see now is the foundation: your manuscript as numbered Markdown files, your metadata in Properties, your research copied, your snapshots as files and as git history. The plugin reads exactly this layout, so nothing needs converting again when it arrives.

## A PDF shows "0 backlinks" with a red icon. Is something broken?

<!-- figure: backlinks -->

No. That chip is Obsidian's own counter: no note links to this file yet. The red icon is the toggle for showing backlinks inside the document, red because there are none to show. In Scrivener, a research file is only connected to a document if you bookmarked it or made an internal link to it, and most research items were never anchored to anything. Where a PDF or image had a synopsis or notes in Scrivener, the importer writes a note of the same name beside it that embeds the file, and that note counts as one backlink. Scrivener bookmarks come across as wikilinks in a note's Properties, which Obsidian also counts. Link a research file from a scene and the number changes.

## Some images did not come over.

If any image is missing, that is a bug; please report it. One was found and fixed on 2026-09-12: inline PNGs came out unreadable because an identifier Scrivener stores next to each picture was being included in the image bytes. Reimport with a build after that date. Inline pictures in document text land in `_attachments/` and are embedded where they were; image items in Research are copied as files.

## Why are files numbered? Can I turn that off?

Obsidian sorts the file tree by name, and your manuscript has an order that is not alphabetical. The two-digit prefix keeps binder order visible everywhere, including on a phone with no plugin. If you would rather not see numbers, run with `-no-prefix` and the order is written to an `order` field in each file's Properties instead. The Longhand plugin honours either.

## There is a note inside a folder with the folder's own name. What is it?

Scrivener lets a folder carry text of its own, typically chapter text above its scenes. Obsidian folders cannot hold text, so that text becomes a note with the folder's name inside the folder, which Obsidian and most plugins treat as the folder's note.

## Where did my Trash go?

Not exported, on purpose, and neither are snapshots of trashed documents. `inspect` shows both counts in their own columns so you know they exist. Run `convert` with `-include-trash` to bring them along under `Trash/`.

## Why does the snapshot count differ from what Scrivener shows?

Two reasons, both reported by `inspect`. Snapshots of trashed documents are skipped with the trash. And Scrivener occasionally saves two files for one snapshot, one for each time zone the Mac was in when it synced; when the bytes are identical the importer keeps one, and when they differ it keeps both.

## My poems lost their line breaks.

Scrivener stores each line of verse as its own paragraph, and prose mode turns paragraphs into paragraphs. Convert verse projects with `-poetry`: single breaks become line breaks and blank lines become stanza breaks. The flag applies to the whole run, so convert poetry projects in their own command.

## Bold and italic survived but my fonts, colours, and styles did not.

By design. Markdown has emphasis and headings, not fonts and colours, and Scrivener's named styles are a compile-time concern. The importer keeps bold, italic, underline, strikethrough, super and subscript, lists, links, footnotes, comments, and images. Tables are not converted yet.

## Where are the snapshots?

Two places, the same content. `_snapshots/<document id>/` inside each project holds one Markdown file per snapshot, named by date and the title you gave it, and syncs with the vault to every device. If you converted with `-git`, each snapshot is also a commit with its original date, which the Longhand Snapshots module will show as a list with compare and restore. Until then, any git client shows them.

## Can I put the vault in iCloud or Dropbox?

The files, yes. The git repository, not directly: a `.git` folder inside a cloud-synced folder gets corrupted. If you used `-git`, either keep the vault outside cloud sync or move the repository out with `git init --separate-git-dir`; the Longhand plugin does this for you. See [Sync and history](https://github.com/Longhand-io/longhand/blob/main/docs/sync-and-history.md).

## Can I run the importer again?

Yes, into an empty folder. Reimporting into an existing vault and merging only what is new is on the roadmap, keyed on each document's `id`. The source package is never modified, so there is no risk in running it as often as you like.
