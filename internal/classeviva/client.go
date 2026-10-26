package classeviva

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	apiBaseURL   = "https://web.spaggiari.eu/rest/v1"
	apiDeviceKey = "Tg1NWEwNGIgIC0K"
	userAgent    = "CVVS/std/4.1.7 Android/10"
	maxBodyBytes = 16 << 20
)

// Client is a small, read-only ClasseViva API client. Credentials are loaded
// from the user's private config directory and never logged or returned.
type Client struct {
	uid      string
	password string
	student  string
	http     *http.Client

	tokenMu sync.Mutex
	token   string
}

// CredentialsPath returns the per-user credentials path. The environment
// override is useful for people who already keep their file elsewhere.
func CredentialsPath() (string, error) {
	if p := strings.TrimSpace(os.Getenv("CLASSEVIVA_CREDENTIALS_FILE")); p != "" {
		return filepath.Clean(p), nil
	}
	root, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate the user configuration directory: %w", err)
	}
	return filepath.Join(root, "classeviva", "credentials.txt"), nil
}

// NewFromUserConfig loads the credentials for the current OS account.
func NewFromUserConfig() (*Client, error) {
	path, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	return NewFromFile(path)
}

// NewFromFile expects the ClasseViva UID on line one and password on line two.
func NewFromFile(path string) (*Client, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("credentials file not found at %s; create it with the ClasseViva UID on line 1 and password on line 2", path)
		}
		return nil, fmt.Errorf("cannot read credentials file metadata: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("credentials path %s is not a regular file", path)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("credentials file permissions are too broad; run chmod 600 %q", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read credentials file: %w", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) < 2 {
		return nil, errors.New("credentials file must contain the ClasseViva UID on line 1 and password on line 2")
	}
	uid := strings.TrimSpace(lines[0])
	password := strings.TrimSuffix(lines[1], "\r")
	if uid == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("credentials file must contain a non-empty UID and password")
	}
	if len(uid) < 2 {
		return nil, errors.New("the ClasseViva UID is too short to identify a student")
	}

	return &Client{
		uid:      uid,
		password: password,
		student:  studentIDFromUID(uid),
		http:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// studentIDFromUID converts the login UID to the numeric student ID expected
// by student endpoints. Some ClasseViva UIDs use S + eight digits + a final
// letter; that final letter is part of the login UID, not the student ID.
func studentIDFromUID(uid string) string {
	if len(uid) == 10 && strings.EqualFold(uid[:1], "S") && isASCIIDigits(uid[1:9]) && isASCIILetter(uid[9]) {
		return uid[1:9]
	}
	if len(uid) > 1 {
		return uid[1:]
	}
	return ""
}

func isASCIIDigits(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func isASCIILetter(value byte) bool {
	return value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

// GetProfile returns the student's card data.
func (c *Client) GetProfile(ctx context.Context) (any, error) {
	return c.getJSON(ctx, "/students/"+url.PathEscape(c.student)+"/card")
}

// GetAgenda returns agenda entries for an inclusive date range.
func (c *Client) GetAgenda(ctx context.Context, from, to time.Time) (any, error) {
	if from.After(to) {
		return nil, errors.New("from must be on or before to")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return nil, errors.New("date range cannot exceed 366 days")
	}
	path := "/students/" + url.PathEscape(c.student) + "/agenda/all/" + from.Format("20060102") + "/" + to.Format("20060102")
	return c.getJSON(ctx, path)
}

// GetLessons returns today's lessons, lessons for one date, or lessons in an
// inclusive date range. If no dates are supplied, it uses ClasseViva's today
// endpoint.
func (c *Client) GetLessons(ctx context.Context, date, from, to *time.Time) (any, error) {
	if date != nil {
		if from != nil || to != nil {
			return nil, errors.New("date cannot be combined with from and to")
		}
		path := "/students/" + url.PathEscape(c.student) + "/lessons/" + date.Format("20060102")
		return c.getJSON(ctx, path)
	}

	if from == nil && to == nil {
		return c.getJSON(ctx, "/students/"+url.PathEscape(c.student)+"/lessons/today")
	}
	if from == nil || to == nil {
		return nil, errors.New("from and to must be provided together")
	}
	if from.After(*to) {
		return nil, errors.New("from must be on or before to")
	}
	if to.Sub(*from) > 366*24*time.Hour {
		return nil, errors.New("date range cannot exceed 366 days")
	}

	path := "/students/" + url.PathEscape(c.student) + "/lessons/" + from.Format("20060102") + "/" + to.Format("20060102")
	return c.getJSON(ctx, path)
}

// GetAbsences returns all absence records, records from a start date, or
// records in an inclusive date range.
func (c *Client) GetAbsences(ctx context.Context, from, to *time.Time) (any, error) {
	if from == nil {
		if to != nil {
			return nil, errors.New("from is required when to is provided")
		}
		return c.getJSON(ctx, "/students/"+url.PathEscape(c.student)+"/absences/details")
	}
	if to == nil {
		return c.getJSON(ctx, "/students/"+url.PathEscape(c.student)+"/absences/details/"+from.Format("20060102"))
	}
	if from.After(*to) {
		return nil, errors.New("from must be on or before to")
	}
	if to.Sub(*from) > 366*24*time.Hour {
		return nil, errors.New("date range cannot exceed 366 days")
	}

	path := "/students/" + url.PathEscape(c.student) + "/absences/details/" + from.Format("20060102") + "/" + to.Format("20060102")
	return c.getJSON(ctx, path)
}

// GetNoticeboard returns current notice-board items without marking them read.
func (c *Client) GetNoticeboard(ctx context.Context) (any, error) {
	return c.getJSON(ctx, "/students/"+url.PathEscape(c.student)+"/noticeboard")
}

// GetDidactics returns the list of didactic materials.
func (c *Client) GetDidactics(ctx context.Context) (any, error) {
	return c.getJSON(ctx, "/students/"+url.PathEscape(c.student)+"/didactics")
}

// GetDidacticsItem reads one didactic item. JSON responses are returned as
// structured data; files are returned with their MIME type and bytes.
func (c *Client) GetDidacticsItem(ctx context.Context, contentID string) (any, string, []byte, error) {
	if err := validatePathPart("content id", contentID); err != nil {
		return nil, "", nil, err
	}
	path := "/students/" + url.PathEscape(c.student) + "/didactics/item/" + url.PathEscape(contentID)
	return c.getJSONOrFile(ctx, path)
}

// GetNoticeAttachment reads an attachment for a specific notice-board item.
func (c *Client) GetNoticeAttachment(ctx context.Context, eventCode, publicationID string, attachmentNumber int) (any, string, []byte, error) {
	if err := validatePathPart("event code", eventCode); err != nil {
		return nil, "", nil, err
	}
	if err := validatePathPart("publication id", publicationID); err != nil {
		return nil, "", nil, err
	}
	if attachmentNumber < 1 {
		return nil, "", nil, fmt.Errorf("attachment number must be at least 1")
	}
	path := "/students/" + url.PathEscape(c.student) + "/noticeboard/attach/" + url.PathEscape(eventCode) + "/" + url.PathEscape(publicationID) + fmt.Sprintf("/%d", attachmentNumber)
	return c.getJSONOrFile(ctx, path)
}

func validatePathPart(label, value string) error {
	if strings.TrimSpace(value) == "" || len(value) > 160 {
		return fmt.Errorf("%s is required and must be at most 160 characters", label)
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, path string) (any, error) {
	data, mimeType, body, err := c.getJSONOrFile(ctx, path)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return data, nil
	}
	return nil, fmt.Errorf("ClasseViva returned %s data where JSON was expected (%d bytes)", mimeType, len(body))
}

func (c *Client) getJSONOrFile(ctx context.Context, path string) (any, string, []byte, error) {
	body, mimeType, err := c.authenticatedRequest(ctx, http.MethodGet, apiBaseURL+path, nil)
	if err != nil {
		return nil, "", nil, err
	}
	if json.Valid(body) {
		var value any
		if err := json.Unmarshal(body, &value); err != nil {
			return nil, "", nil, fmt.Errorf("cannot decode ClasseViva JSON: %w", err)
		}
		if obj, ok := value.(map[string]any); ok {
			if _, failed := obj["error"]; failed {
				return nil, "", nil, errors.New("ClasseViva returned an API error")
			}
		}
		return value, mimeType, nil, nil
	}
	return nil, mimeType, body, nil
}

func (c *Client) authenticatedRequest(ctx context.Context, method, endpoint string, payload any) ([]byte, string, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, "", err
	}
	for attempt := 0; attempt < 2; attempt++ {
		body, mimeType, status, err := c.request(ctx, method, endpoint, payload, token)
		if err != nil {
			return nil, "", err
		}
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			if attempt == 0 {
				c.clearToken(token)
				token, err = c.getToken(ctx)
				if err != nil {
					return nil, "", err
				}
				continue
			}
		}
		if status < 200 || status >= 300 {
			return nil, "", fmt.Errorf("ClasseViva request failed with HTTP %d", status)
		}
		return body, mimeType, nil
	}
	return nil, "", errors.New("ClasseViva authentication failed after retry")
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" {
		return c.token, nil
	}
	loginBody, err := json.Marshal(map[string]string{
		"ident": "",
		"pass":  c.password,
		"uid":   c.uid,
	})
	if err != nil {
		return "", fmt.Errorf("cannot prepare ClasseViva login: %w", err)
	}
	body, _, status, err := c.request(ctx, http.MethodPost, apiBaseURL+"/auth/login", json.RawMessage(loginBody), "")
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("ClasseViva login failed with HTTP %d; verify the local credentials file", status)
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Token == "" {
		return "", errors.New("ClasseViva login response did not contain a session token")
	}
	c.token = result.Token
	return c.token, nil
}

func (c *Client) clearToken(expected string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token == expected {
		c.token = ""
	}
}

func (c *Client) request(ctx context.Context, method, endpoint string, payload any, token string) ([]byte, string, int, error) {
	var requestBody io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, "", 0, fmt.Errorf("cannot encode ClasseViva request: %w", err)
		}
		requestBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return nil, "", 0, fmt.Errorf("cannot prepare ClasseViva request: %w", err)
	}
	req.Header.Set("Z-DEV-ApiKey", apiDeviceKey)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, application/octet-stream, */*")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Z-Auth-Token", token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", 0, fmt.Errorf("ClasseViva request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, "", 0, fmt.Errorf("cannot read ClasseViva response: %w", err)
	}
	if len(body) > maxBodyBytes {
		return nil, "", 0, errors.New("ClasseViva response exceeds the 16 MiB safety limit")
	}
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(body)
	}
	if parsed, _, err := mime.ParseMediaType(mimeType); err == nil {
		mimeType = parsed
	}
	return body, mimeType, resp.StatusCode, nil
}
