import {beforeEach, describe, expect, it} from 'vitest'
import {setGlobalDisplayOptions} from '../../lib/datetime.ts'
import {formatScheduleSummary} from './scheduleSummary.ts'
import type {JobSchedule} from './types.ts'

describe('formatScheduleSummary', () => {
    beforeEach(() => {
        setGlobalDisplayOptions({timeZone: 'UTC', hour12: true})
    })

    it('describes manual jobs directly', () => {
        expect(formatScheduleSummary({kind: 'manual'})).toBe('Manual run only.')
    })

    it('describes a single-run job with date and time', () => {
        const schedule: JobSchedule = {
            kind: 'single',
            runAt: '2026-04-12T18:30:00.000Z',
        }

        const summary = formatScheduleSummary(schedule)
        expect(summary).toContain('Job will run on')
        expect(summary).toContain('2026')
        expect(summary).toContain('6:30')
    })

    it('describes recurring structured schedules with readable cadence text', () => {
        const summary = formatScheduleSummary({
            kind: 'recurring',
            mode: 'structured',
            startAt: '2026-04-12T00:00:00.000Z',
            rule: {
                freq: 'week',
                interval: 1,
                weekdays: [1, 3],
                time: '14:30',
            },
        })

        expect(summary).toContain('job will run each week on Mon and Wed at 2:30pm.')
    })
})
