// Shared types for Scheduled Jobs module

import type {Action} from '../Actions/types';
import type {DbConnection} from '../Connections/types.ts';

export type JobId = string;

export type JobStatus = 'idle' | 'running' | 'success' | 'error';

// Schedule models
export type ScheduleKind = 'single' | 'recurring' | 'manual';

export interface SingleShotSchedule {
    kind: 'single';
    runAt: string; // ISO timestamp (minute granularity)
}

export interface ManualSchedule {
    kind: 'manual';
}

export type RecurringMode = 'structured' | 'cron';

export interface RecurringBase {
    kind: 'recurring';
    startAt: string; // ISO timestamp (required)
    endAt?: string;  // ISO timestamp (optional)
    mode: RecurringMode;
}

// Structured recurring details (no cross-cutting/advanced options)
export type Frequency = 'minute' | 'hour' | 'day' | 'week' | 'month' | 'year';

export interface MinuteEvery {
    freq: 'minute';
    interval: number; // every N minutes
}

export interface HourEvery {
    freq: 'hour';
    interval: number; // every N hours
    minuteMark: number; // 0..59 at which minute past the hour
}

export interface DayEvery {
    freq: 'day';
    interval: number; // every N days
    time: string; // HH:mm (local)
}

export interface WeekEvery {
    freq: 'week';
    interval: number; // every N weeks
    weekdays: number[]; // 0=Sun .. 6=Sat
    time: string; // HH:mm
}

export interface MonthByDate {
    freq: 'month';
    interval: number; // every N months
    mode: 'date';
    days: number[]; // 1..31 (one or more)
    time: string; // HH:mm
}

export interface MonthByOrdinal {
    freq: 'month';
    interval: number;
    mode: 'ordinal';
    ordinal: 'first' | 'second' | 'third' | 'fourth' | 'fifth' | 'next_to_last' | 'last';
    // 0..6 (Sun..Sat) or grouped selectors: 'day' (any day), 'weekday', 'weekend'
    weekday: number | 'day' | 'weekday' | 'weekend';
    time: string; // HH:mm
}

export type MonthEvery = MonthByDate | MonthByOrdinal;

export interface YearByDate {
    freq: 'year';
    interval: number; // every N years
    mode: 'date';
    months: number[]; // 1..12 (one or more months)
    day: number; // 1..31
    time: string; // HH:mm
}

export interface YearByOrdinal {
    freq: 'year';
    interval: number;
    mode: 'ordinal';
    months: number[]; // 1..12 (one or more months)
    ordinal: 'first' | 'second' | 'third' | 'fourth' | 'fifth' | 'next_to_last' | 'last';
    weekday: number | 'day' | 'weekday' | 'weekend'; // support grouped selectors too
    time: string; // HH:mm
}

export type YearEvery = YearByDate | YearByOrdinal;

export type StructuredRecurring = MinuteEvery | HourEvery | DayEvery | WeekEvery | MonthEvery | YearEvery;

export interface RecurringStructuredSchedule extends RecurringBase {
    mode: 'structured';
    rule: StructuredRecurring;
}

export interface RecurringCronSchedule extends RecurringBase {
    mode: 'cron';
    cron: string; // standard 5-field cron string
}

export type JobSchedule = SingleShotSchedule | RecurringStructuredSchedule | RecurringCronSchedule | ManualSchedule;

export interface JobVar {
    name: string;
    value: string;
}

export interface Job {
    id: JobId;
    name: string;
    description?: string;
    notes?: string;
    targetKind?: 'database' | 'shell' | 'webtask';
    schedule: JobSchedule;
    connectionId: string; // FK to DbConnection.id
    shellConnectionId?: string;
    webtaskConnectionId?: string;
    actionId: string; // FK to Action.id
    enabled?: boolean;
    suspended?: boolean;
    variables?: JobVar[]; // values for {{name}} placeholders
    alertEmails?: string;
    alertPhones?: string;
    notifyOnSuccess?: boolean;
    notifyOnError?: boolean;
    notifyIncludeOutput?: boolean;
    // Derived/metadata
    lastRunStatus?: JobStatus;
    lastRunAt?: string; // ISO timestamp
    nextRunAt?: string; // ISO timestamp (computed by server)
    createdAt?: string;
    updatedAt?: string;
    // Optional denormalized display data
    connection?: Pick<DbConnection, 'id' | 'name' | 'driver'>;
    action?: Pick<Action, 'id' | 'name' | 'dialect'>;
}
