package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("ddev-mcp-server " + version)
		os.Exit(0)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "ddev-mcp", Version: version},
		nil,
	)

	tools.RegisterLifecycle(server)
	tools.RegisterInfo(server)
	tools.RegisterExec(server)
	tools.RegisterDatabase(server)
	tools.RegisterComposer(server)
	tools.RegisterLogs(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
