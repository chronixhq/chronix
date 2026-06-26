import {describe, expect, it} from 'vitest'
import {buildDatabaseConnectionDsn, buildDatabaseConnectionPreviewDsn} from './databaseConnectionForm.ts'

describe('databaseConnectionForm', () => {
    it('masks passwords in preview DSNs while preserving other connection details', () => {
        const draft = {
            name: 'Primary Postgres',
            driver: 'postgres' as const,
            host: 'db.example.com',
            port: '5432',
            database: 'chronix',
            username: 'app_user',
            password: 'super-secret',
            sslEnabled: true,
            sslMode: 'require' as const,
            params: 'application_name=chronix',
        }

        expect(buildDatabaseConnectionDsn(draft)).toBe(
            'postgresql://app_user:super-secret@db.example.com:5432/chronix?sslmode=require&application_name=chronix',
        )
        expect(buildDatabaseConnectionPreviewDsn(draft)).toBe(
            'postgresql://app_user:***@db.example.com:5432/chronix?sslmode=require&application_name=chronix',
        )
    })

    it('falls back to legacy extraParams when params is absent', () => {
        const draft = {
            name: 'SQL Server',
            driver: 'mssql' as const,
            host: 'sql.internal',
            port: '1433',
            username: 'chronix',
            password: 'pw',
            database: 'Chronix',
            trustServerCertificate: true,
            extraParams: 'encrypt=true&connection timeout=5',
        }

        expect(buildDatabaseConnectionDsn(draft)).toBe(
            'sqlserver://chronix:pw@sql.internal:1433?database=Chronix&TrustServerCertificate=true&encrypt=true&connectiontimeout=5',
        )
    })
})
