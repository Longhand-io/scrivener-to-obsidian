// Package scrivx parses the .scrivx binder XML of a Scrivener 3 project.
package scrivx

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Binder item types Scrivener writes.
const (
	TypeDraftFolder    = "DraftFolder"
	TypeResearchFolder = "ResearchFolder"
	TypeTrashFolder    = "TrashFolder"
	TypeFolder         = "Folder"
	TypeText           = "Text"
	TypePDF            = "PDF"
	TypeImage          = "Image"
	TypeWebArchive     = "WebArchive"
	TypeOther          = "Other"
)

// Item is one binder entry.
type Item struct {
	UUID             string
	Type             string
	Title            string
	Created          string
	Modified         string
	LabelID          string
	StatusID         string
	KeywordIDs       []string
	Custom           map[string]string // field id -> value
	Bookmarks        []string          // target UUIDs
	IncludeInCompile bool
	Parent           *Item
	Children         []*Item
}

// IsContainer reports whether the item is a folder of any kind.
func (it *Item) IsContainer() bool {
	switch it.Type {
	case TypeDraftFolder, TypeResearchFolder, TypeTrashFolder, TypeFolder:
		return true
	}
	return false
}

// IsAsset reports whether the item's content is a file rather than RTF text.
func (it *Item) IsAsset() bool {
	switch it.Type {
	case TypePDF, TypeImage, TypeWebArchive, TypeOther:
		return true
	}
	return false
}

// InTrash reports whether the item or any ancestor is the Trash folder.
func (it *Item) InTrash() bool {
	for n := it; n != nil; n = n.Parent {
		if n.Type == TypeTrashFolder {
			return true
		}
	}
	return false
}

// Depth is the number of ancestors.
func (it *Item) Depth() int {
	d := 0
	for n := it.Parent; n != nil; n = n.Parent {
		d++
	}
	return d
}

// Walk calls fn for the item and every descendant, depth first, in binder order.
func (it *Item) Walk(fn func(*Item)) {
	fn(it)
	for _, c := range it.Children {
		c.Walk(fn)
	}
}

// Project is a parsed binder plus its definitions.
type Project struct {
	Path         string // the .scriv directory
	Name         string
	Creator      string
	Roots        []*Item
	Labels       map[string]string // id -> title
	Statuses     map[string]string
	Keywords     map[string]string
	CustomFields map[string]string // id -> title
	ByUUID       map[string]*Item
}

// DataDir is Files/Data inside the package.
func (p *Project) DataDir() string { return filepath.Join(p.Path, "Files", "Data") }

// SnapshotsDir is Snapshots inside the package.
func (p *Project) SnapshotsDir() string { return filepath.Join(p.Path, "Snapshots") }

// ItemDir is the per-item folder holding content.rtf and friends.
func (p *Project) ItemDir(it *Item) string { return filepath.Join(p.DataDir(), it.UUID) }

// Walk visits every item in binder order.
func (p *Project) Walk(fn func(*Item)) {
	for _, r := range p.Roots {
		r.Walk(fn)
	}
}

// Count returns the number of items, optionally excluding trash.
func (p *Project) Count(includeTrash bool) int {
	n := 0
	p.Walk(func(it *Item) {
		if includeTrash || !it.InTrash() {
			n++
		}
	})
	return n
}

// ---- XML shapes ----

type xmlProject struct {
	Creator string `xml:"Creator,attr"`
	Binder  struct {
		Items []xmlItem `xml:"BinderItem"`
	} `xml:"Binder"`
	LabelSettings struct {
		Labels []xmlNamed `xml:"Labels>Label"`
	} `xml:"LabelSettings"`
	StatusSettings struct {
		Items []xmlNamed `xml:"StatusItems>Status"`
	} `xml:"StatusSettings"`
	Keywords struct {
		Items []xmlKeyword `xml:"Keyword"`
	} `xml:"Keywords"`
	CustomFields []xmlField `xml:"ProjectSettings>CustomMetaDataSettings>MetaDataField"`
}

type xmlNamed struct {
	ID   string `xml:"ID,attr"`
	Text string `xml:",chardata"`
}

type xmlKeyword struct {
	ID       string       `xml:"ID,attr"`
	Title    string       `xml:"Title"`
	Children []xmlKeyword `xml:"Children>Keyword"`
}

type xmlField struct {
	ID    string `xml:"ID,attr"`
	Title string `xml:"Title"`
}

