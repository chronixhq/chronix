import React, {useEffect, useMemo, useState} from 'react';
import {Box, CircularProgress, Divider, FormControl, IconButton, InputLabel, Link, MenuItem, Paper, Select, TextField, Tooltip, Typography} from '@mui/material';
import {DataGrid, type GridColDef} from '@mui/x-data-grid';
import RefreshIcon from '@mui/icons-material/Refresh';
import {useNavigate, useSearchParams} from 'react-router';
import {useRunsContext} from '../../data/RunsContext';
import {useJobs} from '../../data/JobsContext';
import {formatDateTime, RunStatusChip} from '../../lib/utilities.tsx';
import {SettingsAwareDateTimePicker} from '../../lib/SettingsAwareDateTimePicker';
import {HStack, VStack} from "@dsherwin/mui-kit";


export const RunsList: React.FC = () => {
    const [params] = useSearchParams();
    const navigate = useNavigate();
    const {useRunsList} = useRunsContext();

    const limit = Math.min(Number(params.get('limit') || 25), 100);
    const offset = Number(params.get('offset') || 0);

    // Local UI state for filtering and sorting (client-side only)
    const [search, setSearch] = useState<string>(params.get('q') || '');
    const [debouncedSearch, setDebouncedSearch] = useState<string>(params.get('q') || '');
    const [statusFilter, setStatusFilter] = useState<string>(params.get('status') || '');
    const [jobFilter, setJobFilter] = useState<string>(params.get('job') || '');
    const [startedFrom, setStartedFrom] = useState<string>(params.get('started_from') || '');
    const [startedTo, setStartedTo] = useState<string>(params.get('started_to') || '');

    const {items, total, loading, error, reload} = useRunsList({
        limit,
        offset,
        q: debouncedSearch || undefined,
        status: statusFilter || undefined,
        jobId: jobFilter || undefined,
        startedFrom: startedFrom || undefined,
        startedTo: startedTo || undefined,
    });
    const {items: jobItems, ensureLoaded: ensureJobsLoaded} = useJobs();

    useEffect(() => {
        void ensureJobsLoaded();
    }, [ensureJobsLoaded]);


    // DataGrid columns definition
    const columns: GridColDef[] = useMemo(() => ([
        {
            field: 'status',
            headerName: 'Status',
            width: 120,
            sortable: true,
            renderCell: (params) => <RunStatusChip status={params.row.status}/>,
            sortComparator: (a, b) => String(a || '').localeCompare(String(b || ''), undefined, {numeric: true, sensitivity: 'base'}),
        },
        {
            field: 'runId',
            headerName: 'Run ID',
            flex: 1,
            minWidth: 160,
            sortable: true,
            renderCell: (params) => (
                <Link component="button" underline="hover" onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    navigate(`/runs/${encodeURIComponent(params.row.runId)}`)
                }}>{params.row.runId}</Link>
            ),
        },
        {
            field: 'job',
            headerName: 'Job',
            flex: 1,
            minWidth: 160,
            valueGetter: (_value, row) => row.jobName || String(row.jobId),
            sortComparator: (a, b) => String(a || '').localeCompare(String(b || ''), undefined, {numeric: true, sensitivity: 'base'}),
        },
        {
            field: 'queuedAt',
            headerName: 'Queued',
            minWidth: 170,
            flex: 0.7,
            sortable: true,
            renderCell: (params) => <>{formatDateTime(params.row.queuedAt)}</>,
            sortComparator: (a, b) => String(a || '').localeCompare(String(b || '')),
        },
        {
            field: 'startedAt',
            headerName: 'Started',
            minWidth: 170,
            flex: 0.7,
            sortable: true,
            renderCell: (params) => <>{formatDateTime(params.row.startedAt)}</>,
            sortComparator: (a, b) => String(a || '').localeCompare(String(b || '')),
        },
        {
            field: 'finishedAt',
            headerName: 'Finished',
            minWidth: 170,
            flex: 0.7,
            sortable: true,
            renderCell: (params) => <>{formatDateTime(params.row.finishedAt)}</>,
            sortComparator: (a, b) => String(a || '').localeCompare(String(b || '')),
        },
        {
            field: 'durationMs',
            headerName: 'Duration',
            minWidth: 120,
            flex: 0.5,
            type: 'number',
            valueGetter: (_value, row) => row.durationMs ?? -1,
            renderCell: (params) => {
                const ms = params.row.durationMs;
                return <>{ms != null ? `${Math.round(ms / 1000)}s` : ''}</>
            },
        },
        {
            field: 'message',
            headerName: 'Message',
            flex: 1.5,
            minWidth: 200,
            sortable: true,
            valueGetter: (_value, row) => row.message || '',
        },
    ]), [navigate]);

    const setParamAndNavigate = (next: Record<string, string | number | undefined>) => {
        const np = new URLSearchParams(params);
        Object.entries(next).forEach(([k, v]) => {
            if (v === undefined || v === '') np.delete(k);
            else np.set(k, String(v));
        });
        navigate({pathname: '/runs', search: `?${np.toString()}`});
    };

    const pushFilters = (next?: { q?: string; status?: string; job?: string; started_from?: string; started_to?: string; offset?: number; limit?: number }) => {
        setParamAndNavigate({
            q: next && 'q' in next ? next.q : (search || undefined),
            status: next && 'status' in next ? next.status : (statusFilter || undefined),
            job: next && 'job' in next ? next.job : (jobFilter || undefined),
            started_from: next && 'started_from' in next ? next.started_from : (startedFrom || undefined),
            started_to: next && 'started_to' in next ? next.started_to : (startedTo || undefined),
            offset: next && 'offset' in next ? next.offset : 0,
            limit: next && 'limit' in next ? next.limit! : limit,
        });
    };

    // Debounce raw search input into a debounced value
    useEffect(() => {
        const handle = setTimeout(() => {
            setDebouncedSearch(search);
        }, 500);
        return () => clearTimeout(handle);
    }, [search]);

    // When debounced value changes, apply filter (and reset to first page)
    useEffect(() => {
        pushFilters({q: debouncedSearch || undefined, offset: 0});
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [debouncedSearch]);

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
                        <Typography variant="h5">Runs</Typography>
                        <Box>
                            {loading && <CircularProgress size={20} sx={{mr: 1}}/>}
                            <Tooltip title="Refresh">
                                <IconButton onClick={() => void reload()}><RefreshIcon/></IconButton>
                            </Tooltip>
                        </Box>
                    </Box>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    {error && (
                        <Typography color="error" variant="body2">{error}</Typography>
                    )}

                    <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                        <TextField size="small" label="Search" value={search} onChange={e => setSearch(e.target.value)} placeholder="Run ID, Job, Status, Message"/>
                        <FormControl size="small" sx={{minWidth: 160}}>
                            <InputLabel id="runs-status-label">Status</InputLabel>
                            <Select labelId="runs-status-label" label="Status" value={statusFilter} onChange={e => {
                                const v = String(e.target.value);
                                setStatusFilter(v);
                                pushFilters({status: v || undefined, offset: 0});
                            }}>
                                <MenuItem value="">All</MenuItem>
                                <MenuItem value="queued">Queued</MenuItem>
                                <MenuItem value="running">Running</MenuItem>
                                <MenuItem value="success">Success</MenuItem>
                                <MenuItem value="error">Error</MenuItem>
                                <MenuItem value="canceled">Canceled</MenuItem>
                            </Select>
                        </FormControl>
                        <FormControl size="small" sx={{minWidth: 200}}>
                            <InputLabel id="runs-job-label">Job</InputLabel>
                            <Select labelId="runs-job-label" label="Job" value={jobFilter} onChange={e => {
                                const v = String(e.target.value);
                                setJobFilter(v);
                                pushFilters({job: v || undefined, offset: 0});
                            }}>
                                <MenuItem value="">All Jobs</MenuItem>
                                {jobItems
                                    .slice()
                                    .sort((a, b) => (a.name || '').localeCompare(b.name || ''))
                                    .map(j => (
                                        <MenuItem key={String(j.id)} value={String(j.id)}>
                                            {j.name || `Job ${String(j.id)}`} {j.name ? `(id: ${String(j.id)})` : ''}
                                        </MenuItem>
                                    ))}
                            </Select>
                        </FormControl>
                        <SettingsAwareDateTimePicker
                            label="Started from"
                            valueIso={startedFrom || null}
                            onChangeIso={(iso) => {
                                setStartedFrom(iso || '');
                                pushFilters({started_from: iso || undefined, offset: 0});
                            }}
                            sx={{minWidth: 260}}
                        />
                        <SettingsAwareDateTimePicker
                            label="Started to"
                            valueIso={startedTo || null}
                            onChangeIso={(iso) => {
                                setStartedTo(iso || '');
                                pushFilters({started_to: iso || undefined, offset: 0});
                            }}
                            sx={{minWidth: 260}}
                        />
                    </HStack>
                    <Paper variant="outlined">
                        <DataGrid
                            density="compact"
                            rows={items}
                            columns={columns}
                            getRowId={(row) => row.runId}
                            loading={loading}
                            disableRowSelectionOnClick
                            disableColumnMenu
                            pagination
                            paginationMode="server"
                            sortingMode="client"
                            rowCount={total}
                            pageSizeOptions={[5, 10, 25, 50, 100]}
                            paginationModel={{page: Math.floor(offset / limit), pageSize: limit}}
                            onPaginationModelChange={(model) => {
                                const nextOffset = model.page * model.pageSize;
                                setParamAndNavigate({offset: nextOffset, limit: model.pageSize, q: debouncedSearch || undefined, status: statusFilter || undefined, job: jobFilter || undefined, started_from: startedFrom || undefined, started_to: startedTo || undefined});
                            }}
                            onRowClick={(params) => navigate(`/runs/${encodeURIComponent(params.row.runId)}`)}
                        />
                    </Paper>
                    <Box
                        sx={{
                            display: "flex",
                            justifyContent: "flex-end",
                            alignItems: "center",
                            gap: 2
                        }}>
                        <Typography variant="body2">{items.length} of {total}</Typography>
                    </Box>
                </VStack>
            </Box>
        </Box>
    );
};
