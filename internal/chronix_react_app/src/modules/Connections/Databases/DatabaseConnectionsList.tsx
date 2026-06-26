import {useEffect, useMemo, useState} from "react";
import {useLocation, useNavigate} from "react-router";
import {Alert, Box, Button, Card, Chip, Divider, FormControl, IconButton, InputLabel, LinearProgress, MenuItem, Popover, Select, Snackbar, TextField, Tooltip, Typography} from "@mui/material";
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {Add, Refresh} from "@mui/icons-material";
import {toastAPIError} from "../../../lib/errors.ts";
import {DatabaseConnectionCard} from '../components/ConnectionCards';
import {useFeatureAvailability} from "../../../data/FeatureAvailabilityContext.tsx";
import {useConnections} from "../../../data/ConnectionsContext.tsx";
import {SectionHelp} from "../../../main/SectionHelp";
import {HELP_SECTIONS} from "../../../main/appShellManifest.ts";
import type {DbConnection} from '../types.ts';
import {
    deleteStoredConnection,
    duplicateStoredConnection,
    getConnectionCreatePath,
    getConnectionEditPath,
    setStoredConnectionEnabled,
    testStoredConnection,
    toDatabaseConnectionRow,
} from '../api.ts';

type DbDriver = DbConnection['driver'];

