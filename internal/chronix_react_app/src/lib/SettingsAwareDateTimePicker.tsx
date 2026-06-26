import {useMemo} from 'react';
import {DateTimePicker} from '@mui/x-date-pickers/DateTimePicker';
import {useSettings} from '../data/SettingsContext';
import {dayjs} from './dayjs';
import {useInputSize} from './useInputSize';

export interface SettingsAwareDateTimePickerProps {
    label: string;
    valueIso?: string | null; // ISO UTC string or null
    onChangeIso: (nextIsoUtc: string | null) => void;
    required?: boolean;
    disabled?: boolean;
    errorText?: string;
    minIso?: string;
    maxIso?: string;
    sx?: any;
}

export const SettingsAwareDateTimePicker: React.FC<SettingsAwareDateTimePickerProps> = ({
                                                                                            label,
                                                                                            valueIso,
                                                                                            onChangeIso,
                                                                                            required,
                                                                                            disabled,
                                                                                            errorText,
                                                                                            minIso,
                                                                                            maxIso,
                                                                                            sx,
                                                                                        }) => {
    const {timeZone, timeFormat} = useSettings();
    const ampm = timeFormat === '12h';
    const size = useInputSize();

    // Convert ISO UTC → Dayjs in selected TZ
    const value = useMemo(() => {
        if (!valueIso) return null;
        try {
            return dayjs.utc(valueIso).tz(timeZone);
        } catch {
            return null;
        }
    }, [valueIso, timeZone]);

    const minV = useMemo(() => (minIso ? dayjs.utc(minIso).tz(timeZone) : undefined), [minIso, timeZone]);
    const maxV = useMemo(() => (maxIso ? dayjs.utc(maxIso).tz(timeZone) : undefined), [maxIso, timeZone]);

    return (
        <DateTimePicker
            label={label}
            value={value}
            onChange={(dj) => {
                if (!dj) {
                    onChangeIso(null);
                    return;
                }
                try {
                    // Interpret selected wall time in chosen TZ, then convert to UTC ISO
                    const iso = dj.tz(timeZone, true).utc().toISOString();
                    onChangeIso(iso);
                } catch {
                    onChangeIso(null);
                }
            }}
            ampm={ampm}
            minDateTime={minV}
            maxDateTime={maxV}
            disabled={disabled}
            formatDensity={"spacious"}
            slotProps={{
                textField: {
                    required,
                    error: !!errorText,
                    helperText: errorText,
                    size,
                    sx: [{
                        '& .MuiInputBase-root': {
                            minHeight: {xs: 44, md: undefined},
                        },
                    }, sx],
                },
            }}
        />
    );
};
