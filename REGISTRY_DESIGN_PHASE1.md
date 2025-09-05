# Kubernetes Registry Design - Phase 1 Implementation

This document summarizes the Phase 1 design changes for implementing the Kubernetes Registry functionality in ToolHive, based on the proposal in PR #1641.

## Overview

Phase 1 implements a native Kubernetes registry system using Custom Resource Definitions (CRDs) that provides:
- Registry functionality using CRDs for GitOps compatibility
- Support for both local registry entries and external registry synchronization
- Format conversion between ToolHive and upstream registry formats
- REST API for programmatic access to registry data (moved to Task 2.4)

## Core Components

### 1. MCPRegistry CRD

The primary custom resource that defines registry sources and configurations.

**Required Fields:**
- `source`: Defines where to fetch registry data from (ConfigMap, URL, Git, or Registry sources)
- `type`: Source type (configmap, url, git, registry)

**Optional Configuration:**
- `displayName`: Human-readable registry name
- `format`: Registry data format (toolhive or upstream)
- `syncPolicy`: Synchronization behavior with configurable intervals
- `filter`: Include/exclude criteria for servers

**Example:**
```yaml
apiVersion: toolhive.stacklok.io/v1alpha1
kind: MCPRegistry
metadata:
  name: upstream-community
  namespace: toolhive-system
spec:
  displayName: "MCP Community Registry"
  format: upstream
  source:
    type: url
    url: "https://github.com/modelcontextprotocol/registry/raw/main/registry.json"
```

### 2. Registry Controller

**Primary Responsibilities:**
- Manages MCPRegistry lifecycle and synchronization
- Handles format conversion between ToolHive and upstream formats
- Automatic synchronization with external registry sources
- Status tracking for sync operations and error conditions
- Works with existing MCPServer CRD for complete registry-to-deployment workflow

**Controller Features:**
- ConfigMap-based registry storage for Phase 1 implementation
- Bidirectional format conversion capabilities
- Configurable sync intervals and policies
- Error handling and status reporting

### 3. CLI Integration

New registry management commands:

- `thv registry list` - View available registries and their status
- `thv registry add` - Add new registry sources
- `thv registry sync` - Trigger manual synchronization
- `thv registry info` - Get detailed information about a specific registry

## Architecture Changes

### Registry Storage
- **Phase 1**: ConfigMap-based storage for immediate implementation
- Registry data stored as ConfigMaps in the operator namespace
- Status tracking through CRD status fields

### Format Conversion
- **Bidirectional conversion** between ToolHive and upstream registry formats
- Leverages existing upstream conversion capabilities
- Maintains compatibility with ecosystem tooling

### Multi-Registry Support
- Support for multiple registry sources simultaneously
- Registry hierarchy following upstream model specifications
- Conflict resolution for duplicate server entries

## Implementation Details

### Controller Architecture
```
MCPRegistry CRD → Registry Controller → ConfigMap Storage
                                    ↓
                              Format Conversion
                                    ↓
                              MCPServer CRDs (existing)
```

### Synchronization Model
- **Manual sync**: Triggered via CLI or API calls
- **Automatic sync**: Configurable intervals via syncPolicy
- **Status tracking**: Success/failure states and error messages
- **Conflict handling**: Last-writer-wins for registry conflicts

### Source Types Support (Phase 1)
1. **ConfigMap**: Direct ConfigMap references ✅
2. **URL**: HTTP/HTTPS endpoints serving registry JSON (Phase 2)
3. **Git**: Git repository sources (Phase 2)
4. **Registry**: External registry references (Phase 2)

## Integration Points

### Existing MCPServer CRD
- Registry provides server definitions for manual MCPServer creation
- Users reference registry data when creating MCPServer resources
- Preserves existing MCPServer functionality and workflows
- No automatic server deployment - registry is catalog only

### REST API (Task 2.4)
- HTTP endpoints for single registry instance access
- Support for both ToolHive and upstream format output
- Registry metadata, server listing, and individual server endpoints
- Format conversion via query parameter with ToolHive as default
- **Deployment Architecture**: Separate Registry API service (Option C)
  - New Go package: `cmd/thv-registry-api/` and `pkg/registryapi/`
  - Independent container image: `ghcr.io/stacklok/toolhive/thv-registry-api:latest`
  - One API deployment per MCPRegistry instance for clean separation
  - Same repository for code reuse and synchronized releases
