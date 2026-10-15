import {AppsApi, Configuration as ClientConfig, Notification, PreferencesApi, UsersApi} from './generated'
import type {RawAxiosRequestConfig} from "axios";
import {ServerSentEventTransformStream} from "parse-sse";
import * as api from './generated/api'

export * from './generated/api'


export interface Config {
    basePath: string;
    accessToken: string;
    src?: string;
    autoSrc?: boolean;

    // Advanced Authentication Configs
    // Other services between the Frontend Client and the Backend (e.g. BFF Pattern) may use
    // Session Cookies or other Headers for authentication.
    // Consider the following options if your frontend is not interfacing the backend directly.
    headers?: any;
    withCredentials?: boolean;
}

export const createPrefSyncClient = (config: Config) => {
    const clientConfig = new ClientConfig({
        basePath: config.basePath,
        accessToken: config.accessToken,
        ...(config?.headers ? {
            baseOptions: {
                headers: config.headers,
            }
        } : {})
    })

    const userAPI = new UsersApi(clientConfig)
    const appsAPI = new AppsApi(clientConfig)
    const prefAPI = new PreferencesApi(clientConfig)
    config.autoSrc ??= true

    const withCredOpts = config?.withCredentials !== undefined ? {withCredentials: config.withCredentials} : {}

    const src = config.src || (
        !config.autoSrc ?
            undefined :
            // Random Hex with length of 4 bytes.
            [...crypto.getRandomValues(new Uint8Array(4))]
                .map(m => ('0' + m.toString(16)).slice(-2)).join('')
    )

    return {
        getUser: async () => (await userAPI.getUser({...withCredOpts}))?.data,
        deleteUser: async () => (await userAPI.deleteUser({...withCredOpts}))?.data,
        getApp: async (id: string) => (await appsAPI.getApp(id, {...withCredOpts}))?.data,

        // createApp requires Admin / Apps Creator authority. Regular users may not be authorized for this call.
        createApp: async (name: string, desc: string) => (
            await appsAPI.createApp({name, desc}, {...withCredOpts})
        )?.data,

        authoriseApp: async (appID: string) => (await userAPI.authoriseUser(appID, {...withCredOpts}))?.data,
        deauthoriseApp: async (appID: string) => (await userAPI.deauthoriseUser(appID, {...withCredOpts}))?.data,
        getPref: async (appID: string) => (await prefAPI.getPref(appID, {...withCredOpts}))?.data,
        postPref: async (appID: string, payload: any) => {
            const rsp = src !== undefined ?
                await prefAPI.postPrefFromSource(appID, src, payload, {...withCredOpts}) :
                await prefAPI.postPref(appID, payload, {...withCredOpts})
            return rsp?.data
        },
        subscribeNotification: async (
            appID: string,
            onEvent: (n: Notification) => Promise<any>,
            onError?: (err: any) => Promise<any> | any,
            onClose?: () => Promise<any> | any,
        ) => {
            const controller = new AbortController()
            const axiosOpt = {
                responseType: 'stream',
                headers: {Accept: 'text/event-stream'},
                adapter: 'fetch',
                signal: controller.signal,
                ...withCredOpts,
            } satisfies RawAxiosRequestConfig
            const rsp = src !== undefined ?
                await prefAPI.startNotificationFromSource(appID, src, axiosOpt) :
                await prefAPI.startNotification(appID, axiosOpt)
            const stream = (rsp.data as ReadableStream).pipeThrough(
                new TextDecoderStream()
            ).pipeThrough(
                new ServerSentEventTransformStream()
            );
            const reader = stream.getReader()
            const cancel = async () => {
                controller.abort()
                try {
                    await reader.cancel()
                } catch {
                    // Ignore already-cancelled or aborted reader
                }
            }
            const loop = new Promise<void>(
                async (resolve, reject) => {
                    try {
                        while (true) {
                            const {value, done} = await reader.read();
                            if (done) break;
                            if (value.type === 'server-disconnected') {
                                break
                            }
                            if (value.type === 'message') {
                                const payload: Notification = JSON.parse(value.data)
                                await onEvent(payload)
                            }
                        }
                        resolve()
                    } catch (e: any) {
                        if (e?.name === 'AbortError' || controller.signal.aborted) {
                            resolve()
                        } else if (onError) {
                            await onError(e)
                                .then(() => resolve())
                                .catch((e: any) => {
                                    reject(e)
                                })
                        } else {
                            reject(e)
                        }
                    }
                }
            )
                .finally(async () => {
                    await cancel()
                    if (onClose) {
                        await onClose()
                    }
                })
            return {loop, cancel}
        },
    };
}

export type PrefSyncClient = ReturnType<typeof createPrefSyncClient>