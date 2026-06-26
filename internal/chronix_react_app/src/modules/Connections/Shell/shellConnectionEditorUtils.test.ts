import {describe, expect, it} from 'vitest';
import {buildShellConnectionSavePayload, buildShellConnectionTestPayload, canSaveShellConnection, createDefaultShellConnectionDraft, loadedShellToEditorState} from './shellConnectionEditorUtils';

describe('shellConnectionEditorUtils', () => {
    it('maps loaded shell connections to draft and secret flags', () => {
        const mapped = loadedShellToEditorState({
            id: 1,
            name: 'Prod',
            mode: 'ssh',
            ssh_password: '<redacted>',
            ssh_private_key: null,
            ssh_key_pass: '<redacted>',
            sudo_password: '<redacted>',
            auth_method: 'password',
        });

        expect(mapped.draft.name).toBe('Prod');
        expect(mapped.draft.mode).toBe('ssh');
        expect(mapped.secretFlags).toEqual({
            hasPassword: true,
            hasPrivateKey: false,
            hasKeyPass: true,
            hasSudoPassword: true,
        });
    });

    it('builds save payloads without redacted placeholders', () => {
        const draft = {
            ...createDefaultShellConnectionDraft(),
            name: 'Prod',
            mode: 'ssh' as const,
            host: 'example.com',
            sshUsername: 'ubuntu',
            authMethod: 'key' as const,
            sshPrivateKey: 'PRIVATE',
            sudo: true,
            sudoPassword: 'sudo',
        };
        expect(buildShellConnectionSavePayload(draft)).toEqual(expect.objectContaining({
            name: 'Prod',
            mode: 'ssh',
            host: 'example.com',
            ssh_username: 'ubuntu',
            ssh_private_key: 'PRIVATE',
            sudo_password: 'sudo',
        }));
    });

    it('builds test payloads with redacted placeholders for existing secrets', () => {
        const draft = {
            ...createDefaultShellConnectionDraft(),
            name: 'Prod',
            mode: 'ssh' as const,
            host: 'example.com',
            sshUsername: 'ubuntu',
            authMethod: 'password' as const,
        };
        expect(buildShellConnectionTestPayload(draft, {
            id: '2',
            secretFlags: {
                hasPassword: true,
                hasPrivateKey: false,
                hasKeyPass: false,
                hasSudoPassword: false,
            },
        })).toEqual(expect.objectContaining({
            id: 2,
            ssh_password: '<redacted>',
        }));
    });

    it('validates minimum save requirements', () => {
        expect(canSaveShellConnection(createDefaultShellConnectionDraft())).toBe(false);
        expect(canSaveShellConnection({...createDefaultShellConnectionDraft(), name: 'Local'})).toBe(true);
        expect(canSaveShellConnection({...createDefaultShellConnectionDraft(), name: 'SSH', mode: 'ssh'})).toBe(false);
    });
});
