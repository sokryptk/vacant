package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const ianaTLDListURL = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"

type tldCache struct {
	mu     sync.RWMutex
	tlds   []string
	updated time.Time
}

var tldStore tldCache

func fetchTLDs(ctx context.Context) ([]string, error) {
	tldStore.mu.RLock()
	if len(tldStore.tlds) > 0 && time.Since(tldStore.updated) < 24*time.Hour {
		out := make([]string, len(tldStore.tlds))
		copy(out, tldStore.tlds)
		tldStore.mu.RUnlock()
		return out, nil
	}
	tldStore.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ianaTLDListURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch IANA TLD list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IANA TLD list returned %d", resp.StatusCode)
	}

	var tlds []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tlds = append(tlds, strings.ToLower(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read IANA TLD list: %w", err)
	}

	tldStore.mu.Lock()
	tldStore.tlds = make([]string, len(tlds))
	copy(tldStore.tlds, tlds)
	tldStore.updated = time.Now()
	tldStore.mu.Unlock()

	return tlds, nil
}
