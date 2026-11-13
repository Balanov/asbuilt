// Package sysfs gives every collector a single way to reach the host.
//
// All reads go through a root-relative filesystem so that the exact same
// collectors can run against a real server ("/") and against a recorded
// fixture tree (--root tests/...). Nothing outside this package touches the
// filesystem directly.
package sysfs

import (
	"os"
	"path/filepath"
	"strings"
)

// FS is a read-only view of a filesystem tree rooted at Root.
// An empty Root means the host root.
type FS struct {
	Root string
}

// New returns an FS rooted at root. An empty root means the host root.
func New(root string) *FS {
	return &FS{Root: root}
}

// Host returns the FS that reads the running host.
func Host() *FS { return &FS{} }

// path maps a host-absolute path into the rooted tree.
func (f *FS) path(p string) string {
	p = strings.TrimPrefix(filepath.ToSlash(p), "/")
	if f.Root == "" {
		return "/" + p
	}
	return filepath.Join(f.Root, filepath.FromSlash(p))
}

// ReadFile returns the contents of p, or an error.
func (f *FS) ReadFile(p string) ([]byte, error) {
	return os.ReadFile(f.path(p))
}

// ReadString returns the contents of p as a trimmed string.
func (f *FS) ReadString(p string) (string, error) {
	b, err := f.ReadFile(p)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// Readlink returns the target of the symbolic link p.
func (f *FS) Readlink(p string) (string, error) {
	return os.Readlink(f.path(p))
}

// ReadDir lists the entries of the directory p.
func (f *FS) ReadDir(p string) ([]os.DirEntry, error) {
	return os.ReadDir(f.path(p))
}

// Exists reports whether p exists.
func (f *FS) Exists(p string) bool {
	_, err := os.Stat(f.path(p))
	return err == nil
}

// IsDir reports whether p is a directory.
func (f *FS) IsDir(p string) bool {
	st, err := os.Stat(f.path(p))
	return err == nil && st.IsDir()
}

// Glob returns the host-absolute paths matching pattern.
func (f *FS) Glob(pattern string) ([]string, error) {
	matches, err := filepath.Glob(f.path(pattern))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		rel, err := filepath.Rel(f.path("/"), m)
		if err != nil {
			continue
		}
		out = append(out, "/"+filepath.ToSlash(rel))
	}
	return out, nil
}

// List returns the names of the entries of the directory p, sorted.
// A missing directory yields an empty list, never an error: absent hardware
// is normal on every machine that is not a dual-socket server.
func (f *FS) List(p string) []string {
	entries, err := f.ReadDir(p)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
