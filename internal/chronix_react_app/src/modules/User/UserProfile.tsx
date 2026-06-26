import {useEffect, useState} from 'react';
import {useLocation, useNavigate} from 'react-router';
import {Alert, Box, Button, Card, CardActions, CardContent, Dialog, DialogActions, DialogContent, DialogTitle, Divider, IconButton, InputAdornment, MenuItem, Snackbar, Stack, Switch, TextField, Typography} from '@mui/material';
import {Visibility, VisibilityOff} from '@mui/icons-material';
import {type UserProfileData} from './types';
import {apiGet, apiPut} from '@dsherwin/react-api-interface'
import {useAuthContext} from '../../data/useAuthContext'
import {useSettings} from '../../data/SettingsContext'
import {HStack, VStack} from "@dsherwin/mui-kit";

function supportedTimeZones(): string[] {
    // TS may not know about supportedValuesOf; guard usage
    const anyIntl: any = Intl as any;
    if (anyIntl && typeof anyIntl.supportedValuesOf === 'function') {
        try {
            const vals = anyIntl.supportedValuesOf('timeZone');
            if (Array.isArray(vals) && vals.length > 0) return vals as string[];
        } catch {/* ignore */
        }
    }
    // Fallback minimal list + current browser TZ at top
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
    const base = ['UTC', 'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles', 'Europe/London', 'Europe/Berlin', 'Asia/Tokyo', 'Asia/Singapore', 'Australia/Sydney'];
    const set = new Set([tz, ...base]);
    return Array.from(set);
}

type DisplaySettingsValue = { timeZone: string; timeFormat: '12h' | '24h' };
const DisplaySettingsCard = ({value, onChange}: { value?: DisplaySettingsValue; onChange?: (v: DisplaySettingsValue) => void }) => {
    const {timeFormat, timeZone, setTimeFormat, setTimeZone} = useSettings();
    const v = value ?? {timeZone, timeFormat};
    const change = onChange ?? ((nv: DisplaySettingsValue) => {
        setTimeZone(nv.timeZone);
        setTimeFormat(nv.timeFormat);
    });
    const tzOptions = supportedTimeZones();
    return (
        <Card variant="outlined" sx={{borderRadius: 3}}>
            <CardContent>
                <Typography variant="h6" gutterBottom>Display Settings</Typography>
                <Typography
                    variant="body2"
                    gutterBottom
                    sx={{
                        color: "text.secondary",
                        mb: 2
                    }}>
                    Choose how dates and times are shown in the UI. Server timestamps are stored in UTC; your selection only affects local display.
                </Typography>
                <VStack spacing={2} sx={{mt: 1, maxWidth: 520}}>
                    <TextField
                        select
                        label="Time zone"
                        value={v.timeZone}
                        onChange={(e) => change({...v, timeZone: e.target.value})}
                        fullWidth
                    >
                        {tzOptions.map(tz => (
                            <MenuItem key={tz} value={tz}>{tz}</MenuItem>
                        ))}
                    </TextField>
                    <Stack direction="row" spacing={2} sx={{
                        alignItems: "center"
                    }}>
                        <Typography sx={{minWidth: 160}}>Time format</Typography>
                        <Stack direction="row" spacing={1} sx={{
                            alignItems: "center"
                        }}>
                            <Typography variant="body2">24-hour</Typography>
                            <Switch
                                checked={v.timeFormat === '12h'}
                                onChange={(e) => change({...v, timeFormat: e.target.checked ? '12h' : '24h'})}
                                slotProps={{
                                    input: {
                                        'aria-label': 'toggle 12 hour time',
                                    },
                                }}
                            />
                            <Typography variant="body2">12-hour</Typography>
                        </Stack>
                    </Stack>
                </VStack>
            </CardContent>
        </Card>
    );
};

