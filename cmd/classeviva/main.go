package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"classeviva/internal/classeviva"
	"classeviva/internal/setupui"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type service struct {
	mu           sync.Mutex
	client       *classeviva.Client
	setupSession *setupui.Session
}

func (s *service) getClient() (*classeviva.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		return s.client, nil
	}
	c, err := classeviva.NewFromUserConfig()
	if err != nil {
		return nil, fmt.Errorf("ClasseViva non è configurato o il file credenziali non è valido: %w. Usa get_setup_status e begin_credentials_setup per configurarlo nel browser laterale di Codex", err)
	}
	s.client = c
	return s.client, nil
}

func (s *service) beginSetup() (*setupui.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.setupSession != nil {
		select {
		case <-s.setupSession.Done():
			s.setupSession = nil
		default:
			return s.setupSession, nil
		}
	}
	session, err := setupui.Start(func(uid, password string) error {
		if _, err := classeviva.SaveCredentials(uid, password); err != nil {
			return err
		}
		s.mu.Lock()
		s.client = nil
		s.mu.Unlock()
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.setupSession = session
	go func() {
		<-session.Done()
		s.mu.Lock()
		if s.setupSession == session {
			s.setupSession = nil
		}
		s.mu.Unlock()
	}()
	return session, nil
}

type EmptyInput struct{}

type SetupStatusOutput struct {
	Configured bool   `json:"configured" jsonschema:"Whether a readable ClasseViva credentials file is configured."`
	Path       string `json:"path" jsonschema:"Local path to the credentials file; contains no credentials."`
}

type SetupURLInput struct{}

type SetupURLOutput struct {
	URL              string `json:"url" jsonschema:"Short-lived local URL to open in the Codex browser panel."`
	ExpiresInSeconds int    `json:"expires_in_seconds" jsonschema:"Seconds before the local setup page expires."`
}

type AgendaInput struct {
	From string `json:"from" jsonschema:"Start date in ISO format YYYY-MM-DD. Use the same date as to for one day."`
	To   string `json:"to" jsonschema:"End date in ISO format YYYY-MM-DD, inclusive."`
}

type NoticeAttachmentInput struct {
	EventCode     string `json:"event_code" jsonschema:"Event code from a ClasseViva notice-board item."`
	PublicationID string `json:"publication_id" jsonschema:"Publication ID of the notice-board item."`
}

type DidacticsItemInput struct {
	ContentID string `json:"content_id" jsonschema:"Content ID returned by list_didactics."`
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--credentials-path" {
		path, err := classeviva.CredentialsPath()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)
		return
	}

	log.SetOutput(os.Stderr)
	svc := &service{}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "classeviva",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_setup_status",
		Description: "Controlla se le credenziali ClasseViva locali sono configurate. Non restituisce UID o password.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, any, error) {
		configured, path, err := classeviva.CredentialsConfigured()
		if err != nil {
			return nil, nil, err
		}
		return nil, SetupStatusOutput{Configured: configured, Path: path}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "begin_credentials_setup",
		Description: "Avvia per 10 minuti una pagina temporanea solo locale per inserire UID e password ClasseViva. Restituisce un link senza credenziali; aprilo nel browser laterale di Codex.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ SetupURLInput) (*mcp.CallToolResult, any, error) {
		session, err := svc.beginSetup()
		if err != nil {
			return nil, nil, err
		}
		return nil, SetupURLOutput{URL: session.URL(), ExpiresInSeconds: 600}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_profile",
		Description: "Legge la scheda dello studente ClasseViva autenticato. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, any, error) {
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, err := client.GetProfile(ctx)
		return nil, data, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_agenda",
		Description: "Legge compiti, eventi e lezioni dell'agenda ClasseViva per l'intervallo richiesto. Per un giorno singolo, usa la stessa data in from e to. Date ISO YYYY-MM-DD. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input AgendaInput) (*mcp.CallToolResult, any, error) {
		from, err := time.Parse("2006-01-02", input.From)
		if err != nil {
			return nil, nil, errors.New("from deve essere una data ISO valida YYYY-MM-DD")
		}
		to, err := time.Parse("2006-01-02", input.To)
		if err != nil {
			return nil, nil, errors.New("to deve essere una data ISO valida YYYY-MM-DD")
		}
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, err := client.GetAgenda(ctx, from, to)
		return nil, data, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_noticeboard",
		Description: "Elenca gli avvisi della bacheca ClasseViva senza marcarli come letti. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, any, error) {
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, err := client.GetNoticeboard(ctx)
		return nil, data, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_notice_attachment",
		Description: "Recupera i contenuti di un allegato della bacheca. Usa event_code e publication_id dell'avviso. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input NoticeAttachmentInput) (*mcp.CallToolResult, any, error) {
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, mimeType, body, err := client.GetNoticeAttachment(ctx, input.EventCode, input.PublicationID)
		if err != nil {
			return nil, nil, err
		}
		if data != nil {
			return nil, data, nil
		}
		return embeddedFile("allegato-bacheca", mimeType, body), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_didactics",
		Description: "Elenca i materiali didattici ClasseViva. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, any, error) {
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, err := client.GetDidactics(ctx)
		return nil, data, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_didactics_item",
		Description: "Legge un materiale didattico specifico usando il content_id restituito da list_didactics. Può restituire dati strutturati o un file incorporato. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input DidacticsItemInput) (*mcp.CallToolResult, any, error) {
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, mimeType, body, err := client.GetDidacticsItem(ctx, input.ContentID)
		if err != nil {
			return nil, nil, err
		}
		if data != nil {
			return nil, data, nil
		}
		return embeddedFile("materiale-didattico", mimeType, body), nil, nil
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func embeddedFile(label, mimeType string, body []byte) *mcp.CallToolResult {
	if parsed, _, err := mime.ParseMediaType(mimeType); err == nil {
		mimeType = parsed
	}
	if strings.HasPrefix(mimeType, "text/") || mimeType == "application/json" || mimeType == "application/xml" {
		return &mcp.CallToolResult{Content: []mcp.Content{
			&mcp.TextContent{Text: string(body)},
		}}
	}
	uri := "classeviva://" + url.PathEscape(label) + "/attachment"
	metadata, _ := json.Marshal(map[string]any{"mime_type": mimeType, "size_bytes": len(body)})
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.TextContent{Text: string(metadata)},
		&mcp.EmbeddedResource{Resource: &mcp.ResourceContents{
			URI:      uri,
			MIMEType: mimeType,
			Blob:     body,
		}},
	}}
}
