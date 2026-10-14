import type {ReactNode} from 'react';
import type {Config, Notification, PrefSyncClient} from 'pref-sync';

export {Notification, PrefSyncClient};

export interface PrefSyncCallbacks {
    /**
     * Invoked when a real-time preference update notification is received.
     * If this handler throws, the underlying SDK treats the throw as a stream error:
     * `onError` is invoked, then `onClose`, and the stream terminates.
     */
    onEvent?: (notification: Notification, client: PrefSyncClient) => Promise<void> | void;
    /**
     * Invoked when an error occurs during notification streaming.
     */
    onError?: (error: any, client: PrefSyncClient) => Promise<void> | void;
    /**
     * Invoked when the notification stream disconnects or terminates from the server or due to stream errors.
     * Note: This callback is intentionally suppressed during provider-initiated teardown
     * (such as component unmount or when switching appId/client).
     */
    onClose?: (client: PrefSyncClient) => Promise<void> | void;
}

export type PrefSyncProviderProps = PrefSyncCallbacks & {
    children: ReactNode;
    appId?: string;
    config?: Config;
    client?: PrefSyncClient;
}

export interface PrefSyncContextValue {
    /**
     * The active Preference Sync client instance.
     */
    client: PrefSyncClient;
    /**
     * The application ID currently being watched, if any.
     */
    appId?: string;
    /**
     * Whether an active notification stream subscription is currently established.
     */
    isSubscribed: boolean;
    /**
     * Manually stops the active notification stream subscription.
     * A no-op when no subscription is currently active.
     * After calling, the provider will not automatically re-subscribe unless
     * `appId` changes.
     */
    stopSubscription: () => void;
}
