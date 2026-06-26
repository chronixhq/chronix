import React, {createContext, useCallback, useContext, useEffect, useMemo, useRef, useState} from 'react'
import {useSseContext} from './SseContext'
import {
    fetchRecentRunsForJob,
    fetchRunDetailData,
    fetchRunProgressSnapshot,
    fetchRunsListPage,
} from '../modules/Runs/api.ts'
import type {
    JobFinishedPayload,
    JobProgressPayload,
    RunDetailData,
    RunListItem,
    RunProgressSnapshot,
    RunsListResponse,
} from '../modules/Runs/types.ts'

type ListKey = string

type RunsListParams = {
    limit: number
    offset: number
    q?: string
    status?: string
    jobId?: string | number
    startedFrom?: string
    startedTo?: string
}

type RunsContextValue = {
    useRunsList: (params: RunsListParams) => {
        items: RunListItem[]
        total: number
        loading: boolean
        error?: string
        reload: () => void
    }
    useRun: (runId?: string) => {
        run: RunDetailData['run']
        steps: RunDetailData['steps']
        snapshot: RunDetailData['snapshot']
        loading: boolean
        error?: string
        reload: () => void
    }
    useRecentRunsForJob: (jobId?: string | number, limit?: number) => {
        items: RunListItem[]
        loading: boolean
        error?: string
        reload: () => void
    }
}

interface ListState extends RunsListResponse {
    loading: boolean
    error?: string
}

interface RunsStore {
    lists: Record<ListKey, ListState>
    byId: Record<string, RunListItem>
    details: Record<string, RunDetailData>
    upsertById: (items: RunListItem[]) => void
    fetchRunsList: (params: RunsListParams) => Promise<void> | void
    fetchRunDetail: (runId?: string) => Promise<void> | void
    fetchRunProgress: (runId?: string) => Promise<void> | void
}

const RunsStoreCtx = createContext<RunsStore | undefined>(undefined)
const RunsCtx = createContext<RunsContextValue | undefined>(undefined)

function useRunsStore(): RunsStore {
    const ctx = useContext(RunsStoreCtx)
    if (!ctx) throw new Error('useRunsStore must be used within a RunsProvider')
    return ctx
}

function makeListKey(params: RunsListParams): ListKey {
    const parts = [
        `l=${params.limit}`,
        `o=${params.offset}`,
        params.q ? `q=${encodeURIComponent(params.q)}` : '',
        params.status ? `s=${encodeURIComponent(params.status)}` : '',
        params.jobId != null && String(params.jobId) !== '' ? `j=${encodeURIComponent(String(params.jobId))}` : '',
        params.startedFrom ? `sf=${encodeURIComponent(params.startedFrom)}` : '',
        params.startedTo ? `st=${encodeURIComponent(params.startedTo)}` : '',
    ].filter(Boolean)
    return parts.join('&')
}

function mergeRunItems(current: Record<string, RunListItem>, items: RunListItem[]): Record<string, RunListItem> {
    if (!items.length) return current
    const next = {...current}
    for (const item of items) {
        next[item.runId] = {...(current[item.runId] || {}), ...item}
    }
    return next
}

function parseListKey(key: string): { limit: number; offset: number } {
    let limit = 0
    let offset = 0
    for (const part of key.split('&')) {
        if (part.startsWith('l=')) {
            const parsed = Number(decodeURIComponent(part.slice(2)))
            if (Number.isFinite(parsed)) limit = parsed
        } else if (part.startsWith('o=')) {
            const parsed = Number(decodeURIComponent(part.slice(2)))
            if (Number.isFinite(parsed)) offset = parsed
        }
    }
    return {limit, offset}
}

function getErrorMessage(error: unknown, fallback: string): string {
    if (error instanceof Error && error.message.trim()) return error.message
    return fallback
}

function getRunId(value: string | number | undefined): string {
    return value != null ? String(value) : ''
}

function patchRunSnapshot(snapshot: RunProgressSnapshot | null | undefined, patch: Partial<RunProgressSnapshot>): RunProgressSnapshot {
    return {
        ...(snapshot || {}),
        ...patch,
    }
}

export function useRunsList(params: RunsListParams) {
    const {lists, byId, fetchRunsList} = useRunsStore()
    const {limit, offset, q, status, jobId, startedFrom, startedTo} = params
    const stableParams = useMemo(
        () => ({limit, offset, q, status, jobId, startedFrom, startedTo}),
        [jobId, limit, offset, q, startedFrom, startedTo, status],
    )
    const key = useMemo(() => makeListKey(stableParams), [stableParams])
    const state = lists[key]

    useEffect(() => {
        void fetchRunsList(stableParams)
    }, [fetchRunsList, stableParams])

    const reload = useCallback(() => {
        void fetchRunsList(stableParams)
    }, [fetchRunsList, stableParams])

    return {
        items: (state?.items || []).map((item) => byId[item.runId] || item),
        total: state?.total || 0,
        loading: !!state?.loading,
        error: state?.error,
        reload,
    }
}

