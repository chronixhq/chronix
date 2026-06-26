import {describe, expect, it} from 'vitest';
import {
    createDefaultWebtaskConnectionDraft,
    normalizeLoadedWebtaskConnection,
    snapshotWebtaskConnectionDraft,
    validateWebtaskConnectionDraft,
} from './webtaskConnectionEditorUtils';

describe('webtaskConnectionEditorUtils', () => {
    it('creates the default draft state', () => {
        expect(createDefaultWebtaskConnectionDraft()).toEqual({
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
        });
    });

    it('normalizes loaded connections', () => {
        expect(normalizeLoadedWebtaskConnection({
            id: '1',
            kind: 'webtask',
            name: 'Prod',
            authType: 'bearer',
            authConfig: null,
            autoCheckEnabled: false,
            createdAt: '',
            updatedAt: '',
        } as any)).toEqual(expect.objectContaining({
            name: 'Prod',
            authType: 'bearer',
            authConfig: {},
            autoCheckSeconds: 300,
        }));
    });

    it('creates stable draft snapshots', () => {
        const snapshot = snapshotWebtaskConnectionDraft({
            name: 'Prod',
            authType: 'header',
            authConfig: {header_name: 'X-API-Key', header_value: 'abc'},
        });

        expect(JSON.parse(snapshot)).toEqual(expect.objectContaining({
            name: 'Prod',
            authType: 'header',
            authConfig: {header_name: 'X-API-Key', header_value: 'abc'},
        }));
    });

    it('validates required auth values for new drafts', () => {
        expect(validateWebtaskConnectionDraft({
            ...createDefaultWebtaskConnectionDraft(),
            name: 'Prod',
            authType: 'basic',
            authConfig: {username: 'user', password: ''},
        })).toEqual({password: 'Password is required.'});
    });

    it('allows redacted saved secrets for edits', () => {
        expect(validateWebtaskConnectionDraft({
            ...createDefaultWebtaskConnectionDraft(),
            name: 'Prod',
            authType: 'bearer',
            authConfig: {token: '<redacted>'},
        }, {allowRedactedSecrets: true})).toEqual({});
    });
});
