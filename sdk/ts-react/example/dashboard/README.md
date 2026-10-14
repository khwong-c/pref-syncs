# Dashboard Example

A complete React application demonstrating the `PrefSyncProvider` and `usePrefSync` hook from the `pref-sync-react` SDK with an interactive dashboard UI.
It also demonstrates other RESTful APIs features.  

## What It Demonstrates

This example provides a comprehensive, interactive interface for exploring all aspects of the PrefSync SDK:

- **Provider configuration** — Dynamically configuring the SDK connection with an editable base path and access token, with an "Apply Configuration" button to establish or update the connection.
- **Stub IdP integration** — Fetching a JWT token from the embedded stub IdP via `client_credentials` grant directly from the UI.
- **App management** — Creating new apps, selecting between created apps, and authorizing apps for the current user.
- **Preference CRUD** — Fetching, posting, and viewing preferences with a JSON payload editor and preset templates.
- **Provider integration** — Wrapping the application with `PrefSyncProvider` to manage client state and SSE subscriptions.
- **Hook usage** — Using `usePrefSync` to access the current `client`, `appId`, and subscription status.
- **Event handling** — Receiving real-time preference update notifications via the `onEvent` callback with auto-refresh support.
- **Peer client posting** — Simulating external preference updates by posting from a separate client instance to verify SSE delivery.
- **SDK client explorer** — Directly invoking SDK client operations like `getUser()`, `getApp()`, `authoriseApp()`, and `deauthoriseApp()`.
- **Live event log** — A filterable log panel recording all SSE events, API calls, errors, and stream close events in real time.
- **Subscription lifecycle** — Enabling and disabling the SSE subscription via a toggle control.

## How to Run

From the `sdk/ts-react` root directory:

```bash
# Install dependencies
pnpm install

# Start this dashboard example
pnpm dev:dashboard
```

The example serves on `http://localhost:5173` by default (configured in `vite.config.ts`).

### Prerequisites

The Vite dev server proxies API routes (`/app`, `/user`, `/pref`, `/idp`) to `http://127.0.0.1:7086`.
Ensure the PrefSync backend service is running before starting the examples.

## Files

| File         | Description                                                                 |
|--------------|-----------------------------------------------------------------------------|
| `App.tsx`    | Single-file React component with the full dashboard UI and all logic        |
| `main.tsx`   | React entry point that mounts the `App` component                          |
| `index.html` | HTML template for the Vite dev server                                      |

## Key Concepts

### Provider Configuration

Unlike the minimal example, the dashboard allows runtime configuration of the SDK connection. The user can:

1. Edit the server base path and access token in input fields.
2. Click "Get Dev Token" to fetch a JWT from the stub IdP.
3. Click "Apply Configuration" to establish or update the `PrefSyncProvider` connection.

The provider is conditionally subscribed based on the "Enable Real-Time SSE Subscription" toggle and the selected `appId`.

```tsx
<PrefSyncProvider
    config={appliedConfig}
    appId={enableSubscription && appId ? appId : undefined}
    onEvent={handleEvent}
    onError={handleError}
    onClose={handleClose}
>
    <PrefSyncDashboard ... />
</PrefSyncProvider>
```

### App Management

The dashboard supports creating and managing multiple apps:

- **Create App** — Enter a name and click "+ Create App" to call `client.createApp()`.
- **Select App** — Click on created app chips to switch the target app for operations.
- **Authorise App** — Click the "Authorise" button to call `client.authoriseApp()` for the selected app.

Created apps are tracked as selectable chips, making it easy to switch between them.

### Preference Operations

The dashboard provides a complete preference management workflow:

- **Get Preferences** — Fetch current preferences via `client.getPref()`.
- **Post Preferences** — Edit the JSON payload and save via `client.postPref()`.
- **Post via Peer Client** — Post from a separate client instance to trigger SSE `onEvent` notifications in the same tab.
- **Payload Templates** — Quick-fill buttons for common payloads (Theme, UI Layout, Timestamp).
- **Auto-refresh** — Toggle automatic preference refresh when SSE notifications arrive.

### SDK Client Explorer

A dedicated panel for directly invoking SDK client operations:

- `getUser()` — Fetch the current user profile.
- `getApp(appId)` — Fetch app details by ID.
- `authoriseApp(appId)` — Authorize an app to the current user.
- `deauthoriseApp(appId)` — Remove authorization for an app.

Results are displayed in a code block for inspection.

### Live Event Log

A filterable log panel records all activity in real time:

- **SSE events** — Real-time notifications received via `onEvent`.
- **API success/error** — Results from SDK client calls.
- **Peer API success** — Results from peer client posts.
- **Stream close** — When the SSE connection closes.

Logs can be filtered by type (All, Event, Error, Close, API) and cleared.

### Hook Usage
Please refer to the Documentation from the [Minimal Example](../minimal/README.md).

### Peer Client Pattern
Please refer to the Documentation from the [Minimal Example](../minimal/README.md).

