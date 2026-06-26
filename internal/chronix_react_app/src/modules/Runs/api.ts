import {apiGet, apiPost} from '@dsherwin/react-api-interface'
import type {RunDetailData, RunListItem, RunProgressData, RunProgressEvent, RunProgressSnapshot, RunRecord, RunsListResponse, RunStepDetail} from './types.ts'

type ApiRecord = Record<string, unknown>

function toRecord(value: unknown): ApiRecord {
    return value && typeof value === 'object' ? (value as ApiRecord) : {}
}

function toStringValue(value: unknown): string | undefined {
    if (typeof value === 'string') return value
    if (typeof value === 'number') return String(value)
    return undefined
}

function toNumberValue(value: unknown): number | undefined {
    if (typeof value === 'number' && Number.isFinite(value)) return value
    if (typeof value === 'string' && value.trim()) {
        const parsed = Number(value)
        if (Number.isFinite(parsed)) return parsed
    }
    return undefined
}

function toBooleanValue(value: unknown): boolean | undefined {
    if (typeof value === 'boolean') return value
    if (typeof value === 'number') return value !== 0
    if (typeof value === 'string') return value.toLowerCase() === 'true'
    return undefined
}

function toRunStepDetail(raw: unknown): RunStepDetail {
    const record = toRecord(raw)
    return {
        ...record,
        id: toStringValue(record.id),
        stepOrder: toNumberValue(record.stepOrder) ?? toNumberValue(record.step_order),
        stepName: toStringValue(record.stepName) ?? toStringValue(record.step_name) ?? toStringValue(record.name),
        name: toStringValue(record.name),
        status: toStringValue(record.status),
        startedAt: toStringValue(record.startedAt) ?? toStringValue(record.started_at),
        finishedAt: toStringValue(record.finishedAt) ?? toStringValue(record.finished_at),
        rowsCount: toNumberValue(record.rowsCount) ?? toNumberValue(record.rows_count),
        rowsAffected: toNumberValue(record.rowsAffected) ?? toNumberValue(record.rows_affected),
        exitCode: toNumberValue(record.exitCode) ?? toNumberValue(record.exit_code),
        responseStatus: toNumberValue(record.responseStatus) ?? toNumberValue(record.response_status),
        latencyMs: toNumberValue(record.latencyMs) ?? toNumberValue(record.latency_ms),
        errorCode: toStringValue(record.errorCode) ?? toStringValue(record.error_code),
        errorMessage: toStringValue(record.errorMessage) ?? toStringValue(record.error_message),
        expectOk: toBooleanValue(record.expectOk) ?? toBooleanValue(record.expect_ok),
        expectMessage: toStringValue(record.expectMessage) ?? toStringValue(record.expect_message),
        expectation: toRecord(record.expectation),
        details: toRecord(record.details),
        sqlText: toStringValue(record.sqlText) ?? toStringValue(record.sql_text),
        commandText: toStringValue(record.commandText) ?? toStringValue(record.command_text),
        scriptText: toStringValue(record.scriptText) ?? toStringValue(record.script_text),
        requestUrl: toStringValue(record.requestUrl) ?? toStringValue(record.request_url),
        requestMethod: toStringValue(record.requestMethod) ?? toStringValue(record.request_method),
        requestHeaders: toRecord(record.requestHeaders ?? record.request_headers),
        requestBody: toStringValue(record.requestBody) ?? toStringValue(record.request_body),
        responseHeaders: toRecord(record.responseHeaders ?? record.response_headers),
        responseBody: toStringValue(record.responseBody) ?? toStringValue(record.response_body),
        stdoutText: toStringValue(record.stdoutText) ?? toStringValue(record.stdout_text),
        stderrText: toStringValue(record.stderrText) ?? toStringValue(record.stderr_text),
        stdoutTruncated: toBooleanValue(record.stdoutTruncated) ?? toBooleanValue(record.stdout_truncated),
        stderrTruncated: toBooleanValue(record.stderrTruncated) ?? toBooleanValue(record.stderr_truncated),
    }
}

export function normalizeRunProgressEvent(raw: unknown): RunProgressEvent | null {
    const record = toRecord(raw)
    const runId = toStringValue(record.runId) ?? toStringValue(record.run_id)
    if (!runId) return null

    const fields = toRecord(record.fields)
    return {
        ts: toStringValue(record.ts) ?? new Date().toISOString(),
        type: toStringValue(record.type) ?? 'progress',
        runId,
        jobId: toNumberValue(record.jobId) ?? toNumberValue(record.job_id),
        stepIndex: toNumberValue(record.stepIndex) ?? toNumberValue(record.step_index),
        stepName: toStringValue(record.stepName) ?? toStringValue(record.step_name),
        status: toStringValue(record.status),
        message: toStringValue(record.message),
        fields: Object.keys(fields).length ? fields : undefined,
    }
}

