import {useCallback, useEffect, useMemo, useState} from 'react'
import {Alert, Box, Button, Checkbox, Chip, CircularProgress, Divider, FormControl, FormControlLabel, IconButton, InputLabel, MenuItem, Pagination, Select, Stack, Tooltip, Typography} from '@mui/material'
import {useMuiPrompts} from '@dsherwin/mui-kit'
import {apiGet, apiPost} from '@dsherwin/react-api-interface'
import DeleteIcon from '@mui/icons-material/Delete'
import RefreshIcon from '@mui/icons-material/Refresh'
import type {NotificationItem, NotificationSeverity} from '../../data/types'
import {useNotifications} from '../../data/NotificationsContext'
import {useSseContext} from '../../data/SseContext'
import {formatDateTime} from '../../lib/utilities'
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';

interface ListResponse { items: NotificationItem[]; total: number; page: number; pageSize: number }

export const NotificationsPage = () => {
  const {confirmPrompt} = useMuiPrompts()
  const { markSeen, refresh: refreshRecent } = useNotifications()
  const { addSSEListener } = useSseContext()
  const [items, setItems] = useState<NotificationItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [severity, setSeverity] = useState<'' | NotificationSeverity>('')
  const [category, setCategory] = useState('')
  const [selected, setSelected] = useState<number[]>([])
  const [refreshing, setRefreshing] = useState(false)
  const pageCount = useMemo(() => Math.max(1, Math.ceil(total / pageSize)), [total, pageSize])

  const load = useCallback(async () => {
    const params = new URLSearchParams()
    params.set('page', String(page))
    params.set('pageSize', String(pageSize))
    if (severity) params.set('severity', severity)
    if (category) params.set('category', category)
    const res = await apiGet(`/notifications?${params.toString()}`) as ListResponse
    setItems(res.items)
    setTotal(res.total)
    // mark visible items as seen per spec
    const unseenIds = res.items.filter(i => !i.seen).map(i => i.id)
    if (unseenIds.length) { void markSeen(unseenIds) }
  }, [page, pageSize, severity, category, markSeen])

  useEffect(() => { void load() }, [load])

  // keep selection only for visible items
  useEffect(() => {
    setSelected(prev => prev.filter(id => items.some(i => i.id === id)))
  }, [items])

  // Refresh the page list when a notification SSE event arrives
  useEffect(() => {
    const unsubscribe = addSSEListener('notification', () => { void load() })
    return () => { unsubscribe?.() }
  }, [addSSEListener, load])

  const openConfirm = useCallback(async (ids: number[]) => {
    if (!ids.length) return
    const ok = await confirmPrompt({
        title: ids.length === 1 ? 'Remove Notification' : 'Remove Notifications',
        message: ids.length === 1 ? 'Are you sure you want to remove this notification?' : `Are you sure you want to remove ${ids.length} notifications?`,
        buttonText: 'Remove',
        buttonColor: 'error'
    })
    if (!ok) return
    try {
      await apiPost('/notifications/mark-removed', { ids })
      // reload page list after removal
      await load()
      // refresh recent list and unseen badge as well
      await refreshRecent()
      // clear selection of removed ids
      setSelected(prev => prev.filter(id => !ids.includes(id)))
    } catch {/* ignore; polling/UI reload covers */}
  }, [confirmPrompt, load, refreshRecent])

  const toggleOne = useCallback((id: number) => {
    setSelected(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }, [])

  const allOnPageIds = useMemo(() => items.map(i => i.id), [items])
  const allSelectedOnPage = useMemo(() => allOnPageIds.length > 0 && allOnPageIds.every(id => selected.includes(id)), [allOnPageIds, selected])
  const toggleAllOnPage = useCallback((checked: boolean) => {
    setSelected(prev => {
      const set = new Set(prev)
      if (checked) {
        allOnPageIds.forEach(id => set.add(id))
      } else {
        allOnPageIds.forEach(id => set.delete(id))
      }
      return Array.from(set)
    })
  }, [allOnPageIds])

  const refreshAll = useCallback(async () => {
    setRefreshing(true)
    try {
      await load()
      await refreshRecent()
    } finally {
      setRefreshing(false)
    }
  }, [load, refreshRecent])

  return (
    <Box sx={{p: 2}}>
      <Box sx={{display: 'flex', alignItems: 'center', mb: 2}}>
        <Typography variant="h5">Notifications</Typography>
        <SectionHelp section={HELP_SECTIONS.notifications} />
      </Box>
      <Stack
        direction="row"
        spacing={2}
        sx={{
          alignItems: "center",
          justifyContent: "space-between",
          mb: 2,
          flexWrap: 'wrap'
        }}>
        <Stack direction="row" spacing={2} sx={{
          alignItems: "center"
        }}>
          <FormControl size="small" sx={{minWidth: 160}}>
            <InputLabel id="severity-label">Severity</InputLabel>
            <Select labelId="severity-label" label="Severity" value={severity} onChange={(e) => { setPage(1); setSeverity(e.target.value as any) }}>
              <MenuItem value=""><em>All</em></MenuItem>
              <MenuItem value="info">Info</MenuItem>
              <MenuItem value="success">Success</MenuItem>
              <MenuItem value="warning">Warning</MenuItem>
              <MenuItem value="error">Error</MenuItem>
            </Select>
          </FormControl>
          <FormControl size="small" sx={{minWidth: 160}}>
            <InputLabel id="category-label">Category</InputLabel>
            <Select labelId="category-label" label="Category" value={category} onChange={(e) => { setPage(1); setCategory(e.target.value as any) }}>
              <MenuItem value=""><em>All</em></MenuItem>
              <MenuItem value="job">Job</MenuItem>
              <MenuItem value="system">System</MenuItem>
              <MenuItem value="connection">Connection</MenuItem>
            </Select>
          </FormControl>
        </Stack>
        <Stack direction="row" spacing={2} sx={{
          alignItems: "center"
        }}>
          <FormControlLabel control={<Checkbox checked={allSelectedOnPage} indeterminate={!allSelectedOnPage && selected.some(id => allOnPageIds.includes(id))} onChange={(e) => toggleAllOnPage(e.target.checked)} />} label="Select page" />
          <Tooltip title="Refresh">
            <span>
              <IconButton aria-label="refresh" onClick={() => void refreshAll()} disabled={refreshing} size="small">
                {refreshing ? <CircularProgress size={18} /> : <RefreshIcon />}
              </IconButton>
            </span>
          </Tooltip>
          <Button variant="outlined" color="error" disabled={selected.length === 0} onClick={() => openConfirm(selected)} startIcon={<DeleteIcon/>}>
            Remove Selected ({selected.length})
          </Button>
        </Stack>
      </Stack>
      <Stack spacing={1}>
        {items.length === 0 && (
          <Alert severity="info">No notifications found.</Alert>
        )}
        {items.map(n => (
          <Box key={n.id} sx={{p: 1.5, borderRadius: 2, border: theme => `1px solid ${theme.palette.divider}`}}>
            <Stack
              direction="row"
              spacing={1}
              sx={{
                justifyContent: "space-between",
                alignItems: "center"
              }}>
              <Stack
                direction="row"
                spacing={1}
                sx={{
                  alignItems: "flex-start",
                  flex: 1,
                  minWidth: 0
                }}>
                <Checkbox size="small" checked={selected.includes(n.id)} onChange={() => toggleOne(n.id)} />
                <Stack sx={{minWidth: 0}}>
                  <Typography variant="subtitle2" noWrap>{n.subject}</Typography>
                  {!!(n.data && (n.data as any).message) && (
                    <Typography variant="body2" sx={{ mt: 0.5, whiteSpace: 'pre-wrap' }}>
                      {(n.data as any).message}
                    </Typography>
                  )}
                  <Typography variant="caption" sx={{
                    color: "text.secondary"
                  }}>{formatDateTime(n.createdAt)}</Typography>
                </Stack>
              </Stack>
              <Stack direction="row" spacing={1} sx={{
                alignItems: "center"
              }}>
                {!!n.origin && <Chip size="small" label={n.origin} />}
                <Chip size="small" label={n.category} />
                <Chip size="small" color={n.severity === 'error' ? 'error' : n.severity === 'warning' ? 'warning' : n.severity === 'success' ? 'success' : 'default'} label={n.severity} />
                <Tooltip title="Remove">
                  <IconButton size="small" onClick={() => openConfirm([n.id])}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Stack>
            </Stack>
          </Box>
        ))}
      </Stack>
      <Divider sx={{my: 2, borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}} />
      <Stack
        direction="row"
        sx={{
          justifyContent: "space-between",
          alignItems: "center"
        }}>
        <Typography variant="body2" sx={{
          color: "text.secondary"
        }}>{total} total</Typography>
        <Pagination page={page} count={pageCount} onChange={(_, p) => setPage(p)} />
      </Stack>
    </Box>
  );
}
