import {createContext, type ReactNode, useContext, useEffect, useMemo, useRef, useState} from 'react'
import {type SSEMessage, SSEProvider} from '@dsherwin/react-sse'
import {useAuthContext} from './useAuthContext'
import {apiGet} from '@dsherwin/react-api-interface'
import {isAPIError} from '../lib/errors'
import {baseURL} from '../lib/api_config'

interface SseContextValue {
    // Subscribe to SSE events dispatched by the SSE bus. Returns an unsubscribe function.
    addSSEListener: <T = any>(event: string, handler: (data: T) => void) => () => void
    connectionState: 'connected' | 'reconnecting' | 'disconnected'
    retryCount: number
    showBanner: boolean
}

const Ctx = createContext<SseContextValue | undefined>(undefined)

class SSEBus {
    private target = new EventTarget()

    add<T = any>(event: string, handler: (data: T) => void): () => void {
        const wrapped = (e: Event) => handler((e as CustomEvent).detail as T)
        this.target.addEventListener(event, wrapped as EventListener)
        return () => this.target.removeEventListener(event, wrapped as EventListener)
    }

    emit(event: string, detail: any) {
        this.target.dispatchEvent(new CustomEvent(event, {detail}))
    }
}

export const SseProvider = ({children}: { children: ReactNode }) => {
    const {loggedIn, setUser, logout} = useAuthContext()
    const busRef = useRef<SSEBus | null>(null)
    if (!busRef.current) busRef.current = new SSEBus()

    const [connectionState, setConnectionState] = useState<'connected' | 'reconnecting' | 'disconnected'>('connected')
    const [retryCount, setRetryCount] = useState(0)
    const [showBanner, setShowBanner] = useState(false)
    const graceTimer = useRef<number | null>(null)
    const escalateTimer = useRef<number | null>(null)
    const showWarningAfterMs = 5000
    const escalateAfterMs = 30000

    const url = baseURL + '/sse'

    const clearTimers = () => {
        if (graceTimer.current) {
            window.clearTimeout(graceTimer.current);
            graceTimer.current = null
        }
        if (escalateTimer.current) {
            window.clearTimeout(escalateTimer.current);
            escalateTimer.current = null
        }
    }

    const onEvent = (msg: SSEMessage) => {
        const {type, dataStr} = msg
        let data: any = dataStr
        try {
            data = JSON.parse(dataStr)
        } catch {
        }

        if (type === 'userUpdate' && data && typeof data === 'object' && typeof data.id === 'number') {
            setUser((prev) => {
                if (!prev || (prev as any).id !== data.id) return prev
                return {...prev, email: data.email ?? (prev as any).email, name: data.name ?? (prev as any).name, phone: data.phone ?? (prev as any).phone, admin: typeof data.admin === 'boolean' ? data.admin : (prev as any).admin}
            })
        }
        if (type === 'logout') {
            let reason = 'You have been logged out because your account was changed by an administrator.'
            try {
                if (data && typeof data.reason === 'string' && data.reason.trim()) reason = data.reason
            } catch {
            }
            try {
                window.localStorage.setItem('logoutReason', reason)
            } catch {
            }
            logout()
            data = {reason}
        }

        busRef.current?.emit(type, data)
    }

    const handleOpen = async () => {
        setRetryCount(0)
        setConnectionState('connected')
        setShowBanner(false)
        clearTimers()
        try {
            await apiGet('/checkauth')
        } catch (e) {
            if (isAPIError(e)) {
                const anyErr = e as any
                const httpStatus = Number(anyErr.status)
                const apiCodeRaw = anyErr.APIErrorData?.code
                const apiCode = typeof apiCodeRaw === 'number' ? apiCodeRaw : (typeof apiCodeRaw === 'string' ? apiCodeRaw.toLowerCase() : undefined)
                const isUnauth = httpStatus === 401 || apiCode === 'unauthenticated' || apiCode === 'unauthorized' || apiCode === 401
                if (isUnauth) {
                    logout()
                }
            }
        }
    }

    const handleReconnect = async () => {
        // Same semantics as open, but distinct hook provided by library
        await handleOpen()
    }

    const handleError = () => {
        if (connectionState === 'connected') {
            setConnectionState('reconnecting')
            if (!graceTimer.current) {
                graceTimer.current = window.setTimeout(() => {
                    setShowBanner(true)
                }, showWarningAfterMs) as any
            }
            if (!escalateTimer.current) {
                escalateTimer.current = window.setTimeout(() => {
                    setConnectionState('disconnected')
                }, escalateAfterMs) as any
            }
        }
        if (connectionState === 'disconnected') {
            logout();
        }
        setRetryCount((c) => c + 1)
    }

    // Track browser offline
    useEffect(() => {
        const goOffline = () => handleError()
        const goOnline = () => {
        }
        window.addEventListener('offline', goOffline)
        window.addEventListener('online', goOnline)
        return () => {
            window.removeEventListener('offline', goOffline)
            window.removeEventListener('online', goOnline)
        }
    }, []) // eslint-disable-line react-hooks/exhaustive-deps -- [deps-intentional] mount-only listener setup; handlers are stable

    const addSSEListener = useMemo(() => {
        return (<T = any>(event: string, handler: (data: T) => void) => busRef.current!.add<T>(event, handler))
    }, [])

    const value = useMemo<SseContextValue>(() => ({addSSEListener, connectionState, retryCount, showBanner}), [addSSEListener, connectionState, retryCount, showBanner])

    return (
        <SSEProvider
            enabled={loggedIn}
            connections={[{
                id: 'chronix',
                url,
                withCredentials: true,
                eventTypes: ['notification', 'job_progress', 'job_finished', 'userUpdate', 'logout', 'agent_registration', 'agent_registration_approved', 'agent_registration_denied', 'agent_deleted', 'connection_health'],
            }]}
            onEvent={onEvent}
            onOpen={handleOpen}
            onError={() => {
                handleError()
            }}
            onReconnect={handleReconnect}
            onStatusChange={(_, status, state) => {
                if (status === 'error') {
                    // prefer library-provided count; do not auth-check here to avoid logout on transient errors
                    setRetryCount((state as any)?.consecutiveErrorCount ?? ((c: number) => c + 1) as any)
                }
                // Note: do not call /checkauth on 'closed' either; we'll verify only on a successful reconnect/open.
            }}
        >
            <Ctx.Provider value={value}>{children}</Ctx.Provider>
        </SSEProvider>
    )
}

export function useSseContext(): SseContextValue {
    const ctx = useContext(Ctx)
    if (!ctx) throw new Error('useSseContext must be used within an SseProvider')
    return ctx
}
