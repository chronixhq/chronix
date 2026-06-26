import {type WebtaskConnection} from '../types.ts';

const REDACTED_SECRET = '<redacted>';

function isEmpty(value?: string | number | null) {
    return value === undefined || value === null || String(value).trim() === '';
}

function hasVisibleSecret(value: any) {
    return !isEmpty(value) && value !== REDACTED_SECRET;
}

export function createDefaultWebtaskConnectionDraft(): Partial<WebtaskConnection> {
    return {
        name: '',
        authType: 'none',
        description: '',
        baseUrl: '',
        authConfig: {},
        autoCheckEnabled: true,
        autoCheckSeconds: 300,
        alertEmails: '',
        alertPhones: '',
        notifyOnFailure: true,
    };
}

export function normalizeLoadedWebtaskConnection(connection: WebtaskConnection): Partial<WebtaskConnection> {
    return {
        ...createDefaultWebtaskConnectionDraft(),
        ...connection,
        authConfig: connection.authConfig || {},
        autoCheckSeconds: connection.autoCheckSeconds ?? (connection as any).autoCheckIntervalSeconds ?? 300,
    };
}

export function snapshotWebtaskConnectionDraft(draft: Partial<WebtaskConnection>): string {
    return JSON.stringify({
        ...createDefaultWebtaskConnectionDraft(),
        ...draft,
        authConfig: draft.authConfig || {},
    });
}

export function validateWebtaskConnectionDraft(
    draft: Partial<WebtaskConnection>,
    options?: {allowRedactedSecrets?: boolean},
): Record<string, string> {
    const nextErrors: Record<string, string> = {};
    const authConfig = draft.authConfig || {};
    const allowRedactedSecrets = options?.allowRedactedSecrets ?? false;

    if (isEmpty(draft.name)) nextErrors.name = 'Give this connection a label.';

    if (draft.authType === 'basic') {
        if (isEmpty(authConfig.username)) nextErrors.username = 'Username is required.';
        if (!hasVisibleSecret(authConfig.password) && !(allowRedactedSecrets && authConfig.password === REDACTED_SECRET)) {
            nextErrors.password = 'Password is required.';
        }
    } else if (draft.authType === 'bearer') {
        if (!hasVisibleSecret(authConfig.token) && !(allowRedactedSecrets && authConfig.token === REDACTED_SECRET)) {
            nextErrors.token = 'Token is required.';
        }
    } else if (draft.authType === 'header') {
        if (isEmpty(authConfig.header_name)) nextErrors.header_name = 'Header name is required.';
        if (!hasVisibleSecret(authConfig.header_value) && !(allowRedactedSecrets && authConfig.header_value === REDACTED_SECRET)) {
            nextErrors.header_value = 'Header value is required.';
        }
    }

    return nextErrors;
}
