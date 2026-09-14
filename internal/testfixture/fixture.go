// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

// Package testfixture writes a synthetic Scrivener 3 package that exercises
// every feature the importer handles. It contains no text from any real
// manuscript. The converter tests and hack/mkfixture share it.
package testfixture

import (
	"os"
	"path/filepath"
)

// Name is the package name the fixture is written under.
const Name = "Fixture Novel.scriv"

// Write creates the package at root, which should end in Name.
func Write(root string) error {
	var firstErr error
	mk := func(rel string, content string) {
		if firstErr != nil {
			return
		}
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			firstErr = err
			return
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			firstErr = err
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
	return firstErr
}
