# Phase 09 — Web & Cloud Connection UX

## Role

You are a senior frontend/backend engineer specializing in enterprise infrastructure tooling and cloud connection UX.

You are working on the WeaveLens project as part of an orchestrated multi-agent development workflow.

## Context

WeaveLens is a cloud infrastructure discovery and visualization platform.

AWS is currently the only implemented cloud provider.

Azure and GCP may be supported in the future.

The current goal is NOT to implement multi-cloud functionality.

The goal is to build a clean Web UX that works for AWS today and can naturally expand to additional cloud providers later.

## Objective

Implement the Web UI for:

* AWS connection status;
* AWS account identity;
* AWS region;
* scan controls;
* scan status;
* infrastructure graph;
* resource details;
* AWS setup guidance.

The frontend MUST NOT directly communicate with AWS.

The backend is responsible for AWS authentication and AWS API access.

## Frontend Architecture

Use the project's selected frontend framework.

Keep responsibilities separated:

```text
UI
 ↓
Frontend API Client
 ↓
Backend API / gRPC gateway
 ↓
Application Layer
 ↓
AWS Infrastructure
 ↓
AWS
```

Do NOT import AWS SDKs into the frontend.

## Cloud Provider UX

Use provider-neutral UI concepts where they make sense.

Prefer:

```text
Cloud Connections
```

over hard-coding the entire application around:

```text
AWS Login
```

However, the current implementation should only expose AWS because Azure/GCP are not implemented yet.

Example:

```text
Cloud Connections

AWS
● Connected

Account:
123456789012

Region:
ap-southeast-1

Identity:
arn:aws:iam::123456789012:role/WeaveLensScanner
```

Do not display unsupported Azure/GCP connection options as if they are functional.

## AWS Credentials

The Web UI MUST NOT provide a form for:

* AWS Access Key ID;
* AWS Secret Access Key;
* AWS Session Token.

Do not implement:

```text
AWS Login
Access Key
Secret Key
[Login]
```

The frontend must never receive AWS secret material.

## Credential Model

WeaveLens uses credentials available to the backend/runtime environment.

For local development, the backend may use the AWS SDK default credential chain.

Examples:

```text
AWS_PROFILE
~/.aws/credentials
~/.aws/config
environment credentials
```

For cross-account access, the backend may use STS AssumeRole according to Phase 03.

The frontend does not manage the underlying credentials.

## AWS Connection Status

Provide a clear connection state:

```text
Connected
Connecting
Not Connected
Authentication Error
Access Denied
Configuration Error
Unknown Error
```

Do not expose raw AWS SDK errors directly to users.

Translate errors into useful user-facing messages while preserving technical details in backend logs.

## Connection Information

When connected, display safe identity information such as:

* AWS Account ID;
* caller ARN;
* region;
* credential source type where available;
* connection status.

Never display:

* Access Key;
* Secret Access Key;
* Session Token;
* credential file contents.

## AWS Setup Guide

Provide a concise setup/help experience for users who do not have AWS credentials available to the backend.

This may be:

```text
/setup/aws
```

or an equivalent route following the project's routing conventions.

The setup guide should explain:

### Local development

1. Install/configure AWS CLI.
2. Create or select an AWS profile.
3. Configure credentials through the standard AWS mechanism.
4. Start WeaveLens using that profile.
5. Return to WeaveLens and retry the connection.

Example:

```bash
aws configure --profile weavelens
```

Then:

```bash
AWS_PROFILE=weavelens
```

The guide must NOT ask the user to paste their Secret Access Key into the browser.

### Cross-account

Explain at a high level that WeaveLens can use an IAM Role through STS AssumeRole.

Explain that the role must grant the required permissions and that the runtime identity must be allowed to assume it.

Do not implement an interactive AWS IAM provisioning workflow in this phase.

## Product UX Principle

The setup guide is documentation/help, not an AWS authentication portal.

Do not attempt to replicate AWS login inside WeaveLens.

Do not ask users to enter long-lived AWS credentials into the Web UI.

## Scan UX

Provide a clear scan workflow.

Example:

```text
AWS Connection
      ↓
Connected
      ↓
Select Region
      ↓
Start Scan
      ↓
Scanning
      ↓
Resources Discovered
      ↓
Graph Visualization
```

The UI should clearly communicate:

* scan started;
* scan in progress;
* scan completed;
* scan partially completed;
* scan failed;
* scan cancelled.

