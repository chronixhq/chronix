import {dayjs} from '../../lib/dayjs';
import {cronLooksValid} from './editorUtils';
import {formatScheduleSummary} from './scheduleSummary';
import {type JobSchedule, type RecurringCronSchedule, type RecurringStructuredSchedule, type SingleShotSchedule} from './types';

export type RepeatPreset = 'day' | 'week' | '2weeks' | 'month' | 'year' | 'custom';
export type CustomFrequency = 'minutes' | 'hours' | 'daily' | 'weekly' | 'monthly' | 'yearly';
export type MonthOrdinal = 'first' | 'second' | 'third' | 'fourth' | 'fifth' | 'next_to_last' | 'last';
export type DaySelector = number | 'day' | 'weekday' | 'weekend';

export interface JobScheduleEditorState {
    schedKind: 'single' | 'recurring' | 'manual';
    singleRunAtIso: string | null;
    recMode: 'structured' | 'cron';
    recStartIso: string | null;
    recEndIso: string | null;
    repeatPreset: RepeatPreset;
    customFreq: CustomFrequency;
    everyMinutes: number;
    everyHours: number;
    everyDays: number;
    everyWeeks: number;
    everyMonths: number;
    weekdays: number[];
    monthMode: 'each' | 'onThe';
    monthDays: number[];
    monthOrdinal: MonthOrdinal;
    monthWeekday: DaySelector;
    yearMonths: number[];
    yearOnThe: boolean;
    yearOrdinal: MonthOrdinal;
    yearWeekday: DaySelector;
    cronStr: string;
}

export interface JobScheduleFieldErrors {
    singleRunAt?: string;
    recStart?: string;
    cronStr?: string;
}

const DEFAULT_SINGLE_ISO = () => dayjs().toISOString();

function getScheduleStartParts(recStartIso: string | null, timeZone: string) {
    if (!recStartIso) {
        return {hhmm: '09:00', minute: 0, weekday: 1, day: 1, month: 1};
    }

    const start = dayjs.utc(recStartIso).tz(timeZone);
    return {
        hhmm: `${String(start.hour()).padStart(2, '0')}:${String(start.minute()).padStart(2, '0')}`,
        minute: start.minute(),
        weekday: start.day(),
        day: start.date(),
        month: start.month() + 1,
    };
}

export function createDefaultJobScheduleEditorState(nowIso = DEFAULT_SINGLE_ISO()): JobScheduleEditorState {
    return {
        schedKind: 'single',
        singleRunAtIso: nowIso,
        recMode: 'structured',
        recStartIso: nowIso,
        recEndIso: null,
        repeatPreset: 'day',
        customFreq: 'daily',
        everyMinutes: 5,
        everyHours: 1,
        everyDays: 1,
        everyWeeks: 1,
        everyMonths: 1,
        weekdays: [1],
        monthMode: 'each',
        monthDays: [1],
        monthOrdinal: 'first',
        monthWeekday: 1,
        yearMonths: [],
        yearOnThe: false,
        yearOrdinal: 'first',
        yearWeekday: 1,
        cronStr: '0 * * * *',
    };
}

