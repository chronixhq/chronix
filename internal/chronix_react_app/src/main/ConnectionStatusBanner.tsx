import {Alert, Snackbar, Stack, Typography, CircularProgress} from '@mui/material'
import WarningAmberRoundedIcon from '@mui/icons-material/WarningAmberRounded'
import {useSseContext} from '../data/SseContext'

export const ConnectionStatusBanner = () => {
  const {connectionState, retryCount, showBanner} = useSseContext()
  const open = showBanner || connectionState === 'disconnected'
  const isWarning = connectionState === 'disconnected'

  const message = isWarning
    ? 'Connection to server lost. Still attempting to reconnect…'
    : 'Reconnecting to server…'

  return (
    <Snackbar
      open={open}
      anchorOrigin={{vertical: 'top', horizontal: 'center'}}
      autoHideDuration={null}
      sx={{
        '& .MuiPaper-root': { minWidth: 360 }
      }}
    >
      <Alert
        severity={isWarning ? 'warning' : 'info'}
        icon={isWarning ? <WarningAmberRoundedIcon/> : <CircularProgress size={18} thickness={5} color="inherit"/>}
        sx={{ alignItems: 'center' }}
      >
        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <Typography variant="body2">{message}</Typography>
          {!isWarning && retryCount > 0 && (
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>(attempt {retryCount})</Typography>
          )}
        </Stack>
      </Alert>
    </Snackbar>
  );
}
