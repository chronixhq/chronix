import {useEffect, useMemo, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, FormControl, FormControlLabel, FormHelperText, InputLabel, MenuItem, Select, Snackbar, Stack, Switch, TextField, Typography} from '@mui/material';
import {type Action} from '../Actions/types';
import {type DbConnection} from '../Connections/types.ts';
import {apiPost} from '@dsherwin/react-api-interface';
import {useNavigate} from 'react-router';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {useJobs} from '../../data/JobsContext.tsx';
import {useSettings} from '../../data/SettingsContext';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {extractVarsFromAction, fetchJobEditorResources} from './editorUtils';
import {JobVariablesEditor} from './JobVariablesEditor';
import {JobScheduleEditor} from './jobScheduleEditor';
import {
    buildJobScheduleFromEditorState,
    createDefaultJobScheduleEditorState,
    type JobScheduleEditorState,
    summarizeJobScheduleEditorState,
    validateJobScheduleEditorState,
} from './jobScheduleState';

type CreateJobErrors = {
    name?: string;
    connectionId?: string;
    actionId?: string;
    singleRunAt?: string;
    recStart?: string;
    cronStr?: string;
};

export const CreateJob = () => {
    const navigate = useNavigate();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const jobLimit = checkLimit('jobs');
    const {timeZone} = useSettings();
    const {reload} = useJobs();

    useEffect(() => {
        if (!jobLimit.allowed) {
            navigate('/jobs/list');
        }
    }, [jobLimit.allowed, navigate]);

    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [notes, setNotes] = useState('');
    const [enabled, setEnabled] = useState<boolean>(true);
    const [alertEmails, setAlertEmails] = useState('');
    const [alertPhones, setAlertPhones] = useState('');
    const [notifyOnSuccess, setNotifyOnSuccess] = useState<boolean>(false);
    const [notifyOnError, setNotifyOnError] = useState<boolean>(true);
    const [notifyIncludeOutput, setNotifyIncludeOutput] = useState<boolean>(false);

    const [connections, setConnections] = useState<DbConnection[]>([]);
    const [shellConnections, setShellConnections] = useState<any[]>([]);
    const [webtaskConnections, setWebtaskConnections] = useState<any[]>([]);
    const [actions, setActions] = useState<Action[]>([]);
    const [targetKind, setTargetKind] = useState<'database' | 'shell' | 'webtask'>('database');
    const [connectionId, setConnectionId] = useState<string>('');
    const [shellConnectionId, setShellConnectionId] = useState<string>('');
    const [webtaskConnectionId, setWebtaskConnectionId] = useState<string>('');
    const [actionId, setActionId] = useState<string>('');

    const [scheduleState, setScheduleState] = useState<JobScheduleEditorState>(() => createDefaultJobScheduleEditorState());
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    const [loading, setLoading] = useState<boolean>(false);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [errors, setErrors] = useState<CreateJobErrors>({});

    const selectedAction = useMemo(() => actions.find((action) => action.id === actionId), [actions, actionId]);
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

    useEffect(() => {
        const load = async () => {
            setLoading(true);
            setLoadError(null);
            try {
                const {connections: connectionOptions, actions: actionOptions} = await fetchJobEditorResources(targetKind);
                if (targetKind === 'database') {
                    const databaseConnections = connectionOptions as DbConnection[];
                    const actionItems = actionOptions as Action[];
                    setConnections(databaseConnections);
                    setActions(actionItems);
                    if (!connectionId && databaseConnections[0]) setConnectionId(databaseConnections[0].id);
                    if (!actionId && actionItems[0]) setActionId(actionItems[0].id);
                } else if (targetKind === 'shell') {
                    const shellItems = connectionOptions as any[];
                    const actionItems = actionOptions as Action[];
                    setShellConnections(shellItems);
                    setActions(actionItems);
                    if (!shellConnectionId && shellItems[0]) setShellConnectionId(shellItems[0].id);
                    if (!actionId && actionItems[0]) setActionId(actionItems[0].id);
                } else {
                    const webtaskItems = connectionOptions as any[];
                    const actionItems = actionOptions as Action[];
                    setWebtaskConnections(webtaskItems);
                    setActions(actionItems);
                    if (!webtaskConnectionId && webtaskItems[0]) setWebtaskConnectionId(webtaskItems[0].id);
                    if (!actionId && actionItems[0]) setActionId(actionItems[0].id);
                }
            } catch (error) {
                console.error(error);
                setLoadError('Failed to load actions/connections');
            } finally {
                setLoading(false);
            }
        };

        void load();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [targetKind]);

    const scheduleSummary = useMemo(
        () => summarizeJobScheduleEditorState(scheduleState, timeZone),
        [scheduleState, timeZone],
    );

    const updateScheduleState = (next: JobScheduleEditorState) => {
        setScheduleState(next);
        setErrors((prev) => ({...prev, singleRunAt: undefined, recStart: undefined, cronStr: undefined}));
    };

    const onSave = async () => {
        const nextErrors: CreateJobErrors = {
            ...validateJobScheduleEditorState(scheduleState),
        };

        if (!name.trim()) nextErrors.name = 'Name is required.';
        if (targetKind === 'database' && !connectionId) nextErrors.connectionId = 'Select a database connection.';
        if (targetKind === 'shell' && !shellConnectionId) nextErrors.connectionId = 'Select a shell connection.';
        if (targetKind === 'webtask' && !webtaskConnectionId) nextErrors.connectionId = 'Select a webtask connection.';
        if (!actionId) nextErrors.actionId = 'Select an action.';

        setErrors(nextErrors);
        if (Object.values(nextErrors).some(Boolean)) {
            setSnack({open: true, message: 'Please fix validation issues.', severity: 'error'});
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
            const response: any = await apiPost('/jobs' as any, payload);
            if (response?.ok === false) throw new Error('Create failed');
            setSnack({open: true, message: 'Job created.', severity: 'success'});
            void reloadFeatureAvailability();
            await reload();
            window.setTimeout(() => navigate('/jobs/list'), 500);
        } catch (error) {
            console.error(error);
            setSnack({open: true, message: 'Failed to create job.', severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Typography variant="h5">Create Scheduled Job</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Button variant="contained" onClick={onSave}>Save Job</Button>
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
                                                setErrors((prev) => ({...prev, connectionId: undefined, actionId: undefined}));
                                            }}
                                        >
                                            <MenuItem value="database">Database</MenuItem>
                                            <MenuItem value="shell">Shell</MenuItem>
                                            <MenuItem value="webtask">Web Task</MenuItem>
                                        </Select>
                                    </FormControl>
                                    <TextField
                                        label="Job name"
                                        value={name}
                                        onChange={(event) => {
                                            setName(event.target.value);
                                            if (errors.name) setErrors((prev) => ({...prev, name: undefined}));
                                        }}
                                        required
                                        fullWidth
                                        error={!!errors.name}
                                        helperText={errors.name}
                                    />
                                </Stack>

                                <Stack direction={{xs: 'column', md: 'row'}} spacing={2}>
                                    {targetKind === 'database' ? (
                                        <FormControl fullWidth error={!!errors.connectionId}>
                                            <InputLabel id="conn-label">Database connection</InputLabel>
                                            <Select
                                                labelId="conn-label"
                                                label="Database connection"
                                                value={connectionId}
                                                onChange={(event) => {
                                                    setConnectionId(event.target.value as string);
                                                    if (errors.connectionId) setErrors((prev) => ({...prev, connectionId: undefined}));
                                                }}
                                            >
                                                {connections.map((connection) => (
                                                    <MenuItem key={connection.id} value={connection.id}>
                                                        {connection.name} {connection.driver ? `(${connection.driver})` : ''}
                                                    </MenuItem>
                                                ))}
                                            </Select>
                                            {errors.connectionId && <FormHelperText>{errors.connectionId}</FormHelperText>}
                                        </FormControl>
                                    ) : targetKind === 'shell' ? (
                                        <FormControl fullWidth error={!!errors.connectionId}>
                                            <InputLabel id="sconn-label">Shell connection</InputLabel>
                                            <Select
                                                labelId="sconn-label"
                                                label="Shell connection"
                                                value={shellConnectionId}
                                                onChange={(event) => {
                                                    setShellConnectionId(event.target.value as string);
                                                    if (errors.connectionId) setErrors((prev) => ({...prev, connectionId: undefined}));
                                                }}
                                            >
                                                {shellConnections.map((connection) => (
                                                    <MenuItem key={connection.id} value={connection.id}>
                                                        {connection.name} {connection.mode ? `(${connection.mode})` : ''}
                                                    </MenuItem>
                                                ))}
                                            </Select>
                                            {errors.connectionId && <FormHelperText>{errors.connectionId}</FormHelperText>}
                                        </FormControl>
                                    ) : (
                                        <FormControl fullWidth error={!!errors.connectionId}>
                                            <InputLabel id="wtconn-label">Web Task connection</InputLabel>
                                            <Select
                                                labelId="wtconn-label"
                                                label="Web Task connection"
                                                value={webtaskConnectionId}
                                                onChange={(event) => {
                                                    setWebtaskConnectionId(event.target.value as string);
                                                    if (errors.connectionId) setErrors((prev) => ({...prev, connectionId: undefined}));
                                                }}
                                            >
                                                {webtaskConnections.map((connection) => (
                                                    <MenuItem key={connection.id} value={connection.id}>
                                                        {connection.name}
                                                    </MenuItem>
                                                ))}
                                            </Select>
                                            {errors.connectionId && <FormHelperText>{errors.connectionId}</FormHelperText>}
                                        </FormControl>
                                    )}

                                    <FormControl fullWidth error={!!errors.actionId}>
                                        <InputLabel id="action-label">Action</InputLabel>
                                        <Select
                                            labelId="action-label"
                                            label="Action"
                                            value={actionId}
                                            onChange={(event) => {
                                                setActionId(event.target.value as string);
                                                if (errors.actionId) setErrors((prev) => ({...prev, actionId: undefined}));
                                            }}
                                        >
                                            {actions.map((action) => (
                                                <MenuItem key={action.id} value={action.id}>
                                                    {action.name}
                                                </MenuItem>
                                            ))}
                                        </Select>
                                        {errors.actionId && <FormHelperText>{errors.actionId}</FormHelperText>}
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

                                <HStack>
                                    <FormControlLabel control={<Switch checked={enabled} onChange={(event) => setEnabled(event.target.checked)}/>} label="Enabled"/>
                                </HStack>

                                <JobScheduleEditor
                                    state={scheduleState}
                                    onChange={updateScheduleState}
                                    errors={{
                                        singleRunAt: errors.singleRunAt,
                                        recStart: errors.recStart,
                                        cronStr: errors.cronStr,
                                    }}
                                />

                                <Box sx={{mt: -1}}>
                                    {scheduleSummary ? (
                                        <Alert severity="info" variant="outlined" sx={{py: 0}}>{scheduleSummary}</Alert>
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
                    <CardActions sx={{justifyContent: 'space-between', flexWrap: 'wrap'}}>
                        <Typography variant="caption" sx={{color: 'text.secondary', ml: 1}}>
                            Jobs use exactly one Action and one Connection in v1.
                        </Typography>
                        <Button variant="contained" onClick={onSave}>Save Job</Button>
                    </CardActions>
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
