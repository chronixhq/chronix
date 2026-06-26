import {memo, useState} from 'react';
import {Alert, Button, Card, CardContent, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControl, FormHelperText, IconButton, InputLabel, MenuItem, Select, Switch, TextField, Tooltip, Typography} from '@mui/material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import {Add, Delete} from '@mui/icons-material';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {type Dialect, type ExpectationKind, type ExpectRowsAffected, type FailurePolicy, type StepDraft, type StepExpectation, type ValidationIssue} from './types';

export interface ActionStepFieldErrors {
    name?: string;
    sql?: string;
    timeout?: string;
    expectationQuery?: string;
    expectationColumn?: string;
    expectationValue?: string;
    expectationOp?: string;
    expectationValueNum?: string;
}

const ExpectationEditor = ({
    step,
    onChange,
    fieldErrors,
}: {
    step: StepDraft;
    onChange: (expectation: StepExpectation) => void;
    fieldErrors?: ActionStepFieldErrors;
}) => {
    const kind = step.expectation.kind;

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
                            case 'rowExists':
                                next = {kind: 'rowExists'};
                                break;
                            case 'noRowsReturned':
                                next = {kind: 'noRowsReturned'};
                                break;
                            case 'fieldEqualsFirst':
                                next = {kind: 'fieldEqualsFirst', column: (step.expectation as any).column || '', expected: (step.expectation as any).expected || ''};
                                break;
                            case 'fieldEqualsLast':
                                next = {kind: 'fieldEqualsLast', column: (step.expectation as any).column || '', expected: (step.expectation as any).expected || ''};
                                break;
                            case 'fieldEquals':
                                next = {kind: 'fieldEqualsFirst', column: (step.expectation as any).column || '', expected: (step.expectation as any).expected || ''};
                                break;
                            case 'rowsAffected':
                                next = {
                                    kind: 'rowsAffected',
                                    op: step.expectation.kind === 'rowsAffected' ? (step.expectation as any).op : '>=',
                                    value: step.expectation.kind === 'rowsAffected' ? (step.expectation as any).value : '1',
                                };
                                break;
                            default:
                                next = {kind: 'noError'};
                        }
                        onChange(next);
                    }}
                >
                    <MenuItem value="none">None (don’t check result)</MenuItem>
                    <MenuItem value="noError">Success (no errors)</MenuItem>
                    <MenuItem value="rowExists">Row exists (query returns at least one row)</MenuItem>
                    <MenuItem value="noRowsReturned">No rows returned (query must return zero rows)</MenuItem>
                    <MenuItem value="fieldEqualsFirst">First Row Field Equals</MenuItem>
                    <MenuItem value="fieldEqualsLast">Last Row Field Equals</MenuItem>
                    <MenuItem value="rowsAffected">Rows affected (compare against N)</MenuItem>
                </Select>
                <FormHelperText>
                    Determines pass/fail for this step. If expectation fails, the failure policy below applies.
                </FormHelperText>
            </FormControl>

            {(kind === 'fieldEqualsFirst' || kind === 'fieldEqualsLast' || kind === 'fieldEquals') && (
                <HStack spacing={1} sx={{flexWrap: 'wrap'}}>
                    <TextField
                        label="Column name"
                        placeholder="status"
                        value={(step.expectation as any).column || ''}
                        onChange={(event) => onChange({kind, column: event.target.value, expected: (step.expectation as any).expected})}
                        sx={{minWidth: {xs: '100%', md: 240}}}
                        error={!!fieldErrors?.expectationColumn}
                        helperText={fieldErrors?.expectationColumn}
                    />
                    <TextField
                        label="Expected value"
                        placeholder="OK"
                        value={(step.expectation as any).expected ?? ''}
                        onChange={(event) => onChange({kind, column: (step.expectation as any).column, expected: event.target.value})}
                        sx={{minWidth: {xs: '100%', md: 240}}}
                        error={!!fieldErrors?.expectationValue}
                        helperText={fieldErrors?.expectationValue}
                    />
                </HStack>
            )}

            {kind === 'rowsAffected' && (
                <HStack spacing={1} sx={{flexWrap: 'wrap'}}>
                    <FormControl sx={{minWidth: 160}} error={!!fieldErrors?.expectationOp}>
                        <InputLabel id={`rows-op-${step.id}`}>Operator</InputLabel>
                        <Select
                            labelId={`rows-op-${step.id}`}
                            label="Operator"
                            value={(step.expectation as ExpectRowsAffected).op || '>='}
                            onChange={(event) => onChange({kind: 'rowsAffected', op: event.target.value as ('>=' | '==' | '<='), value: (step.expectation as ExpectRowsAffected).value})}
                        >
                            <MenuItem value=">=">≥ (greater or equal)</MenuItem>
                            <MenuItem value="==">= (equal)</MenuItem>
                            <MenuItem value="<=">≤ (less or equal)</MenuItem>
                        </Select>
                        {fieldErrors?.expectationOp && <FormHelperText>{fieldErrors.expectationOp}</FormHelperText>}
                    </FormControl>
                    <TextField
                        label="Value"
                        placeholder="1"
                        value={(step.expectation as ExpectRowsAffected).value || ''}
                        onChange={(event) => onChange({kind: 'rowsAffected', op: (step.expectation as ExpectRowsAffected).op || '>=', value: event.target.value.replace(/[^0-9]/g, '')})}
                        sx={{minWidth: {xs: '100%', md: 160}}}
                        slotProps={{
                            htmlInput: {
                                inputMode: 'numeric',
                                pattern: '[0-9]*',
                            },
                        }}
                        error={!!fieldErrors?.expectationValueNum}
                        helperText={fieldErrors?.expectationValueNum}
                    />
                </HStack>
            )}
        </VStack>
    );
};

