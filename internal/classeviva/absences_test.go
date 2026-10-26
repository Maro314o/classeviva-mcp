package classeviva

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type absenceTestTransport func(*http.Request) (*http.Response, error)

func (f absenceTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetAbsencesUsesSupportedEndpoints(t *testing.T) {
	start := parseAbsenceTestDate(t, "2026-10-26")
	end := parseAbsenceTestDate(t, "2026-11-01")
	tests := []struct {
		name string
		from *time.Time
		to   *time.Time
		want string
	}{
		{
			name: "all absences",
			want: "/rest/v1/students/12345678/absences/details",
		},
		{
			name: "from one date",
			from: &start,
			want: "/rest/v1/students/12345678/absences/details/20261026",
		},
		{
			name: "inclusive date range",
			from: &start,
			to:   &end,
			want: "/rest/v1/students/12345678/absences/details/20261026/20261101",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotPath string
			client := newAbsenceTestClient(func(req *http.Request) {
				if req.Method == http.MethodGet {
					gotPath = req.URL.Path
				}
			})

			if _, err := client.GetAbsences(context.Background(), test.from, test.to); err != nil {
				t.Fatalf("GetAbsences() error = %v", err)
			}
			if gotPath != test.want {
				t.Fatalf("GET path = %q, want %q", gotPath, test.want)
			}
		})
	}
}

func TestGetAbsencesRejectsInvalidRangesBeforeRequest(t *testing.T) {
	start := parseAbsenceTestDate(t, "2026-10-27")
	end := parseAbsenceTestDate(t, "2026-10-26")
	tooFar := parseAbsenceTestDate(t, "2025-10-01")
	farEnd := parseAbsenceTestDate(t, "2026-10-03")
	tests := []struct {
		name string
		from *time.Time
		to   *time.Time
	}{
		{name: "end without start", to: &end},
		{name: "start after end", from: &start, to: &end},
		{name: "range longer than one year", from: &tooFar, to: &farEnd},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var requestCount int
			client := newAbsenceTestClient(func(*http.Request) { requestCount++ })

			if _, err := client.GetAbsences(context.Background(), test.from, test.to); err == nil {
				t.Fatal("GetAbsences() error = nil, want invalid range error")
			}
			if requestCount != 0 {
				t.Fatalf("HTTP request count = %d, want 0", requestCount)
			}
		})
	}
}

func newAbsenceTestClient(record func(*http.Request)) *Client {
	return &Client{
		uid:      "S12345678A",
		password: "test-only",
		student:  "12345678",
		http: &http.Client{Transport: absenceTestTransport(func(req *http.Request) (*http.Response, error) {
			record(req)
			body := `{"absences":[]}`
			if req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/auth/login") {
				body = `{"token":"test-token"}`
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		})},
	}
}

func parseAbsenceTestDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse test date %q: %v", value, err)
	}
	return parsed
}
