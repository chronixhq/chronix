import type {DbConnectionDraft, DbDriver} from '../types.ts'

export const DEFAULT_DB_CONNECTIONS: Record<DbDriver, Partial<DbConnectionDraft>> = {
    postgres: {port: '5432', sslEnabled: true, sslMode: 'prefer'},
    mysql: {port: '3306', sslEnabled: false},
    sqlite: {},
    mssql: {port: '1433', trustServerCertificate: true},
    oracle: {port: '1521'},
    snowflake: {},
}

function sanitizeParameterKey(value?: string): string {
    return (value || '').replaceAll(/\s+/g, '')
}

function appendExtraParams(params: URLSearchParams, raw?: string) {
    if (!raw) return
    raw.split(/[&\n]/).forEach((pair) => {
        const [key, val] = pair.split('=')
        if (key) params.set(sanitizeParameterKey(key), val ?? '')
    })
}

export function buildDatabaseConnectionDsn(draft: DbConnectionDraft, options?: { preview?: boolean }): string {
    const preview = !!options?.preview
    const extraParams = draft.params || draft.extraParams
    switch (draft.driver) {
        case 'postgres': {
            const user = draft.username ? encodeURIComponent(draft.username) : ''
            const pass = draft.password ? (preview ? ':***' : `:${encodeURIComponent(draft.password)}`) : ''
            const auth = user ? `${user}${pass}@` : ''
            const host = draft.host ?? ''
            const port = draft.port ? `:${draft.port}` : ''
            const database = draft.database ? `/${draft.database}` : ''
            const params = new URLSearchParams()
            if (draft.sslEnabled === false) params.set('sslmode', 'disable')
            if (draft.sslEnabled) params.set('sslmode', draft.sslMode || 'prefer')
            appendExtraParams(params, extraParams)
            const query = params.toString()
            return `postgresql://${auth}${host}${port}${database}${query ? `?${query}` : ''}`
        }
        case 'mysql': {
            const user = draft.username ? encodeURIComponent(draft.username) : ''
            const pass = draft.password ? (preview ? ':***' : `:${encodeURIComponent(draft.password)}`) : ''
            const auth = user ? `${user}${pass}@` : ''
            const host = draft.host ?? ''
            const port = draft.port ? `:${draft.port}` : ''
            const database = draft.database ? `/${draft.database}` : ''
            const params = new URLSearchParams()
            if (draft.sslEnabled) params.set('tls', 'true')
            appendExtraParams(params, extraParams)
            const query = params.toString()
            return `mysql://${auth}${host}${port}${database}${query ? `?${query}` : ''}`
        }
        case 'sqlite':
            return draft.filePath ? `file:${draft.filePath}` : 'file:'
        case 'oracle': {
            const user = draft.username ? encodeURIComponent(draft.username) : ''
            const pass = draft.password ? (preview ? ':***' : `:${encodeURIComponent(draft.password)}`) : ''
            const auth = user ? `${user}${pass}@` : ''
            const host = draft.host ?? ''
            const port = draft.port ? `:${draft.port}` : ''
            const database = draft.database ? `/${draft.database}` : ''
            return `oracle://${auth}${host}${port}${database}`
        }
        case 'snowflake': {
            const user = draft.username ? encodeURIComponent(draft.username) : ''
            const pass = draft.password ? (preview ? ':***' : `:${encodeURIComponent(draft.password)}`) : ''
            const auth = user ? `${user}${pass}@` : ''
            const account = draft.host ?? ''
            const database = draft.database ? `/${draft.database}` : ''
            const params = new URLSearchParams()
            appendExtraParams(params, extraParams)
            const query = params.toString()
            return `snowflake://${auth}${account}${database}${query ? `?${query}` : ''}`
        }
        case 'mssql': {
            const user = draft.username ? encodeURIComponent(draft.username) : ''
            const pass = draft.password ? (preview ? ':***' : `:${encodeURIComponent(draft.password)}`) : ''
            const auth = user ? `${user}${pass}@` : ''
            const host = draft.host ?? ''
            const port = draft.port ? `:${draft.port}` : ''
            const params = new URLSearchParams()
            if (draft.database) params.set('database', draft.database)
            if (draft.trustServerCertificate) params.set('TrustServerCertificate', 'true')
            appendExtraParams(params, extraParams)
            const query = params.toString()
            return `sqlserver://${auth}${host}${port}${query ? `?${query}` : ''}`
        }
        default:
            return ''
    }
}

export function buildDatabaseConnectionPreviewDsn(draft: DbConnectionDraft): string {
    return buildDatabaseConnectionDsn(draft, {preview: true})
}
