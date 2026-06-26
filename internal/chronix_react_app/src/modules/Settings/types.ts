// Shared types for Settings module

export type SettingsLoginResponse = {
    ok: boolean;
    message?: string;
    token?: string; // if backend returns a session token/cookie also used
};

export interface GlobalEmailSettings {
    smtpHost: string;
    smtpPort: string; // keep as string for easier input; backend can coerce
    fromName?: string;
    fromEmail?: string;
    smtpLogin?: string;
    smtpPassword?: string;
    secure?: 'none' | 'ssl' | 'starttls';
}

export interface GlobalHttpsSettings {
    mode: 'selfsigned' | 'upload';
    port: string; // typically 443
    enabled?: boolean;
    // Upload mode placeholders (frontend: filenames only; backend handles upload endpoints)
    certFileName?: string;
    keyFileName?: string;
    certInfo?: { subject: string; issuer: string; notBefore?: string; notAfter?: string; validity?: string };
}

export interface GlobalHttpSettings {
    enabled: boolean;
    port: string; // typically 80
}

export interface GlobalAgentSettings {
    enabled: boolean;
    port: string; // typically 5172
}

export type SmsProvider = 'none' | 'twilio';

export interface GlobalSmsSettings {
    provider: SmsProvider;
    fromNumber?: string;
    accountSid?: string;
    authToken?: string;
}

export interface GlobalAlertSettings {
    systemAlertEmails: string;
    systemAlertPhones: string;
    alertOnAgentLost: boolean;
}

export interface BrandingSettings {
    brandLogoUrl: string;
    brandName: string;
}

export interface ServerUrlSettings {
    serverUrl: string;
}

export interface WebhookSettingsItem {
    id: number;
    name: string;
    url: string;
    secret?: string;
    events: string;
    enabled: boolean;
    createdAt: string;
    updatedAt: string;
}

export interface GlobalSettings {
    serverUrl: string;
    email?: GlobalEmailSettings;
    https?: GlobalHttpsSettings;
    http?: GlobalHttpSettings;
    agent?: GlobalAgentSettings;
    sms?: GlobalSmsSettings;
    systemAlertEmails?: string;
    systemAlertPhones?: string;
}

export interface SettingsUser {
    id: string;
    displayName: string;
    email: string;
    phone?: string;
    timeZone?: string;
    timeFormat?: '12h' | '24h';
    disabled?: boolean;
    isAdmin?: boolean;
    forcePasswordChange?: boolean;
    suspended?: boolean;
    created_at?: string;
    last_login_at?: string;
}

export interface SettingsUserActivity {
    id: string;
    userId: string;
    user?: string;
    when: string; // ISO
    action: string;
    details?: string;
}

export interface SettingsSummary {
    serverUrl: string;
    http: { enabled: boolean; port: number | string };
    https: {
        enabled: boolean;
        port: number | string;
        mode: string;
        hasUploadedCert: boolean;
        hasUploadedKey: boolean;
        certInfo?: { subject: string; issuer: string; notBefore: string; notAfter: string };
    };
    email: { smtpHost?: string; smtpPort?: number | string; fromName?: string; fromEmail?: string; secure?: string; configured?: boolean };
    sms: { provider?: string; fromNumber?: string; configured?: boolean };
}

export interface ActiveSetting {
    setting: string;
    value: string;
    description: string;
}

export type UpdaterVersionInfo = {
    version: string;
    release_date: string;
    release_notes: string;
};

export type UpdaterStatus = {
    currentVersion: string;
    availableVersion?: {
        server?: UpdaterVersionInfo;
        'chronix-agent'?: UpdaterVersionInfo;
    };
    lastCheckTime: string;
    enabled: boolean;
    mode: string;
    windowStart: string;
    agentEnabled: boolean;
    agentMode: string;
    agentWindowStart: string;
};

export type UpdateAgentInfo = {
    uuid: string;
    name: string;
    version: string;
    status: string;
    online: boolean;
    lastSeenAt?: string;
};
