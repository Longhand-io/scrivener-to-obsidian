# Contributing to scrivener-to-obsidian

Thank you. This is a small project with one maintainer, so the rules below exist to keep review fast, not to slow you down.

## Before you start

- Anything that changes what is stored in a vault is a spec change. Open an issue first; see [GOVERNANCE.md](GOVERNANCE.md).
- Everything else: a pull request is welcome without an issue.

## Building and testing

Go 1.23 or later, standard library only; `git` on PATH for the snapshot tests.

```
hack/ci.sh
```

That runs the licence-header check, gofmt, go vet, go build, and go test with the race detector. It is the gate named in every PR.

## Developer Certificate of Origin

Every commit must carry a `Signed-off-by:` line matching its author, which certifies the [DCO](DCO):

```
git commit -s
```

That is the whole contributor agreement. There is no CLA, and no real-name requirement: sign off with a name you use consistently and an email that reaches you. Contributions are licensed under Apache-2.0 and the copyright stays with you; the NOTICE file credits the Longhand Authors collectively.

Every Go file starts with the SPDX line; `hack/ci.sh` checks it.

## Commit messages

Conventional Commits with a closed type set: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`, `ci`, `chore`. The scope is the package or module changed. The subject is imperative, lower case, at most 72 characters. One change per commit; a subject that needs "and" is two commits.

```
feat(snapshots): list only snap commits for the active document
fix(rtf): keep hyperlink text when the field has no result group
docs(spec): add the relief field to map notes
```

Breaking changes to the spec or the CLI get a `!` after the type and a `BREAKING CHANGE:` footer.

## Pull requests

1. Branch off `main`, make the change with signed-off commits.
2. Run the gate above; it must be green.
3. Open the PR. The title follows the commit grammar; PRs are squash-merged, so the title becomes the commit on `main`.
4. Keep every heading in the PR template. If a section has nothing to report, write "None."
5. Every number in a PR body has a command behind it. "Should pass" is not a verdict; "not run (why)" is.

## The FAQ

`docs/FAQ.md` is the source of the site's FAQ page. After editing it, render the page into a checkout of `longhand-site` and commit both:

```
go run ./hack/faqgen -md docs/FAQ.md -html ../longhand-site/faq.html
```

CI runs the same tool with `-check` and fails on drift.

## Cutting a release

Releases are built by CI from a tag; nothing is built by hand. To cut `vX.Y.Z`:

1. In `CHANGELOG.md`, rename `## Unreleased: vX.Y.Z` to `## vX.Y.Z (YYYY-MM-DD)` and delete the "Not yet cut" line. That section becomes the release notes verbatim; `hack/release-notes.sh vX.Y.Z` prints what will be published and fails while the heading still says Unreleased.
2. Update the README's install lines and Status section, and the ledger row in `docs/PHASES.md`, in the same PR.
3. Merge, then tag the squash commit on `main` and push the tag:

   ```
   git tag -s vX.Y.Z -m "vX.Y.Z"
   git push origin vX.Y.Z
   ```

   `release.yml` runs the gate, builds darwin arm64 and amd64, linux amd64 and arm64, and windows amd64 with `hack/build-release.sh`, and publishes the archives with a sha256 checksum file.
4. Download one archive, check it against the checksum file, and run the acceptance check, which converts the fixture with the downloaded binary and with a `go install` of the tag and diffs the two vaults:

   ```
   shasum -a 256 -c scriv2obsidian_vX.Y.Z_checksums.txt --ignore-missing
   hack/verify-release.sh ./scriv2obsidian vX.Y.Z
   ```

   Record the result in the ledger.

The version string comes from the tag through `-ldflags`; a `go install` build reads it from the module, and a plain `go build` prints `dev`. `SOURCE_DATE_EPOCH` fixes the import timestamp so two runs match byte for byte, which the verification depends on; `hack/ci.sh`-adjacent CI runs the same check on every PR with a throwaway version.

## What not to include

No real names of collaborators, no paths from your machine, no text from a real manuscript. Test fixtures are synthetic. The maintainer greps for this before merging and will ask you to scrub.

## Reporting security issues

Not in a public issue. See [SECURITY.md](SECURITY.md).
