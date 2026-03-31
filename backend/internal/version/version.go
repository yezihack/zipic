package version

// Version info injected at build time via ldflags
var (
	Version   = "v1.0.0"   // Application version
	BuildDate = "unknown"  // Build date
	GitCommit = "unknown"  // Git commit hash
)

// Info returns version information as a map
func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"build_date": BuildDate,
		"git_commit": GitCommit,
	}
}