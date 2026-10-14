import { createContext, useCallback, useEffect, useMemo, useRef, useState, type JSX } from 'react';
import { createPrefSyncClient, type PrefSyncClient } from 'pref-sync';
import type { PrefSyncCallbacks, PrefSyncContextValue, PrefSyncProviderProps } from './types';

export const PrefSyncContext = createContext<PrefSyncContextValue | null>(null);
PrefSyncContext.displayName = 'PrefSyncContext';

export const PrefSyncProvider = (props: PrefSyncProviderProps): JSX.Element => {
  const { client: directClient, config, appId, onEvent, onError, onClose, children } = props;

  const { basePath, accessToken, src, autoSrc } = config ?? {};
  const normalizedAutoSrc = autoSrc ?? true;

  const client = useMemo(
    () =>
      directClient ??
      (config && basePath !== undefined
        ? createPrefSyncClient({
            basePath,
            accessToken: accessToken!,
            src,
            autoSrc: normalizedAutoSrc,
          })
        : null),
    [directClient, basePath, accessToken, src, normalizedAutoSrc]
  );

  if (!client) {
    throw new Error('PrefSyncProvider requires either a "config" or a "client" prop.');
  }

  const callbacksRef = useRef<PrefSyncCallbacks>({ onEvent, onError, onClose });
  callbacksRef.current = { onEvent, onError, onClose };

  // Ref to hold the current subscription cancel function so it can be called
  // from stopSubscription() outside the useEffect closure.
  const cancelRef = useRef<(() => Promise<void> | void) | null>(null);

  const [activeSubscription, setActiveSubscription] = useState<{
    client: PrefSyncClient;
    appId: string;
  } | null>(null);

  const isSubscribed =
    activeSubscription !== null &&
    activeSubscription.client === client &&
    Boolean(appId) &&
    activeSubscription.appId === appId;

  useEffect(() => {
    if (!client || !appId) {
      setActiveSubscription(null);
      return;
    }

    const currentClient = client;
    const currentAppId = appId;
    let cancelled = false;
    let ended = false;
    let errorReported = false;
    let closeReported = false;

    const startSubscription = async () => {
      try {
        const sub = await currentClient.subscribeNotification(
          currentAppId,
          async (notification) => {
            if (cancelled || ended) return;
            await callbacksRef.current.onEvent?.(notification, currentClient);
          },
          async (error) => {
            if (cancelled || errorReported) return;
            errorReported = true;
            ended = true;
            setActiveSubscription((prev) =>
              prev?.client === currentClient && prev?.appId === currentAppId ? null : prev
            );
            try {
              await callbacksRef.current.onError?.(error, currentClient);
            } catch {
              // Safeguard against errors thrown inside user's onError handler
            }
          },
          async () => {
            if (cancelled || closeReported) return;
            closeReported = true;
            ended = true;
            setActiveSubscription((prev) =>
              prev?.client === currentClient && prev?.appId === currentAppId ? null : prev
            );
            try {
              await callbacksRef.current.onClose?.(currentClient);
            } catch {
              // Safeguard against errors thrown inside user's onClose handler
            }
          }
        );

        if (sub?.loop && typeof (sub.loop as Promise<any>).then === 'function') {
          (async () => {
            try {
              await sub.loop;
              if (cancelled) return;
              ended = true;
              setActiveSubscription((prev) =>
                prev?.client === currentClient && prev?.appId === currentAppId ? null : prev
              );
            } catch {
              if (cancelled) return;
              ended = true;
              setActiveSubscription((prev) =>
                prev?.client === currentClient && prev?.appId === currentAppId ? null : prev
              );
            }
          })();
        }

        // Store the cancel handle in the ref so stopSubscription() and cleanup
        // can both access it.
        cancelRef.current = sub.cancel;

        if (cancelled || ended) {
          try {
            const res = sub.cancel();
            if (res && typeof (res as Promise<any>).catch === 'function') {
              await res.catch(() => {});
            }
          } catch {
            // Ignore cancellation errors
          }
          return;
        }

        setActiveSubscription({ client: currentClient, appId: currentAppId });
      } catch (err) {
        if (cancelled || errorReported) return;
        errorReported = true;
        ended = true;
        setActiveSubscription((prev) =>
          prev?.client === currentClient && prev?.appId === currentAppId ? null : prev
        );
        try {
          await callbacksRef.current.onError?.(err, currentClient);
        } catch {
          // Safeguard against errors thrown inside user's onError handler
        }
      }
    };

    void startSubscription();

    return () => {
      cancelled = true;
      setActiveSubscription((prev) =>
        prev?.client === currentClient && prev?.appId === currentAppId ? null : prev
      );
      const cancel = cancelRef.current;
      cancelRef.current = null;
      if (cancel) {
        try {
          const res = cancel();
          if (res && typeof (res as Promise<any>).catch === 'function') {
            (res as Promise<any>).catch(() => {});
          }
        } catch {
          // Ignore cancellation errors
        }
      }
    };
  }, [client, appId]);

  const stopSubscription = useCallback(() => {
    const cancel = cancelRef.current;
    if (!cancel) return;
    cancelRef.current = null;
    // Suppress stale callbacks by clearing the active subscription.
    // The callbacks check `cancelled` which is set by the cleanup when
    // stopSubscription clears cancelRef — but since `cancelled` is local to
    // the useEffect closure, we rely on the fact that calling cancel() aborts
    // the stream, which triggers the onClose callback that sets ended=true.
    // We also clear activeSubscription so isSubscribed reflects the stopped state.
    setActiveSubscription((prev) => {
      if (prev) {
        try {
          const res = cancel();
          if (res && typeof (res as Promise<any>).catch === 'function') {
            (res as Promise<any>).catch(() => {});
          }
        } catch {
          // Ignore cancellation errors
        }
        return null;
      }
      return prev;
    });
  }, []);

  const contextValue = useMemo<PrefSyncContextValue>(
    () => ({
      client,
      appId,
      isSubscribed,
      stopSubscription,
    }),
    [client, appId, isSubscribed, stopSubscription]
  );

  return (
    <PrefSyncContext.Provider value={contextValue}>
      {children}
    </PrefSyncContext.Provider>
  );
};
