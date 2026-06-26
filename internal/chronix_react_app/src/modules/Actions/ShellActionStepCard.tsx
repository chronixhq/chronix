import {memo, useState} from 'react';
import {Alert, Autocomplete, Button, Card, CardContent, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControl, FormHelperText, IconButton, InputLabel, MenuItem, Select, Switch, TextField, Tooltip, Typography} from '@mui/material';
import {Add, CheckCircle, Delete} from '@mui/icons-material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import {apiPost} from '@dsherwin/react-api-interface';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {COMMON_SHELLS} from './shellActionEditorUtils';
import {type ExpectationKind, type FailurePolicy, type ShellStepDraft, type StepExpectation} from './types';

function ExpectationEditor({step, onChange}: { step: ShellStepDraft; onChange: (expectation: StepExpectation) => void }) {
    const kind = step.expectation?.kind || 'noError';

    return (
        <VStack spacing={1}>
            <FormControl fullWidth>
                <InputLabel id={`exp-kind-${step.id}`}>Result expectation</InputLabel>
                <Select
                    labelId={`exp-kind-${step.id}`}
                    label="Result expectation"
                    value={kind}
                    onChange={(event) => {
                        const nextKind = event.target.value as ExpectationKind;
                        let next: StepExpectation;
                        switch (nextKind) {
                            case 'none':
                                next = {kind: 'none'};
                                break;
                            case 'noError':
                                next = {kind: 'noError'};
                                break;
                            case 'exitCodeEquals':
                                next = {kind: 'exitCodeEquals', value: (step.expectation as any)?.value || '0'};
                                break;
                            case 'contains':
                                next = {kind: 'contains', value: (step.expectation as any)?.value || ''};
                                break;
                            case 'notContains':
                                next = {kind: 'notContains', value: (step.expectation as any)?.value || ''};
                                break;
                            case 'firstLineEquals':
                                next = {kind: 'firstLineEquals', value: (step.expectation as any)?.value || ''};
                                break;
                            case 'lastLineEquals':
                                next = {kind: 'lastLineEquals', value: (step.expectation as any)?.value || ''};
                                break;
                            case 'regexMatch':
                                next = {kind: 'regexMatch', value: (step.expectation as any)?.value || ''};
                                break;
                            default:
                                next = {kind: 'noError'};
                        }
                        onChange(next);
                    }}
                >
                    <MenuItem value="none">None (don’t check result)</MenuItem>
                    <MenuItem value="noError">No errors (exit code 0)</MenuItem>
                    <MenuItem value="exitCodeEquals">Specific exit code</MenuItem>
                    <MenuItem value="contains">Output contains string</MenuItem>
                    <MenuItem value="notContains">Output does NOT contain string</MenuItem>
                    <MenuItem value="firstLineEquals">First line of stdout equals</MenuItem>
                    <MenuItem value="lastLineEquals">Last line of stdout equals</MenuItem>
                    <MenuItem value="regexMatch">Output matches regex</MenuItem>
                </Select>
                <FormHelperText>
                    Determines pass/fail for this step. If expectation fails, the failure policy below applies.
                </FormHelperText>
            </FormControl>

            {kind !== 'none' && kind !== 'noError' && kind !== 'regexMatch' && (
                <TextField
                    label={kind === 'exitCodeEquals' ? 'Exit code' : 'Value'}
                    value={(step.expectation as any)?.value || ''}
                    onChange={(event) => onChange({...step.expectation, value: event.target.value} as any)}
                    fullWidth
                    slotProps={kind === 'exitCodeEquals' ? {htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}} : undefined}
                />
            )}

            {kind === 'regexMatch' && (
                <VStack spacing={1}>
                    <HStack spacing={1}>
                        <TextField
                            label="Regex pattern"
                            placeholder="ID: (\\d+)"
                            value={(step.expectation as any).value || ''}
                            onChange={(event) => onChange({...step.expectation, value: event.target.value} as any)}
                            fullWidth
                        />
                        <TextField
                            label="Group"
                            value={(step.expectation as any).group || '0'}
                            onChange={(event) => onChange({...step.expectation, group: event.target.value.replace(/[^0-9]/g, '')} as any)}
                            sx={{width: 80}}
                            slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                        />
                    </HStack>
                    <TextField
                        label="Expected value (optional)"
                        placeholder="1234"
                        value={(step.expectation as any).expected || ''}
                        onChange={(event) => onChange({...step.expectation, expected: event.target.value} as any)}
                        fullWidth
                        helperText="If specified, the captured group must equal this value."
                    />
                </VStack>
            )}
        </VStack>
    );
}

