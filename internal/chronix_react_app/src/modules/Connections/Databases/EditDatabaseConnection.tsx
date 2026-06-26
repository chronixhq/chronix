import {useEffect, useMemo, useRef, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, Typography} from '@mui/material';
import {useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import ContentCopy from '@mui/icons-material/ContentCopy';
import {useNavigate, useParams} from 'react-router';
import {apiPost, apiPut} from '@dsherwin/react-api-interface';
import {useFeatureAvailability} from '../../../data/FeatureAvailabilityContext.tsx';
import {useConnections} from "../../../data/ConnectionsContext.tsx";
import {confirmOnNavigate, useUnsavedChanges} from '../../../lib/useUnsavedChanges.ts';
import {type DbConnection, type DbConnectionDraft, type DbDriver} from '../types.ts';
import {fetchAgentOptions, mergeSelectedAgent, type AgentOption} from "../../Agents/api.ts";
import {deleteStoredConnection, duplicateStoredConnection, fetchConnectionById} from '../api.ts';
import {DatabaseConnectionEditor} from './DatabaseConnectionEditor.tsx';
import {buildDatabaseConnectionDsn, buildDatabaseConnectionPreviewDsn, DEFAULT_DB_CONNECTIONS} from './databaseConnectionForm.ts';

const isEmpty = (value?: string | number | null) => value === undefined || value === null || String(value).trim() === '';

type ConnectionResponse = {
    ok?: boolean
    message?: string
    id?: string | number
}

export const EditDatabaseConnection = () => {
    const navigate = useNavigate();
    const {id} = useParams();
    const {reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {reload: reloadConnections} = useConnections();
    const {confirmPrompt} = useMuiPrompts();

    const [agents, setAgents] = useState<AgentOption[]>([]);
    const [agentsLoading, setAgentsLoading] = useState(false);
    const [initialAgent, setInitialAgent] = useState<{ uuid: string; name: string } | null>(null);
    const [loading, setLoading] = useState(true);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [showPassword, setShowPassword] = useState(false);
    const [draft, setDraft] = useState<DbConnectionDraft>({
        name: '',
        driver: 'postgres',
    });
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState<null | { ok: boolean; message: string }>(null);
    const [dirty, setDirty] = useState(false);
    const baselineRef = useRef('');
    const loadedDriverRef = useRef<DbDriver | null>(null);

    useEffect(() => {
        const load = async () => {
            setLoading(true);
            setLoadError(null);
            try {
                const data = await fetchConnectionById<DbConnection>('database', id || '');
                if (data.suspended) {
                    navigate('/databases/list');
                    return;
                }

                const nextDraft: DbConnectionDraft = {
                    name: data.name,
                    driver: data.driver,
                    description: data.description,
                    host: data.host,
                    port: data.port,
                    database: data.database,
                    username: data.username,
                    sslEnabled: data.sslEnabled,
                    sslMode: data.sslMode,
                    trustServerCertificate: data.trustServerCertificate,
                    extraParams: data.extraParams,
                    params: data.params,
                    filePath: data.filePath,
                    autoCheckEnabled: data.autoCheckEnabled,
                    autoCheckSeconds: data.autoCheckSeconds,
                    agentUuid: data.agentUuid || undefined,
                    hasPassword: data.hasPassword,
                    alertEmails: data.alertEmails || "",
                    alertPhones: data.alertPhones || "",
                    notifyOnFailure: !!data.notifyOnFailure,
                };

                setDraft(nextDraft);
                loadedDriverRef.current = nextDraft.driver;
                if (data.agentUuid) {
                    setInitialAgent({uuid: data.agentUuid, name: data.agentUuid});
                }
                baselineRef.current = JSON.stringify(nextDraft);
                setDirty(false);
            } catch (e) {
                console.error(e);
                setLoadError('Failed to load connection');
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [id, navigate]);

    useEffect(() => {
        let alive = true;
        (async () => {
            try {
                setAgentsLoading(true);
                const list = await fetchAgentOptions();
                if (!alive) return;
                setAgents(mergeSelectedAgent(list, initialAgent?.uuid, initialAgent?.name));
            } catch (e) {
                console.error(e);
                setAgents([]);
            } finally {
                setAgentsLoading(false);
            }
        })();
        return () => {
            alive = false;
        };
    }, [initialAgent]);

    useEffect(() => {
        if (loading || !loadedDriverRef.current || draft.driver === loadedDriverRef.current) return;
        setDraft((current) => ({
            ...DEFAULT_DB_CONNECTIONS[current.driver],
            ...current,
        }));
    }, [draft.driver, loading]);

    useEffect(() => {
        if (!baselineRef.current) return;
        setDirty(JSON.stringify(draft) !== baselineRef.current);
    }, [draft]);

    useUnsavedChanges(dirty);

    const validate = (): boolean => {
        const nextErrors: Record<string, string> = {};

        if (isEmpty(draft.name)) nextErrors.name = 'Give this connection a label.';
        if (draft.driver === 'sqlite') {
            if (isEmpty(draft.filePath)) nextErrors.filePath = 'Path to the database file is required.';
        } else {
            if (draft.driver === 'snowflake') {
                if (isEmpty(draft.host)) nextErrors.host = 'Account is required.';
            } else {
                if (isEmpty(draft.host)) nextErrors.host = 'Host is required.';
                if (isEmpty(draft.port)) nextErrors.port = 'Port is required.';
            }
            if (isEmpty(draft.database) && draft.driver !== 'mysql' && draft.driver !== 'snowflake') nextErrors.database = 'Database is required.';
            if (isEmpty(draft.username)) nextErrors.username = 'Username is required.';
        }

        setErrors(nextErrors);
        return Object.keys(nextErrors).length === 0;
    };

    const rawDsn = useMemo(() => buildDatabaseConnectionDsn(draft), [draft]);
    const previewDsn = useMemo(() => buildDatabaseConnectionPreviewDsn(draft), [draft]);

    const onDuplicate = async () => {
        try {
            setLoading(true);
            const data = await duplicateStoredConnection({kind: 'database', id: id || ''}) as ConnectionResponse;
            setTestResult({ok: true, message: 'Connection duplicated successfully'});
            await reloadFeatureAvailability();
            await reloadConnections();
            if (data?.id) {
                navigate(`/databases/edit/${data.id}`);
            } else {
                navigate('/databases/list');
            }
        } catch (e) {
            console.error(e);
            setTestResult({ok: false, message: 'Failed to duplicate connection'});
        } finally {
            setLoading(false);
        }
    };

    const onSave = async () => {
        if (!validate()) return;
        try {
            const payload: {
                name: string
                driver: DbDriver
                description?: string
                dsn: string
                autoCheckEnabled: boolean
                autoCheckSeconds: number
                alertEmails: string
                alertPhones: string
                notifyOnFailure: boolean
                agentUuid?: string
            } = {
                name: draft.name,
                driver: draft.driver,
                description: draft.description,
                dsn: rawDsn,
                autoCheckEnabled: !!draft.autoCheckEnabled,
                autoCheckSeconds: draft.autoCheckSeconds || 3600,
                alertEmails: draft.alertEmails || "",
                alertPhones: draft.alertPhones || "",
                notifyOnFailure: !!draft.notifyOnFailure,
            };
            if (draft.agentUuid) payload.agentUuid = draft.agentUuid;

            const response = await apiPut(`/connections/${encodeURIComponent(id || '')}`, payload) as ConnectionResponse;
            if (response?.ok === false) throw new Error('Update failed');
            void reloadFeatureAvailability();
            void reloadConnections();
            setTestResult({ok: true, message: 'Connection updated.'});
            navigate('/databases/list', {state: {refresh: true}});
        } catch (e) {
            console.error(e);
            setTestResult({ok: false, message: 'Failed to update connection.'});
        }
    };

    const onTest = async () => {
        if (!validate()) return;
        setTesting(true);
        setTestResult(null);
        try {
            const body: { id: number; driver: DbDriver; dsn: string; agentUuid?: string } = {
                id: Number(id),
                driver: draft.driver,
                dsn: rawDsn,
            };
            if (draft.agentUuid) body.agentUuid = draft.agentUuid;
            const response = await apiPost('/connections/test', body) as ConnectionResponse;
            setTestResult({ok: !!response?.ok, message: response?.message || (response?.ok ? 'Connection succeeded' : 'Connection failed')});
        } catch (e) {
            console.error(e);
            setTestResult({ok: false, message: 'Network error'});
        } finally {
            setTesting(false);
        }
    };

    const onDelete = async () => {
        const ok = await confirmPrompt({
            title: 'Delete Connection',
            message: `Are you sure you want to delete "${draft.name}"? This action cannot be undone.`,
            buttonText: 'Delete',
            cancelButtonText: 'Cancel'
        });
        if (!ok) return;

        try {
            await deleteStoredConnection({kind: 'database', id: id || ''});
            setTestResult({ok: true, message: 'Connection deleted'});
            void reloadConnections();
            navigate('/databases/list');
        } catch (e) {
            console.error(e);
            setTestResult({ok: false, message: 'Failed to delete connection'});
        }
    };

    if (loading) {
        return (
            <Box sx={{py: 4, display: 'flex', justifyContent: 'center'}}>
                <Typography variant="body2" sx={{color: "text.secondary"}}>Loading…</Typography>
            </Box>
        );
    }

    if (loadError) {
        return (
            <Box sx={{px: {xs: 1, md: 2}, py: 2}}>
                <Alert severity="error">{loadError}</Alert>
            </Box>
        );
    }

    return (
        <DatabaseConnectionEditor
            title="Edit database connection"
            infoTooltip="Used by jobs and explorers."
            draft={draft}
            setDraft={setDraft}
            errors={errors}
            agents={agents}
            agentsLoading={agentsLoading}
            showPassword={showPassword}
            setShowPassword={setShowPassword}
            testing={testing}
            loading={loading}
            testResult={testResult}
            onTest={onTest}
            onCancel={() => confirmOnNavigate(dirty, navigate, confirmPrompt)('/databases/list')}
            onSave={onSave}
            saveLabel="Save Changes"
            previewDsn={previewDsn}
            headerAction={(
                <Button
                    variant="outlined"
                    startIcon={<ContentCopy/>}
                    onClick={onDuplicate}
                    disabled={loading}
                >
                    Duplicate Connection
                </Button>
            )}
            dangerZone={(
                <Card variant="outlined" sx={{borderRadius: 3, border: '1px solid', borderColor: 'error.main', mt: 4}}>
                    <CardActions sx={{p: 2, justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2}}>
                        <VStack spacing={0.5} alignItems="flex-start">
                            <Typography variant="subtitle1" sx={{fontWeight: "bold", color: "error.main"}}>Danger Zone</Typography>
                            <Typography variant="body2" sx={{color: "text.secondary"}}>
                                Once you delete a connection, there is no going back. Please be certain.
                            </Typography>
                        </VStack>
                        <Button variant="outlined" color="error" onClick={onDelete} disabled={loading}>
                            Delete Connection
                        </Button>
                    </CardActions>
                </Card>
            )}
        />
    );
}
