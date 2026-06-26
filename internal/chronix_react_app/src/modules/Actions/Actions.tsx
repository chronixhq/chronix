import {useEffect, useMemo, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Chip, Collapse, Divider, FormControl, IconButton, InputLabel, MenuItem, Select, Snackbar, Switch, TextField, Tooltip, Typography} from '@mui/material';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {Add, Delete, Edit, ExpandMore, Http, Refresh, Storage, Terminal, Warning} from '@mui/icons-material';
import {useLocation, useNavigate, useSearchParams} from 'react-router';
import {type Action, type Dialect} from './types';
import {apiDelete, apiPatch} from '@dsherwin/react-api-interface';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext';
import {toastAPIError} from '../../lib/errors';
import {formatDateTime} from '../../lib/utilities';
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';
import {useActions} from '../../data/ActionsContext.tsx';

export const Actions = () => {
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const {confirmPrompt} = useMuiPrompts();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {items, loading, error: loadError, reload, ensureLoaded} = useActions();
    const actLimit = checkLimit('actions');

    const [search, setSearch] = useState<string>('');
    const [filterDialect, setFilterDialect] = useState<'all' | Dialect>('all');
    const filterType = (searchParams.get('type') || 'all') as 'all' | 'database' | 'shell' | 'webtask';
    const setFilterType = (val: string) => {
        const newParams = new URLSearchParams(searchParams);
        if (val === 'all') {
            newParams.delete('type');
        } else {
            newParams.set('type', val);
        }
        setSearchParams(newParams);
        setPage(1);
    };
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' | 'warning' }>({open: false, message: '', severity: 'info'});
    const [expanded, setExpanded] = useState<Record<string, boolean>>({});
    const toggleExpand = (id: string) => setExpanded((prev) => ({...prev, [id]: !prev[id]}));
    const isExpanded = (id: string) => !!expanded[id];

    useEffect(() => {
        void ensureLoaded();
    }, [ensureLoaded]);

    // If we navigated back here after an edit, trigger a refresh once
    const location = useLocation();
    useEffect(() => {
        const st: any = location.state as any;
        if (st && st.refresh) {
            void reload();
            // clear the state to avoid reloading on future navigations
            (async () => {
                await Promise.resolve();
                // replace current history entry without state
                navigate(location.pathname, {replace: true, state: {}});
            })();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [location.state]);

    const [sortBy, setSortBy] = useState<'name_asc' | 'name_desc' | 'updated_desc'>('name_asc');
    const [page, setPage] = useState<number>(1);
    const [pageSize, setPageSize] = useState<number>(10);

    const filtered = useMemo(() => {
        const q = search.trim().toLowerCase();
        return items.filter(a => {
            if (filterType !== 'all' && a.actionType !== filterType) return false;
            if (filterDialect !== 'all' && a.dialect !== filterDialect) return false;
            if (!q) return true;
            return (
                a.name.toLowerCase().includes(q) ||
                (a.description || '').toLowerCase().includes(q) ||
                (a.notes || '').toLowerCase().includes(q)
            );
        });
    }, [items, search, filterDialect, filterType]);

    const sorted = useMemo(() => {
        const arr = filtered.slice();
        switch (sortBy) {
            case 'name_desc':
                arr.sort((a, b) => b.name.localeCompare(a.name));
                break;
            case 'updated_desc':
                arr.sort((a, b) => new Date((b.updatedAt || b.createdAt || '')).getTime() - new Date((a.updatedAt || a.createdAt || '')).getTime());
                break;
            default:
                arr.sort((a, b) => a.name.localeCompare(b.name));
        }
        return arr;
    }, [filtered, sortBy]);

    const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
    const pageItems = useMemo(() => sorted.slice((page - 1) * pageSize, (page - 1) * pageSize + pageSize), [sorted, page, pageSize]);

    const confirmDelete = async (id: string, name: string, type: string) => {
        const ok = await confirmPrompt({
            title: 'Delete Action',
            message: `Are you sure you want to delete the action "${name}"?`,
            buttonText: 'Delete',
            buttonColor: 'error'
        });
        if (!ok) return;

        try {
            let url = `/actions/${id}`;
            if (type === 'shell') url = `/shell/actions/${id}`;
            if (type === 'webtask') url = `/actions/webtask/${id}`;

            const res: any = await apiDelete(url as any);
            if ((res as any)?.ok === false) throw new Error('Delete failed');
            setSnack({open: true, message: 'Action deleted', severity: 'success'});
            void reloadFeatureAvailability();
            await reload();
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: toastAPIError(e, 'Delete failed'), severity: 'error'});
        }
    };

    const toggleEnabled = async (action: Action) => {
        if (!action.enabled && !actLimit.allowed) {
            setSnack({open: true, message: actLimit.message || 'Action limit reached', severity: 'warning'});
            return;
        }
        try {
            let url = `/actions/${action.id}`;
            if (action.actionType === 'shell') url = `/shell/actions/${action.id}`;
            if (action.actionType === 'webtask') url = `/actions/webtask/${action.id}`;

            const res = await apiPatch(url as any, {enabled: !action.enabled} as any);
            if ((res as any)?.ok === false) throw new Error('Toggle failed');
            setSnack({open: true, message: !action.enabled ? 'Action enabled' : 'Action disabled', severity: 'success'});
            void reloadFeatureAvailability();
            await reload();
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: toastAPIError(e, 'Toggle failed'), severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5">Actions</Typography>
                        <SectionHelp section={HELP_SECTIONS.actions} />
                    </Box>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Tooltip title="Refresh"><IconButton onClick={() => void reload()}><Refresh/></IconButton></Tooltip>
                        <Button startIcon={<Add/>} onClick={() => navigate('/actions/create')} disabled={!actLimit.allowed}>New DB Action</Button>
                        <Button startIcon={<Add/>} onClick={() => navigate('/actions/create-shell')} disabled={!actLimit.allowed}>New Shell Action</Button>
                        <Button startIcon={<Add/>} onClick={() => navigate('/actions/create-webtask')} disabled={!actLimit.allowed}>New Web Task Action</Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {!actLimit.allowed && (
                    <Alert severity="warning">
                        {actLimit.message}
                    </Alert>
                )}

                {/* Filters */}
                <HStack spacing={1} sx={{flexWrap: 'wrap'}} alignItems="center">
                    <TextField
                        size="small"
                        placeholder="Search by name, description, or notes"
                        value={search}
                        onChange={(e) => {
                            setSearch(e.target.value);
                            setPage(1);
                        }}
                        sx={{minWidth: {xs: '100%', sm: 320}}}
                    />
                    <FormControl size="small" sx={{minWidth: 160}}>
                        <InputLabel id="action-type-filter-label">Action Type</InputLabel>
                        <Select labelId="action-type-filter-label" label="Action Type" value={filterType} onChange={(e) => {
                            setFilterType(e.target.value as any);
                        }}>
                            <MenuItem value="all">All Types</MenuItem>
                            <MenuItem value="database">Database</MenuItem>
                            <MenuItem value="shell">Shell</MenuItem>
                            <MenuItem value="webtask">Web Task</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 160}}>
                        <InputLabel id="dialect-filter-label">Dialect</InputLabel>
                        <Select labelId="dialect-filter-label" label="Dialect" value={filterDialect} onChange={(e) => {
                            setFilterDialect(e.target.value as 'all' | Dialect);
                            setPage(1);
                        }}>
                            <MenuItem value="all">All</MenuItem>
                            <MenuItem value="postgres">Postgres</MenuItem>
                            <MenuItem value="mysql">MySQL</MenuItem>
                            <MenuItem value="sqlite">SQLite</MenuItem>
                            <MenuItem value="tsql">SQL Server</MenuItem>
                            <MenuItem value="generic">Generic</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 180}}>
                        <InputLabel id="sort-label">Sort by</InputLabel>
                        <Select labelId="sort-label" label="Sort by" value={sortBy} onChange={(e) => {
                            setSortBy(e.target.value as any);
                            setPage(1);
                        }}>
                            <MenuItem value="name_asc">Name (A→Z)</MenuItem>
                            <MenuItem value="name_desc">Name (Z→A)</MenuItem>
                            <MenuItem value="updated_desc">Last updated</MenuItem>
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
                    <Alert severity="error" action={<Button color="inherit" size="small" onClick={() => void reload()}>Retry</Button>}>
                        {loadError}
                    </Alert>
                )}

                {/* Content */}
                {loading ? (
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
                        <Typography variant="h6">No actions found</Typography>
                        <Typography
                            sx={{
                                color: "text.secondary",
                                mt: 1
                            }}>Try adjusting your search or create a new action.</Typography>
                        <HStack spacing={1} sx={{mt: 2, justifyContent: 'center', flexWrap: 'wrap'}}>
                            <Button startIcon={<Add/>} onClick={() => navigate('/actions/create')} disabled={!actLimit.allowed}>New DB Action</Button>
                            <Button startIcon={<Add/>} onClick={() => navigate('/actions/create-shell')} disabled={!actLimit.allowed}>New Shell Action</Button>
                            <Button startIcon={<Add/>} onClick={() => navigate('/actions/create-webtask')} disabled={!actLimit.allowed}>New Web Task Action</Button>
                        </HStack>
                    </Card>
                ) : (
                    <VStack spacing={2}>
                        {pageItems.map((a) => {
                            const borderColor = a.actionType === 'database' ? '#1976d2' : a.actionType === 'shell' ? '#9c27b0' : a.actionType === 'webtask' ? '#ed6c02' : 'divider';
                            const icon = a.actionType === 'database' ? <Storage fontSize="small" sx={{color: '#1976d2'}}/> : a.actionType === 'shell' ? <Terminal fontSize="small" sx={{color: '#9c27b0'}}/> : <Http fontSize="small" sx={{color: '#ed6c02'}}/>;
                            return (
                                <Card key={a.id} variant="outlined" sx={{borderRadius: 3, borderLeft: `6px solid ${borderColor}`}}>
                                    <CardContent sx={{pb: 1}}>
                                        <HStack justifyContent="space-between" alignItems="center" sx={{gap: 2, flexWrap: 'wrap'}}>
                                            <HStack alignItems="center" sx={{gap: 1.5, minWidth: 240, flex: 1}}>
                                                {icon}
                                                <Typography
                                                    variant="subtitle1"
                                                    sx={{
                                                        fontWeight: 600,
                                                        opacity: a.suspended ? 0.5 : 1
                                                    }}>{a.name}</Typography>
                                                <Chip size="small" variant="outlined" label={a.actionType === 'shell' ? 'Shell' : (a.actionType === 'webtask' ? 'Web Task' : (a.dialect || 'Generic'))} sx={{textTransform: 'capitalize'}}/>
                                                <Chip size="small" label={`${a.steps?.length ?? 0} step${(a.steps?.length ?? 0) === 1 ? '' : 's'}`}/>
                                                {a.suspended && (
                                                    <Tooltip title="Suspended. This action is temporarily inactive.">
                                                        <Chip size="small" color="warning" icon={<Warning fontSize="small"/>} label="Suspended"/>
                                                    </Tooltip>
                                                )}
                                                {a.updatedAt && (
                                                    <Typography variant="caption" sx={{
                                                        color: "text.secondary"
                                                    }}>Updated {formatDateTime(a.updatedAt)}</Typography>
                                                )}
                                            </HStack>
                                            <HStack alignItems="center" sx={{gap: 0.5}}>
                                                <Tooltip title={a.suspended ? "Suspended" : (a.enabled ? "Enabled" : "Disabled")}>
                                                    <span>
                                                        <Switch size="small" checked={!!a.enabled && !a.suspended} onChange={() => toggleEnabled(a)} disabled={a.suspended}/>
                                                    </span>
                                                </Tooltip>
                                                <Tooltip title={a.suspended ? "Cannot edit suspended action" : "Edit"}>
                                                    <span>
                                                        <IconButton size="small" onClick={() => {
                                                            let path = `/actions/edit/${encodeURIComponent(a.id)}`;
                                                            if (a.actionType === 'shell') path = `/actions/edit-shell/${encodeURIComponent(a.id)}`;
                                                            if (a.actionType === 'webtask') path = `/actions/edit-webtask/${encodeURIComponent(a.id)}`;
                                                            navigate(path);
                                                        }} disabled={a.suspended}><Edit/></IconButton>
                                                    </span>
                                                </Tooltip>
                                                <Tooltip title="Delete"><IconButton size="small" onClick={() => void confirmDelete(a.id, a.name, a.actionType || 'database')}><Delete/></IconButton></Tooltip>
                                            </HStack>
                                        </HStack>
                                        {a.description && (
                                            <Typography variant="body2" sx={{mt: 1}}>{a.description}</Typography>
                                        )}
                                        {a.notes && (
                                            <Typography
                                                variant="body2"
                                                sx={{
                                                    color: "text.secondary",
                                                    mt: 0.5
                                                }}>{a.notes}</Typography>
                                        )}
                                    </CardContent>
                                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                    <CardActions sx={{justifyContent: 'space-between'}}>
                                        <Typography
                                            variant="caption"
                                            sx={{
                                                color: "text.secondary",
                                                ml: 1
                                            }}>ID: {a.id}</Typography>
                                        <IconButton size="small" onClick={() => toggleExpand(a.id)}>
                                            <ExpandMore sx={{transform: isExpanded(a.id) ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s'}}/>
                                        </IconButton>
                                    </CardActions>
                                    <Collapse in={isExpanded(a.id)} timeout="auto" unmountOnExit>
                                        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                        <CardContent>
                                            <Typography variant="subtitle2" gutterBottom>Steps</Typography>
                                            {(!a.steps || a.steps.length === 0) ? (
                                                <Typography variant="body2" sx={{
                                                    color: "text.secondary"
                                                }}>No steps defined.</Typography>
                                            ) : (
                                                <VStack spacing={0.5}>
                                                    {a.steps
                                                        .slice()
                                                        .sort((s1, s2) => s1.order - s2.order)
                                                        .map((s) => (
                                                            <Typography key={s.id} variant="body2">
                                                                {s.order + 1}. {s.name} — on failure: {s.onFailure || 'exit'}
                                                            </Typography>
                                                        ))}
                                                </VStack>
                                            )}
                                        </CardContent>
                                    </Collapse>
                                </Card>
                            );
                        })}
                    </VStack>
                )}

                <Typography
                    variant="caption"
                    sx={{
                        color: "text.secondary",
                        mt: 1
                    }}>
                    {sorted.length} action{sorted.length !== 1 ? 's' : ''} — page {page} of {totalPages}
                </Typography>
                <HStack spacing={1} sx={{justifyContent: 'flex-end'}}>
                    <Button size="small" disabled={page <= 1} onClick={() => setPage(p => Math.max(1, p - 1))}>Prev</Button>
                    <Typography variant="caption" sx={{alignSelf: 'center'}}>Page {page} / {totalPages}</Typography>
                    <Button size="small" disabled={page >= totalPages} onClick={() => setPage(p => Math.min(totalPages, p + 1))}>Next</Button>
                </HStack>


                {/* Snackbar */}
                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack(s => ({...s, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack(s => ({...s, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
