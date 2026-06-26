import {memo} from 'react';
import {Add, Delete} from '@mui/icons-material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import {Box, Button, Card, CardContent, Divider, FormControl, IconButton, InputLabel, MenuItem, Select, TextField, Tooltip, Typography} from '@mui/material';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {type ExpectationKind, type FailurePolicy, type StepExpectation, type WebTaskMethod, type WebtaskStepDraft} from './types';

function ExpectationEditor({
    step,
    onChange,
    onOpenHelp,
}: {
    step: WebtaskStepDraft;
    onChange: (expectation: StepExpectation) => void;
    onOpenHelp: () => void;
}) {
    const kind = step.expectation.kind;

    return (
        <VStack spacing={1}>
            <HStack alignItems="center" spacing={0.5}>
                <Typography variant="subtitle2">Result Expectation</Typography>
                <Tooltip title="Help">
                    <IconButton size="small" onClick={onOpenHelp}>
                        <InfoOutlinedIcon fontSize="small"/>
                    </IconButton>
                </Tooltip>
            </HStack>
            <FormControl fullWidth>
                <InputLabel id={`exp-kind-${step.id}`}>Expectation Kind</InputLabel>
                <Select
                    labelId={`exp-kind-${step.id}`}
                    label="Expectation Kind"
                    value={kind}
                    onChange={(event) => {
                        const nextKind = event.target.value as ExpectationKind;
                        let next: StepExpectation;
                        switch (nextKind) {
                            case 'none':
                                next = {kind: 'none'};
                                break;
                            case 'statusCode':
                                next = {kind: 'statusCode', op: '==', value: '200'};
                                break;
                            case 'bodyContains':
                                next = {kind: 'bodyContains', value: ''};
                                break;
                            case 'jsonPath':
                                next = {kind: 'jsonPath', path: '', value: ''};
                                break;
                            case 'latency':
                                next = {kind: 'latency', value: '1000'};
                                break;
                            case 'bodyRegex':
                                next = {kind: 'bodyRegex', value: '', group: '0', expected: ''};
                                break;
                            default:
                                next = {kind: 'none'};
                        }
                        onChange(next);
                    }}
                >
                    <MenuItem value="none">None (don’t check result)</MenuItem>
                    <MenuItem value="statusCode">HTTP Status Code</MenuItem>
                    <MenuItem value="bodyContains">Response body contains</MenuItem>
                    <MenuItem value="bodyRegex">Response body matches regex</MenuItem>
                    <MenuItem value="jsonPath">JSONPath match</MenuItem>
                    <MenuItem value="latency">Latency (ms) less than</MenuItem>
                </Select>
            </FormControl>

            {kind === 'statusCode' && (
                <HStack spacing={1}>
                    <Select
                        size="small"
                        value={(step.expectation as any).op || '=='}
                        onChange={(event) => onChange({...step.expectation, op: event.target.value} as any)}
                        sx={{minWidth: 80}}
                    >
                        <MenuItem value="==">==</MenuItem>
                        <MenuItem value="!=">!=</MenuItem>
                        <MenuItem value=">">&gt;</MenuItem>
                        <MenuItem value="<">&lt;</MenuItem>
                        <MenuItem value=">=">&gt;=</MenuItem>
                        <MenuItem value="<=">&lt;=</MenuItem>
                    </Select>
                    <TextField
                        label="Code"
                        size="small"
                        value={(step.expectation as any).value || '200'}
                        onChange={(event) => onChange({...step.expectation, value: event.target.value} as any)}
                        slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                    />
                </HStack>
            )}

            {kind === 'bodyContains' && (
                <TextField
                    label="Expected string"
                    size="small"
                    value={(step.expectation as any).value || ''}
                    onChange={(event) => onChange({...step.expectation, value: event.target.value} as any)}
                    fullWidth
                />
            )}

            {kind === 'bodyRegex' && (
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

            {kind === 'jsonPath' && (
                <VStack spacing={1}>
                    <TextField
                        label="JSONPath"
                        size="small"
                        placeholder="$.data.status"
                        value={(step.expectation as any).path || ''}
                        onChange={(event) => onChange({...step.expectation, path: event.target.value} as any)}
                        fullWidth
                    />
                    <TextField
                        label="Expected value"
                        size="small"
                        value={(step.expectation as any).value || ''}
                        onChange={(event) => onChange({...step.expectation, value: event.target.value} as any)}
                        fullWidth
                    />
                </VStack>
            )}

            {kind === 'latency' && (
                <TextField
                    label="Max latency (ms)"
                    size="small"
                    value={(step.expectation as any).value || ''}
                    onChange={(event) => onChange({...step.expectation, value: event.target.value} as any)}
                    fullWidth
                    slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                />
            )}
        </VStack>
    );
}

function HeadersEditor({
    headers,
    onChange,
    onOpenHelp,
}: {
    headers: Record<string, string>;
    onChange: (headers: Record<string, string>) => void;
    onOpenHelp: () => void;
}) {
    const entries = Object.entries(headers || {});

    const onAdd = () => onChange({...headers, '': ''});
    const onRemove = (key: string) => {
        const next = {...headers};
        delete next[key];
        onChange(next);
    };
    const onUpdateKey = (currentKey: string, nextKey: string) => {
        if (currentKey === nextKey) return;
        const next = {...headers};
        const value = next[currentKey];
        delete next[currentKey];
        next[nextKey] = value;
        onChange(next);
    };
    const onUpdateValue = (key: string, value: string) => onChange({...headers, [key]: value});

    return (
        <VStack spacing={1}>
            <HStack alignItems="center" spacing={0.5}>
                <Typography variant="subtitle2">Request Headers</Typography>
                <Tooltip title="Help">
                    <IconButton size="small" onClick={onOpenHelp}>
                        <InfoOutlinedIcon fontSize="small"/>
                    </IconButton>
                </Tooltip>
            </HStack>
            {entries.length === 0 && (
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    No custom headers.
                </Typography>
            )}
            {entries.map(([key, value], index) => (
                <HStack key={index} spacing={1}>
                    <TextField label="Name" size="small" value={key} onChange={(event) => onUpdateKey(key, event.target.value)}/>
                    <TextField label="Value" size="small" value={value} onChange={(event) => onUpdateValue(key, event.target.value)} fullWidth/>
                    <IconButton size="small" onClick={() => onRemove(key)} color="error">
                        <Delete fontSize="small"/>
                    </IconButton>
                </HStack>
            ))}
            <Button size="small" startIcon={<Add/>} onClick={onAdd} sx={{alignSelf: 'flex-start'}}>
                Add Header
            </Button>
        </VStack>
    );
}

function ResponseCaptureEditor({
    capture,
    onChange,
    onOpenHelp,
}: {
    capture: Record<string, any>;
    onChange: (capture: Record<string, any>) => void;
    onOpenHelp: () => void;
}) {
    const entries = Object.entries(capture || {});

    const onAdd = () => onChange({...capture, '': {source: 'jsonpath', path: ''}});
    const onRemove = (key: string) => {
        const next = {...capture};
        delete next[key];
        onChange(next);
    };
    const onUpdateKey = (currentKey: string, nextKey: string) => {
        if (currentKey === nextKey) return;
        const next = {...capture};
        const value = next[currentKey];
        delete next[currentKey];
        next[nextKey] = value;
        onChange(next);
    };
    const onUpdateValue = (key: string, field: string, value: any) => {
        onChange({...capture, [key]: {...capture[key], [field]: value}});
    };

    return (
        <VStack spacing={1}>
            <HStack alignItems="center" spacing={0.5}>
                <Typography variant="subtitle2">Response Capture (Variables)</Typography>
                <Tooltip title="Help">
                    <IconButton size="small" onClick={onOpenHelp}>
                        <InfoOutlinedIcon fontSize="small"/>
                    </IconButton>
                </Tooltip>
            </HStack>
            {entries.length === 0 && (
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    No variables captured from response.
                </Typography>
            )}
            {entries.map(([key, value], index) => (
                <VStack key={index} spacing={1} sx={{p: 1, border: '1px dashed grey', borderRadius: 1}}>
                    <HStack spacing={1} alignItems="center">
                        <TextField label="Variable Name" size="small" value={key} onChange={(event) => onUpdateKey(key, event.target.value)}/>
                        <Select size="small" value={value.source} onChange={(event) => onUpdateValue(key, 'source', event.target.value)}>
                            <MenuItem value="jsonpath">JSONPath</MenuItem>
                            <MenuItem value="header">Header</MenuItem>
                            <MenuItem value="regex">Regex</MenuItem>
                        </Select>
                        <IconButton size="small" onClick={() => onRemove(key)} color="error">
                            <Delete fontSize="small"/>
                        </IconButton>
                    </HStack>
                    {value.source === 'jsonpath' && (
                        <TextField label="JSONPath" size="small" value={value.path || ''} onChange={(event) => onUpdateValue(key, 'path', event.target.value)} fullWidth/>
                    )}
                    {value.source === 'header' && (
                        <TextField label="Header Name" size="small" value={value.name || ''} onChange={(event) => onUpdateValue(key, 'name', event.target.value)} fullWidth/>
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
            <Button size="small" startIcon={<Add/>} onClick={onAdd} sx={{alignSelf: 'flex-start'}}>
                Add Variable Capture
            </Button>
        </VStack>
    );
}

export const WebtaskActionStepCard = memo(({
    step,
    index,
    stepsLength,
    updateStep,
    onRemove,
    onMove,
    onOpenHelp,
}: {
    step: WebtaskStepDraft;
    index: number;
    stepsLength: number;
    updateStep: (id: string, patch: Partial<WebtaskStepDraft>) => void;
    onRemove: (id: string) => void;
    onMove: (id: string, direction: number) => void;
    onOpenHelp: () => void;
}) => (
    <Card variant="outlined" sx={{borderRadius: 3}}>
        <CardContent>
            <VStack spacing={2}>
                <HStack justifyContent="space-between" alignItems="center">
                    <HStack spacing={1} alignItems="center">
                        <Typography variant="h6">Step {index + 1}</Typography>
                        <TextField
                            label="Step Name"
                            size="small"
                            value={step.name}
                            onChange={(event) => updateStep(step.id, {name: event.target.value})}
                            sx={{minWidth: 300}}
                        />
                    </HStack>
                    <HStack spacing={0.5}>
                        <Button size="small" disabled={index === 0} onClick={() => onMove(step.id, -1)}>Move Up</Button>
                        <Button size="small" disabled={index === stepsLength - 1} onClick={() => onMove(step.id, 1)}>Move Down</Button>
                        <IconButton size="small" onClick={() => onRemove(step.id)} color="error">
                            <Delete/>
                        </IconButton>
                    </HStack>
                </HStack>

                <Divider/>

                <VStack spacing={1}>
                    <HStack alignItems="center" spacing={0.5}>
                        <Typography variant="subtitle2">URL</Typography>
                        <Tooltip title="Variable help">
                            <IconButton size="small" onClick={onOpenHelp}>
                                <InfoOutlinedIcon fontSize="small"/>
                            </IconButton>
                        </Tooltip>
                    </HStack>
                    <HStack spacing={2}>
                        <FormControl sx={{minWidth: 120}}>
                            <InputLabel>Method</InputLabel>
                            <Select
                                value={step.method}
                                label="Method"
                                onChange={(event) => updateStep(step.id, {method: event.target.value as WebTaskMethod})}
                            >
                                <MenuItem value="GET">GET</MenuItem>
                                <MenuItem value="POST">POST</MenuItem>
                                <MenuItem value="PUT">PUT</MenuItem>
                                <MenuItem value="DELETE">DELETE</MenuItem>
                                <MenuItem value="PATCH">PATCH</MenuItem>
                            </Select>
                        </FormControl>
                        <TextField
                            label="URL or relative path"
                            placeholder="https://api.example.com/endpoint"
                            value={step.url}
                            onChange={(event) => updateStep(step.id, {url: event.target.value})}
                            fullWidth
                            required
                        />
                    </HStack>
                </VStack>

                <HeadersEditor headers={step.headers} onChange={(headers) => updateStep(step.id, {headers})} onOpenHelp={onOpenHelp}/>

                <VStack spacing={1}>
                    <HStack alignItems="center" spacing={0.5}>
                        <Typography variant="subtitle2">Request Body</Typography>
                        <Tooltip title="Variable help">
                            <IconButton size="small" onClick={onOpenHelp}>
                                <InfoOutlinedIcon fontSize="small"/>
                            </IconButton>
                        </Tooltip>
                    </HStack>
                    <TextField
                        label="Request Body"
                        multiline
                        rows={3}
                        value={step.body}
                        onChange={(event) => updateStep(step.id, {body: event.target.value})}
                        fullWidth
                        placeholder="JSON or text payload"
                    />
                </VStack>

                <Divider/>

                <VStack spacing={2}>
                    <Box sx={{flex: 1}}>
                        <ExpectationEditor
                            step={step}
                            onChange={(expectation) => updateStep(step.id, {expectation})}
                            onOpenHelp={onOpenHelp}
                        />
                    </Box>
                    <HStack spacing={3} alignItems="flex-start">
                        <Box sx={{flex: 1}}>
                            <FormControl fullWidth>
                                <InputLabel>On Failure</InputLabel>
                                <Select
                                    value={step.onFailure || 'exit'}
                                    label="On Failure"
                                    onChange={(event) => updateStep(step.id, {onFailure: event.target.value as FailurePolicy})}
                                >
                                    <MenuItem value="exit">Exit Action (Mark Failed)</MenuItem>
                                    <MenuItem value="continue">Continue to Next Step (Warning)</MenuItem>
                                </Select>
                            </FormControl>
                        </Box>
                        <TextField
                            label="Timeout (sec)"
                            value={step.timeout}
                            onChange={(event) => updateStep(step.id, {timeout: event.target.value})}
                            sx={{width: 120}}
                            slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                        />
                    </HStack>
                </VStack>

                <Divider/>

                <ResponseCaptureEditor
                    capture={step.responseCapture}
                    onChange={(responseCapture) => updateStep(step.id, {responseCapture})}
                    onOpenHelp={onOpenHelp}
                />
            </VStack>
        </CardContent>
    </Card>
));

WebtaskActionStepCard.displayName = 'WebtaskActionStepCard';
