import dayjsOrig from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';

// Initialize plugins once for the app
if (!(dayjsOrig as any).__chronix_dayjs_inited) {
  dayjsOrig.extend(utc);
  dayjsOrig.extend(timezone);
  (dayjsOrig as any).__chronix_dayjs_inited = true;
}

export const dayjs = dayjsOrig;

export function setDefaultTimeZone(tz?: string) {
  try {
    if (tz && tz.length > 0) {
      dayjs.tz.setDefault(tz);
    }
  } catch {/* ignore */}
}
