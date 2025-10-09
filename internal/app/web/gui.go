// Package web provides functionality for the GoBE application.
package web

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:embedded/guiweb
var guiWebFS embed.FS

// GUIGoBE analyzes GUI-related metrics and provides insights
type GUIGoBE struct {
	guiWebFS *embed.FS
}

// NewGUIGoBE creates a new instance of GUIGoBE
func NewGUIGoBE() *GUIGoBE {
	return &GUIGoBE{
		guiWebFS: &guiWebFS,
	}
}

// GetWebFS returns the embedded filesystem for GUI web assets
func (g *GUIGoBE) GetWebFS() *embed.FS {
	if g == nil {
		return nil
	}
	if g.guiWebFS == nil {
		g.guiWebFS = &guiWebFS
	}
	return g.guiWebFS
}

// GetWebRoot returns the root directory for GUI web assets
func (g *GUIGoBE) GetWebRoot(path string) os.DirEntry {
	if g == nil {
		return nil
	}
	path = g.normalizePath(path)
	embedFS := g.GetWebFS()
	if embedFS == nil {
		return nil
	}
	dirEntries, err := embedFS.ReadDir("embedded/guiweb")
	if err != nil || len(dirEntries) == 0 {
		return nil
	}
	for _, entry := range dirEntries {
		if entry.Name() == path {
			return entry
		}
	}
	return nil
}

// GetWebFile retrieves a specific file from the embedded GUI web assets
func (g *GUIGoBE) GetWebFile(path string) ([]byte, error) {
	if g == nil {
		return nil, os.ErrNotExist
	}
	embedFS := g.GetWebFS()
	if embedFS == nil {
		return nil, os.ErrNotExist
	}
	cleanPath := g.normalizePath(path)
	fullPath := filepath.Join("embedded/guiweb", cleanPath)
	if _, err := embedFS.Open(fullPath); err != nil {
		return nil, os.ErrNotExist
	}
	data, err := embedFS.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ListWebFiles lists all files in the embedded GUI web assets
func (g *GUIGoBE) ListWebFiles() ([]string, error) {
	if g == nil {
		return nil, os.ErrNotExist
	}
	embedFS := g.GetWebFS()
	if embedFS == nil {
		return nil, os.ErrNotExist
	}
	var files []string
	err := fs.WalkDir(*embedFS, "embedded/guiweb", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, strings.TrimPrefix(path, "embedded/guiweb/"))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetFS returns a sub-filesystem for the GUI web assets
func (g *GUIGoBE) GetFS() fs.FS {
	if g == nil {
		return nil
	}
	embedFS := g.GetWebFS()
	if embedFS == nil {
		return nil
	}
	subFS, err := fs.Sub(*embedFS, "embedded/guiweb")
	if err != nil {
		return nil
	}
	return subFS
}

// ReadFile reads a file from the embedded GUI web assets
func (g *GUIGoBE) ReadFile(path string) ([]byte, error) {
	if g == nil {
		return nil, os.ErrNotExist
	}
	fsys := g.GetFS()
	if fsys == nil {
		return nil, os.ErrNotExist
	}
	cleanPath := g.normalizePath(path)
	data, err := fs.ReadFile(fsys, cleanPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// OpenFile opens a file from the embedded GUI web assets
func (g *GUIGoBE) OpenFile(path string) (fs.File, error) {
	if g == nil {
		return nil, os.ErrNotExist
	}
	fsys := g.GetFS()
	if fsys == nil {
		return nil, os.ErrNotExist
	}
	cleanPath := g.normalizePath(path)
	file, err := fsys.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Exists checks if a file exists in the embedded GUI web assets
func (g *GUIGoBE) Exists(path string) bool {
	if g == nil {
		return false
	}
	fsys := g.GetFS()
	if fsys == nil {
		return false
	}
	cleanPath := g.normalizePath(path)
	_, err := fsys.Open(cleanPath)
	return err == nil
}

func (g *GUIGoBE) normalizePath(path string) string {
	if path == "" {
		return path
	}
	clean := strings.TrimSpace(path)
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	return clean
}
