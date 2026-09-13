package convert

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/longhand-io/scrivener-to-obsidian/internal/scrivx"
)

// Snapshot is one saved version of one document.
type Snapshot struct {
	Item  *scrivx.Item
	Time  time.Time
	Title string
	Path  string // the .rtf file
}

type xmlSnapshots struct {
	Snapshots []struct {
		Title string `xml:"Title"`
		Date  string `xml:"Date"`
	} `xml:"Snapshot"`
}

// ReadSnapshots lists every snapshot in the package, oldest first.
func ReadSnapshots(p *scrivx.Project) ([]Snapshot, error) {
	dir := p.SnapshotsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Snapshot
	for _, e := range entries {
		if !e.IsDir() || !strings.HasSuffix(e.Name(), ".snapshots") {
			continue
		}
		uuid := strings.TrimSuffix(e.Name(), ".snapshots")
		item := p.ByUUID[uuid]
		if item == nil {
			continue
		}
		sdir := filepath.Join(dir, e.Name())
		titles := map[int64]string{}
		if data, err := os.ReadFile(filepath.Join(sdir, "index.xml")); err == nil {
			var x xmlSnapshots
			if xml.Unmarshal(data, &x) == nil {
				for _, s := range x.Snapshots {
					if t, ok := ParseScrivTime(s.Date); ok {
						titles[t.Unix()] = strings.TrimSpace(s.Title)
					}
				}
			}
		}
		files, _ := os.ReadDir(sdir)
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".rtf") {
				continue
			}
			t, err := time.Parse("2006-01-02-15-04-05-0700", strings.TrimSuffix(f.Name(), ".rtf"))
			if err != nil {
				continue
			}
			title := titles[t.Unix()]
			if title == "" {
				title = "Untitled"
			}
			out = append(out, Snapshot{Item: item, Time: t, Title: title, Path: filepath.Join(sdir, f.Name())})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Time.Equal(out[j].Time) {
			return out[i].Item.UUID < out[j].Item.UUID
		}
		return out[i].Time.Before(out[j].Time)
	})
	// Scrivener can leave two files for one instant, written under different zone offsets.
	// Keep the first when the bytes are identical; keep both when they differ.
	var dedup []Snapshot
	for _, s := range out {
		if n := len(dedup); n > 0 && dedup[n-1].Item == s.Item && dedup[n-1].Time.Equal(s.Time) && sameBytes(dedup[n-1].Path, s.Path) {
			continue
		}
		dedup = append(dedup, s)
	}
	return dedup, nil
}

// InTrash reports whether the snapshot belongs to a trashed item.
func (s Snapshot) InTrash() bool { return s.Item.InTrash() }

func sameBytes(a, b string) bool {
	x, err1 := os.ReadFile(a)
	y, err2 := os.ReadFile(b)
	return err1 == nil && err2 == nil && string(x) == string(y)
}

// FileName is the snapshot's name in the files store.
func (s Snapshot) FileName() string {
	return s.Time.UTC().Format("2006-01-02T15-04-05Z") + " " + SafeName(s.Title) + ".md"
}
