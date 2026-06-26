import {useMemo} from 'react';
import {useSettings} from '../data/SettingsContext';

export interface DateTimeOptions {
  timeZone?: string; // IANA TZ
  hour12?: boolean;
}

export function formatDateTimeTZ(value: string | number | Date | undefined | null, opts?: DateTimeOptions): string {
  if (!value) return '';
  const d = value instanceof Date ? value : new Date(value);
  try {
    return new Intl.DateTimeFormat(undefined, {
      year: 'numeric', month: 'short', day: '2-digit',
      hour: '2-digit', minute: '2-digit', second: '2-digit',
      hour12: opts?.hour12, timeZone: opts?.timeZone,
    }).format(d);
  } catch {
    return d.toLocaleString();
  }
}

// Like formatDateTimeTZ but without seconds (HH:mm only)
export function formatDateTimeTZHM(value: string | number | Date | undefined | null, opts?: DateTimeOptions): string {
  if (!value) return '';
  const d = value instanceof Date ? value : new Date(value);
  try {
    return new Intl.DateTimeFormat(undefined, {
      year: 'numeric', month: 'short', day: '2-digit',
      hour: '2-digit', minute: '2-digit',
      hour12: opts?.hour12, timeZone: opts?.timeZone,
    }).format(d);
  } catch {
    try {
      // Fallback without seconds via locale options if available
      return d.toLocaleString(undefined, { hour: '2-digit', minute: '2-digit' });
    } catch {
      return d.toString();
    }
  }
}

export function formatDateTZ(value: string | number | Date | undefined | null, opts?: DateTimeOptions): string {
  if (!value) return '';
  const d = value instanceof Date ? value : new Date(value);
  try {
    return new Intl.DateTimeFormat(undefined, {
      year: 'numeric', month: 'short', day: '2-digit',
      timeZone: opts?.timeZone,
    }).format(d);
  } catch {
    return d.toLocaleDateString();
  }
}

export function formatTimeTZ(value: string | number | Date | undefined | null, opts?: DateTimeOptions): string {
  if (!value) return '';
  const d = value instanceof Date ? value : new Date(value);
  try {
    return new Intl.DateTimeFormat(undefined, {
      hour: '2-digit', minute: '2-digit', second: '2-digit',
      hour12: opts?.hour12, timeZone: opts?.timeZone,
    }).format(d);
  } catch {
    return d.toLocaleTimeString();
  }
}

// Module-level display options cache that can be used outside React
let GLOBAL_DISPLAY_OPTIONS: { timeZone: string; hour12: boolean } | null = null;

function detectTZ(): string {
  try { return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'; } catch { return 'UTC'; }
}
function detectHour12(): boolean {
  try {
    const p = new Intl.DateTimeFormat(undefined, {hour: 'numeric'}).formatToParts?.(new Date());
    return Array.isArray(p) && p.some(part => part.type === 'dayPeriod');
  } catch { return false; }
}

export function setGlobalDisplayOptions(opts?: { timeZone?: string; hour12?: boolean }) {
  GLOBAL_DISPLAY_OPTIONS = {
    timeZone: opts?.timeZone && opts.timeZone.length > 0 ? opts.timeZone : detectTZ(),
    hour12: typeof opts?.hour12 === 'boolean' ? opts.hour12 : detectHour12(),
  };
}

// Safe to call anywhere; returns last known options or sensible defaults
export function getCurrentDisplayOptions(): { timeZone: string; hour12: boolean } {
  if (!GLOBAL_DISPLAY_OPTIONS) {
    GLOBAL_DISPLAY_OPTIONS = { timeZone: detectTZ(), hour12: detectHour12() };
  }
  return GLOBAL_DISPLAY_OPTIONS;
}

export function useDateTimeFormatters() {
  const {timeZone, timeFormat} = useSettings();
  const hour12 = timeFormat === '12h';
  return useMemo(() => ({
    formatDateTime: (v: string | number | Date | undefined | null) => formatDateTimeTZ(v, {timeZone, hour12}),
    formatDate: (v: string | number | Date | undefined | null) => formatDateTZ(v, {timeZone}),
    formatTime: (v: string | number | Date | undefined | null) => formatTimeTZ(v, {timeZone, hour12}),
    options: {timeZone, hour12},
  }), [timeZone, hour12]);
}
