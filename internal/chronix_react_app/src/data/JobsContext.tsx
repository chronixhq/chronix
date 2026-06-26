import React, {createContext, useCallback, useContext, useMemo, useRef, useState} from 'react'
import {apiDelete, apiPost} from '@dsherwin/react-api-interface'
import type {Job} from '../modules/Jobs/types'
import {fetchJobs} from '../modules/Jobs/api.ts'

interface JobsCtxValue {
    items: Job[]
    byId: Record<string, Job>
    loading: boolean
    refreshing: boolean
    hasLoaded: boolean
    error?: string
    reload: (opts?: { silent?: boolean }) => Promise<void>
    ensureLoaded: () => Promise<void>
    setEnabled: (id: string, enabled: boolean) => Promise<boolean>
    deleteJob: (id: string) => Promise<boolean>
}

const JobsCtx = createContext<JobsCtxValue | undefined>(undefined)

export function useJobs(): JobsCtxValue {
    const ctx = useContext(JobsCtx)
    if (!ctx) throw new Error('useJobs must be used within a JobsProvider')
    return ctx
}

export function JobsProvider({children}: { children: React.ReactNode }) {
    const [items, setItems] = useState<Job[]>([])
    const [loading, setLoading] = useState(false)
    const [refreshing, setRefreshing] = useState(false)
    const [hasLoaded, setHasLoaded] = useState(false)
    const [error, setError] = useState<string | undefined>(undefined)
    const inflightRef = useRef<Promise<void> | null>(null)

    const reload = useCallback(async (opts?: { silent?: boolean }) => {
        if (inflightRef.current) return inflightRef.current
        const silent = !!opts?.silent
        if (silent) setRefreshing(true); else setLoading(true)
        setError(undefined)
        const task = (async () => {
            try {
                setItems(await fetchJobs())
                setHasLoaded(true)
            } catch (e: any) {
                console.error(e)
                setItems([])
                setError('Failed to load jobs')
            } finally {
                if (silent) setRefreshing(false); else setLoading(false)
                inflightRef.current = null
            }
        })()
        inflightRef.current = task
        return task
    }, [])

    const ensureLoaded = useCallback(async () => {
        if (hasLoaded) return
        await reload()
    }, [hasLoaded, reload])

    const byId = useMemo(() => {
        const map: Record<string, Job> = {}
        for (const j of items) map[String(j.id)] = j
        return map
    }, [items])

    const setEnabled = useCallback(async (id: string, enabled: boolean) => {
        try {
            const path = `/jobs/${encodeURIComponent(id)}/${enabled ? 'enable' : 'disable'}`
            const res: any = await apiPost(path as any, {} as any)
            // Treat a successful call (no thrown error) as success. Some endpoints may return 204 with no body.
            if (res && typeof res === 'object' && 'ok' in res && (res as any).ok === false) {
                throw new Error('Request failed')
            }
            // optimistic update
            setItems(prev => prev.map(j => String(j.id) === String(id) ? {...j, enabled} : j))
            return true
        } catch (e) {
            console.error(e)
            return false
        }
    }, [])

    const deleteJob = useCallback(async (id: string) => {
        try {
            await apiDelete(`/jobs/${encodeURIComponent(id)}` as any)
            setItems(prev => prev.filter(j => String(j.id) !== String(id)))
            return true
        } catch (e) {
            console.error(e)
            return false
        }
    }, [])

    const value = useMemo(() => ({items, byId, loading, refreshing, hasLoaded, error, reload, ensureLoaded, setEnabled, deleteJob}), [items, byId, loading, refreshing, hasLoaded, error, reload, ensureLoaded, setEnabled, deleteJob])
    return <JobsCtx.Provider value={value}>{children}</JobsCtx.Provider>
}
