import {beforeAll, describe, expect, it} from 'vitest';
import {type Config, createPrefSyncClient} from '../src';
import {getConfig} from './auth';
import type {Notification} from '../src/generated';
import {BasePath} from './config'


describe('Source ID Resolution & Client Modes (Live Server)', () => {
    let config: Config;
    let testAppId: string;

    beforeAll(async () => {
        config = await getConfig(BasePath);
        const setupClient = createPrefSyncClient(config);
        const app = await setupClient.createApp(`source-res-app-${Date.now()}`, 'Source resolution test app');
        testAppId = app!.id!;
        await setupClient.authoriseApp(testAppId);
    });

    it('generates an 8-character hexadecimal source identifier by default', async () => {
        const observer = createPrefSyncClient({
            ...config,
            autoSrc: false,
        });
        const autoClient = createPrefSyncClient(config);

        let receivedNotification: Notification | null = null;
        let resolveNotification: () => void;
        const notificationPromise = new Promise<void>((resolve) => {
            resolveNotification = resolve;
        });

        const {loop, cancel} = await observer.subscribeNotification(testAppId, async (n) => {
            receivedNotification = n;
            resolveNotification();
        });

        await new Promise((r) => setTimeout(r, 100));

        await autoClient.postPref(testAppId, 'auto-src-payload');
        await notificationPromise;

        expect(receivedNotification).toBeDefined();
        expect(receivedNotification!.Src).toBeDefined();
        expect(typeof receivedNotification!.Src).toBe('string');
        expect(receivedNotification!.Src).toMatch(/^[0-9a-f]{8}$/);

        await cancel();
        await loop;
    });

    it('uses explicit source identifier and delivers it on wire', async () => {
        const observer = createPrefSyncClient({
            ...config,
            autoSrc: false,
        });
        const explicitSrc = 'custom-worker-instance-42';
        const explicitClient = createPrefSyncClient({
            ...config,
            src: explicitSrc,
        });

        let receivedNotification: Notification | null = null;
        let resolveNotification: () => void;
        const notificationPromise = new Promise<void>((resolve) => {
            resolveNotification = resolve;
        });

        const {loop, cancel} = await observer.subscribeNotification(testAppId, async (n) => {
            receivedNotification = n;
            resolveNotification();
        });

        await new Promise((r) => setTimeout(r, 100));

        await explicitClient.postPref(testAppId, 'explicit-src-payload');
        await notificationPromise;

        expect(receivedNotification).toBeDefined();
        expect(receivedNotification!.Src).toBe(explicitSrc);

        await cancel();
        await loop;
    });

    it('generates distinct hex sources for different client instances', async () => {
        const observer = createPrefSyncClient({
            ...config,
            autoSrc: false,
        });
        const clientA = createPrefSyncClient(config);
        const clientB = createPrefSyncClient(config);

        const receivedSources: string[] = [];
        let resolveDone: () => void;
        const donePromise = new Promise<void>((resolve) => {
            resolveDone = resolve;
        });

        const {loop, cancel} = await observer.subscribeNotification(testAppId, async (n) => {
            if (n.Src) receivedSources.push(n.Src);
            if (receivedSources.length === 2) {
                resolveDone();
            }
        });

        await new Promise((r) => setTimeout(r, 100));

        await clientA.postPref(testAppId, 'payload-a');
        await clientB.postPref(testAppId, 'payload-b');

        await donePromise;

        expect(receivedSources.length).toBe(2);
        expect(receivedSources[0]).toMatch(/^[0-9a-f]{8}$/);
        expect(receivedSources[1]).toMatch(/^[0-9a-f]{8}$/);
        expect(receivedSources[0]).not.toBe(receivedSources[1]);

        await cancel();
        await loop;
    });

    it('handles empty string src with autoSrc: true by generating random hex source', async () => {
        const observer = createPrefSyncClient({
            ...config,
            autoSrc: false,
        });
        const emptyAutoClient = createPrefSyncClient({
            ...config,
            src: '',
            autoSrc: true,
        });

        let receivedNotification: Notification | null = null;
        let resolveNotification: () => void;
        const notificationPromise = new Promise<void>((resolve) => {
            resolveNotification = resolve;
        });

        const {loop, cancel} = await observer.subscribeNotification(testAppId, async (n) => {
            receivedNotification = n;
            resolveNotification();
        });

        await new Promise((r) => setTimeout(r, 100));

        await emptyAutoClient.postPref(testAppId, 'empty-auto-payload');
        await notificationPromise;

        expect(receivedNotification).toBeDefined();
        expect(receivedNotification!.Src).toMatch(/^[0-9a-f]{8}$/);

        await cancel();
        await loop;
    });
});
