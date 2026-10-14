import {type Config, createPrefSyncClient} from "../src";
import * as oidc from "openid-client";

async function getConfig(
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
        {scope: 'api:stub'}
    );

    return {
        basePath,
        accessToken: tokens.access_token,
    } as Config;
}


async function index() {
    const config = await getConfig("http://127.0.0.1:7086")
    const c = createPrefSyncClient(config)
    const c2 = createPrefSyncClient(config)
    try {
        const app = (await c.createApp("test", "test")).id!
        await c.authoriseApp(app)
        const {loop, cancel} = await c.subscribeNotification(app, async (n) => {
            const rsp = await c.getPref(app)
            console.log(n, rsp)
        })
        await c2.postPref(app, "Fugu")
        await c2.postPref(app, "Fugu Fish")
        await c2.postPref(app, "Fugu")
        await c2.postPref(app, "Fugu Fish")
        await Promise.all([loop, cancel()])
    } catch (e) {
        console.log(e)
    }
}

index()