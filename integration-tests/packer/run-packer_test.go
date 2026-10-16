package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestPackerVariable(t *testing.T) {
	schematic, err := packerVariable("hcloud-talos.pkr.hcl", "talos_schematic_id", nil)
	if err != nil || len(schematic) != 64 {
		t.Fatalf("read default schematic: %q, %v", schematic, err)
	}
	version, err := packerVariable("hcloud-talos.pkr.hcl", "talos_version", []string{"talos_version=v1.0.0", "talos_version=v1.14.2"})
	if err != nil || version != "v1.14.2" {
		t.Fatalf("read overridden version: %q, %v", version, err)
	}
}

func TestSnapshotExists(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		want       bool
		wantErr    bool
	}{
		{"found", `{"images":[{"id":123}]}`, http.StatusOK, true, false},
		{"missing", `{"images":[]}`, http.StatusOK, false, false},
		{"api error", `{"error":"unauthorized"}`, http.StatusUnauthorized, false, true},
		{"bad response", `invalid`, http.StatusOK, false, true},
		{"missing images", `{}`, http.StatusOK, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("Authorization") != "Bearer test-token" || req.URL.Query().Get("label_selector") != "os=talos,version=v1.14.2" || req.URL.Query().Get("type") != "snapshot" {
					t.Errorf("unexpected request: %s", req.URL)
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})}
			got, err := snapshotExists(context.Background(), client, "test-token", "os=talos,version=v1.14.2")
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("snapshotExists = %v, %v; want %v, error %v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}
