// Package setupui serves a short-lived, loopback-only page for entering
// ClasseViva credentials directly into the user's local configuration.
package setupui

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const sessionLifetime = 10 * time.Minute

var page = template.Must(template.New("setup").Parse(`<!doctype html>
<html lang="it">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Configura ClasseViva</title>
  <style>
    :root { color-scheme: light dark; font-family: system-ui, sans-serif; }
    body { max-width: 34rem; margin: 3rem auto; padding: 0 1.25rem; line-height: 1.5; }
    h1 { font-size: 1.6rem; }
    p { color: color-mix(in srgb, currentColor 78%, transparent); }
    label { display: block; margin: 1.15rem 0 .35rem; font-weight: 600; }
    input { box-sizing: border-box; width: 100%; padding: .7rem; font: inherit; }
    button { margin-top: 1.4rem; padding: .7rem 1rem; font: inherit; cursor: pointer; }
    .note { padding: .8rem; border: 1px solid color-mix(in srgb, currentColor 25%, transparent); border-radius: .5rem; }
  </style>
</head>
<body>
  <h1>Configura ClasseViva</h1>
  <p class="note">Inserisci le credenziali del tuo account. Il modulo le invia al programma locale ClasseViva in esecuzione su questo computer e le salva nel file di configurazione privato. Non vengono inviate nella chat.</p>
  <form method="post" action="{{.Action}}" autocomplete="on">
    <label for="uid">Nome utente / UID ClasseViva</label>
    <input id="uid" name="uid" type="text" autocomplete="username" required minlength="2" maxlength="256" autofocus>
    <label for="password">Password</label>
    <input id="password" name="password" type="password" autocomplete="current-password" required maxlength="4096">
    <button type="submit">Salva credenziali</button>
  </form>
</body>
</html>`))

var resultPage = template.Must(template.New("result").Parse(`<!doctype html>
<html lang="it"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Configurazione ClasseViva</title></head>
<body style="font: 1rem system-ui, sans-serif; max-width: 34rem; margin: 3rem auto; padding: 0 1.25rem; line-height: 1.5">
  <h1>{{.Title}}</h1><p>{{.Message}}</p>
</body></html>`))

// Session is a one-time setup page served only on the IPv4 loopback interface.
type Session struct {
	url       string
	addr      string
	token     string
	server    *http.Server
	done      chan struct{}
	closeOnce sync.Once
	mu        sync.Mutex
	consumed  bool
}

// Start opens a temporary setup page. save receives the values from the local
// form directly; neither value is returned by this package or logged.
func Start(save func(uid, password string) error) (*Session, error) {
	if save == nil {
		return nil, errors.New("credential save callback is required")
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("cannot start the local credential setup page: %w", err)
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		listener.Close()
		return nil, fmt.Errorf("cannot create a private setup link: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(rawToken)
	addr := listener.Addr().String()
	s := &Session{
		url:   "http://" + addr + "/setup/" + token,
		addr:  addr,
		token: token,
		done:  make(chan struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/setup/", s.handle(save))
	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		_ = s.server.Serve(listener)
		s.finish()
	}()
	time.AfterFunc(sessionLifetime, s.Close)
	return s, nil
}

// URL is the temporary local URL to open in the Codex browser panel.
func (s *Session) URL() string { return s.url }

// Done closes when the one-time page expires, succeeds, or is otherwise shut down.
func (s *Session) Done() <-chan struct{} { return s.done }

// Close stops serving the local page.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.done)
		_ = s.server.Close()
	})
}

func (s *Session) finish() {
	s.closeOnce.Do(func() { close(s.done) })
}

func (s *Session) handle(save func(uid, password string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)
		if r.Host != s.addr || !remoteIsLoopback(r.RemoteAddr) {
			http.Error(w, "Richiesta non valida.", http.StatusForbidden)
			return
		}
		if subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(r.URL.Path, "/setup/")), []byte(s.token)) != 1 || r.URL.Path != "/setup/"+s.token {
			http.NotFound(w, r)
			return
		}
		if r.URL.RawQuery != "" {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = page.Execute(w, struct{ Action string }{Action: r.URL.Path})
		case http.MethodPost:
			s.handleSave(w, r, save)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "Metodo non consentito.", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Session) handleSave(w http.ResponseWriter, r *http.Request, save func(uid, password string) error) {
	if !allowedSetupOrigin(r.Header.Get("Origin"), s.addr) {
		http.Error(w, "Origine della richiesta non valida. Riapri la pagina dal link fornito da Codex.", http.StatusForbidden)
		return
	}
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/x-www-form-urlencoded" {
		http.Error(w, "Formato del modulo non valido.", http.StatusUnsupportedMediaType)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Il modulo non è valido o è troppo grande.", http.StatusBadRequest)
		return
	}
	if len(r.Form["uid"]) != 1 || len(r.Form["password"]) != 1 || len(r.Form) != 2 {
		http.Error(w, "Inserisci UID e password nel modulo.", http.StatusBadRequest)
		return
	}
	uid, password := r.Form.Get("uid"), r.Form.Get("password")
	if len(uid) > 256 || len(password) > 4096 || strings.TrimSpace(uid) == "" || strings.TrimSpace(password) == "" {
		http.Error(w, "Controlla UID e password e riprova.", http.StatusBadRequest)
		return
	}
	if strings.ContainsAny(uid, "\r\n\x00") || strings.ContainsAny(password, "\r\n\x00") {
		http.Error(w, "Il modulo contiene caratteri non validi.", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	if s.consumed {
		s.mu.Unlock()
		http.Error(w, "Questo link di configurazione è già stato usato.", http.StatusGone)
		return
	}
	s.consumed = true
	s.mu.Unlock()

	if err := save(uid, password); err != nil {
		s.mu.Lock()
		s.consumed = false
		s.mu.Unlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = resultPage.Execute(w, struct{ Title, Message string }{
			Title:   "Configurazione non salvata",
			Message: "Non è stato possibile salvare il file di configurazione. Dettaglio: " + err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = resultPage.Execute(w, struct{ Title, Message string }{
		Title:   "Credenziali salvate",
		Message: "Le credenziali sono state salvate nel file locale di configurazione. Puoi tornare a Codex.",
	})
	time.AfterFunc(time.Second, s.Close)
}

// allowedSetupOrigin accepts the direct page origin and loopback origins used
// by embedded-browser proxies. Some embedded browsers submit forms with an
// opaque or absent Origin. In those cases the unguessable, one-time token in
// the setup path is the CSRF capability; the handler still requires the exact
// local Host and a loopback client before reaching this check.
func allowedSetupOrigin(origin, addr string) bool {
	if origin == "" || origin == "null" || origin == "http://"+addr {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil || u.Opaque != "" || u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return false
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return false
	}

	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
}

func remoteIsLoopback(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
