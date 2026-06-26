import React, {createContext, useCallback, useContext, useEffect, useMemo, useState} from 'react'
import {apiGet, apiPost} from '@dsherwin/react-api-interface'
import type {NotificationItem} from './types'
import {useSseContext} from './SseContext'

interface RecentResponse { items: NotificationItem[]; unseenCount: number }

interface NotificationsCtx {
  items: NotificationItem[]
  unseenCount: number
  refresh: () => Promise<void>
  markSeen: (ids: number[]) => Promise<void>
}

const Ctx = createContext<NotificationsCtx | undefined>(undefined)

export const useNotifications = (): NotificationsCtx => {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('useNotifications must be used within NotificationsProvider')
  return ctx
}

export const NotificationsProvider = ({children}: {children: React.ReactNode}) => {
  const { addSSEListener } = useSseContext()
  const [items, setItems] = useState<NotificationItem[]>([])
  const [unseenCount, setUnseenCount] = useState(0)

  const refresh = useCallback(async () => {
    try {
      const res = await apiGet('/notifications/recent?limit=20') as RecentResponse
      setItems(res.items)
      setUnseenCount(res.unseenCount)
    } catch (e) {
      console.error(e);
    }
  }, [])

  const markSeen = useCallback(async (ids: number[]) => {
    if (!ids.length) return
    try {
      await apiPost('/notifications/mark-seen', { ids })
      // optimistic update
      setItems(prev => prev.map(it => ids.includes(it.id) ? {...it, seen: true} : it))
      setUnseenCount(prev => Math.max(0, prev - ids.filter(id => items.find(it => it.id === id && !it.seen)).length))
    } catch (e) {
      console.error(e);
      // ignore and let polling correct later
    }
  }, [items])

  useEffect(() => {
    // One-time fetch on mount/login to populate notifications before SSE events arrive
    void refresh()
  }, [refresh])

  // Subscribe to SSE notification events from AuthContext (hybrid approach)
  useEffect(() => {
    interface NotificationSSEPayload {
      id: number
      item?: NotificationItem
      unseenDelta?: number
      unseenCount?: number
    }
    const unsubscribe = addSSEListener<NotificationSSEPayload>('notification', (payload) => {
      try {
        if (!payload || typeof payload !== 'object') { void refresh(); return }
        // Upsert the item if provided
        if (payload.item) {
          setItems(prev => {
            const exists = prev.some(i => i.id === payload.item!.id)
            const next = exists
              ? prev.map(i => i.id === payload.item!.id ? { ...i, ...payload.item! } : i)
              : [payload.item!, ...prev]
            return next.slice(0, 20)
          })
        }
        // Update unseen count
        if (typeof payload.unseenCount === 'number') {
          setUnseenCount(payload.unseenCount)
        } else if (typeof payload.unseenDelta === 'number') {
          setUnseenCount(prev => Math.max(0, prev + payload.unseenDelta!))
        } else if (!payload.item) {
          // Fallback for legacy events: refresh list
          void refresh()
        }
      } catch {
        void refresh()
      }
    })
    return () => { unsubscribe?.() }
  }, [addSSEListener, refresh])

  const value = useMemo(() => ({ items, unseenCount, refresh, markSeen }), [items, unseenCount, refresh, markSeen])
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}
