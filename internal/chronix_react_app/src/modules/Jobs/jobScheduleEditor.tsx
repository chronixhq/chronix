import {Box, Checkbox, Chip, FormControl, FormControlLabel, FormHelperText, InputLabel, MenuItem, Select, TextField, ToggleButton, ToggleButtonGroup, Typography} from '@mui/material';
import {SettingsAwareDateTimePicker} from '../../lib/SettingsAwareDateTimePicker';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {ordinal} from './editorUtils';
import {type CustomFrequency, type DaySelector, type JobScheduleEditorState, type JobScheduleFieldErrors, type MonthOrdinal, type RepeatPreset} from './jobScheduleState';

const WEEKDAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] as const;
const FULL_WEEKDAY_LABELS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'] as const;

export const JobScheduleEditor = ({
    state,
    onChange,
    errors = {},
}: {
    state: JobScheduleEditorState;
    onChange: (next: JobScheduleEditorState) => void;
    errors?: JobScheduleFieldErrors;
}) => {
    const update = (patch: Partial<JobScheduleEditorState>) => onChange({...state, ...patch});
    const toggleWeekday = (day: number) => update({
        weekdays: state.weekdays.includes(day)
            ? state.weekdays.filter((value) => value !== day)
            : [...state.weekdays, day].sort(),
    });
    const toggleMonthDay = (day: number) => update({
        monthDays: state.monthDays.includes(day)
            ? state.monthDays.filter((value) => value !== day)
            : [...state.monthDays, day],
    });
    const toggleYearMonth = (month: number) => update({
        yearMonths: state.yearMonths.includes(month)
            ? state.yearMonths.filter((value) => value !== month)
            : [...state.yearMonths, month],
    });

    return (
        <VStack spacing={2}>
            <Typography variant="h6">Schedule</Typography>
            <FormControl sx={{minWidth: 220}}>
                <InputLabel id="skind-label">Type</InputLabel>
                <Select
                    labelId="skind-label"
                    label="Type"
                    value={state.schedKind}
                    onChange={(event) => update({schedKind: event.target.value as JobScheduleEditorState['schedKind']})}
                >
                    <MenuItem value="manual">Manual only</MenuItem>
                    <MenuItem value="single">Single-shot</MenuItem>
                    <MenuItem value="recurring">Recurring</MenuItem>
                </Select>
            </FormControl>

            {state.schedKind === 'manual' ? (
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    This job will only run when triggered manually.
                </Typography>
            ) : state.schedKind === 'single' ? (
                <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                    <SettingsAwareDateTimePicker
                        label="Run at"
                        valueIso={state.singleRunAtIso}
                        onChangeIso={(value) => update({singleRunAtIso: value})}
                        errorText={errors.singleRunAt}
                        sx={{minWidth: 260}}
                    />
                </HStack>
            ) : (
                <VStack spacing={2}>
                    <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                        <SettingsAwareDateTimePicker
                            label="Start"
                            valueIso={state.recStartIso}
                            onChangeIso={(value) => update({recStartIso: value})}
                            errorText={errors.recStart}
                            sx={{minWidth: 260}}
                        />
                        <SettingsAwareDateTimePicker
                            label="End (optional)"
                            valueIso={state.recEndIso}
                            onChangeIso={(value) => update({recEndIso: value})}
                            sx={{minWidth: 260}}
                        />
                        <FormControl sx={{minWidth: 220}}>
                            <InputLabel id="rmode-label">Mode</InputLabel>
                            <Select
                                labelId="rmode-label"
                                label="Mode"
                                value={state.recMode}
                                onChange={(event) => update({recMode: event.target.value as JobScheduleEditorState['recMode']})}
                            >
                                <MenuItem value="structured">Structured</MenuItem>
                                <MenuItem value="cron">Cron string</MenuItem>
                            </Select>
                        </FormControl>
                    </HStack>

                    {state.recMode === 'structured' ? (
                        <VStack spacing={2}>
                            <HStack spacing={2} sx={{alignItems: 'flex-end', flexWrap: 'wrap'}}>
                                <FormControl sx={{minWidth: 220, width: 'auto', alignSelf: 'flex-start'}}>
                                    <InputLabel id="repeat-label">Repeat</InputLabel>
                                    <Select
                                        labelId="repeat-label"
                                        label="Repeat"
                                        value={state.repeatPreset}
                                        onChange={(event) => update({repeatPreset: event.target.value as RepeatPreset})}
                                    >
                                        <MenuItem value="day">Every Day</MenuItem>
                                        <MenuItem value="week">Every Week</MenuItem>
                                        <MenuItem value="2weeks">Every 2 Weeks</MenuItem>
                                        <MenuItem value="month">Every Month</MenuItem>
                                        <MenuItem value="year">Every Year</MenuItem>
                                        <MenuItem value="custom">Custom</MenuItem>
                                    </Select>
                                </FormControl>
                                {state.repeatPreset === 'custom' && (
                                    <FormControl sx={{minWidth: 220, width: 'auto', alignSelf: 'flex-start'}}>
                                        <InputLabel id="cfreq-label">Frequency</InputLabel>
                                        <Select
                                            labelId="cfreq-label"
                                            label="Frequency"
                                            value={state.customFreq}
                                            onChange={(event) => update({customFreq: event.target.value as CustomFrequency})}
                                        >
                                            <MenuItem value="minutes">By Minutes</MenuItem>
                                            <MenuItem value="hours">By Hours</MenuItem>
                                            <MenuItem value="daily">Daily</MenuItem>
                                            <MenuItem value="weekly">Weekly</MenuItem>
                                            <MenuItem value="monthly">Monthly</MenuItem>
                                            <MenuItem value="yearly">Yearly</MenuItem>
                                        </Select>
                                    </FormControl>
                                )}
                            </HStack>

                            {state.repeatPreset === 'custom' ? (
                                <VStack spacing={2}>
                                    {state.customFreq === 'minutes' && (
                                        <HStack spacing={1} sx={{alignItems: 'flex-end', flexWrap: 'wrap'}}>
                                            <FormControl sx={{minWidth: 140, width: 'auto', alignSelf: 'flex-start'}}>
                                                <InputLabel id="every-min">Every</InputLabel>
                                                <Select labelId="every-min" label="Every" value={state.everyMinutes} onChange={(event) => update({everyMinutes: Number(event.target.value)})}>
                                                    {Array.from({length: 59}).map((_, index) => (
                                                        <MenuItem key={index + 1} value={index + 1}>{index + 1}</MenuItem>
                                                    ))}
                                                </Select>
                                            </FormControl>
                                            <Typography sx={{pb: 1.2}}>{state.everyMinutes === 1 ? 'minute' : 'minutes'}</Typography>
                                        </HStack>
                                    )}

                                    {state.customFreq === 'hours' && (
                                        <HStack spacing={1} sx={{alignItems: 'flex-end', flexWrap: 'wrap'}}>
                                            <FormControl sx={{minWidth: 140, width: 'auto', alignSelf: 'flex-start'}}>
                                                <InputLabel id="every-hour">Every</InputLabel>
                                                <Select labelId="every-hour" label="Every" value={state.everyHours} onChange={(event) => update({everyHours: Number(event.target.value)})}>
                                                    {Array.from({length: 23}).map((_, index) => (
                                                        <MenuItem key={index + 1} value={index + 1}>{index + 1}</MenuItem>
                                                    ))}
                                                </Select>
                                            </FormControl>
                                            <Typography sx={{pb: 1.2}}>{state.everyHours === 1 ? 'hour' : 'hours'}</Typography>
                                        </HStack>
                                    )}

                                    {state.customFreq === 'daily' && (
                                        <HStack spacing={1} sx={{alignItems: 'flex-end', flexWrap: 'wrap'}}>
                                            <FormControl sx={{minWidth: 140, width: 'auto', alignSelf: 'flex-start'}}>
                                                <InputLabel id="every-day">Every</InputLabel>
                                                <Select labelId="every-day" label="Every" value={state.everyDays} onChange={(event) => update({everyDays: Number(event.target.value)})}>
                                                    {Array.from({length: 31}).map((_, index) => (
                                                        <MenuItem key={index + 1} value={index + 1}>{index + 1}</MenuItem>
                                                    ))}
                                                </Select>
                                            </FormControl>
                                            <Typography sx={{pb: 1.2}}>{state.everyDays === 1 ? 'day' : 'days'}</Typography>
                                        </HStack>
                                    )}

                                    {state.customFreq === 'weekly' && (
                                        <VStack spacing={1}>
                                            <HStack spacing={1} sx={{alignItems: 'flex-end', flexWrap: 'wrap'}}>
                                                <FormControl sx={{minWidth: 140, width: 'auto', alignSelf: 'flex-start'}}>
                                                    <InputLabel id="every-week">Every</InputLabel>
                                                    <Select labelId="every-week" label="Every" value={state.everyWeeks} onChange={(event) => update({everyWeeks: Number(event.target.value)})}>
                                                        {Array.from({length: 52}).map((_, index) => (
                                                            <MenuItem key={index + 1} value={index + 1}>{ordinal(index + 1)}</MenuItem>
                                                        ))}
                                                    </Select>
                                                </FormControl>
                                                <Typography sx={{pb: 1.2}}>week</Typography>
                                            </HStack>
                                            <Typography variant="body2">Days of the Week</Typography>
                                            <HStack spacing={1} sx={{flexWrap: 'wrap'}}>
                                                {WEEKDAY_LABELS.map((label, day) => (
                                                    <Chip key={day} size="small" variant={state.weekdays.includes(day) ? 'filled' : 'outlined'} label={label} onClick={() => toggleWeekday(day)}/>
                                                ))}
                                            </HStack>
                                        </VStack>
                                    )}

                                    {state.customFreq === 'monthly' && (
                                        <VStack spacing={1}>
                                            <HStack spacing={1} sx={{alignItems: 'flex-end', flexWrap: 'wrap'}}>
                                                <FormControl sx={{minWidth: 140, width: 'auto', alignSelf: 'flex-start'}}>
                                                    <InputLabel id="every-month">Every</InputLabel>
                                                    <Select labelId="every-month" label="Every" value={state.everyMonths} onChange={(event) => update({everyMonths: Number(event.target.value)})}>
                                                        {Array.from({length: 11}).map((_, index) => (
                                                            <MenuItem key={index + 1} value={index + 1}>{index + 1}</MenuItem>
                                                        ))}
                                                    </Select>
                                                </FormControl>
                                                <Typography sx={{pb: 1.2}}>{state.everyMonths === 1 ? 'month' : 'months'}</Typography>
                                            </HStack>
                                            <ToggleButtonGroup exclusive value={state.monthMode} onChange={(_, value) => value && update({monthMode: value})} size="small">
                                                <ToggleButton value="each">Each</ToggleButton>
                                                <ToggleButton value="onThe">On the</ToggleButton>
                                            </ToggleButtonGroup>
                                            {state.monthMode === 'each' ? (
                                                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 0.5, maxWidth: 360}}>
                                                    {Array.from({length: 31}).map((_, index) => {
                                                        const day = index + 1;
                                                        const selected = state.monthDays.includes(day);
                                                        return (
                                                            <Chip
                                                                key={day}
                                                                label={day}
                                                                color={selected ? 'primary' : 'default'}
                                                                variant={selected ? 'filled' : 'outlined'}
                                                                onClick={() => toggleMonthDay(day)}
                                                                size="small"
                                                                sx={{justifySelf: 'center'}}
                                                            />
                                                        );
                                                    })}
                                                </Box>
                                            ) : (
                                                <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                                                    <FormControl sx={{minWidth: 160}}>
                                                        <InputLabel id="mord-label"> </InputLabel>
                                                        <Select labelId="mord-label" value={state.monthOrdinal} onChange={(event) => update({monthOrdinal: event.target.value as MonthOrdinal})} displayEmpty>
                                                            <MenuItem value="first">First</MenuItem>
                                                            <MenuItem value="second">Second</MenuItem>
                                                            <MenuItem value="third">Third</MenuItem>
                                                            <MenuItem value="fourth">Fourth</MenuItem>
                                                            <MenuItem value="fifth">Fifth</MenuItem>
                                                            <MenuItem value="next_to_last">Next to last</MenuItem>
                                                            <MenuItem value="last">Last</MenuItem>
                                                        </Select>
                                                    </FormControl>
                                                    <FormControl sx={{minWidth: 180}}>
                                                        <InputLabel id="mwd-label"> </InputLabel>
                                                        <Select labelId="mwd-label" value={state.monthWeekday} onChange={(event) => update({monthWeekday: event.target.value as DaySelector})} displayEmpty>
                                                            {FULL_WEEKDAY_LABELS.map((label, day) => (
                                                                <MenuItem key={day} value={day}>{label}</MenuItem>
                                                            ))}
                                                            <MenuItem value="day">Day</MenuItem>
                                                            <MenuItem value="weekday">Weekday</MenuItem>
                                                            <MenuItem value="weekend">Weekend day</MenuItem>
                                                        </Select>
                                                    </FormControl>
                                                </HStack>
                                            )}
                                        </VStack>
                                    )}

                                    {state.customFreq === 'yearly' && (
                                        <VStack spacing={1}>
                                            <Typography variant="body2">Months</Typography>
                                            <HStack spacing={0.5} sx={{flexWrap: 'wrap'}}>
                                                {Array.from({length: 12}).map((_, index) => {
                                                    const month = index + 1;
                                                    const selected = state.yearMonths.includes(month);
                                                    const label = new Date(2000, month - 1, 1).toLocaleString(undefined, {month: 'short'});
                                                    return <Chip key={month} size="small" label={label} variant={selected ? 'filled' : 'outlined'} color={selected ? 'primary' : 'default'} onClick={() => toggleYearMonth(month)}/>;
                                                })}
                                            </HStack>
                                            <FormControlLabel control={<Checkbox checked={state.yearOnThe} onChange={(event) => update({yearOnThe: event.target.checked})}/>} label="On the"/>
                                            {state.yearOnThe && (
                                                <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                                                    <FormControl sx={{minWidth: 160}}>
                                                        <InputLabel id="yord-label"> </InputLabel>
                                                        <Select labelId="yord-label" value={state.yearOrdinal} onChange={(event) => update({yearOrdinal: event.target.value as MonthOrdinal})} displayEmpty>
                                                            <MenuItem value="first">First</MenuItem>
                                                            <MenuItem value="second">Second</MenuItem>
                                                            <MenuItem value="third">Third</MenuItem>
                                                            <MenuItem value="fourth">Fourth</MenuItem>
                                                            <MenuItem value="fifth">Fifth</MenuItem>
                                                            <MenuItem value="next_to_last">Next to last</MenuItem>
                                                            <MenuItem value="last">Last</MenuItem>
                                                        </Select>
                                                    </FormControl>
                                                    <FormControl sx={{minWidth: 180}}>
                                                        <InputLabel id="ywd-label"> </InputLabel>
                                                        <Select labelId="ywd-label" value={state.yearWeekday} onChange={(event) => update({yearWeekday: event.target.value as DaySelector})} displayEmpty>
                                                            {FULL_WEEKDAY_LABELS.map((label, day) => (
                                                                <MenuItem key={day} value={day}>{label}</MenuItem>
                                                            ))}
                                                            <MenuItem value="day">Day</MenuItem>
                                                            <MenuItem value="weekday">Weekday</MenuItem>
                                                            <MenuItem value="weekend">Weekend day</MenuItem>
                                                        </Select>
                                                    </FormControl>
                                                </HStack>
                                            )}
                                        </VStack>
                                    )}
                                </VStack>
                            ) : null}
                        </VStack>
                    ) : (
                        <VStack spacing={1}>
                            <TextField
                                label="Cron (5 fields)"
                                placeholder="e.g., */30 * * * *"
                                value={state.cronStr}
                                onChange={(event) => update({cronStr: event.target.value})}
                                fullWidth
                                error={!!errors.cronStr}
                                helperText={errors.cronStr || 'Examples: "0 * * * *" (hourly), "*/15 * * * *" (every 15 minutes)'}
                            />
                            {errors.cronStr && <FormHelperText error>{errors.cronStr}</FormHelperText>}
                        </VStack>
                    )}
                </VStack>
            )}
        </VStack>
    );
};
