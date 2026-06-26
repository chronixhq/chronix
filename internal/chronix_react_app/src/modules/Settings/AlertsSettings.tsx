import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, FormControlLabel, Snackbar, Switch, TextField, Typography} from '@mui/material';
import {HStack, VStack} from "@dsherwin/mui-kit";
import type {GlobalAlertSettings} from './types';
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';
import {fetchAlertSettings, saveAlertSettings} from './api.ts';

export const AlertsSettingsPage = () => {
  const [alerts, setAlerts] = useState<GlobalAlertSettings>({ systemAlertEmails: '', systemAlertPhones: '', alertOnAgentLost: true });
  const [snack, setSnack] = useState<{open:boolean; message:string; severity:'success'|'error'|'info'}>({open:false, message:'', severity:'info'});

  useEffect(() => {
    (async () => {
      try {
        setAlerts(await fetchAlertSettings());
      } catch (e) {
        console.error(e);
        setSnack({open:true, message:'Failed to load alert settings.', severity:'error'});
      }
    })();
  }, []);

  const onSave = async () => {
    try {
      await saveAlertSettings(alerts);
      setSnack({open:true, message:'Alert settings saved.', severity:'success'});
    } catch (e) {
      console.error(e);
      setSnack({open:true, message:'Failed to save alert settings.', severity:'error'});
    }
  };

  return (
    <Box sx={{px:{xs:1, md:2}, py:2, width:'100%'}}>
      <VStack spacing={2} sx={{maxWidth: 1000, width:'100%', mx:'auto'}}>
        <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap:'wrap'}}>
          <Box sx={{display: 'flex', alignItems: 'center'}}>
            <Typography variant="h5">Alerts & Notifications</Typography>
            <SectionHelp section={HELP_SECTIONS.notifications} />
          </Box>
          <Button variant="contained" onClick={onSave}>Save</Button>
        </HStack>
        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

        <Card variant="outlined" sx={{borderRadius:3}}>
          <CardContent>
            <VStack spacing={3}>
              <Typography variant="subtitle1" sx={{
                fontWeight: "bold"
              }}>System Level Alerts</Typography>
              <Typography variant="body2" sx={{
                color: "text.secondary"
              }}>
                Configure destinations for system-wide alerts, such as agent disconnection or critical server errors.
              </Typography>
              
              <TextField 
                label="System Alert Emails" 
                placeholder="email1@example.com, email2@example.com"
                value={alerts.systemAlertEmails} 
                onChange={(e)=>setAlerts(s=>({...s, systemAlertEmails: e.target.value}))} 
                fullWidth
                helperText="Comma-separated list of email addresses."
              />

              <TextField 
                label="System Alert Phones (SMS)" 
                placeholder="+15550001111, +15550002222"
                value={alerts.systemAlertPhones} 
                onChange={(e)=>setAlerts(s=>({...s, systemAlertPhones: e.target.value}))} 
                fullWidth
                helperText="Comma-separated list of E.164 phone numbers."
              />

              <Divider />

              <FormControlLabel
                control={
                  <Switch 
                    checked={alerts.alertOnAgentLost} 
                    onChange={(e)=>setAlerts(s=>({...s, alertOnAgentLost: e.target.checked}))} 
                  />
                }
                label="Alert when an agent is lost (disconnected)"
              />
              
              <Alert severity="info">
                Ensure Email and/or SMS provider settings are configured correctly in their respective tabs.
              </Alert>
            </VStack>
          </CardContent>
          <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
          <CardActions sx={{justifyContent:'flex-end'}}>
            <Button variant="contained" onClick={onSave}>Save</Button>
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
