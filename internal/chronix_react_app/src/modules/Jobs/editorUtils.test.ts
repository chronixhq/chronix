import {describe, expect, it} from 'vitest'
import {cronLooksValid, extractVarsFromAction, ordinal} from './editorUtils.ts'
import type {Action} from '../Actions/types.ts'

describe('job editor utils', () => {
    it('extracts template variables from mixed action step content', () => {
        const action: Action = {
            id: 'a1',
            name: 'Test action',
            actionType: 'database',
            steps: [
                {
                    sqlText: 'select * from jobs where id={{job_id}} and owner=${owner}',
                    headers: {'X-Trace-Id': '{{trace_id}}'},
                    expectation: {kind: 'fieldEqualsFirst', column: 'status', expected: '{{expected_status}}'},
                    env: {TOKEN: '${token}'},
                }
            ],
        }

        expect(extractVarsFromAction(action)).toEqual(['expected_status', 'job_id', 'owner', 'token', 'trace_id'])
    })

    it('validates simple cron strings and ordinal labels', () => {
        expect(cronLooksValid('*/5 * * * *')).toBe(true)
        expect(cronLooksValid('0 12 * * 1-5')).toBe(true)
        expect(cronLooksValid('0 12 * *')).toBe(false)

        expect(ordinal(1)).toBe('1st')
        expect(ordinal(2)).toBe('2nd')
        expect(ordinal(3)).toBe('3rd')
        expect(ordinal(4)).toBe('4th')
        expect(ordinal(11)).toBe('11th')
        expect(ordinal(22)).toBe('22nd')
    })
})
