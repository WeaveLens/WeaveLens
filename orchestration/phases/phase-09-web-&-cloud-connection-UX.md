# Phase 09 — Web & Cloud Connection UX

## Role

You are a senior frontend/backend engineer specializing in enterprise infrastructure tooling, cloud connection UX, responsive desktop applications, and infrastructure visualization.

You are working on the WeaveLens project as part of an orchestrated multi-agent development workflow.

## Context

WeaveLens is a cloud infrastructure discovery and visualization platform.

AWS is currently the only implemented cloud provider.

Azure and GCP may be supported in the future.

The current goal is NOT to implement multi-cloud functionality.

The goal is to build a clean, enterprise-oriented Web UX that works for AWS today and can naturally expand to additional cloud providers later.

WeaveLens is primarily a desktop-oriented infrastructure visualization tool.

The primary target display is Full HD:

```text
1920 × 1080
```

The UI must also support larger desktop displays such as:

```text
2560 × 1440
3840 × 2160
```

The layout must be responsive and must not be designed around a single fixed screen resolution.

---

# Objective

Implement the Web UI for:

* AWS connection status;
* AWS account identity;
* AWS region;
* scan controls;
* scan status;
* infrastructure graph;
* resource details;
* resource legend;
* AWS setup guidance;
* responsive desktop layout.

The frontend MUST NOT directly communicate with AWS.

The backend is responsible for AWS authentication and AWS API access.

---

# 1. Frontend Architecture

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

Do NOT allow frontend components to contain AWS API calls.

---

# 2. Desktop & Responsive Design

## Primary Resolution

The primary design baseline is:

```text
1920 × 1080 Full HD
```

The application must be designed and visually validated for this resolution.

The UI must also work correctly on larger displays, including:

```text
2560 × 1440
3840 × 2160
```

## Responsive Requirement

The UI MUST NOT assume a fixed:

```text
width: 1920px;
height: 1080px;
```

Do not hard-code the entire application layout to Full HD dimensions.

Instead, use responsive layout mechanisms such as:

* CSS Grid;
* Flexbox;
* relative sizing;
* viewport-aware sizing;
* responsive breakpoints;
* min/max constraints where appropriate.

The application should adapt to available viewport space.

## Large Screens

On larger displays:

```text
2560 × 1440
3840 × 2160
```

the application must:

* use additional available space effectively;
* avoid excessive empty space;
* keep important controls readable;
* prevent content from becoming unnecessarily stretched;
* maintain a consistent visual hierarchy.

Do not simply scale every UI element proportionally with the viewport.

## Maximum Content Width

Use sensible maximum widths for content-heavy sections where appropriate.

For example:

```text
Settings
Connection configuration
Resource details
Documentation
Dialogs
```

These areas should remain readable instead of becoming excessively wide on large screens.

However, infrastructure visualization areas such as the graph/canvas may use most or all of the available viewport.

Conceptually:

```text
┌────────────────────────────────────────────────────────────┐
│ Header                                                     │
├──────────────┬─────────────────────────────────────────────┤
│              │                                             │
│ Navigation   │              Graph / Canvas                 │
│              │                                             │
│              │                                             │
│              │                                             │
├──────────────┴─────────────────────────────────────────────┤
│ Status / Information                                       │
└────────────────────────────────────────────────────────────┘
```

The graph area should be allowed to expand with the viewport.

## Ultrawide Displays

The UI should remain usable on wide and ultrawide desktop screens.

Do not allow:

* excessively stretched forms;
* unreadably long text lines;
* controls drifting too far from their related content;
* graph controls becoming inaccessible.

Use max-width constraints for UI panels where appropriate while allowing the graph to use the available space.

## Smaller Screens

Although desktop is the primary target, the UI must not completely break at smaller viewport sizes.

At narrower widths:

* side panels may collapse;
* navigation may become compact;
* resource details may move into a drawer/modal;
* graph controls may reposition;
* horizontal overflow should be avoided where possible.

Do not sacrifice desktop visualization quality merely to optimize for mobile.

## Browser Zoom

The UI should remain usable under common browser zoom levels.

Do not rely on absolute pixel positioning for core functionality.

---

# 3. Design System

Use the project's existing design system if one exists.

If no design system exists, establish a minimal consistent system for:

* spacing;
* typography;
* headings;
* buttons;
* forms;
* cards;
* badges;
* status indicators;
* dialogs;
* navigation;
* panels.

Avoid introducing multiple UI libraries without justification.

Keep visual language consistent throughout WeaveLens.

---

# 4. Cloud Provider UX

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

Do NOT display unsupported Azure/GCP connection options as if they are functional.

