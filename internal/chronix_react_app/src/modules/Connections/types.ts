// Shared types for Database DatabaseConnectionsList module

export type DbDriver = 'postgres' | 'mysql' | 'sqlite' | 'mssql' | 'oracle' | 'snowflake';

export type SSLMode = 'disable' | 'prefer' | 'require' | 'verify-ca' | 'verify-full';

export type ConnectionKind = 'database' | 'shell' | 'webtask';

// Common connection fields
export interface CommonConnection {
    id: string;
    name: string;
    kind: ConnectionKind;
    description?: string;
    agentUuid?: string | null;
    autoCheckEnabled?: boolean;
    autoCheckSeconds?: number;
    alertEmails?: string;
    alertPhones?: string;
    notifyOnFailure?: boolean;
    enabled?: boolean;
    suspended?: boolean;
    lastStatus?: string;
    lastError?: string;
    lastCheckedAt?: string;
    createdAt?: string;
    updatedAt?: string;
}

export interface ShellConnection extends CommonConnection {
    kind: 'shell';
    mode: 'localhost' | 'ssh';
    run_as_user?: string;
    sudo?: boolean;
    host?: string;
    port?: number;
    ssh_username?: string;
    auth_method?: 'password' | 'key';
}

export type AnyConnection = DbConnection | ShellConnection | WebtaskConnection;

// API model of a stored database connection
export interface DbConnection extends CommonConnection {
    kind: 'database';
    driver: DbDriver;
    host?: string;
    port?: string;
    database?: string;
    username?: string;
    // password is not returned by API typically; keep optional typing for drafts
    password?: string;
    hasPassword?: boolean;
    // driver-specific
    filePath?: string; // sqlite
    sslEnabled?: boolean; // pg/mysql
    sslMode?: SSLMode; // pg
    trustServerCertificate?: boolean; // sqlserver
    params?: string; // free-form k=v&k2=v2
    extraParams?: string; // alias for params if API differs
}

// Draft used in create/edit forms
export interface DbConnectionDraft {
    name: string; // user-facing label
    driver: DbDriver;
    description?: string; // long, optional description shown in lists/details
    host?: string;
    port?: string;
    database?: string;
    username?: string;
    password?: string;
    hasPassword?: boolean;
    // driver-specific
    filePath?: string; // for sqlite only
    sslEnabled?: boolean; // pg/mysql
    sslMode?: SSLMode; // pg
    trustServerCertificate?: boolean; // sqlserver
    params?: string; // free-form k=v&k2=v2
    extraParams?: string; // DSN params as key=value; pairs (alias)
    autoCheckEnabled?: boolean;
    autoCheckSeconds?: number;
    agentUuid?: string | null; // when set, route via this agent
    alertEmails?: string;
    alertPhones?: string;
    notifyOnFailure?: boolean;
}

export type WebTaskAuthType = 'none' | 'basic' | 'bearer' | 'header';

export interface WebtaskConnection extends CommonConnection {
    kind: 'webtask';
    baseUrl?: string;
    authType: WebTaskAuthType;
    authConfig?: any;
}

export interface WebtaskConnectionDraft {
    name: string;
    description?: string;
    baseUrl?: string;
    authType: WebTaskAuthType;
    // For basic: { username, password }
    // For bearer: { token }
    // For header: { name, value }
    authConfig?: any;
    agentUuid?: string | null;
    autoCheckEnabled?: boolean;
    autoCheckSeconds?: number;
    alertEmails?: string;
    alertPhones?: string;
    notifyOnFailure?: boolean;
}

