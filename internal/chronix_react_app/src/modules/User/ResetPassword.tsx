import {useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, IconButton, InputAdornment, Snackbar, TextField, Typography} from '@mui/material';
import {Visibility, VisibilityOff} from '@mui/icons-material';
import {useNavigate} from 'react-router';
import {apiPost} from "@dsherwin/react-api-interface";
import {VStack} from "@dsherwin/mui-kit";

export const ResetPassword = () => {
  const navigate = useNavigate();
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [showPw, setShowPw] = useState(false);
  const [snack, setSnack] = useState<{open:boolean; message:string; severity:'success'|'error'|'info'}>({open:false,message:'',severity:'info'});
  const [loading, setLoading] = useState(false);

  const onReset = async () => {
    if (!code.trim() || !password.trim() || !confirm.trim()) { setSnack({open:true, message:'Complete all fields.', severity:'error'}); return; }
    if (password !== confirm) { setSnack({open:true, message:'Passwords do not match.', severity:'error'}); return; }
    setLoading(true);
    try {
      const res = await apiPost('/reset', {code, password});
      if ((res as any)?.ok === false) throw new Error('Reset failed');
      setSnack({open:true, message:'Password has been reset.', severity:'success'});
    } catch (e) {
      console.error(e);
      setSnack({open:true, message:'Invalid or expired code.', severity:'error'});
    } finally { setLoading(false); }
  };

  return (
    <Box sx={{px:{xs:1, md:2}, py:2, width:'100%'}}>
      <VStack spacing={2} sx={{maxWidth: 480, width:'100%', mx:'auto'}}>
        <Typography variant="h5">Reset password</Typography>
        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
        <Card variant="outlined" sx={{borderRadius:3}}>
          <CardContent>
            <VStack spacing={2}>
              <TextField label="Confirmation code" value={code} onChange={(e)=>setCode(e.target.value)} fullWidth/>
              <TextField 
                label="New password" 
                type={showPw ? 'text' : 'password'} 
                value={password} 
                onChange={(e)=>setPassword(e.target.value)} 
                fullWidth
                slotProps={{
                  input: {
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton onClick={() => setShowPw(!showPw)} edge="end" aria-label="toggle password visibility">
                          {showPw ? <VisibilityOff/> : <Visibility/>}
                        </IconButton>
                      </InputAdornment>
                    )
                  }
                }}
              />
              <TextField 
                label="Confirm password" 
                type={showPw ? 'text' : 'password'} 
                value={confirm} 
                onChange={(e)=>setConfirm(e.target.value)} 
                fullWidth
                slotProps={{
                  input: {
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton onClick={() => setShowPw(!showPw)} edge="end" aria-label="toggle password visibility">
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
          <CardActions sx={{justifyContent:'flex-end'}}>
            <Button variant="outlined" color="error" onClick={() => navigate(-1)}>Cancel</Button>
            <Button variant="contained" onClick={onReset} disabled={loading}>{loading? 'Resetting…' : 'Reset password'}</Button>
          </CardActions>
        </Card>

        <Snackbar open={snack.open} autoHideDuration={3000} onClose={()=>setSnack(s=>({...s, open:false}))} anchorOrigin={{vertical:'top', horizontal:'center'}}>
          <Alert onClose={()=>setSnack(s=>({...s, open:false}))} severity={snack.severity} variant="filled" sx={{width:'100%'}}>
            {snack.message}
          </Alert>
        </Snackbar>
      </VStack>
    </Box>
  );
};
