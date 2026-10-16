// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Oakfall Authors.

package convert

import "github.com/oakfall/scrivener-to-obsidian/internal/scrivx"

func openProject(path string) (*scrivx.Project, error) { return scrivx.Open(path) }
