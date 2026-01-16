/*
Healthcheck probes the local API liveness endpoint and exits with a non-zero
status when the service is unavailable. It is designed for container health
checks where only the port may vary.

Usage:

healthcheck [flags]

Flags:

-port <1-65535>
Port used to call http://localhost:<port>/liveness

Exit codes:

0   HTTP 200/204 received
>0  connection error or unexpected status code

Related files:
- cmd/webapi/main.go (hosts the /liveness endpoint)
- service/api/liveness.go (handler implementation)
*/
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

// main runs the CLI healthcheck probe against the API server.
func main() {
	port := flag.Int("port", 3000, "HTTP port for healthcheck")

	flag.Parse()

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/liveness", *port))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		_ = res.Body.Close()
		_, _ = fmt.Fprintln(os.Stderr, "Healthcheck request not OK: ", res.Status)
		os.Exit(1)
	}
	_ = res.Body.Close()
	os.Exit(0)
}
