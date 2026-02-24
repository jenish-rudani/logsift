// Example: runtime log configuration via HTTP.
//
// Start the server, then adjust logging in real time with curl:
//
//	curl "localhost:9090/log?level=debug"
//	curl "localhost:9090/log?format=json"
//	curl "localhost:9090/log?filter=auth,db&allowEmptyFilter=false"
//	curl "localhost:9090/log?resetFilter=true"
//	curl "localhost:9090/log?sourceFormat=long"
//
// Run: go run ./examples/httpconfig
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jenish-rudani/logsift"
)

func main() {
	logsift.SetLevel("info")
	logsift.SetFormat("forceColor")

	// Mount the config handler
	http.Handle("/log", logsift.Handler())

	// A simple app endpoint
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		logsift.Info("received /ping")
		logsift.DebugFilter("verbose", "detailed ping diagnostics")
		fmt.Fprintln(w, "pong")
	})

	// Background loop to show config changes in real time
	go func() {
		for {
			logsift.Infof("tick — level=%s format=%s", logsift.GetLevel(), logsift.GetFormat())
			logsift.Debug("this only shows when level is debug")
			logsift.DebugFilter("metrics", "filtered debug: only when 'metrics' filter is active")
			time.Sleep(3 * time.Second)
		}
	}()

	addr := ":9090"
	logsift.Infof("starting server on %s", addr)
	logsift.Info("try: curl localhost:9090/log?level=debug")
	logsift.Info("try: curl localhost:9090/log?format=json")
	logsift.Info("try: curl localhost:9090/log?filter=metrics")
	if err := http.ListenAndServe(addr, nil); err != nil {
		logsift.Error("server failed: ", err)
	}
}
