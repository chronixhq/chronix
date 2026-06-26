import React, {useCallback, useEffect, useMemo, useState} from 'react'
import {Box, Button, Divider, FormControl, IconButton, InputLabel, Menu, MenuItem, Select, TextField, Tooltip, Typography} from '@mui/material'
import {DataGrid, type GridColDef} from '@mui/x-data-grid'
import RefreshIcon from '@mui/icons-material/Refresh'
import {useNavigate, useSearchParams} from 'react-router'
import {apiGet} from '@dsherwin/react-api-interface'
import {formatDateTime} from '../../lib/utilities'
import {useAuthContext} from '../../data/useAuthContext'
import {HStack, VStack} from '@dsherwin/mui-kit'
import {useFeatureAvailability} from "../../data/FeatureAvailabilityContext.tsx";
import {SectionHelp} from "../../main/SectionHelp";
import {Download} from "@mui/icons-material";
import {HELP_SECTIONS} from "../../main/appShellManifest.ts";

type ActivityItem = {
    id: string
    when: string // ISO
    action: string
    details?: string
    userId?: string
    user?: string
}

function toIsoString(value: unknown): string {
    if (typeof value === 'string') return value
    if (typeof value === 'number' || value instanceof Date) return new Date(value).toISOString()
    return ''
}

