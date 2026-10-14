import {beforeAll, describe, expect, it} from 'vitest';
import {type Config, createPrefSyncClient} from '../src';
import {getConfig} from './auth';
import {BasePath} from './config'


describe('User Lifecycle & Application Authorization (Live Server)', () => {
    let config: Config;

    beforeAll(async () => {
        config = await getConfig(BasePath);
    });

    it('retrieves the authenticated user profile', async () => {
        const client = createPrefSyncClient(config);
        const user = await client.getUser();

        expect(user).toBeDefined();
        expect(user?.id).toBeDefined();
        expect(typeof user?.id).toBe('string');
        expect(Array.isArray(user?.apps)).toBe(true);
    });

    it('creates an application, retrieves it, and handles authorization lifecycle', async () => {
        const client = createPrefSyncClient(config);
        const appName = `test-app-${Date.now()}`;
        const appDesc = 'Functional integration test app';

        // 1. Create App
        const createdApp = await client.createApp(appName, appDesc);
        expect(createdApp).toBeDefined();
        expect(createdApp?.id).toBeDefined();
        expect(createdApp?.name).toBe(appName);
        expect(createdApp?.desc).toBe(appDesc);

        const appId = createdApp!.id!;

        // 2. Retrieve App
        const fetchedApp = await client.getApp(appId);
        expect(fetchedApp).toBeDefined();
        expect(fetchedApp?.id).toBe(appId);
        expect(fetchedApp?.name).toBe(appName);
        expect(fetchedApp?.desc).toBe(appDesc);

        // 3. Authorise App for User
        const authResult = await client.authoriseApp(appId);
        expect(authResult).toBeDefined();
        expect(authResult?.id).toBeDefined();
        expect(authResult?.apps?.some((a) => (typeof a === 'string' ? a === appId : a.id === appId))).toBe(true);

        const userAfterAuth = await client.getUser();
        expect(userAfterAuth?.apps?.some((a) => (typeof a === 'string' ? a === appId : a.id === appId))).toBe(true);

        // 4. Deauthorise App for User
        const deauthResult = await client.deauthoriseApp(appId);
        expect(deauthResult).toBeDefined();
        expect(deauthResult?.apps?.some((a) => (typeof a === 'string' ? a === appId : a.id === appId))).toBe(false);
    });

    it('handles non-existent app retrieval cleanly', async () => {
        const client = createPrefSyncClient(config);
        await expect(client.getApp('non-existent-app-id-12345')).rejects.toThrow();
    });
});
