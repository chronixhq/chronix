import type React from 'react';
import {useEffect, useMemo, useState} from 'react';
import {apiDelete, apiGet, apiPost} from '@dsherwin/react-api-interface';
import {Alert, Box, Chip, CircularProgress, Divider, IconButton, Link, Menu, MenuItem, Paper, Snackbar, Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Tooltip, Typography} from '@mui/material';
import {HStack, useMuiPrompts} from '@dsherwin/mui-kit';
import RefreshIcon from '@mui/icons-material/Refresh';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import Warning from '@mui/icons-material/Warning';
import {DataGrid, type GridColDef} from '@mui/x-data-grid';
import {useNavigate, useSearchParams} from 'react-router';
import {formatDate, formatDateTime} from '../../lib/utilities';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext';
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';

const platformDescriptions: Record<string, string> = {
    'darwin-amd64': 'macOS (Intel)',
    'darwin-arm64': 'macOS (Apple Silicon)',
    'linux-amd64': 'Linux (x64)',
    'linux-arm64': 'Linux (ARM64)',
    'linux-386': 'Linux (x86)',
    'linux-armv5': 'Linux (ARMv5)',
    'linux-armv6': 'Linux (ARMv6)',
    'linux-armv7': 'Linux (ARMv7)',
    'windows-amd64': 'Windows (x64)',
    'windows-arm64': 'Windows (ARM64)',
    'windows-386': 'Windows (x86)',
    'freebsd-amd64': 'FreeBSD (x64)',
    'freebsd-arm64': 'FreeBSD (ARM64)',
};

const getPlatformDescription = (platform: string) => platformDescriptions[platform] || platform;