## Graph UI

Display discovered resources and relationships visually.

Use the existing graph model/API from previous phases.

Resources should have visual differentiation by resource type.

For example:

```text
VPC
Subnet
EC2
RDS
ALB
```

Use a consistent visual legend.

Do not encode AWS-specific knowledge directly into generic UI components where avoidable.

## Legend

Provide a clear legend explaining resource visualization.

Example:

```text
Legend

● VPC
● Subnet
● EC2
● RDS
● Load Balancer
```

The exact visual representation should follow the project's existing design system.

## Resource Details

Selecting a resource should show useful non-secret metadata.

Examples:

```text
Resource
────────────────────
Type: EC2 Instance
ID: i-0123456789
Region: ap-southeast-1
VPC: vpc-0123456789
State: running
```

Do not expose credentials or sensitive configuration unnecessarily.

## API Boundary

Frontend should communicate with backend APIs only.

Conceptually:

```text
Frontend
   │
   ├── GET connection status
   ├── GET AWS identity
   ├── POST scan
   ├── GET scan status
   ├── GET resources
   └── GET graph
```

The exact endpoints must follow the existing backend/API contracts.

Do not invent a second API contract if one already exists.

## Error UX

Errors should be actionable.

Examples:

```text
AWS credentials not found.

Configure an AWS profile for the WeaveLens runtime
and retry the connection.
```

or:

```text
AWS access denied.

The current AWS identity does not have permission
to perform this discovery operation.
```

Do not display:

```text
AccessDeniedException: ...
```

as the primary user-facing message.

## Multi-Cloud Future Compatibility

Design reusable UI components where it provides real value.

For example:

```text
CloudConnectionCard
ResourceCard
ResourceTypeBadge
ConnectionStatus
SetupGuide
```

Provider-specific implementation may be composed inside them.

Do NOT create:

```text
AzureConnection
GCPConnection
```

without actual Azure/GCP functionality.

Do NOT add disabled fake providers merely for architectural appearance.

## Security Requirements

1. Frontend never calls AWS directly.
2. Frontend never stores AWS secret credentials.
3. Frontend never receives AWS credentials from backend.
4. No credentials appear in browser localStorage.
5. No credentials appear in URL parameters.
6. No credentials appear in frontend logs.
7. Backend errors exposed to frontend must not contain secrets.
8. HTTPS must be used in production deployment.

## Responsive UX

The graph visualization must remain usable on desktop screens.

Handle:

* large graphs;
* zoom;
* pan;
* resource selection;
* relationship visibility.

Do not attempt to solve arbitrary graph-scale optimization in this phase.

## Testing

Add appropriate tests for:

* connection status;
* scan states;
* error states;
* setup guide rendering;
* resource details;
* graph rendering;
* legend;
* API error handling.

Do not require real AWS credentials for frontend tests.

## Accessibility

Use accessible:

* buttons;
* labels;
* dialogs;
* status indicators;
* keyboard interactions.

Do not rely only on color to communicate resource type.

The legend and UI should provide text labels.

## Constraints

Do NOT implement:

* Azure integration;
* GCP integration;
* AWS credential storage;
* Access Key login form;
* Secret Key login form;
* direct frontend-to-AWS calls;
* database;
* new authentication system for WeaveLens users unless already defined by the project architecture.

## Acceptance Criteria

1. Web UI can display AWS connection status.
2. Web UI can display safe AWS identity information.
3. Web UI can select/configure the scan region through the existing backend contract.
4. Web UI can start and monitor a scan.
5. Web UI can display discovered resources.
6. Web UI can display resource relationships as a graph.
7. Web UI includes a resource legend.
8. Web UI provides resource details.
9. Web UI provides an AWS setup guide.
10. No AWS Access Key/Secret Key login form exists.
11. No AWS credential material reaches the frontend.
12. Frontend never calls AWS directly.
13. AWS errors are translated into useful user-facing messages.
14. UI components are reusable where appropriate for future cloud providers.
15. Azure/GCP functionality is NOT implemented.

## Verification

Run:

```bash
go build ./...
go test ./...
```

Run the frontend's configured build and test commands.

Run linting and static analysis.

Review the complete diff.

Verify that no credential material is present in source code, test fixtures, browser storage, or logs.

## Git

Create one focused commit:

```text
feat(web): add cloud connection and aws scan ux
```

Do NOT automatically proceed to the next phase.
