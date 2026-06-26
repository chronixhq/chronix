import {useEffect, useRef, useState} from "react";
import {Box, Button, Card, CardActions, CircularProgress, Typography} from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import {apiPost, apiPut} from "@dsherwin/react-api-interface";
import {useNavigate, useParams} from 'react-router';
import {useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {useFeatureAvailability} from "../../../data/FeatureAvailabilityContext.tsx";
import {confirmOnNavigate, useUnsavedChanges} from "../../../lib/useUnsavedChanges.ts";
import {useConnections} from "../../../data/ConnectionsContext.tsx";
import {fetchAgentOptions, mergeSelectedAgent, type AgentOption} from "../../Agents/api.ts";
import {
    deleteStoredConnection,
    duplicateStoredConnection,
    fetchConnectionById,
    getConnectionApiItemPath,
    getConnectionEditPath,
    getConnectionListPath,
} from '../api.ts';
import type {WebtaskConnection} from '../types.ts';
import {WebtaskConnectionEditor} from './WebtaskConnectionEditor.tsx';
import {
    createDefaultWebtaskConnectionDraft,
    normalizeLoadedWebtaskConnection,
    snapshotWebtaskConnectionDraft,
    validateWebtaskConnectionDraft,
} from './webtaskConnectionEditorUtils';

export const EditWebtaskConnection = () => {
    const navigate = useNavigate();
    const {id} = useParams();
    const connectionId = id || '';
    const {reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {reload: reloadConnections} = useConnections();
    const {confirmPrompt} = useMuiPrompts();
    const [loading, setLoading] = useState(true);
    const [agents, setAgents] = useState<AgentOption[]>([]);
    const [draft, setDraft] = useState<Partial<WebtaskConnection>>(() => createDefaultWebtaskConnectionDraft());
    const [showPassword, setShowPassword] = useState(false);
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState<null | { ok: boolean; message: string }>(null);
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [dirty, setDirty] = useState<boolean>(false);
    const baselineRef = useRef<string>('');
    useUnsavedChanges(dirty);

    useEffect(() => {
        let alive = true;
        (async () => {
            try {
                setLoading(true);
                const loaded = await fetchConnectionById<WebtaskConnection>('webtask', connectionId);
                const normalized = normalizeLoadedWebtaskConnection(loaded);
                if (!alive) return;
                if ((normalized as any).suspended) {
                    navigate(getConnectionListPath('webtask'));
                    return;
                }
                const agentOptions = mergeSelectedAgent(
                    await fetchAgentOptions(),
                    normalized.agentUuid,
                    (loaded as any).agentName,
                );
                if (!alive) return;
                setDraft(normalized);
                baselineRef.current = snapshotWebtaskConnectionDraft(normalized);
                setAgents(agentOptions);
            } catch (e: any) {
                setErrors({api: e.message || 'Failed to load connection.'});
            } finally {
                if (alive) setLoading(false);
            }
        })();
        return () => {
            alive = false
        };
    }, [connectionId, navigate]);

    useEffect(() => {
        if (baselineRef.current && !loading) {
            setDirty(snapshotWebtaskConnectionDraft(draft) !== baselineRef.current);
        }
    }, [draft, loading]);

    const validate = (): boolean => {
        const nextErrors = validateWebtaskConnectionDraft(draft, {allowRedactedSecrets: true});
        setErrors(nextErrors);
        return Object.keys(nextErrors).length === 0;
    };

    const handleTest = async () => {
        if (!validate()) return;
        setTesting(true);
        setTestResult(null);
        try {
            const data = await apiPost(`${getConnectionApiItemPath('webtask', connectionId)}/test`, draft) as any;
            setTestResult({ok: !!data?.ok, message: data?.message || (data?.ok ? 'Connection successful!' : 'Connection failed.')});
        } catch (e: any) {
            setTestResult({ok: false, message: e.message || 'Network error testing connection.'});
        } finally {
            setTesting(false);
        }
    };

    const handleDuplicate = async () => {
        try {
            setLoading(true);
            const data = await duplicateStoredConnection({kind: 'webtask', id: connectionId}) as any;
            await reloadFeatureAvailability();
            await reloadConnections();
            if (data?.id) {
                navigate(getConnectionEditPath('webtask', data.id));
            } else {
                navigate(getConnectionListPath('webtask'));
            }
        } catch (e: any) {
            setErrors({api: e.message || 'Failed to duplicate connection'});
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        if (!validate()) return;
        try {
            await apiPut(getConnectionApiItemPath('webtask', connectionId), draft);
            setDirty(false);
            void reloadFeatureAvailability();
            void reloadConnections();
            navigate(getConnectionListPath('webtask'), {state: {refresh: true}});
        } catch (e: any) {
            setErrors({api: e.message || 'Network error saving connection.'});
        }
    };

    const handleDelete = async () => {
        const ok = await confirmPrompt({
            title: 'Delete Connection',
            message: `Are you sure you want to delete "${draft.name}"? This action cannot be undone.`,
            buttonText: 'Delete',
            cancelButtonText: 'Cancel'
        });
        if (!ok) return;

        try {
            await deleteStoredConnection({kind: 'webtask', id: connectionId});
            void reloadConnections();
            navigate(getConnectionListPath('webtask'));
        } catch (e: any) {
            console.error(e);
            setErrors({api: e.message || 'Failed to delete connection'});
        }
    };

    const dangerZone = (
        <Card variant="outlined" sx={{borderRadius: 3, border: '1px solid', borderColor: 'error.main', mt: 4}}>
            <CardActions sx={{p: 2, justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2}}>
                <VStack spacing={0.5} alignItems="flex-start">
                    <Typography
                        variant="subtitle1"
                        sx={{
                            fontWeight: "bold",
                            color: "error.main"
                        }}>Danger Zone</Typography>
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Once you delete a connection, there is no going back. Please be certain.</Typography>
                </VStack>
                <Button variant="outlined" color="error" onClick={handleDelete} disabled={loading}>Delete Connection</Button>
            </CardActions>
        </Card>
    )

    if (loading) {
        return (
            <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px'}}>
                <CircularProgress/>
            </Box>
        );
    }

    return (
        <WebtaskConnectionEditor
            title="Edit Web Task Connection"
            infoTooltip="Update configuration for this API endpoint."
            draft={draft}
            setDraft={setDraft}
            errors={errors}
            agents={agents}
            showPassword={showPassword}
            setShowPassword={setShowPassword}
            loading={loading}
            testing={testing}
            testResult={testResult}
            onDismissTestResult={() => setTestResult(null)}
            onTest={handleTest}
            onCancel={() => confirmOnNavigate(dirty, navigate, confirmPrompt)(getConnectionListPath('webtask'))}
            onSave={handleSave}
            saveLabel="Save Changes"
            headerAction={(
                <Button
                    variant="outlined"
                    startIcon={<ContentCopyIcon/>}
                    onClick={handleDuplicate}
                    disabled={loading}
                >
                    Duplicate Connection
                </Button>
            )}
            dangerZone={dangerZone}
            apiError={errors.api}
        />
    );
};
