import React, {createContext, useCallback, useContext, useMemo, useRef, useState} from 'react'
import type {AnyConnection} from '../modules/Connections/types.ts'
import {applyConnectionHealthPatch, fetchAllConnections} from '../modules/Connections/api.ts'
import {useSseContext} from './SseContext.tsx'

interface ConnectionsCtxValue {
    items: AnyConnection[]
    byId: Record<string, AnyConnection>
    loading: boolean
    hasLoaded: boolean
    error?: string
    reload: () => Promise<void>
    ensureLoaded: () => Promise<void>
}

const ConnectionsCtx = createContext<ConnectionsCtxValue | undefined>(undefined)

export function useConnections(): ConnectionsCtxValue {
    const ctx = useContext(ConnectionsCtx)
    if (!ctx) throw new Error('useConnections must be used within a ConnectionsProvider')
    return ctx
}

export function ConnectionsProvider({children}: { children: React.ReactNode }) {
    const {addSSEListener} = useSseContext()
    const [items, setItems] = useState<AnyConnection[]>([])
    const [loading, setLoading] = useState(false)
    const [hasLoaded, setHasLoaded] = useState(false)
    const [error, setError] = useState<string | undefined>(undefined)
    const inflightRef = useRef<Promise<void> | null>(null)

    const reload = useCallback(async () => {
        if (inflightRef.current) return inflightRef.current
        setLoading(true)
        setError(undefined)
        const task = (async () => {
            try {
                setItems(await fetchAllConnections());
                setHasLoaded(true)
            } catch (e: any) {
                console.error(e)
                setItems([])
                setError('Failed to load connections')
            } finally {
                setLoading(false)
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

    React.useEffect(() => {
        return addSSEListener<{id?: string | number; lastStatus?: string; lastError?: string | null; lastCheckedAt?: string | Date}>('connection_health', (patch) => {
            if (!patch?.id) return
            setItems((prev) => prev.map((connection) => applyConnectionHealthPatch(connection, patch)))
        })
    }, [addSSEListener])

    const byId = useMemo(() => {
        const map: Record<string, AnyConnection> = {}
        for (const c of items) map[`${c.kind}-${c.id}`] = c
        return map
    }, [items])

    const value = useMemo(() => ({items, byId, loading, hasLoaded, error, reload, ensureLoaded}), [items, byId, loading, hasLoaded, error, reload, ensureLoaded])
    return <ConnectionsCtx.Provider value={value}>{children}</ConnectionsCtx.Provider>
}
