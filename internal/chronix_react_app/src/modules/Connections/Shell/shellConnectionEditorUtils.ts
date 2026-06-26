export type Mode = 'localhost' | 'ssh';
export type AuthMethod = 'password' | 'key';
export type KeyFormat = 'openssh' | 'pkcs8';

export type LoadedShell = {
    id: number;
    name: string;
    description?: string | null;
    agent_uuid?: string | null;
    agent_name?: string | null;
    mode: Mode;
    run_as_user?: string | null;
    sudo?: boolean | null;
    host?: string | null;
    port?: number | null;
    ssh_username?: string | null;
    auth_method?: AuthMethod | null;
    ssh_password?: string | null;
    ssh_private_key?: string | null;
    ssh_key_pass?: string | null;
    sudo_password?: string | null;
    auto_check_enabled?: number | null;
    auto_check_interval_seconds?: number | null;
    alert_emails?: string | null;
    alert_phones?: string | null;
    notify_on_failure?: boolean | null;
    suspended?: boolean | null;
};

export interface ShellConnectionDraft {
    name: string;
    description: string;
    agentUUID: string;
    mode: Mode;
    runAsUser: string;
    sudo: boolean;
    host: string;
    port: string;
    sshUsername: string;
    authMethod: AuthMethod;
    sshPassword: string;
    sshPrivateKey: string;
    sshKeyPass: string;
    sudoPassword: string;
    autoCheckEnabled: boolean;
    autoCheckInterval: string;
    alertEmails: string;
    alertPhones: string;
    notifyOnFailure: boolean;
}

export interface ShellConnectionSecretFlags {
    hasPassword: boolean;
    hasPrivateKey: boolean;
    hasKeyPass: boolean;
    hasSudoPassword: boolean;
}

export interface ShellConnectionUiState {
    showSudoPassword: boolean;
    showSshPassword: boolean;
    showSshKeyPass: boolean;
    generatedPublicKey: string | null;
    copied: boolean;
    keyFormat: KeyFormat;
    showFormatDialog: boolean;
}

export function createDefaultShellConnectionDraft(): ShellConnectionDraft {
    return {
        name: '',
        description: '',
        agentUUID: '',
        mode: 'localhost',
        runAsUser: '',
        sudo: false,
        host: '',
        port: '22',
        sshUsername: '',
        authMethod: 'key',
        sshPassword: '',
        sshPrivateKey: '',
        sshKeyPass: '',
        sudoPassword: '',
        autoCheckEnabled: false,
        autoCheckInterval: '300',
        alertEmails: '',
        alertPhones: '',
        notifyOnFailure: true,
    };
}

export function createDefaultShellConnectionSecretFlags(): ShellConnectionSecretFlags {
    return {
        hasPassword: false,
        hasPrivateKey: false,
        hasKeyPass: false,
        hasSudoPassword: false,
    };
}

export function createDefaultShellConnectionUiState(): ShellConnectionUiState {
    return {
        showSudoPassword: false,
        showSshPassword: false,
        showSshKeyPass: false,
        generatedPublicKey: null,
        copied: false,
        keyFormat: 'openssh',
        showFormatDialog: false,
    };
}

export function canSaveShellConnection(draft: ShellConnectionDraft): boolean {
    if (!draft.name.trim()) return false;
    if (draft.mode === 'ssh' && (!draft.host.trim() || !draft.sshUsername.trim())) return false;
    return true;
}

