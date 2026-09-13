// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 0xSpectra LLC and the Longhand Authors.

// Package gitlog drives the git binary as an argument list. No shell, no library.
package gitlog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Repo is a working tree with git available.
type Repo struct {
	Dir    string
	Author string // "Name <email>" or ""
	Git    string // binary; "git" when empty
}

// ErrNoGit is returned when git is not on PATH.
var ErrNoGit = errors.New("git not found on PATH")

// Ensure returns a Repo for dir, initialising a repository there if dir is not inside one.
func Ensure(dir, author string) (*Repo, error) {
	bin, err := exec.LookPath("git")
	if err != nil {
		return nil, ErrNoGit
	}
	r := &Repo{Dir: dir, Author: author, Git: bin}
	if _, err := r.run(nil, "rev-parse", "--show-toplevel"); err != nil {
		if _, err := r.run(nil, "init", "-q", "-b", "main"); err != nil {
			return nil, fmt.Errorf("git init: %w", err)
		}
	}
	return r, nil
}

func (r *Repo) args(cmd ...string) []string {
	var out []string
	if r.Author != "" {
		name, email := splitAuthor(r.Author)
		out = append(out, "-c", "user.name="+name, "-c", "user.email="+email)
	}
	return append(out, cmd...)
}

func (r *Repo) run(env []string, cmd ...string) (string, error) {
	c := exec.Command(r.Git, r.args(cmd...)...)
	c.Dir = r.Dir
	c.Env = append(os.Environ(), env...)
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	if err := c.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = strings.TrimSpace(out.String())
		}
		return out.String(), fmt.Errorf("git %s: %s", cmd[0], msg)
	}
	return out.String(), nil
}

// Add stages paths relative to Dir.
func (r *Repo) Add(paths ...string) error {
	_, err := r.run(nil, append([]string{"add", "--"}, paths...)...)
	return err
}

// Commit makes a commit with the subject, an optional list of trailer lines, and a date for both author and committer.
func (r *Repo) Commit(subject string, trailers []string, when time.Time) error {
	env := []string{}
	if !when.IsZero() {
		d := when.Format(time.RFC3339)
		env = append(env, "GIT_AUTHOR_DATE="+d, "GIT_COMMITTER_DATE="+d)
	}
	cmd := []string{"commit", "-q", "--no-verify", "-m", subject}
	if len(trailers) > 0 {
		cmd = append(cmd, "-m", strings.Join(trailers, "\n"))
	}
	_, err := r.run(env, cmd...)
	return err
}

// HasChanges reports whether anything is staged.
func (r *Repo) HasChanges() bool {
	_, err := r.run(nil, "diff", "--cached", "--quiet")
	return err != nil
}

func splitAuthor(s string) (string, string) {
	i := strings.Index(s, "<")
	j := strings.LastIndex(s, ">")
	if i < 0 || j < i {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1 : j])
}
