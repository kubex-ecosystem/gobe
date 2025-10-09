// Package web provides functionality for the GoBE application.
package web

import (
	"embed"
	"os"
	"path/filepath"
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
func (g *GUIGoBE) GetWebRoot(path string) *os.DirEntry {
	if g == nil {
		return nil
	}
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
			return &entry
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
	fullPath := filepath.Join("embedded/guiweb", path)
	if _, err := embedFS.Open(fullPath); err != nil {
		return nil, os.ErrNotExist
	}
	data, err := embedFS.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}
