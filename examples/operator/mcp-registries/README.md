# MCPRegistry Examples

This directory contains examples of how to use the MCPRegistry custom resource to manage MCP server registries in Kubernetes.

## Overview

MCPRegistry provides a Kubernetes-native way to manage MCP server registries using Custom Resource Definitions (CRDs). It supports multiple source types for registry data including ConfigMaps, HTTP URLs, Git repositories, and external registries.

## Files in this directory

- `sample-configmap-registry.yaml` - Example MCPRegistry using ConfigMap as source
- `registry-configmap.yaml` - Example ConfigMap containing registry data
- `README.md` - This file

## Quick Start

### Prerequisites

1. ToolHive Operator must be installed in your Kubernetes cluster
2. The `toolhive-system` namespace should exist
3. MCPRegistry CRD should be installed

### Method 1: Initialize ConfigMap from existing registry.json

If you want to use the official ToolHive registry data, you can create a ConfigMap directly from the registry.json file:

```bash
# Create the ConfigMap from the registry.json file
kubectl create configmap toolhive-registry-data \
  --namespace=toolhive-system \
  --from-file=registry.json=pkg/registry/data/registry.json

# Apply the MCPRegistry resource
kubectl apply -f examples/operator/mcp-registries/sample-configmap-registry.yaml
```

### Method 2: Use the provided example ConfigMap

Alternatively, you can use the example ConfigMap that contains a subset of servers:

```bash
# Apply the example ConfigMap
kubectl apply -f examples/operator/mcp-registries/registry-configmap.yaml

# Apply the MCPRegistry resource
kubectl apply -f examples/operator/mcp-registries/sample-configmap-registry.yaml
```

## Verifying the Setup

After applying the resources, you can verify the setup:

```bash
# Check the MCPRegistry status
kubectl get mcpregistry -n toolhive-system

# Get detailed information about the registry
kubectl describe mcpregistry toolhive-community-registry -n toolhive-system

# Check if the ConfigMap was created
kubectl get configmap toolhive-registry-data -n toolhive-system

# View the registry data
kubectl get configmap toolhive-registry-data -n toolhive-system -o jsonpath='{.data.registry\.json}' | jq .
```

## Understanding the MCPRegistry Configuration

### Source Configuration

The example uses a ConfigMap source:

```yaml
source:
  type: configmap
  configmap:
    name: toolhive-registry-data  # Name of the ConfigMap
    key: registry.json            # Key within the ConfigMap containing registry data
```

### Sync Policy

The registry is configured for automatic synchronization:

```yaml
syncPolicy:
  type: automatic          # Can be "automatic" or "manual"
  interval: "1h"           # Sync interval using Go duration format
                          # Valid units: s, m, h
                          # Examples: "3m", "1h", "24h", "168h"
  retryPolicy:
    maxAttempts: 3
    backoffInterval: "30s"  # Retry delay (same format as interval)
    backoffMultiplier: "2.0"
```

### Filtering

The example includes tag-based filtering:

```yaml
filter:
  tags:
    include:              # Only include servers with these tags
      - "database"
      - "api"
      - "filesystem"
    exclude:              # Exclude servers with these tags
      - "experimental"
```

## Sync Interval Formats

MCPRegistry uses Go's duration format for specifying time intervals. Here are the supported formats:

### Time Units
- `s` - seconds
- `m` - minutes
- `h` - hours

### Example Intervals
```yaml
# Common intervals
interval: "30s"     # 30 seconds
interval: "3m"      # 3 minutes  
interval: "1h"      # 1 hour
interval: "12h"     # 12 hours

# Longer intervals (using hours)
interval: "24h"     # 1 day
interval: "168h"    # 1 week (7 × 24h)
interval: "720h"    # 1 month (30 × 24h)
interval: "8760h"   # 1 year (365 × 24h)

# Combined units
interval: "1h30m"   # 1 hour 30 minutes
interval: "2h15m30s" # 2 hours 15 minutes 30 seconds

# Decimal values
interval: "1.5h"    # 1.5 hours (90 minutes)
interval: "0.5h"    # 30 minutes
```

### Recommended Intervals
- **Development**: `"1m"` - `"5m"` (frequent testing)
- **Staging**: `"15m"` - `"1h"` (moderate frequency)  
- **Production**: `"1h"` - `"24h"` (conservative, stable)
- **Long-term**: `"168h"` (weekly), `"720h"` (monthly) for stable registries

