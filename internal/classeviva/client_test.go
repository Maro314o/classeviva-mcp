package classeviva

import (
	"context"
	"net/http"
	"testing"
)

func TestMarkNoticeReadSendsPost(t *testing.T) {
	var method, path string
	client := newAbsenceTestClient(func(req *http.Request) {
		if req.URL.Path != "/rest/v1/auth/login" {
			method, path = req.Method, req.URL.Path
		}
	})
	if err := client.MarkNoticeRead(context.Background(), "CF", "12345"); err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || path != "/rest/v1/students/12345678/noticeboard/read/CF/12345/101" {
		t.Fatalf("got %s %s", method, path)
	}
}

func TestMarkNoticeReadRejectsMissingIdentifiers(t *testing.T) {
	var requests int
	client := newAbsenceTestClient(func(*http.Request) { requests++ })
	if err := client.MarkNoticeRead(context.Background(), "", "12345"); err == nil {
		t.Fatal("expected validation error")
	}
	if requests != 0 {
		t.Fatalf("got %d requests", requests)
	}
}

func TestGetNoticeAttachmentUsesRequestedAttachmentNumber(t *testing.T) {
	tests := []struct {
		name   string
		number int
		want   string
	}{
		{name: "first attachment", number: 1, want: "/rest/v1/students/12345678/noticeboard/attach/CF/40602674/1"},
		{name: "second attachment", number: 2, want: "/rest/v1/students/12345678/noticeboard/attach/CF/40602674/2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotPath string
			client := newAbsenceTestClient(func(req *http.Request) {
				if req.Method == http.MethodGet {
					gotPath = req.URL.Path
				}
			})

			if _, _, err := client.GetNoticeAttachment(context.Background(), "CF", "40602674", test.number); err != nil {
				t.Fatalf("GetNoticeAttachment() error = %v", err)
			}
			if gotPath != test.want {
				t.Fatalf("GET path = %q, want %q", gotPath, test.want)
			}
		})
	}
}

func TestGetNoticeAttachmentRejectsNonPositiveNumberBeforeRequest(t *testing.T) {
	var requestCount int
	client := newAbsenceTestClient(func(*http.Request) { requestCount++ })
	if _, _, err := client.GetNoticeAttachment(context.Background(), "CF", "40602674", 0); err == nil {
		t.Fatal("GetNoticeAttachment() error = nil, want invalid attachment number error")
	}
	if requestCount != 0 {
		t.Fatalf("HTTP request count = %d, want 0", requestCount)
	}
}

func TestGetNoticeAttachmentPreservesJSONFileBytes(t *testing.T) {
	client := newAbsenceTestClient(func(*http.Request) {})
	mimeType, body, err := client.GetNoticeAttachment(context.Background(), "CF", "40602674", 1)
	if err != nil {
		t.Fatalf("GetNoticeAttachment() error = %v", err)
	}
	if mimeType != "application/json" {
		t.Fatalf("MIME type = %q, want application/json", mimeType)
	}
	if string(body) != `{"absences":[]}` {
		t.Fatalf("body = %q, want original JSON file bytes", body)
	}
}

func TestGetDidacticsItemPreservesJSONFileBytes(t *testing.T) {
	client := newAbsenceTestClient(func(*http.Request) {})
	mimeType, body, err := client.GetDidacticsItem(context.Background(), "material-123")
	if err != nil {
		t.Fatalf("GetDidacticsItem() error = %v", err)
	}
	if mimeType != "application/json" {
		t.Fatalf("MIME type = %q, want application/json", mimeType)
	}
	if string(body) != `{"absences":[]}` {
		t.Fatalf("body = %q, want original JSON file bytes", body)
	}
}

func TestStudentIDFromUID(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want string
	}{
		{
			name: "UID with final letter",
			uid:  "S12345678A",
			want: "12345678",
		},
		{
			name: "numeric suffix remains compatible",
			uid:  "S123456789",
			want: "123456789",
		},
		{
			name: "lowercase prefix with final letter",
			uid:  "s12345678a",
			want: "12345678",
		},
		{
			name: "other UID formats retain legacy conversion",
			uid:  "P12AB34",
			want: "12AB34",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := studentIDFromUID(test.uid); got != test.want {
				t.Fatalf("studentIDFromUID(%q) = %q, want %q", test.uid, got, test.want)
			}
		})
	}
}
