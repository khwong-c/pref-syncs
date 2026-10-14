import { afterEach, describe, it, expect, vi } from 'vitest';
import { render, screen, renderHook, fireEvent, act } from '@testing-library/react';
import React, { useState } from 'react';
import { createPrefSyncClient, PrefSyncProvider, usePrefSync } from '../src';
import type { Config, PrefSyncClient } from '../src';

describe('PrefSyncContext and usePrefSync', () => {
  it('throws a descriptive error when usePrefSync is called outside of PrefSyncProvider', () => {
    // Suppress React console.error during expected throw
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      renderHook(() => usePrefSync());
    }).toThrow('usePrefSync must be used within a PrefSyncProvider.');

    consoleErrorSpy.mockRestore();
  });

  it('provides client and state when initialized with a direct client prop', () => {
    const customClient = createPrefSyncClient({
      basePath: 'http://localhost:8080',
      accessToken: 'test-token',
    });

    const Consumer = () => {
      const { client, appId, isSubscribed } = usePrefSync();
      return (
        <div>
          <span data-testid="is-same-client">{String(client === customClient)}</span>
          <span data-testid="app-id">{appId ?? 'none'}</span>
          <span data-testid="is-subscribed">{String(isSubscribed)}</span>
        </div>
      );
    };

    render(
      <PrefSyncProvider client={customClient} appId="app-123">
        <Consumer />
      </PrefSyncProvider>
    );

    expect(screen.getByTestId('is-same-client').textContent).toBe('true');
    expect(screen.getByTestId('app-id').textContent).toBe('app-123');
    expect(screen.getByTestId('is-subscribed').textContent).toBe('false');
  });

  it('instantiates and provides client when initialized with a config prop', () => {
    const Consumer = () => {
      const { client } = usePrefSync();
      return <span data-testid="has-client">{String(typeof client?.getUser === 'function')}</span>;
    };

    const config: Config = {
      basePath: 'http://localhost:8080',
      accessToken: 'test-token',
    };

    render(
      <PrefSyncProvider config={config}>
        <Consumer />
      </PrefSyncProvider>
    );

    expect(screen.getByTestId('has-client').textContent).toBe('true');
  });

  it('preserves stable client reference across re-renders with fresh but equal config objects', () => {
    let capturedClients: PrefSyncClient[] = [];

    const Consumer = () => {
      const { client } = usePrefSync();
      capturedClients.push(client);
      return <span data-testid="client-rendered">rendered</span>;
    };

    const TestApp = () => {
      const [, setTick] = useState(0);
      return (
        <div>
          <button onClick={() => setTick((t) => t + 1)}>Rerender</button>
          <PrefSyncProvider
            config={{
              basePath: 'http://localhost:8080',
              accessToken: 'test-token',
              src: 'src-1',
            }}
          >
            <Consumer />
          </PrefSyncProvider>
        </div>
      );
    };

    render(<TestApp />);
    expect(capturedClients.length).toBe(1);

    fireEvent.click(screen.getByText('Rerender'));
    expect(capturedClients.length).toBe(2);
    expect(capturedClients[0]).toBe(capturedClients[1]);
  });

  it('does not mutate frozen config objects passed by caller', () => {
    const frozenConfig = Object.freeze({
      basePath: 'http://localhost:8080',
      accessToken: 'test-token',
    });

    expect(() => {
      render(
        <PrefSyncProvider config={frozenConfig}>
          <div>Child</div>
        </PrefSyncProvider>
      );
    }).not.toThrow();
  });

  it('creates a new client when config values change', () => {
    let capturedClients: PrefSyncClient[] = [];

    const Consumer = () => {
      const { client } = usePrefSync();
      capturedClients.push(client);
      return <span>rendered</span>;
    };

    const TestApp = () => {
      const [token, setToken] = useState('token-1');
      return (
        <div>
          <button onClick={() => setToken('token-2')}>Change Token</button>
          <PrefSyncProvider config={{ basePath: 'http://localhost:8080', accessToken: token }}>
            <Consumer />
          </PrefSyncProvider>
        </div>
      );
    };

    render(<TestApp />);
    expect(capturedClients.length).toBe(1);

    fireEvent.click(screen.getByText('Change Token'));
    expect(capturedClients.length).toBe(2);
    expect(capturedClients[0]).not.toBe(capturedClients[1]);
  });

  it('creates a new client when basePath or src change', () => {
    let capturedClients: PrefSyncClient[] = [];

    const Consumer = () => {
      const { client } = usePrefSync();
      capturedClients.push(client);
      return <span>rendered</span>;
    };

    const TestApp = () => {
      const [basePath, setBasePath] = useState('http://localhost:8080');
      const [src, setSrc] = useState<string | undefined>(undefined);
      return (
        <div>
          <button onClick={() => setBasePath('http://localhost:9090')}>Change BasePath</button>
          <button onClick={() => setSrc('custom-src')}>Change Src</button>
          <PrefSyncProvider config={{ basePath, accessToken: 'token', src }}>
            <Consumer />
          </PrefSyncProvider>
        </div>
      );
    };

    render(<TestApp />);
    expect(capturedClients.length).toBe(1);

    fireEvent.click(screen.getByText('Change BasePath'));
    expect(capturedClients.length).toBe(2);
    expect(capturedClients[1]).not.toBe(capturedClients[0]);

    fireEvent.click(screen.getByText('Change Src'));
    expect(capturedClients.length).toBe(3);
    expect(capturedClients[2]).not.toBe(capturedClients[1]);
  });

  it('treats autoSrc undefined and true as equal, and creates new client for autoSrc false', () => {
    let capturedClients: PrefSyncClient[] = [];

    const Consumer = () => {
      const { client } = usePrefSync();
      capturedClients.push(client);
      return <span>rendered</span>;
    };

    const TestApp = () => {
      const [autoSrc, setAutoSrc] = useState<boolean | undefined>(undefined);
      return (
        <div>
          <button onClick={() => setAutoSrc(true)}>Set AutoSrc True</button>
          <button onClick={() => setAutoSrc(false)}>Set AutoSrc False</button>
          <PrefSyncProvider config={{ basePath: 'http://localhost:8080', accessToken: 'token', autoSrc }}>
            <Consumer />
          </PrefSyncProvider>
        </div>
      );
    };

    render(<TestApp />);
    expect(capturedClients.length).toBe(1);

    // autoSrc undefined -> true (both normalize to true) => client reference should remain stable
    fireEvent.click(screen.getByText('Set AutoSrc True'));
    expect(capturedClients.length).toBe(2);
    expect(capturedClients[1]).toBe(capturedClients[0]);

    // autoSrc true -> false => new client instance
    fireEvent.click(screen.getByText('Set AutoSrc False'));
    expect(capturedClients.length).toBe(3);
    expect(capturedClients[2]).not.toBe(capturedClients[1]);
  });

  it('correctly handles switching between config and client modes, and between different client instances', () => {
    let capturedClients: PrefSyncClient[] = [];

    const Consumer = () => {
      const { client } = usePrefSync();
      capturedClients.push(client);
      return <span>rendered</span>;
    };

    const clientA = createPrefSyncClient({
      basePath: 'http://localhost:8080',
      accessToken: 'token-a',
    });
    const clientB = createPrefSyncClient({
      basePath: 'http://localhost:8080',
      accessToken: 'token-b',
    });

    const TestApp = () => {
      const [mode, setMode] = useState<'config' | 'clientA' | 'clientB'>('config');
      return (
        <div>
          <button onClick={() => setMode('clientA')}>Use Client A</button>
          <button onClick={() => setMode('clientB')}>Use Client B</button>
          <button onClick={() => setMode('config')}>Use Config</button>
          {mode === 'clientA' && (
            <PrefSyncProvider client={clientA}>
              <Consumer />
            </PrefSyncProvider>
          )}
          {mode === 'clientB' && (
            <PrefSyncProvider client={clientB}>
              <Consumer />
            </PrefSyncProvider>
          )}
          {mode === 'config' && (
            <PrefSyncProvider config={{ basePath: 'http://localhost:8080', accessToken: 'config-token' }}>
              <Consumer />
            </PrefSyncProvider>
          )}
        </div>
      );
    };

    render(<TestApp />);
    expect(capturedClients.length).toBe(1);
    const initialConfigClient = capturedClients[0];
    expect(initialConfigClient).not.toBe(clientA);
    expect(initialConfigClient).not.toBe(clientB);

    fireEvent.click(screen.getByText('Use Client A'));
    expect(capturedClients.length).toBe(2);
    expect(capturedClients[1]).toBe(clientA);

    fireEvent.click(screen.getByText('Use Client B'));
    expect(capturedClients.length).toBe(3);
    expect(capturedClients[2]).toBe(clientB);

    fireEvent.click(screen.getByText('Use Config'));
    expect(capturedClients.length).toBe(4);
    expect(capturedClients[3]).not.toBe(clientA);
    expect(capturedClients[3]).not.toBe(clientB);
  });
});

