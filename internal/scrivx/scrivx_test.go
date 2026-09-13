// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

package scrivx

import "testing"

const fixture = `<?xml version="1.0" encoding="UTF-8"?>
<ScrivenerProject Identifier="P1" Version="2.0" Creator="SCRMAC-3.5.2-17486">
  <Binder>
    <BinderItem UUID="D" Type="DraftFolder" Created="2021-05-22 19:12:38 -0700" Modified="2021-05-22 19:12:38 -0700">
      <Title>Draft</Title>
      <Children>
        <BinderItem UUID="F1" Type="Folder"><Title>Part One</Title>
          <MetaData><LabelID>1</LabelID><IncludeInCompile>Yes</IncludeInCompile></MetaData>
          <Children>
            <BinderItem UUID="T1" Type="Text"><Title> Chapter One </Title>
              <MetaData><LabelID>2</LabelID><StatusID>3</StatusID><IncludeInCompile>Yes</IncludeInCompile>
                <CustomMetaData><MetaDataItem><FieldID>source</FieldID><Value>Grandma</Value></MetaDataItem></CustomMetaData>
              </MetaData>
              <Keywords><KeywordID>10</KeywordID><KeywordID>11</KeywordID></Keywords>
              <Bookmarks><Bookmark BinderUUID="T2" Destination="[Internal Link]">Two</Bookmark></Bookmarks>
            </BinderItem>
            <BinderItem UUID="T2" Type="Text"><Title>Chapter Two</Title><MetaData><IncludeInCompile>No</IncludeInCompile></MetaData></BinderItem>
          </Children>
        </BinderItem>
      </Children>
    </BinderItem>
    <BinderItem UUID="R" Type="ResearchFolder"><Title>Research</Title>
      <Children><BinderItem UUID="P" Type="PDF"><Title>Source</Title></BinderItem></Children>
    </BinderItem>
    <BinderItem UUID="X" Type="TrashFolder"><Title>Trash</Title>
      <Children><BinderItem UUID="T9" Type="Text"><Title>Old</Title></BinderItem></Children>
    </BinderItem>
  </Binder>
  <LabelSettings><Title>Label</Title><Labels><Label ID="-1">No Label</Label><Label ID="1" Color="1 0 0">Red</Label><Label ID="2">Blue</Label></Labels></LabelSettings>
  <StatusSettings><Title>Status</Title><StatusItems><Status ID="-1">No Status</Status><Status ID="3">First Draft</Status></StatusItems></StatusSettings>
  <Keywords><Keyword ID="10"><Title>winter</Title><Children><Keyword ID="11"><Title>vegan</Title></Keyword></Children></Keyword></Keywords>
  <ProjectSettings><CustomMetaDataSettings><MetaDataField ID="source" Type="Text"><Title>Source</Title></MetaDataField></CustomMetaDataSettings></ProjectSettings>
</ScrivenerProject>`

func TestParse(t *testing.T) {
	p, err := Parse([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	if p.Creator != "SCRMAC-3.5.2-17486" {
		t.Errorf("creator %q", p.Creator)
	}
	if got := p.Count(true); got != 8 {
		t.Errorf("count with trash = %d, want 8", got)
	}
	if got := p.Count(false); got != 6 {
		t.Errorf("count without trash = %d, want 6", got)
	}
	if p.Labels["1"] != "Red" || p.Labels["2"] != "Blue" || p.Labels["-1"] != "No Label" {
		t.Errorf("labels %v", p.Labels)
	}
	if p.Statuses["3"] != "First Draft" {
		t.Errorf("statuses %v", p.Statuses)
	}
	if p.Keywords["10"] != "winter" || p.Keywords["11"] != "vegan" {
		t.Errorf("keywords (nested) %v", p.Keywords)
	}
	if p.CustomFields["source"] != "Source" {
		t.Errorf("custom fields %v", p.CustomFields)
	}
	t1 := p.ByUUID["T1"]
	if t1 == nil {
		t.Fatal("T1 missing")
	}
	if t1.Title != "Chapter One" {
		t.Errorf("title not trimmed: %q", t1.Title)
	}
	if t1.LabelID != "2" || t1.StatusID != "3" || !t1.IncludeInCompile {
		t.Errorf("metadata %+v", t1)
	}
	if len(t1.KeywordIDs) != 2 || t1.Custom["source"] != "Grandma" || len(t1.Bookmarks) != 1 || t1.Bookmarks[0] != "T2" {
		t.Errorf("keywords/custom/bookmarks %+v", t1)
	}
	if t1.Parent == nil || t1.Parent.UUID != "F1" || t1.Depth() != 2 {
		t.Errorf("parent/depth wrong: %+v", t1.Parent)
	}
	if p.ByUUID["T2"].IncludeInCompile {
		t.Error("T2 should not be included")
	}
	if !p.ByUUID["T9"].InTrash() || p.ByUUID["T1"].InTrash() {
		t.Error("trash detection wrong")
	}
	if !p.ByUUID["P"].IsAsset() || p.ByUUID["P"].IsContainer() || !p.ByUUID["F1"].IsContainer() {
		t.Error("type predicates wrong")
	}
	var order []string
	p.Walk(func(it *Item) { order = append(order, it.UUID) })
	want := "D F1 T1 T2 R P X T9"
	if got := join(order); got != want {
		t.Errorf("walk order %q, want %q", got, want)
	}
}

func TestParseRejectsEmpty(t *testing.T) {
	if _, err := Parse([]byte(`<ScrivenerProject><Binder/></ScrivenerProject>`)); err == nil {
		t.Error("expected error for empty binder")
	}
}

func join(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}
