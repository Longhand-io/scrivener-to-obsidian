// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

// faqgen renders docs/FAQ.md into the FAQ region of the Longhand site page, or checks that it already matches.
//
//	go run ./hack/faqgen -md docs/FAQ.md -html ../longhand-site/faq.html            # rewrite the region
//	go run ./hack/faqgen -md docs/FAQ.md -html ../longhand-site/faq.html -check     # exit 1 on drift
//
// The markdown subset: "## " headings, paragraphs, `code`, [text](url), and a line
// "<!-- figure: name -->" which inlines <figure-dir>/name.html from the site repo.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	startMark = "<!-- faq:start -->"
	endMark   = "<!-- faq:end -->"
)

var (
	reCode   = regexp.MustCompile("`([^`]+)`")
	reLink   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reFigure = regexp.MustCompile(`^<!-- figure: ([a-z0-9-]+) -->$`)
	reSlug   = regexp.MustCompile(`[^a-z0-9]+`)
)

func inline(s string) string {
	s = html.EscapeString(s)
	s = reCode.ReplaceAllString(s, "<code>$1</code>")
	s = reLink.ReplaceAllString(s, `<a href="$2">$1</a>`)
	s = strings.ReplaceAll(s, "&#34;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return s
}

func slug(h string) string {
	s := strings.ToLower(h)
	s = reSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = strings.TrimRight(s[:40], "-")
	}
	return s
}

type entry struct {
	title, id string
	body      []string
}

func render(md string, figureDir string) (string, error) {
	var entries []entry
	var intro []string
	var cur *entry
	var para []string
	flush := func() {
		if len(para) == 0 {
			return
		}
		p := "<p>" + inline(strings.Join(para, " ")) + "</p>"
		if cur == nil {
			intro = append(intro, p)
		} else {
			cur.body = append(cur.body, p)
		}
		para = nil
	}
	for _, line := range strings.Split(md, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "# "):
			continue
		case strings.HasPrefix(t, "## "):
			flush()
			entries = append(entries, entry{title: strings.TrimPrefix(t, "## ")})
			cur = &entries[len(entries)-1]
			cur.id = slug(cur.title)
		case reFigure.MatchString(t):
			flush()
			name := reFigure.FindStringSubmatch(t)[1]
			b, err := os.ReadFile(filepath.Join(figureDir, name+".html"))
			if err != nil {
				return "", fmt.Errorf("figure %q: %w", name, err)
			}
			if cur != nil {
				cur.body = append(cur.body, strings.TrimSpace(string(b)))
			}
		case strings.HasPrefix(t, "<!--"):
			// an ordinary HTML comment in the markdown: notes for editors, not for the page
			flush()
		case t == "":
			flush()
		default:
			para = append(para, t)
		}
	}
	flush()
	var b strings.Builder
	b.WriteString(startMark + "\n")
	for _, p := range intro {
		b.WriteString("  " + strings.Replace(p, "<p>", `<p class="lede">`, 1) + "\n")
	}
	b.WriteString("  <div class=\"faq\">\n  <ul class=\"toc\">\n")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("    <li><a href=\"#%s\">%s</a></li>\n", e.id, inline(e.title)))
	}
	b.WriteString("  </ul>\n")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("\n  <h2 id=\"%s\">%s</h2>\n", e.id, inline(e.title)))
		for _, p := range e.body {
			b.WriteString("  " + p + "\n")
		}
	}
	b.WriteString("  </div>\n" + endMark)
	return b.String(), nil
}

func main() {
	mdPath := flag.String("md", "docs/FAQ.md", "markdown source")
	htmlPath := flag.String("html", "", "site page holding the faq:start/faq:end region")
	figures := flag.String("figures", "", "directory of figure fragments; defaults to <html dir>/figures")
	check := flag.Bool("check", false, "do not write; exit 1 if the page region differs from the rendering")
	flag.Parse()
	if *htmlPath == "" {
		fmt.Fprintln(os.Stderr, "faqgen: -html is required")
		os.Exit(2)
	}
	if *figures == "" {
		*figures = filepath.Join(filepath.Dir(*htmlPath), "figures")
	}
	md, err := os.ReadFile(*mdPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	page, err := os.ReadFile(*htmlPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	region, err := render(string(md), *figures)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	i := bytes.Index(page, []byte(startMark))
	j := bytes.Index(page, []byte(endMark))
	if i < 0 || j < i {
		fmt.Fprintf(os.Stderr, "faqgen: %s has no %s ... %s region\n", *htmlPath, startMark, endMark)
		os.Exit(1)
	}
	current := page[i : j+len(endMark)]
	if *check {
		if string(current) == region {
			fmt.Println("faq: site page matches docs/FAQ.md")
			return
		}
		fmt.Fprintln(os.Stderr, "faq: DRIFT. The site page does not match docs/FAQ.md. Regenerate with: go run ./hack/faqgen -md docs/FAQ.md -html <path to longhand-site/faq.html>")
		os.Exit(1)
	}
	out := append(append(append([]byte{}, page[:i]...), []byte(region)...), page[j+len(endMark):]...)
	if err := os.WriteFile(*htmlPath, out, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("faq: wrote %d entries into %s\n", strings.Count(region, "<h2 "), *htmlPath)
}
