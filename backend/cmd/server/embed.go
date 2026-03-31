package main

import (
	"embed"
	"io/fs"
)

//go:embed webdist
var webdist embed.FS

// EmbeddedWebdist returns the embedded webdist filesystem
// This is called by the cmd package to serve frontend
func EmbeddedWebdist() (fs.FS, error) {
	return fs.Sub(webdist, "webdist")
}