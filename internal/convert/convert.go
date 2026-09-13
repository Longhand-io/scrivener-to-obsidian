// Package convert turns a parsed Scrivener project into a vault that follows the Longhand spec.
package convert

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/longhand-io/scrivener-to-obsidian/internal/gitlog"
	"github.com/longhand-io/scrivener-to-obsidian/internal/rtf"
	"github.com/longhand-io/scrivener-to-obsidian/internal/scrivx"
)

// Options control a conversion.
type Options struct {
	OutDir       string // the vault directory; the project folder is created inside it
	Git          bool   // replay snapshots as commits and commit the import
	Author       string // git identity "Name <email>", optional
	Poetry       bool   // single paragraph breaks become line breaks
	NoPrefix     bool   // order: frontmatter instead of numeric file prefixes
	IncludeTrash bool
	DryRun       bool
	Now          time.Time // import timestamp; zero means time.Now()
}

// Result summarises what was written.
type Result struct {
	ProjectDir  string
	Manifest    Manifest
	Written     int
	Warnings    []string
	Snapshots   int
	GitCommits  int
	Attachments int
}

// Manifest maps every Scrivener item to its output and records counts.
type Manifest struct {
	Project     string         `json:"project"`
	Source      string         `json:"source"`
	Creator     string         `json:"creator,omitempty"`
	Generated   string         `json:"generated"`
	Spec        int            `json:"longhand"`
	Counts      map[string]int `json:"counts"`
	Items       []ManifestItem `json:"items"`
	Warnings    []string       `json:"warnings"`
	Snapshots   []ManifestSnap `json:"snapshots"`
	Attachments []string       `json:"attachments"`
	Skipped     []ManifestItem `json:"skipped"`
	Extra       map[string]any `json:"-"`
}

