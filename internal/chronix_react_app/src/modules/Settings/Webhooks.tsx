import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardContent, Chip, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControl, FormControlLabel, IconButton, InputLabel, MenuItem, Select, Snackbar, Switch, TextField, Tooltip, Typography} from '@mui/material';
import {Add, Delete, Edit, NotificationsActive, Refresh, Send} from '@mui/icons-material';
import {HStack, useMuiPrompts, VStack} from "@dsherwin/mui-kit";
import {formatAPIError} from "../../lib/errors";
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';
import {createWebhook, deleteWebhook, fetchWebhooks, testWebhook, updateWebhook} from './api.ts';
import type {WebhookSettingsItem} from './types.ts';

const EVENT_OPTIONS = [
    { value: 'job', label: 'Job Events' },
    { value: 'connection', label: 'Connection Events' },
    { value: 'system', label: 'System Events' },
    { value: 'security', label: 'Security Events' },
    { value: 'worker', label: 'Worker Events' },
];

export const WebhooksPage = () => {
    const {confirmPrompt} = useMuiPrompts();
    const [items, setItems] = useState<WebhookSettingsItem[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({open: false, message: '', severity: 'success'});

    const [editDlg, setEditDlg] = useState<{ open: boolean; item?: WebhookSettingsItem }>({open: false});
    const [name, setName] = useState('');
    const [url, setUrl] = useState('');
    const [secret, setSecret] = useState('');
    const [selectedEvents, setSelectedEvents] = useState<string[]>([]);
    const [enabled, setEnabled] = useState(true);
    const [saving, setSaving] = useState(false);

    const load = async () => {
        setLoading(true);
        try {
            setItems(await fetchWebhooks());
            setError(null);
        } catch (e) {
            console.error(e);
            setError('Failed to load webhooks');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);

    const openEdit = (item?: WebhookSettingsItem) => {
        if (item) {
            setName(item.name);
            setUrl(item.url);
            setSecret(item.secret || '');
            setSelectedEvents(item.events ? item.events.split(',').map(e => e.trim()) : []);
            setEnabled(item.enabled);
        } else {
            setName('');
            setUrl('');
            setSecret('');
            setSelectedEvents(['job', 'connection']);
            setEnabled(true);
        }
        setEditDlg({open: true, item});
    };

    const handleSave = async () => {
        if (!name || !url) return;
        setSaving(true);
        try {
            const payload = {
                name,
                url,
                secret,
                events: selectedEvents.join(','),
                enabled
            };

            if (editDlg.item) {
                await updateWebhook(editDlg.item.id, payload);
                setSnack({open: true, message: 'Webhook updated', severity: 'success'});
            } else {
                await createWebhook(payload);
                setSnack({open: true, message: 'Webhook created', severity: 'success'});
            }
            setEditDlg({open: false});
            void load();
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: formatAPIError(e, 'Failed to save webhook'), severity: 'error'});
        } finally {
            setSaving(false);
        }
    };

    const confirmDelete = async (id: number, name: string) => {
        const ok = await confirmPrompt({
            title: 'Delete Webhook',
            message: `Are you sure you want to delete the webhook “${name}”? This action cannot be undone.`,
            buttonText: 'Delete',
            buttonColor: 'error'
        });
        if (!ok) return;

        try {
            await deleteWebhook(id);
            setSnack({open: true, message: 'Webhook deleted', severity: 'success'});
            void load();
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: formatAPIError(e, 'Failed to delete webhook'), severity: 'error'});
        }
    };

    const handleTest = async (item: WebhookSettingsItem) => {
        try {
            setSnack({open: true, message: 'Sending test ping...', severity: 'success'});
            const res = await testWebhook(item);
            setSnack({open: true, message: res.message || 'Test successful', severity: 'success'});
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: formatAPIError(e, 'Test failed'), severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Box sx={{display: 'flex', alignItems: 'center'}}>
                        <Typography variant="h5">Outgoing Webhooks</Typography>
                        <SectionHelp section={HELP_SECTIONS.notifications} />
                    </Box>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Tooltip title="Refresh"><IconButton onClick={load}><Refresh/></IconButton></Tooltip>
                        <Button startIcon={<Add/>} variant="contained" onClick={() => openEdit()}>New Webhook</Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {error && <Alert severity="error">{error}</Alert>}

                {loading && items.length === 0 ? (
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Loading webhooks…</Typography>
                ) : items.length === 0 ? (
                    <Card variant="outlined" sx={{p: 4, textAlign: 'center', borderRadius: 3}}>
                        <NotificationsActive sx={{fontSize: 48, color: 'text.disabled', mb: 2}}/>
                        <Typography variant="h6" sx={{
                            color: "text.secondary"
                        }}>No webhooks configured</Typography>
                        <Typography
                            variant="body2"
                            sx={{
                                color: "text.secondary",
                                mb: 3
                            }}>
                            Webhooks allow you to send real-time notifications to external services like Slack, Discord, or your own custom API.
                        </Typography>
                        <Button startIcon={<Add/>} variant="outlined" onClick={() => openEdit()}>Create your first webhook</Button>
                    </Card>
                ) : (
                    <VStack spacing={2}>
                        {items.map(wh => (
                            <Card key={wh.id} variant="outlined" sx={{borderRadius: 3}}>
                                <CardContent>
                                    <HStack justifyContent="space-between" alignItems="flex-start">
                                        <VStack spacing={0.5} sx={{flex: 1}}>
                                            <HStack spacing={1} alignItems="center">
                                                <Typography variant="subtitle1" sx={{
                                                    fontWeight: "bold"
                                                }}>{wh.name}</Typography>
                                                {!wh.enabled && <Chip label="Disabled" size="small" variant="outlined"/>}
                                            </HStack>
                                            <Typography
                                                variant="body2"
                                                sx={{
                                                    color: "text.secondary",
                                                    wordBreak: 'break-all'
                                                }}>{wh.url}</Typography>
                                            <HStack spacing={0.5} sx={{mt: 1, flexWrap: 'wrap'}}>
                                                {wh.events.split(',').map(e => (
                                                    <Chip key={e} label={e.trim()} size="small" variant="filled" color="primary"/>
                                                ))}
                                            </HStack>
                                        </VStack>
                                        <HStack>
                                            <Tooltip title="Test Webhook"><IconButton onClick={() => handleTest(wh)} color="primary"><Send fontSize="small"/></IconButton></Tooltip>
                                            <Tooltip title="Edit"><IconButton onClick={() => openEdit(wh)}><Edit fontSize="small"/></IconButton></Tooltip>
                                            <Tooltip title="Delete"><IconButton onClick={() => void confirmDelete(wh.id, wh.name)} color="error"><Delete fontSize="small"/></IconButton></Tooltip>
                                        </HStack>
                                    </HStack>
                                </CardContent>
                            </Card>
                        ))}
                    </VStack>
                )}

                {/* Create/Edit Dialog */}
                <Dialog open={editDlg.open} onClose={() => !saving && setEditDlg({open: false})} maxWidth="sm" fullWidth>
                    <DialogTitle>{editDlg.item ? 'Edit Webhook' : 'New Webhook'}</DialogTitle>
                    <DialogContent dividers>
                        <VStack spacing={2} sx={{mt: 1}}>
                            <TextField
                                label="Name"
                                fullWidth
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                placeholder="Slack Alerts"
                                required
                            />
                            <TextField
                                label="Payload URL"
                                fullWidth
                                value={url}
                                onChange={(e) => setUrl(e.target.value)}
                                placeholder="https://hooks.slack.com/services/..."
                                required
                            />
                            <TextField
                                label="Secret (Optional)"
                                fullWidth
                                type="password"
                                value={secret}
                                onChange={(e) => setSecret(e.target.value)}
                                helperText="If provided, we will sign the payload with this secret using HMAC-SHA256 (X-Chronix-Signature header)."
                            />

                            <FormControl fullWidth>
                                <InputLabel id="events-label">Subscribed Events</InputLabel>
                                <Select
                                    labelId="events-label"
                                    multiple
                                    value={selectedEvents}
                                    onChange={(e) => setSelectedEvents(typeof e.target.value === 'string' ? e.target.value.split(',') : e.target.value)}
                                    label="Subscribed Events"
                                    renderValue={(selected) => (
                                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                                            {selected.map((value) => (
                                                <Chip key={value} label={EVENT_OPTIONS.find(o => o.value === value)?.label || value} size="small" />
                                            ))}
                                        </Box>
                                    )}
                                >
                                    {EVENT_OPTIONS.map((opt) => (
                                        <MenuItem key={opt.value} value={opt.value}>
                                            {opt.label}
                                        </MenuItem>
                                    ))}
                                </Select>
                            </FormControl>

                            <FormControlLabel
                                control={<Switch checked={enabled} onChange={(e) => setEnabled(e.target.checked)}/>}
                                label="Enabled"
                            />
                        </VStack>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={() => setEditDlg({open: false})} disabled={saving}>Cancel</Button>
                        <Button variant="contained" onClick={handleSave} disabled={saving || !name || !url}>
                            {saving ? 'Saving...' : 'Save Webhook'}
                        </Button>
                    </DialogActions>
                </Dialog>


                <Snackbar
                    open={snack.open}
                    autoHideDuration={4000}
                    onClose={() => setSnack(s => ({...s, open: false}))}
                    anchorOrigin={{vertical: 'top', horizontal: 'center'}}
                >
                    <Alert severity={snack.severity} variant="filled" onClose={() => setSnack(s => ({...s, open: false}))}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
