import {Config, createPrefSyncClient, PrefSyncClient} from 'pref-sync';
import * as oidc from 'openid-client';

const STUB_CLIENT = 'stub-client-service';
const STUB_SECRET = 'stub-client-secret';
const STUB_SCOPE = 'api:stub';

/**
 * Fetch a JWT token from the stub IdP via client_credentials grant.
 */
async function fetchDevToken(basePath: string): Promise<string> {
  const url = `${basePath.replace(/\/+$/, '')}/idp/oidc/token`;
  const basic = btoa(`${STUB_CLIENT}:${STUB_SECRET}`);
  const resp = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      Authorization: `Basic ${basic}`,
    },
    body: new URLSearchParams({
      grant_type: 'client_credentials',
      scope: STUB_SCOPE,
    }),
  });
  if (!resp.ok) {
    const body = await resp.text().catch(() => '');
    throw new Error(`Token request failed (${resp.status})${body ? `: ${body}` : ''}`);
  }
  const data = await resp.json();
  return data.access_token;
}

/**
 * Result of running the full auto-initialization routine.
 */
export interface InitResult {
    client: PrefSyncClient;
    config: Config;
    appId: string;
}

/**
 * Run the full auto-initialization: fetch a JWT, create and authorize an app,
 * and post an initial preference. Returns the config and appId for
 * PrefSyncProvider.
 */
export async function initExample(basePath: string = ''): Promise<InitResult> {
    // 1. Fetch JWT token
    const accessToken = await fetchDevToken(basePath);

    // 2. Create a PrefSync client
    // The client generates 'src' automatically by default. Yet for the sake of example, we set it manually.
    const src = 'current-client'
    const client = createPrefSyncClient({basePath, accessToken, src});

    // 3. Create a new demo app
    const created = await client.createApp('Minimal Demo', 'Auto-created for the minimal example');
    const appId = created?.id;
    if (!appId) {
        throw new Error('createApp returned no app ID');
    }

    // 4. Authorize the app
    await client.authoriseApp(appId);

    // 5. Post an initial preference
    await client.postPref(appId, {theme: 'light', greeting: 'Hello from minimal example'});

    return {
        client,
        config: {basePath, accessToken, src},
        appId,
    };
}
