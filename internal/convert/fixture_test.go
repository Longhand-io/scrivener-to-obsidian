package convert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildFixture writes a synthetic Scrivener 3 package. No text from any real manuscript.
func buildFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "Fixture Novel.scriv")
	mk := func(rel string, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("Fixture Novel.scrivx", `<?xml version="1.0" encoding="UTF-8"?>
<ScrivenerProject Identifier="FX" Version="2.0" Creator="SCRMAC-3.5.2-17486">
  <Binder>
    <BinderItem UUID="DRAFT" Type="DraftFolder" Created="2021-05-22 19:12:38 -0700" Modified="2025-08-18 18:25:55 -0700"><Title>Manuscript</Title>
      <Children>
        <BinderItem UUID="P1" Type="Folder" Created="2021-05-22 19:12:38 -0700" Modified="2021-05-22 19:12:38 -0700"><Title>Part One</Title>
          <MetaData><LabelID>1</LabelID><IncludeInCompile>Yes</IncludeInCompile></MetaData>
          <Children>
            <BinderItem UUID="C1" Type="Text" Created="2021-05-22 19:12:38 -0700" Modified="2021-06-01 10:00:00 -0700"><Title>The Letter: A Beginning?</Title>
              <MetaData><LabelID>2</LabelID><StatusID>3</StatusID><IncludeInCompile>Yes</IncludeInCompile>
                <CustomMetaData><MetaDataItem><FieldID>source</FieldID><Value>Attic box</Value></MetaDataItem></CustomMetaData>
              </MetaData>
              <Keywords><KeywordID>10</KeywordID></Keywords>
              <Bookmarks><Bookmark BinderUUID="C2" Destination="[Internal Link]">Two</Bookmark></Bookmarks>
            </BinderItem>
            <BinderItem UUID="C2" Type="Text"><Title>The Tin</Title><MetaData><IncludeInCompile>No</IncludeInCompile></MetaData></BinderItem>
            <BinderItem UUID="C3" Type="Text"><Title>Verse</Title></BinderItem>
          </Children>
        </BinderItem>
        <BinderItem UUID="P2" Type="Folder"><Title>Part Two</Title></BinderItem>
      </Children>
    </BinderItem>
    <BinderItem UUID="RES" Type="ResearchFolder"><Title>Research</Title>
      <Children>
        <BinderItem UUID="PDF1" Type="PDF"><Title>Gazette</Title></BinderItem>
        <BinderItem UUID="N1" Type="Text"><Title>Map notes</Title></BinderItem>
      </Children>
    </BinderItem>
    <BinderItem UUID="TRASH" Type="TrashFolder"><Title>Trash</Title>
      <Children><BinderItem UUID="OLD" Type="Text"><Title>Old</Title></BinderItem></Children>
    </BinderItem>
  </Binder>
  <LabelSettings><Title>Label</Title><Labels><Label ID="-1">No Label</Label><Label ID="1">Red</Label><Label ID="2">Blue</Label></Labels></LabelSettings>
  <StatusSettings><Title>Status</Title><StatusItems><Status ID="-1">No Status</Status><Status ID="3">First Draft</Status></StatusItems></StatusSettings>
  <Keywords><Keyword ID="10"><Title>Winter Fair</Title></Keyword></Keywords>
  <ProjectSettings><CustomMetaDataSettings><MetaDataField ID="source" Type="Text"><Title>Source</Title></MetaDataField></CustomMetaDataSettings></ProjectSettings>
</ScrivenerProject>`)
	hdr := `{\rtf1\ansi\ansicpg1252\cocoartf2870{\fonttbl\f0\fswiss\fcharset0 Helvetica;}{\colortbl;\red255\green255\blue255;}\pard\f0\fs24 `
	// C1: bold, italic, a footnote comment, a plain comment, an internal link, a web link, an inline image, a curly quote
	png := "89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000d4944415478da63f8cfc0f00f0003030101000018dd8db00000000049454e44ae426082"
	mk("Files/Data/C1/content.rtf", hdr+`Mara found the {\b letter} and read it {\i twice}.\
{\field{\*\fldinst{HYPERLINK "scrivcmt://F1"}}{\fldrslt Doctrine}} says so.\
{\field{\*\fldinst{HYPERLINK "scrivcmt://K1"}}{\fldrslt Cut this}} maybe.\
See {\field{\*\fldinst{HYPERLINK "scrivlnk://C2"}}{\fldrslt the tin}} and {\field{\*\fldinst{HYPERLINK "https://example.com/x"}}{\fldrslt the web}}.\
\'93Quoted\'94 text.\
{\pict\pngblip `+png+`}\
}`)
	mk("Files/Data/C1/synopsis.txt", "Mara finds the letter.\n")
	mk("Files/Data/C1/notes.rtf", hdr+`Check the postmark date.\
}`)
	mk("Files/Data/C1/content.comments", `<?xml version="1.0" encoding="UTF-8"?>
<Comments>
  <Comment ID="F1" Footnote="Yes" Color="0.9 0.9 0.9"><![CDATA[{\rtf1\ansi{\fonttbl\f0\fswiss Helvetica;}\f0\fs24 Letters 3:18-19}]]></Comment>
  <Comment ID="K1" Footnote="No" Color="0.9 0.9 0.9"><![CDATA[{\rtf1\ansi{\fonttbl\f0\fswiss Helvetica;}\f0\fs24 Too on the nose}]]></Comment>
</Comments>`)
	mk("Files/Data/C2/content.rtf", hdr+`She decides not to open it. Then she opens it.\
}`)
	mk("Files/Data/C3/content.rtf", hdr+`Line one\
line two\
\
Stanza two\
}`)
	mk("Files/Data/P1/content.rtf", hdr+`Part One begins here.\
}`)
	mk("Files/Data/N1/content.rtf", hdr+`Harbour to the east.\
}`)
	mk("Files/Data/PDF1/content.pdf", "%PDF-1.4 fake\n")
	mk("Files/Data/PDF1/synopsis.txt", "The flood report.")
	mk("Files/Data/OLD/content.rtf", hdr+`Old words.\
}`)
	// snapshots for C1: two titled, one untitled, one in another zone offset
	mk("Snapshots/C1.snapshots/index.xml", `<?xml version="1.0" encoding="UTF-8"?>
<Snapshots Version="1.0">
  <Snapshot><Title>First Draft</Title><Date>2020-01-19 21:27:36 -0700</Date></Snapshot>
  <Snapshot><Title>Second Draft</Title><Date>2021-10-02 18:50:31 -0700</Date></Snapshot>
</Snapshots>`)
	mk("Snapshots/C1.snapshots/2020-01-19-20-27-36-0800.rtf", hdr+`Mara found the letter.\
}`)
	mk("Snapshots/C1.snapshots/2021-10-02-18-50-31-0700.rtf", hdr+`Mara found the letter and read it once.\
}`)
	mk("Snapshots/C1.snapshots/2020-06-01-09-00-00-0700.rtf", hdr+`Untitled middle version.\
}`)
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