function OutputCaptureEditor({step, onChange}: { step: ShellStepDraft; onChange: (capture: Record<string, any>) => void }) {
    const capture = step.outputCapture || {};
    const entries = Object.entries(capture);

    const onAdd = () => onChange({...capture, '': {source: 'jsonpath', path: ''}});
    const onRemove = (key: string) => {
        const next = {...capture};
        delete next[key];
        onChange(next);
    };
    const onUpdateKey = (oldKey: string, newKey: string) => {
        if (oldKey === newKey) return;
        const next = {...capture};
        const value = next[oldKey];
        delete next[oldKey];
        next[newKey] = value;
        onChange(next);
    };
    const onUpdateValue = (key: string, field: string, value: any) => {
        onChange({...capture, [key]: {...capture[key], [field]: value}});
    };

    return (
        <VStack spacing={1}>
            <Typography variant="subtitle2">Output Capture (Variables)</Typography>
            {entries.length === 0 && <Typography variant="body2" sx={{color: 'text.secondary'}}>No variables captured from output.</Typography>}
            {entries.map(([key, value], index) => (
                <VStack key={index} spacing={1} sx={{p: 1, border: '1px dashed grey', borderRadius: 1}}>
                    <HStack spacing={1} alignItems="center">
                        <TextField label="Variable Name" size="small" value={key} onChange={(event) => onUpdateKey(key, event.target.value)} sx={{flexGrow: 1}}/>
                        <Select size="small" value={value.source} onChange={(event) => onUpdateValue(key, 'source', event.target.value)}>
                            <MenuItem value="jsonpath">JSONPath</MenuItem>
                            <MenuItem value="regex">Regex</MenuItem>
                        </Select>
                        <IconButton size="small" onClick={() => onRemove(key)} color="error"><Delete fontSize="small"/></IconButton>
                    </HStack>
                    {value.source === 'jsonpath' && (
                        <TextField label="JSONPath" size="small" value={value.path || ''} onChange={(event) => onUpdateValue(key, 'path', event.target.value)} fullWidth/>
                    )}
                    {value.source === 'regex' && (
                        <HStack spacing={1}>
                            <TextField label="Regex Pattern" size="small" value={value.pattern || ''} onChange={(event) => onUpdateValue(key, 'pattern', event.target.value)} fullWidth/>
                            <TextField
                                label="Group"
                                size="small"
                                value={value.group || '0'}
                                onChange={(event) => onUpdateValue(key, 'group', event.target.value.replace(/[^0-9]/g, ''))}
                                sx={{width: 80}}
                                slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                            />
                        </HStack>
                    )}
                </VStack>
            ))}
            <Button size="small" startIcon={<Add/>} onClick={onAdd} sx={{alignSelf: 'flex-start'}}>Add Variable Capture</Button>
        </VStack>
    );
}

function EnvEditor({env, onChange}: { env: Record<string, string>; onChange: (env: Record<string, string>) => void }) {
    const entries = Object.entries(env);

    const onAdd = () => onChange({...env, '': ''});
    const onRemove = (key: string) => {
        const next = {...env};
        delete next[key];
        onChange(next);
    };
    const onUpdateKey = (oldKey: string, newKey: string) => {
        if (oldKey === newKey) return;
        const next = {...env};
        const value = next[oldKey];
        delete next[oldKey];
        next[newKey] = value;
        onChange(next);
    };

    return (
        <VStack spacing={1}>
            <Typography variant="subtitle2">Environment Variables</Typography>
            {entries.length === 0 && <Typography variant="body2" sx={{color: 'text.secondary'}}>No environment variables defined.</Typography>}
            {entries.map(([key, value], index) => (
                <HStack key={index} spacing={1}>
                    <TextField label="Key" size="small" value={key} onChange={(event) => onUpdateKey(key, event.target.value)}/>
                    <TextField label="Value" size="small" value={value} onChange={(event) => onChange({...env, [key]: event.target.value})} fullWidth/>
                    <IconButton size="small" onClick={() => onRemove(key)}><Delete fontSize="small"/></IconButton>
                </HStack>
            ))}
            <Button size="small" startIcon={<Add/>} onClick={onAdd} sx={{alignSelf: 'flex-start'}}>Add Environment Variable</Button>
        </VStack>
    );
}

