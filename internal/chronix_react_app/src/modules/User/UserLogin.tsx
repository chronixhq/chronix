import React, {useRef, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, IconButton, InputAdornment, Snackbar, TextField, Typography} from '@mui/material';
import {alpha, useTheme} from '@mui/material/styles';
import {useNavigate} from 'react-router';
import ThemeToggle from '../../site/themes/ThemeToggle.tsx';
import ChronixLettersSvg from '../../assets/Chronix-letters.svg?react';
import ChronixGearMark from '../../assets/Chronix-gears.png';
import pkg from '../../../package.json';
import {apiGet, apiPost} from "@dsherwin/react-api-interface";
import {useAuthContext} from "../../data/useAuthContext.ts";
import {Visibility, VisibilityOff} from '@mui/icons-material';
import {HStack, VStack} from "@dsherwin/mui-kit";

export const UserLogin = () => {
    const theme = useTheme();
    const navigate = useNavigate();
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [showPw, setShowPw] = useState(false);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    const [loading, setLoading] = useState(false);
    const {setLoggedIn} = useAuthContext();
    const passwordRef = useRef<HTMLInputElement>(null);

    // Show stored logout reason if present (set by SSE-triggered logout)
    React.useEffect(() => {
        try {
            const reason = window.localStorage.getItem('logoutReason');
            if (reason) {
                setSnack({ open: true, message: reason, severity: 'info' });
                window.localStorage.removeItem('logoutReason');
            }
        } catch {}
    }, []);

    const onLogin = async () => {
        if (!email.trim() || !password.trim()) {
            setSnack({open: true, message: 'Enter email and password.', severity: 'error'});
            return;
        }
        setLoading(true);
        try {
            const res: { status: string } = await apiPost('/auth/login', {email, password});
            if (res.status !== 'logged in') throw new Error('Login failed');
            setSnack({open: true, message: 'Login successful.', severity: 'success'});
            setLoggedIn(true);
            try {
                const me = await apiGet('/me') as { forcePasswordChange?: boolean }
                if (me && (me as any).forcePasswordChange) {
                    navigate('/user/profile?forceChange=1');
                } else {
                    navigate('/');
                }
            } catch {
                navigate('/');
            }
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Invalid credentials.', severity: 'error'});
        } finally {
            setLoading(false);
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 4, width: '100%', position: 'relative', minHeight: '100vh', display: 'flex', alignItems: 'center'}}>
            {/* Theme toggle in corner */}
            <Box sx={{position: 'fixed', top: 12, right: 12}}>
                <ThemeToggle/>
            </Box>
            <VStack spacing={2} sx={{maxWidth: 480, width: '100%', mx: 'auto'}}>
                {/* Logo + Title */}
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
                    <Typography variant="h5">Sign In</Typography>
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Welcome to Chronix — please sign in</Typography>
                </VStack>

                <Card variant="outlined" sx={{borderRadius: 2.5}}>
                    <CardContent>
                        <VStack spacing={2}>
                            <TextField label="Email" value={email} onChange={(e) => setEmail(e.target.value)} fullWidth onKeyDown={(e)=>{ if(e.key==='Enter'){ e.preventDefault(); passwordRef.current?.focus(); } }} />
                            <TextField label="Password" type={showPw ? 'text' : 'password'} value={password} onChange={(e) => setPassword(e.target.value)} fullWidth inputRef={passwordRef} onKeyDown={(e)=>{ if(e.key==='Enter'){ e.preventDefault(); onLogin(); } }} 
                                       slotProps={{
                                           input: {
                                               endAdornment: (
                                                   <InputAdornment position="end">
                                                       <IconButton aria-label="toggle password visibility" onClick={() => setShowPw(v=>!v)} edge="end" size="small">
                                                           {showPw ? <VisibilityOff/> : <Visibility/>}
                                                       </IconButton>
                                                   </InputAdornment>
                                               )
                                           }
                                       }}
                            />
                        </VStack>
                    </CardContent>
                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                    <CardActions sx={{justifyContent: 'flex-end', px: 2, py: 1.5}}>
                        <Button variant="contained" onClick={onLogin} disabled={loading} sx={{minWidth: 100, minHeight: 40}}>
                            {loading ? 'Logging in…' : 'Login'}
                        </Button>
                    </CardActions>
                </Card>

                {/* Footer version */}
                <HStack justifyContent="center" sx={{mt: 1}}>
                    <Typography variant="caption" sx={{
                        color: "text.secondary"
                    }}>v{pkg.version}</Typography>
                </HStack>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack(s => ({...s, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack(s => ({...s, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
