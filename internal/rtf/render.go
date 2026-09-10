package rtf

import (
	"fmt"
	"regexp"
	"strings"
)

// LinkResolver turns a hyperlink into Markdown. text is already escaped Markdown.
type LinkResolver func(url, text string) string

// ImageWriter persists an inline image and returns the Markdown that embeds it.
type ImageWriter func(img *Image, index int) string

// Options control Markdown rendering.
type Options struct {
	Poetry  bool // single paragraph marks become line breaks; blank paragraphs separate stanzas
	Link    LinkResolver
	Image   ImageWriter
	NoUnder bool // drop underline instead of emitting <u>
}

var (
	reLeadEsc = regexp.MustCompile(`^(\s*)([#>+\-]|\d+[.)])(\s)`)
	reHashEsc = regexp.MustCompile(`^(\s*)(#)`)
)

// Escape makes prose safe as Markdown inline text.
func Escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '*', '_', '`', '[', ']', '~', '=', '<', '>', '|':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '%':
			b.WriteString("%")
		default:
			b.WriteRune(r)
		}
	}
	out := strings.ReplaceAll(b.String(), "%%", "%\\%")
	return out
}

func escapeLineStart(s string) string {
	if m := reLeadEsc.FindStringSubmatchIndex(s); m != nil {
		return s[:m[4]] + "\\" + s[m[4]:]
	}
	if m := reHashEsc.FindStringSubmatchIndex(s); m != nil {
		return s[:m[4]] + "\\" + s[m[4]:]
	}
	return s
}

type styled struct {
	text  string
	style Style
	url   string
}

func sameFmt(a, b Run) bool {
	return a.Style == b.Style && a.URL == b.URL
}

// Markdown renders the document.
func (d *Doc) Markdown(o Options) string {
	paras := d.Paras
	if o.Poetry {
		paras = joinPoetry(paras)
	}
	var out []string
	imgIndex := 0
	for _, p := range paras {
		if p.IsEmpty() && p.ListLevel == 0 {
			continue
		}
		line := renderInline(p.Runs, o, &imgIndex)
		line = strings.TrimRight(line, " \t")
		if p.ListLevel > 0 {
			indent := strings.Repeat("\t", p.ListLevel-1)
			marker := "- "
			if p.ListOrdered {
				marker = "1. "
			}
			// continuation lines of a list item need the same indent
			line = strings.ReplaceAll(line, "\n", "\n"+indent+"  ")
			out = append(out, indent+marker+strings.TrimSpace(line))
			continue
		}
		lines := strings.Split(line, "\n")
		for i := range lines {
			lines[i] = escapeLineStart(lines[i])
		}
		out = append(out, strings.Join(lines, "\n"))
	}
	// list items adjacent to each other stay in one block; everything else gets a blank line
	var b strings.Builder
	for i, s := range out {
		if i > 0 {
			prevList := strings.HasPrefix(strings.TrimLeft(out[i-1], "\t"), "- ") || strings.HasPrefix(strings.TrimLeft(out[i-1], "\t"), "1. ")
			curList := strings.HasPrefix(strings.TrimLeft(s, "\t"), "- ") || strings.HasPrefix(strings.TrimLeft(s, "\t"), "1. ")
			if prevList && curList {
				b.WriteString("\n")
			} else {
				b.WriteString("\n\n")
			}
		}
		b.WriteString(s)
	}
	res := b.String()
	if res == "" {
		return ""
	}
	return res + "\n"
}

func joinPoetry(paras []Para) []Para {
	var out []Para
	var cur *Para
	for _, p := range paras {
		if p.IsEmpty() {
			cur = nil
			continue
		}
		if cur == nil || p.ListLevel > 0 {
			out = append(out, p)
			if p.ListLevel == 0 {
				cur = &out[len(out)-1]
			} else {
				cur = nil
			}
			continue
		}
		cur.Runs = append(cur.Runs, Run{Break: true})
		cur.Runs = append(cur.Runs, p.Runs...)
	}
	return out
}

func renderInline(runs []Run, o Options, imgIndex *int) string {
	// merge adjacent runs with identical formatting
	var merged []Run
	for _, r := range runs {
		if r.Break || r.Image != nil {
			merged = append(merged, r)
			continue
		}
		if n := len(merged); n > 0 && !merged[n-1].Break && merged[n-1].Image == nil && sameFmt(merged[n-1], r) {
			merged[n-1].Text += r.Text
			continue
		}
		merged = append(merged, r)
	}
	var b strings.Builder
	for _, r := range merged {
		switch {
		case r.Break:
			b.WriteString("  \n")
		case r.Image != nil:
			*imgIndex++
			if o.Image != nil {
				b.WriteString(o.Image(r.Image, *imgIndex))
			}
		default:
			b.WriteString(renderRun(r, o))
		}
	}
	return b.String()
}

func renderRun(r Run, o Options) string {
	text := strings.ReplaceAll(r.Text, "\t", " ")
	lead := text[:len(text)-len(strings.TrimLeft(text, " "))]
	trail := text[len(strings.TrimRight(text, " ")):]
	core := strings.TrimSpace(text)
	if core == "" {
		return text
	}
	esc := Escape(core)
	s := r.Style
	open, close := "", ""
	if s.Bold {
		open += "**"
		close = "**" + close
	}
	if s.Italic {
		open += "*"
		close = "*" + close
	}
	if s.Strike {
		open += "~~"
		close = "~~" + close
	}
	if s.Underline && !o.NoUnder {
		open += "<u>"
		close = "</u>" + close
	}
	if s.Super {
		open += "<sup>"
		close = "</sup>" + close
	}
	if s.Sub {
		open += "<sub>"
		close = "</sub>" + close
	}
	body := open + esc + close
	if r.URL != "" {
		if o.Link != nil {
			body = o.Link(r.URL, body)
		} else {
			body = fmt.Sprintf("[%s](%s)", body, r.URL)
		}
	}
	return lead + body + trail
}

// PlainText renders the document without formatting, paragraphs joined by newlines.
func (d *Doc) PlainText() string {
	var parts []string
	for _, p := range d.Paras {
		if p.IsEmpty() {
			continue
		}
		parts = append(parts, strings.TrimSpace(p.Text()))
	}
	return strings.Join(parts, "\n")
}

// WordCount counts whitespace-separated words in the document's plain text.
func (d *Doc) WordCount() int {
	return len(strings.Fields(d.PlainText()))
}
