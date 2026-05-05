package main

import (
	"testing"
)

func TestResolveServer(t *testing.T) {
	tests := []struct {
		tld  string
		desc string
	}{
		{"is", "ccTLD via IANA referral"},
		{"io", "hardcoded map"},
		{"ai", "hardcoded map"},
		{"dev", "hardcoded map"},
		{"so", "hardcoded map"},
		{"app", "hardcoded map"},
		{"club", "hardcoded map"},
		{"online", "hardcoded map"},
	}

	for _, tt := range tests {
		t.Run(tt.tld, func(t *testing.T) {
			server, err := resolveServer(tt.tld)
			if err != nil {
				t.Fatalf("resolveServer(%q) error: %v", tt.tld, err)
			}
			if server == "" {
				t.Fatalf("resolveServer(%q) returned empty server", tt.tld)
			}
			t.Logf("%s (%s) -> %s", tt.tld, tt.desc, server)
		})
	}
}
