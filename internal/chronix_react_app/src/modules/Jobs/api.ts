import {apiGet} from '@dsherwin/react-api-interface'
import type {Job} from './types'

export function normalizeJob(raw: any): Job {
    return {
        id: String(raw?.id ?? ''),
        name: raw?.name ?? '',
        description: raw?.description ?? undefined,
        notes: raw?.notes ?? undefined,
        targetKind: raw?.targetKind ?? raw?.target_kind ?? undefined,
        schedule: raw?.schedule,
        connectionId: String(raw?.connectionId ?? raw?.connection_id ?? ''),
        shellConnectionId: String(raw?.shellConnectionId ?? raw?.shell_connection_id ?? ''),
        webtaskConnectionId: String(raw?.webtaskConnectionId ?? raw?.webtask_connection_id ?? ''),
        actionId: String(raw?.actionId ?? raw?.action_id ?? ''),
        enabled: raw?.enabled !== false,
        suspended: raw?.suspended === true,
        variables: Array.isArray(raw?.variables) ? raw.variables.map((variable: any) => ({name: variable.name, value: variable.value})) : undefined,
        alertEmails: raw?.alertEmails ?? raw?.alert_emails ?? undefined,
        alertPhones: raw?.alertPhones ?? raw?.alert_phones ?? undefined,
        lastRunStatus: raw?.lastRunStatus ?? raw?.last_run_status ?? undefined,
        lastRunAt: raw?.lastRunAt ?? raw?.last_run_at ?? undefined,
        nextRunAt: raw?.nextRunAt ?? raw?.next_run_at ?? undefined,
        createdAt: raw?.createdAt ?? raw?.created_at ?? undefined,
        updatedAt: raw?.updatedAt ?? raw?.updated_at ?? undefined,
        notifyOnSuccess: raw?.notifyOnSuccess ?? raw?.notify_on_success ?? undefined,
        notifyOnError: raw?.notifyOnError ?? raw?.notify_on_error ?? undefined,
        notifyIncludeOutput: raw?.notifyIncludeOutput ?? raw?.notify_include_output ?? undefined,
        connection: raw?.connection,
        action: raw?.action,
    } as Job
}

export async function fetchJobs(): Promise<Job[]> {
    const data = await apiGet('/jobs')
    return (Array.isArray(data) ? data : []).map(normalizeJob)
}

export async function fetchJobById(id: string | number): Promise<Job> {
    const data = await apiGet(`/jobs/${encodeURIComponent(String(id))}`)
    return normalizeJob(data)
}
