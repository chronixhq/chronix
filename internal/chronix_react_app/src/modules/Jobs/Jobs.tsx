import {useEffect, useMemo, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Chip, CircularProgress, Collapse, Divider, FormControl, IconButton, InputLabel, MenuItem, Select, Snackbar, Switch, TextField, Tooltip, Typography} from '@mui/material';
import {Add, Delete, Edit, ExpandMore, PlayArrow, Refresh, Warning} from '@mui/icons-material';
import {useLocation, useNavigate} from 'react-router';
import {type Job} from './types';
// Data providers
import {useJobs} from '../../data/JobsContext';
import {useActions} from '../../data/ActionsContext';
import {useConnections} from '../../data/ConnectionsContext';
import {toastAPIError} from '../../lib/errors';
import {useRunNow} from '../../data/RunNowContext';
import {JobStatusChip, scheduleSummaryForJob} from "../../lib/utilities.tsx";
import {formatDateTimeHM} from '../../lib/utilities';
import {useRunsContext} from '../../data/RunsContext';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext';
import {HStack, useMessagesContext, useMuiPrompts, VStack} from "@dsherwin/mui-kit";
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';

function RunsPreview({jobId}: { jobId: string | number }) {
    const {useRecentRunsForJob} = useRunsContext();
    const {items, loading, error, reload} = useRecentRunsForJob(jobId, 10);
    return (
        <VStack spacing={0.5} sx={{mt: 1}}>
            <HStack alignItems="center" justifyContent="space-between">
                <Typography variant="subtitle2">Recent runs</Typography>
                {loading && <CircularProgress size={14}/>} {error && <Button size="small" onClick={reload}>Retry</Button>}
            </HStack>
            {(!items || items.length === 0) && !loading ? (
                <Typography variant="body2" sx={{
                    color: "text.secondary"
                }}>No recent runs</Typography>
            ) : (
                <VStack spacing={0.25}>
                    {items.map((r) => {
                        const when = r.finishedAt || r.startedAt || r.queuedAt;
                        const whenStr = when ? formatDateTimeHM(when as any) : '';
                        return (
                            <HStack key={r.runId} sx={{gap: 1}}>
                                <Chip size="small" label={r.status} color={r.status === 'success' ? 'success' : r.status === 'error' ? 'error' : r.status === 'running' ? 'info' : 'default'} variant="outlined"/>
                                <Typography variant="body2">{whenStr}</Typography>
                                {r.message && <Typography
                                    variant="body2"
                                    noWrap
                                    sx={{
                                        color: "text.secondary",
                                        flex: 1,
                                        minWidth: 80
                                    }}>— {r.message}</Typography>}
                            </HStack>
                        );
                    })}
                </VStack>
            )}
        </VStack>
    );
}

/**
 * Jobs component displays a list of scheduled jobs with filtering, sorting, and pagination.
 * It allows users to manage job lifecycle (enable/disable, delete, run now) and view recent run history.
 */
