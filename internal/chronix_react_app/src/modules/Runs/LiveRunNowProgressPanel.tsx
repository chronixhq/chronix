import {Button, Card, CardContent, Chip, IconButton, LinearProgress, List, ListItem, ListItemText, Stack, Typography} from '@mui/material'
import {Close} from '@mui/icons-material'
import {useNavigate} from 'react-router'
import {useRunProgressSse} from './useRunProgressSse'
import {useEffect, useState} from 'react'
import {cancelRun} from './api.ts'
import type {ChipProps} from '@mui/material'

/**
 * LiveRunNowProgressPanel
 * - Shows a compact, live-updating progress panel for a given runId.
 * - If `external` is provided, the panel renders that state instead of subscribing itself.
 */
export const LiveRunNowProgressPanel = ({
                                            runId,
                                            title = 'Live run progress',
                                            onClose,
                                        }: {
    runId: string | undefined
    title?: string
    closable?: boolean
    onClose?: () => void | undefined | null
}) => {
    const state = useRunProgressSse(runId)
    const navigate = useNavigate()

    const [canceling, setCanceling] = useState(false)
    const [cancelRequested, setCancelRequested] = useState(false)

    // Auto-close when finished after 15 seconds
    useEffect(() => {
        if (state?.finished && onClose) {
            const timer = setTimeout(() => {
                onClose()
            }, 15000)
            return () => clearTimeout(timer)
        }
    }, [state?.finished, onClose])

    if (!runId) return null

    const effectiveStatus = cancelRequested ? 'cancellationRequested' : (state?.status || 'running')
    const color: ChipProps['color'] = (
        effectiveStatus === 'success' ? 'success' :
            effectiveStatus === 'failed' ? 'error' :
                effectiveStatus === 'canceled' ? 'default' :
                    effectiveStatus === 'cancellationRequested' ? 'warning' :
                        'info'
    )

    const onCancel = async () => {
        if (canceling || state?.finished) return
        try {
            setCanceling(true)
            await cancelRun(runId)
            setCancelRequested(true)
        } catch (e) {
            // Optional: surface error somehow; for now, revert canceling flag
            console.error(e)
        } finally {
            setCanceling(false)
        }
    }

    return (
        <Card variant="outlined" sx={{borderRadius: 3}}>
            <CardContent>
                <Stack
                    direction={{xs: 'column', md: 'row'}}
                    spacing={1}
                    sx={{
                        alignItems: {xs: 'stretch', md: 'center'},
                        justifyContent: "space-between",
                        mb: 1
                    }}>
                    <Typography variant="h6">{title}</Typography>
                    <Stack direction="row" spacing={1} sx={{
                        alignItems: "center"
                    }}>
                        <Chip size="small" label={effectiveStatus} color={color}/>
                        {!state?.finished && (
                            <Button size="small" color="warning" disabled={canceling || cancelRequested} onClick={onCancel}>
                                {cancelRequested ? 'Cancel requested' : (canceling ? 'Canceling…' : 'Cancel run')}
                            </Button>
                        )}
                        <Button size="small" onClick={() => navigate(`/runs/${encodeURIComponent(runId)}`)}>Open details</Button>
                        {onClose && (
                            <IconButton size="small" aria-label="close" onClick={onClose}>
                                <Close fontSize="small"/>
                            </IconButton>
                        )}
                    </Stack>
                </Stack>
                {!state?.finished && <LinearProgress sx={{mb: 1}}/>}
                <List dense sx={{maxHeight: 240, overflowY: 'auto'}}>
                    {state?.messages?.length ? state.messages.map((m, i) => (
                        <ListItem key={i} sx={{py: 0}}>
                            <ListItemText
                                primary={<Typography variant="body2"><Typography
                                    component="span"
                                    variant="body2"
                                    sx={{
                                        color: "text.secondary",
                                        mr: 1
                                    }}>{m.ts.toLocaleTimeString()}</Typography>{m.text}</Typography>}
                            />
                        </ListItem>
                    )) : (
                        <ListItem><ListItemText primary={<Typography variant="body2" sx={{
                            color: "text.secondary"
                        }}>Waiting for updates…</Typography>}/></ListItem>
                    )}
                </List>
            </CardContent>
        </Card>
    );
}
