import {apiDelete, apiGet, apiPost, apiPut} from '@dsherwin/react-api-interface'
import type {AnyConnection, ConnectionKind, DbConnection, ShellConnection, WebtaskConnection} from './types'
import type {DbConnRow, ShellConnRow, WebtaskConnRow} from './components/ConnectionCards'

export type ConnectionStatus = 'ok' | 'error' | 'unknown'

export interface AllConnectionsRow {
    id: string
    kind: ConnectionKind
    name: string
    description?: string
    status: ConnectionStatus
    lastChecked?: string
    lastError?: string
    hostPort?: string
    agent?: string
    driver?: DbConnection['driver']
    host?: string
    port?: string
    database?: string
    username?: string
    agentName?: string
    autoCheckEnabled?: boolean
    autoCheckSeconds?: number
    mode?: ShellConnection['mode']
    ssh_username?: string
    enabled?: boolean
    suspended?: boolean
    agent_uuid?: string
    agent_name?: string
    baseUrl?: string
    authType?: WebtaskConnection['authType']
}

export interface ConnectionTarget {
    id: string | number
    kind: ConnectionKind
}

function getConnectionNormalizer(kind: ConnectionKind) {
    if (kind === 'database') return normalizeDbConnection
    if (kind === 'shell') return normalizeShellConnection
    return normalizeWebtaskConnection
}

export function getConnectionApiCollectionPath(kind: ConnectionKind): string {
    if (kind === 'database') return '/connections'
    if (kind === 'shell') return '/shell/connections'
    return '/connections/webtask'
}

export function getConnectionApiItemPath(kind: ConnectionKind, id: string | number): string {
    return `${getConnectionApiCollectionPath(kind)}/${encodeURIComponent(String(id))}`
}

export function getConnectionListPath(kind: ConnectionKind): string {
    if (kind === 'database') return '/databases/list'
    if (kind === 'shell') return '/shell/list'
    return '/webtasks/list'
}

export function getConnectionCreatePath(kind: ConnectionKind): string {
    if (kind === 'database') return '/databases/add'
    if (kind === 'shell') return '/shell/add'
    return '/webtasks/add'
}

export function getConnectionEditPath(kind: ConnectionKind, id: string | number): string {
    if (kind === 'database') return `/databases/edit/${encodeURIComponent(String(id))}`
    if (kind === 'shell') return `/shell/edit/${encodeURIComponent(String(id))}`
    return `/webtasks/edit/${encodeURIComponent(String(id))}`
}

function normalizeCommonConnection(raw: any) {
    return {
        id: String(raw?.id ?? ''),
        name: String(raw?.name ?? ''),
        description: raw?.description ?? undefined,
        agentUuid: raw?.agentUuid ?? raw?.agent_uuid ?? undefined,
        autoCheckEnabled: raw?.autoCheckEnabled ?? raw?.auto_check_enabled ?? undefined,
        autoCheckSeconds: raw?.autoCheckSeconds ?? raw?.auto_check_interval_seconds ?? undefined,
        alertEmails: raw?.alertEmails ?? raw?.alert_emails ?? undefined,
        alertPhones: raw?.alertPhones ?? raw?.alert_phones ?? undefined,
        notifyOnFailure: raw?.notifyOnFailure ?? raw?.notify_on_failure ?? undefined,
        enabled: raw?.enabled ?? undefined,
        suspended: raw?.suspended ?? undefined,
        lastStatus: raw?.lastStatus ?? undefined,
        lastError: raw?.lastError ?? undefined,
        lastCheckedAt: raw?.lastCheckedAt ?? undefined,
        createdAt: raw?.createdAt ?? raw?.created_at ?? undefined,
        updatedAt: raw?.updatedAt ?? raw?.updated_at ?? undefined,
    }
}

export function getConnectionStatus(connection: {lastStatus?: string; lastError?: string}): ConnectionStatus {
    if (connection.lastStatus?.toLowerCase() === 'ok') return 'ok'
    if (connection.lastStatus?.toLowerCase() === 'error' || connection.lastError) return 'error'
    return 'unknown'
}