export function loadedShellToEditorState(data: LoadedShell) {
    return {
        draft: {
            name: data.name,
            description: data.description || '',
            agentUUID: data.agent_uuid || '',
            mode: (data.mode || 'localhost') as Mode,
            runAsUser: data.run_as_user || '',
            sudo: Boolean(data.sudo),
            host: data.host || '',
            port: data.port != null ? String(data.port) : '22',
            sshUsername: data.ssh_username || '',
            authMethod: (data.auth_method || 'key') as AuthMethod,
            sshPassword: '',
            sshPrivateKey: '',
            sshKeyPass: '',
            sudoPassword: '',
            autoCheckEnabled: data.auto_check_enabled === 1,
            autoCheckInterval: String(data.auto_check_interval_seconds || '300'),
            alertEmails: data.alert_emails || '',
            alertPhones: data.alert_phones || '',
            notifyOnFailure: data.notify_on_failure !== false,
        } as ShellConnectionDraft,
        secretFlags: {
            hasPassword: data.ssh_password === '<redacted>',
            hasPrivateKey: data.ssh_private_key === '<redacted>',
            hasKeyPass: data.ssh_key_pass === '<redacted>',
            hasSudoPassword: data.sudo_password === '<redacted>',
        } as ShellConnectionSecretFlags,
    };
}

export function snapshotShellConnectionDraft(draft: ShellConnectionDraft): string {
    return JSON.stringify({
        ...draft,
        port: draft.port || '22',
        autoCheckInterval: draft.autoCheckInterval || '300',
    });
}

export function buildShellConnectionSavePayload(draft: ShellConnectionDraft) {
    const payload: any = {
        name: draft.name,
        description: draft.description || undefined,
        agent_uuid: draft.agentUUID || undefined,
        mode: draft.mode,
        run_as_user: draft.runAsUser || undefined,
        sudo: draft.sudo,
        auto_check_enabled: draft.autoCheckEnabled,
        auto_check_interval_seconds: Number(draft.autoCheckInterval || '300'),
        alert_emails: draft.alertEmails || '',
        alert_phones: draft.alertPhones || '',
        notify_on_failure: draft.notifyOnFailure,
    };

    if (draft.sudoPassword) payload.sudo_password = draft.sudoPassword;

    if (draft.mode === 'ssh') {
        payload.host = draft.host;
        payload.port = Number(draft.port || '22');
        payload.ssh_username = draft.sshUsername;
        payload.auth_method = draft.authMethod;
        if (draft.authMethod === 'password') {
            if (draft.sshPassword) payload.ssh_password = draft.sshPassword;
        } else {
            if (draft.sshPrivateKey) payload.ssh_private_key = draft.sshPrivateKey;
            if (draft.sshKeyPass) payload.ssh_key_pass = draft.sshKeyPass;
        }
    }

    return payload;
}

export function buildShellConnectionTestPayload(
    draft: ShellConnectionDraft,
    options?: {
        id?: string;
        nameFallback?: string;
        secretFlags?: ShellConnectionSecretFlags;
    },
) {
    const payload: any = {
        name: draft.name || options?.nameFallback || '(unsaved) test',
        description: draft.description || undefined,
        agent_uuid: draft.agentUUID || undefined,
        mode: draft.mode,
        run_as_user: draft.runAsUser || undefined,
        sudo: draft.sudo,
    };

    if (options?.id) payload.id = Number(options.id);

    if (draft.sudoPassword) payload.sudo_password = draft.sudoPassword;
    else if (options?.secretFlags?.hasSudoPassword) payload.sudo_password = '<redacted>';

    if (draft.mode === 'ssh') {
        payload.host = draft.host;
        payload.port = Number(draft.port || '22');
        payload.ssh_username = draft.sshUsername;
        payload.auth_method = draft.authMethod;
        if (draft.authMethod === 'password') {
            if (draft.sshPassword) payload.ssh_password = draft.sshPassword;
            else if (options?.secretFlags?.hasPassword) payload.ssh_password = '<redacted>';
        } else {
            if (draft.sshPrivateKey) payload.ssh_private_key = draft.sshPrivateKey;
            else if (options?.secretFlags?.hasPrivateKey) payload.ssh_private_key = '<redacted>';

            if (draft.sshKeyPass) payload.ssh_key_pass = draft.sshKeyPass;
            else if (options?.secretFlags?.hasKeyPass) payload.ssh_key_pass = '<redacted>';
        }
    }

    return payload;
}