// ManifestItem is one binder item and where it went.
type ManifestItem struct {
	UUID  string `json:"uuid"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Path  string `json:"path,omitempty"`
	Words int    `json:"words,omitempty"`
}

// ManifestSnap is one replayed snapshot.
type ManifestSnap struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Date  string `json:"date"`
	Path  string `json:"path"`
	Words int    `json:"words"`
}

type planned struct {
	item     *scrivx.Item
	relPath  string // relative to the vault (OutDir)
	dirPath  string // the folder itself, when isFolder
	isFolder bool
	noteBase string // wikilink target name (file name without .md)
	assetExt string
	order    int
}

type doc struct {
	p        *planned
	front    []Field
	body     string
	words    int
	notes    string
	footDefs []string
}

// Run performs the conversion.
func Run(project *scrivx.Project, opts Options) (*Result, error) {
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	res := &Result{}
	projectDir := SafeName(project.Name)
	res.ProjectDir = filepath.Join(opts.OutDir, projectDir)
	plans, byUUID := plan(project, projectDir, opts)

	man := Manifest{Project: project.Name, Source: "scrivener", Creator: project.Creator,
		Generated: opts.Now.Format(time.RFC3339), Spec: 1, Warnings: []string{},
		Counts: map[string]int{"items": 0, "documents": 0, "folders": 0, "assets": 0, "attachments": 0, "snapshots": 0, "skipped_trash": 0}}

	// link resolver shared by every document: wikilink by planned note name
	wikiFor := func(uuid string) (string, bool) {
		if p, ok := byUUID[uuid]; ok && p.noteBase != "" {
			return p.noteBase, true
		}
		return "", false
	}

	docs := map[string]*doc{}
	files := map[string][]byte{} // relPath -> content, written at the end
	var attachments []string

	for _, p := range plans {
		it := p.item
		mi := ManifestItem{UUID: it.UUID, Type: it.Type, Title: it.Title, Path: p.relPath}
		if it.InTrash() && !opts.IncludeTrash {
			if it.Type != scrivx.TypeTrashFolder {
				man.Skipped = append(man.Skipped, mi)
			}
			continue
		}
		if it.Type == scrivx.TypeTrashFolder || it.Type == scrivx.TypeDraftFolder || it.Type == scrivx.TypeResearchFolder {
			// root containers: folder only
			man.Items = append(man.Items, mi)
			man.Counts["folders"]++
			continue
		}
		if it.IsAsset() {
			src := assetSource(project.ItemDir(it))
			if src == "" {
				man.Warnings = append(man.Warnings, fmt.Sprintf("%s: asset item has no content file", it.Title))
				continue
			}
			data, err := os.ReadFile(src)
			if err != nil {
				return nil, err
			}
			files[p.relPath] = data
			man.Counts["assets"]++
			// sidecar note when the asset carries a synopsis or notes
			syn := readText(filepath.Join(project.ItemDir(it), "synopsis.txt"))
			notes := readRTFMarkdown(filepath.Join(project.ItemDir(it), "notes.rtf"), opts, nil, nil)
			if syn != "" || notes != "" {
				side := strings.TrimSuffix(p.relPath, filepath.Ext(p.relPath)) + ".md"
				front := baseFront(project, it, p, opts)
				front = append(front, Field{"synopsis", syn})
				body := "![[" + filepath.Base(p.relPath) + "]]\n"
				files[side] = []byte(Frontmatter(front) + "\n" + body + notesCallout(notes))
				man.Counts["documents"]++
			}
			man.Items = append(man.Items, mi)
			continue
		}
		// text or folder
		d, err := buildDoc(project, it, p, opts, wikiFor, &attachments, files)
		if err != nil {
			return nil, err
		}
		if d != nil {
			docs[it.UUID] = d
			mi.Words = d.words
			mi.Path = p.relPath
			files[p.relPath] = []byte(render(d))
			man.Counts["documents"]++
		} else {
			man.Counts["folders"]++
		}
		man.Items = append(man.Items, mi)
	}
	man.Attachments = attachments
	man.Counts["attachments"] = len(attachments)
	man.Counts["skipped_trash"] = len(man.Skipped)

	// project note
	files[filepath.Join(projectDir, "_Project.md")] = []byte(projectNote(project, opts))

	// snapshots
	snaps, err := ReadSnapshots(project)
	if err != nil {
		return nil, err
	}
	type replay struct {
		snap    Snapshot
		docPath string
		content string
		words   int
	}
	var replays []replay
	for _, s := range snaps {
		if s.Item.InTrash() && !opts.IncludeTrash {
			continue
		}
		d := docs[s.Item.UUID]
		if d == nil {
			man.Warnings = append(man.Warnings, fmt.Sprintf("%s: snapshot for an item that produced no document", s.Item.Title))
			continue
		}
		raw, err := os.ReadFile(s.Path)
		if err != nil {
			return nil, err
		}
		parsed := rtf.Parse(raw)
		body := parsed.Markdown(rtf.Options{Poetry: opts.Poetry, Link: plainLinks(wikiFor), NoUnder: false})
		content := Frontmatter(d.front) + "\n" + body
		snapRel := filepath.Join(projectDir, "_snapshots", s.Item.UUID, s.FileName())
		for n := 2; ; n++ {
			if _, taken := files[snapRel]; !taken {
				break
			}
			snapRel = filepath.Join(projectDir, "_snapshots", s.Item.UUID, strings.TrimSuffix(s.FileName(), ".md")+fmt.Sprintf(" (%d).md", n))
		}
		files[snapRel] = []byte(content)
		man.Snapshots = append(man.Snapshots, ManifestSnap{UUID: s.Item.UUID, Title: s.Title, Date: s.Time.Format(time.RFC3339), Path: snapRel, Words: parsed.WordCount()})
		replays = append(replays, replay{snap: s, docPath: d.p.relPath, content: content, words: parsed.WordCount()})
	}
	man.Counts["snapshots"] = len(replays)
	man.Counts["items"] = len(man.Items)
	res.Snapshots = len(replays)
	res.Manifest = man
	res.Warnings = man.Warnings

	if opts.DryRun {
		return res, nil
	}
	defer func() { res.Manifest = man; res.Warnings = man.Warnings }()

	var repo *gitlog.Repo
	if opts.Git {
		if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
			return nil, err
		}
		repo, err = gitlog.Ensure(opts.OutDir, opts.Author)
		if err != nil {
			return nil, err
		}
		for _, r := range replays {
			full := filepath.Join(opts.OutDir, r.docPath)
			if err := writeFile(full, []byte(r.content)); err != nil {
				return nil, err
			}
			if err := repo.Add(r.docPath); err != nil {
				return nil, err
			}
			if !repo.HasChanges() {
				man.Warnings = append(man.Warnings, fmt.Sprintf("%s: snapshot %q at %s is identical to the previous one; no commit made", r.snap.Item.Title, r.snap.Title, r.snap.Time.Format(time.RFC3339)))
				continue
			}
			trailers := []string{
				"Snapshot-Title: " + r.snap.Title,
				"Snapshot-Doc-Id: " + r.snap.Item.UUID,
				"Snapshot-Source: scrivener",
				fmt.Sprintf("Word-Count: %d", r.words),
			}
			if err := repo.Commit(fmt.Sprintf("snap(%s): %s", filepath.ToSlash(r.docPath), r.snap.Title), trailers, r.snap.Time); err != nil {
				return nil, err
			}
			res.GitCommits++
		}
	}

	// final files
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := writeFile(filepath.Join(opts.OutDir, k), files[k]); err != nil {
			return nil, err
		}
		res.Written++
	}
	// empty folders for containers with no documents
	for _, p := range plans {
		if p.isFolder && !(p.item.InTrash() && !opts.IncludeTrash) {
			if err := os.MkdirAll(filepath.Join(opts.OutDir, p.dirPath), 0o755); err != nil {
				return nil, err
			}
		}
	}
	manPath := filepath.Join(res.ProjectDir, ".scriv2obsidian.json")
	mb, _ := json.MarshalIndent(man, "", "  ")
	if err := writeFile(manPath, append(mb, '\n')); err != nil {
		return nil, err
	}
	res.Written++

	if repo != nil {
		if err := repo.Add(projectDir); err != nil {
			return nil, err
		}
		if repo.HasChanges() {
			trailers := []string{
				"Scrivener-Project: " + project.Name,
				"Scrivener-Creator: " + project.Creator,
				fmt.Sprintf("Documents: %d", man.Counts["documents"]),
				fmt.Sprintf("Snapshots: %d", man.Counts["snapshots"]),
			}
			if err := repo.Commit(fmt.Sprintf("feat(%s): import from Scrivener", Tag(project.Name)), trailers, opts.Now); err != nil {
				return nil, err
			}
			res.GitCommits++
		}
	}
	res.Attachments = len(attachments)
	return res, nil
}

// plan assigns every item a path. Two passes are not needed because names depend only on ancestors.
func plan(project *scrivx.Project, projectDir string, opts Options) ([]*planned, map[string]*planned) {
	var out []*planned
	byUUID := map[string]*planned{}
	var walk func(items []*scrivx.Item, parentRel string, rootLevel bool)
	walk = func(items []*scrivx.Item, parentRel string, rootLevel bool) {
		used := map[string]int{}
		n := len(items)
		for i, it := range items {
			name := SafeName(it.Title)
			if !rootLevel && !opts.NoPrefix {
				name = Prefix(i+1, n) + " " + name
			}
			// collisions among siblings
			key := strings.ToLower(name)
			if c := used[key]; c > 0 {
				name = fmt.Sprintf("%s (%d)", name, c+1)
			}
			used[key]++
			p := &planned{item: it, order: i + 1}
			hasChildren := len(it.Children) > 0
			isFolder := it.IsContainer() || (it.Type == scrivx.TypeText && hasChildren)
			switch {
			case isFolder:
				p.isFolder = true
				p.relPath = filepath.Join(parentRel, name)
				p.dirPath = p.relPath
				p.noteBase = name
			case it.IsAsset():
				ext := assetExt(assetSource(project.ItemDir(it)))
				p.assetExt = ext
				p.relPath = filepath.Join(parentRel, name+ext)
				p.noteBase = name
			default:
				p.relPath = filepath.Join(parentRel, name+".md")
				p.noteBase = name
			}
			out = append(out, p)
			byUUID[it.UUID] = p
			if hasChildren {
				walk(it.Children, p.relPath, false)
			}
		}
	}
	walk(project.Roots, projectDir, true)
	return out, byUUID
}

func hasText(project *scrivx.Project, it *scrivx.Item) bool {
	st, err := os.Stat(filepath.Join(project.ItemDir(it), "content.rtf"))
	return err == nil && st.Size() > 0
}

func assetSource(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasPrefix(n, "content.") {
			continue
		}
		switch filepath.Ext(n) {
		case ".rtf", ".comments", ".styles", ".txt", ".plist":
			continue
		}
		return filepath.Join(dir, n)
	}
	return ""
}

func assetExt(src string) string {
	if src == "" {
		return ".bin"
	}
	return strings.ToLower(filepath.Ext(src))
}

func readText(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func readRTFMarkdown(path string, opts Options, link rtf.LinkResolver, img rtf.ImageWriter) string {
	b, err := os.ReadFile(path)
	if err != nil || len(strings.TrimSpace(string(b))) == 0 {
		return ""
	}
	return rtf.Parse(b).Markdown(rtf.Options{Poetry: opts.Poetry, Link: link, Image: img})
}

func plainLinks(wikiFor func(string) (string, bool)) rtf.LinkResolver {
	return func(url, text string) string {
		switch {
		case strings.HasPrefix(url, "scrivlnk://"):
			if name, ok := wikiFor(strings.TrimPrefix(url, "scrivlnk://")); ok {
				return "[[" + name + "|" + text + "]]"
			}
			return text
		case strings.HasPrefix(url, "scrivcmt://"):
			return text
		default:
			return "[" + text + "](" + url + ")"
		}
	}
}

func baseFront(project *scrivx.Project, it *scrivx.Item, p *planned, opts Options) []Field {
	typ := "text"
	switch {
	case p.isFolder:
		typ = "folder"
	case it.Type == scrivx.TypePDF:
		typ = "pdf"
	case it.Type == scrivx.TypeImage:
		typ = "image"
	case it.Type == scrivx.TypeWebArchive:
		typ = "web"
	case it.IsAsset():
		typ = "other"
	}
	f := []Field{
		{"id", it.UUID},
		{"title", it.Title},
		{"type", typ},
	}
	if opts.NoPrefix && it.Parent != nil {
		f = append(f, Field{"order", p.order})
	}
	f = append(f, Field{"created", ScrivTime(it.Created)}, Field{"modified", ScrivTime(it.Modified)})
	return f
}

func buildDoc(project *scrivx.Project, it *scrivx.Item, p *planned, opts Options, wikiFor func(string) (string, bool), attachments *[]string, files map[string][]byte) (*doc, error) {
	dir := project.ItemDir(it)
	rtfPath := filepath.Join(dir, "content.rtf")
	raw, err := os.ReadFile(rtfPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	syn := readText(filepath.Join(dir, "synopsis.txt"))
	comments, err := ReadComments(filepath.Join(dir, "content.comments"))
	if err != nil {
		return nil, fmt.Errorf("%s: comments: %w", it.Title, err)
	}
	notesMD := readRTFMarkdown(filepath.Join(dir, "notes.rtf"), opts, plainLinks(wikiFor), nil)
	hasBody := len(strings.TrimSpace(string(raw))) > 0
	if p.isFolder && !hasBody && syn == "" && notesMD == "" && it.LabelID == "" && it.StatusID == "" && len(it.KeywordIDs) == 0 && len(it.Custom) == 0 {
		// plain folder, nothing to write
		return nil, nil
	}
	d := &doc{p: p}
	// frontmatter
	front := baseFront(project, it, p, opts)
	front = append(front, Field{"synopsis", syn})
	if lbl := project.Labels[it.LabelID]; lbl != "" && it.LabelID != "-1" {
		front = append(front, Field{"label", lbl})
	}
	if st := project.Statuses[it.StatusID]; st != "" && it.StatusID != "-1" {
		front = append(front, Field{"status", st})
	}
	var tags []string
	for _, k := range it.KeywordIDs {
		if t := Tag(project.Keywords[k]); t != "" {
			tags = append(tags, t)
		}
	}
	front = append(front, Field{"tags", tags})
	if !p.isFolder || hasBody {
		front = append(front, Field{"include", it.IncludeInCompile})
	}
	var bms []string
	for _, b := range it.Bookmarks {
		if name, ok := wikiFor(b); ok {
			bms = append(bms, "[["+name+"]]")
		}
	}
	front = append(front, Field{"bookmarks", bms})
	meta := map[string]string{}
	for id, v := range it.Custom {
		title := project.CustomFields[id]
		if title == "" {
			title = id
		}
		meta[title] = v
	}
	front = append(front, Field{"meta", meta})
	d.front = front

	// body with comments, links, images
	footN := 0
	link := func(url, text string) string {
		switch {
		case strings.HasPrefix(url, "scrivcmt://"):
			c, ok := comments[strings.TrimPrefix(url, "scrivcmt://")]
			if !ok {
				return text
			}
			body := strings.TrimSpace(rtf.Parse(c.RTF).PlainText())
			body = strings.ReplaceAll(body, "\n", " ")
			if c.Footnote {
				footN++
				d.footDefs = append(d.footDefs, fmt.Sprintf("[^%d]: %s", footN, body))
				return fmt.Sprintf("%s[^%d]", text, footN)
			}
			return text + "%% " + body + " %%"
		case strings.HasPrefix(url, "scrivlnk://"):
			if name, ok := wikiFor(strings.TrimPrefix(url, "scrivlnk://")); ok {
				return "[[" + name + "|" + text + "]]"
			}
			return text
		default:
			return "[" + text + "](" + url + ")"
		}
	}
	imgN := 0
	image := func(img *rtf.Image, _ int) string {
		imgN++
		name := fmt.Sprintf("%s-%d.%s", it.UUID, imgN, img.Ext)
		top := strings.SplitN(filepath.ToSlash(p.relPath), "/", 2)[0]
		rel := filepath.Join(top, "_attachments", name)
		files[rel] = img.Data
		*attachments = append(*attachments, rel)
		return "![[" + name + "]]"
	}
	if hasBody {
		parsed := rtf.Parse(raw)
		d.body = parsed.Markdown(rtf.Options{Poetry: opts.Poetry, Link: link, Image: image})
		d.words = parsed.WordCount()
	}
	d.notes = notesMD
	if p.isFolder {
		// folder note lives inside its folder under the same name
		p.relPath = filepath.Join(p.relPath, filepath.Base(p.relPath)+".md")
	}
	return d, nil
}

func render(d *doc) string {
	var b strings.Builder
	b.WriteString(Frontmatter(d.front))
	if d.body != "" {
		b.WriteString("\n" + d.body)
	}
	if len(d.footDefs) > 0 {
		b.WriteString("\n" + strings.Join(d.footDefs, "\n") + "\n")
	}
	b.WriteString(notesCallout(d.notes))
	return b.String()
}

func notesCallout(notes string) string {
	notes = strings.TrimSpace(notes)
	if notes == "" {
		return ""
	}
	lines := strings.Split(notes, "\n")
	var b strings.Builder
	b.WriteString("\n> [!note] Notes\n")
	for _, l := range lines {
		b.WriteString("> " + l + "\n")
	}
	return b.String()
}

func projectNote(project *scrivx.Project, opts Options) string {
	f := []Field{
		{"longhand", 1},
		{"title", project.Name},
		{"source", "scrivener"},
		{"source_creator", project.Creator},
		{"imported", opts.Now.Format(time.RFC3339)},
		{"labels", orderedValues(project.Labels)},
		{"statuses", orderedValues(project.Statuses)},
		{"keywords", orderedValues(project.Keywords)},
	}
	return Frontmatter(f) + "\n# " + project.Name + "\n\nImported from Scrivener with scriv2obsidian. This note marks the project root; the Longhand plugin reads it, and the importer's manifest is beside it in `.scriv2obsidian.json`.\n"
}

func orderedValues(m map[string]string) []string {
	var out []string
	for id, v := range m {
		if id == "-1" {
			continue
		}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
