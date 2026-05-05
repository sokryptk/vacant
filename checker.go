package main

import (
	"context"
	"fmt"
	"strings"
)

type Result struct {
	Domain    string `json:"domain"`
	Available bool   `json:"available"`
	Method    string `json:"method"`
	Reason    string `json:"reason,omitempty"`
	Error     string `json:"error,omitempty"`
}

func stripURL(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "https://")
	if i := strings.Index(input, "/"); i >= 0 {
		input = input[:i]
	}
	return input
}

func CheckDomain(ctx context.Context, domain string) Result {
	domain = stripURL(domain)
	res := Result{Domain: domain}

	if !looksLikeDomain(domain) {
		res.Error = "invalid domain syntax"
		return res
	}

	tld := extractTLD(domain)
	if tld == "" {
		res.Error = "could not extract TLD"
		return res
	}

	hasNS, err := hasNSRecords(ctx, domain)
	if err != nil {
		res.Error = fmt.Sprintf("dns error: %v", err)
		return res
	}
	if hasNS {
		res.Method = "dns"
		res.Available = false
		res.Reason = "NS records found"
		return res
	}

	server, err := resolveServer(tld)
	if err != nil {
		res.Method = "dns"
		res.Available = true
		res.Reason = fmt.Sprintf("no NS records; whois lookup skipped: %v", err)
		return res
	}

	resp, finalServer, err := queryWhoisRecursive(server, domain, 0)
	if err != nil {
		res.Method = "dns"
		res.Available = true
		res.Reason = fmt.Sprintf("no NS records; whois confirmation failed: %v", err)
		return res
	}

	avail, matched := isAvailable(resp)
	res.Method = "whois"
	res.Available = avail
	if avail {
		res.Reason = fmt.Sprintf("matched %q on %s", matched, finalServer)
	} else {
		res.Reason = fmt.Sprintf("registration record returned by %s", finalServer)
	}
	return res
}

func looksLikeDomain(d string) bool {
	if len(d) > 253 || !strings.Contains(d, ".") {
		return false
	}
	for _, r := range d {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '.':
		default:
			return false
		}
	}
	return true
}
