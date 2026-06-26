import React from 'react';
import {Card, CardActions, CardContent, Chip, Collapse, Divider, IconButton, Switch, Tooltip, Typography} from '@mui/material';
import {CheckCircleOutlined, ContentCopy, Delete, Edit, ErrorOutlined, ExpandMore, Http, PlayArrow, Storage, Terminal, Warning} from '@mui/icons-material';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {formatDateTime, formatDateTimeHM} from '../../../lib/utilities';

export type ConnStatus = 'ok' | 'error' | 'unknown';

// Base card that renders common chrome, header, actions, footer with expand, and collapse area.
// Specific cards provide summary (children) and details (detailsChildren) content.
export const ConnectionCardBase = ({
                                       id,
                                       title,
                                       icon,
                                       chipLabel,
                                       chipVariant = 'outlined',
                                       status,
                                       lastChecked,
                                       expanded,
                                       onToggleExpand,
                                       onTest,
                                       onEdit,
                                       onDelete,
                                       onDuplicate,
                                       enabled,
                                       onToggleEnabled,
                                       suspended,
                                       extraActions,
                                       children,
                                       detailsChildren,
                                       kind,
                                   }: {
    id: string | number;
    title: string;
    icon?: React.ReactNode;
    chipLabel?: string;
    chipVariant?: 'filled' | 'outlined';
    status: ConnStatus;
    lastChecked?: string;
    expanded: boolean;
    onToggleExpand: () => void;
    onTest?: (ev: React.MouseEvent<HTMLElement>) => void;
    onEdit?: () => void;
    onDelete?: () => void;
    onDuplicate?: () => void;
    enabled?: boolean;
    onToggleEnabled?: () => void;
    suspended?: boolean;
    extraActions?: React.ReactNode;
    children?: React.ReactNode; // summary content under header
    detailsChildren?: React.ReactNode; // content inside collapse area
    kind?: 'database' | 'shell' | 'webtask';
}) => {
    const statusChip = (() => {
        if (suspended) {
            return (
                <Tooltip title="Suspended. This connection is temporarily inactive.">
                    <Chip size="small" color="warning" icon={<Warning fontSize="small"/>} label="Suspended"/>
                </Tooltip>
            );
        }
        const baseTitle = 'Result of the last connectivity check (manual test or auto-check)';
        switch (status) {
            case 'ok':
                return (
                    <Tooltip title={`${baseTitle}: OK. Click “Test” to refresh.`}>
                        <Chip size="small" color="success" icon={<CheckCircleOutlined fontSize="small"/>} label="OK"/>
                    </Tooltip>
                );
            case 'error':
                return (
                    <Tooltip title={`${baseTitle}: Error. Click “Test” to re-check.`}>
                        <Chip size="small" color="error" icon={<ErrorOutlined fontSize="small"/>} label="Error"/>
                    </Tooltip>
                );
            default:
                return (
                    <Tooltip title={`${baseTitle}: Unknown (not checked yet).`}>
                        <Chip size="small" label="Unknown"/>
                    </Tooltip>
                );
        }
    })();

    const borderColor = kind === 'database' ? '#1976d2' : kind === 'shell' ? '#9c27b0' : kind === 'webtask' ? '#ed6c02' : 'divider';

    return (
        <Card
            variant="outlined"
            sx={{
                borderRadius: 3,
                borderLeft: kind ? `6px solid ${borderColor}` : undefined,
                opacity: (enabled === false || suspended) ? 0.65 : 1,
                bgcolor: (enabled === false || suspended) ? 'action.hover' : 'background.paper'
            }}
        >
            <CardContent sx={{pb: 1}}>
                <HStack justifyContent="space-between" alignItems="center" sx={{gap: 2, flexWrap: 'wrap'}}>
                    <HStack alignItems="center" sx={{gap: 1.5, minWidth: 240, flex: 1}}>
                        {icon}
                        <Typography variant="subtitle1" sx={{
                            fontWeight: 600
                        }}>{title}</Typography>
                        {chipLabel && (
                            <Chip size="small" variant={chipVariant} label={chipLabel} sx={{textTransform: 'capitalize'}}/>
                        )}
                        {statusChip}
                        {enabled === false && !suspended && (
                            <Chip size="small" label="Disabled" variant="outlined" sx={{opacity: 0.8}}/>
                        )}
                        <Tooltip title="Timestamp of the last connectivity check recorded by the server">
                            <Typography variant="caption" sx={{
                                color: "text.secondary"
                            }}>
                                Last checked: {lastChecked ? formatDateTimeHM(lastChecked) : 'Never'}
                            </Typography>
                        </Tooltip>
                    </HStack>
                    <HStack alignItems="center" sx={{gap: 0.5}}>
                        {extraActions}
                        {onToggleEnabled && (
                            <Tooltip title={suspended ? "Suspended" : (enabled ? "Enabled" : "Disabled")}>
                    <span>
                        <Switch size="small" checked={!!enabled && !suspended} onChange={onToggleEnabled} disabled={suspended}/>
                    </span>
                            </Tooltip>
                        )}
                        {onTest && (
                            <Tooltip title={suspended ? "Cannot test suspended connection" : "Test"}>
                <span>
                  <IconButton size="small" onClick={(ev) => onTest?.(ev)} disabled={suspended}>
                    <PlayArrow/>
                  </IconButton>
                </span>
                            </Tooltip>
                        )}
                        {onDuplicate && (
                            <Tooltip title="Duplicate Connection">
                                <IconButton size="small" onClick={onDuplicate}><ContentCopy fontSize="small"/></IconButton>
                            </Tooltip>
                        )}
                        {onEdit && (
                            <Tooltip title={suspended ? "Cannot edit suspended connection" : "Edit"}>
                <span>
                  <IconButton size="small" onClick={onEdit} disabled={suspended}><Edit/></IconButton>
                </span>
                            </Tooltip>
                        )}
                        {onDelete && (
                            <Tooltip title="Delete">
                                <IconButton size="small" onClick={onDelete}><Delete/></IconButton>
                            </Tooltip>
                        )}
                    </HStack>
                </HStack>

                {children}
            </CardContent>
            <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
            <CardActions sx={{justifyContent: 'space-between'}}>
                <Typography
                    variant="caption"
                    sx={{
                        color: "text.secondary",
                        ml: 1
                    }}>
                    ID: {String(id)}
                </Typography>
                <IconButton size="small" onClick={onToggleExpand} aria-label="expand">
                    <ExpandMore sx={{transform: expanded ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s'}}/>
                </IconButton>
            </CardActions>
            <Collapse in={expanded} timeout="auto" unmountOnExit>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                <CardContent>
                    {detailsChildren}
                </CardContent>
            </Collapse>
        </Card>
    );
};

// Database card using the base
export type DbConnRow = {
    id: string;
    name: string;
    driver: 'mysql' | 'postgres' | 'sqlite' | 'mssql' | 'oracle' | 'snowflake';
    host: string;
    port?: string;
    database?: string;
    username?: string;
    description?: string;
    status: ConnStatus;
    lastChecked?: string;
    lastError?: string;
    maskedDsn?: string;
    autoCheckEnabled?: boolean;
    autoCheckSeconds?: number;
    agentUuid?: string;
    agentName?: string;
    enabled?: boolean;
    suspended?: boolean;
};

function formatEvery(seconds?: number): string {
    const s = Math.max(60, Math.min(86400, Number(seconds || 0)));
    if (s % 86400 === 0) return `${s / 86400} day` + (s === 86400 ? '' : 's');
    if (s % 43200 === 0) return `${s / 43200} half-day`;
    if (s % 21600 === 0) return `${s / 21600} × 6 hours`;
    if (s % 3600 === 0) return `${s / 3600} hour` + (s === 3600 ? '' : 's');
    if (s % 1800 === 0) return `${s / 1800} × 30 minutes`;
    if (s % 900 === 0) return `${s / 900} × 15 minutes`;
    if (s % 300 === 0) return `${s / 300} × 5 minutes`;
    return `${s / 60} minute` + (s === 60 ? '' : 's');
}

export const DatabaseConnectionCard = ({
                                           row,
                                           expanded,
                                           onToggleExpand,
                                           onTest,
                                           onEdit,
                                           onDelete,
                                           onDuplicate,
                                           onToggleEnabled,
                                           extraActions,
                                       }: {
    row: DbConnRow;
    expanded: boolean;
    onToggleExpand: () => void;
    onTest?: (ev: React.MouseEvent<HTMLElement>) => void;
    onEdit?: () => void;
    onDelete?: () => void;
    onDuplicate?: () => void;
    onToggleEnabled?: () => void;
    extraActions?: React.ReactNode;
}) => {
    return (
        <ConnectionCardBase
            id={row.id}
            title={row.name}
            kind="database"
            icon={<Storage fontSize="small" sx={{color: '#1976d2'}}/>}
            chipLabel={row.driver}
            status={row.status}
            lastChecked={row.lastChecked}
            expanded={expanded}
            onToggleExpand={onToggleExpand}
            onTest={onTest}
            onEdit={onEdit}
            onDelete={onDelete}
            onDuplicate={onDuplicate}
            enabled={row.enabled}
            onToggleEnabled={onToggleEnabled}
            suspended={row.suspended}
            extraActions={extraActions}
            detailsChildren={(
                <>
                    <Typography variant="subtitle2" gutterBottom>Details</Typography>
                    <VStack spacing={0.5}>
                        <Typography variant="body2">Driver: {row.driver}</Typography>
                        <Typography variant="body2">Host: {row.host}</Typography>
                        {row.port && <Typography variant="body2">Port: {row.port}</Typography>}
                        {row.agentName && (
                            <Typography variant="body2" sx={{
                                color: "text.secondary"
                            }}>Via Agent: {row.agentName}</Typography>
                        )}
                        {row.database && <Typography variant="body2">Database: {row.database}</Typography>}
                        {row.username && <Typography variant="body2">Username: {row.username}</Typography>}
                        {row.description && <Typography variant="body2">Description: {row.description}</Typography>}
                        <Typography variant="body2">
                            Auto-check: {row.autoCheckEnabled ? (
                            <>Every {formatEvery(row.autoCheckSeconds || 3600)}</>
                        ) : 'Off'}
                        </Typography>
                        <Typography variant="body2">
                            Last test: {row.lastChecked ? (
                            <>
                                {formatDateTime(row.lastChecked)} — <span style={{color: row.status === 'ok' ? '#2e7d32' : row.status === 'error' ? '#d32f2f' : '#ed6c02', fontWeight: 600}}>
                {row.status === 'ok' ? 'OK' : row.status === 'error' ? (row.lastError || 'Error') : 'Unknown'}
              </span>
                            </>
                        ) : 'Never'}
                        </Typography>
                    </VStack>
                </>
            )}
        >
            {/* Summary */}
            <Typography variant="body2" sx={{mt: 1}}>
                Host: {row.host || '—'}{row.port ? `:${row.port}` : ''}{row.database ? ` / ${row.database}` : ''}
            </Typography>
            {row.username && (
                <Typography variant="body2" sx={{
                    color: "text.secondary"
                }}>User: {row.username}</Typography>
            )}
            {row.agentName && (
                <Typography variant="body2" sx={{
                    color: "text.secondary"
                }}>Via Agent: {row.agentName}</Typography>
            )}
            {row.description && (
                <Typography variant="body2" sx={{mt: 1}}>
                    {row.description}
                </Typography>
            )}
        </ConnectionCardBase>
    );
};

// Shell card using the base
export type ShellConnRow = {
    id: number | string;
    name: string;
    description?: string;
    agent_uuid?: string;
    agent_name?: string;
    mode: 'localhost' | 'ssh';
    host?: string;
    port?: number | string;
    ssh_username?: string;
    lastStatus?: string;
    lastError?: string;
    lastCheckedAt?: string;
    enabled?: boolean;
    suspended?: boolean;
};

export const ShellConnectionCard = ({
                                        row,
                                        expanded,
                                        onToggleExpand,
                                        onTest,
                                        onEdit,
                                        onDelete,
                                        onDuplicate,
                                        onToggleEnabled,
                                    }: {
    row: ShellConnRow;
    expanded: boolean;
    onToggleExpand: () => void;
    onTest?: (ev: React.MouseEvent<HTMLElement>) => void;
    onEdit?: () => void;
    onDelete?: () => void;
    onDuplicate?: () => void;
    onToggleEnabled?: () => void;
}) => {
    const status: ConnStatus = (row.lastStatus && row.lastStatus.toLowerCase() === 'ok') ? 'ok' : ((row.lastStatus && row.lastStatus.toLowerCase() === 'error') || row.lastError) ? 'error' : 'unknown';
    const hostPort = row.mode === 'ssh' ? `${row.host || '—'}${row.port ? `:${row.port}` : ''}` : 'Localhost';
    return (
        <ConnectionCardBase
            id={row.id}
            title={row.name}
            kind="shell"
            icon={<Terminal fontSize="small" sx={{color: '#9c27b0'}}/>}
            chipLabel={row.mode}
            status={status}
            lastChecked={row.lastCheckedAt}
            expanded={expanded}
            onToggleExpand={onToggleExpand}
            onTest={onTest}
            onEdit={onEdit}
            onDelete={onDelete}
            onDuplicate={onDuplicate}
            enabled={row.enabled}
            onToggleEnabled={onToggleEnabled}
            suspended={row.suspended}
            detailsChildren={(
                <>
                    <Typography variant="subtitle2" gutterBottom>Details</Typography>
                    <VStack spacing={0.5}>
                        <Typography variant="body2">Mode: {row.mode}</Typography>
                        {row.host && <Typography variant="body2">Host: {row.host}</Typography>}
                        {row.port != null && <Typography variant="body2">Port: {String(row.port)}</Typography>}
                        {row.ssh_username && <Typography variant="body2">SSH user: {row.ssh_username}</Typography>}
                        {(row.agent_name || row.agent_uuid) && <Typography variant="body2">Agent: {row.agent_name || row.agent_uuid}</Typography>}
                        {row.lastError && <Typography variant="body2" color="error">Last error: {row.lastError}</Typography>}
                    </VStack>
                </>
            )}
        >
            {/* Summary */}
            <Typography variant="body2" sx={{mt: 1}}>
                {row.mode === 'ssh' ? `SSH to ${hostPort}` : 'Localhost'}
            </Typography>
            {row.ssh_username && (
                <Typography variant="body2" sx={{
                    color: "text.secondary"
                }}>User: {row.ssh_username}</Typography>
            )}
            {(row.agent_name || row.agent_uuid) && (
                <Typography variant="body2" sx={{
                    color: "text.secondary"
                }}>Agent: {row.agent_name || row.agent_uuid}</Typography>
            )}
            {row.description && (
                <Typography variant="body2" sx={{mt: 1}}>
                    {row.description}
                </Typography>
            )}
        </ConnectionCardBase>
    );
};

export type WebtaskConnRow = {
    id: number | string;
    name: string;
    description?: string;
    baseUrl?: string;
    authType: 'none' | 'basic' | 'bearer' | 'header';
    agentUuid?: string;
    agentName?: string;
    lastStatus?: string;
    lastError?: string;
    lastCheckedAt?: string;
    enabled?: boolean;
    suspended?: boolean;
};

export const WebtaskConnectionCard = ({
                                          row,
                                          expanded,
                                          onToggleExpand,
                                          onTest,
                                          onEdit,
                                          onDelete,
                                          onDuplicate,
                                          onToggleEnabled,
                                      }: {
    row: WebtaskConnRow;
    expanded: boolean;
    onToggleExpand: () => void;
    onTest?: (ev: React.MouseEvent<HTMLElement>) => void;
    onEdit?: () => void;
    onDelete?: () => void;
    onDuplicate?: () => void;
    onToggleEnabled?: () => void;
}) => {
    const status: ConnStatus = (row.lastStatus && row.lastStatus.toLowerCase() === 'ok') ? 'ok' : ((row.lastStatus && row.lastStatus.toLowerCase() === 'error') || row.lastError) ? 'error' : 'unknown';
    return (
        <ConnectionCardBase
            id={row.id}
            title={row.name}
            kind="webtask"
            icon={<Http fontSize="small" sx={{color: '#ed6c02'}}/>}
            chipLabel={''}
            status={status}
            lastChecked={row.lastCheckedAt}
            expanded={expanded}
            onToggleExpand={onToggleExpand}
            onTest={onTest}
            onEdit={onEdit}
            onDelete={onDelete}
            onDuplicate={onDuplicate}
            enabled={row.enabled}
            onToggleEnabled={onToggleEnabled}
            suspended={row.suspended}
            detailsChildren={(
                <>
                    <Typography variant="subtitle2" gutterBottom>Details</Typography>
                    <VStack spacing={0.5}>
                        <Typography variant="body2">Auth Type: {row.authType}</Typography>
                        {row.baseUrl && <Typography variant="body2">Base URL: {row.baseUrl}</Typography>}
                        {(row.agentName || row.agentUuid) && <Typography variant="body2">Agent: {row.agentName || row.agentUuid}</Typography>}
                        {row.lastError && <Typography variant="body2" color="error">Last error: {row.lastError}</Typography>}
                    </VStack>
                </>
            )}
        >
            {/* Summary */}
            <Typography variant="body2" sx={{mt: 1}}>
                {row.baseUrl ? `Base URL: ${row.baseUrl}` : 'No Base URL'}
            </Typography>
            {(row.agentName || row.agentUuid) && (
                <Typography variant="body2" sx={{
                    color: "text.secondary"
                }}>Via Agent: {row.agentName || row.agentUuid}</Typography>
            )}
            {row.description && (
                <Typography variant="body2" sx={{mt: 1}}>
                    {row.description}
                </Typography>
            )}
        </ConnectionCardBase>
    );
};