export const ShellActionStepCard = memo(({
    step,
    index,
    stepsLength,
    onRemove,
    onMove,
    updateStep,
}: {
    step: ShellStepDraft;
    index: number;
    stepsLength: number;
    onRemove: (id: string) => void;
    onMove: (id: string, dir: number) => void;
    updateStep: (id: string, patch: Partial<ShellStepDraft>) => void;
}) => {
    const [softWrap, setSoftWrap] = useState(true);
    const [validateMsg, setValidateMsg] = useState<string | null>(null);
    const [openHelp, setOpenHelp] = useState(false);

    async function onValidate() {
        setValidateMsg(null);
        if (step.runMode !== 'script' || !step.scriptText) {
            setValidateMsg('Switch step to Script and enter script to validate');
            return;
        }

        try {
            const response = await apiPost('/shell/actions/validate', {shellPath: step.shellPath, scriptText: step.scriptText}) as any;
            if (response?.ok) setValidateMsg('Script OK');
            else setValidateMsg(response?.message || 'Syntax error');
        } catch {
            setValidateMsg('Validation failed');
        }
    }

    return (
        <Card variant="outlined" sx={{borderRadius: 3}}>
            <CardContent>
                <VStack spacing={2}>
                    <HStack alignItems="center" justifyContent="space-between">
                        <Typography variant="subtitle1" sx={{fontWeight: 600}}>Step {index + 1}</Typography>
                        <HStack spacing={1}>
                            <Button size="small" disabled={index === 0} onClick={() => onMove(step.id, -1)}>Move up</Button>
                            <Button size="small" disabled={index === stepsLength - 1} onClick={() => onMove(step.id, 1)}>Move down</Button>
                            <Button size="small" color="error" disabled={stepsLength <= 1} onClick={() => onRemove(step.id)}>Delete</Button>
                        </HStack>
                    </HStack>

                    <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                        <TextField label="Step name" value={step.name} onChange={(event) => updateStep(step.id, {name: event.target.value})} sx={{flex: 1, minWidth: {xs: '100%', md: 360}}}/>
                        <TextField label="Run Mode" select value={step.runMode} onChange={(event) => updateStep(step.id, {runMode: event.target.value as any})} sx={{minWidth: 160}}>
                            <MenuItem value="command">Command</MenuItem>
                            <MenuItem value="script">Script</MenuItem>
                        </TextField>
                    </HStack>

                    <Dialog open={openHelp} onClose={() => setOpenHelp(false)} maxWidth="sm" fullWidth>
                        <DialogTitle>Shell action help</DialogTitle>
                        <DialogContent dividers>
                            <Typography variant="body2" component="p" sx={{mb: 2}}>
                                Shell actions execute commands or scripts on the target connection. You can use template variables to reference values that will be provided when the job runs.
                            </Typography>
                            <Typography variant="body2" component="p" sx={{mb: 2}}>
                                <strong>Shell Path:</strong> The absolute path to the shell that will execute your command or script (e.g., <code>/bin/bash</code> or <code>/bin/sh</code> on Linux, <code>/bin/zsh</code> on macOS,
                                or <code>C:\Windows\System32\cmd.exe</code> or <code>C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe</code> on Windows).
                            </Typography>
                            <Typography variant="body2" component="p" sx={{mb: 2}}>
                                <strong>Variable Substitution:</strong> Reference variables with {'{{name}}'} or {'${name}'}. These are substituted in the command, script text, and expectation values.
                            </Typography>
                            <Typography variant="body2" component="p" sx={{mb: 2}}>
                                <strong>Environment Variables:</strong> You can also define environment variables for each step. These are exported before your command or script runs.
                            </Typography>
                            <Typography variant="body2" component="p">
                                Tips: “Soft wrap” toggles line wrapping in the script editor. “Validate” runs a syntax check (non-execution).
                            </Typography>
                        </DialogContent>
                        <DialogActions>
                            <Button onClick={() => setOpenHelp(false)} autoFocus>Close</Button>
                        </DialogActions>
                    </Dialog>

                    {step.runMode === 'command' ? (
                        <VStack spacing={1}>
                            <HStack alignItems="center" spacing={0.5}>
                                <Typography variant="subtitle2">Command</Typography>
                                <Tooltip title="Variable help">
                                    <IconButton size="small" aria-label="Variable help" onClick={() => setOpenHelp(true)}>
                                        <InfoOutlinedIcon fontSize="small"/>
                                    </IconButton>
                                </Tooltip>
                            </HStack>
                            <TextField label="Command" value={step.command} onChange={(event) => updateStep(step.id, {command: event.target.value})} fullWidth placeholder="echo hello"/>
                        </VStack>
                    ) : (
                        <VStack spacing={1}>
                            <HStack alignItems="center" justifyContent="space-between">
                                <HStack alignItems="center" spacing={0.5}>
                                    <Typography variant="subtitle2">Script</Typography>
                                    <Tooltip title="Variable help">
                                        <IconButton size="small" aria-label="Variable help" onClick={() => setOpenHelp(true)}>
                                            <InfoOutlinedIcon fontSize="small"/>
                                        </IconButton>
                                    </Tooltip>
                                </HStack>
                                <HStack spacing={1} alignItems="center">
                                    <Typography variant="body2" sx={{color: 'text.secondary'}}>Soft wrap</Typography>
                                    <Switch size="small" checked={softWrap} onChange={(event) => setSoftWrap(event.target.checked)}/>
                                    <Button size="small" variant="outlined" startIcon={<CheckCircle/>} onClick={onValidate}>Validate</Button>
                                </HStack>
                            </HStack>
                            <TextField
                                value={step.scriptText}
                                onChange={(event) => updateStep(step.id, {scriptText: event.target.value})}
                                fullWidth
                                multiline
                                minRows={6}
                                maxRows={12}
                                sx={{
                                    '& .MuiInputBase-input': {
                                        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
                                        whiteSpace: softWrap ? 'pre-wrap' : 'pre',
                                        overflowWrap: softWrap ? 'anywhere' : 'normal',
                                    },
                                }}
                            />
                            {validateMsg && <Alert severity={validateMsg === 'Script OK' ? 'success' : 'error'} onClose={() => setValidateMsg(null)}>{validateMsg}</Alert>}
                        </VStack>
                    )}

                    <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                        <Autocomplete
                            freeSolo
                            options={COMMON_SHELLS}
                            groupBy={(option) => (typeof option === 'string' ? 'Custom' : option.os)}
                            getOptionLabel={(option) => (typeof option === 'string' ? option : option.label)}
                            value={step.shellPath}
                            onChange={(_, newValue) => {
                                if (typeof newValue === 'string') {
                                    updateStep(step.id, {shellPath: newValue});
                                } else if (newValue) {
                                    updateStep(step.id, {shellPath: newValue.label});
                                } else {
                                    updateStep(step.id, {shellPath: ''});
                                }
                            }}
                            onInputChange={(_, newInputValue) => updateStep(step.id, {shellPath: newInputValue})}
                            renderInput={(params) => <TextField {...params} label="Shell Path"/>}
                            sx={{flex: 1, minWidth: 200}}
                        />
                        <TextField label="Working Dir" value={step.workingDir} onChange={(event) => updateStep(step.id, {workingDir: event.target.value})} sx={{flex: 1, minWidth: 200}}/>
                        <TextField
                            label="Timeout (sec)"
                            value={step.timeout}
                            onChange={(event) => updateStep(step.id, {timeout: event.target.value.replace(/[^0-9]/g, '')})}
                            sx={{width: 120}}
                            slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                        />
                    </HStack>

                    <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                        <TextField
                            label="Max Output Bytes"
                            value={step.outputCaptureMaxBytes}
                            onChange={(event) => updateStep(step.id, {outputCaptureMaxBytes: event.target.value.replace(/[^0-9]/g, '')})}
                            onBlur={(event) => {
                                let value = parseInt(event.target.value);
                                if (Number.isNaN(value) || value < 1) value = 1;
                                if (value > 1048576) value = 1048576;
                                updateStep(step.id, {outputCaptureMaxBytes: String(value)});
                            }}
                            sx={{flex: 1, minWidth: 160}}
                            slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*', min: 1, max: 1048576}}}
                            helperText="Max 1,048,576 bytes (1MB)"
                        />
                        <TextField label="Truncation" select value={step.outputTruncation} onChange={(event) => updateStep(step.id, {outputTruncation: event.target.value as any})} sx={{flex: 1, minWidth: 160}}>
                            <MenuItem value="head">head</MenuItem>
                            <MenuItem value="tail">tail</MenuItem>
                        </TextField>
                    </HStack>

                    <EnvEditor env={step.env} onChange={(env) => updateStep(step.id, {env})}/>

                    <Divider/>

                    <ExpectationEditor step={step} onChange={(expectation) => updateStep(step.id, {expectation})}/>

                    <Divider/>

                    <OutputCaptureEditor step={step} onChange={(capture) => updateStep(step.id, {outputCapture: capture})}/>

                    <Divider/>

                    <FormControl sx={{minWidth: 260}}>
                        <InputLabel id={`fail-pol-${step.id}`}>On expectation failure</InputLabel>
                        <Select
                            labelId={`fail-pol-${step.id}`}
                            label="On expectation failure"
                            value={step.onFailure || 'exit'}
                            onChange={(event) => updateStep(step.id, {onFailure: event.target.value as FailurePolicy})}
                        >
                            <MenuItem value="exit">Exit action with failure (default)</MenuItem>
                            <MenuItem value="continue">Continue to next step</MenuItem>
                        </Select>
                    </FormControl>
                </VStack>
            </CardContent>
        </Card>
    );
});

ShellActionStepCard.displayName = 'ShellActionStepCard';
