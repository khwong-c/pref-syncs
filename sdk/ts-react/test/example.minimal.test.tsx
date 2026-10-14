import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, fireEvent, render, screen, waitFor} from '@testing-library/react';
import React from 'react';
import {App} from '../example/minimal/App';

// -- Mock state for captured SSE callbacks --
let capturedOnEvent: ((n: any) => Promise<any> | any) | null = null;
let capturedOnError: ((e: any) => Promise<any> | any) | null = null;
let capturedOnClose: (() => Promise<any> | any) | null = null;

const mockSubscribeNotification = vi.fn().mockImplementation(
    async (
        appId: string,
        onEvent: (n: any) => Promise<any> | any,
        onError?: (err: any) => Promise<any> | any,
        onClose?: () => Promise<any> | any
    ) => {
        capturedOnEvent = onEvent;
        capturedOnError = onError ?? null;
        capturedOnClose = onClose ?? null;

        return {
            loop: new Promise<void>(() => {
            }),
            cancel: vi.fn().mockResolvedValue(undefined),
        };
    }
);

const mockClient = {
    getUser: vi.fn().mockResolvedValue({id: 'user-001', name: 'Test User'}),
    deleteUser: vi.fn().mockResolvedValue(undefined),
    getApp: vi.fn().mockImplementation(async (id: string) => ({
        id,
        name: 'Minimal Demo',
        desc: 'Auto-created for the minimal example',
    })),
    createApp: vi.fn().mockImplementation(async (name: string, desc: string) => ({
        id: 'test-app-id-001',
        name,
        desc,
    })),
    authoriseApp: vi.fn().mockResolvedValue({status: 'ok'}),
    deauthoriseApp: vi.fn().mockResolvedValue({status: 'ok'}),
    getPref: vi.fn().mockImplementation(async (appId: string) => ({
        user_id: 'user-001',
        app_id: appId,
        updated_at: '2026-10-14T19:00:00Z',
        data: {theme: 'light', greeting: 'Hello from minimal example'},
    })),
    postPref: vi.fn().mockImplementation(async (appId: string, payload: any) => ({
        user_id: 'user-001',
        app_id: appId,
        updated_at: '2026-10-14T19:00:00Z',
        data: payload,
    })),
    subscribeNotification: mockSubscribeNotification,
};

// Track all client instances created so the provider can detect client changes
const createdClients: any[] = [];

// Mock the pref-sync module — each call returns a fresh proxy that delegates to mockClient
// so the provider sees a different reference and re-runs its subscription effect.
vi.mock('pref-sync', async (importOriginal) => {
    const actual = await importOriginal<typeof import('pref-sync')>();
    return {
        ...actual,
        createPrefSyncClient: vi.fn().mockImplementation(() => {
            const proxy = new Proxy(mockClient, {
                get(target, prop) {
                    const val = (target as any)[prop];
                    if (typeof val === 'function') {
                        return val.bind(target);
                    }
                    return val;
                },
            });
            createdClients.push(proxy);
            return proxy;
        }),
    };
});

