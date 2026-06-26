import type React from 'react';
import {useCallback, useEffect, useRef, useState} from 'react';
import {Alert, Box, Button, Card, CardContent, Chip, Divider, IconButton, Snackbar, Tooltip, Typography} from '@mui/material';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {useNavigate} from 'react-router';
import {Add, ArrowForward, Autorenew, CheckCircleOutlined, Close, Dataset, DirectionsRun, ErrorOutlined, Groups, HelpOutlined, Refresh, Schedule, Work} from '@mui/icons-material';
import {formatDateTime} from '../lib/utilities';
import {apiGet} from '@dsherwin/react-api-interface';
import {useSseContext} from '../data/SseContext';
import {useUpdateAvailability} from '../lib/useUpdateAvailability';
import type {JobFinishedPayload, JobProgressPayload} from '../modules/Runs/types.ts';

const fmt = (iso: string) => formatDateTime(iso);

type DashboardStats = {
    jobs?: number
    running?: number
    actions?: number
    connectionsTotal?: number
    connectionsOk?: number
    connectionsError?: number
    connectionsUnknown?: number
    agentsKnown?: number
    agentsOnline?: number
    agentsOffline?: number
    agentsPending?: number
}

type DashboardSummaryResponse = {
    stats?: DashboardStats
    upcoming?: Array<{ id?: string | number; name?: string; when?: string | number | Date; status?: string }>
    activity?: Array<{ id?: string | number; when?: string | number | Date; text?: string }>
}


// Types for dashboard lists (populated when backend endpoints are available)
interface UpcomingJob {
    id: string;
    name: string;
    when: string; // ISO
    status: 'scheduled' | 'running' | 'error' | 'success';
}

interface ActivityItem {
    id: string;
    when: string; // ISO
    text: string;
}

