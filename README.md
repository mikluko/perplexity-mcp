# perplexity-mcp

A Model Context Protocol (MCP) server providing Perplexity AI tools for web search, reasoning, and deep research.

## Features

Provides 5 Perplexity AI tools via MCP:

- **perplexity_ask** - Quick web search with citations
- **perplexity_reason** - Step-by-step reasoning and problem solving
- **perplexity_research_start** - Start async deep research (returns request_id)
- **perplexity_research_result** - Check async research status
- **perplexity_research_wait** - Blocking wait for async research completion

## Installation

### Kubernetes/Helm

Install via Helm for production deployments:

```bash
# Add the Helm repository (OCI-based)
helm install perplexity-mcp oci://ghcr.io/mikluko/helm-charts/perplexity-mcp \
  --version 0.2.1 \
  --set perplexity.apiKey="pplx-xxxxxxxxxxxx"
```

#### With Basic Authentication

Enable built-in Traefik sidecar for basic auth protection:

```bash
helm install perplexity-mcp oci://ghcr.io/mikluko/helm-charts/perplexity-mcp \
  --set perplexity.apiKey="pplx-xxxxxxxxxxxx" \
  --set auth.enabled=true \
  --set auth.username=admin \
  --set auth.password="your-secure-password"
```

Or use a pre-generated htpasswd entry (recommended for GitOps):

```bash
# Generate htpasswd entry
HTPASSWD=$(docker run --rm httpd:alpine htpasswd -nbB admin your-password)

# Install with htpasswd entry
helm install perplexity-mcp oci://ghcr.io/mikluko/helm-charts/perplexity-mcp \
  --set perplexity.apiKey="pplx-xxxxxxxxxxxx" \
  --set auth.enabled=true \
  --set-string auth.htpasswd="$HTPASSWD"
```

#### With External Secret

Use an existing Kubernetes secret for the API key:

```bash
# Create secret separately
kubectl create secret generic my-perplexity-secret \
  --from-literal=api-key="pplx-xxxxxxxxxxxx"

# Install referencing the secret
helm install perplexity-mcp oci://ghcr.io/mikluko/helm-charts/perplexity-mcp \
  --set perplexity.existingSecret.name=my-perplexity-secret \
  --set perplexity.existingSecret.key=api-key
```

#### With Ingress

Enable ingress for external access:

```bash
helm install perplexity-mcp oci://ghcr.io/mikluko/helm-charts/perplexity-mcp \
  --set perplexity.apiKey="pplx-xxxxxxxxxxxx" \
  --set auth.enabled=true \
  --set auth.password="your-secure-password" \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set ingress.hosts[0].host=perplexity-mcp.example.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix
```

### Binary Installation

Install locally for development or stdio mode:

```bash
go install github.com/mikluko/perplexity-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/mikluko/perplexity-mcp.git
cd perplexity-mcp
go build -o perplexity-mcp .
```

## Configuration

Set your Perplexity API key as an environment variable:

```bash
export PERPLEXITY_API_KEY=pplx-xxxxxxxxxxxx
```

Get your API key from [Perplexity AI](https://www.perplexity.ai/settings/api).

## Usage

### Stdio Mode (Local)

For use with Claude Desktop, Cursor, or other local MCP clients:

```bash
perplexity-mcp -mode stdio
```

### HTTP Mode (Remote)

For remote access over HTTP:

```bash
perplexity-mcp -mode http -listen :8080
```

Default listen address is `:8080`. The server does not include authentication - deploy behind a reverse proxy with appropriate security controls.

## Client Configuration

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "perplexity": {
      "command": "perplexity-mcp",
      "args": ["-mode", "stdio"],
      "env": {
        "PERPLEXITY_API_KEY": "pplx-xxxxxxxxxxxx"
      }
    }
  }
}
```

### Cursor

Add to Cursor Settings → MCP Servers:

```json
{
  "mcpServers": {
    "perplexity": {
      "command": "perplexity-mcp",
      "args": ["-mode", "stdio"],
      "env": {
        "PERPLEXITY_API_KEY": "pplx-xxxxxxxxxxxx"
      }
    }
  }
}
```

### Remote HTTP Server

For accessing a remote HTTP server, configure your MCP client with:

```json
{
  "mcpServers": {
    "perplexity": {
      "url": "http://your-server:8080"
    }
  }
}
```

**Security Note**: When deploying the HTTP server:
- For Kubernetes deployments: Use the built-in Traefik sidecar with `auth.enabled=true`
- For standalone deployments: Protect with network-level controls (firewall, VPN) or reverse proxy authentication
- Always use TLS/HTTPS in production

The binary intentionally does not include built-in authentication to remain lightweight. The Helm chart provides optional basic auth via Traefik sidecar.

## Tools Reference

### perplexity_ask

Quick web search with the Sonar model.

**Input:**
```json
{
  "query": "What is the latest news about AI?"
}
```

**Output:** Answer with citations

### perplexity_reason

Step-by-step reasoning with the Sonar Reasoning model.

**Input:**
```json
{
  "query": "How do I solve this math problem: ..."
}
```

**Output:** Reasoning chain with conclusion

### perplexity_research_start

Start asynchronous deep research.

**Input:**
```json
{
  "query": "Comprehensive analysis of renewable energy trends"
}
```

**Output:** Request ID for tracking

### perplexity_research_result

Check status of async research.

**Input:**
```json
{
  "request_id": "req_xxxxxxxxxxxx"
}
```

**Output:** Status and results (if completed)

### perplexity_research_wait

Wait for async research to complete (blocking, with timeout).

**Input:**
```json
{
  "request_id": "req_xxxxxxxxxxxx",
  "timeout": 300
}
```

**Output:** Research results when completed

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover
```

Test coverage includes:
- Configuration loading and validation
- Server initialization
- Client setup and status handling
- Error handling for missing API keys and invalid modes

### Building

```bash
# Build for current platform
go build -o perplexity-mcp .

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o perplexity-mcp .

# Build for macOS
GOOS=darwin GOARCH=arm64 go build -o perplexity-mcp .
```

### Running Locally

```bash
# Stdio mode (for testing with MCP clients)
export PERPLEXITY_API_KEY=pplx-xxxxxxxxxxxx
./perplexity-mcp -mode stdio

# HTTP mode (for testing HTTP interface)
export PERPLEXITY_API_KEY=pplx-xxxxxxxxxxxx
./perplexity-mcp -mode http -listen localhost:8080
```

## License

[MIT](./LICENSE)
