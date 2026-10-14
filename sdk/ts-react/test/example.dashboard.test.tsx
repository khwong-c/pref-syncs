import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import React from 'react';

// Create captured callback references so tests can trigger real-time SSE stream events
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
      loop: new Promise<void>(() => {}),
      cancel: vi.fn().mockResolvedValue(undefined),
    };
  }
);

const mockClient = {
  getUser: vi.fn().mockResolvedValue({ id: 'user-001', name: 'Test User' }),
  deleteUser: vi.fn().mockResolvedValue(undefined),
  getApp: vi.fn().mockImplementation(async (id: string) => ({
    id,
    name: 'Demo App',
    desc: 'Demo App Description',
  })),
  createApp: vi.fn().mockImplementation(async (name: string, desc: string) => ({
    id: '550e8400-e29b-41d4-a716-446655440000',
    name,
    desc,
  })),
  authoriseApp: vi.fn().mockResolvedValue({ status: 'ok' }),
  deauthoriseApp: vi.fn().mockResolvedValue({ status: 'ok' }),
  getPref: vi.fn().mockImplementation(async (appId: string) => ({
    user_id: 'user-001',
    app_id: appId,
    updated_at: '2026-10-14T19:00:00Z',
    data: { theme: 'dark', fontSize: 18, autoSave: true },
  })),
  postPref: vi.fn().mockImplementation(async (appId: string, payload: any) => ({
    user_id: 'user-001',
    app_id: appId,
    updated_at: '2026-10-14T19:00:00Z',
    data: payload,
  })),
  subscribeNotification: mockSubscribeNotification,
};

vi.mock('pref-sync', async (importOriginal) => {
  const actual = await importOriginal<typeof import('pref-sync')>();
  return {
    ...actual,
    createPrefSyncClient: vi.fn().mockImplementation(() => mockClient),
  };
});

import { App, fetchDevToken } from '../example/dashboard/App';

