package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

func main() {
	log.Printf("Starting algolia MCP server...")

	if os.Getenv("ALGOLIA_APP_ID") == "" {
		log.Fatal("ALGOLIA_APP_ID is required")
	}
	if os.Getenv("ALGOLIA_API_KEY") == "" {
		log.Fatal("ALGOLIA_API_KEY is required")
	}
	if os.Getenv("ALGOLIA_INDEX_NAME") == "" {
		log.Fatal("ALGOLIA_INDEX_NAME is required")
	}

	client, err := search.NewClient(os.Getenv("ALGOLIA_APP_ID"), os.Getenv("ALGOLIA_API_KEY"))
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	log.Printf("Algolia App ID: %q", os.Getenv("ALGOLIA_APP_ID"))
	log.Printf("Algolia Index Name: %q", os.Getenv("ALGOLIA_INDEX_NAME"))

	mcps := server.NewMCPServer(
		"algolia-mcp",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	runQueryTool := mcp.NewTool(
		"run_query",
		mcp.WithDescription("Run a query against the Algolia search index"),
		mcp.WithString(
			"query",
			mcp.Description("The query to run against the index"),
			mcp.Required(),
		),
	)

	mcps.AddTool(runQueryTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// indexName, _ := req.Params.Arguments["index"].(string)
		query, _ := req.Params.Arguments["query"].(string)

		// What a mess.
		searchResp, err := client.Search(
			client.NewApiSearchRequest(
				search.NewEmptySearchMethodParams().SetRequests(
					[]search.SearchQuery{
						*search.SearchForHitsAsSearchQuery(
							search.NewEmptySearchForHits().SetIndexName(os.Getenv("ALGOLIA_INDEX_NAME")).SetQuery(query),
						),
					},
				),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("could not query (don't panic): %w", err)
		}

		b, err := json.Marshal(searchResp)
		if err != nil {
			return nil, fmt.Errorf("could not marshal response: %w", err)
		}

		return mcp.NewToolResultResource("query results", mcp.TextResourceContents{
			MIMEType: "application/json",
			Text:     string(b),
		}), nil
	})

	if err := server.ServeStdio(mcps); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
