import React, {createContext, useCallback, useContext, useMemo, useRef, useState} from 'react'
import type {Action} from '../modules/Actions/types'
import {fetchAllActions} from '../modules/Actions/api.ts'

interface ActionsCtxValue {
  items: Action[]
  byId: Record<string, Action>
  loading: boolean
  hasLoaded: boolean
  error?: string
  reload: () => Promise<void>
  ensureLoaded: () => Promise<void>
}

const ActionsCtx = createContext<ActionsCtxValue | undefined>(undefined)

export function useActions(): ActionsCtxValue {
  const ctx = useContext(ActionsCtx)
  if (!ctx) throw new Error('useActions must be used within an ActionsProvider')
  return ctx
}

export function ActionsProvider({children}: {children: React.ReactNode}) {
  const [items, setItems] = useState<Action[]>([])
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
        setItems(await fetchAllActions());
        setHasLoaded(true)
      } catch (e: any) {
        console.error(e)
        setItems([])
        setError('Failed to load actions')
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

  const byId = useMemo(() => {
    const map: Record<string, Action> = {}
    for (const a of items) map[`${a.actionType}-${a.id}`] = a
    return map
  }, [items])

  const value = useMemo(() => ({ items, byId, loading, hasLoaded, error, reload, ensureLoaded }), [items, byId, loading, hasLoaded, error, reload, ensureLoaded])
  return <ActionsCtx.Provider value={value}>{children}</ActionsCtx.Provider>
}
