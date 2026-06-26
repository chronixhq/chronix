import {describe, expect, it} from 'vitest';
import {buildJobScheduleFromEditorState, createDefaultJobScheduleEditorState, jobScheduleEditorStateFromSchedule, validateJobScheduleEditorState} from './jobScheduleState';

describe('jobScheduleEditor', () => {
    it('builds a yearly ordinal recurring schedule from custom editor state', () => {
        const state = {
            ...createDefaultJobScheduleEditorState('2026-04-12T15:30:00.000Z'),
            schedKind: 'recurring' as const,
            recStartIso: '2026-04-12T15:30:00.000Z',
            repeatPreset: 'custom' as const,
            customFreq: 'yearly' as const,
            yearMonths: [4, 10],
            yearOnThe: true,
            yearOrdinal: 'last' as const,
            yearWeekday: 'weekday' as const,
        };

        expect(buildJobScheduleFromEditorState(state, 'America/Boise')).toEqual({
            kind: 'recurring',
            mode: 'structured',
            startAt: '2026-04-12T15:30:00.000Z',
            endAt: undefined,
            rule: {
                freq: 'year',
                interval: 1,
                mode: 'ordinal',
                months: [4, 10],
                ordinal: 'last',
                weekday: 'weekday',
                time: '09:30',
            },
        });
    });

    it('hydrates preset schedules into editor state', () => {
        const state = jobScheduleEditorStateFromSchedule({
            kind: 'recurring',
            mode: 'structured',
            startAt: '2026-04-12T15:30:00.000Z',
            rule: {
                freq: 'week',
                interval: 2,
                weekdays: [0],
                time: '09:30',
            },
        }, 'America/Boise');

        expect(state.repeatPreset).toBe('2weeks');
        expect(state.recMode).toBe('structured');
        expect(state.weekdays).toEqual([1]);
    });

    it('validates missing single-shot and cron inputs', () => {
        expect(validateJobScheduleEditorState({
            ...createDefaultJobScheduleEditorState(),
            schedKind: 'single',
            singleRunAtIso: null,
        })).toEqual({singleRunAt: 'Choose a date/time.'});

        expect(validateJobScheduleEditorState({
            ...createDefaultJobScheduleEditorState(),
            schedKind: 'recurring',
            recMode: 'cron',
            recStartIso: '2026-04-12T15:30:00.000Z',
            cronStr: 'bad cron',
        })).toEqual({cronStr: 'Enter a valid 5-field cron string.'});
    });
});
