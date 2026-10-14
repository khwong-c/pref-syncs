import React from "react";

export const S: Record<string, React.CSSProperties> = {
    container: {maxWidth: 700, margin: '0 auto', padding: '32px 16px'},
    title: {fontSize: '1.5rem', fontWeight: 700, marginBottom: 24},
    subtitle: {fontSize: '1rem', fontWeight: 600, marginBottom: 12},
    card: {background: '#fff', border: '1px solid #e5e7eb', borderRadius: 8, padding: 16, marginBottom: 16},
    text: {fontSize: '0.9rem', marginBottom: 4},
    error: {color: '#dc2626', fontSize: '0.9rem'},
    button: {
        padding: '8px 16px',
        borderRadius: 6,
        border: 'none',
        background: '#2563eb',
        color: '#fff',
        fontWeight: 600,
        fontSize: '0.9rem',
        cursor: 'pointer',
        marginBottom: 8
    },
    buttonDisabled: {
        padding: '8px 16px',
        borderRadius: 6,
        border: 'none',
        background: '#abbde5',
        color: '#fff',
        fontWeight: 600,
        fontSize: '0.9rem',
        cursor: 'pointer',
        marginBottom: 8
    },
    hint: {fontSize: '0.8rem', color: '#6b7280', marginBottom: 8},
    log: {maxHeight: 300, overflowY: 'auto'},
    logEntry: {padding: 8, marginBottom: 8, background: '#f9fafb', borderRadius: 4, border: '1px solid #e5e7eb'},
    code: {
        fontSize: '0.8rem',
        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-word'
    },
};