export const DatabaseConnectionsList = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const {confirmPrompt} = useMuiPrompts();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {
        items,
        loading,
        error: storeError,
        reload: reloadConnections,
        ensureLoaded,
    } = useConnections();
    const dbLimit = checkLimit('db_connections');
    const [expanded, setExpanded] = useState<Record<string, boolean>>({});
    const [search, setSearch] = useState<string>("");
    const [filterStatus, setFilterStatus] = useState<'all' | 'ok' | 'error' | 'unknown'>('all');
    const [filterDriver, setFilterDriver] = useState<'all' | DbDriver>('all');
    const [sortBy, setSortBy] = useState<'name_asc' | 'name_desc' | 'driver_asc' | 'status'>('name_asc');
    const [page, setPage] = useState<number>(1);
    const [pageSize, setPageSize] = useState<number>(10);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' | 'warning' }>({open: false, message: '', severity: 'info'});
    const [testUI, setTestUI] = useState<{ open: boolean; id?: string; anchorEl: HTMLElement | null; status: 'starting' | 'connecting' | 'success' | 'error'; messages: string[]; timer?: number | null }>({open: false, anchorEl: null, status: 'starting', messages: [], timer: null});

    const EXPAND_KEY = 'dbconn_expanded';

    useEffect(() => {
        try {
            const exp = localStorage.getItem(EXPAND_KEY);
            if (exp) setExpanded(JSON.parse(exp));
        } catch {
        }
    }, []);

    useEffect(() => {
        try {
            localStorage.setItem(EXPAND_KEY, JSON.stringify(expanded));
        } catch {
        }
    }, [expanded]);

    useEffect(() => {
        void ensureLoaded();
    }, [ensureLoaded]);

    useEffect(() => {
        const st: any = location.state as any;
        if (st?.refresh) {
            void reloadConnections();
            navigate(location.pathname, {replace: true, state: {}});
        }
    }, [location.pathname, location.state, navigate, reloadConnections]);

    const rows = useMemo(() => {
        return items
            .filter((item): item is DbConnection => item.kind === 'database')
            .map(toDatabaseConnectionRow);
    }, [items]);

    const filteredRows = useMemo(() => {
        const q = search.trim().toLowerCase();
        return rows.filter((row) => {
            if (filterStatus !== 'all' && row.status !== filterStatus) return false;
            if (filterDriver !== 'all' && row.driver !== filterDriver) return false;
            if (!q) return true;
            return (
                row.name.toLowerCase().includes(q) ||
                row.host.toLowerCase().includes(q) ||
                (row.database || '').toLowerCase().includes(q) ||
                (row.description || '').toLowerCase().includes(q)
            );
        });
    }, [filterDriver, filterStatus, rows, search]);

    const sortedRows = useMemo(() => {
        const next = filteredRows.slice();
        switch (sortBy) {
            case 'name_desc':
                next.sort((a, b) => b.name.localeCompare(a.name));
                break;
            case 'driver_asc':
                next.sort((a, b) => (a.driver || '').localeCompare(b.driver || ''));
                break;
            case 'status':
                next.sort((a, b) => (a.status || '').localeCompare(b.status || ''));
                break;
            default:
                next.sort((a, b) => a.name.localeCompare(b.name));
        }
        return next;
    }, [filteredRows, sortBy]);

    const totalPages = Math.max(1, Math.ceil(sortedRows.length / pageSize));
    const pageRows = useMemo(() => {
        const start = (page - 1) * pageSize;
        return sortedRows.slice(start, start + pageSize);
    }, [page, pageSize, sortedRows]);

    useEffect(() => {
        if (page > totalPages) setPage(totalPages);
    }, [page, totalPages]);

    const toggleExpand = (id: string) => setExpanded((prev) => ({...prev, [id]: !prev[id]}));
    const isExpanded = (id: string) => !!expanded[id];

    const closeTestUI = () => {
        setTestUI((prev) => {
            if (prev.timer) window.clearTimeout(prev.timer);
            return {open: false, anchorEl: null, status: 'starting', messages: [], id: undefined, timer: null};
        });
    };

    const testConn = async (id: string, anchorEl: HTMLElement) => {
        setTestUI({open: true, id, anchorEl, status: 'starting', messages: ['Test started…'], timer: null});
        const connectTimeout = window.setTimeout(() => {
            setTestUI((prev) => ({...prev, status: 'connecting', messages: [...prev.messages, 'Connecting to host…']}));
        }, 400);
        try {
            const res = await testStoredConnection({kind: 'database', id}) as any;
            window.clearTimeout(connectTimeout);
            if (res?.ok === false) {
                setTestUI((prev) => {
                    const timer = window.setTimeout(closeTestUI, 5000);
                    return {...prev, status: 'error', messages: [...prev.messages, 'Error: Connection failed. Please check credentials and network.'], timer};
                });
                return;
            }
            setTestUI((prev) => ({...prev, messages: [...prev.messages, 'Validating credentials…']}));
            await new Promise((resolve) => setTimeout(resolve, 350));
            setTestUI((prev) => {
                const timer = window.setTimeout(closeTestUI, 5000);
                return {...prev, status: 'success', messages: [...prev.messages, 'Success: Connection established.'], timer};
            });
            void reloadConnections();
        } catch (e) {
            console.error(e);
            window.clearTimeout(connectTimeout);
            setTestUI((prev) => {
                const timer = window.setTimeout(closeTestUI, 5000);
                return {...prev, status: 'error', messages: [...prev.messages, 'Error: Network issue while testing connection.'], timer};
            });
        }
    };

    const confirmDelete = async (id: string, name: string) => {
        const ok = await confirmPrompt({
            title: 'Delete Connection',
            message: `Are you sure you want to delete the connection "${name}"?`,
            buttonText: 'Delete',
            buttonColor: 'error'
        });
        if (!ok) return;

        try {
            const res = await deleteStoredConnection({kind: 'database', id}) as any;
            if (res?.ok === false) throw new Error('Delete connection failed');
            setSnack({open: true, message: 'Connection deleted', severity: 'success'});
            void reloadFeatureAvailability();
            await reloadConnections();
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: toastAPIError(e, 'Delete failed'), severity: 'error'});
        }
    };

    const duplicateConn = async (id: string) => {
        try {
            const data = await duplicateStoredConnection({kind: 'database', id});
            setSnack({open: true, message: 'Connection duplicated successfully', severity: 'success'});
            await reloadFeatureAvailability();
            await reloadConnections();
            if ((data as any)?.id) {
                navigate(getConnectionEditPath('database', (data as any).id));
            }
        } catch (e: any) {
            setSnack({open: true, message: toastAPIError(e, 'Failed to duplicate connection'), severity: 'error'});
        }
    };

    const toggleEnabled = async (row: ReturnType<typeof toDatabaseConnectionRow>) => {
        if (!row.enabled && !dbLimit.allowed) {
            setSnack({open: true, message: dbLimit.message || 'Database connection limit reached', severity: 'warning'});
            return;
        }
        try {
            const res = await setStoredConnectionEnabled({kind: 'database', id: row.id}, !row.enabled) as any;
            if (res?.ok === false) throw new Error('Toggle failed');
            setSnack({open: true, message: !row.enabled ? 'Connection enabled' : 'Connection disabled', severity: 'success'});
            void reloadFeatureAvailability();
            await reloadConnections();
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: toastAPIError(e, 'Toggle failed'), severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" spacing={1} sx={{flexWrap: 'wrap'}}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5">Database connections</Typography>
                        <SectionHelp section={HELP_SECTIONS.connections} />
                    </Box>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Tooltip title="Refresh">
                            <span>
                                <IconButton onClick={() => void reloadConnections()} disabled={loading}>
                                    <Refresh/>
                                </IconButton>
                            </span>
                        </Tooltip>
                        <Button startIcon={<Add/>} variant="contained" onClick={() => navigate(getConnectionCreatePath('database'))} disabled={!dbLimit.allowed}>
                            New Database Connection
                        </Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {!dbLimit.allowed && (
                    <Alert severity="warning">
                        {dbLimit.message}
                    </Alert>
                )}

                <HStack spacing={1} sx={{flexWrap: 'wrap'}} alignItems="center">
                    <TextField
                        size="small"
                        placeholder="Search by name, host, db, or description"
                        value={search}
                        onChange={(e) => {
                            setSearch(e.target.value);
                            setPage(1);
                        }}
                        sx={{minWidth: {xs: '100%', sm: 320}}}
                    />
                    <FormControl size="small" sx={{minWidth: 140}}>
                        <InputLabel id="status-filter-label">Status</InputLabel>
                        <Select labelId="status-filter-label" label="Status" value={filterStatus} onChange={(e) => {
                            setFilterStatus(e.target.value as 'all' | 'ok' | 'error' | 'unknown');
                            setPage(1);
                        }}>
                            <MenuItem value="all">All</MenuItem>
                            <MenuItem value="ok">OK</MenuItem>
                            <MenuItem value="error">Error</MenuItem>
                            <MenuItem value="unknown">Unknown</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 160}}>
                        <InputLabel id="driver-filter-label">Driver</InputLabel>
                        <Select labelId="driver-filter-label" label="Driver" value={filterDriver} onChange={(e) => {
                            setFilterDriver(e.target.value as 'all' | DbDriver);
                            setPage(1);
                        }}>
                            <MenuItem value="all">All</MenuItem>
                            <MenuItem value="postgres">Postgres</MenuItem>
                            <MenuItem value="mysql">MySQL</MenuItem>
                            <MenuItem value="sqlite">SQLite</MenuItem>
                            <MenuItem value="mssql">MSSQL</MenuItem>
                            <MenuItem value="oracle">Oracle</MenuItem>
                            <MenuItem value="snowflake">Snowflake</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 180}}>
                        <InputLabel id="sortby-label">Sort by</InputLabel>
                        <Select labelId="sortby-label" label="Sort by" value={sortBy} onChange={(e) => {
                            setSortBy(e.target.value as any);
                            setPage(1);
                        }}>
                            <MenuItem value="name_asc">Name (A→Z)</MenuItem>
                            <MenuItem value="name_desc">Name (Z→A)</MenuItem>
                            <MenuItem value="driver_asc">Driver</MenuItem>
                            <MenuItem value="status">Status</MenuItem>
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

                {storeError && (
                    <Alert severity="error" action={<Button color="inherit" size="small" onClick={() => void reloadConnections()}>Retry</Button>}>
                        {storeError}
                    </Alert>
                )}

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
                ) : filteredRows.length === 0 ? (
                    <Card variant="outlined" sx={{borderRadius: 3, p: 3, textAlign: 'center'}}>
                        <Typography variant="h6">No connections found</Typography>
                        <Typography
                            sx={{
                                color: "text.secondary",
                                mt: 1
                            }}>
                            Try adjusting your search or create a new connection.
                        </Typography>
                        <Button sx={{mt: 2}} variant="contained" startIcon={<Add/>} onClick={() => navigate(getConnectionCreatePath('database'))} disabled={!dbLimit.allowed}>
                            New Connection
                        </Button>
                    </Card>
                ) : (
                    <VStack spacing={2}>
                        {pageRows.map((row) => (
                            <DatabaseConnectionCard
                                key={row.id}
                                row={row}
                                expanded={isExpanded(row.id)}
                                onToggleExpand={() => toggleExpand(row.id)}
                                onTest={(ev) => testConn(row.id, ev.currentTarget as HTMLElement)}
                                onEdit={() => navigate(getConnectionEditPath('database', row.id))}
                                onDelete={() => void confirmDelete(row.id, row.name)}
                                onDuplicate={() => void duplicateConn(row.id)}
                                onToggleEnabled={() => void toggleEnabled(row)}
                            />
                        ))}
                    </VStack>
                )}

                <Typography
                    variant="caption"
                    sx={{
                        color: "text.secondary",
                        mt: 1
                    }}>
                    {sortedRows.length} connection{sortedRows.length !== 1 ? 's' : ''} — page {page} of {totalPages}
                </Typography>
                <HStack spacing={1} sx={{justifyContent: 'flex-end'}}>
                    <Button size="small" disabled={page <= 1} onClick={() => setPage((current) => Math.max(1, current - 1))}>Prev</Button>
                    <Typography variant="caption" sx={{alignSelf: 'center'}}>Page {page} / {totalPages}</Typography>
                    <Button size="small" disabled={page >= totalPages} onClick={() => setPage((current) => Math.min(totalPages, current + 1))}>Next</Button>
                </HStack>

                <Popover
                    open={testUI.open}
                    anchorEl={testUI.anchorEl}
                    onClose={closeTestUI}
                    anchorOrigin={{vertical: 'top', horizontal: 'center'}}
                    transformOrigin={{vertical: 'bottom', horizontal: 'center'}}
                    disableRestoreFocus
                >
                    <Card variant="outlined" sx={{p: 1.5, maxWidth: 320}}>
                        {(testUI.status === 'starting' || testUI.status === 'connecting') && (
                            <LinearProgress sx={{mb: 1}}/>
                        )}
                        <VStack spacing={0.5}>
                            {testUI.messages.map((message, index) => (
                                <Typography key={index} variant="body2">{message}</Typography>
                            ))}
                        </VStack>
                        {testUI.status === 'success' && <Chip size="small" color="success" label="Success" sx={{mt: 1}}/>}
                        {testUI.status === 'error' && <Chip size="small" color="error" label="Error" sx={{mt: 1}}/>}
                        <Typography
                            variant="caption"
                            sx={{
                                color: "text.secondary",
                                mt: 0.5
                            }}>Click anywhere to dismiss</Typography>
                    </Card>
                </Popover>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack((current) => ({...current, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack((current) => ({...current, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
}
