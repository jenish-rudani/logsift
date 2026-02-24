// Example: filter-based conditional logging.
//
// Filters let you selectively enable debug/info output for specific
// modules or topics without changing the global log level.
//
// Run: go run ./examples/filters
package main

import (
	"fmt"

	"github.com/jenish-rudani/logsift"
)

func main() {
	logsift.SetLevel("debug")
	logsift.SetFormat("forceColor")

	// --- Single filter ---
	fmt.Println("=== No filters active yet ===")
	logsift.DebugFilter("auth", "this won't print — 'auth' filter not active")
	logsift.InfoFilter("db", "this won't print either")

	fmt.Println("\n=== Add 'auth' filter ===")
	logsift.AddFilter("auth")
	logsift.DebugFilter("auth", "token refreshed")          // prints
	logsift.InfoFilter("db", "query executed")               // still won't print
	logsift.DebugFilterf("auth", "user %s logged in", "alice") // prints

	fmt.Println("\n=== Add 'db' filter ===")
	logsift.AddFilter("db")
	logsift.InfoFilter("db", "connection pool ready") // now prints
	logsift.DebugFilter("auth", "session validated")  // still prints

	// --- Multiple filters (match any) ---
	fmt.Println("\n=== Multi-filter: match any ===")
	logsift.DebugFilters([]string{"cache", "auth"}, "prints because 'auth' is active")
	logsift.DebugFilters([]string{"cache", "queue"}, "won't print — neither active")

	// --- Remove a filter ---
	fmt.Println("\n=== Remove 'auth' filter ===")
	logsift.RemoveFilter("auth")
	logsift.DebugFilter("auth", "this won't print anymore")
	logsift.InfoFilter("db", "db filter still active")

	// --- Replace all filters at once ---
	fmt.Println("\n=== Replace all filters with UpdateFilter ===")
	logsift.UpdateFilter(map[string]bool{
		"api":   true,
		"cache": true,
	})
	logsift.InfoFilter("api", "request received")   // prints
	logsift.InfoFilter("cache", "cache hit")         // prints
	logsift.InfoFilter("db", "won't print — removed by UpdateFilter")

	// --- AllowEmptyFilter behavior ---
	fmt.Println("\n=== Empty filter behavior ===")
	logsift.UpdateFilter(map[string]bool{}) // clear all

	logsift.SetAllowEmptyFilter(false) // default
	logsift.DebugFilter("anything", "won't print — empty filter, allowEmpty=false")

	logsift.SetAllowEmptyFilter(true)
	logsift.DebugFilter("anything", "prints — empty filter, allowEmpty=true")

	// --- ParseFilters helper ---
	fmt.Println("\n=== ParseFilters from comma-separated string ===")
	filters := logsift.ParseFilters("auth,db,cache")
	logsift.UpdateFilter(filters)
	logsift.InfoFilter("auth", "parsed from string")
	logsift.InfoFilter("db", "parsed from string")
}
