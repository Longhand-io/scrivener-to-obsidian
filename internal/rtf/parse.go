package rtf

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

// Style is the character formatting active for a run.
type Style struct {
	Bold, Italic, Underline, Strike, Super, Sub bool
}

// Image is an inline picture embedded in the RTF as hex data.
type Image struct {
	Data []byte
	Ext  string // "png" or "jpg"
}

// Run is a piece of inline content.
type Run struct {
	Text  string
	Style Style
	URL   string // non-empty when the run is inside a HYPERLINK field
	Break bool   // \line
	Image *Image
}

// Para is a paragraph.
type Para struct {
	Runs        []Run
	ListLevel   int // 0 = not a list item
	ListOrdered bool
}

// Doc is a parsed RTF document.
type Doc struct {
	Paras []Para
}

// IsEmpty reports whether the paragraph has no visible content.
func (p Para) IsEmpty() bool {
	for _, r := range p.Runs {
		if r.Image != nil || strings.TrimSpace(r.Text) != "" {
			return false
		}
	}
	return true
}

// Text returns the paragraph's plain text.
func (p Para) Text() string {
	var b strings.Builder
	for _, r := range p.Runs {
		if r.Break {
			b.WriteByte('\n')
		}
		b.WriteString(r.Text)
	}
	return b.String()
}

type fieldCtx struct {
	url strings.Builder
}

type pictCtx struct {
	hex strings.Builder
	ext string
}

type state struct {
	style     Style
	uc        int
	skip      bool
	inFldinst bool
	inLink    bool
	field     *fieldCtx
	pict      *pictCtx
	listtext  *strings.Builder
}

type parser struct {
	lx        *lexer
	stack     []state
	cur       state
	doc       Doc
	para      Para
	paraLs    int
	paraIlvl  int
	paraLi    int
	listtext  string
	skipChars int
	pending   token
	hasPend   bool
}

// Parse converts RTF bytes into a Doc. It never fails on malformed input; it
// returns whatever it could recover.
func Parse(data []byte) *Doc {
	p := &parser{lx: &lexer{data: data}, cur: state{uc: 1}}
	p.run()
	return &p.doc
}

func (p *parser) next() token {
	if p.hasPend {
		p.hasPend = false
		return p.pending
	}
	return p.lx.next()
}

func (p *parser) peek() token {
	if !p.hasPend {
		p.pending = p.lx.next()
		p.hasPend = true
	}
	return p.pending
}

var skipDests = map[string]bool{
	"fonttbl": true, "colortbl": true, "stylesheet": true, "info": true,
	"listtable": true, "listoverridetable": true, "header": true, "footer": true,
	"headerl": true, "headerr": true, "footerl": true, "footerr": true,
	"nonshppict": true, "xmlnstbl": true, "themedata": true, "colorschememapping": true,
	"latentstyles": true, "datastore": true, "generator": true, "revtbl": true,
	"pntext": true, "object": true, "levelmarker": true,
}

func (p *parser) run() {
	for {
		t := p.next()
		switch t.kind {
		case tokEOF:
			p.flushPara()
			return
		case tokOpen:
			p.stack = append(p.stack, p.cur)
			// Copy-on-open: children inherit but changes do not leak upward.
			p.cur.listtext = nil
			nt := p.peek()
			if nt.kind == tokControl && nt.word == "*" {
				p.next()
				dt := p.peek()
				if dt.kind == tokControl {
					p.next()
					switch dt.word {
					case "fldinst":
						p.cur.inFldinst = true
					case "shppict":
						// contains a \pict group; keep parsing
					default:
						p.cur.skip = true
					}
				}
			}
		case tokClose:
			closed := p.cur
			if len(p.stack) == 0 {
				p.flushPara()
				return
			}
			p.cur = p.stack[len(p.stack)-1]
			p.stack = p.stack[:len(p.stack)-1]
			if closed.pict != nil && p.cur.pict == nil {
				p.finishPict(closed.pict)
			}
			if closed.listtext != nil {
				p.listtext = closed.listtext.String()
			}
		case tokControl:
			p.control(t)
		case tokText:
			p.text(t.text)
		}
	}
}

func (p *parser) text(b []byte) {
	if len(b) == 0 {
		return
	}
	if p.skipChars > 0 {
		s := string(b)
		for p.skipChars > 0 && len(s) > 0 {
			_, size := utf8.DecodeRuneInString(s)
			s = s[size:]
			p.skipChars--
		}
		b = []byte(s)
		if len(b) == 0 {
			return
		}
	}
	switch {
	case p.cur.pict != nil:
		p.cur.pict.hex.Write(b)
	case p.cur.inFldinst:
		if p.cur.field != nil {
			p.cur.field.url.Write(b)
		}
	case p.cur.listtext != nil:
		p.cur.listtext.Write(b)
	case p.cur.skip:
		return
	default:
		p.emit(Run{Text: string(b), Style: p.cur.style, URL: p.linkURL()})
	}
}