const OutputCaptureEditor = ({step, onChange}: { step: StepDraft; onChange: (capture: Record<string, any>) => void }) => {
    const capture = step.outputCapture || {};
    const entries = Object.entries(capture);

    const onAdd = () => onChange({...capture, '': {source: 'column', name: '', row: 'first'}});
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
                            <MenuItem value="column">Column</MenuItem>
                            <MenuItem value="jsonpath">JSONPath</MenuItem>
                        </Select>
                        <IconButton size="small" onClick={() => onRemove(key)} color="error"><Delete fontSize="small"/></IconButton>
                    </HStack>
                    <HStack spacing={1}>
                        <TextField label="Column Name" size="small" value={value.name || ''} onChange={(event) => onUpdateValue(key, 'name', event.target.value)} fullWidth/>
                        <Select size="small" value={value.row || 'first'} onChange={(event) => onUpdateValue(key, 'row', event.target.value)} sx={{minWidth: 120}}>
                            <MenuItem value="first">First Row</MenuItem>
                            <MenuItem value="last">Last Row</MenuItem>
                        </Select>
                    </HStack>
                    {value.source === 'jsonpath' && <TextField label="JSONPath" size="small" value={value.path || ''} onChange={(event) => onUpdateValue(key, 'path', event.target.value)} fullWidth/>}
                </VStack>
            ))}
            <Button size="small" startIcon={<Add/>} onClick={onAdd} sx={{alignSelf: 'flex-start'}}>Add Variable Capture</Button>
        </VStack>
    );
};

