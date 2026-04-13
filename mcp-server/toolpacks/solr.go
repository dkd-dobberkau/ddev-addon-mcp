package toolpacks

import (
	"context"
	"time"

	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/executor"
	"github.com/dkd-dobberkau/ddev-addon-mcp/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SolrPack struct{}

func (p *SolrPack) Name() string            { return "solr" }
func (p *SolrPack) ServiceRequired() string { return "solr" }

func (p *SolrPack) Register(server *mcp.Server) {
	type SolrStatusInput struct{}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_solr_status", Description: "Get Solr core status."}, func(ctx context.Context, req *mcp.CallToolRequest, input SolrStatusInput) (*mcp.CallToolResult, any, error) {
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "solr", "curl", "-s", "http://localhost:8983/solr/admin/cores?action=STATUS"}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type SolrReloadInput struct {
		Core string `json:"core" jsonschema:"Name of the Solr core to reload"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_solr_reload", Description: "Reload a Solr core."}, func(ctx context.Context, req *mcp.CallToolRequest, input SolrReloadInput) (*mcp.CallToolResult, any, error) {
		url := "http://localhost:8983/solr/admin/cores?action=RELOAD&core=" + input.Core
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "solr", "curl", "-s", url}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})

	type SolrQueryInput struct {
		Core  string `json:"core" jsonschema:"Solr core name"`
		Query string `json:"query" jsonschema:"Solr query string (e.g. '*:*')"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "ddev_solr_query", Description: "Execute a Solr query."}, func(ctx context.Context, req *mcp.CallToolRequest, input SolrQueryInput) (*mcp.CallToolResult, any, error) {
		url := "http://localhost:8983/solr/" + input.Core + "/select?q=" + input.Query + "&wt=json"
		result, err := executor.ExecuteDdev(ctx, []string{"exec", "--service", "solr", "curl", "-s", url}, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
		}
		if result.ExitCode != 0 {
			return tools.ToolError(result), nil, nil
		}
		return tools.ToolSuccess(result), nil, nil
	})
}
