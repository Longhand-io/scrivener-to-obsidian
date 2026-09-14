// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

package convert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/longhand-io/scrivener-to-obsidian/internal/testfixture"
)

// buildFixture writes the synthetic package from internal/testfixture into a temp dir.
func buildFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), testfixture.Name)
	if err := testfixture.Write(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return string(b)
}

func TestConvertFixture(t *testing.T) {
	root := buildFixture(t)
	out := filepath.Join(t.TempDir(), "vault")
	p, err := openProject(root)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(p, Options{OutDir: out, Poetry: false})
	if err != nil {
		t.Fatal(err)
	}
	pd := filepath.Join(out, "Fixture Novel")
	// layout
	must := []string{
		"_Project.md",
		".scriv2obsidian.json",
		"Manuscript/01 Part One/01 Part One.md",
		"Manuscript/01 Part One/01 The Letter- A Beginning-.md",
		"Manuscript/01 Part One/02 The Tin.md",
		"Manuscript/01 Part One/03 Verse.md",
		"Manuscript/02 Part Two",
		"Research/01 Gazette.pdf",
		"Research/01 Gazette.md",
		"Research/02 Map notes.md",
		"_attachments/C1-1.png",
		"_snapshots/C1/2020-01-20T04-27-36Z First Draft.md",
		"_snapshots/C1/2020-06-01T16-00-00Z Untitled.md",
		"_snapshots/C1/2021-10-03T01-50-31Z Second Draft.md",
	}
	for _, m := range must {
		if _, err := os.Stat(filepath.Join(pd, m)); err != nil {
			t.Errorf("missing %s", m)
		}
	}
	if _, err := os.Stat(filepath.Join(pd, "Trash")); err == nil {
		t.Error("trash was exported without -include-trash")
	}
	c1 := read(t, filepath.Join(pd, "Manuscript/01 Part One/01 The Letter- A Beginning-.md"))
	for _, want := range []string{
		"id: \"C1\"", "title: \"The Letter: A Beginning?\"", "type: \"text\"",
		"created: \"2021-05-22T19:12:38-07:00\"", "synopsis: \"Mara finds the letter.\"",
		"label: \"Blue\"", "status: \"First Draft\"", "  - \"winter-fair\"", "include: true",
		"  - \"[[02 The Tin]]\"", "  \"Source\": \"Attic box\"",
		"**letter**", "*twice*", "Doctrine[^1]", "[^1]: Letters 3:18-19",
		"Cut this%% Too on the nose %%", "[[02 The Tin|the tin]]", "[the web](https://example.com/x)",
		"“Quoted” text.", "![[C1-1.png]]", "> [!note] Notes\n> Check the postmark date.",
	} {
		if !strings.Contains(c1, want) {
			t.Errorf("C1 missing %q\n---\n%s", want, c1)
		}
	}
	if strings.Contains(c1, "order:") {
		t.Error("order field written without -no-prefix")
	}
	c2 := read(t, filepath.Join(pd, "Manuscript/01 Part One/02 The Tin.md"))
	if !strings.Contains(c2, "include: false") {
		t.Errorf("C2 include: %s", c2)
	}
	if strings.Contains(c2, "label:") || strings.Contains(c2, "tags:") {
		t.Errorf("C2 has unset fields: %s", c2)
	}
	folder := read(t, filepath.Join(pd, "Manuscript/01 Part One/01 Part One.md"))
	if !strings.Contains(folder, "type: \"folder\"") || !strings.Contains(folder, "Part One begins here.") || !strings.Contains(folder, "label: \"Red\"") {
		t.Errorf("folder note: %s", folder)
	}
	side := read(t, filepath.Join(pd, "Research/01 Gazette.md"))
	if !strings.Contains(side, "type: \"pdf\"") || !strings.Contains(side, "![[01 Gazette.pdf]]") || !strings.Contains(side, "The flood report.") {
		t.Errorf("sidecar: %s", side)
	}
	verse := read(t, filepath.Join(pd, "Manuscript/01 Part One/03 Verse.md"))
	if !strings.Contains(verse, "Line one\n\nline two\n\nStanza two") {
		t.Errorf("prose mode verse: %q", verse)
	}
	snap := read(t, filepath.Join(pd, "_snapshots/C1/2020-01-20T04-27-36Z First Draft.md"))
	if !strings.Contains(snap, "id: \"C1\"") || !strings.Contains(snap, "Mara found the letter.") {
		t.Errorf("snapshot content: %s", snap)
	}
	proj := read(t, filepath.Join(pd, "_Project.md"))
	for _, want := range []string{"longhand: 1", "title: \"Fixture Novel\"", "source_creator: \"SCRMAC-3.5.2-17486\"", "  - \"Red\"", "  - \"First Draft\"", "  - \"Winter Fair\""} {
		if !strings.Contains(proj, want) {
			t.Errorf("project note missing %q", want)
		}
	}
	man := read(t, filepath.Join(pd, ".scriv2obsidian.json"))
	for _, want := range []string{`"documents": 6`, `"snapshots": 3`, `"assets": 1`, `"attachments": 1`, `"skipped_trash": 1`, `"uuid": "OLD"`} {
		if !strings.Contains(man, want) {
			t.Errorf("manifest missing %q\n%s", want, man)
		}
	}
	if res.Snapshots != 3 || res.Attachments != 1 {
		t.Errorf("result counts: %+v", res)
	}
}

func TestConvertPoetryAndNoPrefixAndTrash(t *testing.T) {
	root := buildFixture(t)
	out := filepath.Join(t.TempDir(), "vault")
	p, _ := openProject(root)
	if _, err := Run(p, Options{OutDir: out, Poetry: true, NoPrefix: true, IncludeTrash: true}); err != nil {
		t.Fatal(err)
	}
	pd := filepath.Join(out, "Fixture Novel")
	verse := read(t, filepath.Join(pd, "Manuscript/Part One/Verse.md"))
	if !strings.Contains(verse, "Line one  \nline two\n\nStanza two") {
		t.Errorf("poetry mode: %q", verse)
	}
	if !strings.Contains(verse, "order: 3") {
		t.Errorf("order field missing: %q", verse)
	}
	if _, err := os.Stat(filepath.Join(pd, "Trash/Old.md")); err != nil {
		t.Error("trash not exported with -include-trash")
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	root := buildFixture(t)
	out := filepath.Join(t.TempDir(), "vault")
	p, _ := openProject(root)
	res, err := Run(p, Options{OutDir: out, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("dry run created the output directory")
	}
	if res.Manifest.Counts["documents"] != 6 {
		t.Errorf("dry run counts: %v", res.Manifest.Counts)
	}
}