export function useRun(runId?: string) {
    const {details, fetchRunDetail, fetchRunProgress} = useRunsStore()
    const detail = runId ? details[runId] : undefined
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | undefined>(undefined)

    useEffect(() => {
        if (!runId) return
        let alive = true

        const load = async () => {
            setLoading(true)
            setError(undefined)
            try {
                await Promise.all([fetchRunDetail(runId), fetchRunProgress(runId)])
            } catch (err) {
                if (!alive) return
                setError(getErrorMessage(err, 'Failed to load run'))
            } finally {
                if (alive) setLoading(false)
            }
        }

        void load()
        return () => {
            alive = false
        }
    }, [fetchRunDetail, fetchRunProgress, runId])

    const reload = useCallback(() => {
        if (!runId) return
        void fetchRunDetail(runId)
        void fetchRunProgress(runId)
    }, [fetchRunDetail, fetchRunProgress, runId])

    return {
        run: detail?.run ?? null,
        steps: detail?.steps ?? [],
        snapshot: detail?.snapshot ?? null,
        loading,
        error,
        reload,
    }
}

export function useRecentRunsForJob(jobId?: string | number, limit: number = 20) {
    const {upsertById} = useRunsStore()
    const [state, setState] = useState<{ items: RunListItem[]; loading: boolean; error?: string }>({
        items: [],
        loading: !!jobId,
    })

    const load = useCallback(async () => {
        if (!jobId) return
        setState((prev) => ({...prev, loading: true, error: undefined}))
        try {
            const items = await fetchRecentRunsForJob(jobId, limit)
            upsertById(items)
            setState({items, loading: false})
        } catch (err) {
            setState({items: [], loading: false, error: getErrorMessage(err, 'Failed to load runs')})
        }
    }, [jobId, limit, upsertById])

    const key = useMemo(() => `job:${jobId}:limit:${limit}`, [jobId, limit])
    useEffect(() => {
        void load()
    }, [key, load])

    return {items: state.items, loading: state.loading, error: state.error, reload: load}
}

