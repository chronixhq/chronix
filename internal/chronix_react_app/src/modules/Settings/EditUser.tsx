import {useEffect, useMemo, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControlLabel, IconButton, InputAdornment, MenuItem, Snackbar, Switch, TextField, Typography} from '@mui/material';
import {Visibility, VisibilityOff} from '@mui/icons-material';
import {useLocation, useNavigate} from 'react-router';
import {apiGet, apiPost, apiPut} from '@dsherwin/react-api-interface'
import {useAuthContext} from '../../data/useAuthContext'
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext'
import {HStack, VStack} from "@dsherwin/mui-kit";

function useQuery() {
  const {search} = useLocation();
  return useMemo(() => new URLSearchParams(search), [search]);
}


function supportedTimeZones(): string[] {
  const anyIntl: any = Intl as any;
  if (anyIntl && typeof anyIntl.supportedValuesOf === 'function') {
    try {
      const vals = anyIntl.supportedValuesOf('timeZone');
      if (Array.isArray(vals) && vals.length > 0) return vals as string[];
    } catch {/* ignore */}
  }
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
  const base = ['UTC', 'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles', 'Europe/London', 'Europe/Berlin', 'Asia/Tokyo', 'Asia/Singapore', 'Australia/Sydney'];
  const set = new Set([tz, ...base]);
  return Array.from(set);
}