describe('Minimal Example (example/minimal/App.tsx)', () => {
    const originalFetch = globalThis.fetch;

    beforeEach(() => {
        vi.clearAllMocks();
        capturedOnEvent = null;
        capturedOnError = null;
        capturedOnClose = null;
        createdClients.length = 0;

        // Mock fetch for OIDC token endpoint
        globalThis.fetch = vi.fn().mockImplementation(async (url: string | URL | Request) => {
            const urlStr = typeof url === 'string' ? url : url.toString();
            if (urlStr.includes('/idp/oidc/token')) {
                return new Response(
                    JSON.stringify({
                        access_token: 'mock-jwt-token-minimal',
                        token_type: 'Bearer',
                        expires_in: 3600,
                    }),
                    {status: 200, headers: {'Content-Type': 'application/json'}}
                );
            }
            return new Response('Not found', {status: 404});
        });
    });

    afterEach(() => {
        globalThis.fetch = originalFetch;
        vi.restoreAllMocks();
    });

    it('renders the minimal example header after initialization', async () => {
        render(<App/>);

        // Initially shows loading state
        expect(screen.getByText('Initializing...')).toBeTruthy();

        // After async init completes, shows the dashboard
        await waitFor(() => {
            expect(screen.getByText(/Pref-Sync React SDK — Minimal Example/i)).toBeTruthy();
        });
    });

    it('displays the app ID and connected subscription status', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText(/Pref-Sync React SDK — Minimal Example/i)).toBeTruthy();
        });

        // App ID should be displayed
        expect(screen.getByText(/App ID:/i)).toBeTruthy();
        expect(screen.getByText('test-app-id-001')).toBeTruthy();

        // Subscription should show as Connected
        expect(screen.getByText(/Subscription:/i)).toBeTruthy();
        expect(screen.getByText('Connected')).toBeTruthy();
    });

    it('shows initialization error when fetch fails', async () => {
        globalThis.fetch = vi.fn().mockImplementation(async () => {
            return new Response('Unauthorized', {status: 401});
        });

        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText(/Initialization failed/i)).toBeTruthy();
        });
    });

    it('posts a preference from the peer client and logs the event', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText(/Pref-Sync React SDK — Minimal Example/i)).toBeTruthy();
        });

        // Click the "Post Preference from Peer" button
        const postButton = screen.getByRole('button', {name: /Post Preference from Peer/i});
        fireEvent.click(postButton);

        // Verify postPref was called on the peer client
        await waitFor(() => {
            expect(mockClient.postPref).toHaveBeenCalledWith(
                'test-app-id-001',
                expect.objectContaining({message: 'Posted from minimal example'})
            );
        });

        // Simulate receiving the SSE notification after the post
        await act(async () => {
            await capturedOnEvent!({
                AppID: 'test-app-id-001',
                User: 'user-001',
                Src: 'peer-client',
                Data: {timestamp: Date.now(), message: 'Posted from minimal example'},
            });
        });

        // The event log should now show at least one entry
        const eventLogHeaders = screen.getAllByText(/Event Log/i);
        expect(eventLogHeaders.length).toBeGreaterThan(0);
    });

    it('displays event log entries after SSE notifications', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText(/Pref-Sync React SDK — Minimal Example/i)).toBeTruthy();
        });

        // Initially shows "Waiting for events..."
        expect(screen.getByText(/Waiting for events/i)).toBeTruthy();

        // Trigger an SSE event
        await act(async () => {
            await capturedOnEvent!({
                AppID: 'test-app-id-001',
                User: 'user-001',
                Src: 'peer-client',
                Data: {theme: 'dark'},
            });
        });

        // Event log should show the count and the notification data
        await waitFor(() => {
            expect(screen.getByText(/Event Log \(1\)/i)).toBeTruthy();
        });
        expect(screen.getByText(/"AppID": "test-app-id-001"/i)).toBeTruthy();
    });

    it('unsubscribes the client and shows disconnected status', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText('Connected')).toBeTruthy();
        });

        // Click the Unsubscribe button
        const unsubscribeButton = screen.getByRole('button', {name: /Unsubscribe/i});
        fireEvent.click(unsubscribeButton);

        // After unsubscribing, status should change to Disconnected
        await waitFor(() => {
            expect(screen.getByText('Disconnected')).toBeTruthy();
        });

        // The Unsubscribe button should be disabled after unsubscribing
        expect((screen.getByRole('button', {name: /Unsubscribe/i}) as HTMLButtonElement).disabled).toBe(true);
    });

    it('enables re-create client button after unsubscribing', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText('Connected')).toBeTruthy();
        });

        // Re-create client should be disabled while subscribed
        expect((screen.getByRole('button', {name: /Re-create Client/i}) as HTMLButtonElement).disabled).toBe(true);

        // Unsubscribe first
        const unsubscribeButton = screen.getByRole('button', {name: /Unsubscribe/i});
        fireEvent.click(unsubscribeButton);

        await waitFor(() => {
            expect(screen.getByText('Disconnected')).toBeTruthy();
        });

        // Now re-create client should be enabled
        expect((screen.getByRole('button', {name: /Re-create Client/i}) as HTMLButtonElement).disabled).toBe(false);
    });

    it('re-creates the client and re-subscribes after unsubscribe + re-create', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText('Connected')).toBeTruthy();
        });

        // Unsubscribe
        fireEvent.click(screen.getByRole('button', {name: /Unsubscribe/i}));
        await waitFor(() => {
            expect(screen.getByText('Disconnected')).toBeTruthy();
        });

        // Re-create client
        fireEvent.click(screen.getByRole('button', {name: /Re-create Client/i}));

        // After re-creating, subscription should be re-established
        await waitFor(() => {
            expect(screen.getByText('Connected')).toBeTruthy();
        });
    });

    it('calls createPrefSyncClient with correct config during init', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText(/Pref-Sync React SDK — Minimal Example/i)).toBeTruthy();
        });

        // createPrefSyncClient should have been called during init
        expect(mockClient.createApp).toHaveBeenCalledWith(
            'Minimal Demo',
            'Auto-created for the minimal example'
        );
        expect(mockClient.authoriseApp).toHaveBeenCalledWith('test-app-id-001');
        expect(mockClient.postPref).toHaveBeenCalledWith(
            'test-app-id-001',
            {theme: 'light', greeting: 'Hello from minimal example'}
        );
    });

    it('handles SSE stream error and shows disconnected status', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText('Connected')).toBeTruthy();
        });

        // Trigger an SSE error
        await act(async () => {
            await capturedOnError!(new Error('Network stream error'));
        });

        // Should show disconnected after error
        await waitFor(() => {
            expect(screen.getByText('Disconnected')).toBeTruthy();
        });
    });

    it('handles SSE stream close and shows disconnected status', async () => {
        render(<App/>);

        await waitFor(() => {
            expect(screen.getByText('Connected')).toBeTruthy();
        });

        // Trigger an SSE close
        await act(async () => {
            await capturedOnClose!();
        });

        // Should show disconnected after close
        await waitFor(() => {
            expect(screen.getByText('Disconnected')).toBeTruthy();
        });
    });
});
