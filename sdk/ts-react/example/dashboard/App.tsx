import React, {useCallback, useMemo, useState} from 'react';
import type {Notification, PrefSyncClient} from '../../src';
import {Config, createPrefSyncClient, PrefSyncProvider, usePrefSync} from '../../src';

export interface LogEntry {
    id: string;
    timestamp: string;
    type: 'event' | 'error' | 'close' | 'api-success' | 'api-error' | 'peer-api-success';
    title: string;
    details?: unknown;
}

export interface CreatedApp {
    id: string;
    name: string;
    desc?: string;
}

const STUB_CLIENT = 'stub-client-service';
const STUB_SECRET = 'stub-client-secret';
const STUB_SCOPE = 'api:stub';

/**
 * Fetch a developer JWT from the stub IdP via client_credentials grant.
 */
export const fetchDevToken = async (basePath: string = ''): Promise<string> => {
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
        let body = '';
        try {
            body = await resp.text();
        } catch {
            // ignore
        }
        throw new Error(`Token request failed with status ${resp.status}${body ? `: ${body}` : ''}`);
    }
    const data = await resp.json();
    return data.access_token;
};

// Visual Styles
const cardStyle: React.CSSProperties = {
    backgroundColor: 'var(--card-bg)',
    borderRadius: '12px',
    border: '1px solid var(--card-border)',
    padding: '24px',
    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
};

const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '10px 12px',
    borderRadius: '6px',
    border: '1px solid var(--card-border)',
    backgroundColor: 'var(--input-bg)',
    color: 'var(--text-primary)',
    fontSize: '0.95rem',
    outline: 'none',
};

const buttonStyle: React.CSSProperties = {
    padding: '10px 16px',
    borderRadius: '6px',
    border: 'none',
    fontWeight: 600,
    fontSize: '0.9rem',
    cursor: 'pointer',
    transition: 'background-color 0.2s, opacity 0.2s',
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: '8px',
    color: '#ffffff',
};

const secondaryButtonStyle: React.CSSProperties = {
    ...buttonStyle,
    backgroundColor: '#334155',
    color: 'var(--text-primary)',
    border: '1px solid var(--card-border)',
};

const chipButtonStyle: React.CSSProperties = {
    padding: '4px 10px',
    borderRadius: '6px',
    border: '1px solid var(--card-border)',
    backgroundColor: 'transparent',
    color: 'var(--text-secondary)',
    fontSize: '0.8rem',
    cursor: 'pointer',
};

const labelStyle: React.CSSProperties = {
    display: 'block',
    fontSize: '0.85rem',
    fontWeight: 600,
    color: 'var(--text-secondary)',
    marginBottom: '6px',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
};

const codeBlockStyle: React.CSSProperties = {
    backgroundColor: 'var(--code-bg)',
    padding: '12px',
    borderRadius: '6px',
    fontFamily: 'var(--font-mono)',
    fontSize: '0.85rem',
    color: '#38bdf8',
    overflowX: 'auto',
    maxHeight: '220px',
    border: '1px solid var(--card-border)',
};

