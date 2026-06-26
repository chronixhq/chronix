import {type JobSchedule, type RecurringCronSchedule, type RecurringStructuredSchedule, type SingleShotSchedule} from './types';
import {getCurrentDisplayOptions} from '../../lib/datetime';

function ordinal(n: number): string {
  const s = ['th','st','nd','rd'];
  const v = n % 100;
  return `${n}${s[(v - 20) % 10] || s[v] || s[0]}`;
}

function formatDate(d: Date): string {
  const { timeZone } = getCurrentDisplayOptions();
  try {
    return new Intl.DateTimeFormat(undefined, { month: 'long', day: 'numeric', year: 'numeric', timeZone }).format(d);
  } catch {
    const month = d.toLocaleString(undefined, { month: 'long' });
    const day = ordinal(d.getDate());
    const year = d.getFullYear();
    return `${month} ${day}, ${year}`;
  }
}

function formatTimeFromHHMM(hhmm?: string): string | null {
  if (!hhmm) return null;
  const [hhStr, mmStr] = hhmm.split(':');
  if (hhStr == null || mmStr == null) return null;
  const hh = Number(hhStr);
  const mm = Number(mmStr);
  const { hour12 } = getCurrentDisplayOptions();
  if (hour12) {
    let h12 = hh % 12;
    if (h12 === 0) h12 = 12;
    const ampm = hh < 12 ? 'am' : 'pm';
    return `${h12}:${String(mm).padStart(2, '0')}${ampm}`;
  }
  return `${String(hh).padStart(2, '0')}:${String(mm).padStart(2, '0')}`;
}

function formatTimeFromDate(d?: Date): string | null {
  if (!d) return null;
  const { timeZone, hour12 } = getCurrentDisplayOptions();
  try {
    return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', timeZone, hour12 }).format(d);
  } catch {
    const hh = d.getHours();
    const mm = d.getMinutes();
    if (hour12) {
      let h12 = hh % 12;
      if (h12 === 0) h12 = 12;
      const ampm = hh < 12 ? 'am' : 'pm';
      return `${h12}:${String(mm).padStart(2, '0')}${ampm}`;
    }
    return `${String(hh).padStart(2, '0')}:${String(mm).padStart(2, '0')}`;
  }
}

function listJoin(items: string[]): string {
  if (items.length === 0) return '';
  if (items.length === 1) return items[0];
  if (items.length === 2) return `${items[0]} and ${items[1]}`;
  return `${items.slice(0, -1).join(', ')} and ${items[items.length - 1]}`;
}

function weekdayShort(d: number): string {
  return ['Sun','Mon','Tue','Wed','Thu','Fri','Sat'][d] ?? '';
}

function monthShort(m: number): string {
  return new Date(2000, m - 1, 1).toLocaleString(undefined, { month: 'short' });
}

