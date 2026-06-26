import {Chip} from '@mui/material';
import {Autorenew, CheckCircleOutlined, ErrorOutlined, PauseCircleOutlined} from '@mui/icons-material';
import {type JobSchedule} from '../modules/Jobs/types.ts';
import {formatScheduleSummary} from '../modules/Jobs/scheduleSummary.ts';
import {useJobActivitySse} from '../modules/Jobs/useJobActivitySse.ts';
// Generic date formatting helpers for consistent UI (honor user settings)
import {formatDateTimeTZ, formatDateTimeTZHM, formatDateTZ, getCurrentDisplayOptions} from './datetime';

// Shared UI: live job status chip, driven by JobActivity SSE
export function JobStatusChip({jobId, fallbackStatus}: { jobId: string | number; fallbackStatus?: string }) {
    const activity = useJobActivitySse(jobId);
    const status = activity?.isRunning ? 'running' : (activity?.lastStatus || fallbackStatus || 'idle');
    switch (status) {
        case 'success':
            return <Chip size="small" color="success" icon={<CheckCircleOutlined fontSize="small"/>} label="Success"/>;
        case 'failed':
        case 'error':
            return <Chip size="small" color="error" icon={<ErrorOutlined fontSize="small"/>} label={status === 'failed' ? 'Failed' : 'Error'}/>;
        case 'running':
            return <Chip size="small" color="primary" icon={<Autorenew fontSize="small"/>} label="Running"/>;
        case 'canceled':
            return <Chip size="small" variant="outlined" icon={<PauseCircleOutlined fontSize="small"/>} label="Canceled"/>;
        case 'idle':
        default:
            return <Chip size="small" variant="outlined" icon={<PauseCircleOutlined fontSize="small"/>} label="Idle"/>;
    }
}


// Accepts either the UI shape (with `schedule`) or raw DB row fields and returns a summary string.
export function scheduleSummaryForJob(job: any): string {
    if (!job) return '—';
    if (job.schedule) {
        return formatScheduleSummary(job.schedule as JobSchedule) || '—';
    }
    return '—';
}

export function formatDateTime(value: string | number | Date | undefined | null): string {
    if (!value) return '';
    const opts = getCurrentDisplayOptions();
    return formatDateTimeTZ(value, {timeZone: opts.timeZone, hour12: opts.hour12});
}

export function formatDateTimeHM(value: string | number | Date | undefined | null): string {
    if (!value) return '';
    const opts = getCurrentDisplayOptions();
    return formatDateTimeTZHM(value, {timeZone: opts.timeZone, hour12: opts.hour12});
}

export function formatDate(value: string | number | Date | undefined | null): string {
    if (!value) return '';
    const opts = getCurrentDisplayOptions();
    return formatDateTZ(value, {timeZone: opts.timeZone});
}

// Run status chip for list/table contexts
export function RunStatusChip({status}: { status: string }) {
    const color: any = (
        status === 'success' ? 'success' :
            status === 'failed' ? 'error' :
                status === 'canceled' ? 'default' :
                    status === 'running' ? 'info' :
                        'warning'
    );
    return <Chip size="small" label={status} color={color} variant={status === 'running' ? 'filled' : 'outlined'}/>;
}