export const UserProfile = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const [loading, setLoading] = useState<boolean>(true);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});

    const [profile, setProfile] = useState<UserProfileData>({id: '', displayName: '', email: '', phone: '', timeFormat: "12h", timeZone: "UTC"});
    const [emailAvailable, setEmailAvailable] = useState<boolean | undefined>(undefined);
    const [originalEmail, setOriginalEmail] = useState<string>('');
    const [pwDialogOpen, setPwDialogOpen] = useState<boolean>(false);
    const [currentPw, setCurrentPw] = useState<string>('');
    const [newPw, setNewPw] = useState<string>('');
    const [verifyPw, setVerifyPw] = useState<string>('');
    const [showCurrent, setShowCurrent] = useState<boolean>(false);
    const [showNew, setShowNew] = useState<boolean>(false);
    const [showVerify, setShowVerify] = useState<boolean>(false);

    const {setUser, setLoggedIn} = useAuthContext()
    const {timeFormat, timeZone} = useSettings();
    const [displaySettings, setDisplaySettings] = useState<DisplaySettingsValue>({timeZone, timeFormat});

    // Keep local display settings initialized from context when component mounts/when context changes
    useEffect(() => {
        setDisplaySettings({timeZone, timeFormat});
    }, [timeZone, timeFormat]);

    useEffect(() => {
        const search = new URLSearchParams(location.search)
        if (search.get('forceChange') === '1') {
            setPwDialogOpen(true)
        }
        const load = async () => {
            setLoading(true);
            setLoadError(null);
            try {
                const me = await apiGet('/me') as { id: number; name: string; email: string; phone?: string; timeZone?: string; timeFormat?: '12h' | '24h' }
                const p = {id: String(me.id), displayName: me.name, email: me.email, phone: me.phone ?? '', timeZone: me.timeZone ?? 'UTC', timeFormat: me.timeFormat ?? '12h'};
                setProfile(p)
                setOriginalEmail(me.email)
            } catch (e) {
                console.error(e);
                setLoadError('Failed to load profile')
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [location.search]);

    useEffect(() => {
        const e = profile.email.trim();
        if (!e) {
            setEmailAvailable(undefined);
            return;
        }
        if (originalEmail && e.toLowerCase() === originalEmail.trim().toLowerCase()) {
            setEmailAvailable(true);
            return;
        }
        const t = window.setTimeout(async () => {
            try {
                const res = await apiGet(`/me/check-email?email=${encodeURIComponent(e)}${profile.id ? `&excludeId=${encodeURIComponent(profile.id)}` : ''}`) as { available: boolean }
                setEmailAvailable(res.available)
            } catch {
                setEmailAvailable(undefined)
            }
        }, 400)
        return () => {
            window.clearTimeout(t)
        }
    }, [profile.email, profile.id, originalEmail])

    const onSave = async () => {
        if (!profile.displayName.trim() || !profile.email.trim()) {
            setSnack({open: true, message: 'Display name and email are required.', severity: 'error'});
            return;
        }
        if (emailAvailable === false) {
            setSnack({open: true, message: 'Email is already in use.', severity: 'error'});
            return;
        }
        try {
            // Save core profile fields
            await apiPut('/me', {
                id: Number(profile.id),
                name: profile.displayName,
                email: profile.email,
                phone: profile.phone ?? undefined,
                timeZone: displaySettings.timeZone,
                timeFormat: displaySettings.timeFormat,
            });
            // Refresh current user info after save
            try {
                const me = await apiGet('/me') as any;
                if (me && typeof me === 'object' && 'id' in me) {
                    setUser({
                        id: me.id,
                        name: me.name,
                        email: me.email,
                        phone: me.phone ?? '',
                        admin: me.admin ?? false,
                        forcePasswordChange: me.forcePasswordChange ?? false,
                        timeZone: me.timeZone,
                        timeFormat: me.timeFormat,
                    });
                }
            } catch {/* ignore */
            }
            setSnack({open: true, message: 'Profile saved.', severity: 'success'});
            // Navigate back to previous view after successful save
            navigate(-1);
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Failed to save profile.', severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 800, width: '100%', mx: 'auto'}}>
                <Typography variant="h5">My Profile</Typography>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {loadError && (<Alert severity="error">{loadError}</Alert>)}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        {loading ? (
                            <Typography variant="body2" sx={{
                                color: "text.secondary"
                            }}>Loading…</Typography>
                        ) : (
                            <VStack spacing={2}>
                                <TextField label="Display name" value={profile.displayName} onChange={(e) => setProfile(p => ({...p, displayName: e.target.value}))} fullWidth/>
                                <TextField
                                    label="Email"
                                    value={profile.email}
                                    onChange={(e) => setProfile(p => ({...p, email: e.target.value}))}
                                    fullWidth
                                    error={profile.email.trim() !== '' && emailAvailable === false}
                                    helperText={profile.email.trim() !== '' && emailAvailable === false ? 'Email is already in use' : ''}
                                />
                                <TextField label="Phone (optional)" placeholder="e.g., +1 555 123 4567" value={profile.phone ?? ''} onChange={(e) => setProfile(p => ({...p, phone: e.target.value}))} fullWidth/>
                            </VStack>
                        )}
                    </CardContent>
                </Card>

                {/* Display Settings */}
                <DisplaySettingsCard value={displaySettings} onChange={setDisplaySettings}/>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        <HStack width={"100%"} justifyContent={"space-between"}>
                            <Button variant="outlined" onClick={() => setPwDialogOpen(true)} sx={{alignSelf: 'flex-start'}}>Change Password</Button>
                            <HStack spacing={2}>
                                <Button variant="outlined" color="error" onClick={() => window.history.back()}>Cancel</Button>
                                <Button variant="contained" onClick={onSave}>Save</Button>
                            </HStack>
                        </HStack>
                    </CardActions>
                </Card>

                <Dialog open={pwDialogOpen} onClose={(_e, reason) => {
                    if (reason === "backdropClick" || reason === "escapeKeyDown") return;
                    setPwDialogOpen(false);
                }}>
                    <DialogTitle sx={{paddingBottom: 0}}>Change Password</DialogTitle>
                    <DialogTitle sx={{fontWeight: 'normal', fontSize: '0.875rem', color: 'text.secondary', paddingTop: 0}}>Must be at least 8 characters in length.</DialogTitle>
                    <DialogContent sx={{pt: 1}}>
                        <VStack spacing={2} sx={{mt: 1, minWidth: {xs: 280, sm: 360}}}>
                            <TextField
                                label="Current password"
                                type={showCurrent ? 'text' : 'password'}
                                value={currentPw}
                                onChange={(e) => setCurrentPw(e.target.value)}
                                fullWidth
                                slotProps={{
                                    input: {
                                        endAdornment: (
                                            <InputAdornment position="end">
                                                <IconButton onClick={() => setShowCurrent(s => !s)} aria-label="toggle current password visibility">
                                                    {showCurrent ? <VisibilityOff/> : <Visibility/>}
                                                </IconButton>
                                            </InputAdornment>
                                        )
                                    }
                                }}
                            />
                            <TextField
                                label="New password"
                                type={showNew ? 'text' : 'password'}
                                value={newPw}
                                error={!!newPw && newPw.length < 8}
                                onChange={(e) => setNewPw(e.target.value)}
                                fullWidth
                                slotProps={{
                                    input: {
                                        endAdornment: (
                                            <InputAdornment position="end">
                                                <IconButton onClick={() => setShowNew(s => !s)} aria-label="toggle new password visibility">
                                                    {showNew ? <VisibilityOff/> : <Visibility/>}
                                                </IconButton>
                                            </InputAdornment>
                                        )
                                    }
                                }}
                            />
                            <TextField
                                label="Verify password"
                                type={showVerify ? 'text' : 'password'}
                                value={verifyPw}
                                onChange={(e) => setVerifyPw(e.target.value)}
                                error={!!newPw && !!verifyPw && (newPw !== verifyPw || verifyPw.length < 8)}
                                helperText={!!newPw && !!verifyPw && newPw !== verifyPw ? 'Passwords do not match' : ''}
                                fullWidth
                                slotProps={{
                                    input: {
                                        endAdornment: (
                                            <InputAdornment position="end">
                                                <IconButton onClick={() => setShowVerify(s => !s)} aria-label="toggle verify password visibility">
                                                    {showVerify ? <VisibilityOff/> : <Visibility/>}
                                                </IconButton>
                                            </InputAdornment>
                                        )
                                    }
                                }}
                            />
                        </VStack>
                    </DialogContent>
                    <DialogActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="outlined" color="error" onClick={() => {
                            setPwDialogOpen(false);
                            setCurrentPw('');
                            setNewPw('');
                            setVerifyPw('');
                        }}>Cancel</Button>
                        <Button variant="contained" onClick={async () => {
                            if (!currentPw || !newPw || !verifyPw) {
                                setSnack({open: true, message: 'Please fill in all password fields.', severity: 'error'});
                                return;
                            }
                            if (newPw.length < 8) {
                                setSnack({open: true, message: 'Password must be at least 8 characters.', severity: 'error'});
                                return;
                            }
                            if (newPw !== verifyPw) {
                                setSnack({open: true, message: 'New and Verify passwords must match.', severity: 'error'});
                                return;
                            }
                            try {
                                const res = await apiPut('/me/password', {currentPassword: currentPw, newPassword: newPw})
                                if ((res as any).ok === false) throw new Error('Change password failed')
                                setSnack({open: true, message: 'Password changed.', severity: 'success'});
                                setPwDialogOpen(false);
                                setCurrentPw('');
                                setNewPw('');
                                setVerifyPw('');
                                const search = new URLSearchParams(location.search)
                                if (search.get('forceChange') === '1') {
                                    // backend logs us out; ensure client state reflects and go to login
                                    setLoggedIn(false)
                                    setUser(null)
                                    navigate('/')
                                }
                            } catch (e) {
                                console.error(e);
                                setSnack({open: true, message: 'Failed to change password.', severity: 'error'});
                            }
                        }}>Save</Button>
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
