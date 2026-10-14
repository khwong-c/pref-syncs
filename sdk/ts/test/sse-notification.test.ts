import {beforeAll, describe, expect, it} from 'vitest';
import {type Config, createPrefSyncClient} from '../src';
import {getConfig} from './auth';
import type {Notification} from '../src/generated';
import {BasePath} from './config'


describe('Real-Time SSE Notification Streaming (Live Server)', () => {
    let config: Config;
    let testAppId: string;

    beforeAll(async () => {
        config = await getConfig(BasePath);
        const setupClient = createPrefSyncClient(config);
        const app = await setupClient.createApp(`sse-app-${Date.now()}`, 'SSE test app');
        testAppId = app!.id!;
        await setupClient.authoriseApp(testAppId);
    });

    it('receives real-time preference update notifications across clients', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'listener-client',
        });
        const publisher = createPrefSyncClient({
            ...config,
            src: 'publisher-client',
        });

        const receivedNotifications: Notification[] = [];
        let resolveFirstEvent: (n: Notification) => void;
        const firstEventPromise = new Promise<Notification>((resolve) => {
            resolveFirstEvent = resolve;
        });

        const {loop, cancel} = await subscriber.subscribeNotification(
            testAppId,
            async (notification) => {
                receivedNotifications.push(notification);
                resolveFirstEvent(notification);
            }
        );

        // Give SSE connection a small moment to establish
        await new Promise((r) => setTimeout(r, 100));

        await publisher.postPref(testAppId, 'update-from-publisher');
        const firstEvent = await firstEventPromise;

        expect(firstEvent).toBeDefined();
        expect(firstEvent.App).toBe(testAppId);
        expect(firstEvent.Src).toBe('publisher-client');

        await cancel();
        await loop;
    });

    it('excludes self-echo when src is configured (source-filtering mode)', async () => {
        const selfSrc = `source-filter-${Date.now()}`;
        const client = createPrefSyncClient({
            ...config,
            src: selfSrc,
        });
        const observer = createPrefSyncClient({
            ...config,
            src: 'observer-client',
        });

        let clientReceivedCount = 0;
        let observerReceivedNotification: Notification | null = null;
        let resolveObserver: () => void;
        const observerPromise = new Promise<void>((resolve) => {
            resolveObserver = resolve;
        });

        const subClient = await client.subscribeNotification(testAppId, async () => {
            clientReceivedCount++;
        });

        const subObserver = await observer.subscribeNotification(testAppId, async (n) => {
            observerReceivedNotification = n;
            resolveObserver();
        });

        await new Promise((r) => setTimeout(r, 100));

        // Client posts an update tagged with its own src
        await client.postPref(testAppId, 'filtered-self-update');

        // Observer should receive it
        await observerPromise;
        expect(observerReceivedNotification).toBeDefined();
        expect(observerReceivedNotification!.Src).toBe(selfSrc);

        // Wait a brief window to confirm filtered client received nothing
        await new Promise((r) => setTimeout(r, 150));
        expect(clientReceivedCount).toBe(0);

        await subClient.cancel();
        await subObserver.cancel();
        await Promise.all([subClient.loop, subObserver.loop]);
    });

    it('delivers self-echo notifications in Broadcast mode (autoSrc: false, no src)', async () => {
        const broadcastClient = createPrefSyncClient({
            ...config,
            autoSrc: false,
        });

        let receivedNotification: Notification | null = null;
        let resolveNotification: () => void;
        const notificationPromise = new Promise<void>((resolve) => {
            resolveNotification = resolve;
        });

        const {loop, cancel} = await broadcastClient.subscribeNotification(
            testAppId,
            async (n) => {
                receivedNotification = n;
                resolveNotification();
            }
        );

        await new Promise((r) => setTimeout(r, 100));

        // Post preference without source tag
        await broadcastClient.postPref(testAppId, 'broadcast-self-echo-update');

        await notificationPromise;
        expect(receivedNotification).toBeDefined();
        expect(receivedNotification!.App).toBe(testAppId);

        await cancel();
        await loop;
    });

    it('terminates stream cleanly when cancel() is invoked', async () => {
        const client = createPrefSyncClient(config);
        const {loop, cancel} = await client.subscribeNotification(testAppId, async () => {
        });

        await new Promise((r) => setTimeout(r, 50));
        await cancel();
        await expect(loop).resolves.toBeUndefined();
    });

    it('handles multiple sequential notifications in order', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'multi-sub',
        });
        const publisher = createPrefSyncClient({
            ...config,
            src: 'multi-pub',
        });

        const receivedPayloads: any[] = [];
        let resolveComplete: () => void;
        const completePromise = new Promise<void>((resolve) => {
            resolveComplete = resolve;
        });

        const {loop, cancel} = await subscriber.subscribeNotification(
            testAppId,
            async (_n) => {
                const pref = await subscriber.getPref(testAppId);
                receivedPayloads.push(pref?.data);
                if (receivedPayloads.length === 3) {
                    resolveComplete();
                }
            }
        );

        await new Promise((r) => setTimeout(r, 100));

        await publisher.postPref(testAppId, 'msg-1');
        await publisher.postPref(testAppId, 'msg-2');
        await publisher.postPref(testAppId, 'msg-3');

        await completePromise;

        expect(receivedPayloads.length).toBe(3);
        expect(receivedPayloads).toContain('msg-3');

        await cancel();
        await loop;
    });

    it('handles stream callback errors gracefully via onError callback', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'error-sub-1',
        });
        const publisher = createPrefSyncClient({
            ...config,
            src: 'error-pub-1',
        });

        const callbackError = new Error('Callback processing failed');
        let capturedError: any = null;
        let resolveError: () => void;
        const errorPromise = new Promise<void>((resolve) => {
            resolveError = resolve;
        });

        const {loop} = await subscriber.subscribeNotification(
            testAppId,
            async () => {
                throw callbackError;
            },
            async (err) => {
                capturedError = err;
                resolveError();
            }
        );

        await new Promise((r) => setTimeout(r, 100));
        await publisher.postPref(testAppId, 'error-test-payload');

        await errorPromise;
        expect(capturedError).toBe(callbackError);
        // Since onError handled it gracefully, loop should resolve cleanly
        await expect(loop).resolves.toBeUndefined();
    });

    it('propagates error when onError rethrows or throws', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'error-sub-2',
        });
        const publisher = createPrefSyncClient({
            ...config,
            src: 'error-pub-2',
        });

        // const customHandledError = new Error('Rethrown from onError');

        const {loop} = await subscriber.subscribeNotification(
            testAppId,
            async () => {
                throw new Error('Initial callback error');
            },
            async () => {
                throw new Error('Rethrown from onError');
            }
        );

        const loopPromise = expect(loop).rejects.toThrow('Rethrown from onError');
        await new Promise((r) => setTimeout(r, 100));
        await publisher.postPref(testAppId, 'rethrow-test-payload');

        await loopPromise;
    });

    it('rejects loop promise when error occurs and no onError is provided', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'error-sub-3',
        });
        const publisher = createPrefSyncClient({
            ...config,
            src: 'error-pub-3',
        });

        const unhandledError = new Error('Unhandled stream error');

        const {loop} = await subscriber.subscribeNotification(
            testAppId,
            async () => {
                throw unhandledError;
            }
        );

        const loopPromise = expect(loop).rejects.toThrow('Unhandled stream error');
        await new Promise((r) => setTimeout(r, 100));
        await publisher.postPref(testAppId, 'unhandled-error-payload');

        await loopPromise;
    });

    it('invokes onClose callback when subscription stream is cancelled', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'close-sub-1',
        });

        let closed = false;
        const {loop, cancel} = await subscriber.subscribeNotification(
            testAppId,
            async () => {
            },
            undefined,
            async () => {
                closed = true;
            }
        );

        await new Promise((r) => setTimeout(r, 100));
        await cancel();
        await loop;

        expect(closed).toBe(true);
    });

    it('invokes onClose callback when error is handled by onError', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'close-sub-2',
        });
        const publisher = createPrefSyncClient({
            ...config,
            src: 'close-pub-2',
        });

        let closed = false;
        let errorHandled = false;

        const {loop} = await subscriber.subscribeNotification(
            testAppId,
            async () => {
                throw new Error('Failure in event handler');
            },
            async (_err) => {
                errorHandled = true;
            },
            async () => {
                closed = true;
            }
        );

        await new Promise((r) => setTimeout(r, 100));
        await publisher.postPref(testAppId, 'close-test-payload');

        await loop;
        expect(errorHandled).toBe(true);
        expect(closed).toBe(true);
    });

    it('invokes onClose callback when unhandled error occurs', async () => {
        const subscriber = createPrefSyncClient({
            ...config,
            src: 'close-sub-3',
        });
        const publisher = createPrefSyncClient({
            ...config,
            src: 'close-pub-3',
        });

        let closed = false;

        const {loop} = await subscriber.subscribeNotification(
            testAppId,
            async () => {
                throw new Error('Fatal stream error');
            },
            undefined,
            async () => {
                closed = true;
            }
        );

        const loopPromise = expect(loop).rejects.toThrow('Fatal stream error');
        await new Promise((r) => setTimeout(r, 100));
        await publisher.postPref(testAppId, 'close-fatal-payload');

        await loopPromise;
        expect(closed).toBe(true);
    });
});