func (p *parser) linkURL() string {
	if p.cur.inLink && p.cur.field != nil {
		return strings.TrimSpace(extractHyperlink(p.cur.field.url.String()))
	}
	return ""
}

func extractHyperlink(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "HYPERLINK")
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	if i := strings.Index(s, "\""); i >= 0 {
		s = s[:i]
	}
	return s
}

func (p *parser) emit(r Run) {
	p.para.Runs = append(p.para.Runs, r)
}

func (p *parser) flushPara() {
	if p.paraLs > 0 || p.paraIlvl > 0 {
		level := p.paraIlvl + 1
		if p.paraIlvl == 0 && p.paraLi > 720 {
			level = p.paraLi / 720
			if level < 1 {
				level = 1
			}
		}
		p.para.ListLevel = level
		lt := strings.TrimSpace(p.listtext)
		p.para.ListOrdered = strings.ContainsAny(lt, "0123456789") || strings.HasSuffix(lt, ".") && !strings.ContainsAny(lt, "•◦▪-–")
	}
	if len(p.para.Runs) > 0 || p.para.ListLevel > 0 {
		p.doc.Paras = append(p.doc.Paras, p.para)
	} else {
		p.doc.Paras = append(p.doc.Paras, Para{})
	}
	p.para = Para{}
	p.listtext = ""
}

func (p *parser) finishPict(pc *pictCtx) {
	raw := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			return r
		}
		return -1
	}, pc.hex.String())
	data, err := hex.DecodeString(raw)
	if err != nil || len(data) == 0 || pc.ext == "" {
		return
	}
	p.emit(Run{Image: &Image{Data: data, Ext: pc.ext}})
}

func (p *parser) control(t token) {
	on := !t.hasNum || t.param != 0
	switch t.word {
	case "par":
		if !p.cur.skip && p.cur.pict == nil {
			p.flushPara()
		}
	case "pard":
		p.paraLs, p.paraIlvl, p.paraLi = 0, 0, 0
	case "line":
		if !p.cur.skip {
			p.emit(Run{Break: true})
		}
	case "tab":
		p.text([]byte("\t"))
	case "b":
		p.cur.style.Bold = on
	case "i":
		p.cur.style.Italic = on
	case "ul", "uld", "uldb", "ulw", "ulth":
		p.cur.style.Underline = on
	case "ulnone":
		p.cur.style.Underline = false
	case "strike", "striked":
		p.cur.style.Strike = on
	case "super":
		p.cur.style.Super, p.cur.style.Sub = true, false
	case "sub":
		p.cur.style.Sub, p.cur.style.Super = true, false
	case "nosupersub":
		p.cur.style.Super, p.cur.style.Sub = false, false
	case "plain":
		p.cur.style = Style{}
	case "uc":
		p.cur.uc = t.param
	case "u":
		p.skipChars = p.cur.uc
		n := t.param
		if n < 0 {
			n += 65536
		}
		p.text([]byte(string(rune(n))))
	case "emdash":
		p.text([]byte("—"))
	case "endash":
		p.text([]byte("–"))
	case "lquote":
		p.text([]byte("‘"))
	case "rquote":
		p.text([]byte("’"))
	case "ldblquote":
		p.text([]byte("“"))
	case "rdblquote":
		p.text([]byte("”"))
	case "bullet":
		p.text([]byte("•"))
	case "ls":
		p.paraLs = t.param
	case "ilvl":
		p.paraIlvl = t.param
	case "li":
		p.paraLi = t.param
	case "listtext":
		p.cur.listtext = &strings.Builder{}
		p.cur.skip = true
	case "field":
		p.cur.field = &fieldCtx{}
	case "fldrslt":
		p.cur.inLink = true
	case "pict":
		p.cur.pict = &pictCtx{}
	case "pngblip":
		if p.cur.pict != nil {
			p.cur.pict.ext = "png"
		}
	case "jpegblip":
		if p.cur.pict != nil {
			p.cur.pict.ext = "jpg"
		}
	case "footnote", "chftn":
		// RTF-native footnotes are not produced by Scrivener; skip the group body.
		if t.word == "footnote" {
			p.cur.skip = true
		}
	default:
		if skipDests[t.word] {
			p.cur.skip = true
		}
	}
}