- Integration with existing auth mechanisms (planned for Phase 2)
- OpenAPI documentation for client integration

### GitOps Compatibility
- Declarative registry management through CRD-based operations
- Version control for registry configurations
- Kubernetes-native resource management

## Development Tasks for Phase 1

### Core Implementation
1. Define MCPRegistry CRD with comprehensive field validation
2. Implement Registry Controller with ConfigMap source support
3. Add format conversion logic for upstream compatibility
4. Implement status tracking and error reporting

### CLI Commands
1. Implement `thv registry list` command
2. Implement `thv registry add` command
3. Implement `thv registry sync` command
4. Implement `thv registry info` command

### Testing
1. Unit tests for Registry Controller
2. Integration tests for format conversion
3. E2E tests for CLI commands
4. Operator deployment tests

### Documentation
1. CRD field reference documentation
2. Usage examples and workflows
3. Migration guide from existing registry system
4. API documentation updates

## Prioritized Implementation Tasks

This section provides a phased delivery approach for Phase 1, organized by priority and dependencies.

### Sprint 1: Foundation (Weeks 1-2)
**Priority: Critical - Foundation for all other work**

1. **Task 1.1: MCPRegistry CRD Definition** ⭐ **COMPLETED** ✅
   - ✅ Define CRD schema with all required and optional fields
   - ✅ Add field validation rules and constraints
   - ✅ Create initial status subresource structure
   - ✅ Generate CRD manifests and examples
   - **Dependencies:** None
   - **Deliverable:** CRD YAML files in `api/v1alpha1/`
   - **Status:** COMPLETED - MCPRegistry CRD fully implemented with comprehensive validation

2. **Task 1.2: Basic Controller Scaffolding** ⭐ **COMPLETED** ✅
   - ✅ Set up controller structure using controller-runtime
   - ✅ Implement basic reconcile loop with logging
   - ✅ Add RBAC permissions for MCPRegistry resources
   - **Dependencies:** Task 1.1
   - **Deliverable:** Basic controller structure in `controllers/`
   - **Status:** COMPLETED - Full controller implementation with reconciliation logic

3. **Task 1.3: ConfigMap Storage Foundation** ⭐ **COMPLETED** ✅
   - ✅ Implement ConfigMap creation/update logic
   - ✅ Add basic error handling and status updates
   - ✅ Create helper functions for ConfigMap operations
   - **Dependencies:** Task 1.2
   - **Deliverable:** ConfigMap storage implementation
   - **Status:** COMPLETED - StorageManager interface with ConfigMap implementation

### Sprint 2: Core Registry Logic (Weeks 3-4)
**Priority: High - Core functionality**

4. **Task 2.1: Format Conversion Engine** ⭐ **COMPLETED** ✅
   - ✅ Create FormatConverter interface for decoupling
   - ✅ Implement ToolHive to upstream format conversion (leverages existing pkg/registry)
   - ✅ Implement upstream to ToolHive format conversion (leverages existing pkg/registry)
   - ✅ Add validation and format detection for both formats
   - ✅ Integrate format conversion into ConfigMap source handler
   - ✅ Add comprehensive unit tests with 100% coverage
   - **Dependencies:** Task 1.3
   - **Deliverable:** Format conversion package with interface and tests
   - **Status:** COMPLETED - Full format conversion system implemented with validation and integration

5. **Task 2.2: ConfigMap Source Support** ⭐ **COMPLETED** ✅
   - ✅ Implement ConfigMap source type handling
   - ✅ Add ConfigMap watching and updates
   - ✅ Implement sync logic for ConfigMap sources
   - ✅ Add comprehensive unit tests
   - **Dependencies:** Task 2.1
   - **Deliverable:** ConfigMap source controller logic
   - **Status:** COMPLETED - Full ConfigMap source handler with validation and testing

6. **Task 2.3: URL Source Support** ⭐ **MOVED TO PHASE 2** 📋
   - ❌ Implement HTTP client for registry fetching (Phase 2)
   - ❌ Add retry logic and error handling (Phase 2)  
   - ❌ Implement caching for URL sources (Phase 2)
   - **Dependencies:** Task 2.1
   - **Deliverable:** URL source controller logic (Phase 2)
   - **Status:** MOVED TO PHASE 2 - External sources planned for next phase