export function normalizeDbConnection(raw: any): DbConnection {
    const common = normalizeCommonConnection(raw)
    return {
        ...common,
        kind: 'database',
        driver: raw?.driver,
        host: raw?.host ?? undefined,
        port: raw?.port != null ? String(raw.port) : undefined,
        database: raw?.database ?? undefined,
        username: raw?.username ?? undefined,
        password: raw?.password ?? undefined,
        hasPassword: raw?.hasPassword ?? undefined,
        filePath: raw?.filePath ?? raw?.file_path ?? undefined,
        sslEnabled: raw?.sslEnabled ?? raw?.ssl_enabled ?? undefined,
        sslMode: raw?.sslMode ?? raw?.ssl_mode ?? undefined,
        trustServerCertificate: raw?.trustServerCertificate ?? raw?.trust_server_certificate ?? undefined,
        params: raw?.params ?? undefined,
        extraParams: raw?.extraParams ?? raw?.extra_params ?? undefined,
    }
}

export function normalizeShellConnection(raw: any): ShellConnection {
    const common = normalizeCommonConnection(raw)
    return {
        ...common,
        kind: 'shell',
        mode: raw?.mode ?? 'localhost',
        run_as_user: raw?.run_as_user ?? undefined,
        sudo: raw?.sudo ?? undefined,
        host: raw?.host ?? undefined,
        port: raw?.port ?? undefined,
        ssh_username: raw?.ssh_username ?? undefined,
        auth_method: raw?.auth_method ?? undefined,
    }
}

export function normalizeWebtaskConnection(raw: any): WebtaskConnection {
    const common = normalizeCommonConnection(raw)
    return {
        ...common,
        kind: 'webtask',
        baseUrl: raw?.baseUrl ?? raw?.base_url ?? undefined,
        authType: raw?.authType ?? raw?.auth_type ?? 'none',
        authConfig: raw?.authConfig ?? raw?.auth_config ?? undefined,
    }
}

export async function fetchConnectionsByKind(kind: ConnectionKind): Promise<AnyConnection[]> {
    const data = await apiGet(getConnectionApiCollectionPath(kind))
    if (!Array.isArray(data)) return []
    if (kind === 'database') return data.map(normalizeDbConnection)
    if (kind === 'shell') return data.map(normalizeShellConnection)
    return data.map(normalizeWebtaskConnection)
}

export async function fetchAllConnections(): Promise<AnyConnection[]> {
    const [database, shell, webtask] = await Promise.all([
        fetchConnectionsByKind('database'),
        fetchConnectionsByKind('shell'),
        fetchConnectionsByKind('webtask'),
    ])
    return [...database, ...shell, ...webtask]
}

export async function fetchConnectionById<TConnection extends AnyConnection = AnyConnection>(kind: ConnectionKind, id: string | number): Promise<TConnection> {
    const data = await apiGet(getConnectionApiItemPath(kind, id))
    const normalize = getConnectionNormalizer(kind)
    return normalize(data) as TConnection
}

export async function testStoredConnection({kind, id}: ConnectionTarget): Promise<unknown> {
    return apiPost(`${getConnectionApiItemPath(kind, id)}/test`, {})
}

export async function duplicateStoredConnection({kind, id}: ConnectionTarget): Promise<any> {
    return apiPost(`${getConnectionApiItemPath(kind, id)}/duplicate`, {})
}

export async function deleteStoredConnection({kind, id}: ConnectionTarget): Promise<unknown> {
    return apiDelete(getConnectionApiItemPath(kind, id))
}

export async function setStoredConnectionEnabled({kind, id}: ConnectionTarget, enabled: boolean): Promise<unknown> {
    return apiPut(getConnectionApiItemPath(kind, id), {enabled} as any)
}

export function applyConnectionHealthPatch(connection: AnyConnection, patch: {id?: string | number; lastStatus?: string; lastError?: string | null; lastCheckedAt?: string | Date}): AnyConnection {
    if (String(patch.id ?? '') !== connection.id) return connection
    const lastCheckedAt = patch.lastCheckedAt
        ? (typeof patch.lastCheckedAt === 'string' ? patch.lastCheckedAt : patch.lastCheckedAt.toString())
        : connection.lastCheckedAt
    return {
        ...connection,
        lastStatus: patch.lastStatus ?? connection.lastStatus,
        lastError: patch.lastError !== undefined ? (patch.lastError || undefined) : connection.lastError,
        lastCheckedAt,
    }
}