type xmlItem struct {
	UUID     string `xml:"UUID,attr"`
	ID       string `xml:"ID,attr"`
	Type     string `xml:"Type,attr"`
	Created  string `xml:"Created,attr"`
	Modified string `xml:"Modified,attr"`
	Title    string `xml:"Title"`
	MetaData struct {
		LabelID          string `xml:"LabelID"`
		StatusID         string `xml:"StatusID"`
		IncludeInCompile string `xml:"IncludeInCompile"`
		Custom           []struct {
			FieldID string `xml:"FieldID"`
			Value   string `xml:"Value"`
		} `xml:"CustomMetaData>MetaDataItem"`
	} `xml:"MetaData"`
	KeywordIDs []string `xml:"Keywords>KeywordID"`
	Bookmarks  []struct {
		Target string `xml:"BinderUUID,attr"`
	} `xml:"Bookmarks>Bookmark"`
	Children []xmlItem `xml:"Children>BinderItem"`
}

// ---- parsing ----

// ErrNoScrivx is returned when the package holds no .scrivx file.
var ErrNoScrivx = errors.New("no .scrivx file in package")

// Open finds the .scrivx in a .scriv package and parses it.
func Open(scrivPath string) (*Project, error) {
	matches, err := filepath.Glob(filepath.Join(scrivPath, "*.scrivx"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoScrivx, scrivPath)
	}
	sort.Strings(matches)
	data, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, err
	}
	p, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", matches[0], err)
	}
	p.Path = scrivPath
	p.Name = strings.TrimSuffix(filepath.Base(scrivPath), ".scriv")
	return p, nil
}

// Parse parses .scrivx bytes. Path and Name are left empty.
func Parse(data []byte) (*Project, error) {
	var x xmlProject
	if err := xml.Unmarshal(data, &x); err != nil {
		return nil, err
	}
	p := &Project{
		Creator:      x.Creator,
		Labels:       map[string]string{},
		Statuses:     map[string]string{},
		Keywords:     map[string]string{},
		CustomFields: map[string]string{},
		ByUUID:       map[string]*Item{},
	}
	for _, l := range x.LabelSettings.Labels {
		if t := strings.TrimSpace(l.Text); t != "" {
			p.Labels[l.ID] = t
		}
	}
	for _, s := range x.StatusSettings.Items {
		if t := strings.TrimSpace(s.Text); t != "" {
			p.Statuses[s.ID] = t
		}
	}
	var addKW func(ks []xmlKeyword)
	addKW = func(ks []xmlKeyword) {
		for _, k := range ks {
			if t := strings.TrimSpace(k.Title); t != "" {
				p.Keywords[k.ID] = t
			}
			addKW(k.Children)
		}
	}
	addKW(x.Keywords.Items)
	for _, f := range x.CustomFields {
		if t := strings.TrimSpace(f.Title); t != "" {
			p.CustomFields[f.ID] = t
		}
	}
	var build func(xi xmlItem, parent *Item) *Item
	build = func(xi xmlItem, parent *Item) *Item {
		it := &Item{
			UUID:             firstNonEmpty(xi.UUID, xi.ID),
			Type:             firstNonEmpty(xi.Type, TypeText),
			Title:            strings.TrimSpace(xi.Title),
			Created:          xi.Created,
			Modified:         xi.Modified,
			LabelID:          strings.TrimSpace(xi.MetaData.LabelID),
			StatusID:         strings.TrimSpace(xi.MetaData.StatusID),
			IncludeInCompile: strings.EqualFold(strings.TrimSpace(xi.MetaData.IncludeInCompile), "Yes"),
			Parent:           parent,
			Custom:           map[string]string{},
		}
		if it.Title == "" {
			it.Title = "Untitled"
		}
		for _, k := range xi.KeywordIDs {
			if k = strings.TrimSpace(k); k != "" {
				it.KeywordIDs = append(it.KeywordIDs, k)
			}
		}
		for _, c := range xi.MetaData.Custom {
			if id := strings.TrimSpace(c.FieldID); id != "" && strings.TrimSpace(c.Value) != "" {
				it.Custom[id] = strings.TrimSpace(c.Value)
			}
		}
		for _, b := range xi.Bookmarks {
			if b.Target != "" {
				it.Bookmarks = append(it.Bookmarks, b.Target)
			}
		}
		p.ByUUID[it.UUID] = it
		for _, c := range xi.Children {
			it.Children = append(it.Children, build(c, it))
		}
		return it
	}
	for _, xi := range x.Binder.Items {
		p.Roots = append(p.Roots, build(xi, nil))
	}
	if len(p.Roots) == 0 {
		return nil, errors.New("binder has no items")
	}
	return p, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
