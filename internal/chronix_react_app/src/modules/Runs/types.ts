export type RunStatus = 'queued' | 'running' | 'success' | 'error' | 'warning' | 'canceled' | 'cancelled' | 'unknown'

export interface RunListItem {
    runId: string
    jobId: number
    jobName?: string
    status: RunStatus | string
    queuedAt: string
    startedAt?: string
    finishedAt?: string
    durationMs?: number
    message?: string
}

export interface RunRecord extends RunListItem {
    actionName?: string
    connectionName?: string
    rowsAffected?: number
    [key: string]: unknown
}

export interface RunStepDetail {
    id?: string
    stepOrder?: number
    stepName?: string
    name?: string
    status?: string
    startedAt?: string
    finishedAt?: string
    rowsCount?: number
    rowsAffected?: number
    exitCode?: number
    responseStatus?: number
    latencyMs?: number
    errorCode?: string
    errorMessage?: string
    expectOk?: boolean
    expectMessage?: string
    expectation?: Record<string, unknown>
    details?: Record<string, unknown>
    sqlText?: string
    commandText?: string
    scriptText?: string
    requestUrl?: string
    requestMethod?: string
    requestHeaders?: Record<string, unknown>
    requestBody?: string
    responseHeaders?: Record<string, unknown>
    responseBody?: string
    stdoutText?: string
    stderrText?: string
    stdoutTruncated?: boolean
    stderrTruncated?: boolean
    [key: string]: unknown
}

export interface RunProgressSnapshot {
    status?: string
    message?: string
    updatedAt?: string
    [key: string]: unknown
}

export interface RunDetailData {
    run: RunRecord | null
    steps: RunStepDetail[]
    snapshot: RunProgressSnapshot | null
}

export interface RunsListResponse {
    items: RunListItem[]
    total: number
}

export interface JobProgressPayload {
    run_id?: string | number
    type?: string
    step_index?: number
    step_name?: string
    message?: string
}

export interface JobFinishedPayload {
    run_id?: string | number
    status?: string
    message?: string
}

export interface RunProgressEvent {
    ts: string
    type: string
    runId: string
    jobId?: number
    stepIndex?: number
    stepName?: string
    status?: string
    message?: string
    fields?: Record<string, unknown>
}

export interface RunProgressData {
    snapshot: RunProgressSnapshot | null
    events: RunProgressEvent[]
}

export interface RunProgressMessage {
    ts: Date
    text: string
    type: string
    stepIndex?: number
    stepName?: string
    message?: string
}
