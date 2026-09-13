// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

package convert

import (
	"encoding/xml"
	"os"
)

// Comment is one entry from content.comments.
type Comment struct {
	ID       string
	Footnote bool
	RTF      []byte
}

type xmlComments struct {
	Comments []struct {
		ID       string `xml:"ID,attr"`
		Footnote string `xml:"Footnote,attr"`
		Body     string `xml:",chardata"`
	} `xml:"Comment"`
}

// ReadComments parses a content.comments file. A missing file is not an error.
func ReadComments(path string) (map[string]Comment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]Comment{}, nil
		}
		return nil, err
	}
	var x xmlComments
	if err := xml.Unmarshal(data, &x); err != nil {
		return nil, err
	}
	out := make(map[string]Comment, len(x.Comments))
	for _, c := range x.Comments {
		out[c.ID] = Comment{ID: c.ID, Footnote: c.Footnote == "Yes", RTF: []byte(c.Body)}
	}
	return out, nil
}
