import {Button, Dialog, DialogActions, DialogContent, DialogTitle, Stack, Typography} from '@mui/material'
import {useRunNow} from '../data/RunNowContext'
import {useNavigate} from 'react-router'

// Shows a confirmation dialog when a Run Now finishes, including the result status and message.
export const GlobalRunFinishedDialog = () => {
    const {nextCompleted, popCompleted} = useRunNow()
    const navigate = useNavigate()

    const open = Boolean(nextCompleted)
    const status = nextCompleted?.status?.toLowerCase() || 'finished'
    const isSuccess = status === 'success'

    const handleClose = () => popCompleted()

    return (
        <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
            <DialogTitle>Run Now {isSuccess ? 'Completed' : 'Finished'}</DialogTitle>
            <DialogContent>
                <Stack spacing={1}>
                    <Typography variant="body1">
                        Run <Typography component="span" variant="body1" sx={{fontFamily: 'monospace'}}>{nextCompleted?.runId}</Typography> {isSuccess ? 'completed successfully.' : 'finished.'}
                    </Typography>
                    <Typography variant="body2" sx={{
                        color: "text.secondary"
                    }}>
                        Status: {nextCompleted?.status}
                    </Typography>
                    {nextCompleted?.message && (
                        <Typography variant="body2">{nextCompleted.message}</Typography>
                    )}
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose}>Dismiss</Button>
                <Button onClick={() => { if (nextCompleted?.runId) navigate(`/runs/${encodeURIComponent(nextCompleted.runId)}`); handleClose() }} variant="contained">
                    View details
                </Button>
            </DialogActions>
        </Dialog>
    );
}