export function formatScheduleSummary(schedule: JobSchedule | null | undefined): string | null {
  if (!schedule) return null;
  if (schedule.kind === 'manual') {
    return 'Manual run only.';
  }
  if (schedule.kind === 'single') {
    const ss = schedule as SingleShotSchedule;
    try {
      const d = new Date(ss.runAt);
      return `Job will run on ${formatDate(d)} at ${formatTimeFromDate(d)}.`;
    } catch {
      return null;
    }
  }
  // recurring
  if (schedule.kind !== 'recurring') return null;

  if (schedule.mode === 'cron') {
    const rec = schedule as RecurringCronSchedule;
    const start = rec.startAt ? new Date(rec.startAt) : undefined;
    const end = rec.endAt ? new Date(rec.endAt) : undefined;
    const startStr = start ? `Starting on ${formatDate(start)}` : 'Starting';
    const endStr = end ? ` and ending on ${formatDate(end)}` : '';
    return `${startStr}${endStr}, job will run according to cron "${rec.cron}".`;
  }
  const rec = schedule as RecurringStructuredSchedule;
  const start = rec.startAt ? new Date(rec.startAt) : undefined;
  const end = rec.endAt ? new Date(rec.endAt) : undefined;
  const startStr = start ? `Starting on ${formatDate(start)}` : 'Starting';
  const endStr = end ? ` and ending on ${formatDate(end)}` : '';
  const rule = rec.rule as any;

  if (!rule) return null;

  let detail = '';
  if (rule.freq === 'minute') {
    const n = Number(rule.interval) || 1;
    detail = `every ${n} ${n === 1 ? 'minute' : 'minutes'}`;
  } else if (rule.freq === 'hour') {
    const n = Number(rule.interval) || 1;
    const mm = typeof rule.minuteMark === 'number' ? rule.minuteMark : 0;
    const mmPad = String(mm).padStart(2, '0');
    detail = `every ${n} ${n === 1 ? 'hour' : 'hours'} at :${mmPad} past the hour`;
  } else if (rule.freq === 'day') {
    const n = Number(rule.interval) || 1;
    const t = formatTimeFromHHMM(rule.time) || '';
    detail = `every ${n} ${n === 1 ? 'day' : 'days'}${t ? ` at ${t}` : ''}`;
  } else if (rule.freq === 'week') {
    const n = Number(rule.interval) || 1;
    const wds = (Array.isArray(rule.weekdays) ? rule.weekdays : []).map((d: number) => weekdayShort(d));
    const t = formatTimeFromHHMM(rule.time) || '';
    const everyPart = n === 1 ? 'each week' : `every ${n} weeks`;
    const onPart = wds.length ? ` on ${listJoin(wds)}` : '';
    detail = `${everyPart}${onPart}${t ? ` at ${t}` : ''}`;
  } else if (rule.freq === 'month') {
    const n = Number(rule.interval) || 1;
    const t = formatTimeFromHHMM(rule.time) || '';
    if (rule.mode === 'date') {
      const days: number[] = Array.isArray(rule.days) ? rule.days : (typeof rule.day === 'number' ? [rule.day] : []);
      const dayLabels = days.sort((a,b)=>a-b).map(d => ordinal(d));
      detail = `every ${n} ${n === 1 ? 'month' : 'months'}${dayLabels.length ? ` on the ${listJoin(dayLabels)}` : ''} of the month${t ? ` at ${t}` : ''}`;
    } else {
      // ordinal form
      const ordMap: Record<string,string> = { first:'First', second:'Second', third:'Third', fourth:'Fourth', fifth:'Fifth', last:'Last', next_to_last:'Next to last' };
      const ord = ordMap[String(rule.ordinal)] || 'First';
      let wd: string;
      if (rule.weekday === 'day') wd = 'day';
      else if (rule.weekday === 'weekday') wd = 'weekday';
      else if (rule.weekday === 'weekend') wd = 'weekend day';
      else wd = ['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'][Number(rule.weekday) || 0];
      detail = `every ${n} ${n === 1 ? 'month' : 'months'} on the ${ord} ${wd.toLowerCase()} of the month${t ? ` at ${t}` : ''}`;
    }
  } else if (rule.freq === 'year') {
    const t = formatTimeFromHHMM(rule.time) || '';
    const months: number[] = Array.isArray(rule.months) ? rule.months : (typeof rule.month === 'number' ? [rule.month] : []);
    const mPart = months.length ? ` in ${listJoin(months.sort((a,b)=>a-b).map(m => monthShort(m)))}` : '';
    if (rule.mode === 'date') {
      const day = typeof rule.day === 'number' ? rule.day : undefined;
      const dPart = day ? ` on the ${ordinal(day)}` : '';
      detail = `every year${mPart}${dPart}${t ? ` at ${t}` : ''}`;
    } else {
      const ordMap: Record<string,string> = { first:'First', second:'Second', third:'Third', fourth:'Fourth', fifth:'Fifth', last:'Last', next_to_last:'Next to last' };
      const ord = ordMap[String(rule.ordinal)] || 'First';
      let wd: string;
      if (rule.weekday === 'day') wd = 'day';
      else if (rule.weekday === 'weekday') wd = 'weekday';
      else if (rule.weekday === 'weekend') wd = 'weekend day';
      else wd = ['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'][Number(rule.weekday) || 0];
      detail = `every year${mPart} on the ${ord} ${wd.toLowerCase()}${t ? ` at ${t}` : ''}`;
    }
  }

  if (!detail) return null;
  return `${startStr}${endStr}, job will run ${detail}.`;
}
