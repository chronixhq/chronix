import {useEffect, useMemo, useRef, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, FormControl, FormControlLabel, InputLabel, Link, MenuItem, Select, Snackbar, Stack, Switch, TextField, Typography} from '@mui/material';
import {formatDateTime, formatDateTimeHM} from '../../lib/utilities';
import {useSettings} from '../../data/SettingsContext';
import {useLocation, useNavigate, useParams} from 'react-router';
import {type Job} from './types';
import {type Action} from '../Actions/types';
import {type DbConnection} from '../Connections/types.ts';
import {apiPut} from '@dsherwin/react-api-interface';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {confirmOnNavigate, useUnsavedChanges} from '../../lib/useUnsavedChanges.ts';
import {useRunsContext} from '../../data/RunsContext';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {useRunNow} from '../../data/RunNowContext';
import {extractVarsFromAction, fetchJobEditorResources} from './editorUtils';
import {fetchJobById} from './api.ts';
import {JobVariablesEditor} from './JobVariablesEditor';
import {JobScheduleEditor} from './jobScheduleEditor';
import {
    buildJobScheduleFromEditorState,
    createDefaultJobScheduleEditorState,
    jobScheduleEditorStateFromSchedule,
    type JobScheduleEditorState,
    summarizeJobScheduleEditorState,
} from './jobScheduleState';

function useQuery() {
    const {search} = useLocation();
    return useMemo(() => new URLSearchParams(search), [search]);
}

