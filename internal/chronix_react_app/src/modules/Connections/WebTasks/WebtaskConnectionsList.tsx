import {useEffect, useMemo, useState} from "react";
import {useLocation, useNavigate} from "react-router";
import {Alert, Box, Button, Card, LinearProgress, Snackbar, TextField, Typography} from "@mui/material";
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {Add, Refresh} from "@mui/icons-material";
import {toastAPIError} from "../../../lib/errors.ts";
import {WebtaskConnectionCard} from '../components/ConnectionCards';
import {useFeatureAvailability} from "../../../data/FeatureAvailabilityContext.tsx";
import {useConnections} from "../../../data/ConnectionsContext.tsx";
import {SectionHelp} from "../../../main/SectionHelp";
import {HELP_SECTIONS} from "../../../main/appShellManifest.ts";
import type {WebtaskConnection} from '../types.ts';
import {
    deleteStoredConnection,
    duplicateStoredConnection,
    getConnectionCreatePath,
    getConnectionEditPath,
    setStoredConnectionEnabled,
    testStoredConnection,
    toWebtaskConnectionRow,
} from '../api.ts';

export const WebtaskConnectionsList = () => {
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
    const wtLimit = checkLimit('webtask_connections');
    const [expanded, setExpanded] = useState<Record<string, boolean>>({});
    const [search, setSearch] = useState<string>("");
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' | 'warning' }>({open: false, message: '', severity: 'info'});

    const EXPAND_KEY = 'webtaskconn_expanded';

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
            .filter((item): item is WebtaskConnection => item.kind === 'webtask')
            .map(toWebtaskConnectionRow);
    }, [items]);

    const filtered = useMemo(() => rows.filter((row) => {
        if (!search) return true;
        const value = search.toLowerCase();
        return (
            row.name.toLowerCase().includes(value) ||
            (row.description || "").toLowerCase().includes(value) ||
            (row.baseUrl || "").toLowerCase().includes(value)
        );
    }), [rows, search]);

    const toggleExpand = (id: string | number) => {
        setExpanded((prev) => ({...prev, [String(id)]: !prev[String(id)]}));
    };

    const handleDelete = async (id: string, name: string) => {
        const ok = await confirmPrompt({
            title: 'Delete Connection',
            message: `Are you sure you want to delete the connection "${name}"?`,
            buttonText: 'Delete',
            buttonColor: 'error'
        });
        if (!ok) return;

        try {
            await deleteStoredConnection({kind: 'webtask', id});
            setSnack({open: true, message: `Connection "${name}" deleted.`, severity: 'success'});
            void reloadFeatureAvailability();
            await reloadConnections();
        } catch (e: any) {
            setSnack({open: true, message: toastAPIError(e, `Delete "${name}" failed`), severity: 'error'});
        }
    };

    const handleTest = async (id: string) => {
        try {
            setSnack({open: true, message: 'Starting connectivity test...', severity: 'info'});
            const res = await testStoredConnection({kind: 'webtask', id}) as any;
            if (res?.ok === true) {
                setSnack({open: true, message: `Test successful: ${res.message || 'reachable'}`, severity: 'success'});
            } else if (res?.ok === false) {
                setSnack({open: true, message: `Test failed: ${res.message || 'unknown error'}`, severity: 'error'});
            } else {
                setSnack({open: true, message: 'Test request sent.', severity: 'info'});
            }
            void reloadFeatureAvailability();
            void reloadConnections();
        } catch (e: any) {
            setSnack({open: true, message: toastAPIError(e, "Test failed"), severity: 'error'});
        }
    };

    const handleDuplicate = async (id: string) => {
        try {
            const data = await duplicateStoredConnection({kind: 'webtask', id}) as any;
            setSnack({open: true, message: 'Connection duplicated successfully', severity: 'success'});
            await reloadFeatureAvailability();
            await reloadConnections();
            if (data?.id) {
                navigate(getConnectionEditPath('webtask', data.id));
            }
        } catch (e: any) {
            setSnack({open: true, message: toastAPIError(e, 'Failed to duplicate connection'), severity: 'error'});
        }
    };

    const toggleEnabled = async (row: ReturnType<typeof toWebtaskConnectionRow>) => {
        if (!row.enabled && !wtLimit.allowed) {
            setSnack({open: true, message: wtLimit.message || 'Webtask connection limit reached', severity: 'warning'});
            return;
        }
        try {
            const res = await setStoredConnectionEnabled({kind: 'webtask', id: row.id}, !row.enabled) as any;
            if (res?.ok === false) throw new Error('Toggle failed');
            setSnack({open: true, message: !row.enabled ? 'Connection enabled' : 'Connection disabled', severity: 'success'});
            void reloadFeatureAvailability();
            await reloadConnections();
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: toastAPIError(e, 'Toggle failed'), severity: 'error'});
        }
    };

    return (
        <VStack spacing={2} sx={{p: 3, maxWidth: 1200, mx: 'auto', width: '100%'}}>
            <HStack justifyContent="space-between" alignItems="center" sx={{flexWrap: 'wrap'}}>
                <VStack spacing={0.5}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5" sx={{
                            fontWeight: 700
                        }}>Web Task Connections</Typography>
                        <SectionHelp section={HELP_SECTIONS.connections} />
                    </Box>
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Manage external and internal API endpoints.</Typography>
                </VStack>
                <HStack spacing={1}>
                    <Button variant="outlined" startIcon={<Refresh/>} onClick={() => void reloadConnections()} disabled={loading}>Reload</Button>
                    <Button variant="contained" startIcon={<Add/>} onClick={() => navigate(getConnectionCreatePath('webtask'))} disabled={!wtLimit.allowed}>New Connection</Button>
                </HStack>
            </HStack>
            {!wtLimit.allowed && (
                <Alert severity="warning">
                    {wtLimit.message}
                </Alert>
            )}
            <Card sx={{p: 2}}>
                <HStack spacing={2} alignItems="center">
                    <TextField
                        size="small"
                        placeholder="Search connections..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        sx={{flex: 1}}
                    />
                </HStack>
            </Card>
            {loading && <LinearProgress sx={{borderRadius: 1}}/>}
            {storeError && (
                <Alert severity="error" action={<Button color="inherit" size="small" onClick={() => void reloadConnections()}>Retry</Button>}>
                    {storeError}
                </Alert>
            )}
            {!loading && filtered.length === 0 && (
                <Box sx={{py: 8, textAlign: 'center', opacity: 0.6}}>
                    <Typography variant="h6">No connections found</Typography>
                    <Typography variant="body2">Try a different search or create a new Web Task connection.</Typography>
                    <Button sx={{mt: 2}} variant="contained" startIcon={<Add/>} onClick={() => navigate(getConnectionCreatePath('webtask'))} disabled={!wtLimit.allowed}>
                        New Connection
                    </Button>
                </Box>
            )}
            <VStack spacing={2}>
                {filtered.map((row) => (
                    <WebtaskConnectionCard
                        key={row.id}
                        row={row}
                        expanded={!!expanded[row.id]}
                        onToggleExpand={() => toggleExpand(row.id)}
                        onTest={() => void handleTest(String(row.id))}
                        onEdit={() => navigate(getConnectionEditPath('webtask', String(row.id)))}
                        onDelete={() => void handleDelete(String(row.id), row.name)}
                        onDuplicate={() => void handleDuplicate(String(row.id))}
                        onToggleEnabled={() => void toggleEnabled(row)}
                    />
                ))}
            </VStack>
            <Snackbar
                open={snack.open}
                autoHideDuration={6000}
                onClose={() => setSnack((prev) => ({...prev, open: false}))}
                anchorOrigin={{vertical: 'top', horizontal: 'center'}}
            >
                <Alert severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                    {snack.message}
                </Alert>
            </Snackbar>
        </VStack>
    );
};
