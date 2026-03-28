package main

import (
	"embed"
	"io/fs"
)

//go:embed webdist
var webdist embed.FS

// WebdistFS returns the embedded webdist filesystem
func WebdistFS() (fs.FS, error) {
	return fs.Sub(webdist, "webdist")
}