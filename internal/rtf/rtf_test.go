// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

package rtf

import (
	"strings"
	"testing"
)

const hdr = `{\rtf1\ansi\ansicpg1252\cocoartf2870{\fonttbl\f0\fswiss\fcharset0 Helvetica;}{\colortbl;\red255\green255\blue255;}\pard\f0\fs24 `

func md(t *testing.T, body string, o Options) string {
	t.Helper()
	return Parse([]byte(hdr + body + "}")).Markdown(o)
}

func TestInlineFormatting(t *testing.T) {
	got := md(t, `Plain {\b bold} and {\i italic} and {\b\i both} and {\strike gone} and {\ul under} x{\super 2} H{\sub 2}O.\`+"\n", Options{})
	for _, w := range []string{"**bold**", "*italic*", "***both***", "~~gone~~", "<u>under</u>", "x<sup>2</sup>", "H<sub>2</sub>O"} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in %q", w, got)
		}
	}
}

func TestParagraphsAndLineBreaks(t *testing.T) {
	got := md(t, `One\`+"\n"+`Two\line still two\`+"\n"+`\`+"\n"+`Four\`+"\n", Options{})
	if got != "One\n\nTwo  \nstill two\n\nFour\n" {
		t.Errorf("got %q", got)
	}
}

func TestPoetryMode(t *testing.T) {
	got := md(t, `Line one\`+"\n"+`line two\`+"\n"+`\`+"\n"+`Stanza two\`+"\n", Options{Poetry: true})
	if got != "Line one  \nline two\n\nStanza two\n" {
		t.Errorf("got %q", got)
	}
}

func TestSpecialCharacters(t *testing.T) {
	got := md(t, `\'93Quoted\'94 \'96 dash \emdash{} em \uc1\u8217 ? apostrophe caf\'e9 50% off\`+"\n", Options{})
	for _, w := range []string{"“Quoted”", "–", "—", "’", "café", "50% off"} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in %q", w, got)
		}
	}
}

func TestMarkdownEscaping(t *testing.T) {
	got := md(t, `a * b _c_ [d]\`+"\n"+`# not heading\`+"\n"+`- not a list\`+"\n"+`1. not ordered\`+"\n", Options{})
	for _, w := range []string{`a \* b \_c\_ \[d\]`, `\# not heading`, `\- not a list`, `1\. not ordered`} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in %q", w, got)
		}
	}
}

func TestLists(t *testing.T) {
	body := `{\*\listtable{\list\listtemplateid1{\listlevel\levelnfc23\levelmarker \{disc\}}\listid1}}` +
		`\pard\ls1\ilvl0\li720\fi-360{\listtext	\'95	}First\` + "\n" +
		`\pard\ls1\ilvl0\li720\fi-360{\listtext	\'95	}Second\` + "\n" +
		`\pard\ls2\ilvl0\li720\fi-360{\listtext	1.	}Numbered\` + "\n"
	got := md(t, body, Options{})
	if !strings.Contains(got, "- First\n- Second") {
		t.Errorf("bullets: %q", got)
	}
	if !strings.Contains(got, "1. Numbered") {
		t.Errorf("ordered: %q", got)
	}
}

func TestHyperlinkFieldsSurvive(t *testing.T) {
	body := `See {\field{\*\fldinst{HYPERLINK "https://example.com/a_b"}}{\fldrslt the site}} and ` +
		`{\field{\*\fldinst{HYPERLINK "scrivcmt://ABC"}}{\fldrslt anchored}} and ` +
		`{\field{\*\fldinst{HYPERLINK "scrivlnk://DEF"}}{\fldrslt \cf0 linked}}.\` + "\n"
	var seen []string
	got := md(t, body, Options{Link: func(url, text string) string {
		seen = append(seen, url+"|"+text)
		return "<" + text + ">"
	}})
	if len(seen) != 3 {
		t.Fatalf("resolver calls %v", seen)
	}
	if seen[0] != "https://example.com/a_b|the site" || seen[1] != "scrivcmt://ABC|anchored" || seen[2] != "scrivlnk://DEF|linked" {
		t.Errorf("resolver args %v", seen)
	}
	if !strings.Contains(got, "See <the site> and <anchored> and <linked>.") {
		t.Errorf("got %q", got)
	}
	// default resolver
	got = md(t, `{\field{\*\fldinst{HYPERLINK "https://x.y"}}{\fldrslt here}}\`+"\n", Options{})
	if !strings.Contains(got, "[here](https://x.y)") {
		t.Errorf("default link: %q", got)
	}
}

func TestInlineImageIgnoresBlipUID(t *testing.T) {
	// 1x1 PNG. The blipuid group carries a hex id that must not leak into the image bytes.
	png := "89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000d4944415478da63f8cfc0f00f0003030101000018dd8db00000000049454e44ae426082"
	body := `Before {\*\shppict {\pict\pngblip\picw1\pich1\picwgoal20\pichgoal20\bliptag-12345{\*\blipuid aaaaaaaabbbbbbbbccccccccdddddddd}` + "\n" + png + `}} after\` + "\n"
	var img *Image
	got := md(t, body, Options{Image: func(i *Image, n int) string { img = i; return "![[img]]" }})
	if img == nil {
		t.Fatal("no image emitted")
	}
	if img.Ext != "png" || len(img.Data) != len(png)/2 || string(img.Data[:4]) != "\x89PNG" {
		t.Errorf("image bytes wrong: ext=%s len=%d head=%x", img.Ext, len(img.Data), img.Data[:4])
	}
	if !strings.Contains(got, "Before ![[img]] after") {
		t.Errorf("got %q", got)
	}
}

func TestSkipsHeaderTablesAndUnknownDestinations(t *testing.T) {
	got := md(t, `{\*\expandedcolortbl;;}{\info{\title secret}}{\*\unknownthing hidden}Visible\`+"\n", Options{})
	if strings.Contains(got, "secret") || strings.Contains(got, "hidden") || !strings.Contains(got, "Visible") {
		t.Errorf("got %q", got)
	}
}

func TestPlainTextAndWordCount(t *testing.T) {
	d := Parse([]byte(hdr + `One {\b two} three.\` + "\n" + `Four\` + "\n}"))
	if d.PlainText() != "One two three.\nFour" || d.WordCount() != 4 {
		t.Errorf("plain=%q words=%d", d.PlainText(), d.WordCount())
	}
}

func TestMalformedInputDoesNotPanic(t *testing.T) {
	for _, in := range []string{"", "{", "}", `{\rtf1 \'zz 香9 {\pict\pngblip zz}`, `{\field{\*\fldinst{HYPERLINK "x"}}`} {
		_ = Parse([]byte(in)).Markdown(Options{})
	}
}