export function RunsProvider({children}: { children: React.ReactNode }) {
    const {addSSEListener} = useSseContext()

    const [lists, setLists] = useState<Record<ListKey, ListState>>({})
    const [byId, setById] = useState<Record<string, RunListItem>>({})
    const [details, setDetails] = useState<Record<string, RunDetailData>>({})

    const inflightLists = useRef(new Map<ListKey, Promise<void>>())
    const inflightDetails = useRef(new Map<string, Promise<void>>())
    const inflightUpserts = useRef(new Map<string, Promise<void>>())
    const byIdRef = useRef<Record<string, RunListItem>>({})

    useEffect(() => {
        byIdRef.current = byId
    }, [byId])

    const upsertById = useCallback((items: RunListItem[]) => {
        setById((prev) => mergeRunItems(prev, items))
    }, [])

    const fetchRunsList = useCallback(async (params: RunsListParams) => {
        const key = makeListKey(params)
        setLists((prev) => (prev[key] ? prev : {...prev, [key]: {items: [], total: 0, loading: true}}))

        if (inflightLists.current.has(key)) {
            return inflightLists.current.get(key)
        }

        const task = (async () => {
            try {
                const response = await fetchRunsListPage(params)
                setById((prev) => mergeRunItems(prev, response.items))
                setLists((prev) => ({
                    ...prev,
                    [key]: {
                        items: response.items,
                        total: response.total,
                        loading: false,
                    },
                }))
            } catch (err) {
                setLists((prev) => ({
                    ...prev,
                    [key]: {
                        items: [],
                        total: 0,
                        loading: false,
                        error: getErrorMessage(err, 'Failed to load runs'),
                    },
                }))
            } finally {
                inflightLists.current.delete(key)
            }
        })()

        inflightLists.current.set(key, task)
        return task
    }, [])

    const fetchRunDetail = useCallback(async (runId?: string) => {
        if (!runId) return
        if (inflightDetails.current.has(runId)) {
            return inflightDetails.current.get(runId)
        }

        const task = (async () => {
            try {
                const detail = await fetchRunDetailData(runId)
                const runItem = detail.run
                setDetails((prev) => ({
                    ...prev,
                    [runId]: {
                        ...(prev[runId] || {run: null, steps: [], snapshot: null}),
                        run: runItem,
                        steps: detail.steps,
                    },
                }))
                if (runItem) {
                    setById((prev) => mergeRunItems(prev, [runItem]))
                }
            } finally {
                inflightDetails.current.delete(runId)
            }
        })()

        inflightDetails.current.set(runId, task)
        return task
    }, [])

    const ensureRunPresent = useCallback(async (runId: string) => {
        if (!runId || byIdRef.current[runId]) return
        if (inflightUpserts.current.has(runId)) {
            return inflightUpserts.current.get(runId)
        }

        const task = (async () => {
            try {
                const detail = await fetchRunDetailData(runId)
                const runItem = detail.run
                if (!runItem) return

                setById((prev) => mergeRunItems(prev, [runItem]))
                setDetails((prev) => ({
                    ...prev,
                    [runId]: {
                        ...(prev[runId] || {run: null, steps: [], snapshot: null}),
                        run: runItem,
                        steps: detail.steps,
                    },
                }))
                setLists((prev) => {
                    const next = {...prev}
                    for (const [key, state] of Object.entries(prev)) {
                        const {limit, offset} = parseListKey(key)
                        if (offset !== 0) continue

                        const exists = state.items.some((item) => item.runId === runItem.runId)
                        const items = exists
                            ? state.items.map((item) => (item.runId === runItem.runId ? {...item, ...runItem} : item))
                            : [runItem, ...state.items]

                        next[key] = {
                            ...state,
                            items: limit > 0 ? items.slice(0, limit) : items,
                            total: state.total + (exists ? 0 : 1),
                        }
                    }
                    return next
                })
            } finally {
                inflightUpserts.current.delete(runId)
            }
        })()

        inflightUpserts.current.set(runId, task)
        return task
    }, [])

    const fetchRunProgress = useCallback(async (runId?: string) => {
        if (!runId) return
        try {
            const snapshot = await fetchRunProgressSnapshot(runId)
            setDetails((prev) => ({
                ...prev,
                [runId]: {
                    ...(prev[runId] || {run: null, steps: [], snapshot: null}),
                    snapshot,
                },
            }))
        } catch {
            // ignore progress fetch failures
        }
    }, [])

    useEffect(() => {
        const onProgress = (payload: JobProgressPayload) => {
            const runId = getRunId(payload.run_id)
            if (!runId) return

            const nowIso = new Date().toISOString()
            const eventType = String(payload.type || '')
            const nextStatus = eventType.toLowerCase().includes('queued') ? 'queued' : 'running'
            let hadItem = false

            setById((prev) => {
                const current = prev[runId]
                if (!current) return prev

                hadItem = true
                const normalizedStatus = String(current.status || '').toLowerCase()
                if (normalizedStatus && normalizedStatus !== 'running' && normalizedStatus !== 'queued') {
                    return prev
                }

                return {
                    ...prev,
                    [runId]: {
                        ...current,
                        status: nextStatus,
                        startedAt: nextStatus === 'running' ? (current.startedAt || nowIso) : current.startedAt,
                    },
                }
            })

            if (!hadItem) {
                void ensureRunPresent(runId)
            }

            setDetails((prev) => ({
                ...prev,
                [runId]: {
                    ...(prev[runId] || {run: null, steps: [], snapshot: null}),
                    snapshot: patchRunSnapshot(prev[runId]?.snapshot, {status: nextStatus, updatedAt: nowIso}),
                },
            }))
        }

        const onFinished = (payload: JobFinishedPayload) => {
            const runId = getRunId(payload.run_id)
            if (!runId) return

            const nowIso = new Date().toISOString()
            const status = String(payload.status || 'success').toLowerCase()
            const message = payload.message
            let hadItem = false

            setById((prev) => {
                const current = prev[runId]
                if (!current) return prev

                hadItem = true
                return {
                    ...prev,
                    [runId]: {
                        ...current,
                        status,
                        finishedAt: current.finishedAt || nowIso,
                        message: message ?? current.message,
                    },
                }
            })

            if (!hadItem) {
                void ensureRunPresent(runId)
            }

            setDetails((prev) => ({
                ...prev,
                [runId]: {
                    ...(prev[runId] || {run: null, steps: [], snapshot: null}),
                    snapshot: patchRunSnapshot(prev[runId]?.snapshot, {status, message, updatedAt: nowIso}),
                },
            }))
        }

        const unsubscribeProgress = addSSEListener<JobProgressPayload>('job_progress', onProgress)
        const unsubscribeFinished = addSSEListener<JobFinishedPayload>('job_finished', onFinished)

        return () => {
            unsubscribeProgress?.()
            unsubscribeFinished?.()
        }
    }, [addSSEListener, ensureRunPresent])

    const store = useMemo<RunsStore>(() => ({
        lists,
        byId,
        details,
        upsertById,
        fetchRunsList,
        fetchRunDetail,
        fetchRunProgress,
    }), [lists, byId, details, upsertById, fetchRunsList, fetchRunDetail, fetchRunProgress])

    const value = useMemo<RunsContextValue>(() => ({
        useRunsList,
        useRun,
        useRecentRunsForJob,
    }), [])

    return (
        <RunsStoreCtx.Provider value={store}>
            <RunsCtx.Provider value={value}>{children}</RunsCtx.Provider>
        </RunsStoreCtx.Provider>
    )
}

export function useRunsContext(): RunsContextValue {
    const ctx = useContext(RunsCtx)
    if (!ctx) throw new Error('useRunsContext must be used within a RunsProvider')
    return ctx
}