export const AgentsList = () => {
    const [rows, setRows] = useState<any[] | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | undefined>();
    const navigate = useNavigate();
    const {confirmPrompt} = useMuiPrompts();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const agentLimit = checkLimit('agents');
    const [params] = useSearchParams()
    const limit = Math.min(Number(params.get('limit') || 25), 100)
    const offset = Number(params.get('offset') || 0)
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({
        open: false,
        message: '',
        severity: 'info'
    });
    const [updating, setUpdating] = useState<Record<string, boolean>>({});
    const [agentBuilds, setAgentBuilds] = useState<any>(null);

    const load = async () => {
        try {
            setLoading(true);
            setError(undefined);
            const list = await apiGet('/agents') as any[];
            setRows(list || []);
        } catch (e: any) {
            setError(e?.message || 'Failed to load agents');
            setRows([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void load();
        const fetchBuilds = async () => {
            try {
                const response = await fetch('https://dist.chronixhq.com/latest.json');
                const data = await response.json();
                setAgentBuilds(data['chronix-agent']);
            } catch (e) {
                console.error('Failed to fetch agent builds', e);
            }
        };
        void fetchBuilds();
    }, []);

    const [menuEl, setMenuEl] = useState<null | HTMLElement>(null);
    const [menuAgent, setMenuAgent] = useState<any | null>(null);

    const openMenu = (evt: React.MouseEvent<HTMLElement>, agent: any) => {
        evt.stopPropagation();
        setMenuEl(evt.currentTarget);
        setMenuAgent(agent);
    };
    const closeMenu = () => {
        setMenuEl(null);
        setMenuAgent(null);
    };

    const doDelete = async (agent: any) => {
        try {
            const ok = await confirmPrompt({
                title: 'Delete agent',
                message: `Delete agent '${agent.name}'?`,
                buttonText: 'Delete',
                cancelButtonText: 'Cancel'
            });
            if (!ok) return;

            // Preview the impact first to avoid relying on error responses
            const preview: any = await apiDelete(`/agents/${encodeURIComponent(agent.uuid)}?preview=true` as any);
            const inUseCount = Number(preview?.inUseCount || 0);

            let res: any;
            if (inUseCount > 0) {
                const names: string[] = Array.isArray(preview?.connections) ? preview.connections.map((c: any) => c?.name).filter(Boolean) : [];
                const shown = names.slice(0, 5).join(', ');
                const more = names.length > 5 ? `, and ${names.length - 5} more` : '';
                const also = await confirmPrompt({
                    title: 'Agent in use',
                    message: names.length ? `Agent is used by ${inUseCount} connection(s): ${shown}${more}.\nAlso remove these mappings and delete the agent?` : 'Agent is used by one or more connections. Also remove these mappings and delete the agent?',
                    buttonText: 'Delete and remove mappings',
                    cancelButtonText: 'Cancel'
                });
                if (!also) return;
                res = await apiDelete(`/agents/${encodeURIComponent(agent.uuid)}?removeMappings=true` as any);
            } else {
                res = await apiDelete(`/agents/${encodeURIComponent(agent.uuid)}` as any);
            }

            if (res?.ok === false) {
                setError('Delete failed');
            } else {
                void reloadFeatureAvailability();
                await load();
            }
        } catch (e: any) {
            setError(e?.message || 'Delete failed');
        } finally {
            closeMenu();
        }
    };

    const onUpdateAgent = async (agent: any) => {
        const targetVersion = agent.updateAvailable;
        if (!targetVersion) return;

        const ok = await confirmPrompt({
            title: 'Update Agent',
            message: `Are you sure you want to update agent '${agent.name}' to version ${targetVersion}? The agent will restart.`,
            buttonText: 'Update Agent',
            cancelButtonText: 'Cancel'
        });
        if (!ok) return;

        setUpdating(s => ({...s, [agent.uuid]: true}));
        closeMenu();
        try {
            await apiPost(`/agents/${agent.uuid}/update`, {});
            setSnack({open: true, message: `Update initiated for ${agent.name}`, severity: 'success'});

            // Poll for agent to come back online with new version
            let attempts = 0;
            const maxAttempts = 60; // 2 minutes with 2s interval
            const poll = async () => {
                try {
                    const agentRes = await apiGet('/agents') as any[];
                    const updatedAgent = (Array.isArray(agentRes) ? agentRes : []).find(a => a.uuid === agent.uuid);

                    if (updatedAgent && updatedAgent.online && updatedAgent.version === targetVersion) {
                        setSnack({open: true, message: `Agent ${agent.name} updated successfully to ${targetVersion}`, severity: 'success'});
                        setUpdating(s => ({...s, [agent.uuid]: false}));
                        void load();
                    } else {
                        attempts++;
                        if (attempts < maxAttempts) {
                            setTimeout(poll, 2000);
                        } else {
                            setSnack({open: true, message: `Agent ${agent.name} update verification timed out.`, severity: 'error'});
                            setUpdating(s => ({...s, [agent.uuid]: false}));
                            void load();
                        }
                    }
                } catch {
                    attempts++;
                    if (attempts < maxAttempts) {
                        setTimeout(poll, 2000);
                    } else {
                        setUpdating(s => ({...s, [agent.uuid]: false}));
                        void load();
                    }
                }
            };
            setTimeout(poll, 5000);
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: `Failed to update agent: ${e?.message || agent.name}`, severity: 'error'});
            setUpdating(s => ({...s, [agent.uuid]: false}));
        }
    };

    const onRestartAgent = async (agent: any) => {
        if (!agent || !agent.online) return;

        const ok = await confirmPrompt({
            title: 'Restart Agent',
            message: `Are you sure you want to restart agent '${agent.name}'?`,
            buttonText: 'Restart Agent',
            cancelButtonText: 'Cancel'
        });
        if (!ok) return;

        setUpdating(s => ({...s, [agent.uuid]: true}));
        closeMenu();
        try {
            await apiPost(`/agents/${agent.uuid}/restart`, {});
            setSnack({open: true, message: `Restart initiated for ${agent.name}`, severity: 'info'});

            // Poll for agent to come back online
            let attempts = 0;
            const maxAttempts = 60; // 2 minutes with 2s interval
            const poll = async () => {
                try {
                    const agentRes = await apiGet('/agents') as any[];
                    const updatedAgent = (Array.isArray(agentRes) ? agentRes : []).find(a => a.uuid === agent.uuid);

                    if (updatedAgent && updatedAgent.online) {
                        setSnack({open: true, message: `Agent ${agent.name} is back online`, severity: 'success'});
                        setUpdating(s => ({...s, [agent.uuid]: false}));
                        void load();
                    } else {
                        attempts++;
                        if (attempts < maxAttempts) {
                            setTimeout(poll, 2000);
                        } else {
                            setSnack({open: true, message: `Agent ${agent.name} restart verification timed out.`, severity: 'error'});
                            setUpdating(s => ({...s, [agent.uuid]: false}));
                            void load();
                        }
                    }
                } catch {
                    attempts++;
                    if (attempts < maxAttempts) {
                        setTimeout(poll, 2000);
                    } else {
                        setUpdating(s => ({...s, [agent.uuid]: false}));
                        void load();
                    }
                }
            };
            setTimeout(poll, 5000);
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: `Failed to restart agent: ${e?.message || agent.name}`, severity: 'error'});
            setUpdating(s => ({...s, [agent.uuid]: false}));
        }
    };

    const columns: GridColDef[] = useMemo(() => ([
        {
            field: 'name', headerName: 'Name', flex: 1, minWidth: 180, sortable: true,
            renderCell: (params) => (
                <Link component="button" underline="hover" onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    navigate(`/agents/${encodeURIComponent(params.row.uuid)}`)
                }}>{params.row.name}</Link>
            )
        },
        {
            field: 'version', headerName: 'Version', minWidth: 120, flex: 0.5, sortable: true,
            renderCell: (p) => (
                <HStack spacing={1} alignItems="center" sx={{height: '100%'}}>
                    <Typography variant="body2">{p.row.version || 'Unknown'}</Typography>
                    {p.row.updateAvailable && (
                        <Tooltip title={`Update available: ${p.row.updateAvailable}`}>
                            <Chip size="small" color="warning" label="Update" sx={{height: 20, fontSize: '0.65rem', fontWeight: 'bold'}}/>
                        </Tooltip>
                    )}
                </HStack>
            )
        },
        {
            field: 'platform', headerName: 'Platform', minWidth: 180, flex: 1, sortable: true,
            valueGetter: (_v, row) => {
                const parts = [];
                if (row.osType) {
                    parts.push(row.osType);
                } else if (row.os) {
                    parts.push(row.os);
                }
                if (row.osVersion) {
                    parts.push(row.osVersion);
                }
                if (row.arch) {
                    parts.push(`(${row.arch})`);
                }
                return parts.join(' ') || 'Unknown';
            }
        },
        {
            field: 'runningUser', headerName: 'Running As', minWidth: 140, flex: 1, sortable: true,
            valueGetter: (_v, row) => row.runningUser || ''
        },
        {
            field: 'status', headerName: 'Status', minWidth: 100, flex: 0.4, sortable: true,
            renderCell: (p) => {
                if (p.row.suspended) {
                    return (
                        <Tooltip title="Suspended. This agent is temporarily disconnected and inactive.">
                            <Chip size="small" color="warning" icon={<Warning fontSize="small"/>} label="Suspended"/>
                        </Tooltip>
                    );
                }
                return <Chip size="small" variant="outlined" label={p.row.status} sx={{textTransform: 'capitalize'}}/>;
            }
        },
        {
            field: 'online', headerName: 'Online', minWidth: 80, flex: 0.4, sortable: true,
            renderCell: (p) => p.row.online ? <Chip size="small" color="success" label="Online"/> : <Chip size="small" label="Offline"/>
        },
        {
            field: 'lastSeenAt', headerName: 'Last Seen At', minWidth: 180, flex: 0.8, sortable: true,
            valueGetter: (_v, row) => row.lastSeenAt ? formatDateTime(row.lastSeenAt) : ''
        },
        {
            field: 'lastSeenIp', headerName: 'Last Seen IP', minWidth: 140, flex: 0.6, sortable: true,
            valueGetter: (_v, row) => row.lastSeenIp || ''
        },
        {
            field: 'actions', headerName: 'Actions', minWidth: 90, flex: 0.4, sortable: false, filterable: false,
            renderCell: (params) => (
                <IconButton size="small" onClick={(e) => openMenu(e, params.row)}>
                    <MoreVertIcon/>
                </IconButton>
            )
        },
    ]), [navigate]);

    return (
        <Box sx={{
            p: 2
        }}>
            <Box sx={{width: '100%', maxWidth: 1280, mx: 'auto'}}>
                <Stack spacing={2}>
                    <Box
                        sx={{
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "space-between"
                        }}>
                        <Box
                            sx={{
                                display: "flex",
                                alignItems: "center"
                            }}>
                            <Typography variant="h5">Agents</Typography>
                            <SectionHelp section={HELP_SECTIONS.agents}/>
                        </Box>
                        <Box>
                            {loading && <CircularProgress size={20} sx={{mr: 1}}/>}
                            <Tooltip title="Refresh">
                                <IconButton onClick={() => void load()}><RefreshIcon/></IconButton>
                            </Tooltip>
                        </Box>
                    </Box>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    {error && (<Typography color="error" variant="body2">{error}</Typography>)}

                    {!agentLimit.allowed && (
                        <Alert severity="warning">
                            {agentLimit.message}
                        </Alert>
                    )}

                    <Paper variant="outlined">
                        <DataGrid
                            density="compact"
                            rows={rows || []}
                            columns={columns}
                            getRowId={(row) => row.uuid}
                            loading={loading}
                            disableRowSelectionOnClick
                            disableColumnMenu
                            sortingMode="client"
                            onRowClick={(params) => navigate(`/agents/${encodeURIComponent(params.row.uuid)}`)}
                            paginationModel={{
                                pageSize: limit,
                                page: Math.floor(offset / Math.max(limit, 1))
                            }}
                            pageSizeOptions={[5, 10, 25, 50, 100]}
                            pagination
                        />
                    </Paper>
                    <Box
                        sx={{
                            display: "flex",
                            justifyContent: "flex-end",
                            alignItems: "center",
                            gap: 2
                        }}>
                        <Typography variant="body2">{(rows || []).length} total</Typography>
                    </Box>

                    <Box
                        sx={{
                            mt: 4,
                            borderTop: (theme) => `1px solid ${theme.palette.divider}`,
                            pt: 4
                        }}>
                        <Typography variant="h6" gutterBottom>What is an Agent?</Typography>
                        <Typography
                            variant="body2"
                            component="p"
                            sx={{
                                color: "text.secondary",
                                mb: 2
                            }}>
                            An Agent is a lightweight service that runs on your infrastructure (servers, VMs, or containers) to facilitate communication between Chronix and your network devices. Agents securely tunnel traffic, perform local health checks, and execute commands, allowing you to manage devices even in
                            isolated or private networks without complex VPNs or firewall rules.
                        </Typography>

                        {agentBuilds && (
                            <Box sx={{
                                mt: 2
                            }}>
                                <Typography variant="subtitle2" gutterBottom sx={{fontWeight: 'bold'}}>
                                    Latest Agent Builds ({agentBuilds.version}) — Released {formatDate(agentBuilds.release_date)}
                                </Typography>
                                {agentBuilds.release_notes && (
                                    <Typography
                                        variant="caption"
                                        sx={{
                                            color: "text.secondary",
                                            display: 'block',
                                            mb: 1,
                                            fontStyle: 'italic'
                                        }}>
                                        {agentBuilds.release_notes}
                                    </Typography>
                                )}
                                <TableContainer component={Paper} variant="outlined" sx={{maxWidth: 800}}>
                                    <Table size="small">
                                        <TableHead>
                                            <TableRow sx={{backgroundColor: (theme) => theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.04)' : 'rgba(255, 255, 255, 0.04)'}}>
                                                <TableCell sx={{fontWeight: 'bold'}}>Platform</TableCell>
                                                <TableCell sx={{fontWeight: 'bold'}}>Download Link</TableCell>
                                            </TableRow>
                                        </TableHead>
                                        <TableBody>
                                            {Object.entries(agentBuilds.binaries || {}).map(([platform, info]: [string, any]) => {
                                                const fileName = info.url.split('/').pop();
                                                return (
                                                    <TableRow key={platform} hover>
                                                        <TableCell sx={{py: 1}}>{getPlatformDescription(platform)}</TableCell>
                                                        <TableCell sx={{py: 1}}>
                                                            <Link
                                                                href={info.url}
                                                                target="_blank"
                                                                rel="noopener noreferrer"
                                                                variant="body2"
                                                                underline="hover"
                                                                sx={{fontFamily: 'monospace'}}
                                                            >
                                                                {fileName}
                                                            </Link>
                                                        </TableCell>
                                                    </TableRow>
                                                );
                                            })}
                                        </TableBody>
                                    </Table>
                                </TableContainer>
                            </Box>
                        )}
                    </Box>
                </Stack>
            </Box>
            <Menu anchorEl={menuEl} open={!!menuEl} onClose={closeMenu} anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}>
                {menuAgent?.online && (
                    <MenuItem
                        onClick={() => onRestartAgent(menuAgent)}
                        disabled={updating[menuAgent?.uuid]}
                    >
                        {updating[menuAgent?.uuid] ? 'Busy...' : 'Restart Agent'}
                    </MenuItem>
                )}
                {menuAgent?.updateAvailable && menuAgent?.online && (
                    <MenuItem
                        onClick={() => onUpdateAgent(menuAgent)}
                        sx={{color: 'warning.main', fontWeight: 'bold'}}
                        disabled={updating[menuAgent?.uuid]}
                    >
                        {updating[menuAgent?.uuid] ? 'Updating...' : `Update Agent to ${menuAgent.updateAvailable}`}
                    </MenuItem>
                )}
                <MenuItem onClick={() => menuAgent && doDelete(menuAgent)}>Delete Agent…</MenuItem>
            </Menu>
            <Snackbar
                open={snack.open}
                autoHideDuration={6000}
                onClose={() => setSnack(s => ({...s, open: false}))}
                message={snack.message}
            />
        </Box>
    );
};
