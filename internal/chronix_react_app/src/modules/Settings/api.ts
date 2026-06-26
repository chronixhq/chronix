import {apiDelete, apiGet, apiPost, apiPut} from '@dsherwin/react-api-interface'
import type {
    ActiveSetting,
    BrandingSettings,
    GlobalAlertSettings,
    GlobalAgentSettings,
    GlobalEmailSettings,
    GlobalHttpSettings,
    GlobalHttpsSettings,
    GlobalSmsSettings,
    SettingsSummary,
    ServerUrlSettings,
    UpdateAgentInfo,
    UpdaterStatus,
    WebhookSettingsItem,
} from './types.ts'

type ApiRecord = Record<string, unknown>

function toRecord(value: unknown): ApiRecord {
    return value && typeof value === 'object' ? (value as ApiRecord) : {}
}

function toStringValue(value: unknown, fallback: string = ''): string {
    if (typeof value === 'string') return value
    if (typeof value === 'number') return String(value)
    return fallback
}

function toBooleanValue(value: unknown, fallback: boolean = false): boolean {
    if (typeof value === 'boolean') return value
    if (typeof value === 'number') return value !== 0
    if (typeof value === 'string') return value.toLowerCase() === 'true'
    return fallback
}

function toNumberValue(value: unknown): number | undefined {
    if (typeof value === 'number' && Number.isFinite(value)) return value
    if (typeof value === 'string' && value.trim()) {
        const parsed = Number(value)
        if (Number.isFinite(parsed)) return parsed
    }
    return undefined
}

function toNumberOrString(value: unknown, fallback: string = ''): string | number {
    if (typeof value === 'number' && Number.isFinite(value)) return value
    if (typeof value === 'string') return value
    return fallback
}

function normalizeHttpSettings(raw: unknown): GlobalHttpSettings {
    const record = toRecord(raw)
    return {
        enabled: toBooleanValue(record.enabled),
        port: toStringValue(record.port, '80'),
    }
}

function normalizeHttpsSettings(raw: unknown): GlobalHttpsSettings {
    const record = toRecord(raw)
    const certInfo = toRecord(record.certInfo)
    return {
        enabled: toBooleanValue(record.enabled, true),
        mode: record.mode === 'upload' ? 'upload' : 'selfsigned',
        port: toStringValue(record.port, '443'),
        certFileName: toStringValue(record.certFileName),
        keyFileName: toStringValue(record.keyFileName),
        certInfo: Object.keys(certInfo).length ? {
            subject: toStringValue(certInfo.subject),
            issuer: toStringValue(certInfo.issuer),
            notBefore: toStringValue(certInfo.notBefore),
            notAfter: toStringValue(certInfo.notAfter),
            validity: toStringValue(certInfo.validity),
        } : undefined,
    }
}

function normalizeAgentSettings(raw: unknown): GlobalAgentSettings {
    const record = toRecord(raw)
    return {
        enabled: toBooleanValue(record.enabled, true),
        port: toStringValue(record.port, '5172'),
    }
}

function normalizeEmailSettings(raw: unknown): GlobalEmailSettings {
    const record = toRecord(raw)
    const secure = record.secure === 'ssl' || record.secure === 'starttls' ? record.secure : 'none'
    return {
        smtpHost: toStringValue(record.smtpHost),
        smtpPort: toStringValue(record.smtpPort, '587'),
        fromName: toStringValue(record.fromName),
        fromEmail: toStringValue(record.fromEmail),
        smtpLogin: toStringValue(record.smtpLogin),
        smtpPassword: toStringValue(record.smtpPassword),
        secure,
    }
}

function normalizeSmsSettings(raw: unknown): GlobalSmsSettings {
    const record = toRecord(raw)
    return {
        provider: record.provider === 'twilio' ? 'twilio' : 'none',
        fromNumber: toStringValue(record.fromNumber),
        accountSid: toStringValue(record.accountSid),
        authToken: toStringValue(record.authToken),
    }
}

function normalizeAlertSettings(raw: unknown): GlobalAlertSettings {
    const record = toRecord(raw)
    return {
        systemAlertEmails: toStringValue(record.systemAlertEmails),
        systemAlertPhones: toStringValue(record.systemAlertPhones),
        alertOnAgentLost: toBooleanValue(record.alertOnAgentLost, true),
    }
}

function normalizeServerUrlSettings(raw: unknown): ServerUrlSettings {
    const record = toRecord(raw)
    return {
        serverUrl: toStringValue(record.serverUrl),
    }
}

function normalizeBrandingSettings(raw: unknown): BrandingSettings {
    const record = toRecord(raw)
    return {
        brandLogoUrl: toStringValue(record.brandLogoUrl),
        brandName: toStringValue(record.brandName),
    }
}

function normalizeWebhook(raw: unknown): WebhookSettingsItem {
    const record = toRecord(raw)
    return {
        id: toNumberValue(record.id) ?? 0,
        name: toStringValue(record.name),
        url: toStringValue(record.url),
        secret: toStringValue(record.secret) || undefined,
        events: toStringValue(record.events),
        enabled: toBooleanValue(record.enabled, true),
        createdAt: toStringValue(record.createdAt),
        updatedAt: toStringValue(record.updatedAt),
    }
}

