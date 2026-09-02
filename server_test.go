package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleCheckAcceptsMaximumBatchSize(t *testing.T) {
	domains := make([]string, maxDomainsPerRequest)
	for i := range domains {
		domains[i] = "invalid"
	}

	body, err := json.Marshal(checkRequest{Domains: domains})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/check", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handleCheck(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var response checkResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.Results) != maxDomainsPerRequest {
		t.Fatalf("results = %d, want %d", len(response.Results), maxDomainsPerRequest)
	}
}

func TestHandleCheckRejectsBatchAboveMaximum(t *testing.T) {
	domains := make([]string, maxDomainsPerRequest+1)
	body, err := json.Marshal(checkRequest{Domains: domains})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/check", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handleCheck(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "max 1000 domains per batch") {
		t.Fatalf("unexpected response body: %q", recorder.Body.String())
	}
}