export const ActionStepCard = memo(({
    step,
    index,
    stepsLength,
    softWrap,
    onToggleSoftWrap,
    onMoveStep,
    onRemoveStep,
    updateStep,
    onCheckStep,
    dialect,
    onDialectChange,
    stepCheck,
    fieldErrors,
}: {
    step: StepDraft;
    index: number;
    stepsLength: number;
    softWrap: boolean;
    onToggleSoftWrap: (value: boolean) => void;
    onMoveStep: (id: string, dir: -1 | 1) => void;
    onRemoveStep: (id: string) => void;
    updateStep: (id: string, patch: Partial<StepDraft>) => void;
    onCheckStep: (step: StepDraft) => void;
    dialect: Dialect;
    onDialectChange: (value: Dialect) => void;
    stepCheck?: { checking: boolean; issues: ValidationIssue[]; touched: boolean };
    fieldErrors?: ActionStepFieldErrors;
}) => {
    const canUp = index > 0;
    const canDown = index < stepsLength - 1;
    const [openSqlHelp, setOpenSqlHelp] = useState(false);

    return (
        <Card variant="outlined" sx={{borderRadius: 3}}>
            <CardContent>
                <VStack spacing={2}>
                    <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                        <HStack spacing={1} alignItems="center">
                            <Typography variant="subtitle1" sx={{fontWeight: 600}}>Step {index + 1}</Typography>
                        </HStack>
                        <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                            <Button size="small" disabled={!canUp} onClick={() => onMoveStep(step.id, -1)}>Move up</Button>
                            <Button size="small" disabled={!canDown} onClick={() => onMoveStep(step.id, +1)}>Move down</Button>
                            <Button size="small" color="error" onClick={() => onRemoveStep(step.id)} disabled={stepsLength <= 1}>Delete</Button>
                        </HStack>
                    </HStack>

                    <HStack spacing={2} sx={{flexWrap: 'wrap'}}>
                        <TextField
                            label="Step name"
                            value={step.name}
                            onChange={(event) => updateStep(step.id, {name: event.target.value})}
                            sx={{minWidth: {xs: '100%', md: 360}}}
                            error={!!fieldErrors?.name}
                            helperText={fieldErrors?.name}
                        />
                        <TextField
                            label="Timeout (seconds)"
                            value={step.timeout}
                            onChange={(event) => updateStep(step.id, {timeout: event.target.value.replace(/[^0-9]/g, '')})}
                            placeholder="60"
                            slotProps={{
                                htmlInput: {
                                    inputMode: 'numeric',
                                    pattern: '[0-9]*',
                                    min: 1,
                                    max: 3600,
                                },
                            }}
                            helperText={fieldErrors?.timeout || 'Empty uses default 60s (allowed range 1–3600)'}
                            error={!!fieldErrors?.timeout}
                            sx={{minWidth: {xs: '100%', md: 260}}}
                        />
                    </HStack>

                    <VStack spacing={1}>
                        <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                            <HStack alignItems="center" spacing={0.5}>
                                <Typography variant="subtitle1" sx={{fontWeight: 600}}>SQL</Typography>
                                <Tooltip title="SQL help">
                                    <IconButton size="small" aria-label="SQL help" onClick={() => setOpenSqlHelp(true)}>
                                        <InfoOutlinedIcon fontSize="small"/>
                                    </IconButton>
                                </Tooltip>
                            </HStack>
                            <HStack spacing={1} alignItems="center">
                                <Typography variant="body2" sx={{color: 'text.secondary'}}>Soft wrap</Typography>
                                <Switch size="small" checked={softWrap} onChange={(event) => onToggleSoftWrap(event.target.checked)}/>
                                <Select
                                    size="small"
                                    value={dialect}
                                    onChange={(event) => onDialectChange(event.target.value as Dialect)}
                                    sx={{minWidth: 120, height: 32}}
                                >
                                    <MenuItem value="postgres">Postgres</MenuItem>
                                    <MenuItem value="mysql">MySQL</MenuItem>
                                    <MenuItem value="sqlite">SQLite</MenuItem>
                                    <MenuItem value="tsql">SQL Server</MenuItem>
                                    <MenuItem value="generic">Generic</MenuItem>
                                </Select>
                                <Button size="small" variant="outlined" onClick={() => onCheckStep(step)}>Check SQL</Button>
                            </HStack>
                        </HStack>

                        <TextField
                            value={step.sql}
                            onChange={(event) => updateStep(step.id, {sql: event.target.value})}
                            placeholder={'-- Write your SQL for this step.\nSELECT 1;'}
                            fullWidth
                            multiline
                            minRows={8}
                            error={!!fieldErrors?.sql}
                            helperText={fieldErrors?.sql}
                            sx={{
                                '& .MuiInputBase-input': {
                                    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
                                    whiteSpace: softWrap ? 'pre-wrap' : 'pre',
                                    overflowWrap: softWrap ? 'anywhere' : 'normal',
                                },
                            }}
                        />

                        <Dialog open={openSqlHelp} onClose={() => setOpenSqlHelp(false)} maxWidth="sm" fullWidth>
                            <DialogTitle>SQL editor help</DialogTitle>
                            <DialogContent dividers>
                                <Typography variant="body2" component="p" sx={{mb: 2}}>
                                    Write the SQL to execute for this step. Use template variables to reference values that will be provided when the job/action runs. You assign these variables when you create or define the job (e.g., in Scheduled Jobs); Chronix then binds them at runtime.
                                </Typography>
                                <Typography variant="body2" component="p" sx={{mb: 2}}>
                                    <strong>Variable Substitution:</strong> Reference variables with {'{{name}}'} or {'${name}'}. These are substituted in the SQL text and also in <strong>Result Expectation</strong> values (like "Expected value").
                                </Typography>
                                <Typography variant="body2" component="p">
                                    Tips: “Soft wrap” toggles line wrapping in the editor. “Check SQL” runs a quick lint to catch common mistakes.
                                </Typography>
                            </DialogContent>
                            <DialogActions>
                                <Button onClick={() => setOpenSqlHelp(false)} autoFocus>Close</Button>
                            </DialogActions>
                        </Dialog>

                        {stepCheck?.checking && <Alert severity="info">Checking…</Alert>}
                        {!stepCheck?.checking && stepCheck?.touched && (
                            stepCheck.issues.length > 0 ? (
                                <Alert severity="warning">
                                    {stepCheck.issues.map((issue, issueIndex) => (
                                        <Typography key={issueIndex} variant="body2">[{issue.code}] {issue.message}</Typography>
                                    ))}
                                </Alert>
                            ) : (
                                <Alert severity="success">No issues found.</Alert>
                            )
                        )}
                    </VStack>

                    <ExpectationEditor step={step} onChange={(expectation) => updateStep(step.id, {expectation})} fieldErrors={fieldErrors}/>

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

ActionStepCard.displayName = 'ActionStepCard';
