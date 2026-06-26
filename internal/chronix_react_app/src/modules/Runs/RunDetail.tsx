import React from 'react';
import {useRunsContext} from '../../data/RunsContext';
import {Alert, Box, Card, CardContent, Chip, Divider, IconButton, List, ListItem, ListItemText, Paper, Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Tooltip, Typography} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import CancelIcon from '@mui/icons-material/Cancel';
import ReplayIcon from '@mui/icons-material/Replay';
import {useNavigate, useParams} from 'react-router';
import {formatDateTime, RunStatusChip} from '../../lib/utilities.tsx';
import {cancelRun, rerunRun} from './api.ts';
import {useRunProgressSse} from './useRunProgressSse.ts';
import type {RunStepDetail} from './types.ts';

function toRecord(value: unknown): Record<string, unknown> {
    return value && typeof value === 'object' ? (value as Record<string, unknown>) : {}
}

function toStringValue(value: unknown): string | undefined {
    if (typeof value === 'string') return value
    if (typeof value === 'number') return String(value)
    return undefined
}

function toNumberValue(value: unknown): number | undefined {
    if (typeof value === 'number' && Number.isFinite(value)) return value
    if (typeof value === 'string' && value.trim()) {
        const parsed = Number(value)
        if (Number.isFinite(parsed)) return parsed
    }
    return undefined
}

function toRecordArray(value: unknown): Array<Record<string, unknown>> {
    return Array.isArray(value)
        ? value.map((item) => toRecord(item)).filter((item) => Object.keys(item).length > 0)
        : []
}


