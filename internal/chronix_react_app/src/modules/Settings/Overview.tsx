import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, Grid, IconButton, Typography} from '@mui/material';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {BugReport, Close, Dns, Email, Http, Https, Lightbulb, Sms} from '@mui/icons-material';
import {DataGrid, type GridColDef} from "@mui/x-data-grid";
import {SectionHelp} from '../../main/SectionHelp';
import {useNavigate, useSearchParams} from 'react-router';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext';
import {useUpdateAvailability} from '../../lib/useUpdateAvailability';
import {fetchActiveSettings, fetchSettingsSummary, restartServer, shutdownServer} from './api.ts';
import type {ActiveSetting, SettingsSummary} from './types.ts';

export const AdminOverview = () => {
    const navigate = useNavigate();
    const {data: featureData} = useFeatureAvailability();
    const [data, setData] = useState<SettingsSummary | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [confirmRestart, setConfirmRestart] = useState(false);
    const [confirmShutdown, setConfirmShutdown] = useState(false);
    const [busy, setBusy] = useState<{ restarting?: boolean; shuttingDown?: boolean }>({});
    const [activeSettings, setActiveSettings] = useState<ActiveSetting[] | null>(null);
    const {status: updaterStatus, hasServerUpdateNotice, hasAgentUpdateNotice, shouldShowUpdateNotice, dismissUpdateNotice} = useUpdateAvailability();
    const [params] = useSearchParams()

    // URL param-backed state (keep consistent with Runs grid patterns)
    const limit = Math.min(Number(params.get('limit') || 25), 100)
    const offset = Number(params.get('offset') || 0)

    useEffect(() => {
        (async () => {
            try {
                setData(await fetchSettingsSummary());
            } catch (e) {
                console.error(e);
                setError('Failed to load summary.');
            }
            try {
                setActiveSettings(await fetchActiveSettings());
            } catch (e) {
                console.error(e);
                // Don't fail the page if settings list fails; just show an inline message below.
                setActiveSettings([]);
            }
        })();
    }, []);

    const doRestart = async () => {
        setBusy(s => ({...s, restarting: true}));
        try {
            await restartServer();
            setConfirmRestart(false);
        } catch (e) {
            console.error(e);
            setError('Failed to restart server.');
        } finally {
            setBusy(s => ({...s, restarting: false}));
        }
    };

    const doShutdown = async () => {
        setBusy(s => ({...s, shuttingDown: true}));
        try {
            await shutdownServer();
            setConfirmShutdown(false);
        } catch (e) {
            console.error(e);
            setError('Failed to shutdown server.');
        } finally {
            setBusy(s => ({...s, shuttingDown: false}));
        }
    };

    const settingsColumns: GridColDef[] = [
        {field: 'setting', headerName: 'Setting', width: 250, valueGetter: (_v, r) => r.setting.replace(/_/g, ' ')},
        {field: 'value', headerName: 'Value', width: 200, valueGetter: (_v, r) => r.value || '—'},
        {field: 'description', headerName: 'Description', flex: 1},
    ];

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1100, width: '100%', mx: 'auto'}}>
                <Box sx={{display: 'flex', alignItems: 'center'}}>
                    <Typography variant="h5">Settings Overview</Typography>
                    <SectionHelp/>
                </Box>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                {error && <Alert severity="error">{error}</Alert>}
                {(hasServerUpdateNotice || hasAgentUpdateNotice) && shouldShowUpdateNotice && (
                    <Alert
                        severity="warning"
                        action={(
                            <HStack spacing={0.5} alignItems="flex-start">
                                <Button color="inherit" size="small" onClick={() => navigate('/settings/updates')}>Go To Updates</Button>
                                <IconButton
                                    aria-label="Dismiss update notice"
                                    color="inherit"
                                    size="small"
                                    onClick={dismissUpdateNotice}
                                    sx={{mt: '-4px'}}
                                >
                                    <Close fontSize="small"/>
                                </IconButton>
                            </HStack>
                        )}
                        sx={{width: '100%'}}
                    >
                        <VStack spacing={0.5} alignItems="flex-start">
                            {hasServerUpdateNotice && (
                                <Typography variant="body2">
                                    A Chronix app update is available{updaterStatus?.availableVersion?.server?.version ? `: ${updaterStatus.availableVersion.server.version}` : ''}.
                                </Typography>
                            )}
                            {hasAgentUpdateNotice && (
                                <Typography variant="body2">
                                    Agent updates are available and automatic agent updating is not enabled.
                                </Typography>
                            )}
                        </VStack>
                    </Alert>
                )}
                {!data ? (
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Loading…</Typography>
                ) : (
                    <Grid container spacing={2} sx={{
                        alignItems: "stretch"
                    }}>
                        <Grid size={{xs: 12, md: 6}}>
                            <Card variant="outlined" sx={{borderRadius: 3, height: '100%', display: 'flex', flexDirection: 'column'}}>
                                <CardContent sx={{flexGrow: 1}}>
                                    <HStack spacing={1} sx={{mb: 1}}>
                                        <Dns color="primary"/>
                                        <Typography variant="h6">Server</Typography>
                                    </HStack>
                                    <Typography variant="body2" sx={{
                                        color: "text.secondary"
                                    }}>Server URL</Typography>
                                    <Typography>{data.serverUrl || '—'}</Typography>
                                </CardContent>
                                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                <CardActions sx={{gap: 1}}>
                                    <Button variant="contained" color="warning" onClick={() => setConfirmRestart(true)} disabled={busy.restarting}>Restart Server</Button>
                                    <Button variant="outlined" color="error" onClick={() => setConfirmShutdown(true)} disabled={busy.shuttingDown}>Shutdown Server</Button>
                                </CardActions>
                            </Card>
                        </Grid>
                        <Grid size={{xs: 12, md: 6}}>
                            <Card variant="outlined" sx={{borderRadius: 3, height: '100%', display: 'flex', flexDirection: 'column'}}>
                                <CardContent sx={{flexGrow: 1}}>
                                    <HStack spacing={1} sx={{mb: 1}}>
                                        <Http color="primary"/>
                                        <Typography variant="h6">HTTP</Typography>
                                    </HStack>
                                    <Typography>Enabled: {data.http?.enabled ? 'Yes' : 'No'}</Typography>
                                    <Typography>Port: {data.http?.port || '—'}</Typography>
                                </CardContent>
                            </Card>
                        </Grid>
                        <Grid size={{xs: 12, md: 6}}>
                            <Card variant="outlined" sx={{borderRadius: 3, height: '100%', display: 'flex', flexDirection: 'column'}}>
                                <CardContent sx={{flexGrow: 1}}>
                                    <HStack spacing={1} sx={{mb: 1}}>
                                        <Https color="primary"/>
                                        <Typography variant="h6">HTTPS</Typography>
                                    </HStack>
                                    <VStack spacing={1}>
                                        <Typography>Enabled: {data.https?.enabled ? 'Yes' : 'No'}</Typography>
                                        <Typography>Port: {data.https?.port || '—'}</Typography>
                                        <Typography>Certificate Mode: {data.https?.mode}</Typography>
                                        {!!data.https?.certInfo && (
                                            <>
                                                <VStack spacing={0}>
                                                    <Typography sx={{
                                                        color: "text.secondary"
                                                    }}>Certificate Information</Typography>
                                                    <HStack gap={1}>
                                                        <Typography>Subject:</Typography>
                                                        <Typography>{data.https.certInfo.subject}</Typography>
                                                    </HStack>
                                                    <HStack gap={1}>
                                                        <Typography>Issuer:</Typography>
                                                        <Typography>{data.https.certInfo.issuer}</Typography>
                                                    </HStack>
                                                    <HStack gap={1}>
                                                        <Typography>Valid:</Typography>
                                                        <Typography>{data.https.certInfo.notBefore} — {data.https.certInfo.notAfter}</Typography>
                                                    </HStack>
                                                </VStack>
                                            </>
                                        )}

                                    </VStack>
                                </CardContent>
                            </Card>
                        </Grid>
                        <Grid size={{xs: 12, md: 6}}>
                            <Card variant="outlined" sx={{borderRadius: 3, height: '100%', display: 'flex', flexDirection: 'column'}}>
                                <CardContent sx={{flexGrow: 1}}>
                                    <HStack spacing={1} sx={{mb: 1}}>
                                        <Email color="primary"/>
                                        <Typography variant="h6">Email</Typography>
                                    </HStack>
                                    <Typography>Configured: {data.email?.configured ? 'Yes' : 'No'}</Typography>
                                    <Typography>Host: {data.email?.smtpHost || '—'}</Typography>
                                    <Typography>From: {data.email?.fromName ? `${data.email.fromName} <${data.email.fromEmail}>` : (data.email?.fromEmail || '—')}</Typography>
                                </CardContent>
                            </Card>
                        </Grid>
                        <Grid size={{xs: 12, md: 6}}>
                            <Card variant="outlined" sx={{borderRadius: 3, height: '100%', display: 'flex', flexDirection: 'column'}}>
                                <CardContent sx={{flexGrow: 1}}>
                                    <HStack spacing={1} sx={{mb: 1}}>
                                        <Sms color="primary"/>
                                        <Typography variant="h6">SMS</Typography>
                                    </HStack>
                                    <Typography>Configured: {data.sms?.configured ? 'Yes' : 'No'}</Typography>
                                    {!!data.sms?.configured && <Typography>Provider: {data.sms?.provider}</Typography>}
                                    {!!data.sms?.configured && <Typography>From: {data.sms?.fromNumber || '—'}</Typography>}
                                </CardContent>
                            </Card>
                        </Grid>
                        {featureData?.feedbackEnabled && (
                            <>
                                <Grid size={{xs: 12, md: 6}}>
                                    <Card variant="outlined" sx={{borderRadius: 3, height: '100%', display: 'flex', flexDirection: 'column'}}>
                                        <CardContent sx={{flexGrow: 1}}>
                                            <HStack spacing={1} sx={{mb: 1}}>
                                                <BugReport color="primary"/>
                                                <Typography variant="h6">Bug Reports</Typography>
                                            </HStack>
                                            <Typography variant="body2" sx={{
                                                color: "text.secondary"
                                            }}>Review submitted bug reports from users.</Typography>
                                        </CardContent>
                                        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                        <CardActions>
                                            <Button variant="outlined" onClick={() => navigate('/settings/bug-reports')}>View Reports</Button>
                                        </CardActions>
                                    </Card>
                                </Grid>
                                <Grid size={{xs: 12, md: 6}}>
                                    <Card variant="outlined" sx={{borderRadius: 3, height: '100%', display: 'flex', flexDirection: 'column'}}>
                                        <CardContent sx={{flexGrow: 1}}>
                                            <HStack spacing={1} sx={{mb: 1}}>
                                                <Lightbulb color="primary"/>
                                                <Typography variant="h6">Feature Requests</Typography>
                                            </HStack>
                                            <Typography variant="body2" sx={{
                                                color: "text.secondary"
                                            }}>Review feature requests submitted by users.</Typography>
                                        </CardContent>
                                        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                        <CardActions>
                                            <Button variant="outlined" onClick={() => navigate('/settings/feature-requests')}>View Requests</Button>
                                        </CardActions>
                                    </Card>
                                </Grid>
                            </>
                        )}
                    </Grid>
                )}

                {/* Active Settings table at the bottom, full width */}
                <Divider sx={{my: 2, borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                <Typography variant="h6">Advanced Settings</Typography>
                {activeSettings === null ? (
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Loading settings…</Typography>
                ) : activeSettings.length === 0 ? (
                    <Alert severity="info">No settings to display.</Alert>
                ) : (
                    <Card variant="outlined" sx={{borderRadius: 3}}>
                        <CardContent sx={{p: 0}}>
                            <div style={{
                                width: '100%'
                            }}>
                                <DataGrid
                                    rows={activeSettings}
                                    columns={settingsColumns}
                                    getRowId={(r) => r.setting}
                                    density="compact"
                                    disableRowSelectionOnClick
                                    paginationModel={{
                                        pageSize: limit,
                                        page: Math.floor(offset / Math.max(limit, 1))
                                    }}
                                    pageSizeOptions={[5, 10, 25, 50, 100]}
                                    pagination
                                />
                            </div>
                        </CardContent>
                    </Card>
                )}

                <Dialog open={confirmRestart} onClose={() => setConfirmRestart(false)}>
                    <DialogTitle>Restart Server</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            Are you sure you want to restart the server now? Active connections will be dropped briefly.
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={() => setConfirmRestart(false)}>Cancel</Button>
                        <Button color="warning" variant="contained" onClick={doRestart} disabled={busy.restarting}>Restart</Button>
                    </DialogActions>
                </Dialog>

                <Dialog open={confirmShutdown} onClose={() => setConfirmShutdown(false)}>
                    <DialogTitle>Shutdown Server</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            Are you sure you want to shutdown the server? You will need to manually start it again.
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={() => setConfirmShutdown(false)}>Cancel</Button>
                        <Button color="error" variant="contained" onClick={doShutdown} disabled={busy.shuttingDown}>Shutdown</Button>
                    </DialogActions>
                </Dialog>
            </VStack>
        </Box>
    );
};
