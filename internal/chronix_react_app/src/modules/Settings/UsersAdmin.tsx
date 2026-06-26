import {useEffect, useMemo, useState} from 'react';
import {formatDateTime} from '../../lib/utilities';
import {Alert, Box, Button, Card, CardActions, CardContent, Chip, Collapse, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControlLabel, IconButton, Menu, MenuItem, Snackbar, Switch, TextField, Tooltip, Typography} from '@mui/material';
import {Add, AdminPanelSettings, Delete, Edit, ExpandMore, History, LockReset, MoreVert, Refresh, Warning} from '@mui/icons-material';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {type SettingsUser} from './types';
import {useNavigate} from 'react-router';
import {apiDelete, apiGet, apiPost} from '@dsherwin/react-api-interface'
import {formatAPIError} from '../../lib/errors'
import {useAuthContext} from '../../data/useAuthContext'
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext'

export const UsersAdmin = () => {
    const navigate = useNavigate();
    const {user, logout} = useAuthContext();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const userLimit = checkLimit('users');
    const [items, setItems] = useState<SettingsUser[]>([]);
    const [loading, setLoading] = useState<boolean>(false);
    const [search, setSearch] = useState<string>('');
    const [loadError, setLoadError] = useState<string | null>(null);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    const [deleteDlg, setDeleteDlg] = useState<{ open: boolean; id?: string; name?: string }>({open: false});
    const [revokeDlg, setRevokeDlg] = useState<{ open: boolean; id?: string; name?: string }>({open: false});
    const [pwDlg, setPwDlg] = useState<{ open: boolean; id?: string; name?: string }>({open: false});
    const [newPw, setNewPw] = useState<string>('');
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const [menuUser, setMenuUser] = useState<SettingsUser | null>(null);

    const handleMenuOpen = (event: React.MouseEvent<HTMLButtonElement>, u: SettingsUser) => {
        setAnchorEl(event.currentTarget);
        setMenuUser(u);
    };
    const handleMenuClose = () => {
        setAnchorEl(null);
        setMenuUser(null);
    };

    const [expanded, setExpanded] = useState<Record<string, boolean>>({});
    const toggleExpand = (id: string) => setExpanded(prev => ({...prev, [id]: !prev[id]}));
    const isExpanded = (id: string) => !!expanded[id];

    const load = async () => {
        setLoading(true);
        setLoadError(null);
        try {
            const data = await apiGet('/users') as { id: number; name: string; email: string; phone?: string; enabled: boolean; admin: boolean; forcePasswordChange: boolean; suspended: boolean }[]
            const mapped: SettingsUser[] = Array.isArray(data)
                ? data.map(u => ({id: String(u.id), displayName: u.name, email: u.email, disabled: !u.enabled, isAdmin: u.admin, forcePasswordChange: u.forcePasswordChange, suspended: u.suspended}))
                : []
            setItems(mapped)
        } catch (e) {
            console.error(e);
            setLoadError('Failed to load users');
            setItems([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);

    const filtered = useMemo(() => {
        const q = search.trim().toLowerCase();
        if (!q) return items;
        return items.filter(u => u.displayName.toLowerCase().includes(q) || u.email.toLowerCase().includes(q));
    }, [items, search]);

    const openDelete = (id: string, name: string) => setDeleteDlg({open: true, id, name});
    const closeDelete = () => setDeleteDlg({open: false});
    const confirmDelete = async () => {
        if (!deleteDlg.id) return;
        try {
            await apiDelete(`/settings/users/${encodeURIComponent(deleteDlg.id)}`);
            setSnack({open: true, message: 'User deleted', severity: 'success'});
            setDeleteDlg({open: false});
            void reloadFeatureAvailability();
            await load();
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: e?.message || 'Failed to delete user', severity: 'error'});
        }
    };

    const toggleDisabled = async (u: SettingsUser) => {
        if (u.disabled && !userLimit.allowed) {
            setSnack({open: true, message: userLimit.message || 'User limit reached', severity: 'info'});
            return;
        }
        try {
            await apiPost('/user', {id: Number(u.id), name: u.displayName, email: u.email, enabled: u.disabled, admin: !!u.isAdmin, forcePasswordChange: !!u.forcePasswordChange});
            void reloadFeatureAvailability();
            await load();
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: formatAPIError(e, 'Failed to update user'), severity: 'error'});
        }
    };

    const toggleAdmin = async (u: SettingsUser) => {
        // If the current user is revoking their own admin, confirm first
        const isSelf = user && String(user.id) === String(u.id);
        const targetIsAdmin = !u.isAdmin;
        if (isSelf && u.isAdmin && !targetIsAdmin) {
            setRevokeDlg({open: true, id: u.id, name: u.displayName});
            return;
        }
        try {
            await apiPost('/user', {id: Number(u.id), name: u.displayName, email: u.email, enabled: !u.disabled, admin: targetIsAdmin, forcePasswordChange: !!u.forcePasswordChange});
            await load();
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: formatAPIError(e, 'Failed to update user'), severity: 'error'});
        }
    };

    const toggleForcePasswordChange = async (u: SettingsUser) => {
        try {
            await apiPost('/user', {id: Number(u.id), name: u.displayName, email: u.email, enabled: !u.disabled, admin: !!u.isAdmin, forcePasswordChange: !u.forcePasswordChange});
            setSnack({open: true, message: `Force password change ${!u.forcePasswordChange ? 'enabled' : 'disabled'}`, severity: 'success'});
            await load();
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: e?.message || 'Failed to update user', severity: 'error'});
        }
    };

    const onResetPassword = async () => {
        if (!pwDlg.id || !newPw) return;
        if (newPw.length < 8) {
            setSnack({open: true, message: 'Password must be at least 8 characters', severity: 'error'});
            return;
        }
        try {
            await apiPost(`/settings/users/${encodeURIComponent(pwDlg.id)}/password`, {newPassword: newPw}, {method: 'PUT'});
            setSnack({open: true, message: 'Password updated', severity: 'success'});
            setPwDlg({open: false});
            setNewPw('');
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: e?.message || 'Failed to update password', severity: 'error'});
        }
    };

    const confirmSelfRevoke = async () => {
        if (!revokeDlg.id) {
            setRevokeDlg({open: false});
            return;
        }
        try {
            // Proceed to revoke admin on self
            const u = items.find(x => x.id === revokeDlg.id);
            if (!u) throw new Error('User not found');
            await apiPost('/user', {id: Number(u.id), name: u.displayName, email: u.email, enabled: !u.disabled, admin: false, forcePasswordChange: !!u.forcePasswordChange});
            
            // If successful, log the user out per requirement
            try {
                await apiGet('/logout');
            } catch (e) {
                console.error(e);
            }
            logout();
        } catch (e: any) {
            console.error(e);
            setSnack({open: true, message: formatAPIError(e, 'Failed to update user'), severity: 'error'});
        } finally {
            setRevokeDlg({open: false});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Typography variant="h5">Users</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Tooltip title="Refresh"><IconButton onClick={load}><Refresh/></IconButton></Tooltip>
                        <Button
                            startIcon={<Add/>}
                            onClick={() => navigate('/settings/users/edit')}
                            disabled={!userLimit.allowed}
                        >
                            New User
                        </Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {!userLimit.allowed && (
                    <Alert severity="warning">
                        {userLimit.message}
                    </Alert>
                )}

                <HStack spacing={1} sx={{flexWrap: 'wrap'}}>
                    <TextField size="small" placeholder="Search by name or email" value={search} onChange={(e) => setSearch(e.target.value)} sx={{minWidth: {xs: '100%', sm: 320}}}/>
                </HStack>

                {loadError && (
                    <Alert severity="error" action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>{loadError}</Alert>
                )}

                {loading ? (
                    <VStack spacing={2}>
                        {[...Array(3)].map((_, i) => (
                            <Card key={i} variant="outlined" sx={{borderRadius: 3, p: 2}}>
                                <Typography variant="body2" sx={{
                                    color: "text.secondary"
                                }}>Loading…</Typography>
                            </Card>
                        ))}
                    </VStack>
                ) : filtered.length === 0 ? (
                    <Card variant="outlined" sx={{borderRadius: 3, p: 3, textAlign: 'center'}}>
                        <Typography variant="h6">No users found</Typography>
                        <Typography
                            sx={{
                                color: "text.secondary",
                                mt: 1
                            }}>Try adjusting your search or create a new user.</Typography>
                        <Button sx={{mt: 2}} startIcon={<Add/>} onClick={() => navigate('/settings/users/edit')} disabled={!userLimit.allowed}>New User</Button>
                    </Card>
                ) : (
                    <VStack spacing={2}>
                        {filtered.map(u => (
                            <Card key={u.id} variant="outlined" sx={{borderRadius: 3}}>
                                <CardContent sx={{pb: 1}}>
                                    <HStack alignItems="center" justifyContent="space-between" sx={{gap: 2, flexWrap: 'wrap'}}>
                                        <HStack alignItems="center" sx={{gap: 1.5, minWidth: 240, flex: 1}}>
                                            <Typography variant="subtitle1" sx={{
                                                fontWeight: 600
                                            }}>{u.displayName}</Typography>
                                            {u.isAdmin && <Chip size="small" color="primary" icon={<AdminPanelSettings fontSize="small"/>} label="Admin"/>}
                                            {u.disabled && <Chip size="small" color="warning" label="Disabled"/>}
                                            {u.suspended && (
                                                <Tooltip title="Suspended. This user is temporarily inactive.">
                                                    <Chip size="small" color="warning" icon={<Warning fontSize="small"/>} label="Suspended"/>
                                                </Tooltip>
                                            )}
                                            {u.forcePasswordChange && <Chip size="small" variant="outlined" label="Force PW change"/>}
                                            {u.last_login_at && <Typography variant="caption" sx={{
                                                color: "text.secondary"
                                            }}>Last login: {formatDateTime(u.last_login_at)}</Typography>}
                                        </HStack>
                                        <HStack alignItems="center" sx={{gap: 0.5}}>
                                            <Tooltip title="More"><IconButton size="small" onClick={(e) => handleMenuOpen(e, u)}><MoreVert/></IconButton></Tooltip>
                                            <Tooltip title={u.suspended ? "Cannot edit suspended user" : "Edit"}>
                        <span>
                          <IconButton size="small" onClick={() => navigate(`/settings/users/edit?id=${encodeURIComponent(u.id)}`)} disabled={u.suspended}><Edit/></IconButton>
                        </span>
                                            </Tooltip>
                                            <Tooltip title="Delete"><IconButton size="small" onClick={() => openDelete(u.id, u.displayName)}><Delete/></IconButton></Tooltip>
                                        </HStack>
                                    </HStack>
                                    <Typography variant="body2" sx={{
                                        color: "text.secondary"
                                    }}>{u.email}</Typography>
                                </CardContent>
                                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                <CardActions sx={{justifyContent: 'space-between'}}>
                                    <HStack spacing={2} sx={{ml: 1}}>
                                        <FormControlLabel control={<Switch checked={!u.disabled} onChange={() => toggleDisabled(u)} disabled={u.suspended}/>} label={"Enabled"}/>
                                        <FormControlLabel control={<Switch checked={!!u.isAdmin} onChange={() => toggleAdmin(u)} disabled={u.suspended}/>} label={"Admin"}/>
                                    </HStack>
                                    <IconButton size="small" onClick={() => toggleExpand(u.id)}>
                                        <ExpandMore sx={{transform: isExpanded(u.id) ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s'}}/>
                                    </IconButton>
                                </CardActions>
                                <Collapse in={isExpanded(u.id)} timeout="auto" unmountOnExit>
                                    <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                                    <CardContent>
                                        <Typography variant="subtitle2" gutterBottom>Details</Typography>
                                        <VStack spacing={0.5}>
                                            <Typography variant="body2">ID: {u.id}</Typography>
                                            {u.created_at && <Typography variant="body2">Created: {formatDateTime(u.created_at)}</Typography>}
                                        </VStack>
                                    </CardContent>
                                </Collapse>
                            </Card>
                        ))}
                    </VStack>
                )}

                <Dialog open={deleteDlg.open} onClose={closeDelete}>
                    <DialogTitle>Delete user</DialogTitle>
                    <DialogContent>
                        <Typography>Are you sure you want to permanently delete “{deleteDlg.name}”? This action cannot be undone.</Typography>
                    </DialogContent>
                    <DialogActions>
                        <Button variant="outlined" color="primary" onClick={closeDelete}>Cancel</Button>
                        <Button color="error" variant="contained" onClick={confirmDelete}>Delete</Button>
                    </DialogActions>
                </Dialog>

                <Dialog open={revokeDlg.open} onClose={() => setRevokeDlg({open: false})}>
                    <DialogTitle>Revoke your own admin access?</DialogTitle>
                    <DialogContent>
                        <Typography>
                            You are about to revoke admin permissions from your own account. You will be logged out immediately after this change.
                            If no other admin accounts exist, this action will be blocked by the server.
                        </Typography>
                    </DialogContent>
                    <DialogActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="outlined" color="error" onClick={() => setRevokeDlg({open: false})}>Cancel</Button>
                        <Button color="error" variant="contained" onClick={confirmSelfRevoke}>Revoke Admin</Button>
                    </DialogActions>
                </Dialog>

                <Dialog open={pwDlg.open} onClose={() => setPwDlg({open: false})}>
                    <DialogTitle>Change Password for {pwDlg.name}</DialogTitle>
                    <DialogContent>
                        <VStack spacing={2} sx={{mt: 1}}>
                            <TextField
                                label="New Password"
                                type="password"
                                fullWidth
                                value={newPw}
                                onChange={(e) => setNewPw(e.target.value)}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter') void onResetPassword();
                                }}
                            />
                        </VStack>
                    </DialogContent>
                    <DialogActions>
                        <Button variant="outlined" color="primary" onClick={() => setPwDlg({open: false})}>Cancel</Button>
                        <Button variant="contained" color="primary" onClick={onResetPassword}>Set Password</Button>
                    </DialogActions>
                </Dialog>

                <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
                    <MenuItem onClick={() => {
                        handleMenuClose();
                        if (menuUser) navigate(`/activity?user=${encodeURIComponent(menuUser.displayName)}`);
                    }}>
                        <HStack spacing={1}><History fontSize="small"/><Typography>View Activity</Typography></HStack>
                    </MenuItem>
                    <MenuItem onClick={() => {
                        handleMenuClose();
                        if (menuUser) setPwDlg({open: true, id: menuUser.id, name: menuUser.displayName});
                    }}>
                        <HStack spacing={1}><LockReset fontSize="small"/><Typography>Change Password</Typography></HStack>
                    </MenuItem>
                    <MenuItem onClick={() => {
                        handleMenuClose();
                        if (menuUser) void toggleForcePasswordChange(menuUser);
                    }}>
                        <HStack spacing={1}>
                            <FormControlLabel
                                control={<Switch size="small" checked={!!menuUser?.forcePasswordChange}/>}
                                label="Force PW Change"
                                sx={{pointerEvents: 'none'}}
                            />
                        </HStack>
                    </MenuItem>
                </Menu>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack(s => ({...s, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack(s => ({...s, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
