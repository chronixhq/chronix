import {useCallback, useEffect, useState} from 'react';
import {useNavigate, useParams} from 'react-router';
import {apiGet, apiPost} from '@dsherwin/react-api-interface';
import {Alert, Box, Button, Card, CardContent, Chip, CircularProgress, Divider, Grid, Snackbar, Tooltip, Typography} from '@mui/material';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {formatDateTime} from '../../lib/utilities';
import {RestartAlt, SystemUpdate, Warning} from "@mui/icons-material";

export const AgentDetail = () => {
    const navigate = useNavigate();
    const {uuid} = useParams();
    const {confirmPrompt} = useMuiPrompts();
    const [data, setData] = useState<any | null>(null);
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | undefined>();
    const [updaterStatus, setUpdaterStatus] = useState<any | null>(null);
    const [busy, setBusy] = useState(false);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' | 'warning' }>({
        open: false,
        message: '',
        severity: 'info'
    });

    const load = useCallback(async () => {
        if (!uuid) return;
        try {
            setLoading(true);
            setError(undefined);
            const [res, statusRes] = await Promise.all([
                apiGet(`/agents/${encodeURIComponent(uuid)}`),
                apiGet('/settings/updater/status')
            ]);
            if ((res as any).suspended) {
                navigate('/agents/list')
                return
            }
            setData(res);
            setUpdaterStatus(statusRes);
        } catch (e: any) {
            setError(e?.message || 'Failed to load agent');
            setData(null);
        } finally {
            setLoading(false);
        }
    }, [uuid, navigate]);

    useEffect(() => {
        void load();
    }, [load]);

    const onUpdate = async () => {
        if (!data || !updaterStatus?.availableVersion?.['chronix-agent']) return;
        const targetVersion = updaterStatus.availableVersion['chronix-agent'].version;

        const ok = await confirmPrompt({
            title: 'Update Agent',
            message: `Are you sure you want to update agent '${data.name}' to version ${targetVersion}? The agent will restart.`,
            buttonText: 'Update Agent',
            cancelButtonText: 'Cancel'
        });
        if (!ok) return;

        setBusy(true);
        try {
            await apiPost(`/agents/${data.uuid}/update`, {});
            setSnack({open: true, message: `Update initiated for ${data.name}`, severity: 'success'});

            // Poll for agent to come back online with new version
            let attempts = 0;
            const maxAttempts = 60; // 2 minutes with 2s interval
            const poll = async () => {
                try {
                    const updatedAgent = await apiGet(`/agents/${encodeURIComponent(data.uuid)}`) as any;
                    // Note: we can't easily check "online" state from just getAgent because it doesn't return it
                    // Wait, cxrestapi/agents.go getAgent doesn't return online status!
                    // But it returns lastSeenAt. We can calculate it.
                    const onlineThreshold = 2 * 60 * 1000;
                    const isOnline = updatedAgent.lastSeenAt && (Date.now() - new Date(updatedAgent.lastSeenAt).getTime()) <= onlineThreshold;

                    if (updatedAgent && isOnline && updatedAgent.version === targetVersion) {
                        setSnack({open: true, message: `Agent ${data.name} updated successfully to ${targetVersion}`, severity: 'success'});
                        setBusy(false);
                        void load();
                    } else {
                        attempts++;
                        if (attempts < maxAttempts) {
                            setTimeout(poll, 2000);
                        } else {
                            setSnack({open: true, message: `Agent ${data.name} update verification timed out.`, severity: 'error'});
                            setBusy(false);
                            void load();
                        }
                    }
                } catch {
                    attempts++;
                    if (attempts < maxAttempts) {
                        setTimeout(poll, 2000);
                    } else {
                        setBusy(false);
                        void load();
                    }
                }
            };
            setTimeout(poll, 5000);
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: `Failed to update agent: ${e?.message || data.name}`, severity: 'error'});
            setBusy(false);
        }
    };

    const onRestart = async () => {
        if (!data || !online) return;

        const ok = await confirmPrompt({
            title: 'Restart Agent',
            message: `Are you sure you want to restart agent '${data.name}'?`,
            buttonText: 'Restart Agent',
            cancelButtonText: 'Cancel'
        });
        if (!ok) return;

        setBusy(true);
        try {
            await apiPost(`/agents/${data.uuid}/restart`, {});
            setSnack({open: true, message: `Restart initiated for ${data.name}`, severity: 'info'});

            // Poll for agent to come back online
            let attempts = 0;
            const maxAttempts = 60; // 2 minutes with 2s interval
            const poll = async () => {
                try {
                    const updatedAgent = await apiGet(`/agents/${encodeURIComponent(data.uuid)}`) as any;
                    const onlineThreshold = 2 * 60 * 1000;
                    const isOnline = updatedAgent.lastSeenAt && (Date.now() - new Date(updatedAgent.lastSeenAt).getTime()) <= onlineThreshold;

                    if (updatedAgent && isOnline) {
                        setSnack({open: true, message: `Agent ${data.name} is back online`, severity: 'success'});
                        setBusy(false);
                        void load();
                    } else {
                        attempts++;
                        if (attempts < maxAttempts) {
                            setTimeout(poll, 2000);
                        } else {
                            setSnack({open: true, message: `Agent ${data.name} restart verification timed out.`, severity: 'error'});
                            setBusy(false);
                            void load();
                        }
                    }
                } catch {
                    attempts++;
                    if (attempts < maxAttempts) {
                        setTimeout(poll, 2000);
                    } else {
                        setBusy(false);
                        void load();
                    }
                }
            };
            setTimeout(poll, 5000);
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: `Failed to restart agent: ${e?.message || data.name}`, severity: 'error'});
            setBusy(false);
        }
    };

    const online = !!data?.lastSeenAt && (Date.now() - new Date(data.lastSeenAt).getTime()) <= (2 * 60 * 1000);
    const updateAvailable = updaterStatus?.availableVersion?.['chronix-agent'] && data?.version && data.version !== updaterStatus.availableVersion['chronix-agent'].version;

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1280, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between">
                    <Typography variant="h5">Agent Detail: {data?.name || '...'}</Typography>
                    <HStack spacing={1}>
                        {online && (
                            <Button
                                variant="outlined"
                                color="primary"
                                startIcon={<RestartAlt/>}
                                onClick={onRestart}
                                disabled={busy || data?.suspended}
                            >
                                {busy ? 'BUSY...' : 'RESTART AGENT'}
                            </Button>
                        )}
                        {updateAvailable && online && (
                            <Button
                                variant="contained"
                                color="warning"
                                startIcon={<SystemUpdate/>}
                                onClick={onUpdate}
                                disabled={busy || data?.suspended}
                            >
                                {busy ? 'UPDATING...' : `UPDATE TO ${updaterStatus.availableVersion['chronix-agent'].version}`}
                            </Button>
                        )}
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {loading && <CircularProgress/>}
                {error && <Alert severity="error" sx={{mb: 2}}>{error}</Alert>}

                {(!loading && data) && (
                    <VStack spacing={3}>
                        {data.suspended && (
                            <Alert severity="warning" variant="filled" icon={<Warning />} sx={{ borderRadius: 3 }}>
                                This agent is <strong>suspended</strong>. It will not be allowed to connect or perform any tasks until it is re-enabled.
                            </Alert>
                        )}
                        <Card variant="outlined" sx={{borderRadius: 3}}>
                            <CardContent>
                                <Grid container spacing={3}>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Name</Typography>
                                            <Typography variant="body1" sx={{
                                                fontWeight: 500
                                            }}>{data.name}</Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Status</Typography>
                                            <HStack spacing={1}>
                                                {data.suspended ? (
                                                    <Tooltip title="Suspended. This agent is temporarily disconnected and inactive.">
                                                        <Chip size="small" color="warning" icon={<Warning fontSize="small"/>} label="Suspended"/>
                                                    </Tooltip>
                                                ) : (
                                                    <Chip size="small" label={data.status} color={String(data.status).toLowerCase() === 'active' ? 'success' : 'default'}/>
                                                )}
                                                {online ? <Chip size="small" color="success" label="Online"/> : <Chip size="small" label="Offline"/>}
                                            </HStack>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Version</Typography>
                                            <Typography variant="body1" sx={{
                                                fontWeight: 500
                                            }}>{data.version || 'Unknown'}</Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>OS / Platform</Typography>
                                            <Typography variant="body1" sx={{
                                                fontWeight: 500
                                            }}>
                                                {data.osType || data.os || '—'} {data.osVersion ? `(${data.osVersion})` : ''}
                                            </Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Architecture</Typography>
                                            <Typography variant="body1" sx={{
                                                fontWeight: 500
                                            }}>{data.arch || '—'}</Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Running As</Typography>
                                            <Typography variant="body1" sx={{
                                                fontWeight: 500
                                            }}>{data.runningUser || '—'}</Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>UUID</Typography>
                                            <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{data.uuid}</Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Last Seen</Typography>
                                            <Typography variant="body1">{data.lastSeenAt ? formatDateTime(data.lastSeenAt) : 'Never'}</Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 6}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Last IP Address</Typography>
                                            <Typography variant="body1">{data.lastSeenIp || '—'}</Typography>
                                        </VStack>
                                    </Grid>
                                    <Grid size={{xs: 12}}>
                                        <VStack spacing={0.5}>
                                            <Typography variant="subtitle2" sx={{
                                                color: "text.secondary"
                                            }}>Public Key</Typography>
                                            <Typography variant="body2" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>{data.publicKey || '—'}</Typography>
                                        </VStack>
                                    </Grid>
                                </Grid>
                            </CardContent>
                        </Card>

                        <VStack spacing={1}>
                            <Typography variant="h6">Metadata</Typography>
                            <Card variant="outlined" sx={{borderRadius: 3}}>
                                <CardContent sx={{p: 0}}>
                                    <Box component="pre" sx={{m: 0, p: 2, backgroundColor: (theme) => theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.2)' : 'grey.100', borderRadius: 0, overflowX: 'auto', fontSize: '0.875rem', fontFamily: 'monospace'}}>
                                        {JSON.stringify(data.metadataJson ?? {}, null, 2)}
                                    </Box>
                                </CardContent>
                            </Card>
                        </VStack>
                    </VStack>
                )}
            </VStack>
            <Snackbar
                open={snack.open}
                autoHideDuration={6000}
                onClose={() => setSnack(s => ({...s, open: false}))}
                message={snack.message}
            />
        </Box>
    );
};
