package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed all:web/dist
var webFS embed.FS

var (
	distFS    fs.FS
	indexHTML []byte
	agentsMD  []byte
	readmeMD  []byte
)

func init() {
	var err error
	distFS, err = fs.Sub(webFS, "web/dist")
	if err != nil {
		panic(err)
	}
	indexHTML, _ = fs.ReadFile(distFS, "index.html")
	agentsMD, _ = fs.ReadFile(distFS, "AGENTS.md")
	readmeMD, _ = fs.ReadFile(distFS, "README.md")
}

type checkRequest struct {
	Domain  string   `json:"domain,omitempty"`
	Domains []string `json:"domains,omitempty"`
}

type checkResponse struct {
	Results []Result `json:"results"`
}

const maxDomainsPerRequest = 1000

func runServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/check", handleCheck)
	mux.HandleFunc("/api/tlds", handleTLDs)
	mux.HandleFunc("/agents.md", handleAgentsMD)
	mux.HandleFunc("/readme.md", handleReadmeMD)
	mux.Handle("/", http.FileServerFS(distFS))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return srv.ListenAndServe()
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func renderTemplate(w http.ResponseWriter, r *http.Request, data []byte) {
	if data == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	out := strings.ReplaceAll(string(data), "{{BASE_URL}}", baseURL(r))
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write([]byte(out))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func handleAgentsMD(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, agentsMD)
}

func handleReadmeMD(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, readmeMD)
}

func handleTLDs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tlds, err := fetchTLDs(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"count": len(tlds),
		"tlds":  tlds,
	})
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req checkRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	domains := req.Domains
	if req.Domain != "" {
		domains = append(domains, req.Domain)
	}
	if len(domains) == 0 {
		http.Error(w, "provide 'domain' or 'domains'", http.StatusBadRequest)
		return
	}
	if len(domains) > maxDomainsPerRequest {
		http.Error(w, "max 1000 domains per batch", http.StatusBadRequest)
		return
	}

	const concurrency = 5
	sem := make(chan struct{}, concurrency)
	results := make([]Result, len(domains))
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	for i, d := range domains {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, d string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = checkDomainSafe(ctx, d)
		}(i, d)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkResponse{Results: results})
}
