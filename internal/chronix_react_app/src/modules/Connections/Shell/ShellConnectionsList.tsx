import {useEffect, useMemo, useState} from 'react'
import {Alert, Box, Button, Card, Divider, FormControl, IconButton, InputLabel, MenuItem, Select, TextField, Tooltip, Typography} from '@mui/material'
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit'
import {Add, Refresh} from '@mui/icons-material'
import {useLocation, useNavigate} from 'react-router'
import {ShellConnectionCard} from '../components/ConnectionCards'
import {useFeatureAvailability} from "../../../data/FeatureAvailabilityContext.tsx";
import {useConnections} from "../../../data/ConnectionsContext.tsx";
import {SectionHelp} from "../../../main/SectionHelp";
import {HELP_SECTIONS} from "../../../main/appShellManifest.ts";
import type {ShellConnection} from '../types.ts'
import {
    deleteStoredConnection,
    duplicateStoredConnection,
    getConnectionCreatePath,
    getConnectionEditPath,
    setStoredConnectionEnabled,
    testStoredConnection,
    toShellConnectionRow,
} from '../api.ts'

export const ShellConnectionsList = () => {
    const navigate = useNavigate()
    const location = useLocation()
    const {confirmPrompt} = useMuiPrompts()
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability()
    const {
        items,
        loading,
        error: storeError,
        reload: reloadConnections,
        ensureLoaded,
    } = useConnections()
    const shLimit = checkLimit('shell_connections')
    const [testMsg, setTestMsg] = useState<string | null>(null)

    const [filterStatus, setFilterStatus] = useState<'all' | 'ok' | 'error' | 'unknown'>('all')
    const [filterMode, setFilterMode] = useState<'all' | 'localhost' | 'ssh'>('all')
    const [search, setSearch] = useState('')
    const [page, setPage] = useState(1)
    const [pageSize, setPageSize] = useState(10)
    const [expanded, setExpanded] = useState<Record<string, boolean>>({})

    useEffect(() => {
        void ensureLoaded()
    }, [ensureLoaded])

    useEffect(() => {
        const st: any = location.state as any
        if (st?.refresh) {
            void reloadConnections()
            navigate(location.pathname, {replace: true, state: {}})
        }
    }, [location.pathname, location.state, navigate, reloadConnections])

    const rows = useMemo(() => {
        return items
            .filter((item): item is ShellConnection => item.kind === 'shell')
            .map(toShellConnectionRow)
    }, [items])

    const toggleExpand = (id: string) => setExpanded((prev) => ({...prev, [id]: !prev[id]}))
    const isExpanded = (id: string) => !!expanded[id]

    function deriveStatus(row: ReturnType<typeof toShellConnectionRow>): 'ok' | 'error' | 'unknown' {
        if (row.lastStatus?.toLowerCase() === 'ok') return 'ok'
        if (row.lastStatus?.toLowerCase() === 'error' || row.lastError) return 'error'
        return 'unknown'
    }

    const filteredRows = useMemo(() => {
        const q = search.trim().toLowerCase()
        return rows.filter((row) => {
            const status = deriveStatus(row)
            if (filterStatus !== 'all' && status !== filterStatus) return false
            if (filterMode !== 'all' && row.mode !== filterMode) return false
            if (q) {
                const hostPort = `${row.host || ''}${row.port ? `:${row.port}` : ''}`.toLowerCase()
                const haystack = [row.name, hostPort, row.description || '', row.agent_uuid || ''].join(' ').toLowerCase()
                if (!haystack.includes(q)) return false
            }
            return true
        })
    }, [filterMode, filterStatus, rows, search])

    const sortedRows = useMemo(() => {
        return [...filteredRows].sort((a, b) => a.name.localeCompare(b.name))
    }, [filteredRows])

    const totalPages = Math.max(1, Math.ceil(sortedRows.length / pageSize))
    const pageRows = useMemo(() => {
        const start = (page - 1) * pageSize
        return sortedRows.slice(start, start + pageSize)
    }, [page, pageSize, sortedRows])

    useEffect(() => {
        if (page > totalPages) setPage(totalPages)
    }, [page, totalPages])

    async function doTest(id: string) {
        setTestMsg(null)
        try {
            const res = await testStoredConnection({kind: 'shell', id}) as any
            setTestMsg(res?.message || (res?.ok ? 'ok' : 'failed'))
        } catch {
            setTestMsg('test failed')
        } finally {
            void reloadConnections()
        }
    }

    async function doDelete(id: string, name: string) {
        const ok = await confirmPrompt({
            title: 'Delete shell connection',
            message: `Are you sure you want to delete "${name}"?`,
            buttonText: 'Delete',
            buttonColor: 'error'
        })
        if (!ok) return
        try {
            await deleteStoredConnection({kind: 'shell', id})
            void reloadFeatureAvailability()
            await reloadConnections()
        } catch {
            setTestMsg('Delete failed')
        }
    }

    async function doDuplicate(id: string) {
        try {
            const data = await duplicateStoredConnection({kind: 'shell', id}) as any
            await reloadFeatureAvailability()
            await reloadConnections()
            if (data?.id) {
                navigate(getConnectionEditPath('shell', data.id))
            }
        } catch (e: any) {
            setTestMsg(e?.message || 'Failed to duplicate connection')
        }
    }

    async function doToggleEnabled(row: ReturnType<typeof toShellConnectionRow>) {
        if (!row.enabled && !shLimit.allowed) {
            setTestMsg(shLimit.message || 'Shell connection limit reached')
            return
        }
        try {
            await setStoredConnectionEnabled({kind: 'shell', id: row.id}, !row.enabled)
            void reloadFeatureAvailability()
            await reloadConnections()
        } catch (e: any) {
            setTestMsg(e?.message || 'Toggle failed')
        }
    }

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{mb: 2, flexWrap: 'wrap'}}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5">Shell connections</Typography>
                        <SectionHelp section={HELP_SECTIONS.connections} />
                    </Box>
                    <HStack spacing={1}>
                        <Tooltip title="Refresh">
                            <span>
                                <IconButton onClick={() => void reloadConnections()} disabled={loading}>
                                    <Refresh/>
                                </IconButton>
                            </span>
                        </Tooltip>
                        <Button startIcon={<Add/>} variant="contained" onClick={() => navigate(getConnectionCreatePath('shell'))} disabled={!shLimit.allowed}>New Connection</Button>
                    </HStack>
                </HStack>

                <Divider sx={{mb: 2, borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {!shLimit.allowed && (
                    <Alert severity="warning" sx={{mb: 2}}>
                        {shLimit.message}
                    </Alert>
                )}

                <HStack spacing={1} alignItems="center" sx={{flexWrap: 'wrap', mb: 2}}>
                    <TextField
                        size="small"
                        placeholder="Search by name, host, description, or agent"
                        value={search}
                        onChange={(e) => {
                            setSearch(e.target.value)
                            setPage(1)
                        }}
                        sx={{minWidth: {xs: '100%', sm: 320}}}
                    />
                    <FormControl size="small" sx={{minWidth: 140}}>
                        <InputLabel id="status-filter-label">Status</InputLabel>
                        <Select labelId="status-filter-label" label="Status" value={filterStatus} onChange={(e) => {
                            setFilterStatus(e.target.value as any)
                            setPage(1)
                        }}>
                            <MenuItem value="all">All</MenuItem>
                            <MenuItem value="ok">OK</MenuItem>
                            <MenuItem value="error">Error</MenuItem>
                            <MenuItem value="unknown">Unknown</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 160}}>
                        <InputLabel id="mode-filter-label">Mode</InputLabel>
                        <Select labelId="mode-filter-label" label="Mode" value={filterMode} onChange={(e) => {
                            setFilterMode(e.target.value as any)
                            setPage(1)
                        }}>
                            <MenuItem value="all">All</MenuItem>
                            <MenuItem value="localhost">localhost</MenuItem>
                            <MenuItem value="ssh">ssh</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 120}}>
                        <InputLabel id="psize-label">Page size</InputLabel>
                        <Select labelId="psize-label" label="Page size" value={String(pageSize)} onChange={(e) => {
                            setPageSize(Number(e.target.value))
                            setPage(1)
                        }}>
                            <MenuItem value={10}>10</MenuItem>
                            <MenuItem value={20}>20</MenuItem>
                            <MenuItem value={50}>50</MenuItem>
                        </Select>
                    </FormControl>
                </HStack>

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
                ) : storeError ? (
                    <Alert severity="error" action={<Button color="inherit" size="small" onClick={() => void reloadConnections()}>Retry</Button>}>
                        {storeError}
                    </Alert>
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
                        <Button sx={{mt: 2}} variant="contained" startIcon={<Add/>} onClick={() => navigate(getConnectionCreatePath('shell'))} disabled={!shLimit.allowed}>New Connection</Button>
                    </Card>
                ) : (
                    <VStack spacing={2}>
                        {pageRows.map((row) => (
                            <ShellConnectionCard
                                key={row.id}
                                row={row}
                                expanded={isExpanded(String(row.id))}
                                onToggleExpand={() => toggleExpand(String(row.id))}
                                onTest={() => void doTest(String(row.id))}
                                onEdit={() => navigate(getConnectionEditPath('shell', String(row.id)))}
                                onDelete={() => void doDelete(String(row.id), row.name)}
                                onDuplicate={() => void doDuplicate(String(row.id))}
                                onToggleEnabled={() => void doToggleEnabled(row)}
                            />
                        ))}
                    </VStack>
                )}

                <HStack sx={{justifyContent: 'space-between', mt: 1, flexWrap: 'wrap', gap: 1}}>
                    <Typography variant="caption" sx={{
                        color: "text.secondary"
                    }}>
                        {sortedRows.length} connection{sortedRows.length !== 1 ? 's' : ''} — page {page} of {totalPages}
                    </Typography>
                    <HStack spacing={1} sx={{justifyContent: 'flex-end'}}>
                        <Button size="small" disabled={page <= 1} onClick={() => setPage((current) => Math.max(1, current - 1))}>Prev</Button>
                        <Typography variant="caption" sx={{alignSelf: 'center'}}>Page {page} / {totalPages}</Typography>
                        <Button size="small" disabled={page >= totalPages} onClick={() => setPage((current) => Math.min(totalPages, current + 1))}>Next</Button>
                    </HStack>
                </HStack>

                {!!testMsg && <Alert sx={{mt: 2}} severity="info" onClose={() => setTestMsg(null)}>{testMsg}</Alert>}
            </VStack>
        </Box>
    );
}