7. **Task 2.4: REST API for Registry Access** ⭐ **HIGH PRIORITY**
   - Create separate Registry API service as new Go package (`cmd/thv-registry-api/`)
   - Implement HTTP endpoints for single registry instance access
   - Add support for both ToolHive and upstream format output via query parameter
   - Create API handlers for registry info, server listing, and individual server access
   - Integrate with existing FormatConverter for format conversion
   - Build independent container image for deployment per MCPRegistry instance
   - Add OpenAPI documentation for registry endpoints
   - **Architecture:** Separate service (Option C) - clean separation from operator
   - **Endpoints:** `/api/v1/registry/info`, `/api/v1/registry/servers`, `/api/v1/registry/servers/{name}`
   - **Format Support:** Default ToolHive format with `?format=upstream` option
   - **Dependencies:** Task 2.2, Task 2.1 (FormatConverter)
   - **Deliverable:** Independent Registry API service with container image
   - **Estimate:** 4-5 days
   - **Claude Code Estimate:** 6-8 hours
   - **Implementation Tasks:**
     1. Create cmd/thv-registry-api Go package structure
     2. Create pkg/registryapi package with HTTP handlers
     3. Implement /api/v1/registry/info endpoint
     4. Implement /api/v1/registry/servers endpoint with format support
     5. Implement /api/v1/registry/servers/{name} endpoint
     6. Integrate FormatConverter for query parameter format conversion
     7. Add Kubernetes client integration to read MCPRegistry and ConfigMap data
     8. Update MCPRegistry controller to deploy Registry API service
     9. Create Deployment and Service manifests for Registry API
     10. Add Registry API URL to MCPRegistry status
     11. Create Dockerfile for registry API container image
     12. Add OpenAPI documentation for registry endpoints
     13. Write unit tests for all API handlers

### Sprint 3: CLI and Integration (Weeks 5-6)
**Priority: High - User interface and integration**

7. **Task 3.1: Basic CLI Commands** ⭐ **HIGH PRIORITY**
   - Implement `thv registry list` command
   - Implement `thv registry info` command
   - Add basic output formatting and error handling
   - **Dependencies:** Task 2.2
   - **Deliverable:** Basic CLI commands in `cmd/thv/app/`
   - **Estimate:** 3-4 days
   - **Claude Code Estimate:** 3-4 hours

8. **Task 3.2: Registry Management CLI** ⭐ **MEDIUM-HIGH PRIORITY**
   - Implement `thv registry add` command
   - Implement `thv registry sync` command
   - Add input validation and user feedback
   - **Dependencies:** Task 3.1 (Task 2.3 moved to Phase 2)
   - **Deliverable:** Complete CLI command set
   - **Estimate:** 3-4 days
   - **Claude Code Estimate:** 4-6 hours

9. **Task 3.3: REMOVED** ❌ **INVALID REQUIREMENT**
   - ❌ Automatic MCPServer creation is not intended
   - **Reason:** Registry is a catalog, not a deployment system
   - **Note:** Users manually create MCPServers referencing registry data
   - **Status:** REMOVED - Invalid architectural assumption

### Sprint 4: Testing and Polish (Weeks 7-8)
**Priority: Medium-High - Quality assurance**

10. **Task 4.1: Unit Tests** ⭐ **HIGH PRIORITY**
    - Write controller unit tests with mocked dependencies
    - Write format conversion unit tests
    - Write CLI command unit tests
    - **Dependencies:** Tasks 2.1, 2.2, 3.1
    - **Deliverable:** Comprehensive unit test suite
    - **Estimate:** 4-5 days
    - **Claude Code Estimate:** 6-8 hours

11. **Task 4.2: Integration Tests** ⭐ **MEDIUM-HIGH PRIORITY**
    - Write controller integration tests with test environment
    - Write end-to-end workflow tests
    - Add test data and fixtures
    - **Dependencies:** Task 4.1, Task 3.3
    - **Deliverable:** Integration test suite
    - **Estimate:** 3-4 days
    - **Claude Code Estimate:** 4-6 hours

12. **Task 4.3: Status Tracking and Error Handling** ⭐ **MEDIUM PRIORITY**
    - Implement comprehensive status reporting
    - Add detailed error messages and conditions
    - Implement retry logic for failed operations
    - **Dependencies:** Task 3.3
    - **Deliverable:** Robust status and error handling
    - **Estimate:** 2-3 days
    - **Claude Code Estimate:** 2-3 hours

### Sprint 5: Documentation and Deployment (Week 9)
**Priority: Medium - Documentation and deployment readiness**

