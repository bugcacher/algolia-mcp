# MCP

This repository contains experimental [Model Context Protocol (or MCP)](https://modelcontextprotocol.io/introduction) servers for interacting with Algolia APIs.

## Installation

First follow the [quick start](https://modelcontextprotocol.io/quickstart/user), which will install Claude Desktop and setup a sample Fileserver MCP server.

## Setup the prototype Algolia MCP server

Requirements:

* Go (https://go.dev/doc/install)

### Clone the repo and build the server

Clone this repo, and then from the repo root:

```shell
$ cd cmd/mcp
$ go build
```
We need to have the full path of the built server:
```shell
$ pwd
/path/to/the/repo/cmd/mcp
```
The full path to the built server is:

```shell
/path/to/the/repo/cmd/mcp/mcp
```

### Update the settings to point to the new server

In Claude desktop edit the settings as per https://modelcontextprotocol.io/quickstart/user#2-add-the-filesystem-mcp-server and this time add the server definition for algolia (using the server path that you found earlier).

```json
{
   "mcpServers": {
      "algolia": {
         "command": "/path/to/the/repo/cmd/mcp/mcp",
         "env": {
            "ALGOLIA_APP_ID": "<APP_ID>",
            "ALGOLIA_INDEX_NAME": "<INDEX_NAME>",
            "ALGOLIA_API_KEY": "<API_KEY>"
         }
      }
   }
}
```

Restart Claude desktop, and you should see a new `"algolia"` tool is available.

## Debugging

You can run the Inspector (see https://modelcontextprotocol.io/docs/tools/inspector) to check the MCP features and run them manually.

From the repo root, setup the environment

```shell
$ export ALGOLIA_APP_ID=""
$ export ALGOLIA_INDEX_NAME=""
$ export ALGOLIA_API_KEY=""
```
Move into the server directory, and rebuild (if necessary):
```shell
$ cd cmd/mcp
$ go build # might already be up-to-date
```
Run the MCP inspector on the server:
```shell
$ npx @modelcontextprotocol/inspector ./mcp
```