export function normalizeRunListItem(raw: unknown): RunListItem {
    const record = toRecord(raw)
    return {
        runId: toStringValue(record.runId) ?? toStringValue(record.run_id) ?? toStringValue(record.id) ?? '',
        jobId: toNumberValue(record.jobId) ?? toNumberValue(record.job_id) ?? toNumberValue(record.job) ?? 0,
        jobName: toStringValue(record.jobName) ?? toStringValue(record.job_name),
        status: (toStringValue(record.status) ?? toStringValue(record.state) ?? 'unknown').toLowerCase(),
        queuedAt: toStringValue(record.queuedAt) ?? toStringValue(record.queued_at) ?? toStringValue(record.createdAt) ?? toStringValue(record.created_at) ?? new Date().toISOString(),
        startedAt: toStringValue(record.startedAt) ?? toStringValue(record.started_at),
        finishedAt: toStringValue(record.finishedAt) ?? toStringValue(record.finished_at),
        durationMs: toNumberValue(record.durationMs) ?? toNumberValue(record.duration_ms),
        message: toStringValue(record.message),
    }
}

export function normalizeRunRecord(raw: unknown): RunRecord {
    const record = toRecord(raw)
    const base = normalizeRunListItem(record)
    return {
        ...record,
        ...base,
        actionName: toStringValue(record.actionName) ?? toStringValue(record.action_name),
        connectionName: toStringValue(record.connectionName) ?? toStringValue(record.connection_name),
        rowsAffected: toNumberValue(record.rowsAffected) ?? toNumberValue(record.rows_affected),
    }
}

export function normalizeRunProgressSnapshot(raw: unknown): RunProgressSnapshot | null {
    const record = toRecord(raw)
    if (!Object.keys(record).length) return null
    return {
        ...record,
        status: toStringValue(record.status),
        message: toStringValue(record.message),
        updatedAt: toStringValue(record.updatedAt) ?? toStringValue(record.updated_at),
    }
}

export async function fetchRunsListPage(params: {
    limit: number
    offset: number
    q?: string
    status?: string
    jobId?: string | number
    startedFrom?: string
    startedTo?: string
}): Promise<RunsListResponse> {
    const qs = new URLSearchParams()
    qs.set('limit', String(params.limit))
    qs.set('offset', String(params.offset))
    if (params.q) qs.set('q', params.q)
    if (params.status) qs.set('status', params.status)
    if (params.jobId != null && String(params.jobId) !== '') qs.set('job_id', String(params.jobId))
    if (params.startedFrom) qs.set('started_from', params.startedFrom)
    if (params.startedTo) qs.set('started_to', params.startedTo)

    const raw = toRecord(await apiGet(`/runs?${qs.toString()}`))
    const items = Array.isArray(raw.items) ? raw.items.map(normalizeRunListItem) : []
    const total = toNumberValue(raw.total) ?? items.length

    return {items, total}
}

export async function fetchRunDetailData(runId: string): Promise<Pick<RunDetailData, 'run' | 'steps'>> {
    const raw = toRecord(await apiGet(`/runs/${encodeURIComponent(runId)}`))
    const run = normalizeRunRecord(raw.run ?? raw)
    const stepsRaw = Array.isArray(raw.steps) ? raw.steps : []
    return {
        run,
        steps: stepsRaw.map(toRunStepDetail),
    }
}

export async function fetchRunProgressSnapshot(runId: string): Promise<RunProgressSnapshot | null> {
    const raw = toRecord(await apiGet(`/runs/${encodeURIComponent(runId)}/progress`))
    return normalizeRunProgressSnapshot(raw.snapshot)
}

export async function fetchRunProgressData(runId: string): Promise<RunProgressData> {
    const raw = toRecord(await apiGet(`/runs/${encodeURIComponent(runId)}/progress`))
    const eventsRaw = Array.isArray(raw.events) ? raw.events : []
    return {
        snapshot: normalizeRunProgressSnapshot(raw.snapshot),
        events: eventsRaw
            .map(normalizeRunProgressEvent)
            .filter((event): event is RunProgressEvent => event !== null),
    }
}

export async function fetchRecentRunsForJob(jobId: string | number, limit: number): Promise<RunListItem[]> {
    const raw = toRecord(await apiGet(`/jobs/${encodeURIComponent(String(jobId))}/runs?limit=${encodeURIComponent(String(limit))}`))
    return Array.isArray(raw.items) ? raw.items.map(normalizeRunListItem) : []
}

export async function cancelRun(runId: string): Promise<void> {
    await apiPost(`/runs/${encodeURIComponent(runId)}/cancel`, {})
}

export async function rerunRun(runId: string): Promise<void> {
    await apiPost(`/runs/${encodeURIComponent(runId)}/rerun`, {})
}
