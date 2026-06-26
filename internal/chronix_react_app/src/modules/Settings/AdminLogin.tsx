import {useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, Snackbar, TextField, Typography} from '@mui/material';
import {alpha, useTheme} from '@mui/material/styles';
import ThemeToggle from '../../site/themes/ThemeToggle.tsx';
import ChronixLettersSvg from '../../assets/Chronix-letters.svg?react';
import ChronixGearMark from '../../assets/Chronix-gears.png';
import pkg from '../../../package.json';
import {useAuthContext} from "../../data/useAuthContext.ts";
import {ServerStatus} from "../../data/types.ts";
import InitialSetup from "./InitialSetup.tsx";
import {apiGet} from "@dsherwin/react-api-interface";
import {HStack, VStack} from "@dsherwin/mui-kit";

export const AdminLogin = () => {
    const theme = useTheme();
    const {serverStatus} = useAuthContext();
    const [code, setCode] = useState('');
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    const [loading, setLoading] = useState(false);
    const [initialSetup, setInitialSetup] = useState(false);
    const {setLoggedIn} = useAuthContext();

    const onLogin = async () => {
        if (!code.trim()) {
            setSnack({open: true, message: 'Enter the admin code.', severity: 'error'});
            return;
        }
        setLoading(true);
        try {
            const res:{status: string} = await apiGet("/auth/settings/"+code);
            if(res.status !== "logged in") {
                throw new Error('Login failed');
            }
            if (serverStatus !== ServerStatus.ACTIVE) {
                setInitialSetup(true);
                return;
            }
            setLoggedIn(true);
        } catch (e) {
            console.error(e);
            setSnack({open: true, message: 'Invalid or expired code. Codes are valid for 10 minutes.', severity: 'error'});
        } finally {
            setLoading(false);
        }
    };

    if(initialSetup) return (<InitialSetup/>)

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
                    <Typography variant="h5">Settings Sign In</Typography>
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>Use your temporary admin code (valid for 10m)</Typography>
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>To obtain admin code, look in the server log after start up, or use the command "chronix adminCode" on the cli</Typography>
                </VStack>

                <Card variant="outlined" sx={{borderRadius: 2.5}}>
                    <CardContent>
                        <VStack spacing={2}>
                            <TextField label="Admin code" value={code} onChange={(e) => setCode(e.target.value)} fullWidth autoFocus onKeyDown={(e)=>{ if(e.key==='Enter'){ e.preventDefault(); onLogin(); } }} />
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
