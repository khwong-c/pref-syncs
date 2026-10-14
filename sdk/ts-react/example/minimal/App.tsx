import React, {useCallback, useEffect, useMemo, useState} from 'react';
import type {Notification, PrefSyncClient} from 'pref-sync';
import {createPrefSyncClient} from 'pref-sync';
import {PrefSyncProvider, usePrefSync} from '../../src';
import type {InitResult} from './init';
import {initExample} from './init';
import {S} from "./style";

type NotificationWithTime = Notification & { time: string };

export function App() {
    const [initResult, setInitResult] = useState<InitResult | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [events, setEvents] = useState<NotificationWithTime[]>([]);

    // Initiate clients and configs at the beginning.
    useEffect(() => {
        initExample('')
            .then((r) => {
                setInitResult(r);
                setLoading(false);
            })
            .catch((err: any) => {
                setError(err?.message || String(err));
                setLoading(false);
            });
    }, []);

    // The setting of the peer client is basically the same with the current client.
    // We set 'src' manually for the peer client so that we can distinguish the two clients.
    const peerClient = useMemo(
        () => createPrefSyncClient(
            {
                basePath: initResult?.config?.basePath || '',
                accessToken: initResult?.config?.accessToken || '',
                src: 'peer-client'
            }),
        [
            initResult?.config?.basePath,
            initResult?.config?.accessToken,
        ]
    );

    // Function to create a new client.
    // By assigning a new client to the provider, we can restart the subscription.
    const recreateClient = useCallback(() => {
        if (!initResult || !initResult.config) return;
        console.log('Updating client');
        setLoading(true);
        const newClient = createPrefSyncClient(initResult.config);
        setInitResult(
            {
                ...initResult,
                client: newClient
            }
        )
        setLoading(false);
    }, [initResult?.config]);

    if (loading) return <div style={S.container}>Initializing...</div>;
    if (error) return (
        <div style={S.container}>
            <h1 style={S.title}>Minimal Example</h1>
            <p style={S.error}>Initialization failed: {error}</p>
        </div>
    );
    if (!initResult) return null;

    return (
        <PrefSyncProvider
            client={initResult.client}
            config={initResult.config}
            appId={initResult.appId}
            onEvent={(n) => setEvents((p) => [{...n, time: new Date().toTimeString()}, ...p])}>
            <Demo events={events} peerClient={peerClient} recreateClient={recreateClient}/>
        </PrefSyncProvider>
    );
}

function Demo({events, peerClient, recreateClient}: {
    events: Notification[],
    peerClient: PrefSyncClient,
    recreateClient: () => void,
}) {
    const {appId, isSubscribed, stopSubscription} = usePrefSync();
    const [posting, setPosting] = useState(false);

    const handlePost = async () => {
        if (!appId) return;
        setPosting(true);
        try {
            // Posting from a peer client to simulate an external posting event.
            await peerClient.postPref(appId, {timestamp: Date.now(), message: 'Posted from minimal example'});
        } catch (err: any) {
            console.error('postPref failed:', err);
        } finally {
            setPosting(false);
        }
    };

    const handleUnsubscribe = async () => {
        if (!isSubscribed) return;
        stopSubscription();
    };

    return (
        <div style={S.container}>
            <h1 style={S.title}>Pref-Sync React SDK — Minimal Example</h1>

            <div style={S.card}>
                <p style={S.text}><strong>App ID:</strong> {appId}</p>
                <p style={S.text}><strong>Subscription:</strong>{' '}
                    <span style={{color: isSubscribed ? '#16a34a' : '#dc2626'}}>
            {isSubscribed ? 'Connected' : 'Disconnected'}
          </span>
                </p>
            </div>

            <div style={S.card}>
                <div>
                    <button onClick={handlePost} disabled={posting} style={S.button}>
                        {posting ? 'Posting...' : 'Post Preference from Peer'}
                    </button>
                    <p style={S.hint}>Click to trigger a preference update from another client. Watch the event log
                        below.</p>
                </div>
                <div>
                    <button
                        onClick={handleUnsubscribe}
                        disabled={!isSubscribed}
                        style={isSubscribed ? S.button : S.buttonDisabled}>
                        Unsubscribe
                    </button>
                    <p style={S.hint}>Unsubscribe the client from receiving preference updates. Try to post a preference
                        update from another client after unsubscribing.</p>
                </div>
                <div>
                    <button onClick={() => recreateClient()}
                            disabled={isSubscribed}
                            style={!isSubscribed ? S.button : S.buttonDisabled}
                    >
                        Re-create Client
                    </button>
                    <p style={S.hint}>Recreate a client to receive updates again. Try to post a preference
                        update from another client after create the client after unsubscribed.</p>
                </div>

            </div>

            <div style={S.card}>
                <h2 style={S.subtitle}>Event Log ({events.length})</h2>
                <div style={S.log}>
                    {events.length === 0 && <p style={S.text}>Waiting for events...</p>}
                    {events.map((evt, i) => (
                        <div key={i} style={S.logEntry}>
                            <code style={S.code}>{JSON.stringify(evt, null, 2)}</code>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
