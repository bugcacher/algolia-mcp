package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

func main() {
	log.Printf("Starting algolia MCP server...")

	var algoliaAppID, algoliaAPIKey, algoliaIndexName string
	if algoliaAppID = os.Getenv("ALGOLIA_APP_ID"); algoliaAppID == "" {
		log.Fatal("ALGOLIA_APP_ID is required")
	}
	if algoliaAPIKey = os.Getenv("ALGOLIA_API_KEY"); algoliaAPIKey == "" {
		log.Fatal("ALGOLIA_API_KEY is required")
	}
	if algoliaIndexName = os.Getenv("ALGOLIA_INDEX_NAME"); algoliaIndexName == "" {
		log.Fatal("ALGOLIA_INDEX_NAME is required")
	}

	client := search.NewClient(algoliaAppID, algoliaAPIKey)
	index := client.InitIndex(algoliaIndexName)

	log.Printf("Algolia App ID: %q", algoliaAppID)
	log.Printf("Algolia Index Name: %q", algoliaIndexName)

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

		resp, err := index.Search(query)
		if err != nil {
			return nil, fmt.Errorf("could not search: %w", err)
		}

		b, err := json.Marshal(resp)
		if err != nil {
			return nil, fmt.Errorf("could not marshal response: %w", err)
		}

		return mcp.NewToolResultResource("query results", mcp.TextResourceContents{
			MIMEType: "application/json",
			Text:     string(b),
		}), nil
	})

	getObjectTool := mcp.NewTool(
		"get_object",
		mcp.WithDescription("Get an object by its object ID"),
		mcp.WithString(
			"objectID",
			mcp.Description("The object ID to look up"),
			mcp.Required(),
		),
	)

	mcps.AddTool(getObjectTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		objectID, _ := req.Params.Arguments["objectID"].(string)

		var x map[string]any
		if err := index.GetObject(objectID, &x); err != nil {
			return nil, fmt.Errorf("could not get object: %w", err)
		}

		b, err := json.Marshal(x)
		if err != nil {
			return nil, fmt.Errorf("could not marshal response: %w", err)
		}

		return mcp.NewToolResultResource("object", mcp.TextResourceContents{
			MIMEType: "application/json",
			Text:     string(b),
		}), nil
	})

	settingsResource := mcp.NewResource(
		"algolia://settings",
		"Index settings",
		mcp.WithResourceDescription("Get the settings for the Algolia index"),
		mcp.WithMIMEType("application/json"),
	)
	mcps.AddResource(settingsResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		settingsResp, err := index.GetSettings()
		if err != nil {
			return nil, fmt.Errorf("could not get settings: %w", err)
		}

		b, err := json.Marshal(settingsResp)
		if err != nil {
			return nil, fmt.Errorf("could not marshal response: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				MIMEType: "application/json",
				Text:     string(b),
			},
		}, nil
	})

	recordResourceTemplate := mcp.NewResourceTemplate(
		"algolia://records/{objectID}",
		"Lookup a record by object ID",
		mcp.WithTemplateDescription("Get a record from the Algolia index by its object ID"),
	)
	mcps.AddResourceTemplate(recordResourceTemplate, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		objectID, _ := req.Params.Arguments["objectID"].(string)

		var x map[string]any
		if err := index.GetObject(objectID, &x); err != nil {
			return nil, fmt.Errorf("could not get object: %w", err)
		}

		b, err := json.Marshal(x)
		if err != nil {
			return nil, fmt.Errorf("could not marshal response: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				MIMEType: "application/json",
				Text:     string(b),
			},
		}, nil
	})

	if err := server.ServeStdio(mcps); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
