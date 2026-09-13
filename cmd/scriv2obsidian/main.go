// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

// Command scriv2obsidian converts Scrivener 3 projects into Longhand vaults.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/longhand-io/scrivener-to-obsidian/internal/convert"
	"github.com/longhand-io/scrivener-to-obsidian/internal/scrivx"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "convert":
		os.Exit(runConvert(os.Args[2:]))
	case "inspect":
		os.Exit(runInspect(os.Args[2:]))
	case "version", "-v", "--version":
		fmt.Println("scriv2obsidian", version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `scriv2obsidian: move Scrivener 3 projects into a Longhand vault, history included.

Usage:
  scriv2obsidian inspect PROJECT.scriv [PROJECT.scriv ...]
  scriv2obsidian convert -out DIR [flags] PROJECT.scriv [PROJECT.scriv ...]
  scriv2obsidian version

Flags for convert:
  -out DIR          vault directory to write into (required)
  -git              initialise a repository if needed and replay snapshots as commits
  -poetry           treat single paragraph breaks as line breaks, blank lines as stanza breaks
  -no-prefix        write order: frontmatter instead of numeric file prefixes
  -include-trash    export the Trash folder too
  -author "N <e>"   git identity for the commits (defaults to your git config)
  -dry-run          print the plan and write nothing

The source package is never modified.
`)
}

func runInspect(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "inspect: give at least one .scriv package")
		return 2
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "project\titems\tdocuments\tfolders\tassets\tsnapshots\ttrashed snaps\tlabels\tstatuses\tkeywords\tcustom\ttrash\tcreator")
	code := 0
	for _, path := range args {
		p, err := scrivx.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			code = 1
			continue
		}
		var docs, folders, assets, trash int
		p.Walk(func(it *scrivx.Item) {
			if it.InTrash() {
				trash++
				return
			}
			switch {
			case it.IsContainer():
				folders++
			case it.IsAsset():
				assets++
			default:
				docs++
			}
		})
		snaps, _ := convert.ReadSnapshots(p)
		live, trashed := 0, 0
		for _, s := range snaps {
			if s.InTrash() {
				trashed++
			} else {
				live++
			}
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n", p.Name, p.Count(true), docs, folders, assets, live, trashed, countReal(p.Labels), countReal(p.Statuses), len(p.Keywords), len(p.CustomFields), trash, p.Creator)
	}
	w.Flush()
	return code
}

func countReal(m map[string]string) int {
	n := 0
	for id := range m {
		if id != "-1" {
			n++
		}
	}
	return n
}

func runConvert(args []string) int {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var opts convert.Options
	fs.StringVar(&opts.OutDir, "out", "", "vault directory to write into")
	fs.BoolVar(&opts.Git, "git", false, "replay snapshots as git commits")
	fs.BoolVar(&opts.Poetry, "poetry", false, "single paragraph breaks become line breaks")
	fs.BoolVar(&opts.NoPrefix, "no-prefix", false, "order: frontmatter instead of numeric prefixes")
	fs.BoolVar(&opts.IncludeTrash, "include-trash", false, "export the Trash folder")
	fs.StringVar(&opts.Author, "author", "", "git identity \"Name <email>\"")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print the plan and write nothing")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if opts.OutDir == "" || fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "convert: -out DIR and at least one .scriv package are required")
		return 2
	}
	abs, err := filepath.Abs(opts.OutDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	opts.OutDir = abs
	code := 0
	for _, path := range fs.Args() {
		p, err := scrivx.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			code = 1
			continue
		}
		res, err := convert.Run(p, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", p.Name, err)
			code = 1
			continue
		}
		c := res.Manifest.Counts
		verb := "wrote"
		if opts.DryRun {
			verb = "would write"
		}
		fmt.Printf("%s: %s %d documents, %d folders, %d assets, %d attachments, %d snapshots", p.Name, verb, c["documents"], c["folders"], c["assets"], c["attachments"], c["snapshots"])
		if opts.Git && !opts.DryRun {
			fmt.Printf(", %d commits", res.GitCommits)
		}
		if c["skipped_trash"] > 0 {
			fmt.Printf(", skipped %d in trash", c["skipped_trash"])
		}
		fmt.Printf(" -> %s\n", res.ProjectDir)
		for _, wmsg := range res.Warnings {
			fmt.Printf("  warning: %s\n", wmsg)
		}
	}
	_ = strings.TrimSpace
	return code
}
