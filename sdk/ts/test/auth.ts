import * as oidc from 'openid-client';
import type { Config } from '../src';

export interface AuthConfigOptions {
  basePath?: string;
  idp?: string;
  clientId?: string;
  clientSecret?: string;
  scope?: string;
}

/**
 * Retrieves client configuration containing a valid access token from the OIDC IdP.
 */
export async function getConfig(
  basePath: string = 'http://127.0.0.1:7086',
  idp?: string
): Promise<Config> {
  const oidcConfig = await oidc.discovery(
    new URL(idp || `${basePath}/idp`),
    'stub-client-service',
    'stub-client-secret',
    oidc.ClientSecretBasic('stub-client-secret'),
    {
      execute:
        !idp && basePath === 'http://127.0.0.1:7086'
          ? [oidc.allowInsecureRequests]
          : [],
    }
  );

  const tokens: oidc.TokenEndpointResponse = await oidc.clientCredentialsGrant(
    oidcConfig,
    { scope: 'api:stub' }
  );

  return {
    basePath,
    accessToken: tokens.access_token,
  } as Config;
}
