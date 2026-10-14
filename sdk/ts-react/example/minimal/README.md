# Minimal Example

A minimal React application demonstrating the `PrefSyncProvider` and `usePrefSync` hook from the `pref-sync-react` SDK.

## What It Demonstrates

This example shows the essential patterns for integrating preference synchronization into a React application:

- **Client initialization** — Fetching a JWT token, creating a `PrefSyncClient`, creating and authorizing an app, and
  posting initial preferences.
- **Provider integration** — Wrapping the application with `PrefSyncProvider` to manage client state and SSE
  subscriptions.
- **Hook usage** — Using `usePrefSync` to access the current `appId` and subscription status.
- **Event handling** — Receiving real-time preference update notifications via the `onEvent` callback.
- **Peer client posting** — Simulating external preference updates by posting from a separate client instance.
- **Subscription lifecycle** — Unsubscribing from and re-subscribing to real-time updates.

## How to Run

From the `sdk/ts-react` root directory:

```bash
# Install dependencies
pnpm install

# Start this minimal example
pnpm dev:minimal
```

The example serves on `http://localhost:5173` by default (configured in `vite.config.ts`).

### Prerequisites

The Vite dev server proxies API routes (`/app`, `/user`, `/pref`, `/idp`) to `http://127.0.0.1:7086`.
Ensure the PrefSync backend service is running before starting the examples.

## Files

| File         | Description                                                                                 |
|--------------|---------------------------------------------------------------------------------------------|
| `App.tsx`    | Main React component with `PrefSyncProvider` and demo UI                                    |
| `init.ts`    | Auto-initialization routine: fetches JWT, creates app, authorizes, posts initial preference |
| `main.tsx`   | React entry point that mounts the `App` component                                           |
| `style.ts`   | Inline styles for the demo layout                                                           |
| `index.html` | HTML template for the Vite dev server                                                       |

## Key Concepts

### Initialization Flow

The `initExample` function in `init.ts` performs the following steps:

1. Fetches a JWT token from the stub IdP via `client_credentials` grant.
2. Creates a `PrefSyncClient` with the token and base path.
3. Creates a new demo app via `client.createApp()`.
4. Authorizes the app via `client.authoriseApp()`.
5. Posts an initial preference via `client.postPref()`.

The returned `InitResult` contains the client, configuration, and `appId` needed by `PrefSyncProvider`.

### Provider Usage

```tsx
<PrefSyncProvider
    client={client}
    config={config}
    appId={appId}
    onEvent={(notification) => {/* handle SSE event */}}>
    <Dashboard/>
</PrefSyncProvider>
```

The provider manages the SSE subscription lifecycle automatically. When the `client` prop changes, it re-subscribes with
the new client.

### Hook Usage

```tsx
const {client, appId, isSubscribed, stopSubscription} = usePrefSync();
```

The `usePrefSync` hook provides access to:

- `client` — The active `PrefSyncClient` instance for making API calls.
- `appId` — The currently active application ID.
- `isSubscribed` — Whether the SSE subscription is active.
- `stopSubscription` — Function to manually stop the subscription.

### Peer Client

To simulate real-time events during development, a second `PrefSyncClient` instance (the "peer client") is used to post
preference updates. This triggers SSE notifications that the main client receives via the `onEvent` callback,
demonstrating the end-to-end event delivery path.
