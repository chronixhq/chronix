import React, {createContext, useCallback, useContext, useEffect, useMemo, useState} from 'react';
import {setDefaultTimeZone} from '../lib/dayjs';
import {useAuthContext} from './useAuthContext';
import {setGlobalDisplayOptions} from '../lib/datetime';

export type TimeFormat = '12h' | '24h';

interface SettingsContextValue {
    timeFormat: TimeFormat;
    timeZone: string; // IANA TZ name
    setTimeFormat: (fmt: TimeFormat) => void;
    setTimeZone: (tz: string) => void;
}

const SettingsContext = createContext<SettingsContextValue | undefined>(undefined);

function detectDefaultTimeFormat(): TimeFormat {
    try {
        const f = new Intl.DateTimeFormat(undefined, {hour: 'numeric'});
        const p = f.formatToParts?.(new Date());
        const hasDayPeriod = Array.isArray(p) && p.some(part => part.type === 'dayPeriod');
        return hasDayPeriod ? '12h' : '24h';
    } catch {
        return '24h';
    }
}

function detectDefaultTimeZone(): string {
    try {
        const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
        return tz || 'UTC';
    } catch {
        return 'UTC';
    }
}

export function SettingsContextProvider({children}: { children: React.ReactNode }) {
    const {user} = useAuthContext();

    // Initialize from AuthContext.user if present, else sensible browser defaults
    const [timeFormat, setTimeFormatState] = useState<TimeFormat>(() => {
        const tf = user?.timeFormat;
        return tf === '12h' || tf === '24h' ? tf : detectDefaultTimeFormat();
    });
    const [timeZone, setTimeZoneState] = useState<string>(() => {
        const tz = user?.timeZone;
        return tz && tz.length > 0 ? tz : detectDefaultTimeZone();
    });

    // When AuthContext.user changes (login/logout or profile refresh), sync our state
    useEffect(() => {
        const tf = user?.timeFormat;
        const tz = user?.timeZone;
        const nextTF: TimeFormat = tf === '12h' || tf === '24h' ? tf : detectDefaultTimeFormat();
        const nextTZ: string = tz && tz.length > 0 ? tz : detectDefaultTimeZone();
        setTimeFormatState(nextTF);
        setTimeZoneState(nextTZ);
    }, [user]);

    // Keep dayjs and global display options in sync
    useEffect(() => {
        try { setDefaultTimeZone(timeZone); } catch {}
        try { setGlobalDisplayOptions({ timeZone, hour12: timeFormat === '12h' }); } catch {}
    }, [timeZone, timeFormat]);

    const setTimeFormat = useCallback((fmt: TimeFormat) => {
        setTimeFormatState(fmt);
    }, []);

    const setTimeZone = useCallback((tz: string) => {
        setTimeZoneState(tz || 'UTC');
    }, []);

    const value = useMemo<SettingsContextValue>(() => ({
        timeFormat,
        timeZone,
        setTimeFormat,
        setTimeZone,
    }), [timeFormat, timeZone, setTimeFormat, setTimeZone]);

    return <SettingsContext.Provider value={value}>{children}</SettingsContext.Provider>;
}

export function useSettings(): SettingsContextValue {
    const ctx = useContext(SettingsContext);
    if (!ctx) throw new Error('useSettings must be used within SettingsContextProvider');
    return ctx;
}
