import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, FormControl, IconButton, InputAdornment, InputLabel, MenuItem, Select, Snackbar, TextField, Typography} from '@mui/material';
import {Visibility, VisibilityOff} from '@mui/icons-material';
import type {GlobalEmailSettings} from './types';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';
import {fetchEmailSettings, saveEmailSettings, testEmailSettings} from './api.ts';

export const EmailNotifierPage = () => {
  const [email, setEmail] = useState<GlobalEmailSettings>({ smtpHost: '', smtpPort: '587', secure: 'starttls' });
  const [snack, setSnack] = useState<{open:boolean; message:string; severity:'success'|'error'|'info'}>({open:false, message:'', severity:'info'});
  const [showPw, setShowPw] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testDialogOpen, setTestDialogOpen] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        setEmail(await fetchEmailSettings());
      } catch (e) {
        console.error(e);
        setSnack({open:true, message:'Failed to load email settings.', severity:'error'});
      }
    })();
  }, []);

  const onSave = async () => {
    try {
      await saveEmailSettings(email);
      setSnack({open:true, message:'Email settings saved.', severity:'success'});
    } catch (e) {
      console.error(e);
      setSnack({open:true, message:'Failed to save email settings.', severity:'error'});
    }
  };

  const onTest = () => {
    setTestDialogOpen(true);
  };

  const runTest = async () => {
    setTestDialogOpen(false);
    setTesting(true);
    try {
      await testEmailSettings(email);
      setSnack({open:true, message:'SMTP test successful! Check your inbox.', severity:'success'});
    } catch (e) {
      console.error(e);
      const message = e instanceof Error && e.message ? e.message : 'SMTP test failed.'
      setSnack({open:true, message, severity:'error'});
    } finally {
      setTesting(false);
    }
  };

  return (
    <Box sx={{px:{xs:1, md:2}, py:2, width:'100%'}}>
      <VStack spacing={2} sx={{maxWidth: 1000, width:'100%', mx:'auto'}}>
        <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap:'wrap'}}>
          <Box sx={{display: 'flex', alignItems: 'center'}}>
            <Typography variant="h5">Email Notifier</Typography>
            <SectionHelp section={HELP_SECTIONS.notifications} />
          </Box>
          <HStack spacing={1}>
            <Button variant="outlined" color="secondary" onClick={onTest} disabled={testing}>
              {testing ? 'Testing...' : 'Test SMTP'}
            </Button>
            <Button variant="contained" onClick={onSave} disabled={testing}>Save</Button>
          </HStack>
        </HStack>
        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

        <Card variant="outlined" sx={{borderRadius:3}}>
          <CardContent>
            <VStack spacing={2}>
              <HStack spacing={2} sx={{flexWrap:'wrap'}}>
                <TextField label="SMTP host" value={email.smtpHost||''} onChange={(e)=>setEmail(s=>({...s, smtpHost: e.target.value}))} sx={{minWidth:{xs:'100%', md:320}}}/>
                <TextField label="SMTP port" value={email.smtpPort||''} onChange={(e)=>setEmail(s=>({...s, smtpPort: e.target.value.replace(/[^0-9]/g,'')}))} sx={{minWidth:{xs:'100%', md:160}}} slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}/>
                <FormControl sx={{minWidth:{xs:'100%', md:200}}}>
                  <InputLabel id="secure-label">Secure</InputLabel>
                  <Select
                    labelId="secure-label"
                    label="Secure"
                    value={email.secure||'none'}
                    onChange={(e)=>setEmail(s=>({...s, secure: e.target.value === 'ssl' || e.target.value === 'starttls' ? e.target.value : 'none'}))}
                  >
                    <MenuItem value="none">None</MenuItem>
                    <MenuItem value="ssl">SSL</MenuItem>
                    <MenuItem value="starttls">StartTLS</MenuItem>
                  </Select>
                </FormControl>
              </HStack>
              <HStack spacing={2} sx={{flexWrap:'wrap'}}>
                <TextField label="From name" value={email.fromName||''} onChange={(e)=>setEmail(s=>({...s, fromName: e.target.value}))} sx={{minWidth:{xs:'100%', md:260}}}/>
                <TextField label="From email" value={email.fromEmail||''} onChange={(e)=>setEmail(s=>({...s, fromEmail: e.target.value}))} sx={{minWidth:{xs:'100%', md:260}}}/>
              </HStack>
              <HStack spacing={2} sx={{flexWrap:'wrap'}}>
                <TextField label="SMTP login" value={email.smtpLogin||''} onChange={(e)=>setEmail(s=>({...s, smtpLogin: e.target.value}))} sx={{minWidth:{xs:'100%', md:260}}}/>
                <TextField 
                  type={showPw ? 'text' : 'password'} 
                  label="SMTP password" 
                  value={email.smtpPassword||''} 
                  onChange={(e)=>setEmail(s=>({...s, smtpPassword: e.target.value}))} 
                  sx={{minWidth:{xs:'100%', md:260}}}
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
              </HStack>
              <Alert severity="info">Use StartTLS on port 587 for most providers.</Alert>
            </VStack>
          </CardContent>
          <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
          <CardActions sx={{justifyContent:'flex-end', gap: 1}}>
            <Button variant="outlined" color="secondary" onClick={onTest} disabled={testing}>
              {testing ? 'Testing...' : 'Test SMTP'}
            </Button>
            <Button variant="contained" onClick={onSave} disabled={testing}>Save</Button>
          </CardActions>
        </Card>

        <Dialog open={testDialogOpen} onClose={() => setTestDialogOpen(false)}>
          <DialogTitle>Confirm SMTP Test</DialogTitle>
          <DialogContent>
            <DialogContentText>
              An email will be sent to verify your SMTP settings.
            </DialogContentText>
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" sx={{ fontWeight: 'bold' }}>To:</Typography>
              <Typography variant="body2" sx={{ mb: 1 }}>{email.fromEmail}</Typography>
              
              <Typography variant="body2" sx={{ fontWeight: 'bold' }}>From:</Typography>
              <Typography variant="body2" sx={{ mb: 1 }}>{email.fromName || 'Chronix'} &lt;{email.fromEmail}&gt;</Typography>
              
              <Typography variant="body2" sx={{ fontWeight: 'bold' }}>Via:</Typography>
              <Typography variant="body2">{email.smtpHost}:{email.smtpPort} ({email.secure})</Typography>
            </Box>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setTestDialogOpen(false)}>Cancel</Button>
            <Button onClick={runTest} variant="contained" color="primary">Send Test Email</Button>
          </DialogActions>
        </Dialog>

        <Snackbar open={snack.open} autoHideDuration={3000} onClose={()=>setSnack(s=>({...s, open:false}))} anchorOrigin={{vertical:'top', horizontal:'center'}}>
          <Alert onClose={()=>setSnack(s=>({...s, open:false}))} severity={snack.severity} variant="filled" sx={{width:'100%'}}>
            {snack.message}
          </Alert>
        </Snackbar>
      </VStack>
    </Box>
  );
};
