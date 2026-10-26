package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
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

type LessonsInput struct {
	Date string `json:"date,omitempty" jsonschema:"Optional single day in ISO format YYYY-MM-DD. Leave empty when using from and to."`
	From string `json:"from,omitempty" jsonschema:"Optional start date in ISO format YYYY-MM-DD. Provide together with to."`
	To   string `json:"to,omitempty" jsonschema:"Optional end date in ISO format YYYY-MM-DD, inclusive. Provide together with from."`
}

type AbsencesInput struct {
	From string `json:"from,omitempty" jsonschema:"Optional start date in ISO format YYYY-MM-DD. Can be used alone or together with to."`
	To   string `json:"to,omitempty" jsonschema:"Optional end date in ISO format YYYY-MM-DD, inclusive. Provide together with from."`
}

type NoticeAttachmentInput struct {
	EventCode       string `json:"event_code" jsonschema:"Event code from a ClasseViva notice-board item."`
	PublicationID   string `json:"publication_id" jsonschema:"Publication ID of the notice-board item."`
	AttachNum       int    `json:"attach_num,omitempty" jsonschema:"Attachment number from the notice attachments[].attachNum metadata. Defaults to 1."`
	DestinationPath string `json:"destination_path" jsonschema:"Absolute local file path including the filename. Existing files are not overwritten."`
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
		Version: "0.1.16",
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
		Name:        "list_lessons",
		Description: "Elenca le lezioni ClasseViva di oggi, di un singolo giorno o di un intervallo inclusivo. Senza date restituisce oggi; usa date per un giorno oppure from e to insieme. Date ISO YYYY-MM-DD. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input LessonsInput) (*mcp.CallToolResult, any, error) {
		date, err := parseOptionalISODate("date", input.Date)
		if err != nil {
			return nil, nil, err
		}
		from, err := parseOptionalISODate("from", input.From)
		if err != nil {
			return nil, nil, err
		}
		to, err := parseOptionalISODate("to", input.To)
		if err != nil {
			return nil, nil, err
		}
		if date != nil && (from != nil || to != nil) {
			return nil, nil, errors.New("date non può essere combinata con from e to")
		}
		if date == nil && (from == nil) != (to == nil) {
			return nil, nil, errors.New("from e to devono essere forniti insieme")
		}
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, err := client.GetLessons(ctx, date, from, to)
		return nil, data, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_absences",
		Description: "Elenca le assenze registrate su ClasseViva. Senza date restituisce tutti i dettagli; con from restituisce i dati da quella data; con from e to restituisce l'intervallo inclusivo. Date ISO YYYY-MM-DD. Strumento di sola lettura.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input AbsencesInput) (*mcp.CallToolResult, any, error) {
		from, err := parseOptionalISODate("from", input.From)
		if err != nil {
			return nil, nil, err
		}
		to, err := parseOptionalISODate("to", input.To)
		if err != nil {
			return nil, nil, err
		}
		if from == nil && to != nil {
			return nil, nil, errors.New("from è obbligatorio quando viene fornito to")
		}
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		data, err := client.GetAbsences(ctx, from, to)
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
		Description: "Scarica l'allegato e lo salva nel percorso assoluto destination_path indicato. Restituisce percorso, MIME type e dimensione; non sovrascrive file esistenti. Usa event_code, publication_id e attach_num da attachments[].attachNum (predefinito 1). Non segna l'avviso come letto. Se restituisce HTTP 404 con identificativi verificati, lo stato non letto può essere una causa: una verifica API diretta vide il codice 152, ma il MCP nasconde il body. Chiedi consenso prima di qualsiasi marcatura separata; questo MCP non espone il relativo tool.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input NoticeAttachmentInput) (*mcp.CallToolResult, any, error) {
		if err := validateDownloadDestination(input.DestinationPath); err != nil {
			return nil, nil, err
		}
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		attachmentNumber := input.AttachNum
		if attachmentNumber == 0 {
			attachmentNumber = 1
		}
		mimeType, body, err := client.GetNoticeAttachment(ctx, input.EventCode, input.PublicationID, attachmentNumber)
		if err != nil {
			return nil, nil, err
		}
		output, err := saveDownloadedFile(input.DestinationPath, mimeType, body)
		if err != nil {
			return nil, nil, err
		}
		return nil, output, nil
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
		Name:        "download_didactics_item",
		Description: "Scarica un materiale didattico e lo salva nel percorso assoluto destination_path indicato. Richiede content_id da list_didactics e destination_path completo di nome file. Restituisce percorso, MIME type e dimensione; non sovrascrive file esistenti.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input DidacticsDownloadInput) (*mcp.CallToolResult, any, error) {
		if err := validateDownloadDestination(input.DestinationPath); err != nil {
			return nil, nil, err
		}
		client, err := svc.getClient()
		if err != nil {
			return nil, nil, err
		}
		mimeType, body, err := client.GetDidacticsItem(ctx, input.ContentID)
		if err != nil {
			return nil, nil, err
		}
		output, err := saveDownloadedFile(input.DestinationPath, mimeType, body)
		if err != nil {
			return nil, nil, err
		}
		return nil, output, nil
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func parseOptionalISODate(field, value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, fmt.Errorf("%s deve essere una data ISO valida YYYY-MM-DD", field)
	}
	return &parsed, nil
}
