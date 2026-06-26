import {useEffect, useMemo, useState} from 'react'
import {useSseContext} from '../../data/SseContext'
import {apiPost} from '@dsherwin/react-api-interface'
import {Alert, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Stack, Typography} from '@mui/material'
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext'

interface AgentRegEvent {
  requestId: string
  uuid: string
  name: string
  ip?: string
  metadata?: Record<string, any>
  expiresAt?: string
}

type State = { open: boolean, data?: AgentRegEvent, countdownSec: number }

export const GlobalAgentRegistrationDialog = () => {
  const {addSSEListener} = useSseContext()
  const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability()
  const agentLimit = checkLimit('agents')
  const [state, setState] = useState<State>({open: false, data: undefined, countdownSec: 300})

  useEffect(() => {
    const unsub = addSSEListener<AgentRegEvent>('agent_registration', (data) => {
      setState({open: true, data, countdownSec: 300})
    })
    const unsubApproved = addSSEListener<any>('agent_registration_approved', (d) => {
      setState((s) => {
        if (s.data && d && d.requestId === s.data.requestId) {
          return {...s, open: false}
        }
        return s
      })
    })
    const unsubDenied = addSSEListener<any>('agent_registration_denied', (d) => {
      setState((s) => {
        if (s.data && d && d.requestId === s.data.requestId) {
          return {...s, open: false}
        }
        return s
      })
    })
    return () => { unsub(); unsubApproved(); unsubDenied() }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!state.open) return
    const t = window.setInterval(() => {
      setState((s) => ({...s, countdownSec: Math.max(0, s.countdownSec - 1)}))
    }, 1000)
    return () => window.clearInterval(t)
  }, [state.open])

  // Auto-close and alert when timed out
  useEffect(() => {
    if (!state.open) return
    if (state.countdownSec <= 0) {
      setState((s) => ({...s, open: false}))
      // Simple alert for MVP; can be replaced with app snackbar/toast later
      window.alert('Agent registration request timed out. Please try again from the agent.')
    }
  }, [state.open, state.countdownSec])

  const onApprove = async () => {
    if (!state.data) return
    await apiPost(`/agents/requests/${encodeURIComponent(state.data.requestId)}/approve`, {})
    void reloadFeatureAvailability()
    setState((s) => ({...s, open: false}))
  }
  const onDeny = async () => {
    if (!state.data) return
    await apiPost(`/agents/requests/${encodeURIComponent(state.data.requestId)}/deny`, {})
    setState((s) => ({...s, open: false}))
  }

  const details = useMemo(() => {
    const d: Array<[string, string]> = []
    if (state.data?.uuid) d.push(['UUID', state.data.uuid])
    if (state.data?.ip) d.push(['IP', state.data.ip])
    if (state.data?.metadata) {
      const m = state.data.metadata
      const add = (k: string, label?: string) => { const v = (m as any)[k]; if (typeof v === 'string') d.push([label || k, v]) }
      add('os', 'OS');
      add('os_type', 'OS Type');
      add('os_version', 'OS Version');
      add('arch', 'Arch');
      add('user', 'Running As');
    }
    return d
  }, [state.data])

  // Prevent closing by ESC or backdrop click
  const handleDialogClose = (_e: unknown, reason?: string) => {
    if (reason === 'backdropClick' || reason === 'escapeKeyDown') return
    setState((s) => ({...s, open: false}))
  }

  return (
    <Dialog open={state.open} onClose={handleDialogClose} maxWidth="sm" fullWidth>
      <DialogTitle>Agent '{state.data?.name}' is requesting to register</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>
          Agent '{state.data?.name}' (UUID {state.data?.uuid}) from {state.data?.ip ?? 'unknown IP'} is requesting to register with Chronix. Approve?
        </DialogContentText>

        {!agentLimit.allowed && (
            <Alert severity="warning" sx={{mb: 2}}>
                {agentLimit.message}
            </Alert>
        )}

        <Stack spacing={1} sx={{mt: 2}}>
          {details.map(([k,v]) => (<Typography key={k} variant="body2"><strong>{k}:</strong> {v}</Typography>))}
          <Typography variant="caption" sx={{
            color: "text.secondary"
          }}>Expires in {state.countdownSec}s</Typography>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onDeny} color="inherit">Deny</Button>
        <Button onClick={onApprove} variant="contained" disabled={!agentLimit.allowed}>Approve</Button>
      </DialogActions>
    </Dialog>
  );
}