export const ActivityList: React.FC = () => {
    const navigate = useNavigate()
    const [params] = useSearchParams()
    const {user} = useAuthContext()
    const {data} = useFeatureAvailability()

    // URL param-backed state (keep consistent with Runs grid patterns)
    const limit = Math.min(Number(params.get('limit') || 25), 100)
    const offset = Number(params.get('offset') || 0)
    const [search, setSearch] = useState<string>(params.get('q') || '')
    const [debouncedSearch, setDebouncedSearch] = useState<string>(params.get('q') || '')
    const [actionFilter, setActionFilter] = useState<string>(params.get('action') || '')
    const [userFilter, setUserFilter] = useState<string>(params.get('user') || '')
    const [from, setFrom] = useState<string>(params.get('from') || '')
    const [to, setTo] = useState<string>(params.get('to') || '')

    const [items, setItems] = useState<ActivityItem[]>([])
    const [total, setTotal] = useState<number>(0)
    const [loading, setLoading] = useState<boolean>(false)
    const [error, setError] = useState<string | undefined>()

    const [exportAnchor, setExportAnchor] = useState<null | HTMLElement>(null)

    const onExport = (format: string) => {
        setExportAnchor(null)
        const q = params.get('q') || ''
        const action = params.get('action') || ''
        const u = params.get('user') || ''
        const fromIso = params.get('from') || ''
        const toIso = params.get('to') || ''

        const query = new URLSearchParams({
            format,
            q,
            action,
            user: u,
            from: fromIso,
            to: toIso
        })

        const url = `/activity/export?${query.toString()}`
        window.open(url, '_blank')
    }

    const columns: GridColDef[] = useMemo(() => [
        {field: 'when', headerName: 'When', minWidth: 180, flex: 0.8, sortable: false, valueGetter: (_v, r) => r.when, renderCell: (p) => <>{formatDateTime(p.row.when)}</>},
        ...(user?.admin ? [{field: 'user', headerName: 'User', minWidth: 160, flex: 0.8, sortable: false, valueGetter: (_v, r) => r.user || (r.userId === '0' || r.userId === 0 ? 'Chronix System' : (r.userId ? `User ${r.userId}` : ''))}] as GridColDef[] : []),
        {field: 'action', headerName: 'Action', minWidth: 160, flex: 0.8, sortable: false},
        {field: 'details', headerName: 'Details', minWidth: 220, flex: 1.5, sortable: false, valueGetter: (_v, r) => r.details || ''},
    ], [user?.admin])

    const setParamAndNavigate = (next: Record<string, string | number | undefined>) => {
        const np = new URLSearchParams(params)
        Object.entries(next).forEach(([k, v]) => {
            if (v === undefined || v === '') np.delete(k)
            else np.set(k, String(v))
        })
        navigate({pathname: '/activity', search: `?${np.toString()}`})
    }

    const pushFilters = (next?: { q?: string; action?: string; user?: string; from?: string; to?: string; offset?: number; limit?: number }) => {
        setParamAndNavigate({
            q: next && 'q' in next ? next.q : (search || undefined),
            action: next && 'action' in next ? next.action : (actionFilter || undefined),
            user: next && 'user' in next ? next.user : (userFilter || undefined),
            from: next && 'from' in next ? next.from : (from || undefined),
            to: next && 'to' in next ? next.to : (to || undefined),
            offset: next && 'offset' in next ? next.offset : 0,
            limit: next && 'limit' in next ? next.limit! : limit,
        })
    }

    // Debounce search
    useEffect(() => {
        const t = setTimeout(() => {
            if (search !== debouncedSearch) {
                setDebouncedSearch(search)
            }
        }, 500)
        return () => clearTimeout(t)
    }, [search, debouncedSearch])

    useEffect(() => {
        pushFilters({q: debouncedSearch || undefined, offset: 0})
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [debouncedSearch])

    const load = useCallback(async () => {
        setLoading(true);
        setError(undefined)
        try {
            const q = params.get('q') || ''
            const action = params.get('action') || ''
            const u = params.get('user') || ''
            const fromIso = params.get('from') || ''
            const toIso = params.get('to') || ''

            const query = new URLSearchParams({
                limit: String(limit),
                offset: String(offset),
                q,
                action,
                user: u,
                from: fromIso,
                to: toIso
            })

            const res = await apiGet(`/activity?${query.toString()}`) as {
                items?: Array<Record<string, unknown>>
                total?: number
            }
            const data = res.items || []
            const totalCount = res.total || 0

            const arr: ActivityItem[] = Array.isArray(data) ? data.map((d) => ({
                id: String(d.id ?? ''),
                when: toIsoString(d.when),
                action: String(d.action ?? ''),
                details: d.details != null ? String(d.details) : undefined,
                userId: d.userId != null ? String(d.userId) : undefined,
                user: d.user != null ? String(d.user) : undefined,
            })) : []
            setItems(arr)
            setTotal(totalCount)
        } catch (e) {
            console.error(e)
            setItems([])
            setTotal(0)
            setError('Failed to load activity')
        } finally {
            setLoading(false)
        }
    }, [limit, offset, params])

    // Initial + param-backed loads
    useEffect(() => {
        void load()
    }, [load])

    // Build a small list of distinct actions and users for filter dropdowns (based on current page)
    const actionOptions = useMemo(() => Array.from(new Set(items.map(i => i.action).filter(Boolean))).sort(), [items])
    const userOptions = useMemo(() => Array.from(new Set(items.map(i => i.user || i.userId).filter(Boolean))).sort(), [items])

    return (
        <Box sx={{
            p: 2
        }}>
            <Box sx={{width: '100%', maxWidth: 1280, mx: 'auto'}}>
                <VStack spacing={2}>
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
                            <Typography variant="h5">Activity</Typography>
                            <SectionHelp section={HELP_SECTIONS.activity}/>
                        </Box>
                        <HStack spacing={1}>
                            <Button
                                variant="outlined"
                                startIcon={<Download/>}
                                onClick={(e) => setExportAnchor(e.currentTarget)}
                            >
                                Export
                            </Button>
                            <Tooltip title="Refresh"><IconButton onClick={() => void load()}><RefreshIcon/></IconButton></Tooltip>
                        </HStack>
                    </Box>
                    <Menu
                        anchorEl={exportAnchor}
                        open={Boolean(exportAnchor)}
                        onClose={() => setExportAnchor(null)}
                    >
                        <MenuItem onClick={() => onExport('csv')} disabled={!data?.features.csvReports}>Export CSV</MenuItem>
                        <MenuItem onClick={() => onExport('html')} disabled={!data?.features.htmlReports}>Export HTML</MenuItem>
                        <MenuItem onClick={() => onExport('pdf')} disabled={!data?.features.pdfReports}>Export PDF</MenuItem>
                    </Menu>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                    {error && <Typography color="error" variant="body2">{error}</Typography>}

                    <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                        <TextField size="small" label="Search" value={search} onChange={e => setSearch(e.target.value)} placeholder="Action, Details, User"/>
                        <FormControl size="small" sx={{minWidth: 180}}>
                            <InputLabel id="activity-action-label">Action</InputLabel>
                            <Select labelId="activity-action-label" label="Action" value={actionFilter} onChange={e => {
                                const v = String(e.target.value);
                                setActionFilter(v);
                                pushFilters({action: v || undefined, offset: 0})
                            }}>
                                <MenuItem value="">All</MenuItem>
                                {actionOptions.map(opt => (<MenuItem key={opt} value={opt}>{opt}</MenuItem>))}
                            </Select>
                        </FormControl>
                        {user?.admin && (
                            <FormControl size="small" sx={{minWidth: 200}}>
                                <InputLabel id="activity-user-label">User</InputLabel>
                                <Select labelId="activity-user-label" label="User" value={userFilter} onChange={e => {
                                    const v = String(e.target.value);
                                    setUserFilter(v);
                                    pushFilters({user: v || undefined, offset: 0})
                                }}>
                                    <MenuItem value="">All</MenuItem>
                                    {userOptions.map(opt => (<MenuItem key={String(opt)} value={String(opt)}>{String(opt)}</MenuItem>))}
                                </Select>
                            </FormControl>
                        )}
                        <TextField
                            size="small"
                            label="From"
                            type="datetime-local"
                            value={from}
                            onChange={e => {
                                const v = e.target.value;
                                setFrom(v);
                                pushFilters({from: v || undefined, offset: 0})
                            }}
                            sx={{minWidth: 220}}
                            slotProps={{htmlInput: {step: 1}}}
                        />
                        <TextField
                            size="small"
                            label="To"
                            type="datetime-local"
                            value={to}
                            onChange={e => {
                                const v = e.target.value;
                                setTo(v);
                                pushFilters({to: v || undefined, offset: 0})
                            }}
                            sx={{minWidth: 220}}
                            slotProps={{htmlInput: {step: 1}}}
                        />
                    </HStack>

                    <div style={{width: '100%'}}>
                        <DataGrid
                            autoHeight
                            rows={items}
                            columns={columns}
                            density="compact"
                            getRowId={(r) => r.id}
                            loading={loading}
                            rowCount={total}
                            pageSizeOptions={[5, 10, 25, 50, 100]}
                            disableRowSelectionOnClick
                            disableColumnMenu
                            pagination
                            paginationMode="server"
                            paginationModel={{
                                pageSize: limit,
                                page: Math.floor(offset / Math.max(limit, 1))
                            }}
                            onPaginationModelChange={(m) => {
                                const nextLimit = m.pageSize
                                const nextOffset = m.page * m.pageSize
                                pushFilters({limit: nextLimit, offset: nextOffset})
                            }}
                        />
                    </div>
                </VStack>
            </Box>
        </Box>
    );
}
