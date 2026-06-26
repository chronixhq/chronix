import React, {createContext, useCallback, useContext, useMemo} from 'react'
import type {FeatureAvailabilityStatus, FeatureUsageKind} from './types.ts'

interface FeatureAvailabilityCtxValue {
    data: FeatureAvailabilityStatus
    loading: boolean
    error?: string
    reload: () => Promise<void>
    checkLimit: (kind: FeatureUsageKind) => { allowed: boolean; message?: string }
}

const FeatureAvailabilityCtx = createContext<FeatureAvailabilityCtxValue | undefined>(undefined)

const featureAvailability: FeatureAvailabilityStatus = {
    usage: {
        agents: 0,
        jobs: 0,
        db_connections: 0,
        shell_connections: 0,
        webtask_connections: 0,
        actions: 0,
        users: 0,
    },
    features: {
        sms: true,
        webhooks: true,
        csvReports: true,
        htmlReports: true,
        pdfReports: true,
        branding: true,
    },
    feedbackEnabled: true,
    branding: {},
}

export function useFeatureAvailability(): FeatureAvailabilityCtxValue {
    const ctx = useContext(FeatureAvailabilityCtx)
    if (!ctx) throw new Error('useFeatureAvailability must be used within a FeatureAvailabilityProvider')
    return ctx
}

export function FeatureAvailabilityProvider({children}: { children: React.ReactNode }) {
    const reload = useCallback(async () => {
        return Promise.resolve()
    }, [])

    const checkLimit = useCallback((_kind: FeatureUsageKind): { allowed: boolean; message?: string } => {
        return { allowed: true }
    }, [])

    const value = useMemo(() => ({
        data: featureAvailability,
        loading: false,
        error: undefined,
        reload,
        checkLimit,
    }), [reload, checkLimit])

    return <FeatureAvailabilityCtx.Provider value={value}>{children}</FeatureAvailabilityCtx.Provider>
}
