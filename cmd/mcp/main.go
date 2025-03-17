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

	client, err := search.NewClient(algoliaAppID, algoliaAPIKey)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

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

		// What a mess.
		mps := search.NewEmptySearchMethodParams().
			SetRequests(
				[]search.SearchQuery{
					*search.SearchForHitsAsSearchQuery(
						search.NewEmptySearchForHits().
							SetIndexName(algoliaIndexName).
							SetQuery(query),
					),
				},
			)

		br, err := json.Marshal(mps)
		if err != nil {
			return nil, fmt.Errorf("could not marshal request: %w", err)
		}
		log.Printf("run_query Request: %s", string(br))

		searchResp, err := client.Search(
			client.NewApiSearchRequest(
				mps,
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

	settingsResource := mcp.NewResource(
		"algolia://settings",
		"Index settings",
		mcp.WithResourceDescription("Get the settings for the Algolia index"),
		mcp.WithMIMEType("application/json"),
	)
	mcps.AddResource(settingsResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		settingsResp, err := client.GetSettings(client.NewApiGetSettingsRequest(algoliaIndexName))
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

		// Slightly better than the search endpoing, but still...
		recordResp, err := client.GetObjects(
			client.NewApiGetObjectsRequest(
				search.NewEmptyGetObjectsParams().SetRequests(
					[]search.GetObjectsRequest{
						*search.NewEmptyGetObjectsRequest().
							SetObjectID(objectID).
							SetIndexName(algoliaIndexName),
					},
				),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("could not get record: %w", err)
		}

		b, err := json.Marshal(recordResp)
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
