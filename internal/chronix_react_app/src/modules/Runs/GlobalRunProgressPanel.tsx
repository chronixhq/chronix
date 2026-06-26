import {Box, Stack} from '@mui/material'
import {useRunNow} from '../../data/RunNowContext'
import {LiveRunNowProgressPanel} from './LiveRunNowProgressPanel'

// Renders floating, global run progress panels for all active Run Now executions.
export const GlobalRunProgressPanel = () => {
    const {activeRuns, dismiss, getRunTitle} = useRunNow()
    if (!activeRuns.length) return null
    return (
        <Box sx={{
            position: 'fixed',
            right: 16,
            bottom: 16,
            zIndex: (theme) => theme.zIndex.snackbar + 1,
            width: {xs: 'calc(100% - 32px)', sm: 420, md: 480},
            maxWidth: 'calc(100% - 32px)'
        }}>
            <Stack spacing={1}>
                {activeRuns.map(rid => (
                    <LiveRunNowProgressPanel key={rid} runId={rid} title={getRunTitle(rid) || 'Live run progress'} onClose={() => dismiss(rid)} />
                ))}
            </Stack>
        </Box>
    )
}
