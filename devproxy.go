//go:build ignore

// devproxy.go is a local-only dev server for testing the site against the live
// Sumo MCP API without CORS. It serves the static site and reverse-proxies /mcp
// to https://api.sumo-mcp.com, stripping browser Origin/Sec-Fetch headers so the
// upstream treats the call as same-origin/non-browser.
//
// Usage:
//
//	go run devproxy.go        # then open http://localhost:8000
//
// This file is not part of the deployed site (build tag "ignore").
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	const addr = ":8000"

	target, err := url.Parse("https://api.sumo-mcp.com")
	if err != nil {
		log.Fatal(err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Host = target.Host
			// Make the upstream see a plain server-to-server request so its
			// cross-origin protection allows it.
			r.Out.Header.Del("Origin")
			r.Out.Header.Del("Sec-Fetch-Site")
			r.Out.Header.Del("Sec-Fetch-Mode")
			r.Out.Header.Del("Sec-Fetch-Dest")
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/mcp", proxy)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	log.Printf("serving site + /mcp proxy on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