export function toDatabaseConnectionRow(connection: DbConnection): DbConnRow {
    return {
        id: connection.id,
        name: connection.name,
        driver: connection.driver,
        host: connection.host || '',
        port: connection.port,
        database: connection.database,
        username: connection.username,
        description: connection.description,
        status: getConnectionStatus(connection),
        lastChecked: connection.lastCheckedAt,
        lastError: connection.lastError,
        maskedDsn: (connection as any).maskedDsn || undefined,
        autoCheckEnabled: connection.autoCheckEnabled,
        autoCheckSeconds: connection.autoCheckSeconds,
        agentUuid: connection.agentUuid || undefined,
        agentName: (connection as any).agentName || undefined,
        enabled: connection.enabled,
        suspended: connection.suspended,
    }
}

export function toShellConnectionRow(connection: ShellConnection): ShellConnRow {
    return {
        id: connection.id,
        name: connection.name,
        description: connection.description,
        agent_uuid: connection.agentUuid || undefined,
        agent_name: (connection as any).agent_name || (connection as any).agentName || undefined,
        mode: connection.mode,
        host: connection.host,
        port: connection.port,
        ssh_username: connection.ssh_username,
        lastStatus: connection.lastStatus,
        lastError: connection.lastError,
        lastCheckedAt: connection.lastCheckedAt,
        enabled: connection.enabled,
        suspended: connection.suspended,
    }
}

export function toWebtaskConnectionRow(connection: WebtaskConnection): WebtaskConnRow {
    return {
        id: connection.id,
        name: connection.name,
        description: connection.description,
        baseUrl: connection.baseUrl,
        authType: connection.authType,
        agentUuid: connection.agentUuid || undefined,
        agentName: (connection as any).agentName || undefined,
        lastStatus: connection.lastStatus,
        lastError: connection.lastError,
        lastCheckedAt: connection.lastCheckedAt,
        enabled: connection.enabled,
        suspended: connection.suspended,
    }
}

export function toAllConnectionsRow(connection: AnyConnection): AllConnectionsRow {
    if (connection.kind === 'database') {
        return {
            id: connection.id,
            kind: 'database',
            name: connection.name,
            description: connection.description,
            status: getConnectionStatus(connection),
            lastChecked: connection.lastCheckedAt,
            lastError: connection.lastError,
            hostPort: [connection.host ?? '', connection.port ?? ''].filter(Boolean).join(':'),
            agent: (connection as any).agentName || connection.agentUuid || undefined,
            driver: connection.driver,
            host: connection.host,
            port: connection.port,
            database: connection.database,
            username: connection.username,
            agentName: (connection as any).agentName || undefined,
            autoCheckEnabled: connection.autoCheckEnabled,
            autoCheckSeconds: connection.autoCheckSeconds,
            enabled: connection.enabled,
            suspended: connection.suspended,
        }
    }
    if (connection.kind === 'shell') {
        return {
            id: connection.id,
            kind: 'shell',
            name: connection.name,
            description: connection.description,
            status: getConnectionStatus(connection),
            lastChecked: connection.lastCheckedAt,
            lastError: connection.lastError,
            hostPort: [connection.host ?? '', connection.port != null ? String(connection.port) : ''].filter(Boolean).join(':'),
            agent: (connection as any).agent_name || connection.agentUuid || undefined,
            mode: connection.mode,
            host: connection.host,
            port: connection.port != null ? String(connection.port) : undefined,
            ssh_username: connection.ssh_username,
            enabled: connection.enabled,
            suspended: connection.suspended,
            agent_uuid: connection.agentUuid || undefined,
            agent_name: (connection as any).agent_name || undefined,
        }
    }
    return {
        id: connection.id,
        kind: 'webtask',
        name: connection.name,
        description: connection.description,
        status: getConnectionStatus(connection),
        lastChecked: connection.lastCheckedAt,
        lastError: connection.lastError,
        hostPort: connection.baseUrl || undefined,
        agent: (connection as any).agentName || connection.agentUuid || undefined,
        baseUrl: connection.baseUrl,
        authType: connection.authType,
        enabled: connection.enabled,
        suspended: connection.suspended,
    }
}