function normalizeUpdaterStatus(raw: unknown): UpdaterStatus {
    const record = toRecord(raw)
    const availableVersion = toRecord(record.availableVersion)
    const serverVersion = toRecord(availableVersion.server)
    const agentVersion = toRecord(availableVersion['chronix-agent'])

    return {
        currentVersion: toStringValue(record.currentVersion),
        availableVersion: {
            server: Object.keys(serverVersion).length ? {
                version: toStringValue(serverVersion.version),
                release_date: toStringValue(serverVersion.release_date),
                release_notes: toStringValue(serverVersion.release_notes),
            } : undefined,
            'chronix-agent': Object.keys(agentVersion).length ? {
                version: toStringValue(agentVersion.version),
                release_date: toStringValue(agentVersion.release_date),
                release_notes: toStringValue(agentVersion.release_notes),
            } : undefined,
        },
        lastCheckTime: toStringValue(record.lastCheckTime),
        enabled: toBooleanValue(record.enabled),
        mode: toStringValue(record.mode),
        windowStart: toStringValue(record.windowStart),
        agentEnabled: toBooleanValue(record.agentEnabled),
        agentMode: toStringValue(record.agentMode),
        agentWindowStart: toStringValue(record.agentWindowStart),
    }
}

function normalizeUpdaterAgent(raw: unknown): UpdateAgentInfo {
    const record = toRecord(raw)
    return {
        uuid: toStringValue(record.uuid),
        name: toStringValue(record.name) || toStringValue(record.uuid),
        version: toStringValue(record.version),
        status: toStringValue(record.status),
        online: toBooleanValue(record.online),
        lastSeenAt: toStringValue(record.lastSeenAt) || undefined,
    }
}

export async function fetchNetworkSettings(): Promise<{
    http: GlobalHttpSettings
    https: GlobalHttpsSettings
    agent: GlobalAgentSettings
}> {
    const [http, https, agent] = await Promise.all([
        apiGet('/settings/settings/http'),
        apiGet('/settings/settings/https'),
        apiGet('/settings/settings/agent'),
    ])

    return {
        http: normalizeHttpSettings(http),
        https: normalizeHttpsSettings(https),
        agent: normalizeAgentSettings(agent),
    }
}

export async function saveHttpSettings(settings: GlobalHttpSettings): Promise<void> {
    await apiPut('/settings/settings/http', settings)
}

export async function saveHttpsSettings(settings: GlobalHttpsSettings): Promise<void> {
    await apiPut('/settings/settings/https', settings)
}

export async function saveAgentSettings(settings: GlobalAgentSettings): Promise<void> {
    await apiPut('/settings/settings/agent', settings)
}

export async function uploadHttpsCertificatePair(cert: File, key: File): Promise<GlobalHttpsSettings> {
    const formData = new FormData()
    formData.append('cert', cert)
    formData.append('key', key)
    const response = await apiPost('/settings/settings/https/upload', formData)
    return normalizeHttpsSettings(response)
}

export async function removeHttpsCertificatePair(): Promise<void> {
    await apiDelete('/settings/settings/https')
}

export async function fetchSettingsSummary(): Promise<SettingsSummary> {
    const raw = toRecord(await apiGet('/settings/settings/summary'))
    const http = toRecord(raw.http)
    const https = toRecord(raw.https)
    const certInfo = toRecord(https.certInfo)
    const email = toRecord(raw.email)
    const sms = toRecord(raw.sms)

    return {
        serverUrl: toStringValue(raw.serverUrl),
        http: {
            enabled: toBooleanValue(http.enabled),
            port: toNumberOrString(http.port, '80'),
        },
        https: {
            enabled: toBooleanValue(https.enabled),
            port: toNumberOrString(https.port, '443'),
            mode: toStringValue(https.mode),
            hasUploadedCert: toBooleanValue(https.hasUploadedCert),
            hasUploadedKey: toBooleanValue(https.hasUploadedKey),
            certInfo: Object.keys(certInfo).length ? {
                subject: toStringValue(certInfo.subject),
                issuer: toStringValue(certInfo.issuer),
                notBefore: toStringValue(certInfo.notBefore),
                notAfter: toStringValue(certInfo.notAfter),
            } : undefined,
        },
        email: {
            smtpHost: toStringValue(email.smtpHost) || undefined,
            smtpPort: toNumberOrString(email.smtpPort, ''),
            fromName: toStringValue(email.fromName) || undefined,
            fromEmail: toStringValue(email.fromEmail) || undefined,
            secure: toStringValue(email.secure) || undefined,
            configured: toBooleanValue(email.configured),
        },
        sms: {
            provider: toStringValue(sms.provider) || undefined,
            fromNumber: toStringValue(sms.fromNumber) || undefined,
            configured: toBooleanValue(sms.configured),
        },
    }
}

