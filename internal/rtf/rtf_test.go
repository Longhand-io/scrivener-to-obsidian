package rtf

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMarkdown(t *testing.T) {
	png := func(img *Image, n int) string {
		return fmt.Sprintf("![](%s-%d:%x)", img.Ext, n, img.Data)
	}
	tests := []struct {
		name string
		rtf  string
		opt  Options
		want string
	}{
		{
			name: "bold",
			rtf:  `{\rtf1 {\b bold} plain}`,
			want: "**bold** plain\n",
		},
		{
			name: "italic",
			rtf:  `{\rtf1 {\i italic}}`,
			want: "*italic*\n",
		},
		{
			name: "bold and italic",
			rtf:  `{\rtf1 {\b\i both}}`,
			want: "***both***\n",
		},
		{
			name: "bold toggle",
			rtf:  `{\rtf1 \b on\b0 off}`,
			want: "**on**off\n",
		},
		{
			name: "lists/bullet",
			rtf:  `{\rtf1\ls1\ilvl0{\listtext\bullet \tab}one\par\ls1\ilvl0{\listtext\bullet \tab}two}`,
			want: "- one\n- two\n",
		},
		{
			name: "lists/numbered",
			rtf:  `{\rtf1\ls1\ilvl0{\listtext 1.\tab}one\par\ls1\ilvl0{\listtext 2.\tab}two}`,
			want: "1. one\n1. two\n",
		},
		{
			name: "lists/nested ilvl",
			rtf:  `{\rtf1\ls1\ilvl0{\listtext\bullet \tab}one\par\ls1\ilvl1{\listtext\bullet \tab}nested}`,
			want: "- one\n\t- nested\n",
		},
		{
			name: "lists/nested li",
			rtf:  `{\rtf1\ls1\ilvl0\li1440{\listtext\bullet \tab}nested}`,
			want: "\t- nested\n",
		},
		{
			name: "lists/then paragraph",
			rtf:  `{\rtf1\ls1\ilvl0{\listtext\bullet \tab}one\par\pard two}`,
			want: "- one\n\ntwo\n",
		},
		{
			name: "link/http",
			rtf:  `{\rtf1 {\field{\*\fldinst{HYPERLINK "https://example.com"}}{\fldrslt{click here}}}}`,
			want: "[click here](https://example.com)\n",
		},
		{
			name: "link/http no inner braces",
			rtf:  `{\rtf1 {\field{\*\fldinst HYPERLINK "https://ex.com"}{\fldrslt click}}}`,
			want: "[click](https://ex.com)\n",
		},
		{
			name: "link/bold text",
			rtf:  `{\rtf1 {\field{\*\fldinst{HYPERLINK "https://ex.com"}}{\fldrslt{\b click}}}}`,
			want: "[**click**](https://ex.com)\n",
		},
		{
			name: "comment anchor",
			rtf:  `{\rtf1 see {\field{\*\fldinst{HYPERLINK "scrivcmt://C1"}}{\fldrslt{this}}} here}`,
			want: "see [this](scrivcmt://C1) here\n",
		},
		{
			name: "internal link",
			rtf:  `{\rtf1 {\field{\*\fldinst{HYPERLINK "scrivlnk://DEADBEEF"}}{\fldrslt{Chapter Two}}}}`,
			want: "[Chapter Two](scrivlnk://DEADBEEF)\n",
		},
		{
			name: "comment anchor via Link",
			rtf:  `{\rtf1 {\field{\*\fldinst{HYPERLINK "scrivcmt://C1"}}{\fldrslt{word}}}}`,
			opt: Options{Link: func(url, text string) string {
				return text + "%%" + url + "%%"
			}},
			want: "word%%scrivcmt://C1%%\n",
		},
		{
			name: "unicode/u hex fallback",
			rtf:  `{\rtf1 caf\u233\'e9}`,
			want: "café\n",
		},
		{
			name: "unicode/uc0",
			rtf:  `{\rtf1 \uc0\u233}`,
			want: "é\n",
		},
		{
			name: "unicode/hex byte",
			rtf:  `{\rtf1 caf\'e9}`,
			want: "café\n",
		},
		{
			name: "unicode/cp1252 euro",
			rtf:  `{\rtf1 \'80}`,
			want: "€\n",
		},
		{
			name: "unicode/negative",
			rtf:  `{\rtf1 \uc0\u-4964}`,
			want: "\uec9c\n",
		},
		{
			name: "unicode/named dashes and quotes",
			rtf:  `{\rtf1 \emdash\endash\lquote\rquote\ldblquote\rdblquote\bullet}`,
			want: "—–‘’“”•\n",
		},
		{
			name: "image/png",
			rtf:  `{\rtf1 before{\pict\pngblip 89504e470d0a1a0a}after}`,
			opt:  Options{Image: png},
			want: "before![](png-1:89504e470d0a1a0a)after\n",
		},
		{
			name: "image/jpeg",
			rtf:  `{\rtf1 {\pict\jpegblip ffd8ffe0}}`,
			opt:  Options{Image: png},
			want: "![](jpg-1:ffd8ffe0)\n",
		},
		{
			name: "image/shppict and index",
			rtf:  `{\rtf1 {\*\shppict{\pict\pngblip aaaa}}{\pict\jpegblip bbbb}}`,
			opt:  Options{Image: png},
			want: "![](png-1:aaaa)![](jpg-2:bbbb)\n",
		},
		{
			name: "image/without writer",
			rtf:  `{\rtf1 before{\pict\pngblip 89504e470d0a1a0a}after}`,
			want: "beforeafter\n",
		},
		{
			name: "poetry/joins",
			rtf:  `{\rtf1 line one\par line two\par\par stanza two}`,
			opt:  Options{Poetry: true},
			want: "line one  \nline two\n\nstanza two\n",
		},
		{
			name: "poetry/off",
			rtf:  `{\rtf1 line one\par line two\par\par stanza two}`,
			want: "line one\n\nline two\n\nstanza two\n",
		},
		{
			name: "poetry/keeps style",
			rtf:  `{\rtf1 line one\par {\b line two}}`,
			opt:  Options{Poetry: true},
			want: "line one  \n**line two**\n",
		},
		{
			name: "poetry/does not join lists",
			rtf:  `{\rtf1 verse\par more\par\ls1\ilvl0{\listtext\bullet \tab}item}`,
			opt:  Options{Poetry: true},
			want: "verse  \nmore\n\n- item\n",
		},
		{
			name: "fonttbl skipped",
			rtf:  `{\rtf1\ansi\ansicpg1252{\fonttbl\f0\fswiss Helvetica;}{\colortbl;\red255\green255\blue255;}\f0 hello {\b world}}`,
			want: "hello **world**\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse([]byte(tt.rtf)).Markdown(tt.opt)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		rtf   string
		check func(*testing.T, *Doc)
	}{
		{
			name: "bold italic flags",
			rtf:  `{\rtf1 {\b bold} {\i italic} {\b\i both}}`,
			check: func(t *testing.T, d *Doc) {
				runs := d.Paras[0].Runs
				if !runs[0].Style.Bold || runs[0].Style.Italic || runs[0].Text != "bold" {
					t.Errorf("bold run %+v", runs[0])
				}
				if runs[2].Style.Bold || !runs[2].Style.Italic || runs[2].Text != "italic" {
					t.Errorf("italic run %+v", runs[2])
				}
				if !runs[4].Style.Bold || !runs[4].Style.Italic || runs[4].Text != "both" {
					t.Errorf("both run %+v", runs[4])
				}
			},
		},
		{
			name: "lists levels and ordered",
			rtf:  `{\rtf1\ls1\ilvl0{\listtext\bullet \tab}a\par\ls1\ilvl0{\listtext 1.\tab}b\par\ls1\ilvl1{\listtext\bullet \tab}c}`,
			check: func(t *testing.T, d *Doc) {
				if len(d.Paras) != 3 {
					t.Fatalf("paras %d", len(d.Paras))
				}
				if d.Paras[0].ListLevel != 1 || d.Paras[0].ListOrdered || d.Paras[0].Text() != "a" {
					t.Errorf("bullet %+v %q", d.Paras[0], d.Paras[0].Text())
				}
				if d.Paras[1].ListLevel != 1 || !d.Paras[1].ListOrdered || d.Paras[1].Text() != "b" {
					t.Errorf("numbered %+v %q", d.Paras[1], d.Paras[1].Text())
				}
				if d.Paras[2].ListLevel != 2 || d.Paras[2].Text() != "c" {
					t.Errorf("nested %+v %q", d.Paras[2], d.Paras[2].Text())
				}
			},
		},
		{
			name: "comment and link urls",
			rtf:  `{\rtf1 {\field{\*\fldinst{HYPERLINK "scrivcmt://C1"}}{\fldrslt{cmt}}} {\field{\*\fldinst{HYPERLINK "https://ex.com"}}{\fldrslt{web}}} {\field{\*\fldinst{HYPERLINK "scrivlnk://ID"}}{\fldrslt{doc}}}}`,
			check: func(t *testing.T, d *Doc) {
				var urls []string
				for _, r := range d.Paras[0].Runs {
					if r.URL != "" {
						urls = append(urls, r.URL+"|"+r.Text)
					}
				}
				want := []string{"scrivcmt://C1|cmt", "https://ex.com|web", "scrivlnk://ID|doc"}
				if len(urls) != 3 || urls[0] != want[0] || urls[1] != want[1] || urls[2] != want[2] {
					t.Errorf("urls %v, want %v", urls, want)
				}
			},
		},
		{
			name: "unicode runs",
			rtf:  `{\rtf1 caf\u233\'e9 \uc0\u8212 \'80}`,
			check: func(t *testing.T, d *Doc) {
				got := d.Paras[0].Text()
				want := "café —€"
				if got != want {
					t.Errorf("text %q, want %q", got, want)
				}
			},
		},
		{
			name: "png image bytes",
			rtf:  `{\rtf1 {\pict\pngblip 89504e470d0a1a0a}}`,
			check: func(t *testing.T, d *Doc) {
				img := firstImage(d)
				if img == nil {
					t.Fatal("missing image")
				}
				if img.Ext != "png" || !bytes.Equal(img.Data, []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) {
					t.Errorf("image ext=%s data=%x", img.Ext, img.Data)
				}
			},
		},
		{
			name: "jpeg and whitespace hex",
			rtf:  "{\\rtf1 {\\pict\\jpegblip ff d8 ff e0}}",
			check: func(t *testing.T, d *Doc) {
				img := firstImage(d)
				if img == nil {
					t.Fatal("missing image")
				}
				if img.Ext != "jpg" || !bytes.Equal(img.Data, []byte{0xff, 0xd8, 0xff, 0xe0}) {
					t.Errorf("image ext=%s data=%x", img.Ext, img.Data)
				}
			},
		},
		{
			name: "image dropped without ext or odd hex",
			rtf:  `{\rtf1 {\pict 89504e470d0a1a0a}{\pict\pngblip 89504e4}}`,
			check: func(t *testing.T, d *Doc) {
				if img := firstImage(d); img != nil {
					t.Errorf("unexpected image %+v", img)
				}
			},
		},
		{
			name: "poetry join is a render option",
			rtf:  `{\rtf1 a\par b}`,
			check: func(t *testing.T, d *Doc) {
				if len(d.Paras) != 2 || d.Paras[0].Text() != "a" || d.Paras[1].Text() != "b" {
					t.Errorf("parse should keep separate paras: %+v", d.Paras)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, Parse([]byte(tt.rtf)))
		})
	}
}

func TestPlainText(t *testing.T) {
	d := Parse([]byte(`{\rtf1 {\b bold} {\field{\*\fldinst{HYPERLINK "scrivcmt://C1"}}{\fldrslt{cmt}}}\par two}`))
	if got, want := d.PlainText(), "bold cmt\ntwo"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if d.WordCount() != 3 {
		t.Errorf("WordCount = %d, want 3", d.WordCount())
	}
}

func firstImage(d *Doc) *Image {
	for _, p := range d.Paras {
		for _, r := range p.Runs {
			if r.Image != nil {
				return r.Image
			}
		}
	}
	return nil
}