export function jobScheduleEditorStateFromSchedule(
    schedule: JobSchedule | undefined,
    timeZone: string,
    nowIso = DEFAULT_SINGLE_ISO(),
): JobScheduleEditorState {
    const next = createDefaultJobScheduleEditorState(nowIso);
    if (!schedule) {
        return next;
    }

    if (schedule.kind === 'manual') {
        next.schedKind = 'manual';
        return next;
    }

    if (schedule.kind === 'single') {
        next.schedKind = 'single';
        next.singleRunAtIso = schedule.runAt || nowIso;
        return next;
    }

    next.schedKind = 'recurring';
    next.recStartIso = schedule.startAt || nowIso;
    next.recEndIso = schedule.endAt ?? null;

    if (schedule.mode === 'cron') {
        next.recMode = 'cron';
        next.cronStr = schedule.cron || next.cronStr;
        return next;
    }

    next.recMode = 'structured';
    const startParts = getScheduleStartParts(next.recStartIso, timeZone);
    const rule = schedule.rule as RecurringStructuredSchedule['rule'];
    next.repeatPreset = 'custom';

    if (rule.freq === 'minute') {
        next.customFreq = 'minutes';
        next.everyMinutes = Number(rule.interval) || 1;
        return next;
    }

    if (rule.freq === 'hour') {
        next.customFreq = 'hours';
        next.everyHours = Number(rule.interval) || 1;
        return next;
    }

    if (rule.freq === 'day') {
        if (Number(rule.interval) === 1) {
            next.repeatPreset = 'day';
        } else {
            next.customFreq = 'daily';
            next.everyDays = Number(rule.interval) || 1;
        }
        return next;
    }

    if (rule.freq === 'week') {
        const weekdays = Array.isArray(rule.weekdays) && rule.weekdays.length > 0 ? rule.weekdays : [startParts.weekday];
        if ((Number(rule.interval) === 1 || Number(rule.interval) === 2) && weekdays.length === 1 && weekdays[0] === startParts.weekday) {
            next.repeatPreset = Number(rule.interval) === 2 ? '2weeks' : 'week';
        } else {
            next.customFreq = 'weekly';
            next.everyWeeks = Number(rule.interval) || 1;
            next.weekdays = weekdays;
        }
        return next;
    }

    if (rule.freq === 'month') {
        next.customFreq = 'monthly';
        next.everyMonths = Number(rule.interval) || 1;

        if (rule.mode === 'date') {
            const days = Array.isArray(rule.days) && rule.days.length > 0 ? rule.days : [startParts.day];
            if (Number(rule.interval) === 1 && days.length === 1 && days[0] === startParts.day) {
                next.repeatPreset = 'month';
            } else {
                next.monthMode = 'each';
                next.monthDays = days;
            }
            return next;
        }

        next.monthMode = 'onThe';
        next.monthOrdinal = rule.ordinal ?? 'first';
        next.monthWeekday = rule.weekday ?? 1;
        return next;
    }

    const months = Array.isArray(rule.months) && rule.months.length > 0 ? rule.months : [startParts.month];
    next.customFreq = 'yearly';
    next.yearMonths = months;

    if (rule.mode === 'date') {
        if (Number(rule.interval) === 1 && months.length === 1 && months[0] === startParts.month && Number(rule.day) === startParts.day) {
            next.repeatPreset = 'year';
        } else {
            next.yearOnThe = false;
        }
        return next;
    }

    next.yearOnThe = true;
    next.yearOrdinal = rule.ordinal ?? 'first';
    next.yearWeekday = rule.weekday ?? 1;
    return next;
}

export function validateJobScheduleEditorState(state: JobScheduleEditorState): JobScheduleFieldErrors {
    if (state.schedKind === 'single' && !state.singleRunAtIso) {
        return {singleRunAt: 'Choose a date/time.'};
    }

    if (state.schedKind !== 'recurring') {
        return {};
    }

    if (!state.recStartIso) {
        return {recStart: 'Start time is required.'};
    }

    if (state.recMode === 'cron' && !cronLooksValid(state.cronStr)) {
        return {cronStr: 'Enter a valid 5-field cron string.'};
    }

    return {};
}

