import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, FormControl, IconButton, InputAdornment, InputLabel, MenuItem, Select, Snackbar, TextField, Typography} from '@mui/material';
import {Visibility, VisibilityOff} from '@mui/icons-material';
import type {GlobalSmsSettings} from './types';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';
import {fetchSmsSettings, saveSmsSettings, testSmsSettings} from './api.ts';

export const SmsNotifierPage = () => {
  const [sms, setSms] = useState<GlobalSmsSettings>({ provider: 'none' });
  const [snack, setSnack] = useState<{open:boolean; message:string; severity:'success'|'error'|'info'}>({open:false, message:'', severity:'info'});
  const [showPw, setShowPw] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testDialogOpen, setTestDialogOpen] = useState(false);
  const [testNumber, setTestNumber] = useState('');

  useEffect(() => {
    (async () => {
      try {
        setSms(await fetchSmsSettings());
      } catch (e) {
        console.error(e);
        setSnack({open:true, message:'Failed to load SMS settings.', severity:'error'});
      }
    })();
  }, []);

  const onSave = async () => {
    try {
      await saveSmsSettings(sms);
      setSnack({open:true, message:'SMS settings saved.', severity:'success'});
    } catch (e) {
      console.error(e);
      setSnack({open:true, message:'Failed to save SMS settings.', severity:'error'});
    }
  };

  const onTest = () => {
    setTestDialogOpen(true);
  };

  const runTest = async () => {
    setTestDialogOpen(false);
    setTesting(true);
    try {
      await testSmsSettings({
        ...sms,
        testNumber: testNumber,
      });
      setSnack({open:true, message:'SMS test successful!', severity:'success'});
    } catch (e) {
      console.error(e);
      const message = e instanceof Error && e.message ? e.message : 'SMS test failed.'
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
            <Typography variant="h5">SMS Notifier</Typography>
            <SectionHelp section={HELP_SECTIONS.notifications} />
          </Box>
          <HStack spacing={1}>
            <Button variant="outlined" color="secondary" onClick={onTest} disabled={testing || sms.provider === 'none'}>
              {testing ? 'Testing...' : 'Test SMS'}
            </Button>
            <Button variant="contained" onClick={onSave} disabled={testing}>Save</Button>
          </HStack>
        </HStack>
        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

        <Card variant="outlined" sx={{borderRadius:3}}>
          <CardContent>
            <VStack spacing={2}>
              <FormControl sx={{minWidth:{xs:'100%', md:240}}}>
                <InputLabel id="sms-provider-label">Provider</InputLabel>
                <Select
                  labelId="sms-provider-label"
                  label="Provider"
                  value={sms.provider}
                  onChange={(e)=>setSms(s=>({...s, provider: e.target.value === 'twilio' ? 'twilio' : 'none'}))}
                >
                  <MenuItem value="none">None</MenuItem>
                  <MenuItem value="twilio">Twilio</MenuItem>
                </Select>
              </FormControl>
              {sms.provider === 'twilio' && (
                <HStack spacing={2} sx={{flexWrap:'wrap'}}>
                  <TextField label="From number" value={sms.fromNumber||''} onChange={(e)=>setSms(s=>({...s, fromNumber: e.target.value}))} sx={{minWidth:{xs:'100%', md:220}}}/>
                  <TextField label="Account SID" value={sms.accountSid||''} onChange={(e)=>setSms(s=>({...s, accountSid: e.target.value}))} sx={{minWidth:{xs:'100%', md:220}}}/>
                  <TextField 
                    type={showPw ? 'text' : 'password'} 
                    label="Auth Token or API Secret" 
                    value={sms.authToken||''} 
                    onChange={(e)=>setSms(s=>({...s, authToken: e.target.value}))} 
                    sx={{minWidth:{xs:'100%', md:300}}}
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
              )}
            </VStack>
          </CardContent>
          <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
          <CardActions sx={{justifyContent:'flex-end', gap: 1}}>
            <Button variant="outlined" color="secondary" onClick={onTest} disabled={testing || sms.provider === 'none'}>
              {testing ? 'Testing...' : 'Test SMS'}
            </Button>
            <Button variant="contained" onClick={onSave} disabled={testing}>Save</Button>
          </CardActions>
        </Card>

        <Dialog open={testDialogOpen} onClose={() => setTestDialogOpen(false)}>
          <DialogTitle>Test SMS Notifier</DialogTitle>
          <DialogContent>
            <DialogContentText sx={{mb: 2}}>
              Enter a phone number to send a test SMS message to.
            </DialogContentText>
            <TextField
              autoFocus
              label="Phone Number"
              fullWidth
              variant="outlined"
              value={testNumber}
              onChange={(e) => setTestNumber(e.target.value)}
              placeholder="+15550001212"
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setTestDialogOpen(false)}>Cancel</Button>
            <Button onClick={runTest} variant="contained" color="primary" disabled={!testNumber.trim()}>Send Test SMS</Button>
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
