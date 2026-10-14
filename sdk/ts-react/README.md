# pref-sync-react

React Context Provider and hooks for the [Preference Sync SDK](../ts). This library wraps the `pref-sync` client in a
React Context, making it easy to subscribe to real-time preference update notifications in React applications.

## Features

- **React Context integration** — Share a single `PrefSyncClient` across your component tree
- **Automatic subscription management** — Subscribe to preference notifications per `appId` with automatic connection
  cleanup
- **Real-time event callbacks** — Receive `onEvent`, `onError`, and `onClose` notifications
- **Manual subscription control** — Stop subscriptions programmatically via `stopSubscription()`
- **Stable client references** — Client instances are memoized and only recreated when config values change

## Installation

```bash
pnpm add pref-sync-react
# or
npm install pref-sync-react
# or
yarn add pref-sync-react
```

## Quick Start

Wrap your application (or a section of it) with `<PrefSyncProvider>`, then use the `usePrefSync` hook in any child
component.

```tsx
import {PrefSyncProvider, usePrefSync} from 'pref-sync-react';
import {createPrefSyncClient} from 'pref-sync-react';

function MyApp() {
    return (
        <PrefSyncProvider
            config={{
                basePath: 'https://api.example.com',
                accessToken: 'your-access-token',
            }}
            appId="my-app-id"
            onEvent={(notification, client) => {
                console.log('Preference update:', notification);
            }}
        >
            <Dashboard/>
        </PrefSyncProvider>
    );
}

function Dashboard() {
    const {client, appId, isSubscribed, stopSubscription} = usePrefSync();

    return (
        <div>
            <p>App: {appId}</p>
            <p>Status: {isSubscribed ? 'Connected' : 'Disconnected'}</p>
            <button onClick={stopSubscription}>Unsubscribe</button>
        </div>
    );
}
```

## API Reference

### `<PrefSyncProvider>`

A React Context Provider that manages a `PrefSyncClient` and its notification subscription lifecycle.

#### Props

| Prop       | Type                                                           | Required                    | Description                                                                                                        |
|------------|----------------------------------------------------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------------|
| `children` | `ReactNode`                                                    | Yes                         | Child components that can access the context via `usePrefSync()`                                                   |
| `config`   | `Config`                                                       | One of `config` or `client` | Configuration object used to create a client internally                                                            |
| `client`   | `PrefSyncClient`                                               | One of `config` or `client` | A pre-created client instance (useful for advanced control)                                                        |
| `appId`    | `string`                                                       | No                          | The application ID to subscribe to. Without it, no subscription is started                                         |
| `onEvent`  | `(notification: Notification, client: PrefSyncClient) => void` | No                          | Called when a real-time preference update is received                                                              |
| `onError`  | `(error: any, client: PrefSyncClient) => void`                 | No                          | Called when a stream error occurs                                                                                  |
| `onClose`  | `(client: PrefSyncClient) => void`                             | No                          | Called when the stream disconnects. Suppressed during provider-initiated teardown (unmount, `appId`/client switch) |

#### Config Object

| Field         | Type      | Required | Description                                                      |
|---------------|-----------|----------|------------------------------------------------------------------|
| `basePath`    | `string`  | Yes      | Base URL of the Preference Sync server                           |
| `accessToken` | `string`  | Yes      | Access token for authentication                                  |
| `src`         | `string`  | No       | Identifier for this client, used to distinguish multiple clients |
| `autoSrc`     | `boolean` | No       | Auto-generate a source identifier. Defaults to `true`            |

#### Provider Modes

You can initialize the provider in two ways:

**1. Config mode** — Provider creates and manages the client:

```tsx
<PrefSyncProvider
    config={{
        basePath: 'https://api.example.com',
        accessToken: 'your-token',
    }}
    appId="my-app-uuid"
>
    <App/>
</PrefSyncProvider>
```

**2. Direct client mode** — You create the client yourself:

```tsx
const client = createPrefSyncClient({
    basePath: 'https://api.example.com',
    accessToken: 'your-token',
});

<PrefSyncProvider client={client} appId="my-app-uuid">
    <App/>
</PrefSyncProvider>
```

### `usePrefSync()`

A hook that returns the current Preference Sync context. Must be used within a `<PrefSyncProvider>`.

#### Returns

| Property           | Type                  | Description                                                                                                                  |
|--------------------|-----------------------|------------------------------------------------------------------------------------------------------------------------------|
| `client`           | `PrefSyncClient`      | The active Preference Sync client instance                                                                                   |
| `appId`            | `string \| undefined` | The application ID currently being watched                                                                                   |
| `isSubscribed`     | `boolean`             | Whether an active notification stream subscription is established                                                            |
| `stopSubscription` | `() => void`          | Manually stop the active subscription. No-op when not subscribed. Provider will not auto-re-subscribe unless `appId` changes |

#### Usage

```tsx
const {client, appId, isSubscribed, stopSubscription} = usePrefSync();
```

Throws an error if called outside `<PrefSyncProvider>`.

### Callback Behavior

- **`onEvent`** — Receives every preference update notification. If this handler throws, the SDK treats it as a stream
  error: `onError` is invoked, then `onClose`, and the stream terminates.
- **`onError`** — Called when a stream error occurs. `isSubscribed` is set to `false`.
- **`onClose`** — Called when the stream disconnects. Intentionally suppressed during provider-initiated teardown
  (unmount, `appId`/client switch) to avoid stale callbacks.

## Examples

See the [`example/`](example/) directory for working demos:

- **[Minimal](example/minimal/)** — A simple example showing subscription, event logging, manual unsubscribe, and client
  recreation. Quick recipe for integration
- **[Dashboard](example/dashboard/)** — A more comprehensive dashboard example

Run the examples locally:

```bash
pnpm dev:minimal   # Opens the minimal example
pnpm dev:dashboard # Opens the dashboard example
```

## Development

```bash
pnpm install        # Install dependencies
pnpm build          # Build the package
pnpm typecheck      # Type-check the source
pnpm test           # Run tests
pnpm test:watch     # Run tests in watch mode
```

## License

MIT