export const EditJob = () => {
    const navigate = useNavigate();
    const {reload: reloadFeatureAvailability} = useFeatureAvailability();
    const query = useQuery();
    const {timeZone} = useSettings();
    const {runNow} = useRunNow();
    const {confirmPrompt} = useMuiPrompts();

    const [loading, setLoading] = useState<boolean>(true);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});

    const [connections, setConnections] = useState<DbConnection[]>([]);
    const [shellConnections, setShellConnections] = useState<any[]>([]);
    const [webtaskConnections, setWebtaskConnections] = useState<any[]>([]);
    const [actions, setActions] = useState<Action[]>([]);

    const [id, setId] = useState<string>('');
    const [name, setName] = useState<string>('');
    const [description, setDescription] = useState<string>('');
    const [notes, setNotes] = useState<string>('');
    const [targetKind, setTargetKind] = useState<'database' | 'shell' | 'webtask'>('database');
    const [connectionId, setConnectionId] = useState<string>('');
    const [shellConnectionId, setShellConnectionId] = useState<string>('');
    const [webtaskConnectionId, setWebtaskConnectionId] = useState<string>('');
    const [actionId, setActionId] = useState<string>('');
    const [enabled, setEnabled] = useState<boolean>(true);
    const [alertEmails, setAlertEmails] = useState<string>('');
    const [alertPhones, setAlertPhones] = useState<string>('');
    const [notifyOnSuccess, setNotifyOnSuccess] = useState<boolean>(false);
    const [notifyOnError, setNotifyOnError] = useState<boolean>(true);
    const [notifyIncludeOutput, setNotifyIncludeOutput] = useState<boolean>(false);
    const [nextRunAt, setNextRunAt] = useState<string | null>(null);
    const [lastRunAt, setLastRunAt] = useState<string | null>(null);

    const [scheduleState, setScheduleState] = useState<JobScheduleEditorState>(() => createDefaultJobScheduleEditorState());
    const [dirty, setDirty] = useState<boolean>(false);
    const baselineRef = useRef<string>('');

    const {useRecentRunsForJob} = useRunsContext();
    const {items: runs, reload: reloadRuns} = useRecentRunsForJob(id, 20);

    const selectedAction = useMemo(() => actions.find((action) => String(action.id) === String(actionId)), [actions, actionId]);
    const detectedVars = useMemo(() => {
        if (!selectedAction) return [];
        return extractVarsFromAction(selectedAction);
    }, [selectedAction]);
    const [varValues, setVarValues] = useState<Record<string, string>>({});

    useEffect(() => {
        setVarValues((prev) => {
            const next: Record<string, string> = {};
            detectedVars.forEach((variable) => {
                next[variable] = prev[variable] ?? '';
            });
            return next;
        });
    }, [detectedVars]);

    const params = useParams();
    useEffect(() => {
        const selectedId = (params as any)?.id || query.get('id') || '';

        const load = async () => {
            setLoading(true);
            setLoadError(null);
            try {
                const job: Job | undefined = selectedId ? await fetchJobById(selectedId) : undefined;
                if (!job) throw new Error('Job not found');
                if ((job as any).suspended) {
                    navigate('/jobs/list');
                    return;
                }

                const jobAny: any = job;
                const nextTargetKind = jobAny.targetKind || jobAny.target_kind || 'database';
                const nextScheduleState = jobScheduleEditorStateFromSchedule(jobAny.schedule, timeZone);
                const nextVariables: Record<string, string> = {};
                (jobAny.variables || []).forEach((variable: any) => {
                    if (variable?.name) nextVariables[variable.name] = variable.value ?? '';
                });

                const {connections: connectionOptions, actions: actionOptions} = await fetchJobEditorResources(nextTargetKind);
                if (nextTargetKind === 'database') {
                    setConnections(connectionOptions as DbConnection[]);
                } else if (nextTargetKind === 'shell') {
                    setShellConnections(connectionOptions as any[]);
                } else {
                    setWebtaskConnections(connectionOptions as any[]);
                }
                setActions(actionOptions as Action[]);

                setId(String(jobAny.id));
                setName(jobAny.name);
                setDescription(jobAny.description || '');
                setNotes(jobAny.notes || '');
                setTargetKind(nextTargetKind);
                setConnectionId(String(jobAny.connectionId ?? jobAny.connection_id ?? ''));
                setShellConnectionId(String(jobAny.shellConnectionId ?? jobAny.shell_connection_id ?? ''));
                setWebtaskConnectionId(String(jobAny.webtaskConnectionId ?? jobAny.webtask_connection_id ?? ''));
                setActionId(String(jobAny.actionId ?? jobAny.action_id ?? ''));
                setEnabled(jobAny.enabled !== false);
                setAlertEmails(jobAny.alertEmails ?? jobAny.alert_emails ?? '');
                setAlertPhones(jobAny.alertPhones ?? jobAny.alert_phones ?? '');
                setNotifyOnSuccess(jobAny.notifyOnSuccess ?? !!jobAny.notify_on_success);
                setNotifyOnError(jobAny.notifyOnError ?? jobAny.notify_on_error !== false);
                setNotifyIncludeOutput(jobAny.notifyIncludeOutput ?? !!jobAny.notify_include_output);
                setNextRunAt(jobAny.nextRunAt ?? jobAny.next_run_at ?? null);
                setLastRunAt(jobAny.lastRunAt ?? jobAny.last_run_at ?? null);
                setVarValues(nextVariables);
                setScheduleState(nextScheduleState);

                baselineRef.current = JSON.stringify({
                    name: jobAny.name,
                    description: jobAny.description || '',
                    notes: jobAny.notes || '',
                    targetKind: nextTargetKind,
                    connectionId: String(jobAny.connectionId ?? jobAny.connection_id ?? ''),
                    shellConnectionId: String(jobAny.shellConnectionId ?? jobAny.shell_connection_id ?? ''),
                    webtaskConnectionId: String(jobAny.webtaskConnectionId ?? jobAny.webtask_connection_id ?? ''),
                    actionId: String(jobAny.actionId ?? jobAny.action_id ?? ''),
                    enabled: jobAny.enabled !== false,
                    alertEmails: jobAny.alertEmails ?? jobAny.alert_emails ?? '',
                    alertPhones: jobAny.alertPhones ?? jobAny.alert_phones ?? '',
                    notifyOnSuccess: jobAny.notifyOnSuccess ?? !!jobAny.notify_on_success,
                    notifyOnError: jobAny.notifyOnError ?? jobAny.notify_on_error !== false,
                    notifyIncludeOutput: jobAny.notifyIncludeOutput ?? !!jobAny.notify_include_output,
                    scheduleState: nextScheduleState,
                    varValues: nextVariables,
                });
                setDirty(false);
            } catch (error) {
                console.error(error);
                setLoadError('Failed to load job');
            } finally {
                setLoading(false);
            }
        };

        void load();
    }, [navigate, params, query, timeZone]);

    useEffect(() => {
        if (baselineRef.current === '') return;
        const current = JSON.stringify({
            name,
            description,
            notes,
            targetKind,
            connectionId,
            shellConnectionId,
            webtaskConnectionId,
            actionId,
            enabled,
            alertEmails,
            alertPhones,
            notifyOnSuccess,
            notifyOnError,
            notifyIncludeOutput,
            scheduleState,
            varValues,
        });
        setDirty(current !== baselineRef.current);
    }, [
        name,
        description,
        notes,
        targetKind,
        connectionId,
        shellConnectionId,
        webtaskConnectionId,
        actionId,
        enabled,
        alertEmails,
        alertPhones,
        notifyOnSuccess,
        notifyOnError,
        notifyIncludeOutput,
        scheduleState,
        varValues,
    ]);

    useUnsavedChanges(dirty);

    useEffect(() => {
        if (id) {
            void reloadRuns();
        }
    }, [id, reloadRuns]);

    const scheduleSummary = useMemo(
        () => summarizeJobScheduleEditorState(scheduleState, timeZone),
        [scheduleState, timeZone],
    );

    const onSave = async () => {
        if (!name.trim()) {
            setSnack({open: true, message: 'Name is required.', severity: 'error'});
            return;
        }
        if (targetKind === 'database' && !connectionId) {
            setSnack({open: true, message: 'Select a database connection.', severity: 'error'});
            return;
        }
        if (targetKind === 'shell' && !shellConnectionId) {
            setSnack({open: true, message: 'Select a shell connection.', severity: 'error'});
            return;
        }
        if (targetKind === 'webtask' && !webtaskConnectionId) {
            setSnack({open: true, message: 'Select a webtask connection.', severity: 'error'});
            return;
        }
        if (!actionId) {
            setSnack({open: true, message: 'Select an action.', severity: 'error'});
            return;
        }

        const schedule = buildJobScheduleFromEditorState(scheduleState, timeZone);
        if (!schedule) {
            setSnack({open: true, message: 'Please complete the schedule configuration.', severity: 'error'});
            return;
        }

        const variables = detectedVars.map((variable) => ({
            name: variable,
            value: (varValues[variable] ?? '').trim() || undefined,
        }));

        const payload = {
            name,
            description: description || undefined,
            notes: notes || undefined,
            target_kind: targetKind,
            connection_id: targetKind === 'database' ? Number(connectionId) : undefined,
            shell_connection_id: targetKind === 'shell' ? Number(shellConnectionId) : undefined,
            webtask_connection_id: targetKind === 'webtask' ? Number(webtaskConnectionId) : undefined,
            action_id: Number(actionId),
            schedule,
            enabled,
            variables,
            alert_emails: alertEmails || undefined,
            alert_phones: alertPhones || undefined,
            notify_on_success: notifyOnSuccess,
            notify_on_error: notifyOnError,
            notify_include_output: notifyIncludeOutput,
        } as const;

        try {
            const response: any = await apiPut(`/jobs/${encodeURIComponent(id)}` as any, payload as any);
            if (response?.ok === false) throw new Error('Update failed');
            void reloadFeatureAvailability();
            setSnack({open: true, message: 'Job updated.', severity: 'success'});
            navigate('/jobs/list', {state: {refresh: true}});
        } catch (error) {
            console.error(error);
            setSnack({open: true, message: 'Failed to update job.', severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Typography variant="h5">Edit Scheduled Job</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Button
                            variant="outlined"
                            onClick={async () => {
                                try {
                                    const runId = await runNow(id, {jobName: name});
                                    setSnack({open: true, message: `Run queued${runId ? ` (id: ${runId})` : ''}`, severity: 'info'});
                                    try {
                                        await reloadRuns();
                                    } catch {
                                        // best effort refresh
                                    }
                                } catch (error) {
                                    console.error(error);
                                    setSnack({open: true, message: 'Run failed', severity: 'error'});
                                }
                            }}
                        >
                            Run now
                        </Button>
                        <Button variant="outlined" onClick={() => confirmOnNavigate(dirty, navigate, confirmPrompt)('/jobs/list')}>Cancel</Button>
                        <Button variant="contained" onClick={onSave}>Save Changes</Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {loadError && <Alert severity="error">{loadError}</Alert>}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        {loading ? (
                            <Typography variant="body2" sx={{color: 'text.secondary'}}>Loading…</Typography>
                        ) : (
                            <VStack spacing={3}>
                                <TextField label="ID" value={id} disabled fullWidth/>

                                <Stack direction={{xs: 'column', md: 'row'}} spacing={2}>
                                    <FormControl sx={{minWidth: 160}}>
                                        <InputLabel id="tk-label">Target type</InputLabel>
                                        <Select
                                            labelId="tk-label"
                                            label="Target type"
                                            value={targetKind}
                                            onChange={(event) => {
                                                setTargetKind(event.target.value as typeof targetKind);
                                                setConnectionId('');
                                                setShellConnectionId('');
                                                setWebtaskConnectionId('');
                                                setActionId('');
                                            }}
                                        >
                                            <MenuItem value="database">Database</MenuItem>
                                            <MenuItem value="shell">Shell</MenuItem>
                                            <MenuItem value="webtask">Web Task</MenuItem>
                                        </Select>
                                    </FormControl>
                                    <TextField label="Job name" value={name} onChange={(event) => setName(event.target.value)} required fullWidth/>
                                </Stack>

                                <Stack direction={{xs: 'column', md: 'row'}} spacing={2}>
                                    {targetKind === 'database' ? (
                                        <FormControl fullWidth>
                                            <InputLabel id="conn-label">Database connection</InputLabel>
                                            <Select labelId="conn-label" label="Database connection" value={connectionId} onChange={(event) => setConnectionId(event.target.value as string)}>
                                                {connections.map((connection) => (
                                                    <MenuItem key={connection.id} value={connection.id}>
                                                        {connection.name} {connection.driver ? `(${connection.driver})` : ''}
                                                    </MenuItem>
                                                ))}
                                            </Select>
                                        </FormControl>
                                    ) : targetKind === 'shell' ? (
                                        <FormControl fullWidth>
                                            <InputLabel id="sconn-label">Shell connection</InputLabel>
                                            <Select labelId="sconn-label" label="Shell connection" value={shellConnectionId} onChange={(event) => setShellConnectionId(event.target.value as string)}>
                                                {shellConnections.map((connection) => (
                                                    <MenuItem key={connection.id} value={connection.id}>
                                                        {connection.name} {connection.mode ? `(${connection.mode})` : ''}
                                                    </MenuItem>
                                                ))}
                                            </Select>
                                        </FormControl>
                                    ) : (
                                        <FormControl fullWidth>
                                            <InputLabel id="wtconn-label">Web Task connection</InputLabel>
                                            <Select labelId="wtconn-label" label="Web Task connection" value={webtaskConnectionId} onChange={(event) => setWebtaskConnectionId(event.target.value as string)}>
                                                {webtaskConnections.map((connection) => (
                                                    <MenuItem key={connection.id} value={connection.id}>
                                                        {connection.name}
                                                    </MenuItem>
                                                ))}
                                            </Select>
                                        </FormControl>
                                    )}

                                    <FormControl fullWidth>
                                        <InputLabel id="action-label">Action</InputLabel>
                                        <Select labelId="action-label" label="Action" value={actionId} onChange={(event) => setActionId(event.target.value as string)}>
                                            {actions.map((action) => (
                                                <MenuItem key={action.id} value={action.id}>
                                                    {action.name}
                                                </MenuItem>
                                            ))}
                                        </Select>
                                    </FormControl>
                                </Stack>

                                <TextField label="Description (optional)" value={description} onChange={(event) => setDescription(event.target.value)} fullWidth multiline minRows={2}/>
                                <TextField label="Notes (optional)" value={notes} onChange={(event) => setNotes(event.target.value)} fullWidth multiline minRows={2}/>

                                <Divider/>
                                <Typography variant="h6">Alerts</Typography>
                                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                                    Configure specific destinations for alerts related to this job&apos;s execution.
                                </Typography>
                                <TextField
                                    label="Alert Emails"
                                    placeholder="email1@example.com, email2@example.com"
                                    value={alertEmails}
                                    onChange={(event) => setAlertEmails(event.target.value)}
                                    fullWidth
                                    helperText="Comma-separated list of email addresses. If empty, system defaults are used."
                                />
                                <TextField
                                    label="Alert Phones (SMS)"
                                    placeholder="+15550001111"
                                    value={alertPhones}
                                    onChange={(event) => setAlertPhones(event.target.value)}
                                    fullWidth
                                    helperText="Comma-separated list of E.164 phone numbers. If empty, system defaults are used."
                                />
                                <HStack spacing={4}>
                                    <FormControlLabel control={<Switch checked={notifyOnSuccess} onChange={(event) => setNotifyOnSuccess(event.target.checked)}/>} label="Notify on Success"/>
                                    <FormControlLabel control={<Switch checked={notifyOnError} onChange={(event) => setNotifyOnError(event.target.checked)}/>} label="Notify on Error"/>
                                    <FormControlLabel control={<Switch checked={notifyIncludeOutput} onChange={(event) => setNotifyIncludeOutput(event.target.checked)}/>} label="Include Output"/>
                                </HStack>

                                <Divider/>

                                <HStack sx={{alignItems: 'center', gap: 2, flexWrap: 'wrap'}}>
                                    <FormControlLabel control={<Switch checked={enabled} onChange={(event) => setEnabled(event.target.checked)}/>} label="Enabled"/>
                                    {lastRunAt && <Typography variant="body2" sx={{color: 'text.secondary'}}>Last run: {formatDateTimeHM(lastRunAt)}</Typography>}
                                    {nextRunAt && <Typography variant="body2" sx={{color: 'text.secondary'}}>Next run: {formatDateTimeHM(nextRunAt)}</Typography>}
                                </HStack>

                                <JobScheduleEditor state={scheduleState} onChange={setScheduleState}/>

                                <Box sx={{mt: -1}}>
                                    {scheduleSummary ? (
                                        <Alert severity="info" variant="standard" sx={{py: 0}}>{scheduleSummary}</Alert>
                                    ) : (
                                        <Typography variant="body2" sx={{color: 'text.secondary'}}>
                                            Complete schedule fields to see a summary.
                                        </Typography>
                                    )}
                                </Box>

                                <JobVariablesEditor
                                    detectedVars={detectedVars}
                                    varValues={varValues}
                                    onChange={(variable, value) => setVarValues((prev) => ({...prev, [variable]: value}))}
                                />
                            </VStack>
                        )}
                    </CardContent>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="outlined" onClick={() => confirmOnNavigate(dirty, navigate, confirmPrompt)('/jobs/list')}>Cancel</Button>
                        <Button variant="contained" onClick={onSave}>Save Changes</Button>
                    </CardActions>
                </Card>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <Typography variant="h6" gutterBottom>Recent runs</Typography>
                        {runs.length === 0 ? (
                            <Typography variant="body2" sx={{color: 'text.secondary'}}>No runs yet.</Typography>
                        ) : (
                            <VStack spacing={0.5}>
                                {runs.map((run: any, index: number) => (
                                    <Typography key={index} variant="body2">
                                        <Link
                                            href={`/runs/${encodeURIComponent(run.runId)}`}
                                            underline="hover"
                                            onClick={(event) => {
                                                event.preventDefault();
                                                navigate(`/runs/${encodeURIComponent(run.runId)}`);
                                            }}
                                        >
                                            #{run.runId}
                                        </Link>
                                        {` — ${run.status} — queued ${formatDateTime(run.queuedAt as any)}`}
                                        {run.finishedAt ? ` — finished ${formatDateTime(run.finishedAt as any)}` : ''}
                                    </Typography>
                                ))}
                            </VStack>
                        )}
                    </CardContent>
                </Card>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack((current) => ({...current, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack((current) => ({...current, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
