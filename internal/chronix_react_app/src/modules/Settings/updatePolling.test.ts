import {afterEach, describe, expect, it, vi} from 'vitest'
import {waitForServerVersion} from './updatePolling.ts'
import type {UpdaterStatus} from './types.ts'

function makeStatus(currentVersion: string): UpdaterStatus {
    return {
        currentVersion,
        availableVersion: {
            server: {
                version: '0.0.30',
                release_date: '2026-04-12',
                release_notes: 'Release notes',
            },
        },
        lastCheckTime: '2026-04-12T00:00:00Z',
        enabled: true,
        mode: 'notify',
        windowStart: '',
        agentEnabled: true,
        agentMode: 'notify',
        agentWindowStart: '',
    }
}

describe('waitForServerVersion', () => {
    afterEach(() => {
        vi.useRealTimers()
    })

    it('keeps polling until the target version is reported', async () => {
        vi.useFakeTimers()
        const fetchStatus = vi.fn<() => Promise<UpdaterStatus>>()
            .mockResolvedValueOnce(makeStatus('0.0.29'))
            .mockRejectedValueOnce(new Error('server restarting'))
            .mockResolvedValueOnce(makeStatus('0.0.30'))

        const waitPromise = waitForServerVersion(fetchStatus, '0.0.30', {
            initialDelayMs: 1000,
            intervalMs: 2000,
            maxAttempts: 5,
        })
        const assertion = expect(waitPromise).resolves.toMatchObject({currentVersion: '0.0.30'})

        await vi.runAllTimersAsync()

        await assertion
        expect(fetchStatus).toHaveBeenCalledTimes(3)
    })

    it('fails with a useful error after exhausting attempts', async () => {
        vi.useFakeTimers()
        const fetchStatus = vi.fn<() => Promise<UpdaterStatus>>().mockResolvedValue(makeStatus('0.0.29'))

        const waitPromise = waitForServerVersion(fetchStatus, '0.0.30', {
            initialDelayMs: 0,
            intervalMs: 1000,
            maxAttempts: 3,
        })
        const assertion = expect(waitPromise).rejects.toThrow('last reported version was 0.0.29')

        await vi.runAllTimersAsync()

        await assertion
        expect(fetchStatus).toHaveBeenCalledTimes(3)
    })
})