export async function fetchActiveSettings(): Promise<ActiveSetting[]> {
    const response = await apiGet('/settings/settings/active')
    const record = toRecord(response)
    const settings = Array.isArray(response)
        ? response
        : Array.isArray(record.settings)
            ? record.settings
            : []

    return settings.map((item) => {
        const normalized = toRecord(item)
        return {
            setting: toStringValue(normalized.setting),
            value: toStringValue(normalized.value),
            description: toStringValue(normalized.description),
        }
    })
}

export async function restartServer(): Promise<void> {
    await apiPost('/settings/restart', {})
}

export async function shutdownServer(): Promise<void> {
    await apiPost('/settings/shutdown', {})
}

export async function restartNetworkListeners(): Promise<void> {
    await apiPost('/settings/restart-network', {})
}

function withCacheBust(path: string): string {
    const separator = path.includes('?') ? '&' : '?'
    return `${path}${separator}_=${Date.now()}`
}

export async function fetchUpdaterStatus(options?: { fresh?: boolean }): Promise<UpdaterStatus> {
    const path = options?.fresh ? withCacheBust('/settings/updater/status') : '/settings/updater/status'
    return normalizeUpdaterStatus(await apiGet(path))
}

export async function fetchUpdaterAgents(): Promise<UpdateAgentInfo[]> {
    const response = await apiGet('/agents')
    return (Array.isArray(response) ? response : []).map(normalizeUpdaterAgent)
}

export async function checkForUpdates(): Promise<ApiRecord> {
    return toRecord(await apiPost('/settings/updater/check', {}))
}

export async function applyUpdate(): Promise<void> {
    await apiPost('/settings/updater/apply', {})
}

export async function saveAppUpdaterSettings(settings: Pick<UpdaterStatus, 'enabled' | 'mode' | 'windowStart'>): Promise<void> {
    await apiPut('/settings/settings/updater/app', settings)
}

export async function saveAgentUpdaterSettings(settings: Pick<UpdaterStatus, 'agentEnabled' | 'agentMode' | 'agentWindowStart'>): Promise<void> {
    await apiPut('/settings/settings/updater/agent', {
        enabled: settings.agentEnabled,
        mode: settings.agentMode,
        windowStart: settings.agentWindowStart,
    })
}

export async function updateAgentNow(uuid: string): Promise<void> {
    await apiPost(`/agents/${uuid}/update`, {})
}

export async function fetchEmailSettings(): Promise<GlobalEmailSettings> {
    return normalizeEmailSettings(await apiGet('/settings/settings/email'))
}

export async function saveEmailSettings(settings: GlobalEmailSettings): Promise<void> {
    await apiPut('/settings/settings/email', settings)
}

export async function testEmailSettings(settings: GlobalEmailSettings): Promise<void> {
    await apiPost('/settings/settings/email/test', settings)
}

export async function fetchSmsSettings(): Promise<GlobalSmsSettings> {
    return normalizeSmsSettings(await apiGet('/settings/settings/sms'))
}

export async function saveSmsSettings(settings: GlobalSmsSettings): Promise<void> {
    await apiPut('/settings/settings/sms', settings)
}

export async function testSmsSettings(settings: GlobalSmsSettings & { testNumber: string }): Promise<void> {
    await apiPost('/settings/settings/sms/test', settings)
}

export async function fetchAlertSettings(): Promise<GlobalAlertSettings> {
    return normalizeAlertSettings(await apiGet('/settings/settings/alerts'))
}

export async function saveAlertSettings(settings: GlobalAlertSettings): Promise<void> {
    await apiPut('/settings/settings/alerts', settings)
}

export async function fetchServerUrlSettings(): Promise<ServerUrlSettings> {
    return normalizeServerUrlSettings(await apiGet('/settings/settings/server-url'))
}

export async function saveServerUrlSettings(settings: ServerUrlSettings): Promise<void> {
    await apiPut('/settings/settings/server-url', settings)
}

export async function fetchBrandingSettings(): Promise<BrandingSettings> {
    return normalizeBrandingSettings(await apiGet('/settings/settings/branding'))
}

export async function saveBrandingSettings(settings: BrandingSettings): Promise<void> {
    await apiPut('/settings/settings/branding', settings)
}

export async function fetchWebhooks(): Promise<WebhookSettingsItem[]> {
    const response = await apiGet('/webhooks')
    return Array.isArray(response) ? response.map(normalizeWebhook) : []
}

export async function createWebhook(payload: Omit<WebhookSettingsItem, 'id' | 'createdAt' | 'updatedAt'>): Promise<void> {
    await apiPost('/webhooks', payload)
}

export async function updateWebhook(id: number, payload: Omit<WebhookSettingsItem, 'id' | 'createdAt' | 'updatedAt'>): Promise<void> {
    await apiPut(`/webhooks/${id}`, payload)
}

export async function deleteWebhook(id: number): Promise<void> {
    await apiDelete(`/webhooks/${id}`)
}

export async function testWebhook(item: WebhookSettingsItem): Promise<{ message?: string }> {
    const response = await apiPost('/webhooks/test', item)
    const record = toRecord(response)
    return {message: toStringValue(record.message) || undefined}
}
