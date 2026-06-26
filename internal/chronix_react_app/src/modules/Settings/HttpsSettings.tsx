import {useEffect, useMemo, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, FormControl, FormHelperText, InputLabel, MenuItem, Select, Snackbar, Switch, TextField, Typography} from '@mui/material';
import type {GlobalAgentSettings, GlobalHttpSettings, GlobalHttpsSettings} from './types';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {
    fetchNetworkSettings,
    removeHttpsCertificatePair,
    restartNetworkListeners,
    saveAgentSettings,
    saveHttpSettings,
    saveHttpsSettings,
    uploadHttpsCertificatePair,
} from './api.ts';
import {SectionHelp} from '../../main/SectionHelp';

export const HttpsSettingsPage = () => {
    const [http, setHttp] = useState<GlobalHttpSettings>({enabled: false, port: '80'});
    const [https, setHttps] = useState<GlobalHttpsSettings>({mode: 'selfsigned', port: '443', enabled: true});
    const [agent, setAgent] = useState<GlobalAgentSettings>({enabled: true, port: '5172'});
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    // Local staged files (not uploaded until Save)
    const [selectedCert, setSelectedCert] = useState<File | null>(null);
    const [selectedKey, setSelectedKey] = useState<File | null>(null);
    const [fieldErrors, setFieldErrors] = useState<{ certKeyPair?: string; httpPort?: string; httpsPort?: string; agentPort?: string } | null>(null);
    const [confirmRemoveOpen, setConfirmRemoveOpen] = useState(false);
    const [restartPromptOpen, setRestartPromptOpen] = useState(false);
    const [restarting, setRestarting] = useState(false);

    useEffect(() => {
        (async () => {
            try {
                const network = await fetchNetworkSettings();
                setHttp(network.http);
                setHttps(network.https);
                setAgent(network.agent);
            } catch (e) {
                console.error(e);
                setSnack({open: true, message: 'Failed to load Network settings.', severity: 'error'});
            }
        })();
    }, []);

    const hasServerUploadedPair = useMemo(() => !!https.certFileName && !!https.keyFileName, [https.certFileName, https.keyFileName]);
    const hasLocalPair = useMemo(() => !!selectedCert && !!selectedKey, [selectedCert, selectedKey]);

    const portsConflict = useMemo(() => {
        const activePorts = [
            http.enabled ? http.port : null,
            https.enabled ? https.port : null,
            agent.enabled ? agent.port : null,
        ].filter(p => p !== null);
        return new Set(activePorts).size !== activePorts.length;
    }, [http.enabled, http.port, https.enabled, https.port, agent.enabled, agent.port]);

    // HTTPS-specific save enablement
    const canSaveHttps = useMemo(() => {
        if (https.mode !== 'upload') return true;
        if (hasServerUploadedPair) return true;
        return hasLocalPair;
    }, [https.mode, hasServerUploadedPair, hasLocalPair]);

    // Save only HTTP block
    const onSaveHttp = async () => {
        setFieldErrors(null);
        try {
            // Basic client validation: ports must differ when both enabled
            if (portsConflict) {
                setFieldErrors({
                    httpPort: 'Conflict with other service port',
                    httpsPort: 'Conflict with other service port',
                    agentPort: 'Conflict with other service port'
                });
                setSnack({open: true, message: 'All enabled service ports must be unique.', severity: 'error'});
                return;
            }
            await saveHttpSettings(http);
            setSnack({open: true, message: 'HTTP settings saved.', severity: 'success'});
            // These settings require a restart to take effect
            setRestartPromptOpen(true);
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Failed to save HTTP settings.', severity: 'error'});
        }
    };

    // Save only HTTPS block
    const onSaveHttps = async () => {
        setFieldErrors(null);
        try {
            // Basic client validation: ports must differ when both enabled
            if (portsConflict) {
                setFieldErrors({
                    httpPort: 'Conflict with other service port',
                    httpsPort: 'Conflict with other service port',
                    agentPort: 'Conflict with other service port'
                });
                setSnack({open: true, message: 'All enabled service ports must be unique.', severity: 'error'});
                return;
            }
            // If mode is upload and server does not yet have a pair, require local pair and upload both first
            if (https.mode === 'upload' && !hasServerUploadedPair) {
                if (!hasLocalPair) {
                    setFieldErrors({certKeyPair: 'Both certificate and key are required before saving.'});
                    setSnack({open: true, message: 'Select both certificate and key before saving.', severity: 'error'});
                    return;
                }
                const uploaded = await uploadHttpsCertificatePair(selectedCert!, selectedKey!);
                setHttps((current) => ({
                    ...current,
                    ...uploaded,
                    certFileName: uploaded.certFileName || selectedCert!.name || 'uploaded',
                    keyFileName: uploaded.keyFileName || selectedKey!.name || 'uploaded',
                }));
                // clear local staged files after successful upload
                setSelectedCert(null);
                setSelectedKey(null);
            }

            await saveHttpsSettings(https);
            setSnack({open: true, message: 'HTTPS settings saved.', severity: 'success'});
            // These settings require a restart to take effect
            setRestartPromptOpen(true);
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Failed to save HTTPS settings.', severity: 'error'});
        }
    };
    
    // Save only Agent block
    const onSaveAgent = async () => {
        setFieldErrors(null);
        try {
            // Basic client validation: ports must differ when both enabled
            if (portsConflict) {
                setFieldErrors({
                    httpPort: 'Conflict with other service port',
                    httpsPort: 'Conflict with other service port',
                    agentPort: 'Conflict with other service port'
                });
                setSnack({open: true, message: 'All enabled service ports must be unique.', severity: 'error'});
                return;
            }
            await saveAgentSettings(agent);
            setSnack({open: true, message: 'Agent connection settings saved.', severity: 'success'});
            // These settings require a restart to take effect
            setRestartPromptOpen(true);
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Failed to save Agent settings.', severity: 'error'});
        }
    };

    const onSelect = async (kind: 'cert' | 'key') => {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = kind === 'cert' ? '.pem,.crt,.cer,.cert,text/plain' : '.pem,.key,text/plain';
        input.onchange = () => {
            if (!input.files || input.files.length === 0) return;
            const file = input.files[0]!;
            if (kind === 'cert') {
                setSelectedCert(file);
            } else {
                setSelectedKey(file);
            }
            setFieldErrors(null);
        };
        input.click();
    };

    const onRemoveBoth = async () => {
        try {
            if (hasServerUploadedPair) {
                await removeHttpsCertificatePair();
                setHttps(s => ({...s, certFileName: '', keyFileName: '', certInfo: undefined}));
            }
            setSelectedCert(null);
            setSelectedKey(null);
            setSnack({open: true, message: 'Certificate and key removed.', severity: 'success'});
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Failed to remove certificate and key.', severity: 'error'});
        }
    };

    const onConfirmRemove = async () => {
        await onRemoveBoth();
        setConfirmRemoveOpen(false);
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 900, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5">HTTP / HTTPS / Agent</Typography>
                        <SectionHelp />
                    </Box>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {/* HTTP Block */}
                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={2} alignItems="flex-start">
                            <Typography variant="h6">HTTP</Typography>
                            <VStack spacing={2} alignItems={"flex-start"} flexWrap={"wrap"}>
                                <HStack spacing={1} alignItems={"center"}>
                                    <Switch checked={!!http.enabled} onChange={(e) => setHttp(s => ({...s, enabled: e.target.checked}))}/>
                                    <Typography>Enable HTTP</Typography>
                                </HStack>
                                <TextField
                                    label="Port"
                                    value={http.port}
                                    onChange={(e) => setHttp(s => ({...s, port: e.target.value.replace(/[^0-9]/g, '')}))}
                                    sx={{minWidth: {xs: '100%', md: 160}}}
                                    slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                                    error={!!fieldErrors?.httpPort}
                                    helperText={fieldErrors?.httpPort}
                                />
                            </VStack>
                            <Alert severity="info">HTTP serves the UI/API without TLS. Recommended only behind a reverse proxy or for local development.</Alert>
                        </VStack>
                    </CardContent>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="contained" onClick={onSaveHttp}>Save</Button>
                    </CardActions>
                </Card>

                {/* HTTPS Block */}
                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={2} alignItems="flex-start">
                            <Typography variant="h6">HTTPS</Typography>
                            <VStack spacing={2} alignItems={"flex-start"} flexWrap={"wrap"}>
                                <HStack spacing={1} sx={{alignItems: 'center'}}>
                                    <Switch checked={!!https.enabled} onChange={(e) => setHttps(s => ({...s, enabled: e.target.checked}))}/>
                                    <Typography>Enable HTTPS</Typography>
                                </HStack>
                                <TextField
                                    label="Port"
                                    value={https.port}
                                    onChange={(e) => setHttps(s => ({...s, port: e.target.value.replace(/[^0-9]/g, '')}))}
                                    sx={{minWidth: {xs: '100%', md: 160}}}
                                    slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                                    error={!!fieldErrors?.httpsPort}
                                    helperText={fieldErrors?.httpsPort}
                                />
                                <FormControl sx={{minWidth: {xs: '100%', md: 240}}}>
                                    <InputLabel id="https-mode-label">Certificate Mode</InputLabel>
                                    <Select
                                        labelId="https-mode-label"
                                        label="Mode"
                                        value={https.mode}
                                        onChange={(e) => setHttps(s => ({...s, mode: e.target.value === 'upload' ? 'upload' : 'selfsigned'}))}
                                    >
                                        <MenuItem value="selfsigned">Self Signed</MenuItem>
                                        <MenuItem value="upload">Upload</MenuItem>
                                    </Select>
                                </FormControl>
                            </VStack>
                            {https.mode === 'upload' && (
                                <Alert severity="info">
                                    Provide your TLS certificate and private key in PEM format.
                                    {/* When a pair is already uploaded on the server, hide selectors and show cert info */}
                                    {hasServerUploadedPair ? (
                                        <VStack spacing={1} sx={{mt: 1}}>
                                            <Typography variant="subtitle1">Certificate Installed</Typography>
                                            <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                                                <Typography sx={{minWidth: 140, color: 'text.secondary'}}>Subject</Typography>
                                                <Typography sx={{flex: 1}}>{https.certInfo?.subject || '—'}</Typography>
                                            </HStack>
                                            <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                                                <Typography sx={{minWidth: 140, color: 'text.secondary'}}>Issuer</Typography>
                                                <Typography sx={{flex: 1}}>{https.certInfo?.issuer || '—'}</Typography>
                                            </HStack>
                                            <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                                                <Typography sx={{minWidth: 140, color: 'text.secondary'}}>Validity Period</Typography>
                                                <Typography sx={{flex: 1}}>{https.certInfo?.validity || '—'}</Typography>
                                            </HStack>
                                        </VStack>
                                    ) : (
                                        <VStack spacing={1} sx={{mt: 1}}>
                                            <HStack spacing={1} sx={{flexWrap: 'wrap', alignItems: 'center'}}>
                                                <Button variant="outlined" onClick={() => onSelect('cert')}>Select Certificate</Button>
                                                <FormHelperText>Certificate: {selectedCert?.name || 'not selected'}</FormHelperText>
                                            </HStack>
                                            <HStack spacing={1} sx={{flexWrap: 'wrap', alignItems: 'center'}}>
                                                <Button variant="outlined" onClick={() => onSelect('key')}>Select Key</Button>
                                                <FormHelperText>Key: {selectedKey?.name || 'not selected'}</FormHelperText>
                                            </HStack>
                                            {!!fieldErrors?.certKeyPair && (
                                                <FormHelperText error>{fieldErrors.certKeyPair}</FormHelperText>
                                            )}
                                        </VStack>
                                    )}
                                </Alert>
                            )}
                        </VStack>
                    </CardContent>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        {hasServerUploadedPair && (
                            <Button color="error" onClick={() => setConfirmRemoveOpen(true)}>Remove Certificates</Button>
                        )}
                        <Button variant="contained" onClick={onSaveHttps} disabled={!canSaveHttps}>Save</Button>
                    </CardActions>
                </Card>
                
                {/* Agent Block */}
                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={2} alignItems="flex-start">
                            <Typography variant="h6">Agent Connection</Typography>
                            <VStack spacing={2} alignItems={"flex-start"} flexWrap={"wrap"}>
                                <HStack spacing={1} alignItems={"center"}>
                                    <Switch checked={!!agent.enabled} onChange={(e) => setAgent((s: GlobalAgentSettings) => ({...s, enabled: e.target.checked}))}/>
                                    <Typography>Enable Agent Connections</Typography>
                                </HStack>
                                <TextField
                                    label="Port"
                                    value={agent.port}
                                    onChange={(e) => setAgent((s: GlobalAgentSettings) => ({...s, port: e.target.value.replace(/[^0-9]/g, '')}))}
                                    sx={{minWidth: {xs: '100%', md: 160}}}
                                    slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                                    error={!!fieldErrors?.agentPort}
                                    helperText={fieldErrors?.agentPort}
                                />
                            </VStack>
                            <Alert severity="info">Agent connections are always secure (TLS). They use the same certificate/key as HTTPS when in Upload mode, or a self-signed certificate otherwise.</Alert>
                        </VStack>
                    </CardContent>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="contained" onClick={onSaveAgent}>Save</Button>
                    </CardActions>
                </Card>

                <Dialog open={confirmRemoveOpen} onClose={() => setConfirmRemoveOpen(false)}>
                    <DialogTitle>Remove certificates?</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            This will remove the currently installed TLS certificate and private key from the server. HTTPS will require a new certificate/key to be uploaded before it can use the Upload mode again. Do you want to proceed?
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={() => setConfirmRemoveOpen(false)}>Cancel</Button>
                        <Button color="error" variant="contained" onClick={onConfirmRemove}>Remove Certificates</Button>
                    </DialogActions>
                </Dialog>

                <Dialog open={restartPromptOpen} onClose={() => setRestartPromptOpen(false)}>
                    <DialogTitle>Restart listeners required</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            The HTTP/HTTPS listeners must be restarted for these network settings to take effect. Do you want to restart them now?
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={() => setRestartPromptOpen(false)}>Later</Button>
                        <Button variant="contained" color="warning" disabled={restarting} onClick={async () => {
                            try {
                                setRestarting(true);
                                await restartNetworkListeners();
                                setSnack({open: true, message: 'Restarting HTTP/HTTPS listeners…', severity: 'info'});
                            } catch (e) {
                                console.error(e);
                                setSnack({open: true, message: 'Failed to restart HTTP/HTTPS listeners.', severity: 'error'});
                            } finally {
                                setRestarting(false);
                                setRestartPromptOpen(false);
                            }
                        }}>Restart Now</Button>
                    </DialogActions>
                </Dialog>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack(s => ({...s, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack(s => ({...s, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
