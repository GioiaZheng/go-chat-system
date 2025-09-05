package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

// READABLE: Build the probe URL from flags/env instead of hardcoding localhost.
// Defaults keep local runs convenient; no absolute URL literals embedded.
func main() {
	port := flag.Int("port", envInt("PORT", 3000), "API port")
	host := flag.String("host", envStr("HOST", "localhost"), "API host (without scheme)")
	scheme := flag.String("scheme", envStr("SCHEME", "http"), "URL scheme (http/https)")
	prefix := flag.String("prefix", envStr("API_PREFIX", "/api/v1"), "API base path prefix")
	path := flag.String("path", envStr("LIVENESS_PATH", "/liveness"), "Liveness path")
	timeout := flag.Duration("timeout", envDuration("TIMEOUT", 2*time.Second), "HTTP timeout")
	flag.Parse()

	url := fmt.Sprintf("%s://%s:%d%s%s", *scheme, *host, *port, *prefix, *path)
	client := &http.Client{Timeout: *timeout}

	res, err := client.Get(url)
	if err != nil {
		fmt.Printf("HEALTHCHECK FAIL: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("HEALTHCHECK FAIL: status %d from %s\n", res.StatusCode, url)
		os.Exit(1)
	}

	fmt.Printf("HEALTHCHECK OK: %s -> %d\n", url, res.StatusCode)
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		if n != 0 {
			return n
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