13. **Task 5.1: API Documentation** ⭐ **MEDIUM PRIORITY**
    - Generate CRD reference documentation
    - Update REST API documentation
    - Add code examples and usage patterns
    - **Dependencies:** Task 4.2
    - **Deliverable:** Complete API documentation
    - **Estimate:** 2-3 days
    - **Claude Code Estimate:** 2-3 hours

14. **Task 5.2: User Documentation** ⭐ **MEDIUM PRIORITY**
    - Write user guides and tutorials
    - Create migration documentation
    - Add troubleshooting guides
    - **Dependencies:** Task 5.1
    - **Deliverable:** User-facing documentation
    - **Estimate:** 2-3 days
    - **Claude Code Estimate:** 3-4 hours

15. **Task 5.3: Deployment Testing** ⭐ **MEDIUM PRIORITY**
    - Test operator deployment with new CRDs
    - Verify upgrade/downgrade scenarios
    - Test multi-namespace scenarios
    - **Dependencies:** Task 4.2
    - **Deliverable:** Deployment validation
    - **Estimate:** 1-2 days
    - **Claude Code Estimate:** 1-2 hours

### Risk Mitigation Tasks (Parallel to other sprints)

16. **Task R.1: Backup/Rollback Strategy** ⭐ **LOW-MEDIUM PRIORITY**
    - Implement registry backup mechanisms
    - Add rollback capabilities for failed syncs
    - Create data migration utilities
    - **Dependencies:** Task 2.2
    - **Deliverable:** Backup and rollback tools
    - **Estimate:** 2-3 days
    - **Claude Code Estimate:** 3-4 hours

17. **Task R.2: Performance Optimization** ⭐ **LOW PRIORITY**
    - Optimize controller reconcile loops
    - Add caching for frequently accessed data
    - Implement batching for large registries
    - **Dependencies:** Task 4.1
    - **Deliverable:** Performance improvements
    - **Estimate:** 2-3 days
    - **Claude Code Estimate:** 2-4 hours

### Critical Path Analysis

**Critical Path:** Tasks 1.1 → 1.2 → 1.3 → 2.2 → 2.4 → 3.1 → 3.2 → 4.1 → 4.2
**Total Estimated Duration:** 7-9 weeks
**Claude Code Total Duration:** 35-50 hours (1-1.5 weeks)
**Key Milestones:**
- Week 2: Basic CRD and controller foundation
- Week 4: Core registry functionality working
- Week 6: CLI integration complete
- Week 8: Full testing coverage
- Week 9: Documentation and deployment ready

### Delivery Gates

**Gate 1 (End of Sprint 1):** CRD deployed, basic controller running
**Gate 2 (End of Sprint 2):** ConfigMap and URL sources working with format conversion
**Gate 3 (End of Sprint 3):** CLI commands functional, MCPServer integration working
**Gate 4 (End of Sprint 4):** All tests passing, error handling robust
**Gate 5 (End of Sprint 5):** Documentation complete, deployment validated

## Success Criteria

Phase 1 is considered complete when:
- MCPRegistry CRD is deployed and functional
- Registry Controller can sync from ConfigMap sources
- Format conversion works bidirectionally
- CLI commands are functional and tested
- Basic multi-registry support is operational
- Documentation is complete and accurate

## Future Phases

Phase 1 establishes the foundation for:

### **Phase 2: External Sources and API Security**
- Git source support and advanced synchronization
- URL source implementation with retry logic and caching
- **Task 2.5: Authentication and Authorization Integration**
  - Integrate REST API with existing ToolHive auth mechanisms
  - Add RBAC support for registry access (read/write permissions)
  - Implement API key authentication for external clients
  - Add JWT token validation for user-based access
  - Create authorization policies using existing Cedar framework
  - Support namespace-based access control for multi-tenant scenarios
  - Add audit logging for registry API access
  - **Dependencies:** Task 2.4 (REST API foundation)
  - **Deliverable:** Secure, multi-tenant registry API with comprehensive auth

### **Phase 3: Enhanced Catalog and Trust**
- Enhanced catalog system with trust levels
- Server verification and signing support
- Publisher authentication and metadata validation

### **Phase 4: Advanced Features**
- Advanced filtering and server selection
- Cross-registry federation and discovery
- Performance optimization and caching

### **Phase 5: Production Readiness**
- Complete GitOps integration and scaling features
- Monitoring, alerting, and observability
- High availability and disaster recovery