// Subscribed inner component that accesses the PrefSync context
export const PrefSyncDashboard = ({
                                      onAddLog,
                                      config,
                                      createdApps,
                                      onAppCreated,
                                      onSelectAppId,
                                      autoRefreshOnEvent,
                                      setAutoRefreshOnEvent,
                                      onAutoRefreshRegister,
                                  }: {
    onAddLog: (entry: Omit<LogEntry, 'id' | 'timestamp'>) => void;
    config: Config,
    createdApps: CreatedApp[];
    onAppCreated: (app: CreatedApp) => void;
    onSelectAppId: (appId: string) => void;
    autoRefreshOnEvent: boolean;
    setAutoRefreshOnEvent: (val: boolean) => void;
    onAutoRefreshRegister?: (handler: (data: unknown) => void) => void;
}) => {
    const {client, appId, isSubscribed} = usePrefSync();

    // Local state for app management
    const [newAppName, setNewAppName] = useState('My Workspace');
    const [isCreatingApp, setIsCreatingApp] = useState(false);
    const [targetAppId, setTargetAppId] = useState<string>(appId || '');

    // Local state for preference operations
    const [prefPayloadText, setPrefPayloadText] = useState<string>(
        JSON.stringify({theme: 'dark', fontSize: 16, autoSave: true}, null, 2)
    );
    const [currentPref, setCurrentPref] = useState<{
        data: unknown;
        updated_at?: string;
        user_id?: string;
        app_id?: string;
    } | null>(null);
    const [isFetchingPref, setIsFetchingPref] = useState(false);
    const [isPostingPref, setIsPostingPref] = useState(false);
    const [isPostingPeerPref, setIsPostingPeerPref] = useState(false);

    // Status banner/message
    const [statusMessage, setStatusMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    // SDK Explorer state
    const [apiResult, setApiResult] = useState<{ title: string; data: unknown } | null>(null);
    const [isCallingApi, setIsCallingApi] = useState(false);

    // Sync default target appId when provider appId changes
    React.useEffect(() => {
        if (appId) {
            setTargetAppId(appId);
        }
    }, [appId]);

    const showStatus = (type: 'success' | 'error', text: string) => {
        setStatusMessage({type, text});
        setTimeout(() => {
            setStatusMessage((prev) => (prev?.text === text ? null : prev));
        }, 5000);
    };

    // Create App
    const handleCreateApp = async () => {
        if (!newAppName.trim()) {
            showStatus('error', 'Please enter a name for the new app.');
            return;
        }
        setIsCreatingApp(true);
        try {
            const created = await client.createApp(
                newAppName.trim(),
                `Demo application: ${newAppName.trim()}`
            );
            const createdId = created?.id || '';
            const createdName = created?.name || newAppName.trim();
            const newApp: CreatedApp = {
                id: createdId,
                name: createdName,
                desc: created?.desc,
            };
            if (createdId) {
                onAppCreated(newApp);
                setTargetAppId(createdId);
                onSelectAppId(createdId);
            }

            onAddLog({
                type: 'api-success',
                title: `createApp("${newAppName.trim()}") created UUID: ${createdId}`,
                details: created,
            });
            showStatus('success', `Created App "${createdName}" (${createdId}). Automatically selected for sync.`);
        } catch (err: any) {
            const errorMsg = err?.response?.data?.message || err?.message || String(err);
            onAddLog({
                type: 'api-error',
                title: `createApp("${newAppName.trim()}") failed`,
                details: errorMsg,
            });
            showStatus('error', `Failed to create app: ${errorMsg}`);
        } finally {
            setIsCreatingApp(false);
        }
    };

    // Authorise App
    const handleAuthoriseApp = async (idToAuth: string = targetAppId) => {
        const trimmedId = idToAuth.trim();
        if (!trimmedId) {
            showStatus('error', 'Please enter or select an App UUID to authorise.');
            return;
        }
        setIsCallingApi(true);
        try {
            await client.authoriseApp(trimmedId);
            onAddLog({
                type: 'api-success',
                title: `authoriseApp("${trimmedId}") succeeded`,
            });
            showStatus('success', `App ${trimmedId} successfully authorised for your user.`);
        } catch (err: any) {
            const errorMsg = err?.response?.data?.message || err?.message || String(err);
            onAddLog({
                type: 'api-error',
                title: `authoriseApp("${trimmedId}") failed`,
                details: errorMsg,
            });
            showStatus('error', `Failed to authorise app: ${errorMsg}`);
        } finally {
            setIsCallingApi(false);
        }
    };

    // Normalize preference response
    const normalizePref = (resp: unknown) => {
        if (resp && typeof resp === 'object' && 'data' in (resp as any)) {
            return {
                data: (resp as any).data,
                updated_at: (resp as any).updated_at,
                user_id: (resp as any).user_id,
                app_id: (resp as any).app_id,
            };
        }
        return {data: resp};
    };

    // Fetch preferences
    const handleGetPref = async (appIdToFetch: string = targetAppId) => {
        const trimmedId = appIdToFetch.trim();
        if (!trimmedId) {
            showStatus('error', 'Please enter or select an App UUID.');
            return;
        }
        setIsFetchingPref(true);
        try {
            const resp = await client.getPref(trimmedId);
            const formattedPref = normalizePref(resp);
            setCurrentPref(formattedPref);
            onAddLog({
                type: 'api-success',
                title: `getPref("${trimmedId}") succeeded`,
                details: resp,
            });
            showStatus('success', `Fetched preferences for app ${trimmedId}`);
        } catch (err: any) {
            const errorMsg = err?.response?.data?.message || err?.message || String(err);
            onAddLog({
                type: 'api-error',
                title: `getPref("${trimmedId}") failed`,
                details: errorMsg,
            });
            showStatus('error', `Failed to get preferences: ${errorMsg}`);
        } finally {
            setIsFetchingPref(false);
        }
    };

    // Post preferences with current client
    const handlePostPref = async () => {
        const trimmedId = targetAppId.trim();
        if (!trimmedId) {
            showStatus('error', 'Please enter or select an App UUID.');
            return;
        }
        let parsedPayload: unknown;
        try {
            parsedPayload = JSON.parse(prefPayloadText);
        } catch {
            showStatus('error', 'Invalid JSON payload. Please verify syntax.');
            return;
        }

        setIsPostingPref(true);
        try {
            const resp = await client.postPref(trimmedId, parsedPayload);
            const formattedPref = normalizePref(resp);
            setCurrentPref(formattedPref);
            onAddLog({
                type: 'api-success',
                title: `postPref("${trimmedId}") succeeded`,
                details: {sent: parsedPayload, response: resp},
            });
            showStatus('success', `Saved preferences for app ${trimmedId}`);
        } catch (err: any) {
            const errorMsg = err?.response?.data?.message || err?.message || String(err);
            onAddLog({
                type: 'api-error',
                title: `postPref("${trimmedId}") failed`,
                details: errorMsg,
            });
            showStatus('error', `Failed to post preferences: ${errorMsg}`);
        } finally {
            setIsPostingPref(false);
        }
    };

    // Post preferences with a peer client (distinct session source) to verify real-time SSE delivery in a single tab
    const handlePostPeerPref = async () => {
        const trimmedId = targetAppId.trim();
        if (!trimmedId) {
            showStatus('error', 'Please enter or select an App UUID.');
            return;
        }
        let parsedPayload: unknown;
        try {
            parsedPayload = JSON.parse(prefPayloadText);
        } catch {
            showStatus('error', 'Invalid JSON payload. Please verify syntax.');
            return;
        }

        setIsPostingPeerPref(true);
        try {
            // Create a lightweight peer client instance using current config or token with distinct source
            console.log((client as any).config)
            const peerClient = createPrefSyncClient({
                basePath: config.basePath || window.location.origin,
                accessToken: config.accessToken || '',
            });

            const resp = await peerClient.postPref(trimmedId, parsedPayload);
            onAddLog({
                type: 'peer-api-success',
                title: `PeerClient.postPref("${trimmedId}") succeeded (triggers SSE onEvent!)`,
                details: {sent: parsedPayload, response: resp},
            });
            showStatus('success', `Posted from peer client. Check the Live Event Log for onEvent!`);
        } catch (err: any) {
            const errorMsg = err?.response?.data?.message || err?.message || String(err);
            onAddLog({
                type: 'api-error',
                title: `PeerClient.postPref("${trimmedId}") failed`,
                details: errorMsg,
            });
            showStatus('error', `Peer client post failed: ${errorMsg}`);
        } finally {
            setIsPostingPeerPref(false);
        }
    };

    // Generic helper for SDK API calls
    const handleApiCall = async (actionName: string, callFn: () => Promise<unknown>) => {
        setIsCallingApi(true);
        try {
            const res = await callFn();
            setApiResult({title: actionName, data: res});
            onAddLog({
                type: 'api-success',
                title: `${actionName} succeeded`,
                details: res,
            });
            showStatus('success', `${actionName} succeeded`);
        } catch (err: any) {
            const errorMsg = err?.response?.data?.message || err?.message || String(err);
            setApiResult({title: actionName, data: {error: errorMsg}});
            onAddLog({
                type: 'api-error',
                title: `${actionName} failed`,
                details: errorMsg,
            });
            showStatus('error', `${actionName} failed: ${errorMsg}`);
        } finally {
            setIsCallingApi(false);
        }
    };

    const setTemplate = (template: Record<string, unknown>) => {
        setPrefPayloadText(JSON.stringify(template, null, 2));
    };

    // Register the auto-refresh handler with the parent via callback ref
    React.useEffect(() => {
        if (onAutoRefreshRegister) {
            onAutoRefreshRegister((data: unknown) => {
                const formattedPref = normalizePref(data);
                setCurrentPref(formattedPref);
            });
        }
    }, [onAutoRefreshRegister]);

    return (
        <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(360px, 1fr))',
            gap: '20px',
            marginTop: '20px'
        }}>
            {/* App Workflow & Preference Management Card */}
            <div style={cardStyle}>
                {/* Header & Status Indicator */}
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '16px'
                }}>
                    <h2 style={{fontSize: '1.25rem', fontWeight: 600}}>Preference Manager</h2>
                    <div style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
            <span
                style={{
                    display: 'inline-block',
                    width: '10px',
                    height: '10px',
                    borderRadius: '50%',
                    backgroundColor: isSubscribed ? 'var(--success)' : 'var(--danger)',
                    boxShadow: isSubscribed ? '0 0 8px var(--success)' : 'none',
                }}
            />
                        <span style={{
                            fontSize: '0.875rem',
                            color: isSubscribed ? 'var(--success)' : 'var(--text-secondary)'
                        }}>
              {isSubscribed ? `Stream Subscribed: ${appId ? `${appId.slice(0, 8)}...` : 'None'}` : 'Not Subscribed'}
            </span>
                    </div>
                </div>

                {/* Status Toast Banner */}
                {statusMessage && (
                    <div
                        style={{
                            padding: '10px 14px',
                            borderRadius: '6px',
                            marginBottom: '16px',
                            fontSize: '0.875rem',
                            backgroundColor: statusMessage.type === 'success' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                            border: `1px solid ${statusMessage.type === 'success' ? 'var(--success)' : 'var(--danger)'}`,
                            color: statusMessage.type === 'success' ? '#34d399' : '#f87171',
                        }}
                    >
                        {statusMessage.text}
                    </div>
                )}

                {/* App Creation & Selection */}
                <div style={{
                    backgroundColor: 'var(--input-bg)',
                    padding: '14px',
                    borderRadius: '8px',
                    border: '1px solid var(--card-border)',
                    marginBottom: '16px'
                }}>
                    <label style={{...labelStyle, color: 'var(--accent)'}}>App Management & Selection</label>
                    <div style={{display: 'flex', gap: '8px', marginBottom: '10px'}}>
                        <input
                            type="text"
                            value={newAppName}
                            onChange={(e) => setNewAppName(e.target.value)}
                            placeholder="App Name (e.g. Workspace Settings)"
                            style={{...inputStyle, flex: 2}}
                        />
                        <button
                            onClick={handleCreateApp}
                            disabled={isCreatingApp}
                            style={{...buttonStyle, backgroundColor: '#8b5cf6', flex: 1, whiteSpace: 'nowrap'}}
                        >
                            {isCreatingApp ? 'Creating...' : '+ Create App'}
                        </button>
                    </div>

                    {/* Target App ID Input with Authorise button */}
                    <div style={{display: 'flex', gap: '8px', alignItems: 'center'}}>
                        <input
                            type="text"
                            value={targetAppId}
                            onChange={(e) => {
                                setTargetAppId(e.target.value);
                                onSelectAppId(e.target.value);
                            }}
                            placeholder="Target App UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
                            style={{...inputStyle, fontFamily: 'var(--font-mono)', fontSize: '0.85rem'}}
                        />
                        <button
                            type="button"
                            onClick={() => handleAuthoriseApp(targetAppId)}
                            disabled={isCallingApi || !targetAppId.trim()}
                            style={{...secondaryButtonStyle, padding: '10px 12px', whiteSpace: 'nowrap'}}
                            title="Authorise app for current user"
                        >
                            Authorise
                        </button>
                    </div>

                    {/* List of Created App UUIDs */}
                    {createdApps.length > 0 && (
                        <div style={{
                            display: 'flex',
                            flexWrap: 'wrap',
                            gap: '6px',
                            marginTop: '10px',
                            alignItems: 'center'
                        }}>
                            <span style={{fontSize: '0.75rem', color: 'var(--text-secondary)'}}>Created Apps:</span>
                            {createdApps.map((a) => (
                                <button
                                    key={a.id}
                                    type="button"
                                    onClick={() => {
                                        setTargetAppId(a.id);
                                        onSelectAppId(a.id);
                                    }}
                                    style={{
                                        ...chipButtonStyle,
                                        borderColor: targetAppId === a.id ? 'var(--accent)' : 'var(--card-border)',
                                        color: targetAppId === a.id ? 'var(--accent)' : 'var(--text-secondary)',
                                    }}
                                    title={a.id}
                                >
                                    {a.name} ({a.id.slice(0, 8)}...)
                                </button>
                            ))}
                        </div>
                    )}
                </div>

                {/* Action Buttons for Get / Post / Peer Post */}
                <div style={{display: 'flex', gap: '10px', marginBottom: '16px', flexWrap: 'wrap'}}>
                    <button
                        onClick={() => handleGetPref(targetAppId)}
                        disabled={isFetchingPref}
                        style={{...buttonStyle, backgroundColor: '#3b82f6', flex: 1, minWidth: '130px'}}
                    >
                        {isFetchingPref ? 'Fetching...' : 'Get Preferences'}
                    </button>
                    <button
                        onClick={handlePostPref}
                        disabled={isPostingPref}
                        style={{...buttonStyle, backgroundColor: 'var(--success)', flex: 1, minWidth: '130px'}}
                    >
                        {isPostingPref ? 'Saving...' : 'Post Preferences'}
                    </button>
                    <button
                        onClick={handlePostPeerPref}
                        disabled={isPostingPeerPref}
                        style={{...buttonStyle, backgroundColor: '#f59e0b', flex: 1, minWidth: '160px'}}
                        title="Post from a peer session with a distinct source to trigger onEvent in this tab"
                    >
                        {isPostingPeerPref ? 'Posting...' : '⚡ Post via Peer Client'}
                    </button>
                </div>

                {/* Current Fetched Prefs */}
                {currentPref !== null && (
                    <div style={{marginBottom: '16px'}}>
                        <div style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            marginBottom: '4px'
                        }}>
                            <label style={labelStyle}>Remote Preference Value</label>
                            {currentPref.updated_at && (
                                <span style={{fontSize: '0.75rem', color: 'var(--text-secondary)'}}>
                  Updated: {new Date(currentPref.updated_at).toLocaleTimeString()}
                </span>
                            )}
                        </div>
                        <pre style={codeBlockStyle}>{JSON.stringify(currentPref.data, null, 2)}</pre>
                    </div>
                )}

                {/* Payload Editor */}
                <div style={{marginBottom: '12px'}}>
                    <div style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        marginBottom: '6px'
                    }}>
                        <label style={labelStyle}>Payload (JSON)</label>
                        <div style={{display: 'flex', gap: '6px'}}>
                            <button
                                type="button"
                                onClick={() => setTemplate({theme: 'dark', fontSize: 16, autoSave: true})}
                                style={chipButtonStyle}
                            >
                                Theme
                            </button>
                            <button
                                type="button"
                                onClick={() => setTemplate({view: 'grid', sortBy: 'updatedAt', itemsPerPage: 50})}
                                style={chipButtonStyle}
                            >
                                UI Layout
                            </button>
                            <button
                                type="button"
                                onClick={() => setTemplate({
                                    timestamp: new Date().toISOString(),
                                    randomSeed: Math.random()
                                })}
                                style={chipButtonStyle}
                            >
                                Timestamp
                            </button>
                        </div>
                    </div>
                    <textarea
                        rows={5}
                        value={prefPayloadText}
                        onChange={(e) => setPrefPayloadText(e.target.value)}
                        style={{...inputStyle, fontFamily: 'var(--font-mono)', fontSize: '0.85rem'}}
                    />
                </div>

                {/* Auto Refresh Toggle on Event */}
                <div style={{display: 'flex', alignItems: 'center', gap: '8px', marginTop: '10px'}}>
                    <label style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        cursor: 'pointer',
                        fontSize: '0.85rem',
                        color: 'var(--text-secondary)'
                    }}>
                        <input
                            type="checkbox"
                            checked={autoRefreshOnEvent}
                            onChange={(e) => setAutoRefreshOnEvent(e.target.checked)}
                            style={{width: '15px', height: '15px', accentColor: 'var(--accent)'}}
                        />
                        Auto-fetch preferences on incoming <code
                        style={{color: 'var(--accent)'}}>onEvent</code> notification
                    </label>
                </div>
            </div>

            {/* API Explorer Card */}
            <div style={cardStyle}>
                <h2 style={{fontSize: '1.25rem', fontWeight: 600, marginBottom: '12px'}}>SDK Client Explorer</h2>
                <p style={{fontSize: '0.875rem', color: 'var(--text-secondary)', marginBottom: '16px'}}>
                    Directly invoke operations on the active <code
                    style={{color: 'var(--accent)'}}>PrefSyncClient</code> instance.
                </p>

                <div
                    style={{display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '10px', marginBottom: '20px'}}>
                    <button
                        onClick={() => handleApiCall('getUser()', () => client.getUser())}
                        disabled={isCallingApi}
                        style={secondaryButtonStyle}
                    >
                        Get User
                    </button>
                    <button
                        onClick={() => handleApiCall(`getApp("${targetAppId}")`, () => client.getApp(targetAppId.trim()))}
                        disabled={isCallingApi || !targetAppId.trim()}
                        style={secondaryButtonStyle}
                    >
                        Get App Info
                    </button>
                    <button
                        onClick={() => handleApiCall(`authoriseApp("${targetAppId}")`, () => client.authoriseApp(targetAppId.trim()))}
                        disabled={isCallingApi || !targetAppId.trim()}
                        style={secondaryButtonStyle}
                    >
                        Authorise App
                    </button>
                    <button
                        onClick={() => handleApiCall(`deauthoriseApp("${targetAppId}")`, () => client.deauthoriseApp(targetAppId.trim()))}
                        disabled={isCallingApi || !targetAppId.trim()}
                        style={secondaryButtonStyle}
                    >
                        Deauthorise App
                    </button>
                </div>

                {apiResult && (
                    <div>
                        <div style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            marginBottom: '6px'
                        }}>
                            <label style={labelStyle}>Last Result: {apiResult.title}</label>
                            <button onClick={() => setApiResult(null)}
                                    style={{...chipButtonStyle, color: 'var(--danger)'}}>
                                Dismiss
                            </button>
                        </div>
                        <pre style={codeBlockStyle}>{JSON.stringify(apiResult.data, null, 2)}</pre>
                    </div>
                )}
            </div>
        </div>
    );
};