describe('Interactive Demo Application (example/App.tsx)', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.clearAllMocks();
    capturedOnEvent = null;
    capturedOnError = null;
    capturedOnClose = null;

    // Mock fetch for OIDC token endpoint
    globalThis.fetch = vi.fn().mockImplementation(async (url: string | URL | Request) => {
      const urlStr = typeof url === 'string' ? url : url.toString();
      if (urlStr.includes('/idp/oidc/token')) {
        return new Response(
          JSON.stringify({
            access_token: 'mock-secret-jwt-token-12345',
            token_type: 'Bearer',
            expires_in: 3600,
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        );
      }
      return new Response('Not found', { status: 404 });
    });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('fetchDevToken issues correct client_credentials POST request', async () => {
    const token = await fetchDevToken('http://127.0.0.1:7086');
    expect(token).toBe('mock-secret-jwt-token-12345');
    expect(globalThis.fetch).toHaveBeenCalledWith(
      'http://127.0.0.1:7086/idp/oidc/token',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/x-www-form-urlencoded',
        }),
      })
    );
  });

  it('renders header, configuration section, and controls with safe defaults', () => {
    render(<App />);

    expect(screen.getByText(/Preference Sync React Demo/i)).toBeTruthy();
    expect(screen.getByText(/Provider Configuration/i)).toBeTruthy();
    expect(screen.getByText(/Preference Manager/i)).toBeTruthy();
    expect(screen.getByText(/SDK Client Explorer/i)).toBeTruthy();
    expect(screen.getByText(/Live Event & Activity Log/i)).toBeTruthy();
    expect(screen.getByRole('button', { name: /Get Dev Token/i })).toBeTruthy();
    expect(screen.getByRole('button', { name: /Connected \/ Applied/i })).toBeTruthy();
  });

  it('allows fetching dev token and updating config without leaking token to logs', async () => {
    render(<App />);

    const getDevTokenButton = screen.getByRole('button', { name: /Get Dev Token/i });
    fireEvent.click(getDevTokenButton);

    await waitFor(() => {
      expect(screen.getByText(/Acquired developer JWT from stub IdP/i)).toBeTruthy();
    });

    // Verify token value is populated in password input
    const tokenInput = screen.getByPlaceholderText(/Bearer JWT token/i) as HTMLInputElement;
    expect(tokenInput.value).toBe('mock-secret-jwt-token-12345');
    expect(tokenInput.type).toBe('password');

    // Verify token is NOT logged in plain text in the log entries
    const logsContainer = screen.getByText(/Live Event & Activity Log/i).closest('section');
    expect(logsContainer?.textContent).not.toContain('mock-secret-jwt-token-12345');
  });

  it('allows editing draft config and applying without reconnecting on keystroke', () => {
    render(<App />);

    const basePathInput = screen.getByPlaceholderText(/http:\/\/localhost:5173/i) as HTMLInputElement;
    fireEvent.change(basePathInput, { target: { value: 'http://custom-host:9000' } });

    // Button indicates dirty state
    const applyButton = screen.getByRole('button', { name: /Apply Configuration/i });
    expect(applyButton.textContent).toContain('Apply Configuration');

    fireEvent.click(applyButton);
    expect(screen.getByText(/Applied client configuration/i)).toBeTruthy();
  });

  it('allows toggling token visibility', () => {
    render(<App />);

    const toggleButton = screen.getByRole('button', { name: /Show token/i });
    const tokenInput = screen.getByPlaceholderText(/Bearer JWT token/i) as HTMLInputElement;

    expect(tokenInput.type).toBe('password');
    fireEvent.click(toggleButton);
    expect(tokenInput.type).toBe('text');
    expect(screen.getByRole('button', { name: /Hide token/i })).toBeTruthy();
  });

  it('creates an app with UUID and selects it for preferences and subscription', async () => {
    render(<App />);

    const createAppInput = screen.getByPlaceholderText(/App Name/i);
    fireEvent.change(createAppInput, { target: { value: 'Production Workspace' } });

    const createAppButton = screen.getByRole('button', { name: /\+ Create App/i });
    fireEvent.click(createAppButton);

    await waitFor(() => {
      expect(mockClient.createApp).toHaveBeenCalledWith(
        'Production Workspace',
        'Demo application: Production Workspace'
      );
      expect(screen.getAllByText(/550e8400-e29b-41d4-a716-446655440000/i).length).toBeGreaterThan(0);
    });

    const targetIdInput = screen.getByPlaceholderText(/Target App UUID/i) as HTMLInputElement;
    expect(targetIdInput.value).toBe('550e8400-e29b-41d4-a716-446655440000');
  });

  it('authorises app successfully', async () => {
    render(<App />);

    const targetIdInput = screen.getByPlaceholderText(/Target App UUID/i);
    fireEvent.change(targetIdInput, { target: { value: '550e8400-e29b-41d4-a716-446655440000' } });

    const authoriseButton = screen.getByRole('button', { name: 'Authorise' });
    fireEvent.click(authoriseButton);

    await waitFor(() => {
      expect(mockClient.authoriseApp).toHaveBeenCalledWith('550e8400-e29b-41d4-a716-446655440000');
      expect(screen.getByText(/authoriseApp\("550e8400-e29b-41d4-a716-446655440000"\) succeeded/i)).toBeTruthy();
    });
  });

  it('fetches remote preferences and updates the view and logs', async () => {
    render(<App />);

    const targetIdInput = screen.getByPlaceholderText(/Target App UUID/i);
    fireEvent.change(targetIdInput, { target: { value: '550e8400-e29b-41d4-a716-446655440000' } });

    const getPrefButton = screen.getByRole('button', { name: /Get Preferences/i });
    fireEvent.click(getPrefButton);

    await waitFor(() => {
      expect(mockClient.getPref).toHaveBeenCalledWith('550e8400-e29b-41d4-a716-446655440000');
      expect(screen.getByText(/getPref\("550e8400-e29b-41d4-a716-446655440000"\) succeeded/i)).toBeTruthy();
      expect(screen.getByText(/Remote Preference Value/i)).toBeTruthy();
    });
  });

  it('posts preferences and updates logs', async () => {
    render(<App />);

    const targetIdInput = screen.getByPlaceholderText(/Target App UUID/i);
    fireEvent.change(targetIdInput, { target: { value: '550e8400-e29b-41d4-a716-446655440000' } });

    const postPrefButton = screen.getByRole('button', { name: /Post Preferences/i });
    fireEvent.click(postPrefButton);

    await waitFor(() => {
      expect(mockClient.postPref).toHaveBeenCalledWith(
        '550e8400-e29b-41d4-a716-446655440000',
        expect.objectContaining({ theme: 'dark' })
      );
      expect(screen.getByText(/postPref\("550e8400-e29b-41d4-a716-446655440000"\) succeeded/i)).toBeTruthy();
    });
  });

  it('posts via peer client to verify single-tab event delivery path', async () => {
    render(<App />);

    const targetIdInput = screen.getByPlaceholderText(/Target App UUID/i);
    fireEvent.change(targetIdInput, { target: { value: '550e8400-e29b-41d4-a716-446655440000' } });

    const peerPostButton = screen.getByRole('button', { name: /⚡ Post via Peer Client/i });
    fireEvent.click(peerPostButton);

    await waitFor(() => {
      expect(screen.getByText(/PeerClient\.postPref\("550e8400-e29b-41d4-a716-446655440000"\) succeeded/i)).toBeTruthy();
    });
  });

  it('records onEvent, onError, and onClose SSE callbacks in the live activity log', async () => {
    render(<App />);

    // Enable the SSE subscription toggle (disabled by default)
    const subscriptionCheckbox = screen.getByLabelText(/Enable Real-Time SSE Subscription/i) as HTMLInputElement;
    fireEvent.click(subscriptionCheckbox);

    // Create an app to trigger subscription
    const createAppButton = screen.getByRole('button', { name: /\+ Create App/i });
    fireEvent.click(createAppButton);

    await waitFor(() => {
      expect(mockSubscribeNotification).toHaveBeenCalled();
    });

    expect(capturedOnEvent).toBeTypeOf('function');
    expect(capturedOnError).toBeTypeOf('function');
    expect(capturedOnClose).toBeTypeOf('function');

    // Trigger onEvent
    await act(async () => {
      await capturedOnEvent!({
        App: '550e8400-e29b-41d4-a716-446655440000',
        User: 'user-001',
        Src: 'peer-src',
      });
    });

    expect(screen.getByText(/SSE Notification \[App: 550e8400-e29b-41d4-a716-446655440000, User: user-001\]/i)).toBeTruthy();

    // Trigger onError
    await act(async () => {
      await capturedOnError!(new Error('Test network stream disconnect'));
    });

    expect(screen.getByText(/SSE Stream Error/i)).toBeTruthy();
    expect(screen.getByText(/Test network stream disconnect/i)).toBeTruthy();

    // Trigger onClose
    await act(async () => {
      await capturedOnClose!();
    });

    expect(screen.getByText(/SSE Stream Closed/i)).toBeTruthy();
  });

  it('invokes SDK explorer operations (getUser, getApp, deauthoriseApp)', async () => {
    render(<App />);

    const getUserButton = screen.getByRole('button', { name: 'Get User' });
    fireEvent.click(getUserButton);

    await waitFor(() => {
      expect(mockClient.getUser).toHaveBeenCalled();
      expect(screen.getAllByText(/getUser\(\) succeeded/i).length).toBeGreaterThan(0);
    });
  });

  it('allows payload templates and log filtering/clearing', () => {
    render(<App />);

    const themeButton = screen.getByRole('button', { name: 'Theme' });
    const layoutButton = screen.getByRole('button', { name: 'UI Layout' });

    fireEvent.click(themeButton);
    expect(screen.getByDisplayValue(/autoSave/i)).toBeTruthy();

    fireEvent.click(layoutButton);
    expect(screen.getByDisplayValue(/itemsPerPage/i)).toBeTruthy();

    const clearButton = screen.getByRole('button', { name: /Clear Logs/i });
    fireEvent.click(clearButton);
    expect(screen.getByText(/No log events recorded yet/i)).toBeTruthy();
  });
});
