package convert

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	badChars   = regexp.MustCompile(`[\\/:*?"<>|#^\[\]]+`)
	multiSpace = regexp.MustCompile(`\s+`)
	tagBad     = regexp.MustCompile(`[^a-z0-9_/-]+`)
	multiDash  = regexp.MustCompile(`-+`)
)

const maxNameLen = 120

// SafeName turns a Scrivener title into a file or folder name.
func SafeName(title string) string {
	s := badChars.ReplaceAllString(title, "-")
	s = multiSpace.ReplaceAllString(strings.TrimSpace(s), " ")
	s = strings.Trim(s, ". ")
	if len(s) > maxNameLen {
		s = strings.TrimSpace(s[:maxNameLen])
	}
	if s == "" {
		return "Untitled"
	}
	return s
}

// Tag turns a Scrivener keyword into an Obsidian tag.
func Tag(kw string) string {
	s := strings.ToLower(strings.TrimSpace(kw))
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '-'
		}
		return r
	}, s)
	s = tagBad.ReplaceAllString(s, "")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// Prefix formats a 1-based position with the width needed for count siblings.
func Prefix(pos, count int) string {
	width := len(strconv.Itoa(count))
	if width < 2 {
		width = 2
	}
	return leftPad(strconv.Itoa(pos), width)
}

func leftPad(s string, width int) string {
	for len(s) < width {
		s = "0" + s
	}
	return s
}

// ScrivTime parses Scrivener's "2006-01-02 15:04:05 -0700" and returns RFC 3339, or "" if it does not parse.
func ScrivTime(s string) string {
	t, ok := ParseScrivTime(s)
	if !ok {
		return ""
	}
	return t.Format(time.RFC3339)
}

// ParseScrivTime parses the binder's date attributes.
func ParseScrivTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05 -0700", "2006-01-02 15:04:05 Z07:00", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
