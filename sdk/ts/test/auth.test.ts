import {describe, expect, it} from 'vitest';
import {getConfig} from './auth';
import {BasePath} from './config'


describe('Auth & Token Routine', () => {
    it('obtains a valid access token from the local OIDC IdP', async () => {
        const config = await getConfig(BasePath);

        expect(config).toBeDefined();
        expect(config.basePath).toBe(BasePath);
        expect(typeof config.accessToken).toBe('string');
        expect(config.accessToken.length).toBeGreaterThan(0);
    });

    it('uses default basePath if none provided', async () => {
        const config = await getConfig();

        expect(config.basePath).toBe(BasePath);
        expect(typeof config.accessToken).toBe('string');
        expect(config.accessToken.length).toBeGreaterThan(0);
    });
});