// Main App Component with Provider, Auth Configuration, and Event Log
export const App = () => {
    // Draft configuration states (to prevent reconnect-on-every-keystroke)
    const defaultBasePath = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5173';
    const [draftBasePath, setDraftBasePath] = useState<string>(defaultBasePath);
    const [draftAccessToken, setDraftAccessToken] = useState<string>('');
    const [showToken, setShowToken] = useState<boolean>(false);
    const [isFetchingToken, setIsFetchingToken] = useState<boolean>(false);

    // Active applied configuration
    const [appliedConfig, setAppliedConfig] = useState<{
        basePath: string;
        accessToken: string;
    }>({
        basePath: defaultBasePath,
        accessToken: '',
    });

    // App subscription state
    const [appId, setAppId] = useState<string>('');
    const [enableSubscription, setEnableSubscription] = useState<boolean>(false);
    const [createdApps, setCreatedApps] = useState<CreatedApp[]>([]);
    const [autoRefreshOnEvent, setAutoRefreshOnEvent] = useState<boolean>(true);

    // Ref to hold the dashboard's auto-refresh handler
    const dashboardRefreshRef = React.useRef<(data: unknown) => void>(() => {
    });

    // Event Logs state
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [logFilter, setLogFilter] = useState<'all' | 'event' | 'error' | 'close' | 'api'>('all');

    const isConfigDirty =
        draftBasePath !== appliedConfig.basePath ||
        draftAccessToken !== appliedConfig.accessToken;

    const addLog = useCallback((entry: Omit<LogEntry, 'id' | 'timestamp'>) => {
        const newEntry: LogEntry = {
            id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
            timestamp: new Date().toLocaleTimeString(),
            ...entry,
        };
        setLogs((prev) => [newEntry, ...prev.slice(0, 99)]);
    }, []);

    // Fetch token handler
    const handleGetDevToken = async () => {
        setIsFetchingToken(true);
        try {
            const token = await fetchDevToken(draftBasePath);
            setDraftAccessToken(token);
            setAppliedConfig((prev) => ({
                ...prev,
                basePath: draftBasePath.trim(),
                accessToken: token.trim(),
            }));
            addLog({
                type: 'api-success',
                title: 'Acquired developer JWT from stub IdP (/idp/oidc/token)',
            });
        } catch (err: any) {
            const errorMsg = err?.message || String(err);
            addLog({
                type: 'api-error',
                title: 'Failed to fetch developer token from stub IdP',
                details: errorMsg,
            });
        } finally {
            setIsFetchingToken(false);
        }
    };

    // Apply configuration
    const handleApplyConfig = () => {
        setAppliedConfig({
            basePath: draftBasePath.trim(),
            accessToken: draftAccessToken.trim(),
        });
        addLog({
            type: 'api-success',
            title: `Applied client configuration (basePath: ${draftBasePath.trim()})`,
        });
    };

    // Event handlers for PrefSyncProvider
    const handleEvent = useCallback(
        async (notification: Notification, client: PrefSyncClient) => {
            addLog({
                type: 'event',
                title: `SSE Notification [App: ${notification.App || 'unknown'}, User: ${notification.User || 'unknown'}]`,
                details: notification,
            });

            if (autoRefreshOnEvent && notification.App) {
                try {
                    const freshData = await client.getPref(notification.App);
                    addLog({
                        type: 'api-success',
                        title: `Auto-refreshed preferences on notification for ${notification.App}`,
                        details: freshData,
                    });
                    // Update the displayed preference in the dashboard
                    dashboardRefreshRef.current(freshData);
                } catch {
                    // ignore background refresh error
                }
            }
        },
        [addLog, autoRefreshOnEvent]
    );

    const handleError = useCallback(
        (error: unknown) => {
            const errorMsg = error instanceof Error ? error.message : String(error);
            addLog({
                type: 'error',
                title: 'SSE Stream Error',
                details: errorMsg,
            });
        },
        [addLog]
    );

    const handleClose = useCallback(() => {
        addLog({
            type: 'close',
            title: 'SSE Stream Closed',
        });
    }, [addLog]);

    const filteredLogs = useMemo(() => {
        if (logFilter === 'all') return logs;
        if (logFilter === 'api') return logs.filter((l) => l.type === 'api-success' || l.type === 'api-error' || l.type === 'peer-api-success');
        return logs.filter((l) => l.type === logFilter);
    }, [logs, logFilter]);

    return (
        <div style={{maxWidth: '1200px', margin: '0 auto'}}>
            {/* Header */}
            <header style={{marginBottom: '24px', borderBottom: '1px solid var(--card-border)', paddingBottom: '16px'}}>
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    flexWrap: 'wrap',
                    gap: '12px'
                }}>
                    <div>
                        <h1 style={{
                            fontSize: '1.75rem',
                            fontWeight: 700,
                            display: 'flex',
                            alignItems: 'center',
                            gap: '10px'
                        }}>
                            <span style={{fontSize: '1.75rem'}}>🔄</span> Preference Sync React Demo
                        </h1>
                        <p style={{color: 'var(--text-secondary)', fontSize: '0.95rem', marginTop: '4px'}}>
                            Interactive demonstration of <code
                            style={{color: 'var(--accent)'}}>PrefSyncProvider</code>,{' '}
                            <code style={{color: 'var(--accent)'}}>usePrefSync</code>, and real-time SSE stream
                            synchronization.
                        </p>
                    </div>
                    <div style={{display: 'flex', gap: '8px'}}>
            <span style={{
                fontSize: '0.8rem',
                padding: '4px 10px',
                borderRadius: '9999px',
                backgroundColor: 'var(--card-border)'
            }}>
              React 19
            </span>
                        <span style={{
                            fontSize: '0.8rem',
                            padding: '4px 10px',
                            borderRadius: '9999px',
                            backgroundColor: 'var(--card-border)'
                        }}>
              Port 7086 / Proxied
            </span>
                    </div>
                </div>
            </header>

            {/* Provider Configuration Section */}
            <section style={{...cardStyle, marginBottom: '24px'}}>
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '16px',
                    flexWrap: 'wrap',
                    gap: '10px'
                }}>
                    <div>
                        <h2 style={{fontSize: '1.25rem', fontWeight: 600}}>Provider Configuration</h2>
                        <p style={{fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '2px'}}>
                            Edit values and click &quot;Apply Configuration&quot; to establish or update the SDK
                            connection.
                        </p>
                    </div>
                    <div style={{display: 'flex', gap: '10px'}}>
                        <button
                            type="button"
                            onClick={handleGetDevToken}
                            disabled={isFetchingToken}
                            style={{...buttonStyle, backgroundColor: '#6366f1'}}
                            title="Obtain a JWT from the embedded stub IdP via client_credentials"
                        >
                            {isFetchingToken ? 'Fetching Token...' : '🔑 Get Dev Token (Stub IdP)'}
                        </button>
                        <button
                            type="button"
                            onClick={handleApplyConfig}
                            style={{
                                ...buttonStyle,
                                backgroundColor: isConfigDirty ? 'var(--accent)' : '#475569',
                                color: isConfigDirty ? '#0f172a' : '#ffffff',
                            }}
                        >
                            {isConfigDirty ? '⚡ Apply Configuration' : '✓ Connected / Applied'}
                        </button>
                    </div>
                </div>

                <div
                    style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '16px'}}>
                    <div>
                        <label style={labelStyle}>Server Base Path</label>
                        <input
                            type="text"
                            value={draftBasePath}
                            onChange={(e) => setDraftBasePath(e.target.value)}
                            placeholder="e.g. http://localhost:5173 or http://127.0.0.1:7086"
                            style={inputStyle}
                        />
                    </div>

                    <div>
                        <div style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            marginBottom: '6px'
                        }}>
                            <label htmlFor="access-token-input" style={{...labelStyle, marginBottom: 0}}>
                                Access Token (In-Memory Only)
                            </label>
                            <button
                                type="button"
                                aria-label={showToken ? 'Hide token' : 'Show token'}
                                onClick={() => setShowToken(!showToken)}
                                style={{...chipButtonStyle, padding: '1px 6px', fontSize: '0.75rem'}}
                            >
                                {showToken ? 'Hide' : 'Show'}
                            </button>
                        </div>
                        <input
                            id="access-token-input"
                            type={showToken ? 'text' : 'password'}
                            value={draftAccessToken}
                            onChange={(e) => setDraftAccessToken(e.target.value)}
                            placeholder="Bearer JWT token or click 'Get Dev Token'"
                            style={inputStyle}
                        />
                    </div>
                </div>

                {/* Toggles */}
                <div style={{
                    display: 'flex',
                    flexWrap: 'wrap',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginTop: '16px',
                    gap: '12px'
                }}>
                    <div style={{display: 'flex', alignItems: 'center', gap: '20px'}}>
                        <label style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            cursor: 'pointer',
                            fontSize: '0.9rem'
                        }}>
                            <input
                                type="checkbox"
                                checked={enableSubscription}
                                onChange={(e) => setEnableSubscription(e.target.checked)}
                                style={{width: '16px', height: '16px', accentColor: 'var(--accent)'}}
                            />
                            Enable Real-Time SSE Subscription
                        </label>

                        <label style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            cursor: 'pointer',
                            fontSize: '0.9rem'
                        }}>
                        </label>
                    </div>
                </div>
            </section>

            {/* Wrapped PrefSyncProvider */}
            <PrefSyncProvider
                config={appliedConfig}
                appId={enableSubscription && appId ? appId : undefined}
                onEvent={handleEvent}
                onError={handleError}
                onClose={handleClose}
            >
                <PrefSyncDashboard
                    onAddLog={addLog}
                    config={appliedConfig}
                    createdApps={createdApps}
                    onAppCreated={(app) => setCreatedApps((prev) => [app, ...prev])}
                    onSelectAppId={(id) => setAppId(id)}
                    autoRefreshOnEvent={autoRefreshOnEvent}
                    setAutoRefreshOnEvent={setAutoRefreshOnEvent}
                    onAutoRefreshRegister={(handler) => {
                        dashboardRefreshRef.current = handler;
                    }}
                />
            </PrefSyncProvider>

            {/* Real-time Event Log Panel */}
            <section style={{...cardStyle, marginTop: '24px'}}>
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '16px',
                    flexWrap: 'wrap',
                    gap: '10px'
                }}>
                    <div style={{display: 'flex', alignItems: 'center', gap: '12px'}}>
                        <h2 style={{fontSize: '1.25rem', fontWeight: 600}}>Live Event & Activity Log</h2>
                        <span style={{
                            fontSize: '0.8rem',
                            padding: '2px 8px',
                            borderRadius: '12px',
                            backgroundColor: 'var(--card-border)'
                        }}>
              {logs.length} entries
            </span>
                    </div>

                    <div style={{display: 'flex', alignItems: 'center', gap: '10px'}}>
                        {/* Filter Chips */}
                        <div style={{display: 'flex', gap: '6px'}}>
                            {(['all', 'event', 'error', 'close', 'api'] as const).map((filter) => (
                                <button
                                    key={filter}
                                    onClick={() => setLogFilter(filter)}
                                    style={{
                                        ...chipButtonStyle,
                                        backgroundColor: logFilter === filter ? 'var(--accent)' : 'transparent',
                                        color: logFilter === filter ? '#0f172a' : 'var(--text-secondary)',
                                        fontWeight: logFilter === filter ? 600 : 400,
                                    }}
                                >
                                    {filter.toUpperCase()}
                                </button>
                            ))}
                        </div>

                        <button
                            onClick={() => setLogs([])}
                            style={{...chipButtonStyle, color: 'var(--danger)', borderColor: 'var(--card-border)'}}
                        >
                            Clear Logs
                        </button>
                    </div>
                </div>

                {/* Log Entries */}
                <div
                    style={{
                        maxHeight: '320px',
                        overflowY: 'auto',
                        backgroundColor: 'var(--input-bg)',
                        borderRadius: '8px',
                        border: '1px solid var(--card-border)',
                        padding: '12px',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '8px',
                    }}
                >
                    {filteredLogs.length === 0 ? (
                        <div style={{
                            textAlign: 'center',
                            padding: '32px 0',
                            color: 'var(--text-secondary)',
                            fontSize: '0.9rem'
                        }}>
                            No log events recorded yet. Perform an action or trigger preference changes to see live
                            activity.
                        </div>
                    ) : (
                        filteredLogs.map((log) => (
                            <div
                                key={log.id}
                                style={{
                                    padding: '10px 12px',
                                    borderRadius: '6px',
                                    backgroundColor: 'var(--card-bg)',
                                    border: '1px solid var(--card-border)',
                                    borderLeft: `4px solid ${
                                        log.type === 'event'
                                            ? 'var(--success)'
                                            : log.type === 'error'
                                                ? 'var(--danger)'
                                                : log.type === 'close'
                                                    ? 'var(--warning)'
                                                    : log.type === 'api-success'
                                                        ? '#3b82f6'
                                                        : '#f87171'
                                    }`,
                                }}
                            >
                                <div style={{
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center',
                                    marginBottom: log.details ? '6px' : '0'
                                }}>
                                    <div style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
                    <span
                        style={{
                            fontSize: '0.7rem',
                            fontWeight: 700,
                            padding: '2px 6px',
                            borderRadius: '4px',
                            textTransform: 'uppercase',
                            backgroundColor:
                                log.type === 'event'
                                    ? 'rgba(16, 185, 129, 0.2)'
                                    : log.type === 'error'
                                        ? 'rgba(239, 68, 68, 0.2)'
                                        : log.type === 'close' || log.type === 'peer-api-success'
                                            ? 'rgba(245, 158, 11, 0.2)'
                                            : log.type === 'api-success'
                                                ? 'rgba(59, 130, 246, 0.2)'
                                                : 'rgba(248, 113, 113, 0.2)',
                            color:
                                log.type === 'event'
                                    ? 'var(--success)'
                                    : log.type === 'error'
                                        ? 'var(--danger)'
                                        : log.type === 'close' || log.type === 'peer-api-success'
                                            ? 'var(--warning)'
                                            : log.type === 'api-success'
                                                ? '#3b82f6'
                                                : '#f87171',
                        }}
                    >
                      {log.type}
                    </span>
                                        <span style={{fontSize: '0.9rem', fontWeight: 600}}>{log.title}</span>
                                    </div>
                                    <span style={{
                                        fontSize: '0.75rem',
                                        color: 'var(--text-secondary)'
                                    }}>{log.timestamp}</span>
                                </div>
                                {log.details !== undefined && (
                                    <pre
                                        style={{
                                            marginTop: '4px',
                                            padding: '8px',
                                            borderRadius: '4px',
                                            backgroundColor: 'var(--code-bg)',
                                            fontSize: '0.8rem',
                                            fontFamily: 'var(--font-mono)',
                                            color: 'var(--text-secondary)',
                                            overflowX: 'auto',
                                        }}
                                    >
                    {typeof log.details === 'string' ? log.details : JSON.stringify(log.details, null, 2)}
                  </pre>
                                )}
                            </div>
                        ))
                    )}
                </div>
            </section>
        </div>
    );
};
