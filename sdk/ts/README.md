# Preference Sync TypeScript SDK (`pref-sync`)

TypeScript client SDK for the Preference Sync API. This library provides a high-level client for managing users, applications, preferences, and real-time preference change notifications via Server-Sent Events (SSE).

---

## Installation

Install via your preferred package manager:

```bash
# pnpm
pnpm add pref-sync

# npm
npm install pref-sync

# yarn
yarn add pref-sync
```

---

## Configuration

Initialize the client with the `Config` interface:

```typescript
export interface Config {
  basePath: string;     // Base URL of the Preference Sync API server
  accessToken: string;  // Bearer authentication token
  src?: string;         // Optional source identifier for client tracking
  autoSrc?: boolean;    // Automatically generate a random hex source ID if src is not provided (default: true)
}
```

### Source Tracking (`src` and `autoSrc`)

- **`autoSrc` (default: `true`)**: When enabled and `src` is not explicitly provided, the SDK generates a random 4-byte hex identifier. This allows the API to distinguish distinct client instances and avoid echoing updates back to the sender during notification streaming.
- **`src`**: Manually specify an identifier string.
- If `autoSrc` is set to `false` and `src` is not provided, requests will be sent without a source ID.

---

## Quick Start

```typescript
import { createPrefSyncClient } from "pref-sync";

async function run() {
  const client = createPrefSyncClient({
    basePath: "https://api.example.com",
    accessToken: "your-access-token",
  });
  const remoteClient = createPrefSyncClient({
    basePath: "https://api.example.com",
    accessToken: "your-access-token",
  });

  // 1. Create and authorize an application
  const app = await client.createApp("My App", "Configuration sync for My App");
  await client.authoriseApp(app.id);

  // 2. Subscribe to real-time preference change notifications
  const { loop, cancel } = await remoteClient.subscribeNotification(app.id, async (notification) => {
    console.log("Preference updated:", notification);
    const updatedPref = await remoteClient.getPref(app.id);
    console.log("Current preferences:", updatedPref);
  });

  // 3. Post a new preference
  await client.postPref(app.id, {
    theme: "dark",
    fontSize: 14,
  });

  // 4. Retrieve current preferences
  const pref = await client.getPref(app.id);
  console.log("Preferences:", pref);

  // 5. Clean up subscription when done
  await cancel();
  await loop;
}
```

---

## API Reference

### Client Factory

#### `createPrefSyncClient(config: Config)`

Creates and returns a client instance configured with the specified parameters.

---

### User Operations

#### `client.getUser()`
- Retrieves information about the authenticated user and applications authorized.

#### `client.deleteUser()`
- Deletes the current authenticated user account with all its preferences and applications created.

---

### Application Operations

#### `client.createApp(name: string, desc: string)`
- **Parameters**:
  - `name`: Name of the application.
  - `desc`: Description of the application.
- Registers a new application.
- Required Admin / App Creator Authority.

#### `client.getApp(id: string)`
- **Parameters**:
  - `id`: Unique application identifier.
- Fetches details of a specific application.

#### `client.authoriseApp(appID: string)`
- **Parameters**:
  - `appID`: ID of the application to authorize.
- Authorizes the specified application for the authenticated user.

#### `client.deauthoriseApp(appID: string)`
- **Parameters**:
  - `appID`: ID of the application to deauthorize.
- Revokes authorization for the specified application.

---

### Preferences Operations

#### `client.getPref(appID: string)`
- **Parameters**:
  - `appID`: Application ID.
- Retrieves the saved preferences for the specified application.

#### `client.postPref(appID: string, payload: any)`
- **Parameters**:
  - `appID`: Application ID.
  - `payload`: Preference data to save.
- Updates or creates preferences for the given application. Automatically includes source tracking if configured.
- Other clients will receive an update event.

---

### Real-time Notifications

#### `client.subscribeNotification(appID: string, callback: (n: Notification) => Promise<any>, onError?: (err: any) => Promise<any> | any, onClose?: () => Promise<any> | any)`
- **Parameters**:
  - `appID`: Application ID to watch.
  - `onEvent`: Async handler invoked when a preference change notification is received.
  - `onError`: Optional handler invoked when an error occurs during notification streaming.
  - `onClose`: Optional handler invoked when the stream connection is closed or terminated.
- **Returns**: `Promise<{ loop: Promise<void>, cancel: () => Promise<void> }>`
  - `loop`: A promise that stays active while listening to the SSE stream and resolves when the connection is closed or disconnected (or after being handled by `onError`).
  - `cancel`: A function to cancel the stream reader and terminate the subscription.

---

## Development & Build

```bash
# Install dependencies
pnpm install -d

# Type-check TypeScript sources
pnpm run typecheck

# Build compiled JavaScript output
pnpm run build

# Re-generate API client code from OpenAPI specification
pnpm run generate
```
