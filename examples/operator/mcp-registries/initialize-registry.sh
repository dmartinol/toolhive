#!/bin/bash

# ToolHive MCPRegistry Initialization Script
# This script demonstrates how to initialize an MCPRegistry with ConfigMap source

set -e

NAMESPACE="toolhive-system"
REGISTRY_NAME="toolhive-community-registry"
CONFIGMAP_NAME="toolhive-registry-data"

echo "🚀 Initializing ToolHive MCPRegistry with ConfigMap source..."

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "📦 Creating namespace: $NAMESPACE"
    kubectl create namespace "$NAMESPACE"
else
    echo "✅ Namespace $NAMESPACE already exists"
fi

# Method 1: Create ConfigMap from the official registry.json file
if [ -f "pkg/registry/data/registry.json" ]; then
    echo "📋 Creating ConfigMap from official registry.json..."
    kubectl create configmap "$CONFIGMAP_NAME" \
        --namespace="$NAMESPACE" \
        --from-file=registry.json=pkg/registry/data/registry.json \
        --dry-run=client -o yaml | kubectl apply -f -
    echo "✅ ConfigMap created from official registry data"
else
    # Method 2: Use the example ConfigMap
    echo "📋 Creating ConfigMap from example data..."
    kubectl apply -f examples/operator/mcp-registries/registry-configmap.yaml
    echo "✅ ConfigMap created from example data"
fi

# Apply the MCPRegistry resource
echo "🔧 Creating MCPRegistry resource..."
kubectl apply -f examples/operator/mcp-registries/sample-configmap-registry.yaml

echo "⏳ Waiting for MCPRegistry to be ready..."
kubectl wait --for=condition=Ready mcpregistry/"$REGISTRY_NAME" -n "$NAMESPACE" --timeout=300s || true

echo ""
echo "📊 Registry Status:"
kubectl get mcpregistry "$REGISTRY_NAME" -n "$NAMESPACE" -o wide

echo ""
echo "🔍 Detailed Registry Information:"
kubectl describe mcpregistry "$REGISTRY_NAME" -n "$NAMESPACE"

echo ""
echo "🎉 MCPRegistry initialization complete!"
echo ""
echo "Next steps:"
echo "  1. Check registry status: kubectl get mcpregistry -n $NAMESPACE"
echo "  2. View registry data: kubectl get configmap $CONFIGMAP_NAME -n $NAMESPACE -o jsonpath='{.data.registry\.json}' | jq ."
echo "  3. Deploy MCP servers from the registry"