---

# 5. AWS Credentials

The Web UI MUST NOT provide a form for:

* AWS Access Key ID;
* AWS Secret Access Key;
* AWS Session Token.

Do NOT implement:

```text
AWS Login

Access Key:
[................]

Secret Key:
[................]

[Login]
```

The frontend must never receive AWS secret material.

---

# 6. Credential Model

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

---

# 7. AWS Connection Status

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

---

# 8. Connection Information

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

---

# 9. AWS Setup Guide

Provide a concise setup/help experience for users who do not have AWS credentials available to the backend.

This may be:

```text
/setup/aws
```

or an equivalent route following the project's routing conventions.

The setup guide should explain:

## Local Development

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

```text
AWS_PROFILE=weavelens
```

The guide must NOT ask the user to paste their Secret Access Key into the browser.

## Cross-Account

Explain at a high level that WeaveLens can use an IAM Role through STS AssumeRole.

Explain that the role must grant the required permissions and that the runtime identity must be allowed to assume it.

Do not implement an interactive AWS IAM provisioning workflow in this phase.

---

# 10. Product UX Principle

The setup guide is documentation/help, not an AWS authentication portal.

Do not attempt to replicate AWS login inside WeaveLens.

Do not ask users to enter long-lived AWS credentials into the Web UI.

---

# 11. Scan UX

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

---

# 12. Graph UI

The infrastructure graph is one of the primary features of WeaveLens.

The graph should use the available viewport effectively.

At Full HD:

```text
1920 × 1080
```

the graph must have sufficient visual space to display a realistic infrastructure topology.

At larger resolutions:

```text
2560 × 1440
3840 × 2160
```

the graph should expand naturally rather than remaining constrained to a small fixed canvas.

Use:

* zoom;
* pan;
* fit-to-screen;
* resource selection;
* relationship highlighting;
* appropriate graph controls.

Do not use fixed graph dimensions such as:

```text
width: 1200px;
height: 700px;
```

unless they are only fallback values inside a responsive layout.

The graph container should generally follow the available viewport.

---

# 13. Resource Visualization

Resources should have visual differentiation by resource type.

Examples:

```text
VPC
Subnet
EC2
RDS
Load Balancer
```

Use a consistent visual legend.

Do not rely only on color to distinguish resources.

Each resource type must also have:

* text;
* icon;
* label;
* or another accessible visual indicator.

---

# 14. Legend

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

The legend should remain accessible while exploring the graph.

On smaller screens, the legend may collapse into a panel or control.

---

# 15. Resource Details

Selecting a resource should show useful non-secret metadata.

Example:

```text
Resource
────────────────────────
Type: EC2 Instance
ID: i-0123456789
Region: ap-southeast-1
VPC: vpc-0123456789
State: running
```

The resource detail panel should work appropriately across:

```text
1920 × 1080
2560 × 1440
3840 × 2160
```

On smaller screens, it may become a drawer or modal.

Do not expose credentials or sensitive configuration unnecessarily.

---

# 16. Page Layout

Use a layout appropriate for an enterprise infrastructure tool.

Suggested structure:

```text
┌──────────────────────────────────────────────────────────────┐
│ Header                                                       │
├─────────────┬────────────────────────────────────────────────┤
│             │                                                │
│ Sidebar     │ Main Content                                   │
│             │                                                │
│ Dashboard   │                                                │
│ Connections │                                                │
│ Scans       │                                                │
│ Graph       │                                                │
│ Settings    │                                                │
│             │                                                │
└─────────────┴────────────────────────────────────────────────┘
```

The exact navigation structure may follow the existing application architecture.

Do not create unnecessary pages.

---

# 17. API Boundary

Frontend should communicate with backend APIs only.

Conceptually:

```text
Frontend
   │
   ├── Get connection status
   ├── Get AWS identity
   ├── Start scan
   ├── Get scan status
   ├── Get resources
   └── Get graph
```

The exact endpoints must follow existing backend/API contracts.

Do not invent a second API contract if one already exists.

---

# 18. Error UX

Errors should be actionable.

Example:

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

Do not display raw SDK exceptions as the primary user-facing message.

---

# 19. Loading & Empty States

Every major data-driven page must have appropriate:

* loading state;
* empty state;
* error state;
* success state.

For example:

```text
No scan has been performed yet.

Connect an AWS account and start your first scan.
```

Avoid blank screens while data is loading.

---

# 20. Multi-Cloud Future Compatibility

Design reusable UI components where it provides real value.

For example:

```text
CloudConnectionCard
ResourceCard
ResourceTypeBadge
ConnectionStatus
SetupGuide
```

