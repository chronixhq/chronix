import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, Divider, Tooltip, Typography} from '@mui/material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import {apiPost} from '@dsherwin/react-api-interface';
import {useNavigate} from 'react-router';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {useFeatureAvailability} from '../../../data/FeatureAvailabilityContext.tsx';
import {useConnections} from '../../../data/ConnectionsContext.tsx';
import {fetchAgentOptions} from '../../Agents/api.ts';
import {ShellConnectionForm} from './ShellConnectionForm';
import {buildShellConnectionSavePayload, buildShellConnectionTestPayload, canSaveShellConnection, createDefaultShellConnectionDraft, createDefaultShellConnectionSecretFlags, createDefaultShellConnectionUiState, type ShellConnectionDraft, type ShellConnectionUiState} from './shellConnectionEditorUtils';

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

export const CreateShellConnection = () => {
    const navigate = useNavigate();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {reload: reloadConnections} = useConnections();
    const shLimit = checkLimit('shell_connections');

    useEffect(() => {
        if (!shLimit.allowed) {
            navigate('/shell/list');
        }
    }, [shLimit.allowed, navigate]);

    const [draft, setDraft] = useState<ShellConnectionDraft>(() => createDefaultShellConnectionDraft());
    const [uiState, setUiState] = useState<ShellConnectionUiState>(() => createDefaultShellConnectionUiState());
    const [agents, setAgents] = useState<Array<{ uuid: string; name: string }>>([]);
    const [agentsLoading, setAgentsLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState<null | { ok: boolean; message: string }>(null);
    const [generatingKey, setGeneratingKey] = useState(false);

    useEffect(() => {
        let alive = true;
        (async () => {
            try {
                setAgentsLoading(true);
                const list = await fetchAgentOptions();
                if (!alive) return;
                setAgents(list);
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
    }, []);

    const patchDraft = (patch: Partial<ShellConnectionDraft>) => setDraft((prev) => ({...prev, ...patch}));
    const patchUi = (patch: Partial<ShellConnectionUiState>) => setUiState((prev) => ({...prev, ...patch}));
    const canSave = canSaveShellConnection(draft);

    async function onSave() {
        setSaving(true);
        setError(null);
        try {
            await apiPost('/shell/connections', buildShellConnectionSavePayload(draft)) as any;
            void reloadFeatureAvailability();
            void reloadConnections();
            navigate('/shell/list');
        } catch (error: any) {
            setError(error?.message || 'Failed to create shell connection');
        } finally {
            setSaving(false);
        }
    }

    async function onTest() {
        setTesting(true);
        setTestResult(null);
        try {
            const response = await apiPost('/shell/connections/test', buildShellConnectionTestPayload(draft, {nameFallback: '(unsaved) test'})) as any;
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
            }
        } catch (error: any) {
            setError(error?.message || 'Failed to generate key pair');
        } finally {
            setGeneratingKey(false);
        }
    }

    const renderActionButtons = () => (
        <>
            {testResult && <Alert severity={testResult.ok ? 'success' : 'warning'}>{testResult.message}</Alert>}
            <Card variant="outlined" sx={{borderRadius: 3, bgcolor: 'action.hover'}}>
                <CardActions sx={{p: 2, justifyContent: 'space-between', flexWrap: 'wrap', gap: 2}}>
                    <Typography variant="body2" sx={{color: 'text.secondary'}}>Tip: Test changes before saving.</Typography>
                    <HStack spacing={1}>
                        <Button variant="outlined" onClick={() => void onTest()} disabled={testing || !canSave}>{testing ? 'Testing…' : 'Test connection'}</Button>
                        <Button variant="outlined" onClick={() => navigate('/shell/list')}>Cancel</Button>
                        <Button variant="contained" onClick={() => void onSave()} disabled={saving || !canSave}>Save and Continue</Button>
                    </HStack>
                </CardActions>
            </Card>
        </>
    );

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 800, width: '100%', mx: 'auto'}}>
                <HStack alignItems="flex-end" justifyContent="space-between">
                    <VStack spacing={2} sx={{maxWidth: 800, width: '100%', mx: 'auto'}}>
                        <HStack alignItems="center" spacing={1}>
                            <Typography variant="h5">Add shell connection</Typography>
                            <Tooltip title="Used by jobs and explorers. You can edit later.">
                                <InfoOutlinedIcon fontSize="small"/>
                            </Tooltip>
                        </HStack>
                        <Typography variant="body2" sx={{color: 'text.secondary'}}>
                            Enter the minimum required fields. Advanced options are optional.
                        </Typography>
                    </VStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {error && <Alert severity="error">{error}</Alert>}

                {renderActionButtons()}

                <ShellConnectionForm
                    draft={draft}
                    onDraftChange={patchDraft}
                    secretFlags={createDefaultShellConnectionSecretFlags()}
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
                />

                {renderActionButtons()}
            </VStack>
        </Box>
    );
};
