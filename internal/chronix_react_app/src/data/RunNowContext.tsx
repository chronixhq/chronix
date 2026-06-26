import {createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState} from 'react'
import {apiPost} from '@dsherwin/react-api-interface'
import {useSseContext} from './SseContext'

export type RunCompletion = {
    runId: string
    status: string
    message?: string
}

export type RunNowContextValue = {
    // Multiple concurrently running Run Now panels
    activeRuns: string[]
    // Trigger a Run Now for a job. Returns runId (if available).
    runNow: (jobId: string | number, opts?: { jobName?: string }) => Promise<string | undefined>
    // Manually dismiss a specific run's panel (does not affect the run itself)
    dismiss: (runId: string) => void
    // Provide a friendly title for a run id, if known
    getRunTitle: (runId: string) => string | undefined
    // Whether a runId was initiated in this browser (used to decide dialog vs snack)
    isLocalRun: (runId: string) => boolean
    // Dialog queue of completed runs; show one at a time
    nextCompleted?: RunCompletion
    popCompleted: () => void
}

const Ctx = createContext<RunNowContextValue | undefined>(undefined)

export function RunsCommandsProvider({children}: { children: ReactNode }) {
    const {addSSEListener} = useSseContext()
    const [activeRuns, setActiveRuns] = useState<string[]>([])
    const [completedQueue, setCompletedQueue] = useState<RunCompletion[]>([])
    const [titles, setTitles] = useState<Record<string, string>>({})
    // Track which runs were initiated from this browser instance
    const localRunsRef = useRef<Set<string>>(new Set())

    const addActive = useCallback((rid: string) => {
        setActiveRuns(prev => (prev.includes(rid) ? prev : [...prev, rid]))
    }, [])
    const removeActive = useCallback((rid: string) => {
        setActiveRuns(prev => prev.filter(id => id !== rid))
    }, [])

    const runNow = useCallback(async (jobId: string | number, opts?: { jobName?: string }) => {
        // apiPost returns already-parsed JSON; it will throw on parse/network errors
        const res: any = await apiPost(`/jobs/${encodeURIComponent(String(jobId))}/runNow` as any, {} as any)

        // Our Go handler returns: { status: "queued", runId: <string> }
        // Some endpoints might wrap under { data: {...} }, so support both.
        let rid: string | undefined = res?.runId ?? res?.data?.runId ?? res?.run_id ?? res?.data?.run_id
        if (!rid) {
            const runObj = res?.run ?? res?.data?.run
            rid = runObj?.runId ?? runObj?.run_id ?? runObj?.id
        }
        if (!rid) {
            throw new Error('Run Now did not return a runId')
        }
        const ridStr = String(rid)
        addActive(ridStr)
        // Mark this run as initiated locally so we can show a dialog only here
        localRunsRef.current.add(ridStr)
        if (opts?.jobName) {
            setTitles(prev => ({ ...prev, [ridStr]: `Run now — ${opts.jobName}` }))
        }
        return ridStr
    }, [addActive])

    // Listen globally for run finished events to remove panels and queue a dialog
    const recentFinished = useRef<Set<string>>(new Set())
    useEffect(() => {
        const unsub = addSSEListener<any>('job_finished', (p) => {
            try {
                if (!p || typeof p !== 'object') return
                const rid = String(p.run_id || '')
                if (!rid) return
                if (recentFinished.current.has(rid)) return
                recentFinished.current.add(rid)
                // We no longer remove from activeRuns immediately; 
                // LiveRunNowProgressPanel will handle its own 15s auto-dismiss.
                // removeActive(rid)
                // Only queue a completion dialog if this run was initiated in this browser
                if (localRunsRef.current.has(rid)) {
                    const status = String(p.status || 'finished')
                    const message = p.message ? String(p.message) : undefined
                    setCompletedQueue(prev => [...prev, {runId: rid, status, message}])
                }
                // Garbage collect after some time to allow re-runs
                setTimeout(() => {
                    recentFinished.current.delete(rid)
                }, 60_000)
            } catch {
            }
        })
        return () => {
            unsub?.()
        }
    }, [addSSEListener, removeActive])

    const dismiss = useCallback((runId: string) => {
        removeActive(runId)
    }, [removeActive])

    const popCompleted = useCallback(() => {
        setCompletedQueue(prev => prev.slice(1))
    }, [])

    const getRunTitle = useCallback((runId: string) => titles[runId], [titles])

    const isLocalRun = useCallback((rid: string) => localRunsRef.current.has(rid), [])

    const value = useMemo<RunNowContextValue>(() => ({
        activeRuns,
        runNow,
        dismiss,
        getRunTitle,
        isLocalRun,
        nextCompleted: completedQueue[0],
        popCompleted,
    }), [activeRuns, runNow, dismiss, getRunTitle, isLocalRun, completedQueue, popCompleted])

    return (
        <Ctx.Provider value={value}>{children}</Ctx.Provider>
    )
}

export function useRunNow() {
    const ctx = useContext(Ctx)
    if (!ctx) throw new Error('useRunNow must be used within RunsCommandsProvider')
    return ctx
}
