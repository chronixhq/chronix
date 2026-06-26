import {describe, expect, it} from 'vitest'
import {
    applyConnectionHealthPatch,
    getConnectionApiCollectionPath,
    getConnectionApiItemPath,
    getConnectionCreatePath,
    getConnectionEditPath,
    getConnectionListPath,
    normalizeWebtaskConnection,
} from './api.ts'

describe('connections api helpers', () => {
    it('returns the expected api and app paths for each connection kind', () => {
        expect(getConnectionApiCollectionPath('database')).toBe('/connections')
        expect(getConnectionApiCollectionPath('shell')).toBe('/shell/connections')
        expect(getConnectionApiCollectionPath('webtask')).toBe('/connections/webtask')

        expect(getConnectionApiItemPath('shell', 'abc 123')).toBe('/shell/connections/abc%20123')
        expect(getConnectionListPath('database')).toBe('/databases/list')
        expect(getConnectionCreatePath('shell')).toBe('/shell/add')
        expect(getConnectionEditPath('webtask', '42')).toBe('/webtasks/edit/42')
    })

    it('patches connection health updates without changing unrelated rows', () => {
        const connection = normalizeWebtaskConnection({
            id: 42,
            name: 'API',
            authType: 'none',
            lastStatus: 'ok',
            lastCheckedAt: '2026-04-12T00:00:00Z',
        })

        const updated = applyConnectionHealthPatch(connection, {
            id: '42',
            lastStatus: 'error',
            lastError: 'timeout',
            lastCheckedAt: '2026-04-13T00:00:00Z',
        })

        expect(updated.lastStatus).toBe('error')
        expect(updated.lastError).toBe('timeout')
        expect(updated.lastCheckedAt).toBe('2026-04-13T00:00:00Z')

        const untouched = applyConnectionHealthPatch(connection, {id: '99', lastStatus: 'error'})
        expect(untouched).toBe(connection)
    })
})
