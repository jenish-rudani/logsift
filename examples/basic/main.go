// Example: basic logging with different levels, formats, and source tracking.
//
// Run: go run ./examples/basic
package main

import "github.com/jenish-rudani/logsift"

func main() {
	// --- Log levels ---
	logsift.SetLevel("debug")
	logsift.Info("info level message")
	logsift.Debug("debug level message")
	logsift.Warn("warn level message")
	logsift.Error("error level message")

	// Formatted variants
	logsift.Infof("server listening on port %d", 8080)
	logsift.Debugf("loaded %d items from cache", 42)

	// --- Output formats ---

	// JSON output
	logsift.SetFormat("json")
	logsift.Info("this is JSON formatted")

	// Text without colors (useful for file output)
	logsift.SetFormat("nocolor")
	logsift.Info("this is plain text, no colors")

	// Force colors (useful when piping to a tool that supports ANSI)
	logsift.SetFormat("forceColor")
	logsift.Info("this has forced colors")

	// Back to default text
	logsift.SetFormat("text")
	logsift.Info("back to default text format")

	// --- Source format ---

	// Short source: just filename:line (default)
	logsift.SetSourceFormat("short")
	logsift.Info("short source format")

	// Long source: full file path:line
	logsift.SetSourceFormat("long")
	logsift.Info("long source format — shows full path")

	// --- Query current config ---
	logsift.SetSourceFormat("short")
	logsift.Infof("current level: %s", logsift.GetLevel())
	logsift.Infof("current format: %s", logsift.GetFormat())
	logsift.Infof("current source format: %s", logsift.GetSourceFormat())
	logsift.Infof("debug enabled: %v", logsift.IsDebugEnabled())
}
