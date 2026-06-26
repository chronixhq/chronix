import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, FormHelperText, Snackbar, TextField, Typography} from '@mui/material';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {SectionHelp} from '../../main/SectionHelp';
import {fetchServerUrlSettings, saveServerUrlSettings} from './api.ts';

export const ServerUrlPage = () => {
  const [serverUrl, setServerUrl] = useState<string>('');
  const [snack, setSnack] = useState<{open:boolean; message:string; severity:'success'|'error'|'info'}>({open:false, message:'', severity:'info'});

  useEffect(() => {
    (async () => {
      try {
        const data = await fetchServerUrlSettings();
        setServerUrl(data.serverUrl);
      } catch (e) {
        console.error(e);
        setSnack({open:true, message:'Failed to load server URL.', severity:'error'});
      }
    })();
  }, []);

  const onSave = async () => {
    try {
      await saveServerUrlSettings({serverUrl});
      setSnack({open:true, message:'Server URL saved.', severity:'success'});
    } catch (e) {
      console.error(e);
      setSnack({open:true, message:'Failed to save server URL.', severity:'error'});
    }
  };

  return (
    <Box sx={{px:{xs:1, md:2}, py:2, width:'100%'}}>
      <VStack spacing={2} sx={{maxWidth: 800, width:'100%', mx:'auto'}}>
        <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap:'wrap'}}>
          <Box sx={{display: 'flex', alignItems: 'center'}}>
            <Typography variant="h5">Server URL</Typography>
            <SectionHelp />
          </Box>
          <Button variant="contained" onClick={onSave}>Save</Button>
        </HStack>
        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

        <Card variant="outlined" sx={{borderRadius:3}}>
          <CardContent>
            <VStack spacing={2}>
              <TextField label="Server URL" value={serverUrl} onChange={(e)=>setServerUrl(e.target.value)} fullWidth placeholder="https://chronix.example.com"/>
              <FormHelperText>Used in certificates and notification links.</FormHelperText>
              <Alert severity="info">Set the fully-qualified domain name with protocol (http/https).</Alert>
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
