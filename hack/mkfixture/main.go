// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

// Command mkfixture writes the synthetic Scrivener package into a directory,
// so scripts outside go test can run the CLI against it.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/longhand-io/scrivener-to-obsidian/internal/testfixture"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: mkfixture DIR   (writes DIR/Fixture Novel.scriv)")
		os.Exit(2)
	}
	root := filepath.Join(os.Args[1], testfixture.Name)
	if err := testfixture.Write(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(root)
}
