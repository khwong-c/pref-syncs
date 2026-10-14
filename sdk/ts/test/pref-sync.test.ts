import {beforeAll, describe, expect, it} from 'vitest';
import {type Config, createPrefSyncClient} from '../src';
import {getConfig} from './auth';
import {BasePath} from './config'


describe('Preference Synchronization (Live Server)', () => {
    let config: Config;
    let testAppId: string;

    beforeAll(async () => {
        config = await getConfig(BasePath);
        const client = createPrefSyncClient(config);
        const app = await client.createApp(`pref-test-app-${Date.now()}`, 'Pref test app');
        testAppId = app!.id!;
        await client.authoriseApp(testAppId);
    });

    it('posts and retrieves string preference data', async () => {
        const client = createPrefSyncClient(config);
        const payload = 'Fugu Fish Theme';

        const postRsp = await client.postPref(testAppId, payload);
        expect(postRsp).toBeDefined();

        const getRsp = await client.getPref(testAppId);
        expect(getRsp).toBeDefined();
        expect(getRsp?.app_id).toBe(testAppId);
        expect(getRsp?.data).toBe(payload);
    });

    it('posts and retrieves complex JSON object preference data', async () => {
        const client = createPrefSyncClient(config);
        const complexPayload = {
            theme: 'dark',
            editor: {
                fontSize: 16,
                tabSize: 2,
                wordWrap: true,
            },
            tags: ['typescript', 'sync', 'sse'],
            metadata: {
                nested: {
                    flag: false,
                    nullable: null,
                },
            },
            unicode: '🌟 寿司 🍣 & <custom>',
        };

        const postRsp = await client.postPref(testAppId, complexPayload);
        expect(postRsp).toBeDefined();

        const getRsp = await client.getPref(testAppId);
        expect(getRsp).toBeDefined();
        expect(getRsp?.app_id).toBe(testAppId);
        expect(getRsp?.data).toEqual(complexPayload);
    });

    it('posts preference updates using broadcast mode (autoSrc: false, no src)', async () => {
        const broadcastClient = createPrefSyncClient({
            ...config,
            autoSrc: false,
        });

        const payload = {broadcastUpdate: true, timestamp: Date.now()};
        const postRsp = await broadcastClient.postPref(testAppId, payload);
        expect(postRsp).toBeDefined();

        const getRsp = await broadcastClient.getPref(testAppId);
        expect(getRsp?.data).toEqual(payload);
    });

    it('posts preference updates with explicit custom source', async () => {
        const customSrcClient = createPrefSyncClient({
            ...config,
            src: 'custom-src-node-99',
        });

        const payload = {customSrcUpdate: true, version: 2};
        const postRsp = await customSrcClient.postPref(testAppId, payload);
        expect(postRsp).toBeDefined();

        const getRsp = await customSrcClient.getPref(testAppId);
        expect(getRsp?.data).toEqual(payload);
    });
});