### Note on Long Intervals
Go's duration format doesn't support `d` (days), `w` (weeks), or `M` (months) units directly. 
For longer intervals, use hours:
- 1 day = `"24h"`
- 1 week = `"168h"` 
- 1 month (30 days) = `"720h"`
- 1 year = `"8760h"`

## Available Source Types

MCPRegistry supports multiple source types:

### ConfigMap Source
```yaml
source:
  type: configmap
  configmap:
    name: my-registry-data
    namespace: toolhive-system  # Optional, defaults to MCPRegistry namespace
    key: registry.json          # Optional, defaults to "registry.json"
```

### URL Source
```yaml
source:
  type: url
  url:
    url: "https://raw.githubusercontent.com/stacklok/toolhive/main/pkg/registry/data/registry.json"
    headers:                    # Optional HTTP headers
      Authorization: "Bearer token"
    authentication:             # Optional authentication
      bearerToken:
        secretRef:
          name: registry-auth
          key: token
```

### Git Source (Future)
```yaml
source:
  type: git
  git:
    repository: "https://github.com/stacklok/toolhive"
    ref: "main"                # Optional, defaults to "main"
    path: "pkg/registry/data/registry.json"  # Optional, defaults to "registry.json"
```

## Registry Data Format

The registry data should be in ToolHive format by default. The structure includes:

```json
{
  "$schema": "https://raw.githubusercontent.com/stacklok/toolhive/main/pkg/registry/data/schema.json",
  "version": "1.0.0",
  "last_updated": "2025-09-02T00:17:21Z",
  "servers": {
    "server-name": {
      "description": "Server description",
      "tier": "Official",
      "status": "Active",
      "transport": "stdio",
      "tools": ["tool1", "tool2"],
      "image": "ghcr.io/stacklok/dockyard/uvx/server-name:1.0.0",
      "command": ["uvx", "server-name"],
      "tags": ["tag1", "tag2"]
    }
  }
}
```

## Working with Registry Data

### Updating Registry Data

To update the registry data in a ConfigMap:

```bash
# Update the ConfigMap with new registry data
kubectl create configmap toolhive-registry-data \
  --namespace=toolhive-system \
  --from-file=registry.json=path/to/new/registry.json \
  --dry-run=client -o yaml | kubectl apply -f -

# The MCPRegistry will automatically detect the change and sync
```

### Manual Sync Trigger

If using manual sync policy, you can trigger a sync by annotating the MCPRegistry:

```bash
kubectl annotate mcpregistry toolhive-community-registry \
  -n toolhive-system \
  toolhive.stacklok.dev/sync-trigger="$(date +%s)"
```

## Monitoring and Troubleshooting

### Check Registry Status

```bash
# View registry phase and status
kubectl get mcpregistry -n toolhive-system -o wide

# Get detailed status information
kubectl get mcpregistry toolhive-community-registry -n toolhive-system -o jsonpath='{.status}' | jq .
```

### Common Status Phases

- `Pending` - Registry is being initialized
- `Syncing` - Registry is currently synchronizing data
- `Ready` - Registry is ready and up-to-date
- `Failed` - Registry encountered an error
- `Updating` - Registry is updating its data

### Troubleshooting

1. **Registry stuck in Pending/Syncing:**
   - Check if the ConfigMap exists and contains valid data
   - Verify the ConfigMap key specified in the source configuration

2. **Registry in Failed state:**
   - Check the registry status message: `kubectl describe mcpregistry <name>`
   - Verify the registry data format is valid JSON
   - Check operator logs for detailed error messages

3. **Sync not working:**
   - Verify sync policy configuration
   - Check if the source data has changed
   - Look for controller reconciliation errors in operator logs

## Next Steps

After setting up the registry:

1. **Deploy MCP Servers**: Use the registry to deploy MCP servers with `kubectl create mcpserver`
2. **Configure Filters**: Adjust filtering to show only relevant servers for your use case
3. **Set up Monitoring**: Monitor registry sync status and server deployment health
4. **Integrate with GitOps**: Manage registry configuration through your GitOps workflow

## Additional Resources

- [MCPRegistry CRD API Reference](../../../docs/operator/crd-api.md#mcpregistry)
- [ToolHive Operator Documentation](../../../docs/operator/)
- [MCP Server Registry Format](https://github.com/stacklok/toolhive/blob/main/pkg/registry/data/schema.json)