export function buildJobScheduleFromEditorState(state: JobScheduleEditorState, timeZone: string): JobSchedule | null {
    if (state.schedKind === 'manual') {
        return {kind: 'manual'};
    }

    if (state.schedKind === 'single') {
        if (!state.singleRunAtIso) {
            return null;
        }
        return {kind: 'single', runAt: state.singleRunAtIso} as SingleShotSchedule;
    }

    if (!state.recStartIso) {
        return null;
    }

    if (state.recMode === 'cron') {
        if (!cronLooksValid(state.cronStr)) {
            return null;
        }
        return {
            kind: 'recurring',
            mode: 'cron',
            startAt: state.recStartIso,
            endAt: state.recEndIso ?? undefined,
            cron: state.cronStr.trim(),
        } as RecurringCronSchedule;
    }

    const startParts = getScheduleStartParts(state.recStartIso, timeZone);
    let rule: RecurringStructuredSchedule['rule'] | null = null;

    if (state.repeatPreset !== 'custom') {
        if (state.repeatPreset === 'day') {
            rule = {freq: 'day', interval: 1, time: startParts.hhmm};
        } else if (state.repeatPreset === 'week') {
            rule = {freq: 'week', interval: 1, weekdays: [startParts.weekday], time: startParts.hhmm};
        } else if (state.repeatPreset === '2weeks') {
            rule = {freq: 'week', interval: 2, weekdays: [startParts.weekday], time: startParts.hhmm};
        } else if (state.repeatPreset === 'month') {
            rule = {freq: 'month', interval: 1, mode: 'date', days: [startParts.day], time: startParts.hhmm};
        } else if (state.repeatPreset === 'year') {
            rule = {freq: 'year', interval: 1, mode: 'date', months: [startParts.month], day: startParts.day, time: startParts.hhmm};
        }
    } else if (state.customFreq === 'minutes') {
        rule = {freq: 'minute', interval: Math.max(1, Math.min(59, Number(state.everyMinutes) || 1))};
    } else if (state.customFreq === 'hours') {
        rule = {freq: 'hour', interval: Math.max(1, Math.min(23, Number(state.everyHours) || 1)), minuteMark: startParts.minute};
    } else if (state.customFreq === 'daily') {
        rule = {freq: 'day', interval: Math.max(1, Math.min(31, Number(state.everyDays) || 1)), time: startParts.hhmm};
    } else if (state.customFreq === 'weekly') {
        const weekdays = state.weekdays.length > 0 ? state.weekdays.slice().sort() : [startParts.weekday];
        rule = {freq: 'week', interval: Math.max(1, Math.min(52, Number(state.everyWeeks) || 1)), weekdays, time: startParts.hhmm};
    } else if (state.customFreq === 'monthly') {
        if (state.monthMode === 'each') {
            const monthDays = state.monthDays.length > 0 ? state.monthDays.slice().sort((a, b) => a - b) : [startParts.day];
            rule = {freq: 'month', interval: Math.max(1, Math.min(11, Number(state.everyMonths) || 1)), mode: 'date', days: monthDays, time: startParts.hhmm};
        } else {
            rule = {
                freq: 'month',
                interval: Math.max(1, Math.min(11, Number(state.everyMonths) || 1)),
                mode: 'ordinal',
                ordinal: state.monthOrdinal,
                weekday: state.monthWeekday,
                time: startParts.hhmm,
            };
        }
    } else if (state.customFreq === 'yearly') {
        const months = state.yearMonths.length > 0 ? state.yearMonths.slice().sort((a, b) => a - b) : [startParts.month];
        if (state.yearOnThe) {
            rule = {
                freq: 'year',
                interval: 1,
                mode: 'ordinal',
                months,
                ordinal: state.yearOrdinal,
                weekday: state.yearWeekday,
                time: startParts.hhmm,
            };
        } else {
            rule = {
                freq: 'year',
                interval: 1,
                mode: 'date',
                months,
                day: startParts.day,
                time: startParts.hhmm,
            };
        }
    }

    if (!rule) {
        return null;
    }

    return {
        kind: 'recurring',
        mode: 'structured',
        startAt: state.recStartIso,
        endAt: state.recEndIso ?? undefined,
        rule,
    } as RecurringStructuredSchedule;
}

export function summarizeJobScheduleEditorState(state: JobScheduleEditorState, timeZone: string): string | null {
    try {
        return formatScheduleSummary(buildJobScheduleFromEditorState(state, timeZone));
    } catch {
        return null;
    }
}
