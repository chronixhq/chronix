import {useEffect, useRef, useState} from 'react'
import {useSseContext} from '../../data/SseContext'
import {fetchRunProgressData} from './api.ts'
import type {JobFinishedPayload, JobProgressPayload, RunProgressMessage} from './types.ts'

type CurrentStep = {
    index?: number
    name?: string
}

export type UseRunProgressState = {
    messages: RunProgressMessage[]
    currentStep?: CurrentStep
    status?: string
    finished: boolean
    refreshToken: number
}

function isFinishedStatus(status?: string): boolean {
    const normalized = String(status || '').toLowerCase()
    return normalized === 'success' || normalized === 'error' || normalized === 'canceled' || normalized === 'cancelled'
}

function toMessageText(type: string, stepIndex?: number, stepName?: string, message?: string): string {
    const parts: string[] = []
    if (stepIndex != null) parts.push(`step ${stepIndex}`)
    if (stepName) parts.push(stepName)
    if (message) parts.push(message)
    return parts.join(' - ') || type
}

function toStepIndex(value: unknown): number | undefined {
    return typeof value === 'number' && Number.isFinite(value) ? value : undefined
}

export function useRunProgressSse(runId?: string) {
    const {addSSEListener} = useSseContext()
    const [state, setState] = useState<UseRunProgressState>({messages: [], finished: false, refreshToken: 0})
    const runIdRef = useRef<string | undefined>(runId)
    runIdRef.current = runId

    useEffect(() => {
        setState({messages: [], finished: false, refreshToken: 0})
    }, [runId])

    useEffect(() => {
        if (!runId) return

        let mounted = true

        const loadInitial = async () => {
            try {
                const data = await fetchRunProgressData(runId)
                if (!mounted) return

                const initialMessages = data.events.map((event) => ({
                    ts: new Date(event.ts),
                    text: toMessageText(event.type, event.stepIndex, event.stepName, event.message),
                    type: event.type,
                    stepIndex: event.stepIndex,
                    stepName: event.stepName,
                    message: event.message,
                }))

                setState({
                    messages: initialMessages,
                    finished: isFinishedStatus(data.snapshot?.status),
                    status: data.snapshot?.status,
                    currentStep: undefined,
                    refreshToken: 0,
                })
            } catch (error) {
                console.error('Failed to load initial progress', error)
            }
        }

        void loadInitial()

        const onProgress = (payload: JobProgressPayload) => {
            const payloadRunId = payload.run_id != null ? String(payload.run_id) : ''
            if (!payloadRunId || payloadRunId !== runIdRef.current) return

            const stepIndex = toStepIndex(payload.step_index)
            const stepName = payload.step_name
            const message = payload.message
            const type = String(payload.type || 'progress')
            const status = type.toLowerCase().includes('queued') ? 'queued' : 'running'
            const shouldRefresh = type !== 'StepProgress'

            setState((prev) => ({
                messages: [
                    ...prev.messages,
                    {
                        ts: new Date(),
                        text: toMessageText(type, stepIndex, stepName, message),
                        type,
                        stepIndex,
                        stepName,
                        message,
                    },
                ].slice(-200),
                currentStep: stepIndex != null || stepName ? {index: stepIndex, name: stepName} : prev.currentStep,
                status,
                finished: false,
                refreshToken: prev.refreshToken + (shouldRefresh ? 1 : 0),
            }))
        }

        const onFinished = (payload: JobFinishedPayload) => {
            const payloadRunId = payload.run_id != null ? String(payload.run_id) : ''
            if (!payloadRunId || payloadRunId !== runIdRef.current) return

            const status = String(payload.status || 'finished')
            const message = payload.message

            setState((prev) => ({
                messages: [
                    ...prev.messages,
                    {
                        ts: new Date(),
                        text: `Run ${status}${message ? ` - ${message}` : ''}`,
                        type: 'RunFinished',
                        message,
                    },
                ].slice(-200),
                currentStep: prev.currentStep,
                status,
                finished: true,
                refreshToken: prev.refreshToken + 1,
            }))
        }

        const unsubscribeProgress = addSSEListener<JobProgressPayload>('job_progress', onProgress)
        const unsubscribeFinished = addSSEListener<JobFinishedPayload>('job_finished', onFinished)

        return () => {
            mounted = false
            unsubscribeProgress?.()
            unsubscribeFinished?.()
        }
    }, [addSSEListener, runId])

    return state
}