describe('PrefSyncProvider SSE Subscription Lifecycle', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  const createMockClient = () => {
    let capturedOnEvent: ((n: any) => Promise<any>) | undefined;
    let capturedOnError: ((e: any) => Promise<any>) | undefined;
    let capturedOnClose: (() => Promise<any>) | undefined;
    let cancelCalled = false;

    const cancelMock = vi.fn().mockImplementation(async () => {
      cancelCalled = true;
    });

    const loopPromise = new Promise<void>(() => {});

    const subscribeNotification = vi.fn().mockImplementation(
      async (
        appId: string,
        onEvent: (n: any) => Promise<any>,
        onError?: (e: any) => Promise<any>,
        onClose?: () => Promise<any>
      ) => {
        capturedOnEvent = onEvent;
        capturedOnError = onError;
        capturedOnClose = onClose;
        return {
          loop: loopPromise,
          cancel: cancelMock,
        };
      }
    );

    const client = {
      getUser: vi.fn(),
      deleteUser: vi.fn(),
      getApp: vi.fn(),
      createApp: vi.fn(),
      authoriseApp: vi.fn(),
      deauthoriseApp: vi.fn(),
      getPref: vi.fn(),
      postPref: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    return {
      client,
      subscribeNotification,
      cancelMock,
      triggerEvent: async (n: any) => capturedOnEvent?.(n),
      triggerError: async (e: any) => capturedOnError?.(e),
      triggerClose: async () => capturedOnClose?.(),
    };
  };

  it('does not start subscription when appId is not provided', async () => {
    const { client, subscribeNotification } = createMockClient();

    const Consumer = () => {
      const { isSubscribed, appId } = usePrefSync();
      return (
        <div>
          <span data-testid="is-sub">{String(isSubscribed)}</span>
          <span data-testid="app-id">{appId ?? 'none'}</span>
        </div>
      );
    };

    render(
      <PrefSyncProvider client={client}>
        <Consumer />
      </PrefSyncProvider>
    );

    expect(subscribeNotification).not.toHaveBeenCalled();
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
    expect(screen.getByTestId('app-id').textContent).toBe('none');
  });

  it('starts subscription and sets isSubscribed to true when appId is provided', async () => {
    const { client, subscribeNotification } = createMockClient();

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc">
        <Consumer />
      </PrefSyncProvider>
    );

    expect(subscribeNotification).toHaveBeenCalledWith(
      'app-abc',
      expect.any(Function),
      expect.any(Function),
      expect.any(Function)
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');
  });

  it('forwards onEvent notifications to callback with notification and client', async () => {
    const { client, triggerEvent } = createMockClient();
    const onEvent = vi.fn();

    render(
      <PrefSyncProvider client={client} appId="app-abc" onEvent={onEvent}>
        <div>Child</div>
      </PrefSyncProvider>
    );

    await act(async () => {});

    const notificationPayload = {
      Time: '2026-10-14T12:00:00Z',
      Src: 'src-123',
      AppID: 'app-abc',
    };

    await act(async () => {
      await triggerEvent(notificationPayload);
    });

    expect(onEvent).toHaveBeenCalledTimes(1);
    expect(onEvent).toHaveBeenCalledWith(notificationPayload, client);
  });

  it('forwards stream errors to onError, sets isSubscribed to false, and works with synchronous onError', async () => {
    const { client, triggerError } = createMockClient();
    const onError = vi.fn();

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc" onError={onError}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    const streamError = new Error('Stream connection lost');
    await act(async () => {
      await triggerError(streamError);
    });

    expect(onError).toHaveBeenCalledTimes(1);
    expect(onError).toHaveBeenCalledWith(streamError, client);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
  });

  it('forwards stream close events to onClose and sets isSubscribed to false', async () => {
    const { client, triggerClose } = createMockClient();
    const onClose = vi.fn();

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc" onClose={onClose}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    await act(async () => {
      await triggerClose();
    });

    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onClose).toHaveBeenCalledWith(client);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
  });

  it('maintains active subscription without reconnecting when callbacks change across renders', async () => {
    const { client, subscribeNotification, cancelMock, triggerEvent } = createMockClient();
    const onEvent1 = vi.fn();
    const onEvent2 = vi.fn();

    const TestApp = ({ onEvent }: { onEvent: any }) => {
      return (
        <PrefSyncProvider client={client} appId="app-abc" onEvent={onEvent}>
          <div>Child</div>
        </PrefSyncProvider>
      );
    };

    const { rerender } = render(<TestApp onEvent={onEvent1} />);
    await act(async () => {});

    expect(subscribeNotification).toHaveBeenCalledTimes(1);
    expect(cancelMock).not.toHaveBeenCalled();

    // Rerender with fresh callback
    rerender(<TestApp onEvent={onEvent2} />);
    await act(async () => {});

    // Stream must NOT have reconnected or cancelled
    expect(subscribeNotification).toHaveBeenCalledTimes(1);
    expect(cancelMock).not.toHaveBeenCalled();

    // Triggering event should call the new onEvent2 callback
    await act(async () => {
      await triggerEvent({ AppID: 'app-abc' });
    });

    expect(onEvent1).not.toHaveBeenCalled();
    expect(onEvent2).toHaveBeenCalledTimes(1);
    expect(onEvent2).toHaveBeenCalledWith({ AppID: 'app-abc' }, client);
  });

  it('cancels active stream and suppresses stale onClose/onError when component unmounts', async () => {
    const { client, cancelMock, triggerClose, triggerError } = createMockClient();
    const onClose = vi.fn();
    const onError = vi.fn();

    const { unmount } = render(
      <PrefSyncProvider client={client} appId="app-abc" onClose={onClose} onError={onError}>
        <div>Child</div>
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(cancelMock).not.toHaveBeenCalled();

    unmount();
    expect(cancelMock).toHaveBeenCalledTimes(1);

    // Simulate SDK's loop.finally calling onClose/onError on cancellation
    await act(async () => {
      await triggerClose();
      await triggerError(new Error('post-cancel error'));
    });

    // Stale callbacks must NOT be invoked on unmount
    expect(onClose).not.toHaveBeenCalled();
    expect(onError).not.toHaveBeenCalled();
  });

  it('cancels previous subscription and starts new subscription when appId changes', async () => {
    const { client, subscribeNotification, cancelMock } = createMockClient();

    const TestApp = ({ appId }: { appId?: string }) => {
      return (
        <PrefSyncProvider client={client} appId={appId}>
          <div>Child</div>
        </PrefSyncProvider>
      );
    };

    const { rerender } = render(<TestApp appId="app-1" />);
    await act(async () => {});

    expect(subscribeNotification).toHaveBeenCalledTimes(1);
    expect(subscribeNotification).toHaveBeenLastCalledWith('app-1', expect.any(Function), expect.any(Function), expect.any(Function));
    expect(cancelMock).not.toHaveBeenCalled();

    // Switch to app-2
    rerender(<TestApp appId="app-2" />);
    await act(async () => {});

    expect(cancelMock).toHaveBeenCalledTimes(1);
    expect(subscribeNotification).toHaveBeenCalledTimes(2);
    expect(subscribeNotification).toHaveBeenLastCalledWith('app-2', expect.any(Function), expect.any(Function), expect.any(Function));

    // Clear appId
    rerender(<TestApp appId={undefined} />);
    await act(async () => {});

    expect(cancelMock).toHaveBeenCalledTimes(2);
    expect(subscribeNotification).toHaveBeenCalledTimes(2);
  });

  it('cancels pending acquisition if component unmounts before subscribeNotification resolves', async () => {
    let resolveSubscription!: (value: any) => void;
    const cancelMock = vi.fn().mockResolvedValue(undefined);
    const onClose = vi.fn();

    const subscribeNotification = vi.fn().mockImplementation(() => {
      return new Promise((resolve) => {
        resolveSubscription = resolve;
      });
    });

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const { unmount } = render(
      <PrefSyncProvider client={client} appId="app-abc" onClose={onClose}>
        <div>Child</div>
      </PrefSyncProvider>
    );

    expect(subscribeNotification).toHaveBeenCalledWith('app-abc', expect.any(Function), expect.any(Function), expect.any(Function));

    // Unmount before resolution
    unmount();

    // Late resolution
    await act(async () => {
      resolveSubscription({
        loop: Promise.resolve(),
        cancel: cancelMock,
      });
    });

    // Should immediately cancel late resolution
    expect(cancelMock).toHaveBeenCalledTimes(1);
    expect(onClose).not.toHaveBeenCalled();
  });

  it('cancels pending acquisition for old appId when appId changes before resolution', async () => {
    let resolveSub1!: (value: any) => void;
    const cancel1 = vi.fn().mockResolvedValue(undefined);
    const cancel2 = vi.fn().mockResolvedValue(undefined);

    const subscribeNotification = vi.fn().mockImplementation((appId: string) => {
      if (appId === 'app-1') {
        return new Promise((resolve) => {
          resolveSub1 = resolve;
        });
      }
      return Promise.resolve({
        loop: Promise.resolve(),
        cancel: cancel2,
      });
    });

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const TestApp = ({ appId }: { appId: string }) => {
      return (
        <PrefSyncProvider client={client} appId={appId}>
          <div>Child</div>
        </PrefSyncProvider>
      );
    };

    const { rerender } = render(<TestApp appId="app-1" />);
    expect(subscribeNotification).toHaveBeenCalledWith('app-1', expect.any(Function), expect.any(Function), expect.any(Function));

    // Switch appId before app-1 resolves
    rerender(<TestApp appId="app-2" />);
    await act(async () => {});

    expect(subscribeNotification).toHaveBeenCalledWith('app-2', expect.any(Function), expect.any(Function), expect.any(Function));

    // Resolve app-1 late
    await act(async () => {
      resolveSub1({
        loop: Promise.resolve(),
        cancel: cancel1,
      });
    });

    expect(cancel1).toHaveBeenCalledTimes(1);
    expect(cancel2).not.toHaveBeenCalled();
  });

  it('handles subscribeNotification initial rejection gracefully via onError', async () => {
    const setupError = new Error('Network failure connecting to SSE');
    const subscribeNotification = vi.fn().mockRejectedValue(setupError);
    const onError = vi.fn();

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc" onError={onError}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});

    expect(onError).toHaveBeenCalledTimes(1);
    expect(onError).toHaveBeenCalledWith(setupError, client);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
  });

  it('marks isSubscribed as false and safeguards against loop rejection without causing unhandled errors', async () => {
    let rejectLoop!: (reason?: any) => void;
    const loopPromise = new Promise<void>((_, reject) => {
      rejectLoop = reject;
    });

    const subscribeNotification = vi.fn().mockResolvedValue({
      loop: loopPromise,
      cancel: vi.fn(),
    });

    const onError = vi.fn();
    const onClose = vi.fn();

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc" onError={onError} onClose={onClose}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    // Reject loop
    await act(async () => {
      rejectLoop(new Error('Loop unhandled failure'));
    });

    // Settled loop marks isSubscribed as false and suppresses unhandled promise rejections
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
    expect(onError).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it('marks isSubscribed as false when loop resolves', async () => {
    let resolveLoop!: () => void;
    const loopPromise = new Promise<void>((resolve) => {
      resolveLoop = resolve;
    });

    const subscribeNotification = vi.fn().mockResolvedValue({
      loop: loopPromise,
      cancel: vi.fn(),
    });

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <PrefSyncProvider client={client} appId="app-resolve">
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    await act(async () => {
      resolveLoop();
    });

    expect(screen.getByTestId('is-sub').textContent).toBe('false');
  });

  it('guarantees isSubscribed is false during render immediately when appId or client switches until resolution', async () => {
    let resolveApp2!: (value: any) => void;
    const seenContexts: { appId?: string; isSubscribed: boolean }[] = [];

    const { client } = createMockClient();
    const cancel1 = vi.fn().mockResolvedValue(undefined);
    const cancel2 = vi.fn().mockResolvedValue(undefined);

    (client.subscribeNotification as any).mockImplementation((appId: string) => {
      if (appId === 'app-1') {
        return Promise.resolve({
          loop: new Promise(() => {}),
          cancel: cancel1,
        });
      }
      return new Promise((resolve) => {
        resolveApp2 = () =>
          resolve({
            loop: new Promise(() => {}),
            cancel: cancel2,
          });
      });
    });

    const Consumer = () => {
      const { appId, isSubscribed } = usePrefSync();
      seenContexts.push({ appId, isSubscribed });
      return (
        <div>
          <span data-testid="app-id">{appId}</span>
          <span data-testid="is-sub">{String(isSubscribed)}</span>
        </div>
      );
    };

    const TestApp = ({ appId }: { appId: string }) => {
      return (
        <PrefSyncProvider client={client} appId={appId}>
          <Consumer />
        </PrefSyncProvider>
      );
    };

    const { rerender } = render(<TestApp appId="app-1" />);
    await act(async () => {});

    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    // Switch to app-2
    seenContexts.length = 0;
    rerender(<TestApp appId="app-2" />);

    // On the initial render of app-2, isSubscribed MUST be false (no stale true from app-1)
    expect(seenContexts[0]).toEqual({ appId: 'app-2', isSubscribed: false });
    expect(screen.getByTestId('is-sub').textContent).toBe('false');

    // Now resolve app-2 subscription
    await act(async () => {
      resolveApp2(undefined);
    });

    expect(screen.getByTestId('is-sub').textContent).toBe('true');
  });

  it('cancels old client and subscribes new client when client prop changes while appId stays same', async () => {
    const mock1 = createMockClient();
    const mock2 = createMockClient();

    const TestApp = ({ client }: { client: PrefSyncClient }) => {
      return (
        <PrefSyncProvider client={client} appId="app-fixed">
          <div>Child</div>
        </PrefSyncProvider>
      );
    };

    const { rerender } = render(<TestApp client={mock1.client} />);
    await act(async () => {});

    expect(mock1.subscribeNotification).toHaveBeenCalledTimes(1);
    expect(mock1.cancelMock).not.toHaveBeenCalled();
    expect(mock2.subscribeNotification).not.toHaveBeenCalled();

    // Switch client
    rerender(<TestApp client={mock2.client} />);
    await act(async () => {});

    expect(mock1.cancelMock).toHaveBeenCalledTimes(1);
    expect(mock2.subscribeNotification).toHaveBeenCalledTimes(1);
  });

  it('does not set isSubscribed to true if stream closes before subscribeNotification resolves', async () => {
    let capturedOnClose!: () => Promise<void>;
    const cancelMock = vi.fn().mockResolvedValue(undefined);

    const subscribeNotification = vi.fn().mockImplementation(
      async (
        _appId: string,
        _onEvent: any,
        _onError?: (e: any) => Promise<any>,
        onClose?: () => Promise<any>
      ) => {
        capturedOnClose = onClose!;

        // Trigger early close before returning sub object. Loop never settles so
        // isSubscribed staying false is due to the ended guard, not loop handlers.
        await capturedOnClose();

        return {
          loop: new Promise(() => {}),
          cancel: cancelMock,
        };
      }
    );

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const onClose = vi.fn();

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    const { unmount } = render(
      <PrefSyncProvider client={client} appId="app-early-close" onClose={onClose}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});

    expect(onClose).toHaveBeenCalledTimes(1);
    // isSubscribed must stay false and not be overwritten to true
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
    // Ended stream is cancelled immediately so the live handle is not leaked.
    expect(cancelMock).toHaveBeenCalledTimes(1);

    unmount();
    expect(cancelMock).toHaveBeenCalledTimes(2);
  });

  it('does not set isSubscribed to true if stream errors before subscribeNotification resolves', async () => {
    let capturedOnError!: (e: any) => Promise<void>;
    const cancelMock = vi.fn().mockResolvedValue(undefined);

    const subscribeNotification = vi.fn().mockImplementation(
      async (
        _appId: string,
        _onEvent: any,
        onError?: (e: any) => Promise<any>,
        _onClose?: () => Promise<any>
      ) => {
        capturedOnError = onError!;

        // Trigger early error before returning sub object. Loop never settles so
        // isSubscribed staying false is due to the ended guard, not loop handlers.
        await capturedOnError(new Error('early stream error'));

        return {
          loop: new Promise(() => {}),
          cancel: cancelMock,
        };
      }
    );

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const onError = vi.fn();

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    const { unmount } = render(
      <PrefSyncProvider client={client} appId="app-early-error" onError={onError}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});

    expect(onError).toHaveBeenCalledTimes(1);
    expect(onError).toHaveBeenCalledWith(expect.any(Error), client);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
    // Ended stream is cancelled immediately so the live handle is not leaked.
    expect(cancelMock).toHaveBeenCalledTimes(1);

    unmount();
    expect(cancelMock).toHaveBeenCalledTimes(2);
  });

  it('forwards both onError and onClose in order when stream encounters error and terminates', async () => {
    let capturedOnError!: (e: any) => Promise<void>;
    let capturedOnClose!: () => Promise<void>;

    const subscribeNotification = vi.fn().mockImplementation(
      async (
        _appId: string,
        _onEvent: any,
        onError?: (e: any) => Promise<any>,
        onClose?: () => Promise<any>
      ) => {
        capturedOnError = onError!;
        capturedOnClose = onClose!;

        return {
          loop: new Promise<void>(() => {}),
          cancel: vi.fn(),
        };
      }
    );

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const callOrder: string[] = [];
    const onError = vi.fn().mockImplementation(() => {
      callOrder.push('error');
    });
    const onClose = vi.fn().mockImplementation(() => {
      callOrder.push('close');
    });

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <PrefSyncProvider client={client} appId="app-err-close" onError={onError} onClose={onClose}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    const error = new Error('Mid-stream error');
    await act(async () => {
      await capturedOnError(error);
      await capturedOnClose();
    });

    expect(onError).toHaveBeenCalledTimes(1);
    expect(onError).toHaveBeenCalledWith(error, client);
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onClose).toHaveBeenCalledWith(client);
    expect(callOrder).toEqual(['error', 'close']);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
  });

  it('forwards onError and onClose with real SDK when fetch stream errors', async () => {
    const errorStream = new ReadableStream({
      start(controller) {
        controller.error(new Error('ReadableStream network failure'));
      },
    });

    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(errorStream, {
          status: 200,
          headers: { 'Content-Type': 'text/event-stream' },
        })
      )
    );

    try {
      const realClient = createPrefSyncClient({
        basePath: 'http://localhost:8080',
        accessToken: 'test-token',
      });

      const callOrder: string[] = [];
      const onError = vi.fn().mockImplementation(() => {
        callOrder.push('error');
      });
      const onClose = vi.fn().mockImplementation(() => {
        callOrder.push('close');
      });

      const Consumer = () => {
        const { isSubscribed } = usePrefSync();
        return <span data-testid="is-sub">{String(isSubscribed)}</span>;
      }

      render(
        <PrefSyncProvider client={realClient} appId="app-real-sdk-err" onError={onError} onClose={onClose}>
          <Consumer />
        </PrefSyncProvider>
      );

      await act(async () => {});

      expect(onError).toHaveBeenCalledTimes(1);
      expect(onError).toHaveBeenCalledWith(expect.any(Error), realClient);
      expect(onClose).toHaveBeenCalledTimes(1);
      expect(onClose).toHaveBeenCalledWith(realClient);
      expect(callOrder).toEqual(['error', 'close']);
      expect(screen.getByTestId('is-sub').textContent).toBe('false');
    } finally {
      vi.unstubAllGlobals();
    };
  });

  it('works correctly in React StrictMode double mounting', async () => {
    let cancelCount = 0;
    const cancelMock = vi.fn().mockImplementation(async () => {
      cancelCount++;
    });

    let subCount = 0;
    const subscribeNotification = vi.fn().mockImplementation(async () => {
      subCount++;
      return {
        loop: new Promise<void>(() => {}),
        cancel: cancelMock,
      };
    });

    const client = {
      getUser: vi.fn(),
      subscribeNotification,
    } as unknown as PrefSyncClient;

    const onClose = vi.fn();

    const Consumer = () => {
      const { isSubscribed } = usePrefSync();
      return <span data-testid="is-sub">{String(isSubscribed)}</span>;
    };

    render(
      <React.StrictMode>
        <PrefSyncProvider client={client} appId="app-strict" onClose={onClose}>
          <Consumer />
        </PrefSyncProvider>
      </React.StrictMode>
    );

    await act(async () => {});

    // In StrictMode, mount effect runs twice (mount, cleanup, mount)
    expect(subCount).toBe(2);
    expect(cancelCount).toBe(1);
    expect(onClose).not.toHaveBeenCalled();
    expect(screen.getByTestId('is-sub').textContent).toBe('true');
  });

  it('stopSubscription cancels the active subscription and sets isSubscribed to false', async () => {
    const { client, cancelMock } = createMockClient();

    const Consumer = () => {
      const { isSubscribed, stopSubscription } = usePrefSync();
      return (
        <div>
          <span data-testid="is-sub">{String(isSubscribed)}</span>
          <button data-testid="stop-sub" onClick={() => stopSubscription()}>
            Stop
          </button>
        </div>
      );
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc">
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');
    expect(cancelMock).not.toHaveBeenCalled();

    // Manually stop the subscription
    await act(async () => {
      fireEvent.click(screen.getByTestId('stop-sub'));
    });

    expect(cancelMock).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
  });

  it('stopSubscription is a no-op when no subscription is active', async () => {
    const { client, cancelMock } = createMockClient();

    const Consumer = () => {
      const { stopSubscription } = usePrefSync();
      return (
        <button data-testid="stop-sub" onClick={() => stopSubscription()}>
          Stop
        </button>
      );
    };

    render(
      <PrefSyncProvider client={client}>
        <Consumer />
      </PrefSyncProvider>
    );

    // No appId, so no subscription started
    await act(async () => {
      fireEvent.click(screen.getByTestId('stop-sub'));
    });

    expect(cancelMock).not.toHaveBeenCalled();
  });

  it('stopSubscription is a no-op when called twice', async () => {
    const { client, cancelMock } = createMockClient();

    const Consumer = () => {
      const { isSubscribed, stopSubscription } = usePrefSync();
      return (
        <div>
          <span data-testid="is-sub">{String(isSubscribed)}</span>
          <button data-testid="stop-sub" onClick={() => stopSubscription()}>
            Stop
          </button>
        </div>
      );
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc">
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    // First stop
    await act(async () => {
      fireEvent.click(screen.getByTestId('stop-sub'));
    });

    expect(cancelMock).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');

    // Second stop — should be a no-op
    await act(async () => {
      fireEvent.click(screen.getByTestId('stop-sub'));
    });

    expect(cancelMock).toHaveBeenCalledTimes(1);
  });

  it('stopSubscription triggers cancel which invokes onClose via the SDK loop', async () => {
    const { client, cancelMock } = createMockClient();
    const onClose = vi.fn();

    const Consumer = () => {
      const { isSubscribed, stopSubscription } = usePrefSync();
      return (
        <div>
          <span data-testid="is-sub">{String(isSubscribed)}</span>
          <button data-testid="stop-sub" onClick={() => stopSubscription()}>
            Stop
          </button>
        </div>
      );
    };

    render(
      <PrefSyncProvider client={client} appId="app-abc" onClose={onClose}>
        <Consumer />
      </PrefSyncProvider>
    );

    await act(async () => {});
    expect(screen.getByTestId('is-sub').textContent).toBe('true');

    // Stop the subscription
    await act(async () => {
      fireEvent.click(screen.getByTestId('stop-sub'));
    });

    // cancel() is called, which aborts the stream
    expect(cancelMock).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('is-sub').textContent).toBe('false');
  });
});
