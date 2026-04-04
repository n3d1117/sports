package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sports/internal/buildinfo"
	"sports/internal/mcpserver"
	"sports/internal/provider/sofascore"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("sports-mcp", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	httpAddr := flags.String("http", "", "Serve Streamable HTTP at this address instead of stdio.")
	version := flags.Bool("version", false, "Print version and exit.")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			writeUsage(stdout)
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "sports-mcp: %v\n\n", err)
		writeUsage(stderr)
		return 2
	}

	if *version {
		_, _ = fmt.Fprintln(stdout, buildinfo.Current())
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := sofascoreapi.New("", nil)
	server := mcpserver.New(client)

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		log.Printf("sports-mcp listening at %s", *httpAddr)
		log.Fatal(http.ListenAndServe(*httpAddr, handler))
	}

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}

	return 0
}

func writeUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: sports-mcp [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "MCP server for the sports lookup surface.")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  -h, --help       Show help.")
	_, _ = fmt.Fprintln(w, "      --http       Serve Streamable HTTP at this address instead of stdio.")
	_, _ = fmt.Fprintln(w, "      --version    Print version and exit.")
}
