import {useEffect, useMemo, useRef, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, Divider, Tooltip, Typography} from '@mui/material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import {apiDelete, apiGet, apiPost, apiPut} from '@dsherwin/react-api-interface';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {useNavigate, useParams} from 'react-router';
import {useFeatureAvailability} from '../../../data/FeatureAvailabilityContext.tsx';
import {confirmOnNavigate, useUnsavedChanges} from '../../../lib/useUnsavedChanges';
import {useConnections} from '../../../data/ConnectionsContext.tsx';
import {fetchAgentOptions, mergeSelectedAgent} from '../../Agents/api.ts';
import {ShellConnectionForm} from './ShellConnectionForm';
import {buildShellConnectionSavePayload, buildShellConnectionTestPayload, canSaveShellConnection, createDefaultShellConnectionDraft, createDefaultShellConnectionUiState, loadedShellToEditorState, snapshotShellConnectionDraft, type LoadedShell, type ShellConnectionDraft, type ShellConnectionSecretFlags, type ShellConnectionUiState} from './shellConnectionEditorUtils';

async function copyToClipboard(text: string) {
    if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        return;
    }

    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.left = '-9999px';
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();
    try {
        document.execCommand('copy');
    } finally {
        document.body.removeChild(textarea);
    }
}

export const EditShellConnection = () => {
    const navigate = useNavigate();
    const {id} = useParams();
    const {reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {reload: reloadConnections} = useConnections();
    const {confirmPrompt} = useMuiPrompts();

    const [draft, setDraft] = useState<ShellConnectionDraft>(() => createDefaultShellConnectionDraft());
    const [secretFlags, setSecretFlags] = useState<ShellConnectionSecretFlags>(() => ({
        hasPassword: false,
        hasPrivateKey: false,
        hasKeyPass: false,
        hasSudoPassword: false,
    }));
    const [uiState, setUiState] = useState<ShellConnectionUiState>(() => createDefaultShellConnectionUiState());
    const [agents, setAgents] = useState<Array<{ uuid: string; name: string }>>([]);
    const [agentsLoading, setAgentsLoading] = useState(false);
    const [initialAgent, setInitialAgent] = useState<{ uuid: string; name: string } | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState<null | { ok: boolean; message: string }>(null);
    const [generatingKey, setGeneratingKey] = useState(false);
    const [dirty, setDirty] = useState(false);
    const baselineRef = useRef<string>('');

    const patchDraft = (patch: Partial<ShellConnectionDraft>) => setDraft((prev) => ({...prev, ...patch}));
    const patchUi = (patch: Partial<ShellConnectionUiState>) => setUiState((prev) => ({...prev, ...patch}));
    const canSave = useMemo(() => canSaveShellConnection(draft), [draft]);

    useEffect(() => {
        let alive = true;
        (async () => {
            setLoading(true);
            setLoadError(null);
            try {
                const data = await apiGet(`/shell/connections/${encodeURIComponent(id || '')}`) as LoadedShell;
                if (!alive) return;
                if (data.suspended) {
                    navigate('/shell/list');
                    return;
                }

                const next = loadedShellToEditorState(data);
                setDraft(next.draft);
                setSecretFlags(next.secretFlags);
                if (data.agent_uuid && data.agent_name) {
                    setInitialAgent({uuid: data.agent_uuid, name: data.agent_name});
                }
                baselineRef.current = snapshotShellConnectionDraft(next.draft);
                setDirty(false);
            } catch (error) {
                console.error(error);
                setLoadError('Failed to load shell connection');
            } finally {
                setLoading(false);
            }
        })();
        return () => {
            alive = false;
        };
    }, [id, navigate]);

    useEffect(() => {
        let alive = true;
        (async () => {
            try {
                setAgentsLoading(true);
                const list = await fetchAgentOptions();
                if (!alive) return;
                setAgents(mergeSelectedAgent(list, initialAgent?.uuid, initialAgent?.name));
            } catch (error) {
                console.log(error);
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
        if (baselineRef.current === '') return;
        setDirty(snapshotShellConnectionDraft(draft) !== baselineRef.current);
    }, [draft]);

    useUnsavedChanges(dirty);

    async function onClearSecret(field: 'ssh_password' | 'ssh_private_key' | 'ssh_key_pass' | 'sudo_password', label: string) {
        const ok = await confirmPrompt({
            title: `Clear ${label}?`,
            message: `This will permanently remove the saved ${label} for this connection from the server. You cannot undo this action.`,
            buttonText: 'Clear Now',
            buttonColor: 'error',
        });
        if (!ok) return;

        try {
            await apiPost(`/shell/connections/${id}/clear-secret`, {field});
            if (field === 'ssh_password') {
                setSecretFlags((prev) => ({...prev, hasPassword: false}));
                patchDraft({sshPassword: ''});
            } else if (field === 'ssh_private_key') {
                setSecretFlags((prev) => ({...prev, hasPrivateKey: false}));
                patchDraft({sshPrivateKey: ''});
            } else if (field === 'ssh_key_pass') {
                setSecretFlags((prev) => ({...prev, hasKeyPass: false}));
                patchDraft({sshKeyPass: ''});
            } else if (field === 'sudo_password') {
                setSecretFlags((prev) => ({...prev, hasSudoPassword: false}));
                patchDraft({sudoPassword: ''});
            }
        } catch (error: any) {
            setError(error?.message || `Failed to clear ${label}`);
        }
    }

    async function onDuplicate() {
        setSaving(true);
        setError(null);
        try {
            const data = await apiPost(`/shell/connections/${id}/duplicate`) as any;
            await reloadFeatureAvailability();
            await reloadConnections();
            if (data && data.id) navigate(`/shell/edit/${data.id}`);
            else navigate('/shell/list');
        } catch (error: any) {
            setError(error?.message || 'Failed to duplicate connection');
        } finally {
            setSaving(false);
        }
    }

    async function onSave() {
        setSaving(true);
        setError(null);
        try {
            await apiPut(`/shell/connections/${encodeURIComponent(id || '')}`, buildShellConnectionSavePayload(draft)) as any;
            void reloadFeatureAvailability();
            void reloadConnections();
            navigate('/shell/list');
        } catch (error: any) {
            setError(error?.message || 'Failed to save shell connection');
        } finally {
            setSaving(false);
        }
    }

    async function onTest() {
        setTesting(true);
        setTestResult(null);
        try {
            const response = await apiPost('/shell/connections/test', buildShellConnectionTestPayload(draft, {id, secretFlags})) as any;
            setTestResult({ok: !!response?.ok, message: response?.message || (response?.ok ? 'Connection ok' : 'Test failed')});
        } catch (error: any) {
            setTestResult({ok: false, message: error?.message || 'Test failed'});
        } finally {
            setTesting(false);
        }
    }

    async function onGenerateKeyPair() {
        patchUi({showFormatDialog: false});
        setGeneratingKey(true);
        patchUi({generatedPublicKey: null});
        try {
            const response = await apiPost('/shell/connections/generate-keypair', {format: uiState.keyFormat}) as any;
            if (response?.privateKey && response?.publicKey) {
                patchDraft({sshPrivateKey: response.privateKey});
                patchUi({generatedPublicKey: response.publicKey});
                setSecretFlags((prev) => ({...prev, hasPrivateKey: false}));
            }
        } catch (error: any) {
            setError(error?.message || 'Failed to generate key pair');
        } finally {
            setGeneratingKey(false);
        }
    }

    async function onDelete() {
        const ok = await confirmPrompt({
            title: 'Delete Connection',
            message: `Are you sure you want to delete "${draft.name}"? This action cannot be undone.`,
            buttonText: 'Delete',
            cancelButtonText: 'Cancel',
        });
        if (!ok) return;

        try {
            await apiDelete(`/shell/connections/${id}`);
            void reloadConnections();
            navigate('/shell/list');
        } catch (error: any) {
            console.error(error);
            setError(error?.message || 'Failed to delete connection');
        }
    }

    const renderDeleteBar = () => (
        <Card variant="outlined" sx={{borderRadius: 3, border: '1px solid', borderColor: 'error.main', mt: 4}}>
            <CardActions sx={{p: 2, justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2}}>
                <VStack spacing={0.5} alignItems="flex-start">
                    <Typography variant="subtitle1" sx={{fontWeight: 'bold', color: 'error.main'}}>Danger Zone</Typography>
                    <Typography variant="body2" sx={{color: 'text.secondary'}}>Once you delete a connection, there is no going back. Please be certain.</Typography>
                </VStack>
                <Button variant="outlined" color="error" onClick={onDelete} disabled={loading || saving}>Delete Connection</Button>
            </CardActions>
        </Card>
    );

    const renderActionButtons = () => (
        <>
            {testResult && <Alert severity={testResult.ok ? 'success' : 'warning'}>{testResult.message}</Alert>}
            <Card variant="outlined" sx={{borderRadius: 3, bgcolor: 'action.hover'}}>
                <CardActions sx={{p: 2, justifyContent: 'space-between', flexWrap: 'wrap', gap: 2}}>
                    <Typography variant="body2" sx={{color: 'text.secondary'}}>Tip: Test changes before saving.</Typography>
                    <HStack spacing={1}>
                        <Button variant="outlined" onClick={() => void onTest()} disabled={testing || loading}>{testing ? 'Testing…' : 'Test connection'}</Button>
                        <Button variant="outlined" onClick={() => confirmOnNavigate(dirty, navigate, confirmPrompt)('/shell/list')}>Cancel</Button>
                        <Button variant="contained" onClick={() => void onSave()} disabled={saving || !canSave}>Save Changes</Button>
                    </HStack>
                </CardActions>
            </Card>
        </>
    );

    if (loading) {
        return (
            <Box sx={{px: {xs: 1, md: 2}, py: 2}}>
                <Typography variant="body2">Loading…</Typography>
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
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 800, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" spacing={1} sx={{justifyContent: 'space-between', flexWrap: 'wrap'}}>
                    <HStack alignItems="center" spacing={1}>
                        <Typography variant="h5">Edit shell connection</Typography>
                        <Tooltip title="Used by shell jobs.">
                            <InfoOutlinedIcon fontSize="small"/>
                        </Tooltip>
                    </HStack>
                    <Button variant="outlined" startIcon={<ContentCopyIcon/>} onClick={onDuplicate} disabled={loading || saving}>
                        Duplicate Connection
                    </Button>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {error && <Alert severity="error">{error}</Alert>}

                {renderActionButtons()}

                <ShellConnectionForm
                    draft={draft}
                    onDraftChange={patchDraft}
                    secretFlags={secretFlags}
                    uiState={uiState}
                    onUiChange={patchUi}
                    agents={agents}
                    agentsLoading={agentsLoading}
                    generatingKey={generatingKey}
                    onGenerateKeyPair={() => void onGenerateKeyPair()}
                    onCopyGeneratedPublicKey={() => {
                        if (uiState.generatedPublicKey) {
                            void copyToClipboard(uiState.generatedPublicKey);
                            patchUi({copied: true});
                            setTimeout(() => patchUi({copied: false}), 2000);
                        }
                    }}
                    onDismissGeneratedPublicKey={() => patchUi({generatedPublicKey: null})}
                    onClearSecret={(field, label) => void onClearSecret(field, label)}
                />

                {renderActionButtons()}
                {renderDeleteBar()}
            </VStack>
        </Box>
    );
};