Provider-specific behavior may be composed inside these components.

Do NOT create:

```text
AzureConnection
GCPConnection
```

without actual Azure/GCP functionality.

Do NOT add disabled fake providers merely for architectural appearance.

---

# 21. Security Requirements

1. Frontend never calls AWS directly.
2. Frontend never stores AWS secret credentials.
3. Frontend never receives AWS credentials from backend.
4. No credentials appear in browser localStorage.
5. No credentials appear in URL parameters.
6. No credentials appear in frontend logs.
7. Backend errors exposed to frontend must not contain secrets.
8. HTTPS must be used in production deployment.

---

# 22. Accessibility

Use accessible:

* buttons;
* labels;
* dialogs;
* status indicators;
* keyboard interactions.

Do not rely only on color to communicate resource type.

The legend and UI must provide text labels.

Ensure sufficient contrast.

Interactive graph elements should provide accessible labels where technically possible.

---

# 23. Performance

The Web UI should remain responsive during large infrastructure scans and graph rendering.

Do not block the main UI thread unnecessarily.

For graph rendering:

* avoid unnecessary full graph re-renders;
* avoid excessive DOM nodes where a canvas/WebGL approach is appropriate;
* keep selection interactions responsive.

Do not prematurely optimize before measuring.

---

# 24. Testing

Add appropriate tests for:

* connection status;
* scan states;
* error states;
* setup guide rendering;
* resource details;
* graph rendering;
* legend;
* responsive layout behavior;
* API error handling.

Frontend tests must not require real AWS credentials.

Where practical, verify the layout at:

```text
1920 × 1080
2560 × 1440
3840 × 2160
```

and at a narrower viewport to ensure the UI does not break.

---

# 25. Constraints

Do NOT implement:

* Azure integration;
* GCP integration;
* AWS credential storage;
* Access Key login form;
* Secret Key login form;
* direct frontend-to-AWS calls;
* database;
* new authentication system for WeaveLens users unless already defined by the project architecture.

Do NOT redesign the entire application unnecessarily.

Do NOT introduce a UI framework solely for responsive behavior if an existing project framework already provides the required capabilities.

---

# 26. Acceptance Criteria

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
16. The primary desktop layout is designed and validated for 1920×1080.
17. The UI works correctly at 2560×1440.
18. The UI works correctly at 3840×2160.
19. The graph uses available viewport space effectively.
20. Content-heavy panels do not become excessively wide on large screens.
21. The layout does not rely on a fixed 1920px × 1080px canvas.
22. The UI remains usable at narrower viewport sizes.
23. Core functionality does not depend on absolute pixel positioning.
24. Loading, empty, error, and success states are implemented for major data-driven views.

---

# 27. Existing Implementation Review & Layout Fix

This phase is an update/fix to an existing Web implementation.

Do NOT assume that the current Web implementation already satisfies
the requirements in this specification.

Before implementing new functionality:

1. Inspect the existing Web layout and component hierarchy.
2. Identify the root cause of current layout and responsive issues.
3. Fix the existing implementation rather than hiding symptoms.
4. Preserve working functionality unless it conflicts with this specification.

Specifically investigate the following existing issues:

- The UI does not expand correctly on displays larger than Full HD.
- The Resource panel may show a scrollbar even when it contains no content.
- Resource and sidebar/panel areas may disappear when browser zoom is increased.
- Some containers may have inappropriate fixed width/height values.
- Nested Flexbox/Grid containers may have incorrect min-width/min-height behavior.
- Overflow rules may cause content to be clipped.
- Graph dimensions may be unnecessarily fixed.
- Panels may use flex-shrink/flex-grow incorrectly.

Inspect the complete layout hierarchy:

```text
Viewport
  ↓
Application Root
  ↓
Header
  ↓
Main Workspace
  ├── Sidebar
  ├── Graph / Main Content
  └── Resource Panel
```

---

# 28. Verification

Run:

```bash
go build ./...
go test ./...
```

Run the frontend's configured build and test commands.

Run linting and static analysis.

Verify the UI at minimum:

```text
1920 × 1080
2560 × 1440
3840 × 2160
```

Also verify a narrower viewport to ensure the responsive layout does not break.

Review:

* graph usability;
* navigation;
* connection status;
* setup guide;
* resource details;
* legend;
* responsive behavior;
* accessibility.

Verify that no credential material is present in:

* source code;
* test fixtures;
* browser localStorage;
* browser session storage;
* URLs;
* frontend logs.

Review the complete diff.

---

# 28. Git

Create one focused commit:

```text
feat(web): add cloud connection and aws scan ux
```

Do NOT automatically proceed to the next phase.
