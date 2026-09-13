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

That runs gofmt, go vet, go build, and go test with the race detector. It is the gate named in every PR.

## Developer Certificate of Origin

Every commit must carry a `Signed-off-by:` line matching its author, which certifies the [DCO](DCO):

```
git commit -s
```

That is the whole contributor agreement. There is no CLA.

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

## What not to include

No real names of collaborators, no paths from your machine, no text from a real manuscript. Test fixtures are synthetic. The maintainer greps for this before merging and will ask you to scrub.

## Reporting security issues

Not in a public issue. See [SECURITY.md](SECURITY.md).
