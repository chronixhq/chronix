import {useEffect, useMemo, useState} from 'react'
import {Alert, Box, Button, Divider, FormControl, IconButton, InputLabel, LinearProgress, MenuItem, Select, TextField, Tooltip, Typography} from '@mui/material'
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit'
import {Add, Refresh} from '@mui/icons-material'
import {useNavigate} from 'react-router'
import {DatabaseConnectionCard, ShellConnectionCard, WebtaskConnectionCard} from '../components/ConnectionCards'
import {useFeatureAvailability} from '../../../data/FeatureAvailabilityContext.tsx'
import {useConnections} from '../../../data/ConnectionsContext.tsx'
import {SectionHelp} from '../../../main/SectionHelp'
import {HELP_SECTIONS} from '../../../main/appShellManifest.ts'
import {
    deleteStoredConnection,
    duplicateStoredConnection,
    getConnectionCreatePath,
    getConnectionEditPath,
    setStoredConnectionEnabled,
    testStoredConnection,
    type AllConnectionsRow,
    type ConnectionStatus,
    toAllConnectionsRow,
    toDatabaseConnectionRow,
    toShellConnectionRow,
    toWebtaskConnectionRow
} from '../api.ts'

export const AllConnectionsList = () => {
    const navigate = useNavigate()
    const {confirmPrompt} = useMuiPrompts()
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability()
    const {items, loading, error: storeError, reload: reloadConnections, ensureLoaded} = useConnections()
    const dbLimit = checkLimit('db_connections')
    const shLimit = checkLimit('shell_connections')
    const wtLimit = checkLimit('webtask_connections')
    const [localError, setLocalError] = useState<string | null>(null)

    const [filterStatus, setFilterStatus] = useState<'all' | ConnectionStatus>('all')
    const [filterKind, setFilterKind] = useState<'all' | 'database' | 'shell' | 'webtask'>('all')
    const [search, setSearch] = useState('')
    const [page, setPage] = useState(1)
    const [pageSize, setPageSize] = useState(10)
    const [expanded, setExpanded] = useState<Record<string, boolean>>({})

    useEffect(() => {
        void ensureLoaded()
    }, [ensureLoaded])

    const rows = useMemo(() => items.map(toAllConnectionsRow), [items])
    const itemsByKey = useMemo(() => {
        return Object.fromEntries(items.map((item) => [`${item.kind}-${item.id}`, item]))
    }, [items])

    const toggleExpand = (id: string) => setExpanded(prev => ({...prev, [id]: !prev[id]}))
    const isExpanded = (id: string) => !!expanded[id]

    const filteredRows = useMemo(() => {
        const q = search.trim().toLowerCase()
        return rows.filter((row) => {
            if (filterKind !== 'all' && row.kind !== filterKind) return false
            if (filterStatus !== 'all' && row.status !== filterStatus) return false
            if (q) {
                const hay = [row.name, row.description || '', row.hostPort || '', row.agent || ''].join(' ').toLowerCase()
                if (!hay.includes(q)) return false
            }
            return true
        }).sort((a, b) => a.name.localeCompare(b.name))
    }, [rows, filterKind, filterStatus, search])

    const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize))
    const pageRows = useMemo(() => {
        const start = (page - 1) * pageSize
        return filteredRows.slice(start, start + pageSize)
    }, [filteredRows, page, pageSize])

    useEffect(() => {
        if (page > totalPages) setPage(totalPages)
    }, [page, totalPages])

    async function handleTest(row: AllConnectionsRow) {
        setLocalError(null)
        try {
            await testStoredConnection({kind: row.kind, id: row.id})
        } catch (e: any) {
            setLocalError(e?.message || 'Failed to test connection')
        } finally {
            void reloadConnections()
        }
    }

    async function handleDelete(row: AllConnectionsRow) {
        const ok = await confirmPrompt({
            title: `Delete ${row.kind} connection`,
            message: `Delete ${row.kind} connection “${row.name}”?`,
            buttonText: 'Delete',
            buttonColor: 'error'
        })
        if (!ok) return
        setLocalError(null)
        try {
            await deleteStoredConnection({kind: row.kind, id: row.id})
            void reloadFeatureAvailability()
            await reloadConnections()
        } catch (e: any) {
            setLocalError(e?.message || 'Failed to delete connection')
        }
    }

    async function handleToggleEnabled(row: AllConnectionsRow) {
        const nextEnabled = !row.enabled
        if (nextEnabled) {
            if (row.kind === 'database' && !dbLimit.allowed) return
            if (row.kind === 'shell' && !shLimit.allowed) return
            if (row.kind === 'webtask' && !wtLimit.allowed) return
        }

        setLocalError(null)
        try {
            await setStoredConnectionEnabled({kind: row.kind, id: row.id}, nextEnabled)
            void reloadFeatureAvailability()
            await reloadConnections()
        } catch (e: any) {
            setLocalError(e?.message || 'Failed to update connection')
        }
    }

    async function handleDuplicate(row: AllConnectionsRow) {
        setLocalError(null)
        try {
            const res = await duplicateStoredConnection({kind: row.kind, id: row.id}) as any
            void reloadFeatureAvailability()
            void reloadConnections()
            if (res?.id) {
                navigate(getConnectionEditPath(row.kind, res.id))
            }
        } catch (e: any) {
            setLocalError(e?.message || 'Failed to duplicate connection')
        }
    }

    function handleEdit(row: AllConnectionsRow) {
        navigate(getConnectionEditPath(row.kind, row.id))
    }

    const error = localError || storeError

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" spacing={1} sx={{flexWrap: 'wrap'}}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5">Connections</Typography>
                        <SectionHelp section={HELP_SECTIONS.connections}/>
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
                        <Button startIcon={<Add/>} variant="contained" onClick={() => navigate(getConnectionCreatePath('shell'))} disabled={!shLimit.allowed}>
                            New Shell Connection
                        </Button>
                        <Button startIcon={<Add/>} variant="contained" onClick={() => navigate(getConnectionCreatePath('webtask'))} disabled={!wtLimit.allowed}>
                            New Web Task Connection
                        </Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {!dbLimit.allowed && <Alert severity="warning">{dbLimit.message}</Alert>}
                {!shLimit.allowed && <Alert severity="warning">{shLimit.message}</Alert>}
                {!wtLimit.allowed && <Alert severity="warning">{wtLimit.message}</Alert>}

                <HStack spacing={1} sx={{flexWrap: 'wrap'}} alignItems="center">
                    <TextField
                        size="small"
                        placeholder="Search by name, host, or description"
                        value={search}
                        onChange={(e) => {
                            setSearch(e.target.value);
                            setPage(1);
                        }}
                        sx={{minWidth: {xs: '100%', sm: 320}}}
                    />
                    <FormControl size="small" sx={{minWidth: 140}}>
                        <InputLabel id="kind-filter-label">Kind</InputLabel>
                        <Select labelId="kind-filter-label" label="Kind" value={filterKind} onChange={(e) => {
                            setFilterKind(e.target.value as any);
                            setPage(1);
                        }}>
                            <MenuItem value="all">All types</MenuItem>
                            <MenuItem value="database">Database</MenuItem>
                            <MenuItem value="shell">Shell</MenuItem>
                            <MenuItem value="webtask">Web Task</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{minWidth: 140}}>
                        <InputLabel id="status-filter-label">Status</InputLabel>
                        <Select labelId="status-filter-label" label="Status" value={filterStatus} onChange={(e) => {
                            setFilterStatus(e.target.value as any);
                            setPage(1);
                        }}>
                            <MenuItem value="all">All statuses</MenuItem>
                            <MenuItem value="ok">OK</MenuItem>
                            <MenuItem value="error">Error</MenuItem>
                            <MenuItem value="unknown">Unknown</MenuItem>
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

                {loading && (
                    <Box sx={{px: 2}}><LinearProgress/></Box>
                )}

                {error && (
                    <Alert severity="error">{error}</Alert>
                )}

                {!loading && !error && filteredRows.length === 0 && (
                    <Alert severity="info">No connections match the current filters.</Alert>
                )}

                <VStack spacing={2}>
                    {pageRows.map((row) => {
                        const key = `${row.kind}-${row.id}`
                        const source = itemsByKey[key]
                        if (!source) return null

                        if (source.kind === 'database') {
                            return (
                                <DatabaseConnectionCard
                                    key={key}
                                    row={toDatabaseConnectionRow(source)}
                                    expanded={isExpanded(key)}
                                    onToggleExpand={() => toggleExpand(key)}
                                    onTest={() => void handleTest(row)}
                                    onDuplicate={() => void handleDuplicate(row)}
                                    onEdit={() => handleEdit(row)}
                                    onDelete={() => void handleDelete(row)}
                                    onToggleEnabled={() => void handleToggleEnabled(row)}
                                />
                            )
                        }

                        if (source.kind === 'webtask') {
                            return (
                                <WebtaskConnectionCard
                                    key={key}
                                    row={toWebtaskConnectionRow(source)}
                                    expanded={isExpanded(key)}
                                    onToggleExpand={() => toggleExpand(key)}
                                    onTest={() => void handleTest(row)}
                                    onDuplicate={() => void handleDuplicate(row)}
                                    onEdit={() => handleEdit(row)}
                                    onDelete={() => void handleDelete(row)}
                                    onToggleEnabled={() => void handleToggleEnabled(row)}
                                />
                            )
                        }

                        return (
                            <ShellConnectionCard
                                key={key}
                                row={toShellConnectionRow(source)}
                                expanded={isExpanded(key)}
                                onToggleExpand={() => toggleExpand(key)}
                                onTest={() => void handleTest(row)}
                                onDuplicate={() => void handleDuplicate(row)}
                                onEdit={() => handleEdit(row)}
                                onDelete={() => void handleDelete(row)}
                                onToggleEnabled={() => void handleToggleEnabled(row)}
                            />
                        )
                    })}

                    <HStack sx={{justifyContent: 'space-between', mt: 1, flexWrap: 'wrap', gap: 1}}>
                        <Typography variant="caption" sx={{
                            color: "text.secondary"
                        }}>
                            {filteredRows.length} connection{filteredRows.length !== 1 ? 's' : ''} - page {page} of {totalPages}
                        </Typography>
                        <HStack spacing={1} sx={{justifyContent: 'flex-end'}}>
                            <Button size="small" disabled={page <= 1} onClick={() => setPage(p => Math.max(1, p - 1))}>Prev</Button>
                            <Typography variant="caption" sx={{alignSelf: 'center'}}>Page {page} / {totalPages}</Typography>
                            <Button size="small" disabled={page >= totalPages} onClick={() => setPage(p => Math.min(totalPages, p + 1))}>Next</Button>
                        </HStack>
                    </HStack>
                </VStack>
            </VStack>
        </Box>
    );
}

export default AllConnectionsList
