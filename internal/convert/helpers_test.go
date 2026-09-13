package convert

import "github.com/longhand-io/scrivener-to-obsidian/internal/scrivx"

func openProject(path string) (*scrivx.Project, error) { return scrivx.Open(path) }
