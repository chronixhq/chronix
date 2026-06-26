import {useEffect, useMemo, useRef, useState} from 'react'
import {apiGet} from '@dsherwin/react-api-interface'
import {useSseContext} from '../../data/SseContext'

export type JobActivityState = {
  isRunning: boolean
  currentRunId?: string
  lastStatus?: 'running'|'success'|'failed'|'canceled'|'idle'
  lastMessage?: string
  updatedAt?: Date
}

/**
 * useJobActivitySse subscribes to SSE job events for a given jobId and
 * exposes a simple activity state for UI chips. It also seeds from the most
 * recent run via REST as a baseline.
 */
export function useJobActivitySse(jobId?: number | string) {
  const { addSSEListener } = useSseContext()
  const [state, setState] = useState<JobActivityState>({ isRunning: false, lastStatus: 'idle' })
  const jobIdRef = useRef<string | undefined>(jobId ? String(jobId) : undefined)
  jobIdRef.current = jobId ? String(jobId) : undefined
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Seed baseline from last run
  useEffect(() => {
    if (!jobId) return
    let alive = true
    ;(async () => {
      try {
        const res = await apiGet(`/jobs/${encodeURIComponent(String(jobId))}/runs?limit=1`) as any
        const item = Array.isArray(res?.items) && res.items.length > 0 ? res.items[0] : undefined
        if (!alive) return
        if (item) {
          const lastStatus = String(item.status || 'idle') as JobActivityState['lastStatus']
          setState(prev => ({
            ...prev,
            isRunning: lastStatus === 'running',
            lastStatus: lastStatus as any,
            lastMessage: item.message ?? prev.lastMessage,
            currentRunId: item.runId ?? prev.currentRunId,
            updatedAt: new Date(item.startedAt || item.queuedAt || Date.now()),
          }))
        } else {
          setState(prev => ({ ...prev, isRunning: false, lastStatus: 'idle' }))
        }
      } catch {
        // ignore
      }
    })()
    return () => { alive = false }
  }, [jobId])

  // Helper: schedule inactivity expiry
  const bumpExpiry = () => {
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => {
      setState(prev => ({ ...prev, isRunning: false, lastStatus: prev.lastStatus === 'running' ? 'idle' : prev.lastStatus }))
    }, 15000) // 15s inactivity window
  }

  // Listen to SSE
  useEffect(() => {
    if (!jobId) return

    const onProgress = (payload: any) => {
      try {
        if (!payload || typeof payload !== 'object') return
        if (String(payload.job_id) !== jobIdRef.current) return
        bumpExpiry()
        setState(prev => ({
          ...prev,
          isRunning: true,
          lastStatus: 'running',
          currentRunId: payload.run_id ?? prev.currentRunId,
          updatedAt: new Date(),
        }))
      } catch {}
    }

    const onFinished = (payload: any) => {
      try {
        if (!payload || typeof payload !== 'object') return
        if (String(payload.job_id) !== jobIdRef.current) return
        if (timerRef.current) { clearTimeout(timerRef.current); timerRef.current = null }
        setState(prev => ({
          ...prev,
          isRunning: false,
          lastStatus: String(payload.status || 'idle') as any,
          lastMessage: payload.message ?? prev.lastMessage,
          currentRunId: payload.run_id ?? prev.currentRunId,
          updatedAt: new Date(),
        }))
      } catch {}
    }

    const unsub1 = addSSEListener<any>('job_progress', onProgress)
    const unsub2 = addSSEListener<any>('job_finished', onFinished)

    return () => { unsub1?.(); unsub2?.(); if (timerRef.current) clearTimeout(timerRef.current) }
  }, [addSSEListener, jobId])

  return useMemo(() => state, [state])
}
