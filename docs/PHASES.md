---
repo: scrivener-to-obsidian
schema: phases/v1
program: spec-and-importer
current_phase: I8
updated: 2026-09-13
updated_by: grassclaw

phases:
  - id: I0
    title: RTF reader
    status: done
    completed: 2026-09-09
    depends_on: []
    subphases:
      - id: I0.1
        title: Tokenizer, parser, Markdown renderer
        status: done
        deliverables:
          - { id: I0.1-d1, done: true, desc: "internal/rtf: token.go, parse.go, render.go; hyperlink fields preserved; poetry mode; plain text and word count" }
        acceptance:
          - { id: I0.1-a1, met: true, check: "unit tests cover bold, italic, lists, links, comment anchors, unicode, images, poetry joins", method: unit, note: "2026-09-12: internal/rtf/rtf_test.go; the tests found two real bugs on landing (blipuid hex leaking into image bytes, \\u escapes emitting the fallback instead of the character), both fixed in the same change" }

  - id: I1
    title: Binder parsing
    status: done
    completed: 2026-09-12
    depends_on: []
    subphases:
      - id: I1.1
        title: scrivx to a project tree
        status: done
        deliverables:
          - { id: I1.1-d1, done: true, desc: "internal/scrivx: Project and Item; labels, statuses, keywords, custom metadata definitions; per-item label, status, keywords, custom values, include, dates, bookmarks; trash detection" }
        acceptance:
          - { id: I1.1-a1, met: true, check: "fixture scrivx parses to the expected tree; a real project's item count equals grep -c BinderItem", method: unit, note: "2026-09-12: TestParse green; item counts matched grep on 14 real Scrivener 3 projects" }

  - id: I2
    title: Layout and frontmatter
    status: done
    completed: 2026-09-12
    depends_on: [I1]
    subphases:
      - id: I2.1
        title: Paths, ordering, folder notes, project note
        status: done
        deliverables:
          - { id: I2.1-d1, done: true, desc: "sanitised titles, numeric prefixes with width from sibling count, collision suffixes, folder notes, _Project.md, -no-prefix writes order:" }
        acceptance:
          - { id: I2.1-a1, met: true, check: "golden tree for the fixture matches byte for byte", method: unit }
      - id: I2.2
        title: Document frontmatter per spec
        status: done
        deliverables:
          - { id: I2.2-d1, done: true, desc: "id, title, type, created, modified, synopsis, label, status, tags, include, bookmarks, meta; notes callout; YAML strings quoted via JSON encoding" }
        acceptance:
          - { id: I2.2-a1, met: true, check: "every spec field appears in the fixture output where the fixture sets it, and nowhere it does not", method: unit }

  - id: I3
    title: Comments, links, images, research
    status: done
    completed: 2026-09-12
    depends_on: [I2]
    subphases:
      - id: I3.1
        title: content.comments to footnotes and comments
        status: done
        deliverables:
          - { id: I3.1-d1, done: true, desc: "footnote comments to [^n] with definitions at the end; other comments to %% %% after the anchor; comment bodies through the RTF reader as plain text" }
        acceptance:
          - { id: I3.1-a1, met: true, check: "fixture with two footnotes and one comment renders both forms", method: unit }
      - id: I3.2
        title: Links and images
        status: done
        deliverables:
          - { id: I3.2-d1, done: true, desc: "scrivlnk to [[Title]] via the path map, unresolved left as text; http links kept; inline images to _attachments/<id>-<n>.<ext>" }
        acceptance:
          - { id: I3.2-a1, met: true, check: "fixture link resolves to the renamed target; image file exists and is embedded", method: unit }
      - id: I3.3
        title: Research assets
        status: done
        deliverables:
          - { id: I3.3-d1, done: true, desc: "PDF, Image, WebArchive, Other items copied with position; sidecar note when synopsis or notes exist" }
        acceptance:
          - { id: I3.3-a1, met: true, check: "fixture PDF lands at its numbered path with a sidecar", method: unit }

  - id: I4
    title: Snapshots
    status: done
    completed: 2026-09-12
    depends_on: [I2]
    subphases:
      - id: I4.1
        title: Files store
        status: done
        deliverables:
          - { id: I4.1-d1, done: true, desc: "Snapshots/<uuid>.snapshots/index.xml and rtf files matched by instant across time zones; written to _snapshots/<id>/<utc-timestamp> <title>.md with the document's frontmatter; Untitled when the index has no title" }
        acceptance:
          - { id: I4.1-a1, met: true, check: "fixture with three snapshots, one untitled, one with a different zone offset, yields three files in date order", method: unit }
      - id: I4.2
        title: Git replay
        status: done
        deliverables:
          - { id: I4.2-d1, done: true, desc: "-git: init if needed; commits in project-wide date order, one file each, snap(path): title, Snapshot-* trailers, GIT_AUTHOR_DATE and GIT_COMMITTER_DATE set; final feat(<project>): import from Scrivener commit; -author flag" }
        acceptance:
          - { id: I4.2-a1, met: true, check: "git log --format=%aI shows original dates ascending; git log --grep Snapshot-Doc-Id count equals snapshot count", method: integration }

  - id: I5
    title: CLI and manifest
    status: done
    completed: 2026-09-12
    depends_on: [I3, I4]
    subphases:
      - id: I5.1
        title: convert, inspect, manifest, dry-run
        status: done
        deliverables:
          - { id: I5.1-d1, done: true, desc: "cmd/scriv2obsidian: convert flags per README; inspect prints the inventory table; .scriv2obsidian.json manifest with uuid to path, counts, warnings; -dry-run prints the plan; exit codes" }
        acceptance:
          - { id: I5.1-a1, met: true, check: "README flag table equals --help output", method: review }

  - id: I6
    title: Tests and fixture
    status: done
    completed: 2026-09-12
    depends_on: [I1]
    note: "Runs alongside I2 to I5; every phase lands with its tests. Listed separately for the fixture itself."
    subphases:
      - id: I6.1
        title: Synthetic .scriv fixture
        status: done
        deliverables:
          - { id: I6.1-d1, done: true, desc: "testdata/fixture.scriv built by a Go helper: nested folders, folder text, labels, status, keywords, custom meta, footnote, comment, internal link, web link, image, PDF, three snapshots, trash item; no real manuscript text" }
        acceptance:
          - { id: I6.1-a1, met: true, check: "hack/ci.sh green on ubuntu and macos", method: integration }

  - id: I7
    title: Validation on real projects
    status: done
    completed: 2026-09-12
    depends_on: [I5]
    subphases:
      - id: I7.1
        title: Reconcile counts
        status: done
        deliverables:
          - { id: I7.1-d1, done: true, desc: "run against a set of real Scrivener 3 projects; record document, word, snapshot, and asset counts against the package in a private log; fix every discrepancy or document it in README limitations" }
        acceptance:
          - { id: I7.1-a1, met: true, check: "document count exact; snapshot count exact; word count within 1% per project", method: e2e, note: "2026-09-12: 14 real Scrivener 3 projects; text documents, snapshots (after same-instant duplicate removal), and assets matched exactly; word counts within 0.16% of macOS textutil per project; git replay 22 snapshot commits + 14 import commits, per-document dates ascending" }

  - id: I8
    title: Release v0.1.0
    status: in_progress
    depends_on: [I6, I7]
    subphases:
      - id: I8.1
        title: Binaries and notes
        status: in_progress
        deliverables:
          - { id: I8.1-d1, done: true, desc: "release workflow builds darwin/arm64, darwin/amd64, linux/amd64, linux/arm64, windows/amd64 with sha256 sums; CHANGELOG section becomes the release notes", note: "2026-09-13: .github/workflows/release.yml on a v* tag; hack/build-release.sh, hack/release-notes.sh; ci.yml dry-runs the build and the verification on every PR" }
        acceptance:
          - { id: I8.1-a1, met: false, check: "go install and the downloaded binary produce identical output on the fixture", method: e2e, note: "hack/verify-release.sh BINARY vX.Y.Z is the check; passed locally 2026-09-13 against a release-style build with a throwaway version (13 files identical under SOURCE_DATE_EPOCH); met when run against the published v0.1.0 archive" }
      - id: I8.2
        title: Cut v0.1.0
        status: planned
        deliverables:
          - { id: I8.2-d1, done: false, desc: "CHANGELOG heading renamed to v0.1.0 with the date; README install lines point at the release; tag pushed; release published with five archives and the checksum file" }
        acceptance:
          - { id: I8.2-a1, met: false, check: "shasum -c on the downloaded archive passes; go install ...@v0.1.0 prints v0.1.0", method: e2e }
---

# Phase ledger

The YAML above is the ledger. A deliverable is `done` when merged on `main`; an acceptance is `met` when the named check ran and passed by the named method. `current_phase` is the lowest phase with an open deliverable. Every PR that closes a deliverable flips it here under the PR body's Ledger heading.
