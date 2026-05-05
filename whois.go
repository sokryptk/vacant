package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

var tldServers = map[string]string{
	"com":    "whois.verisign-grs.com",
	"net":    "whois.verisign-grs.com",
	"org":    "whois.publicinterestregistry.org",
	"io":     "whois.nic.io",
	"co":     "whois.nic.co",
	"ai":     "whois.nic.ai",
	"app":    "whois.nic.google",
	"dev":    "whois.nic.google",
	"info":   "whois.afilias.net",
	"me":     "whois.nic.me",
	"us":     "whois.nic.us",
	"uk":     "whois.nic.uk",
	"de":     "whois.denic.de",
	"xyz":    "whois.nic.xyz",
	"sh":     "whois.nic.sh",
	"so":     "whois.nic.so",
	"tv":     "whois.nic.tv",
	"cc":     "whois.nic.cc",
	"biz":    "whois.nic.biz",
	"club":   "whois.nic.club",
	"online": "whois.nic.online",
}

var availablePatterns = []string{
	"no match for",
	"not found",
	"no entries found",
	"no data found",
	"domain not found",
	"no object found",
	"status: free",
	"status: available",
	"is available for registration",
	"this query returned 0 objects",
	"no whois server is known",
	"object does not exist",
	"%% no entries found",
}

func queryWhois(server, query string) (string, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(server, "43"), 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("dial %s: %w", server, err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	if _, err := fmt.Fprintf(conn, "%s\r\n", query); err != nil {
		return "", fmt.Errorf("write query: %w", err)
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return sb.String(), nil
}

func resolveServer(tld string) (string, error) {
	if s, ok := tldServers[tld]; ok {
		return s, nil
	}
	resp, err := queryWhois("whois.iana.org", tld)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(resp, "\n") {
		before, after, ok := strings.Cut(strings.ToLower(strings.TrimSpace(line)), "refer:")
		if !ok || before != "" {
			continue
		}
		after = strings.TrimSpace(after)
		if after != "" {
			return after, nil
		}
	}
	return "", fmt.Errorf("no whois server found for tld %q", tld)
}

func isAvailable(response string) (bool, string) {
	lower := strings.ToLower(response)
	for _, p := range availablePatterns {
		if strings.Contains(lower, p) {
			return true, p
		}
	}
	return false, ""
}

func extractTLD(domain string) string {
	domain = strings.TrimSuffix(domain, ".")
	i := strings.LastIndex(domain, ".")
	if i < 0 {
		return ""
	}
	return strings.ToLower(domain[i+1:])
}
