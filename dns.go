package main

import (
	"context"
	"errors"
	"net"
)

func hasNSRecords(ctx context.Context, domain string) (bool, error) {
	ns, err := net.DefaultResolver.LookupNS(ctx, domain)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return false, nil
		}
		return false, err
	}
	return len(ns) > 0, nil
}
