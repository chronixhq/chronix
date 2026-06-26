import React from 'react';
import {Alert, Snackbar} from '@mui/material';
import {useSseContext} from '../data/SseContext';
import {useRunNow} from '../data/RunNowContext';
import type {JobFinishedPayload} from '../modules/Runs/types.ts';

// Displays a global snackbar whenever any run finishes, on browsers that did NOT initiate the run
export const GlobalRunSnack: React.FC = () => {
    const {addSSEListener} = useSseContext();
    const {isLocalRun} = useRunNow();
    const [open, setOpen] = React.useState(false);
    const [message, setMessage] = React.useState('');
    const [severity, setSeverity] = React.useState<'success' | 'error'>('success');

    // Keep a short-lived memory of recently notified runIds to avoid duplicates
    const recentRef = React.useRef<Map<string, number>>(new Map()); // runId -> ts

    React.useEffect(() => {
        const cleanupOld = () => {
            const now = Date.now();
            for (const [k, ts] of recentRef.current.entries()) {
                if (now - ts > 2 * 60 * 1000) { // 2 minutes window
                    recentRef.current.delete(k);
                }
            }
        };

        const onFinished = (payload: JobFinishedPayload) => {
            try {
                const runId = payload.run_id != null ? String(payload.run_id) : '';
                if (!runId) return;
                // If this run was initiated in this browser, suppress the snack (a dialog will show)
                if (isLocalRun(runId)) return;
                cleanupOld();
                if (recentRef.current.has(runId)) return; // dedupe
                recentRef.current.set(runId, Date.now());
                const status = String(payload.status || '').toLowerCase();
                const ok = status === 'success';
                const msg = payload.message ? String(payload.message) : (ok ? 'Run completed successfully' : 'Run finished with errors');
                setSeverity(ok ? 'success' : 'error');
                setMessage(`Run ${runId} ${ok ? 'completed' : 'failed'}${msg ? ` — ${msg}` : ''}`);
                setOpen(true);
            } catch {
                // ignore
            }
        };

        const unsub = addSSEListener<JobFinishedPayload>('job_finished', onFinished);
        return () => {
            unsub?.();
        };
    }, [addSSEListener, isLocalRun]);

    return (
        <Snackbar open={open} autoHideDuration={6000} onClose={() => setOpen(false)} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
            <Alert onClose={() => setOpen(false)} severity={severity} variant="filled" sx={{width: '100%'}}>
                {message}
            </Alert>
        </Snackbar>
    );
};
