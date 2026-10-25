package setupui

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func postSetupForm(t *testing.T, origin string, save func(string, string) error) int {
	t.Helper()
	const addr = "127.0.0.1:43091"
	const token = "test-token"
	session := &Session{addr: addr, token: token, done: make(chan struct{})}
	form := url.Values{"uid": {"test-user"}, "password": {"test-only-value"}}
	req := httptest.NewRequest(http.MethodPost, "http://"+addr+"/setup/"+token, strings.NewReader(form.Encode()))
	req.Header.Set("Origin", origin)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	session.handleSave(resp, req, save)
	return resp.Code
}

func TestSetupFormAcceptsOpaqueCodexBrowserOrigin(t *testing.T) {
	saved := false
	if status := postSetupForm(t, "null", func(string, string) error {
		saved = true
		return nil
	}); status != http.StatusOK {
		t.Fatalf("POST with opaque browser origin returned status %d, want %d", status, http.StatusOK)
	}
	if !saved {
		t.Fatal("credential callback was not called")
	}
}

func TestSetupFormAcceptsMissingBrowserOrigin(t *testing.T) {
	saved := false
	if status := postSetupForm(t, "", func(string, string) error {
		saved = true
		return nil
	}); status != http.StatusOK {
		t.Fatalf("POST without an origin header returned status %d, want %d", status, http.StatusOK)
	}
	if !saved {
		t.Fatal("credential callback was not called")
	}
}

func TestSetupFormAcceptsLoopbackBrowserProxyOrigin(t *testing.T) {
	saved := false
	if status := postSetupForm(t, "http://127.0.0.1:35313", func(string, string) error {
		saved = true
		return nil
	}); status != http.StatusOK {
		t.Fatalf("POST with loopback proxy origin returned status %d, want %d", status, http.StatusOK)
	}
	if !saved {
		t.Fatal("credential callback was not called")
	}
}

func TestSetupFormRejectsExternalOrigin(t *testing.T) {
	saved := false
	if status := postSetupForm(t, "https://attacker.example", func(string, string) error {
		saved = true
		return nil
	}); status != http.StatusForbidden {
		t.Fatalf("POST with external origin returned status %d, want %d", status, http.StatusForbidden)
	}
	if saved {
		t.Fatal("credential callback must not be called for an external origin")
	}
}