export const Dashboard = () => {
    const navigate = useNavigate();
    const {addSSEListener} = useSseContext();

    // KPI stats (mocked in DEV, empty defaults in PROD)
    const [stats, setStats] = useState<{ jobs: number; running: number; actions: number; connectionsTotal: number; connectionsOk: number; connectionsError: number; connectionsUnknown: number; agentsKnown?: number; agentsOnline?: number; agentsOffline?: number; agentsPending?: number }>({
        jobs: 0,
        running: 0,
        actions: 0,
        connectionsTotal: 0,
        connectionsOk: 0,
        connectionsError: 0,
        connectionsUnknown: 0
    });
    const [upcoming, setUpcoming] = useState<UpcomingJob[]>([]);
    const [activity, setActivity] = useState<ActivityItem[]>([]);
    const [loading, setLoading] = useState<boolean>(false);
    const [initialized, setInitialized] = useState<boolean>(false);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    const refreshTimer = useRef<ReturnType<typeof window.setTimeout> | null>(null);
    const {status: updaterStatus, hasServerUpdateNotice, hasAgentUpdateNotice, shouldShowUpdateNotice, dismissUpdateNotice} = useUpdateAvailability();

    const load = useCallback(async () => {
        if (!initialized) setLoading(true);
        setLoadError(null);
        try {
            const data = await apiGet('/dashboard/summary') as DashboardSummaryResponse;
            const s = data?.stats || {};
            setStats({
                jobs: Number(s.jobs ?? 0),
                running: Number(s.running ?? 0),
                actions: Number(s.actions ?? 0),
                connectionsTotal: s.connectionsTotal ?? 0,
                connectionsOk: Number(s.connectionsOk ?? 0),
                connectionsError: Number(s.connectionsError ?? 0),
                connectionsUnknown: Number(s.connectionsUnknown ?? 0),
                agentsKnown: Number(s.agentsKnown ?? 0),
                agentsOnline: Number(s.agentsOnline ?? 0),
                agentsOffline: Number(s.agentsOffline ?? 0),
                agentsPending: Number(s.agentsPending ?? 0),
            });
            const up = Array.isArray(data?.upcoming) ? data.upcoming : [];
            setUpcoming(up.map((j) => ({
                id: String(j.id ?? ''),
                name: String(j.name ?? ''),
                when: typeof j.when === 'string' ? j.when : (j.when ? new Date(j.when).toISOString() : ''),
                status: (j.status === 'running' || j.status === 'error' || j.status === 'success') ? j.status : 'scheduled',
            })));
            const act = Array.isArray(data?.activity) ? data.activity : [];
            setActivity(act.map((a) => ({
                id: String(a.id ?? ''),
                when: typeof a.when === 'string' ? a.when : (a.when ? new Date(a.when).toISOString() : ''),
                text: String(a.text ?? ''),
            })));
            setInitialized(true);
        } catch (e) {
            console.error(e);
            setLoadError('Failed to load dashboard data');
        } finally {
            setLoading(false);
        }
    }, [initialized]);

    useEffect(() => {
        load();
        const interval = setInterval(() => {
            load();
        }, 60000); // periodic refresh every 60s
        return () => clearInterval(interval);
    }, [load]);

    // Debounced refresh on relevant SSE events to keep dashboard live
    useEffect(() => {
        const triggerReload = () => {
            if (refreshTimer.current) window.clearTimeout(refreshTimer.current);
            refreshTimer.current = window.setTimeout(() => {
                load();
            }, 500);
        };
        const unsub1 = addSSEListener<JobProgressPayload>('job_progress', triggerReload);
        const unsub2 = addSSEListener<JobFinishedPayload>('job_finished', triggerReload);
        const unsub3 = addSSEListener<Record<string, unknown>>('connection_health', triggerReload);
        const unsub4 = addSSEListener<Record<string, unknown>>('notification', triggerReload);
        const unsub5 = addSSEListener<Record<string, unknown>>('agent_registration', triggerReload);
        const unsub6 = addSSEListener<Record<string, unknown>>('agent_registration_approved', triggerReload);
        const unsub7 = addSSEListener<Record<string, unknown>>('agent_registration_denied', triggerReload);
        const unsub8 = addSSEListener<Record<string, unknown>>('agent_deleted', triggerReload);
        return () => {
            unsub1?.();
            unsub2?.();
            unsub3?.();
            unsub4?.();
            unsub5?.();
            unsub6?.();
            unsub7?.();
            unsub8?.();
            if (refreshTimer.current) window.clearTimeout(refreshTimer.current);
        };
    }, [addSSEListener, load]);

    const kpiCard = (icon: React.ReactNode, label: string, value: string | number, color?: 'default' | 'success' | 'error' | 'primary' | 'warning') => (
        <Card variant="outlined" sx={{borderRadius: 3, flex: 1, minWidth: 180}}>
            <CardContent>
                <HStack alignItems="center" spacing={1} sx={{justifyContent: 'space-between'}}>
                    <HStack alignItems="center" spacing={1.5}>
                        <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', width: 36, height: 36, borderRadius: '50%', bgcolor: 'action.hover'}}>
                            {icon}
                        </Box>
                        <Typography variant="body2" sx={{
                            color: "text.secondary"
                        }}>{label}</Typography>
                    </HStack>
                    {color === 'success' && <Chip size="small" color="success" label="OK"/>}
                    {color === 'error' && <Chip size="small" color="error" label="Issues"/>}
                </HStack>
                <Typography variant="h5" sx={{mt: 1}}>{value}</Typography>
            </CardContent>
        </Card>
    );


    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1100, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Typography variant="h5">Dashboard</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Tooltip title="Refresh"><IconButton onClick={load}><Refresh/></IconButton></Tooltip>
                        <Button startIcon={<Add/>} onClick={() => navigate('/jobs/create')}>New Job</Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {loadError && (
                    <Alert severity="error" action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>{loadError}</Alert>
                )}

                {(hasServerUpdateNotice || hasAgentUpdateNotice) && shouldShowUpdateNotice && (
                    <Alert
                        severity="warning"
                        action={(
                            <HStack spacing={0.5} alignItems="flex-start">
                                <Button color="inherit" size="small" onClick={() => navigate('/settings/updates')}>Open Updates</Button>
                                <IconButton
                                    aria-label="Dismiss update notice"
                                    color="inherit"
                                    size="small"
                                    onClick={dismissUpdateNotice}
                                    sx={{mt: '-4px'}}
                                >
                                    <Close fontSize="small"/>
                                </IconButton>
                            </HStack>
                        )}
                        sx={{width: '100%'}}
                    >
                        <VStack spacing={0.5} alignItems="flex-start">
                            {hasServerUpdateNotice && (
                                <Typography variant="body2">
                                    A Chronix app update is available{updaterStatus?.availableVersion?.server?.version ? `: ${updaterStatus.availableVersion.server.version}` : ''}.
                                </Typography>
                            )}
                            {hasAgentUpdateNotice && (
                                <Typography variant="body2">
                                    Agent updates are available and automatic agent updating is not enabled.
                                </Typography>
                            )}
                        </VStack>
                    </Alert>
                )}

                {/* KPI Row */}
                <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                    {kpiCard(<Work fontSize="small"/>, 'Jobs', stats.jobs)}
                    {kpiCard(<Autorenew fontSize="small"/>, 'Running', stats.running, stats.running > 0 ? 'primary' : 'default')}
                    {kpiCard(<DirectionsRun fontSize="small"/>, 'Actions', stats.actions)}
                    {kpiCard(<Dataset fontSize="small"/>, 'Connections', stats.connectionsTotal)}
                    {kpiCard(<Groups fontSize="small"/>, 'Agents', `${stats.agentsKnown ?? 0}`)}
                </HStack>

                {/* Middle: Upcoming + Activity */}
                <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                    <Card variant="outlined" sx={{borderRadius: 3, flex: 1, minWidth: 300}}>
                        <CardContent>
                            <HStack alignItems="center" justifyContent="space-between" sx={{mb: 1}}>
                                <HStack spacing={1} alignItems="center"><Schedule fontSize="small"/><Typography variant="subtitle1" sx={{
                                    fontWeight: 600
                                }}>Upcoming jobs</Typography></HStack>
                                <Button size="small" endIcon={<ArrowForward/>} onClick={() => navigate('/jobs/list')}>View all</Button>
                            </HStack>
                            {loading && !initialized ? (
                                <Typography variant="body2" sx={{
                                    color: "text.secondary"
                                }}>Loading…</Typography>
                            ) : upcoming.length === 0 ? (
                                <Typography variant="body2" sx={{
                                    color: "text.secondary"
                                }}>No upcoming jobs scheduled.</Typography>
                            ) : (
                                <VStack spacing={1}>
                                    {upcoming.map(j => (
                                        <HStack key={j.id} sx={{justifyContent: 'space-between'}}>
                                            <Typography variant="body2">{j.name}</Typography>
                                            <Typography variant="caption" sx={{
                                                color: "text.secondary"
                                            }}>{fmt(j.when)}</Typography>
                                        </HStack>
                                    ))}
                                </VStack>
                            )}
                        </CardContent>
                    </Card>

                    <Card variant="outlined" sx={{borderRadius: 3, flex: 1, minWidth: 300}}>
                        <CardContent>
                            <HStack alignItems="center" justifyContent="space-between" sx={{mb: 1}}>
                                <HStack spacing={1} alignItems="center"><DirectionsRun fontSize="small"/><Typography variant="subtitle1" sx={{
                                    fontWeight: 600
                                }}>Recent activity</Typography></HStack>
                                <Button size="small" endIcon={<ArrowForward/>} onClick={() => navigate('/activity')}>View all</Button>
                            </HStack>
                            {loading && !initialized ? (
                                <Typography variant="body2" sx={{
                                    color: "text.secondary"
                                }}>Loading…</Typography>
                            ) : activity.length === 0 ? (
                                <Typography variant="body2" sx={{
                                    color: "text.secondary"
                                }}>No recent activity.</Typography>
                            ) : (
                                <VStack spacing={1}>
                                    {activity.map(a => (
                                        <HStack key={a.id} sx={{justifyContent: 'space-between'}}>
                                            <Typography variant="body2">{a.text}</Typography>
                                            <Typography variant="caption" sx={{
                                                color: "text.secondary"
                                            }}>{fmt(a.when)}</Typography>
                                        </HStack>
                                    ))}
                                </VStack>
                            )}
                        </CardContent>
                    </Card>
                </HStack>

                {/* System status & Quick actions */}
                <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                    <Card variant="outlined" sx={{borderRadius: 3, flex: 1, minWidth: 300}}>
                        <CardContent>
                            <HStack spacing={1} alignItems="center" sx={{mb: 1}}>
                                <Dataset fontSize="small"/>
                                <Typography variant="subtitle1" sx={{
                                    fontWeight: 600
                                }}>System status</Typography>
                            </HStack>
                            {loading && !initialized ? (
                                <Typography variant="body2" sx={{
                                    color: "text.secondary"
                                }}>Loading…</Typography>
                            ) : (
                                <VStack spacing={1}>
                                    <HStack spacing={1} alignItems="center">
                                        <CheckCircleOutlined color="success" fontSize="small"/>
                                        <Typography variant="body2">Connections OK: {stats.connectionsOk}</Typography>
                                    </HStack>
                                    <HStack spacing={1} alignItems="center">
                                        <ErrorOutlined color="error" fontSize="small"/>
                                        <Typography variant="body2">Connections with issues: {stats.connectionsError}</Typography>
                                    </HStack>
                                    <HStack spacing={1} alignItems="center">
                                        <HelpOutlined color="warning" fontSize="small"/>
                                        <Typography variant="body2">Connections unknown: {stats.connectionsUnknown}</Typography>
                                    </HStack>
                                    <Divider sx={{my: 0.5, borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                    <HStack spacing={1} alignItems="center">
                                        <Groups fontSize="small"/>
                                        <Typography variant="body2">Agents known: {stats.agentsKnown ?? 0}</Typography>
                                    </HStack>
                                    <HStack spacing={1} alignItems="center">
                                        <CheckCircleOutlined color="success" fontSize="small"/>
                                        <Typography variant="body2">Agents online: {stats.agentsOnline ?? 0}</Typography>
                                    </HStack>
                                    <HStack spacing={1} alignItems="center">
                                        <ErrorOutlined color="error" fontSize="small"/>
                                        <Typography variant="body2">Agents offline: {stats.agentsOffline ?? 0}</Typography>
                                    </HStack>
                                    {(stats.agentsPending ?? 0) > 0 && (
                                        <HStack spacing={1} alignItems="center">
                                            <Schedule color="warning" fontSize="small"/>
                                            <Typography variant="body2">Agent registrations pending: {stats.agentsPending}</Typography>
                                        </HStack>
                                    )}
                                </VStack>
                            )}
                        </CardContent>
                    </Card>

                    <Card variant="outlined" sx={{borderRadius: 3, flex: 1, minWidth: 300}}>
                        <CardContent>
                            <HStack spacing={1} alignItems="center" sx={{mb: 1}}>
                                <DirectionsRun fontSize="small"/>
                                <Typography variant="subtitle1" sx={{
                                    fontWeight: 600
                                }}>Quick actions</Typography>
                            </HStack>
                            <VStack spacing={1}>
                                <Button startIcon={<Add/>} onClick={() => navigate('/jobs/create')}>New Job</Button>
                                <Button startIcon={<Add/>} onClick={() => navigate('/actions/create')}>New Action</Button>
                                <Button startIcon={<Add/>} onClick={() => navigate('/databases/add')}>New Connection</Button>
                            </VStack>
                        </CardContent>
                    </Card>
                </HStack>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack(s => ({...s, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack(s => ({...s, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>{snack.message}</Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