export const RunDetail: React.FC = () => {
    const {runId} = useParams<{ runId: string }>();
    const navigate = useNavigate();

    const {useRun} = useRunsContext();
    const {run, steps, snapshot, error, reload} = useRun(runId);
    const progress = useRunProgressSse(runId);
    React.useEffect(() => {
        if (!runId) return;
        const reloadTimer = window.setTimeout(() => {
            if (progress.refreshToken > 0) {
                reload();
            }
        }, 300);
        return () => {
            window.clearTimeout(reloadTimer);
        };
    }, [progress.refreshToken, reload, runId]);

    const onCancel = async () => {
        if (!runId) return;
        try {
            await cancelRun(runId);
            reload();
        } catch (e) {
            console.error(e);
        }
    };
    const onRerun = async () => {
        if (!runId) return;
        try {
            await rerunRun(runId);
            // navigate to runs list; new run will appear
            navigate('/runs');
        } catch (e) {
            console.error(e);
        }
    };

    const currentStatus = progress.status || snapshot?.status || run?.status;

    return (
        <Box
            sx={{
                p: 2,
                width: '100%',
                maxWidth: 1000,
                mx: 'auto'
            }}>
            <Stack
                direction="row"
                sx={{
                    alignItems: "center",
                    justifyContent: "space-between"
                }}>
                <Stack direction="row" spacing={3} sx={{
                    alignItems: "center"
                }}>
                    <Stack direction="row" spacing={2} sx={{
                        alignItems: "center"
                    }}>
                        <Typography variant="h5">Run {runId}</Typography>
                        {currentStatus && <RunStatusChip status={String(currentStatus)}/>}
                    </Stack>
                    <Stack>
                        {run?.jobName && <Typography
                            variant="body2"
                            sx={{
                                color: "text.secondary",
                                lineHeight: 1.4
                            }}>Job: {run.jobName}</Typography>}
                        {run?.actionName && <Typography
                            variant="body2"
                            sx={{
                                color: "text.secondary",
                                lineHeight: 1.4
                            }}>Action: {run.actionName}</Typography>}
                        {run?.connectionName && <Typography
                            variant="body2"
                            sx={{
                                color: "text.secondary",
                                lineHeight: 1.4
                            }}>Connection: {run.connectionName}</Typography>}
                    </Stack>
                </Stack>
                <Stack direction="row" spacing={1}>
                    <Tooltip title="Refresh">
                        <IconButton onClick={() => void reload()}><RefreshIcon/></IconButton>
                    </Tooltip>
                    <Tooltip title="Cancel run">
            <span>
              <IconButton onClick={() => void onCancel()} disabled={currentStatus !== 'running'}><CancelIcon/></IconButton>
            </span>
                    </Tooltip>
                    <Tooltip title="Re-run">
                        <IconButton onClick={() => void onRerun()}><ReplayIcon/></IconButton>
                    </Tooltip>
                </Stack>
            </Stack>
            <Divider sx={{my: 2, borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
            {error && <Alert severity="error" sx={{mb: 2}}>{error}</Alert>}
            {/* Snapshot / Summary */}
            <Card variant="outlined" sx={{mb: 2}}>
                <CardContent>
                    <Stack direction={{xs: 'column', md: 'row'}} spacing={2} divider={<Divider flexItem orientation="vertical" sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>}>
                        <Box>
                            <Typography variant="overline">Status</Typography>
                            <Typography variant="body1">{currentStatus || 'unknown'}</Typography>
                        </Box>
                        <Box>
                            <Typography variant="overline">Queued</Typography>
                            <Typography variant="body1">{formatDateTime(run?.queuedAt)}</Typography>
                        </Box>
                        <Box>
                            <Typography variant="overline">Started</Typography>
                            <Typography variant="body1">{formatDateTime(run?.startedAt)}</Typography>
                        </Box>
                        <Box>
                            <Typography variant="overline">Finished</Typography>
                            <Typography variant="body1">{formatDateTime(run?.finishedAt)}</Typography>
                        </Box>
                        <Box>
                            <Typography variant="overline">Rows affected</Typography>
                            <Typography variant="body1">{run?.rowsAffected ?? '-'}</Typography>
                        </Box>
                    </Stack>
                    {run?.message && (
                        <Typography
                            variant="body2"
                            sx={{
                                color: "text.secondary",
                                mt: 1
                            }}>{run.message}</Typography>
                    )}
                </CardContent>
            </Card>
            {/* Steps */}
            <Typography variant="h6" gutterBottom>Steps</Typography>
            <Stack spacing={1.5} sx={{mb: 2}}>
                {steps.map((s: RunStepDetail) => {
                    const status = String(s.status || '')
                    const showDetails = status === 'success' || status === 'error'
                    const e = toRecord(s.expectation)
                    const d = toRecord(s.details)
                    const rowsCount = typeof s.rowsCount === 'number' ? s.rowsCount : toNumberValue(d.rows_count)
                    const rowsAffected = typeof s.rowsAffected === 'number' ? s.rowsAffected : toNumberValue(d.rows_affected)
                    const errorCode = s.errorCode || toStringValue(d.error_code)
                    const errorMessage = s.errorMessage || toStringValue(d.error_message) || s.expectMessage

                    const renderSuccessLines = () => {
                        const lines: string[] = []
                        const kind = String(d.expect_kind || e.kind || '')
                        if (!e || Object.keys(e).length === 0) {
                            lines.push('No errors detected.')
                        } else if (kind === 'rowExists' || e.kind === 'rowExists') {
                            lines.push(`Expected at least one row; received ${rowsCount ?? 0}.`)
                        } else if (kind === 'noRowsReturned' || e.kind === 'noRowsReturned') {
                            lines.push(`Expected zero rows; received ${rowsCount ?? 0}.`)
                        } else if (kind === 'rowsAffected' || e.kind === 'rowsAffected') {
                            const op = String(d.expect_op || e.op || '==')
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Rows affected was ${rowsAffected ?? 0}; met expectation (${op} ${val}).`)
                        } else if (kind === 'fieldEquals' || e.kind === 'fieldEquals' || e.kind === 'fieldEqualsFirst' || e.kind === 'fieldEqualsLast') {
                            const rowSel = String(d.expect_row || (e.kind === 'fieldEqualsLast' ? 'last' : 'first') || 'first')
                            const col = String(d.expect_column || e.column || '')
                            const expected = d.expect_expected !== undefined ? String(d.expect_expected) : (e.expected != null ? String(e.expected) : '')
                            if (col && expected !== '') {
                                lines.push(`${rowSel} row assertion met: ${col} = "${expected}".`)
                            } else {
                                lines.push('Expectation met.')
                            }
                        } else if (kind === 'noError' || e.kind === 'noError') {
                            lines.push('Completed successfully.')
                        } else if (kind === 'exitCodeEquals' || e.kind === 'exitCodeEquals') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '0')
                            lines.push(`Exit code was ${s.exitCode ?? 0}; met expectation (== ${val}).`)
                        } else if (kind === 'contains' || e.kind === 'contains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Output contained expected string: "${val}".`)
                        } else if (kind === 'notContains' || e.kind === 'notContains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Output did not contain forbidden string: "${val}".`)
                        } else if (kind === 'firstLineEquals' || e.kind === 'firstLineEquals') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`First line assertion met: "${val}".`)
                        } else if (kind === 'lastLineEquals' || e.kind === 'lastLineEquals') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Last line assertion met: "${val}".`)
                        } else if (kind === 'regexMatch' || e.kind === 'regexMatch') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Output matched regex pattern: "${val}".`)
                        } else if (kind === 'statusCode' || e.kind === 'statusCode') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '200')
                            lines.push(`Status code was ${s.responseStatus ?? 'unknown'}; met expectation (== ${val}).`)
                        } else if (kind === 'bodyContains' || e.kind === 'bodyContains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Body contained expected string: "${val}".`)
                        } else if (kind === 'bodyNotContains' || e.kind === 'bodyNotContains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Body did not contain forbidden string: "${val}".`)
                        } else if (kind === 'jsonPathMatch' || e.kind === 'jsonPathMatch') {
                            const path = d.expect_path || e.path || ''
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`JSONPath "${path}" matched expected value: "${val}".`)
                        } else if (kind === 'latency' || e.kind === 'latency') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '1000')
                            lines.push(`Latency was ${s.latencyMs ?? 0}ms; met expectation (<= ${val}ms).`)
                        } else {
                            // Legacy expectations
                            if (e.rows_min != null) lines.push(`Rows returned was ${rowsCount ?? 0}; met expectation (>= ${String(e.rows_min)}).`)
                            else if (e.rows_max != null) lines.push(`Rows returned was ${rowsCount ?? 0}; met expectation (<= ${String(e.rows_max)}).`)
                            else if (e.affected_min != null) lines.push(`Rows affected was ${rowsAffected ?? 0}; met expectation (>= ${String(e.affected_min)}).`)
                            else if (e.assert) lines.push(`Assertion met: ${String(e.assert)}`)
                            else lines.push('Expectation met.')
                        }
                        return lines
                    }

                    const renderErrorLines = () => {
                        const lines: string[] = []
                        const kind = String(d.expect_kind || e.kind || '')
                        if ((errorCode && String(errorCode) === 'expectation_eval_error') || (s.expectOk === false && s.expectMessage && String(s.expectMessage).startsWith('evaluation error'))) {
                            lines.push(`Expectation evaluation error: ${String(errorMessage || s.expectMessage || '')}`)
                            return lines
                        }
                        if (kind === 'fieldEquals' || e.kind === 'fieldEquals' || e.kind === 'fieldEqualsFirst' || e.kind === 'fieldEqualsLast') {
                            const rowSel = String(d.expect_row || (e.kind === 'fieldEqualsLast' ? 'last' : 'first') || 'first')
                            const col = String(d.expect_column || e.column || '')
                            const expected = d.expect_expected !== undefined ? String(d.expect_expected) : (e.expected != null ? String(e.expected) : '')
                            const actual = d.expect_actual !== undefined ? String(d.expect_actual) : ''
                            if (col && expected !== '') {
                                lines.push(`The ${rowSel} result row was expected to have:`)
                                lines.push(`${col} = "${expected}"`)
                                lines.push('What was received is:')
                                lines.push(`${col} = "${actual}"`)
                            } else if (s.expectMessage) {
                                lines.push(String(s.expectMessage))
                            }
                        } else if (kind === 'rowExists' || e.kind === 'rowExists') {
                            lines.push(`Expected at least one row; received ${rowsCount ?? 0}.`)
                        } else if (kind === 'rowsAffected' || e.kind === 'rowsAffected') {
                            const op = String(d.expect_op || e.op || '>=')
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Expected rows affected ${op} ${val}; received ${rowsAffected ?? 0}.`)
                        } else if (kind === 'exitCodeEquals' || e.kind === 'exitCodeEquals') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '0')
                            lines.push(`Expected exit code ${val}; received ${s.exitCode ?? 0}.`)
                        } else if (kind === 'contains' || e.kind === 'contains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Output did not contain expected string: "${val}".`)
                        } else if (kind === 'notContains' || e.kind === 'notContains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Output contained forbidden string: "${val}".`)
                        } else if (kind === 'firstLineEquals' || e.kind === 'firstLineEquals') {
                            const expected = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`First line mismatch. Expected: "${expected}"`)
                        } else if (kind === 'lastLineEquals' || e.kind === 'lastLineEquals') {
                            const expected = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Last line mismatch. Expected: "${expected}"`)
                        } else if (kind === 'regexMatch' || e.kind === 'regexMatch') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Output did not match regex pattern: "${val}".`)
                        } else if (kind === 'statusCode' || e.kind === 'statusCode') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '200')
                            lines.push(`Expected status code ${val}; received ${s.responseStatus ?? 'none'}.`)
                        } else if (kind === 'bodyContains' || e.kind === 'bodyContains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Body did not contain expected string: "${val}".`)
                        } else if (kind === 'bodyNotContains' || e.kind === 'bodyNotContains') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`Body contained forbidden string: "${val}".`)
                        } else if (kind === 'jsonPathMatch' || e.kind === 'jsonPathMatch') {
                            const path = d.expect_path || e.path || ''
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '')
                            lines.push(`JSONPath "${path}" did not match expected value: "${val}".`)
                        } else if (kind === 'latency' || e.kind === 'latency') {
                            const val = d.expect_expected != null ? String(d.expect_expected) : (e.value != null ? String(e.value) : '1000')
                            lines.push(`Expected latency <= ${val}ms; received ${s.latencyMs ?? 0}ms.`)
                        } else if (errorCode === 'sql_error' && errorMessage) {
                            lines.push(String(errorMessage))
                        } else if (s.expectMessage) {
                            lines.push(String(s.expectMessage))
                        } else if (errorMessage) {
                            lines.push(String(errorMessage))
                        }
                        return lines
                    }

                    return (
                        <Card key={s.id} variant="outlined" sx={{borderLeft: 4, borderLeftColor: status === 'error' ? 'error.main' : status === 'success' ? 'success.main' : 'divider'}}>
                            <CardContent sx={{py: 1.5}}>
                                {/* Header line: #, Name, Status, Started, Finished */}
                                <Stack
                                    direction={{xs: 'column', md: 'row'}}
                                    spacing={1}
                                    sx={{
                                        alignItems: {xs: 'flex-start', md: 'center'},
                                        justifyContent: "space-between"
                                    }}>
                                    <Stack direction="row" spacing={2} sx={{
                                        alignItems: "center"
                                    }}>
                                        <Typography variant="body2" sx={{minWidth: 24}}>#{s.stepOrder}</Typography>
                                        <Typography variant="subtitle2">{s.stepName}</Typography>
                                        <RunStatusChip status={s.status || 'unknown'}/>
                                    </Stack>
                                    <Stack direction="row" spacing={2}>
                                        <Typography variant="caption" sx={{
                                            color: "text.secondary"
                                        }}>Started: {formatDateTime(s.startedAt)}</Typography>
                                        <Typography variant="caption" sx={{
                                            color: "text.secondary"
                                        }}>Finished: {formatDateTime(s.finishedAt)}</Typography>
                                    </Stack>
                                </Stack>
                                {/* Details lines below header, success/error only */}
                                {showDetails && (
                                    <Box sx={{mt: 1}}>
                                        {status === 'success' && renderSuccessLines().map((t, i) => (
                                            <Typography key={i} variant="body2" sx={{color: 'success.main', whiteSpace: 'pre-wrap'}}>{t}</Typography>
                                        ))}
                                        {status === 'error' && renderErrorLines().map((t, i) => (
                                            <Typography key={i} variant="body2" sx={{color: 'error.main', whiteSpace: 'pre-wrap'}}>{t}</Typography>
                                        ))}
                                        {(rowsCount != null || rowsAffected != null) && (
                                            <Typography
                                                variant="caption"
                                                sx={{
                                                    color: "text.secondary",
                                                    display: 'block',
                                                    mt: 0.5
                                                }}>
                                                {rowsCount != null ? `rows: ${rowsCount}` : ''}{rowsAffected != null ? `${rowsCount != null ? ' • ' : ''}affected: ${rowsAffected}` : ''}
                                            </Typography>
                                        )}
                                        {s.exitCode != null && (
                                            <Typography
                                                variant="caption"
                                                sx={{
                                                    color: "text.secondary",
                                                    display: 'block',
                                                    mt: 0.5
                                                }}>
                                                exit code: {s.exitCode}
                                            </Typography>
                                        )}

                                        {(s.sqlText || s.commandText || s.scriptText) && (
                                            <Box sx={{mt: 1.5}}>
                                                <Typography
                                                    variant="caption"
                                                    sx={{
                                                        fontWeight: 600,
                                                        color: 'text.secondary',
                                                        display: 'block',
                                                        mb: 0.5
                                                    }}>
                                                    {s.sqlText ? 'SQL' : (s.commandText ? 'Command' : 'Script')}
                                                </Typography>
                                                <Box sx={{
                                                    p: 1,
                                                    backgroundColor: 'grey.900',
                                                    color: 'primary.light',
                                                    fontFamily: 'monospace',
                                                    fontSize: '0.75rem',
                                                    borderRadius: 1,
                                                    overflowX: 'auto',
                                                    whiteSpace: 'pre-wrap',
                                                    maxHeight: 200,
                                                    overflowY: 'auto'
                                                }}>
                                                    {s.sqlText || s.commandText || s.scriptText}
                                                </Box>
                                            </Box>
                                        )}

                                        {toRecordArray(d.executed_args).length > 0 && (
                                            <Box sx={{mt: 1.5}}>
                                                <Typography
                                                    variant="caption"
                                                    sx={{
                                                        fontWeight: 600,
                                                        color: 'text.secondary',
                                                        display: 'block',
                                                        mb: 0.5
                                                    }}>
                                                    SQL Arguments
                                                </Typography>
                                                <Box sx={{
                                                    p: 1,
                                                    backgroundColor: 'grey.900',
                                                    color: 'info.light',
                                                    fontFamily: 'monospace',
                                                    fontSize: '0.75rem',
                                                    borderRadius: 1,
                                                    overflowX: 'auto'
                                                }}>
                                                    {toRecordArray(d.executed_args).map((arg, idx: number) => (
                                                        <div key={idx}>${idx + 1}: {JSON.stringify(arg)}</div>
                                                    ))}
                                                </Box>
                                            </Box>
                                        )}

                                        {toRecordArray(d.result_lines).length > 0 && (
                                            <Box sx={{mt: 1.5}}>
                                                <Typography
                                                    variant="caption"
                                                    sx={{
                                                        fontWeight: 600,
                                                        color: 'text.secondary',
                                                        display: 'block',
                                                        mb: 0.5
                                                    }}>
                                                    Result Data
                                                </Typography>
                                                <TableContainer component={Paper} variant="outlined" sx={{maxHeight: 400}}>
                                                    <Table size="small" stickyHeader>
                                                        <TableHead>
                                                            <TableRow>
                                                                {Object.keys(toRecordArray(d.result_lines)[0]).map(k => (
                                                                    <TableCell key={k} sx={{fontWeight: 'bold', bgcolor: 'background.paper'}}>{k}</TableCell>
                                                                ))}
                                                            </TableRow>
                                                        </TableHead>
                                                        <TableBody>
                                                            {toRecordArray(d.result_lines).map((row, i: number) => (
                                                                <TableRow key={i}>
                                                                    {Object.values(row).map((v, j: number) => (
                                                                        <TableCell key={j} sx={{whiteSpace: 'nowrap'}}>{String(v)}</TableCell>
                                                                    ))}
                                                                </TableRow>
                                                            ))}
                                                        </TableBody>
                                                    </Table>
                                                </TableContainer>
                                            </Box>
                                        )}

                                        {(s.requestUrl || s.requestMethod) && (
                                            <Box sx={{mt: 1.5}}>
                                                <Typography
                                                    variant="caption"
                                                    sx={{
                                                        fontWeight: 600,
                                                        color: 'text.secondary',
                                                        display: 'block',
                                                        mb: 0.5
                                                    }}>
                                                    WEB TASK: {s.requestMethod} {s.requestUrl}
                                                </Typography>
                                                {s.requestHeaders && Object.keys(s.requestHeaders).length > 0 && (
                                                    <Box sx={{mb: 1}}>
                                                        <Typography variant="caption" sx={{color: 'text.secondary', display: 'block', mb: 0.5}}>Request Headers</Typography>
                                                        <Box sx={{p: 1, backgroundColor: 'grey.900', color: 'common.white', fontFamily: 'monospace', fontSize: '0.75rem', borderRadius: 1, overflowX: 'auto'}}>
                                                            {JSON.stringify(s.requestHeaders, null, 2)}
                                                        </Box>
                                                    </Box>
                                                )}
                                                {s.requestBody && (
                                                    <Box sx={{mb: 1}}>
                                                        <Typography variant="caption" sx={{color: 'text.secondary', display: 'block', mb: 0.5}}>Request Body</Typography>
                                                        <Box sx={{p: 1, backgroundColor: 'grey.900', color: 'common.white', fontFamily: 'monospace', fontSize: '0.75rem', borderRadius: 1, overflowX: 'auto', whiteSpace: 'pre-wrap', maxHeight: 150, overflowY: 'auto'}}>
                                                            {s.requestBody}
                                                        </Box>
                                                    </Box>
                                                )}
                                                {s.responseStatus && (
                                                    <Box sx={{mb: 1}}>
                                                        <Typography variant="caption" sx={{color: 'text.secondary', display: 'block', mb: 0.5}}>
                                                            Response Status: {s.responseStatus} • Latency: {s.latencyMs}ms
                                                        </Typography>
                                                    </Box>
                                                )}
                                                {s.responseHeaders && Object.keys(s.responseHeaders).length > 0 && (
                                                    <Box sx={{mb: 1}}>
                                                        <Typography variant="caption" sx={{color: 'text.secondary', display: 'block', mb: 0.5}}>Response Headers</Typography>
                                                        <Box sx={{p: 1, backgroundColor: 'grey.900', color: 'common.white', fontFamily: 'monospace', fontSize: '0.75rem', borderRadius: 1, overflowX: 'auto'}}>
                                                            {JSON.stringify(s.responseHeaders, null, 2)}
                                                        </Box>
                                                    </Box>
                                                )}
                                                {s.responseBody && (
                                                    <Box>
                                                        <Typography variant="caption" sx={{color: 'text.secondary', display: 'block', mb: 0.5}}>Response Body</Typography>
                                                        <Box sx={{p: 1, backgroundColor: 'grey.900', color: 'common.white', fontFamily: 'monospace', fontSize: '0.75rem', borderRadius: 1, overflowX: 'auto', whiteSpace: 'pre-wrap', maxHeight: 300, overflowY: 'auto'}}>
                                                            {s.responseBody}
                                                        </Box>
                                                    </Box>
                                                )}
                                            </Box>
                                        )}

                                        {(s.stdoutText || s.stderrText) && (
                                            <Box sx={{mt: 1.5}}>
                                                {s.stdoutText && (
                                                    <Box sx={{mb: 1}}>
                                                        <Typography
                                                            variant="caption"
                                                            sx={{
                                                                fontWeight: 600,
                                                                color: 'text.secondary',
                                                                display: 'block',
                                                                mb: 0.5
                                                            }}>STDOUT {s.stdoutTruncated && '(Truncated)'}</Typography>
                                                        <Box sx={{
                                                            p: 1,
                                                            backgroundColor: 'grey.900',
                                                            color: 'common.white',
                                                            fontFamily: 'monospace',
                                                            fontSize: '0.75rem',
                                                            borderRadius: 1,
                                                            overflowX: 'auto',
                                                            whiteSpace: 'pre-wrap',
                                                            maxHeight: 200,
                                                            overflowY: 'auto'
                                                        }}>
                                                            {s.stdoutText}
                                                        </Box>
                                                    </Box>
                                                )}
                                                {s.stderrText && (
                                                    <Box>
                                                        <Typography
                                                            variant="caption"
                                                            sx={{
                                                                fontWeight: 600,
                                                                color: 'error.main',
                                                                display: 'block',
                                                                mb: 0.5
                                                            }}>STDERR {s.stderrTruncated && '(Truncated)'}</Typography>
                                                        <Box sx={{
                                                            p: 1,
                                                            backgroundColor: 'grey.900',
                                                            color: 'error.light',
                                                            fontFamily: 'monospace',
                                                            fontSize: '0.75rem',
                                                            borderRadius: 1,
                                                            overflowX: 'auto',
                                                            whiteSpace: 'pre-wrap',
                                                            maxHeight: 200,
                                                            overflowY: 'auto'
                                                        }}>
                                                            {s.stderrText}
                                                        </Box>
                                                    </Box>
                                                )}
                                            </Box>
                                        )}
                                    </Box>
                                )}
                            </CardContent>
                        </Card>
                    );
                })}
                {steps.length === 0 && (
                    <Card variant="outlined"><CardContent><Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>No steps found.</Typography></CardContent></Card>
                )}
            </Stack>
            {/* Events (live) */}
            {progress.messages.length > 0 && (
                <>
                    <Typography variant="h6" gutterBottom>Events</Typography>
                    <Card variant="outlined">
                        <CardContent>
                            <List dense sx={{maxHeight: 320, overflowY: 'auto'}}>
                                {progress.messages.map((event, idx: number) => (
                                    <ListItem key={idx} sx={{py: 0}}>
                                        <ListItemText
                                            primary={<>
                                                <Typography
                                                    component="span"
                                                    variant="body2"
                                                    sx={{
                                                        color: "text.secondary",
                                                        mr: 1
                                                    }}>{formatDateTime(event.ts.toISOString())}</Typography>
                                                <Chip size="small" label={event.type} sx={{mr: 1}}/>
                                                {event.stepIndex != null && <Chip size="small" variant="outlined" label={`step ${event.stepIndex}`} sx={{mr: 1}}/>}
                                                {event.stepName && <Typography component="span" variant="body2" sx={{mr: 1}}>{event.stepName}</Typography>}
                                                {event.message && <Typography component="span" variant="body2">{event.message}</Typography>}
                                            </>}
                                        />
                                    </ListItem>
                                ))}
                            </List>
                        </CardContent>
                    </Card>
                </>
            )}
        </Box>
    );
}