export const Jobs = () => {
    const navigate = useNavigate();
    const {items, loading, refreshing, hasLoaded, error: jobsError, reload, ensureLoaded: ensureJobsLoaded, setEnabled, deleteJob} = useJobs();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const jobLimit = checkLimit('jobs');
    const {byId: actionsById, reload: reloadActions, ensureLoaded: ensureActionsLoaded} = useActions();
    const {byId: connsById, reload: reloadConnections, ensureLoaded: ensureConnectionsLoaded} = useConnections();
    const [search, setSearch] = useState<string>('');
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' | 'warning' }>({open: false, message: '', severity: 'info'});
    const [expanded, setExpanded] = useState<Record<string, boolean>>({});
    const {runNow} = useRunNow();
    const loadError = jobsError;
    const {confirmPrompt} = useMuiPrompts();
    const {displayErrorMessage} = useMessagesContext();

    const toggleExpand = (id: string) => setExpanded(prev => ({...prev, [id]: !prev[id]}));
    const isExpanded = (id: string) => !!expanded[id];

    const load = async () => {
        try {
            // refresh supporting lookups and jobs list (keep list visible; show header spinner)
            await Promise.all([
                reloadConnections(),
                reloadActions(),
                reload({silent: true}),
            ]);
        } catch (e) {
            console.error(e);
        }
    };

    useEffect(() => {
        void ensureJobsLoaded();
        void ensureConnectionsLoaded();
        void ensureActionsLoaded();
    }, [ensureActionsLoaded, ensureConnectionsLoaded, ensureJobsLoaded]);

    // If navigated back after an edit, trigger a one-time reload
    const location = useLocation();
    useEffect(() => {
        const st: any = location.state as any;
        if (st && st.refresh) {
            void load();
            (async () => {
                await Promise.resolve();
                navigate(location.pathname, {replace: true, state: {}});
            })();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [location.state]);

    const [sortBy, setSortBy] = useState<'name_asc' | 'last_desc' | 'enabled_first'>('name_asc');
    const [page, setPage] = useState<number>(1);
    const [pageSize, setPageSize] = useState<number>(10);

    const filtered = useMemo(() => {
        const q = search.trim().toLowerCase();
        if (!q) return items;
        return items.filter(j => {
            const jAny: any = j as any;
            const aid = String(jAny.actionId ?? jAny.action_id ?? '');
            const kind = jAny.targetKind || jAny.target_kind || 'database';
            const aName = (j.action?.name || (aid ? actionsById[`${kind}-${aid}`]?.name : '') || '').toLowerCase();
            return (
                j.name.toLowerCase().includes(q) ||
                (j.description || '').toLowerCase().includes(q) ||
                aName.includes(q)
            );
        });
    }, [items, search, actionsById]);

    const sorted = useMemo(() => {
        const arr = filtered.slice();
        switch (sortBy) {
            case 'last_desc':
                arr.sort((a, b) => new Date((b as any).lastRunAt || 0).getTime() - new Date((a as any).lastRunAt || 0).getTime());
                break;
            case 'enabled_first':
                arr.sort((a, b) => (b.enabled === false ? 0 : 1) - (a.enabled === false ? 0 : 1) || a.name.localeCompare(b.name));
                break;
            default:
                arr.sort((a, b) => a.name.localeCompare(b.name));
        }
        return arr;
    }, [filtered, sortBy]);

    const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
    const pageItems = useMemo(() => sorted.slice((page - 1) * pageSize, (page - 1) * pageSize + pageSize), [sorted, page, pageSize]);

    const handleDeleteJob = (id: string, name: string) => {
        if (!id || id.length === 0) {
            displayErrorMessage('Invalid job ID');
            return;
        }
        (async () => {
            let ok = false;
            try {
                ok = await confirmPrompt({
                    message: `Are you sure you want to delete "${name}?"`,
                    title: `Delete Job ${name}`,
                    buttonText: "Delete",
                });
            } catch {
            }
            if (!ok) return;
            try {
                ok = await deleteJob(id);
                if (!ok) throw new Error('Delete failed');
                setSnack({open: true, message: 'Job deleted', severity: 'success'});
                void reloadFeatureAvailability();
                await reload({silent: true});
            } catch (e) {
                console.error(e);
                setSnack({open: true, message: toastAPIError(e, 'Delete failed'), severity: 'error'});
            }
        })();
    }

    const toggleEnabled = async (j: Job) => {
        if (j.enabled === false && !jobLimit.allowed) {
            setSnack({open: true, message: jobLimit.message || 'Job limit reached', severity: 'warning'});
            return;
        }
        try {
            const ok = await setEnabled(j.id, j.enabled === false);
            if (!ok) throw new Error('Toggle failed');
            setSnack({open: true, message: j.enabled === false ? 'Job enabled' : 'Job disabled', severity: 'success'});
            void reloadFeatureAvailability();
            await reload();
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Toggle failed', severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5">Jobs</Typography>
                        <SectionHelp section={HELP_SECTIONS.jobs} />
                    </Box>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        {(loading || refreshing) && <CircularProgress size={20} sx={{mr: 1}}/>}
                        <Tooltip title="Refresh"><IconButton onClick={load}><Refresh/></IconButton></Tooltip>
                        <Button startIcon={<Add/>} onClick={() => navigate('/jobs/create')} disabled={!jobLimit.allowed}>New Job</Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {!jobLimit.allowed && (
                    <Alert severity="warning">
                        {jobLimit.message}
                    </Alert>
                )}

                <HStack spacing={1} sx={{flexWrap: 'wrap'}}>
                    <TextField size="small" placeholder="Search by name, description, or action" value={search} onChange={(e) => {
                        setSearch(e.target.value);
                        setPage(1);
                    }} sx={{minWidth: {xs: '100%', sm: 320}}}/>
                    <FormControl size="small" sx={{minWidth: 180}}>
                        <InputLabel id="sortby-label">Sort by</InputLabel>
                        <Select labelId="sortby-label" label="Sort by" value={sortBy} onChange={(e) => {
                            setSortBy(e.target.value as any);
                            setPage(1);
                        }}>
                            <MenuItem value="name_asc">Name (A→Z)</MenuItem>
                            <MenuItem value="last_desc">Last run</MenuItem>
                            <MenuItem value="enabled_first">Enabled first</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 120}}>
                        <InputLabel id="psize-label">Page size</InputLabel>
                        <Select labelId="psize-label" label="Page size" value={String(pageSize)} onChange={(e) => {
                            setPageSize(Number(e.target.value));
                            setPage(1);
                        }}>
                            <MenuItem value={10}>10</MenuItem>
                            <MenuItem value={20}>20</MenuItem>
                            <MenuItem value={50}>50</MenuItem>
                        </Select>
                    </FormControl>
                </HStack>

                {loadError && (
                    <Alert severity="error" action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>
                        {loadError}
                    </Alert>
                )}

                {(loading && !hasLoaded) ? (
                    <VStack spacing={2}>
                        {[...Array(3)].map((_, i) => (
                            <Card key={i} variant="outlined" sx={{borderRadius: 3, p: 2}}>
                                <Typography variant="body2" sx={{
                                    color: "text.secondary"
                                }}>Loading…</Typography>
                            </Card>
                        ))}
                    </VStack>
                ) : filtered.length === 0 ? (
                    <Card variant="outlined" sx={{borderRadius: 3, p: 3, textAlign: 'center'}}>
                        <Typography variant="h6">No jobs found</Typography>
                        <Typography
                            sx={{
                                color: "text.secondary",
                                mt: 1
                            }}>Try adjusting your search or create a new job.</Typography>
                        <Button sx={{mt: 2}} startIcon={<Add/>} onClick={() => navigate('/jobs/create')} disabled={!jobLimit.allowed}>New Job</Button>
                    </Card>
                ) : (
                    <VStack spacing={2}>
                        {pageItems.map(j => (
                            <Card key={j.id} variant="outlined" sx={{borderRadius: 3}}>
                                <CardContent sx={{pb: 1}}>
                                    <HStack alignItems="center" justifyContent="space-between" sx={{gap: 2, flexWrap: 'wrap'}}>
                                        <HStack alignItems="center" sx={{gap: 1.5, minWidth: 240, flex: 1}}>
                                            <Typography
                                                variant="subtitle1"
                                                sx={{
                                                    fontWeight: 600,
                                                    opacity: j.suspended ? 0.5 : 1
                                                }}>{j.name}</Typography>
                                            <JobStatusChip jobId={j.id} fallbackStatus={j.lastRunStatus}/>
                                            {j.enabled === false && !j.suspended && (
                                                <Chip size="small" variant="outlined" color="warning" label="Disabled"/>
                                            )}
                                            {j.suspended && (
                                                <Tooltip title="Suspended. This job is temporarily inactive.">
                                                    <Chip size="small" color="warning" icon={<Warning fontSize="small"/>} label="Suspended"/>
                                                </Tooltip>
                                            )}
                                            <VStack>
                                                {j.lastRunAt && <Typography variant="caption" sx={{
                                                    color: "text.secondary"
                                                }}>Last run: {formatDateTimeHM(j.lastRunAt)}</Typography>}
                                                {j.nextRunAt && <Typography variant="caption" sx={{
                                                    color: "text.secondary"
                                                }}>Next run: {formatDateTimeHM(j.nextRunAt)}</Typography>}
                                            </VStack>
                                        </HStack>
                                        <HStack alignItems="center" sx={{gap: 0.5}}>
                                            <Tooltip title={j.suspended ? "Suspended" : (j.enabled !== false ? "Enabled" : "Disabled")}>
                                                <span>
                                                    <Switch size="small" checked={j.enabled !== false && !j.suspended} onChange={() => toggleEnabled(j)} disabled={j.suspended}/>
                                                </span>
                                            </Tooltip>
                                            <Tooltip title="Run now">
                                                <IconButton size="small" onClick={async () => {
                                                    try {
                                                        const rid = await runNow(j.id, {jobName: j.name});
                                                        setSnack({open: true, message: `Run queued${rid ? ` (id: ${rid})` : ''}`, severity: 'info'});
                                                        await load();
                                                    } catch (e) {
                                                        console.error(e);
                                                        setSnack({open: true, message: 'Run failed', severity: 'error'});
                                                    }
                                                }}>
                                                    <PlayArrow/>
                                                </IconButton>
                                            </Tooltip>
                                            <Tooltip title={j.suspended ? "Cannot edit suspended job" : "Edit"}>
                                                <span>
                                                    <IconButton size="small" onClick={() => navigate(`/jobs/edit/${encodeURIComponent(j.id)}`)} disabled={j.suspended}><Edit/></IconButton>
                                                </span>
                                            </Tooltip>
                                            <Tooltip title="Delete"><IconButton size="small" onClick={() => handleDeleteJob(j.id, j.name)}><Delete/></IconButton></Tooltip>
                                        </HStack>
                                    </HStack>
                                    {j.description && (
                                        <Typography variant="body2" sx={{mt: 1}}>{j.description}</Typography>
                                    )}
                                    <Alert severity="info" variant="standard" sx={{py: 0, mt: 2, mb: 1}}>{scheduleSummaryForJob(j)}</Alert>
                                </CardContent>
                                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                <CardActions sx={{justifyContent: 'space-between'}}>
                                    <Typography
                                        variant="caption"
                                        sx={{
                                            color: "text.secondary",
                                            ml: 1
                                        }}>ID: {j.id}</Typography>
                                    <IconButton size="small" onClick={() => toggleExpand(j.id)}>
                                        <ExpandMore sx={{transform: isExpanded(j.id) ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s'}}/>
                                    </IconButton>
                                </CardActions>
                                <Collapse in={isExpanded(j.id)} timeout="auto" unmountOnExit>
                                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                    <CardContent>
                                        <Typography variant="subtitle2" gutterBottom>Details</Typography>
                                        <VStack spacing={0.5}>
                                            {(() => {
                                                // Always show Last and Next rows with proper messaging
                                                const jAny: any = j as any;
                                                const last = j.lastRunAt ? formatDateTimeHM(j.lastRunAt) : 'not ran yet';
                                                let next: string = '';
                                                const sched: any = jAny.schedule;
                                                if (j.enabled === false) {
                                                    next = 'job disabled';
                                                } else if (sched && sched.kind === 'single') {
                                                    next = 'single shot';
                                                } else if (sched && sched.kind === 'manual') {
                                                    next = 'manual only';
                                                } else if (sched && sched.kind === 'recurring' && sched.endAt) {
                                                    try {
                                                        const end = new Date(sched.endAt);
                                                        if (!isNaN(end.getTime()) && end.getTime() < Date.now()) {
                                                            next = 'past end date';
                                                        }
                                                    } catch {
                                                        // ignore parse errors
                                                    }
                                                }
                                                if (!next) {
                                                    const nra = j.nextRunAt as any;
                                                    next = nra ? formatDateTimeHM(nra) : 'not scheduled';
                                                }
                                                return (
                                                    <>
                                                        <Typography variant="body2">Last run: {last}</Typography>
                                                        <Typography variant="body2">Next run: {next}</Typography>
                                                    </>
                                                );
                                            })()}
                                            {(() => {
                                                const jAny: any = j as any;
                                                const kind = jAny.targetKind || jAny.target_kind || 'database';
                                                const cid = String(jAny.connectionId ?? jAny.connection_id ?? jAny.shellConnectionId ?? jAny.shell_connection_id ?? jAny.webtaskConnectionId ?? jAny.webtask_connection_id ?? '');
                                                const conn = (j as any).connection || (cid ? connsById[`${kind}-${cid}`] : undefined);
                                                return conn ? (<Typography variant="body2">Connection: {conn.name}{conn.driver ? ` (${conn.driver})` : ''}</Typography>) : null;
                                            })()}
                                            {(() => {
                                                const jAny: any = j as any;
                                                const aid = String(jAny.actionId ?? jAny.action_id ?? '');
                                                const kind = jAny.targetKind || jAny.target_kind || 'database';
                                                const act = (j as any).action || (aid ? actionsById[`${kind}-${aid}`] : undefined);
                                                return act ? (<Typography variant="body2">Action: {act.name}</Typography>) : null;
                                            })()}
                                            {j.variables && j.variables.length > 0 && (
                                                <>
                                                    <Typography variant="body2">Variables:</Typography>
                                                    {j.variables.map(v => (
                                                        <Typography key={v.name} variant="body2" sx={{pl: 2}}>{v.name}: {v.value}</Typography>
                                                    ))}
                                                </>
                                            )}
                                            {(() => {
                                                const jAny: any = j as any;
                                                const inc = jAny.notify_include_output ?? jAny.notifyIncludeOutput;
                                                if (inc) {
                                                    return <Typography variant="body2">Reporting: Include task output in alerts</Typography>;
                                                }
                                                return null;
                                            })()}
                                            {j.notes && <Typography variant="body2">Notes: {j.notes}</Typography>}
                                        </VStack>
                                        <RunsPreview jobId={j.id}/>
                                    </CardContent>
                                </Collapse>
                            </Card>
                        ))}
                    </VStack>
                )}

                <Typography
                    variant="caption"
                    sx={{
                        color: "text.secondary",
                        mt: 1
                    }}>{sorted.length} job{sorted.length !== 1 ? 's' : ''} — page {page} of {totalPages}</Typography>
                <HStack spacing={1} sx={{justifyContent: 'flex-end'}}>
                    <Button size="small" disabled={page <= 1} onClick={() => setPage(p => Math.max(1, p - 1))}>Prev</Button>
                    <Typography variant="caption" sx={{alignSelf: 'center'}}>Page {page} / {totalPages}</Typography>
                    <Button size="small" disabled={page >= totalPages} onClick={() => setPage(p => Math.min(totalPages, p + 1))}>Next</Button>
                </HStack>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack(s => ({...s, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack(s => ({...s, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
