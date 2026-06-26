import {useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, IconButton, InputAdornment, Snackbar, TextField, Typography} from '@mui/material';
import {alpha, useTheme} from '@mui/material/styles';
import ChronixLettersSvg from '../../assets/Chronix-letters.svg?react';
import ChronixGearMark from '../../assets/Chronix-gears.png';
import ThemeToggle from '../../site/themes/ThemeToggle';
import {apiPost} from "@dsherwin/react-api-interface";
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import {useAuthContext} from "../../data/useAuthContext.ts";
import {HStack, VStack} from "@dsherwin/mui-kit";

interface InitPayload {
    serverUrl: string;
    name: string;
    email: string;
    password: string;
}

const isEmail = (v: string) => /.+@.+\..+/.test(v);

export const InitialSetup = () => {
    const theme = useTheme();
    const [serverUrl, setServerUrl] = useState('');
    const [adminName, setAdminName] = useState('');
    const [adminEmail, setAdminEmail] = useState('');
    const [adminPassword, setAdminPassword] = useState('');
    const [adminPassword2, setAdminPassword2] = useState('');
    const [loading, setLoading] = useState(false);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    const [showPassword, setShowPassword] = useState(false);
    const [showPassword2, setShowPassword2] = useState(false);
    const [recoDialogOpen, setRecoDialogOpen] = useState(false);
    const [recommendation, setRecommendation] = useState<'http' | 'https' | ''>('');
    const {logout} = useAuthContext();

    const validate = (): string | null => {
        if (!serverUrl.trim()) return 'Server URL is required.';
        if (!/^https?:\/\//i.test(serverUrl)) return 'Server URL must start with http:// or https://';
        if (!adminName.trim()) return 'Admin name is required.';
        if (!adminEmail.trim()) return 'Admin email is required.';
        if (!isEmail(adminEmail)) return 'Enter a valid email address.';
        if (!adminPassword) return 'Admin password is required.';
        if (adminPassword.length < 8) return 'Password must be at least 8 characters.';
        if (adminPassword !== adminPassword2) return 'Passwords do not match.';
        return null;
    };

    const onFinalize = async (disable: 'http' | 'https' | 'none') => {
        setRecoDialogOpen(false);
        setLoading(true);
        try {
            await apiPost('/initialize/finalize', {disable});
            setSnack({open: true, message: 'Setup complete. Reloading…', severity: 'success'});
            setTimeout(() => logout(true), 800);
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Failed to complete setup.', severity: 'error'});
            setLoading(false);
        }
    };

    const onSubmit = async () => {
        const err = validate();
        if (err) {
            setSnack({open: true, message: err, severity: 'error'});
            return;
        }
        setLoading(true);
        try {
            const payload: InitPayload = {serverUrl: serverUrl.trim(), name: adminName.trim(), email: adminEmail.trim(), password: adminPassword};
            const data: any = await apiPost('/initialize', payload);
            if (data?.recommendation) {
                setRecommendation(data.recommendation);
                setRecoDialogOpen(true);
                setLoading(false);
            } else {
                setSnack({open: true, message: 'Setup complete. Reloading…', severity: 'success'});
                setTimeout(() => logout(true), 800);
            }
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Failed to complete setup.', severity: 'error'});
            setLoading(false);
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 4, width: '100%', position: 'relative', minHeight: '100vh', display: 'flex', alignItems: 'center'}}>
            <Box sx={{position: 'fixed', top: 12, right: 12}}>
                <ThemeToggle/>
            </Box>
            <VStack spacing={2} sx={{maxWidth: 580, width: '100%', mx: 'auto'}}>
                <VStack spacing={1} sx={{alignItems: 'center', textAlign: 'center'}}>
                    <Box
                        component="img"
                        src={ChronixGearMark}
                        alt="Chronix mark"
                        sx={{
                            width: 82,
                            height: 82,
                            borderRadius: 4,
                            objectFit: 'cover',
                            border: `1px solid ${alpha(theme.palette.common.white, 0.14)}`,
                            boxShadow: `0 12px 26px ${alpha(theme.palette.primary.main, 0.18)}`,
                        }}
                    />
                    <Box
                        sx={{
                            color: theme.palette.common.white,
                            '& svg': {display: 'block'},
                            '& svg path': {fill: 'currentColor'},
                        }}
                    >
                        <ChronixLettersSvg style={{display: 'block', width: 214, height: 32}}/>
                    </Box>
                    <Typography variant="h5">Initial Chronix Setup</Typography>
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Provide the server URL and create the first admin user</Typography>
                </VStack>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={2}>
                            <TextField label="Server URL" value={serverUrl} onChange={(e) => setServerUrl(e.target.value)} placeholder="https://chronix.example.com" fullWidth required/>
                            <Divider flexItem sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                            <Typography variant="subtitle1" sx={{
                                fontWeight: 600
                            }}>Primary Administrator</Typography>
                            <HStack spacing={2} sx={{flexWrap: 'nowrap'}}>
                                <TextField label="Full name" value={adminName} onChange={(e) => setAdminName(e.target.value)} sx={{minWidth: {xs: '100%', md: 260}}} fullWidth required/>
                                <TextField label="Email" value={adminEmail} onChange={(e) => setAdminEmail(e.target.value)} sx={{minWidth: {xs: '100%', md: 260}}} fullWidth required/>
                            </HStack>
                            <HStack spacing={2} sx={{flexWrap: 'nowrap'}}>
                                <TextField
                                    type={showPassword ? 'text' : 'password'}
                                    label="Password"
                                    value={adminPassword}
                                    onChange={(e) => setAdminPassword(e.target.value)}
                                    sx={{minWidth: {xs: '100%', md: 260}}}
                                    fullWidth
                                    required
                                    slotProps={{
                                        input: {
                                            endAdornment: (
                                                <InputAdornment position="end">
                                                    <IconButton
                                                        aria-label="toggle password visibility"
                                                        onClick={() => setShowPassword(v => !v)}
                                                        edge="end"
                                                        size="small"
                                                    >
                                                        {showPassword ? <VisibilityOff/> : <Visibility/>}
                                                    </IconButton>
                                                </InputAdornment>
                                            )
                                        }
                                    }}
                                />
                                <TextField
                                    type={showPassword2 ? 'text' : 'password'}
                                    label="Confirm password"
                                    value={adminPassword2}
                                    onChange={(e) => setAdminPassword2(e.target.value)}
                                    sx={{minWidth: {xs: '100%', md: 260}}}
                                    fullWidth
                                    required
                                    slotProps={{
                                        input: {
                                            endAdornment: (
                                                <InputAdornment position="end">
                                                    <IconButton
                                                        aria-label="toggle confirm password visibility"
                                                        onClick={() => setShowPassword2(v => !v)}
                                                        edge="end"
                                                        size="small"
                                                    >
                                                        {showPassword2 ? <VisibilityOff/> : <Visibility/>}
                                                    </IconButton>
                                                </InputAdornment>
                                            )
                                        }
                                    }}
                                />
                            </HStack>
                            <Alert severity="info">These settings are required to complete initial setup. You can change details later in Global Settings.</Alert>
                        </VStack>
                    </CardContent>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="contained" onClick={onSubmit} disabled={loading}>{loading ? 'Saving…' : 'Complete Setup'}</Button>
                    </CardActions>
                </Card>

                <Snackbar open={snack.open} autoHideDuration={3500} onClose={() => setSnack(s => ({...s, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack(s => ({...s, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>

                <Dialog open={recoDialogOpen} onClose={() => {}}>
                    <DialogTitle>Recommended Security Optimization</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            Initialization complete! For ease of setup, both HTTP and HTTPS were enabled. However, for security, it is recommended to only use one protocol in production.
                        </DialogContentText>
                        <DialogContentText sx={{mt: 2, fontWeight: 'bold'}}>
                            Since your Server URL uses {recommendation === 'http' ? 'HTTPS' : 'HTTP'}, we recommend turning off {recommendation.toUpperCase()} now.
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions sx={{px: 3, pb: 2}}>
                        <Button onClick={() => onFinalize('none')} color="inherit">Keep Both Enabled</Button>
                        <Button onClick={() => onFinalize(recommendation as 'http' | 'https')} variant="contained" color="primary" autoFocus>
                            Turn off {recommendation.toUpperCase()}
                        </Button>
                    </DialogActions>
                </Dialog>
            </VStack>
        </Box>
    );
};

export default InitialSetup;
