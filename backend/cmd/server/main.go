package main

import (
	"zipic/internal/cmd"
)

func main() {
	// Get embedded frontend filesystem
	webdistFS, err := EmbeddedWebdist()
	if err != nil {
		// Start server without frontend embedding (headless mode)
		cmd.Execute(nil)
		return
	}

	// Start server with embedded frontend
	cmd.Execute(webdistFS)
}