export const EditUser = () => {
  const navigate = useNavigate();
  const query = useQuery();
  const { user, logout } = useAuthContext();
  const { checkLimit, reload: reloadFeatureAvailability } = useFeatureAvailability();
  const userLimit = checkLimit('users');

  const [loading, setLoading] = useState<boolean>(true);
  const [emailAvailable, setEmailAvailable] = useState<boolean | undefined>(undefined);
  const [loadError, setLoadError] = useState<string|null>(null);
  const [snack, setSnack] = useState<{open:boolean; message:string; severity:'success'|'error'|'info'}>({open:false,message:'',severity:'info'});

  const [id, setId] = useState<string>('');
  const [displayName, setDisplayName] = useState<string>('');
  const [email, setEmail] = useState<string>('');
  const [originalEmail, setOriginalEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  // Profile fields
  const [phone, setPhone] = useState<string>('');
  const [timeZone, setTimeZone] = useState<string>(Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC');
  const [timeFormat, setTimeFormat] = useState<'12h'|'24h'>('24h');

  // Change password dialog state
  const [pwDialogOpen, setPwDialogOpen] = useState<boolean>(false);
  const [showPw, setShowPw] = useState<boolean>(false);
  const [newPw, setNewPw] = useState<string>('');
  const [verifyPw, setVerifyPw] = useState<string>('');

  const [forceChange, setForceChange] = useState<boolean>(false);
  const [disabled, setDisabled] = useState<boolean>(false);
  const [isAdmin, setIsAdmin] = useState<boolean>(false);
  const [originalIsAdmin, setOriginalIsAdmin] = useState<boolean>(false);
  const [confirmRevokeOpen, setConfirmRevokeOpen] = useState<boolean>(false);

  useEffect(() => {
    const userId = query.get('id') || '';
    if (!userId && !userLimit.allowed && !loading) {
      navigate('/settings/users');
    }
  }, [userLimit.allowed, loading, navigate, query]);

  useEffect(()=>{
    const userId = query.get('id') || '';
    const load = async () => {
      setLoading(true); setLoadError(null);
      try {
        if (!userId) {
          // Create mode
          setId(''); setDisplayName(''); setEmail(''); setOriginalEmail(''); setPassword(''); setIsAdmin(false); setOriginalIsAdmin(false); setDisabled(false); setForceChange(false);
          setPhone('');
          setTimeZone(Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC');
          setTimeFormat('24h');
        } else {
          // No single-user endpoint; fetch list and find user
          const data = await apiGet('/users') as any[]
          const u = data.find(u => String(u.id) === userId)
          if (!u) throw new Error('User not found')
          if (u.suspended) {
            navigate('/settings/users')
            return
          }
          setId(String(u.id)); setDisplayName(u.name); setEmail(u.email); setOriginalEmail(u.email); setIsAdmin(!!u.admin); setOriginalIsAdmin(!!u.admin); setDisabled(!u.enabled); setForceChange(!!u.forcePasswordChange);
          setPhone(u.phone || '');
          setTimeZone(u.timeZone || Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC');
          setTimeFormat((u.timeFormat === '12h' || u.timeFormat === '24h') ? u.timeFormat : '24h');
        }
      } catch (e) {
        console.error(e);
        setLoadError('Failed to load user');
      } finally { setLoading(false); }
    };
    load();
  }, [query, navigate]);

  useEffect(() => {
    const e = email.trim();
    if (!e) { setEmailAvailable(undefined); return; }
    if (originalEmail && e.toLowerCase() === originalEmail.trim().toLowerCase()) { setEmailAvailable(true); return; }
    const t = window.setTimeout(async () => {
      try {
        const res = await apiGet(`/users/check-email?email=${encodeURIComponent(e)}${id ? `&excludeId=${encodeURIComponent(id)}` : ''}`) as { available: boolean }
        setEmailAvailable(res.available)
      } catch {
        setEmailAvailable(undefined)
      }
    }, 400)
    return () => { window.clearTimeout(t) }
  }, [email, id, originalEmail])

  const onSave = async () => {
    if (!displayName.trim() || !email.trim()) { setSnack({open:true, message:'Display name and Email are required.', severity:'error'}); return; }
    if (emailAvailable === false) { setSnack({open:true, message:'Email is already in use.', severity:'error'}); return; }
    const isSelf = user && id && String(user.id) === String(id);
    const isSelfRevoking = isSelf && originalIsAdmin && !isAdmin;
    if (isSelfRevoking && !confirmRevokeOpen) {
      setConfirmRevokeOpen(true);
      return;
    }
    try {
      const payload: any = { id: id? Number(id): undefined, name: displayName, email, admin: isAdmin, enabled: !disabled, forcePasswordChange: forceChange };
      if (!id) {
        payload.password = password || undefined;
      }
      // Include profile fields; if empty strings, backend preserves prior values due to pointer semantics
      payload.phone = phone || undefined;
      payload.timeZone = timeZone || undefined;
      payload.timeFormat = timeFormat || undefined;
      await apiPost('/user', payload);
      setSnack({open:true, message:`User ${id? 'updated':'created'}.`, severity:'success'});
      void reloadFeatureAvailability();
      if (isSelfRevoking) {
        try { await apiGet('/logout'); } catch (e) { console.error(e); }
        logout();
        return;
      }
      navigate('/settings/users');
    } catch (e: any) {
      console.error(e);
      // If server provided message (e.g., preventing self-revoke with no other admins), surface it
      setSnack({open:true, message: e?.message || 'Failed to save user.', severity:'error'});
    } finally {
      setConfirmRevokeOpen(false);
    }
  };

  const onChangePassword = async () => {
    if (!id) { setSnack({open:true, message:'Save the user first, then set a password.', severity:'error'}); return; }
    if (newPw.length < 8) { setSnack({open:true, message:'Password must be at least 8 characters.', severity:'error'}); return; }
    if (newPw !== verifyPw) { setSnack({open:true, message:'Passwords do not match.', severity:'error'}); return; }
    try {
      const res = await apiPut(`/settings/users/${encodeURIComponent(id)}/password`, { newPassword: newPw });
      if (!res.ok) throw new Error('Failed to change password');
      setPwDialogOpen(false);
      setNewPw(''); setVerifyPw('');
      setSnack({open:true, message:'Password updated.', severity:'success'});
    } catch (e:any) {
      console.error(e);
      setSnack({open:true, message: e?.message || 'Failed to change password.', severity:'error'});
    }
  }

  return (
    <Box sx={{px:{xs:1, md:2}, py:2, width:'100%'}}>
      <VStack spacing={2} sx={{maxWidth: 800, width:'100%', mx:'auto'}}>
        <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap:'wrap'}}>
          <Typography variant="h5">{id ? 'Edit User' : 'Create User'}</Typography>
          <HStack spacing={1} sx={{mt:{xs:1, sm:0}}}>
            <Button variant="outlined" onClick={()=>navigate('/settings/users')}>Cancel</Button>
            <Button variant="contained" onClick={onSave}>Save</Button>
          </HStack>
        </HStack>
        <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

        {loadError && (<Alert severity="error">{loadError}</Alert>)}

        <Card variant="outlined" sx={{borderRadius:3}}>
          <CardContent>
            {loading ? (
              <Typography variant="body2" sx={{
                color: "text.secondary"
              }}>Loading…</Typography>
            ) : (
              <VStack spacing={2}>
                <TextField label="Display name" value={displayName} onChange={(e)=>setDisplayName(e.target.value)} fullWidth required/>
                <TextField 
                  label="Email" 
                  value={email} 
                  onChange={(e)=>setEmail(e.target.value)} 
                  fullWidth 
                  required 
                  error={email.trim() !== '' && emailAvailable === false}
                  helperText={email.trim() !== '' && emailAvailable === false ? 'Email is already in use' : ''}
                />
                {!id && (
                  <TextField 
                    label="Password" 
                    type={showPw ? 'text' : 'password'} 
                    value={password} 
                    onChange={(e)=>setPassword(e.target.value)} 
                    fullWidth 
                    required 
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
                )}
                <TextField label="Phone" value={phone} onChange={(e)=>setPhone(e.target.value)} fullWidth/>
                <HStack spacing={2} sx={{flexWrap:'wrap'}}>
                  <TextField select label="Time zone" value={timeZone} onChange={(e)=>setTimeZone(e.target.value)} sx={{minWidth:260}}>
                    {supportedTimeZones().map(tz => (
                      <MenuItem key={tz} value={tz}>{tz}</MenuItem>
                    ))}
                  </TextField>
                  <TextField select label="Time format" value={timeFormat} onChange={(e)=>setTimeFormat(e.target.value as any)} sx={{minWidth:160}}>
                    <MenuItem value="12h">12-hour</MenuItem>
                    <MenuItem value="24h">24-hour</MenuItem>
                  </TextField>
                  <Button variant="outlined" onClick={()=>setPwDialogOpen(true)} disabled={!id} sx={{ml:'auto'}}>Change Password…</Button>
                </HStack>
                <HStack spacing={2} sx={{flexWrap:'wrap'}}>
                  <FormControlLabel control={<Switch checked={forceChange} onChange={(e)=>setForceChange(e.target.checked)} />} label="Force password change at next login"/>
                  <FormControlLabel control={<Switch checked={!disabled} onChange={(e)=>setDisabled(!e.target.checked)} />} label={'Enabled'}/>
                  <FormControlLabel control={<Switch checked={isAdmin} onChange={(e)=>setIsAdmin(e.target.checked)} />} label={'Admin'}/>
                </HStack>
              </VStack>
            )}
          </CardContent>
          <CardActions sx={{justifyContent:'flex-end'}}>
            <Button variant="contained" onClick={onSave}>Save</Button>
          </CardActions>
        </Card>

        <Dialog open={confirmRevokeOpen} onClose={()=>setConfirmRevokeOpen(false)}>
          <DialogTitle>Revoke your own admin access?</DialogTitle>
          <DialogContent>
            <Typography>
              You are about to revoke admin permissions from your own account. You will be logged out immediately after this change.
              If no other admin accounts exist, this action will be blocked by the server.
            </Typography>
          </DialogContent>
          <DialogActions sx={{justifyContent:'flex-end'}}>
            <Button variant="outlined" color="error" onClick={()=>setConfirmRevokeOpen(false)}>Cancel</Button>
            <Button color="error" variant="contained" onClick={onSave}>Confirm</Button>
          </DialogActions>
        </Dialog>

        <Dialog open={pwDialogOpen} onClose={()=>setPwDialogOpen(false)}>
          <DialogTitle>Change Password</DialogTitle>
          <DialogContent>
            <VStack spacing={2} sx={{mt:1, minWidth: 360}}>
              <TextField 
                label="New password" 
                type={showPw ? 'text' : 'password'} 
                value={newPw} 
                onChange={(e)=>setNewPw(e.target.value)} 
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
                label="Confirm new password" 
                type={showPw ? 'text' : 'password'} 
                value={verifyPw} 
                onChange={(e)=>setVerifyPw(e.target.value)} 
                fullWidth 
                error={!!verifyPw && newPw !== verifyPw} 
                helperText={!!verifyPw && newPw !== verifyPw ? 'Passwords do not match' : ''}
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
          </DialogContent>
          <DialogActions sx={{justifyContent:'flex-end'}}>
            <Button variant="outlined" onClick={()=>setPwDialogOpen(false)}>Cancel</Button>
            <Button variant="contained" onClick={onChangePassword} disabled={!id}>Save Password</Button>
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
