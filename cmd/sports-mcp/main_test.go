package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sports/internal/lookups"
	"sports/internal/mcpserver"
	"sports/internal/provider/sofascore"
)

const (
	helperProcessEnv = "SPORTS_TEST_HELPER_PROCESS"
	helperBaseURLEnv = "SPORTS_TEST_BASE_URL"
)

func TestStdioSmoke(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/search/all":
			_, _ = w.Write([]byte(`{"results":[{"entity":{"id":2693,"name":"Fiorentina","sport":{"slug":"football"},"country":{"name":"Italy"},"team":{}},"score":1,"type":"team"}]}`))
		case "/event/13981704":
			_, _ = w.Write([]byte(`{"event":{"id":13981704,"startTimestamp":1773690300,"status":{"type":"notstarted","description":"Not started"},"homeTeam":{"name":"Cremonese"},"awayTeam":{"name":"Fiorentina"},"tournament":{"name":"Serie A","category":{"sport":{"slug":"football"}}}}}`))
		case "/event/13981704/statistics":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)
				return
			}
			_, _ = w.Write([]byte(`{"shots":10}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess", "--")
	cmd.Env = append(os.Environ(),
		helperProcessEnv+"=1",
		helperBaseURLEnv+"="+api.URL,
	)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	var names []string
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	if !slices.Contains(names, "search") || !slices.Contains(names, "event") {
		t.Fatalf("unexpected tool list: %v", names)
	}

	searchResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "search",
		Arguments: mcpserver.SearchInput{Query: "fiorentina", Limit: 1},
	})
	if err != nil {
		t.Fatalf("search tool failed: %v", err)
	}
	if searchResult.IsError {
		t.Fatalf("search tool returned error result: %+v", searchResult)
	}
	search := decodeStructured[lookups.SearchResponse](t, searchResult.StructuredContent)
	if len(search.Results) != 1 || search.Results[0].ID != 2693 {
		t.Fatalf("unexpected search response: %+v", search)
	}

	eventResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "event",
		Arguments: mcpserver.EventInput{EventID: 13981704},
	})
	if err != nil {
		t.Fatalf("event tool failed: %v", err)
	}
	if eventResult.IsError {
		t.Fatalf("event tool returned error result: %+v", eventResult)
	}
	event := decodeStructured[lookups.EventResponse](t, eventResult.StructuredContent)
	if event.EventID != 13981704 || event.Home != "Cremonese" || event.Away != "Fiorentina" {
		t.Fatalf("unexpected event response: %+v", event)
	}
}

func TestHelpOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Usage: sports-mcp [flags]") {
		t.Fatalf("unexpected help output: %q", output)
	}
	if !strings.Contains(output, "--version") {
		t.Fatalf("help output missing version flag: %q", output)
	}
}

func TestVersionOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "dev" {
		t.Fatalf("unexpected version output: %q", got)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv(helperProcessEnv) != "1" {
		return
	}

	baseURL := os.Getenv(helperBaseURLEnv)
	if baseURL == "" {
		t.Fatal("missing helper base URL")
	}

	server := mcpserver.New(sofascoreapi.New(baseURL, nil))
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func decodeStructured[T any](t *testing.T, value any) T {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var decoded T
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	return decoded
}
