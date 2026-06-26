import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, Snackbar, Stack, TextField, Typography} from '@mui/material';
import {apiPost} from '@dsherwin/react-api-interface';
import {useNavigate} from 'react-router';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {TestActionDialog} from './TestActionDialog';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {ActionSqlVariableDialog} from './ActionSqlVariableDialog';
import {ActionStepCard, type ActionStepFieldErrors} from './ActionStepCard';
import {createActionDraftStep, extractActionTemplateVars, validateActionStepSql} from './actionEditorUtils';
import {type Dialect, type StepDraft, type ValidationIssue} from './types';

export const CreateAction = () => {
    const navigate = useNavigate();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const actionLimit = checkLimit('actions');

    useEffect(() => {
        if (!actionLimit.allowed) {
            navigate('/actions/list');
        }
    }, [actionLimit.allowed, navigate]);

    const {confirmPrompt} = useMuiPrompts();

    const [name, setName] = useState('');
    const [dialect, setDialect] = useState<Dialect>('generic');
    const [description, setDescription] = useState('');
    const [notes, setNotes] = useState('');
    const [softWrap, setSoftWrap] = useState<boolean>(true);
    const [steps, setSteps] = useState<StepDraft[]>([createActionDraftStep(1)]);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});
    const [testOpen, setTestOpen] = useState(false);
    const [stepChecks, setStepChecks] = useState<Record<string, { checking: boolean; issues: ValidationIssue[]; touched: boolean }>>({});
    const [actionErrors, setActionErrors] = useState<{ name?: string }>({});
    const [stepFieldErrors, setStepFieldErrors] = useState<Record<string, ActionStepFieldErrors>>({});
    const [varDialog, setVarDialog] = useState<{ open: boolean; step: StepDraft | null; vars: string[]; values: Record<string, string> }>({open: false, step: null, vars: [], values: {}});

    function applyValidationIssues(issues: ValidationIssue[]) {
        const nextActionErrors: { name?: string } = {};
        const nextStepErrors: Record<string, ActionStepFieldErrors> = {};

        for (const issue of issues) {
            if (!issue.stepId) {
                if (issue.code === 'NAME_REQUIRED') nextActionErrors.name = 'Action name is required.';
                continue;
            }

            const current = nextStepErrors[issue.stepId] || {};
            switch (issue.code) {
                case 'STEP_NAME_REQUIRED':
                    current.name = 'Name is required.';
                    break;
                case 'STEP_SQL_REQUIRED':
                    current.sql = 'SQL is required.';
                    break;
                case 'TIMEOUT_RANGE':
                    current.timeout = 'Timeout must be 1–3600 seconds.';
                    break;
                case 'EXPECT_COLUMN_REQUIRED':
                    current.expectationColumn = 'Column is required.';
                    break;
                case 'EXPECT_VALUE_REQUIRED':
                    current.expectationValue = 'Expected value is required.';
                    break;
                case 'EXPECT_OP_REQUIRED':
                    current.expectationOp = 'Choose an operator.';
                    break;
                case 'EXPECT_VALUE_NUM':
                    current.expectationValueNum = 'Provide a numeric value.';
                    break;
                default:
                    break;
            }
            nextStepErrors[issue.stepId] = current;
        }

        setActionErrors(nextActionErrors);
        setStepFieldErrors(nextStepErrors);
    }

    const onAddStep = () => {
        setSteps((prev) => [...prev, createActionDraftStep(prev.length + 1)]);
    };

    const onRemoveStep = async (id: string) => {
        const ok = await confirmPrompt({
            title: 'Delete Step',
            message: 'Are you sure you want to delete this step? This cannot be undone.',
            buttonText: 'Delete',
        });
        if (!ok) return;
        setSteps((prev) => (prev.length <= 1 ? prev : prev.filter((step) => step.id !== id)));
    };

    const moveStep = (id: string, dir: -1 | 1) => {
        setSteps((prev) => {
            const index = prev.findIndex((step) => step.id === id);
            if (index === -1) return prev;
            const nextIndex = index + dir;
            if (nextIndex < 0 || nextIndex >= prev.length) return prev;
            const next = prev.slice();
            const [item] = next.splice(index, 1);
            next.splice(nextIndex, 0, item);
            return next;
        });
    };

    const updateStep = (id: string, patch: Partial<StepDraft>) => {
        setSteps((prev) => prev.map((step) => (step.id === id ? {...step, ...patch} : step)));
        setStepFieldErrors((prev) => {
            const existing = {...(prev[id] || {})};
            if ('name' in patch) delete existing.name;
            if ('sql' in patch) delete existing.sql;
            if ('timeout' in patch) delete existing.timeout;
            if ('expectation' in patch) {
                delete existing.expectationQuery;
                delete existing.expectationColumn;
                delete existing.expectationValue;
                delete existing.expectationOp;
                delete existing.expectationValueNum;
            }

            const next = {...prev};
            if (Object.keys(existing).length === 0) delete next[id];
            else next[id] = existing;
            return next;
        });
    };

    const validateAll = (): ValidationIssue[] => {
        const issues: ValidationIssue[] = [];
        if (!name.trim()) issues.push({code: 'NAME_REQUIRED', message: 'Action name is required.'});
        if (steps.length === 0) issues.push({code: 'NO_STEPS', message: 'At least one step is required.'});

        steps.forEach((step, index) => {
            if (!step.name.trim()) issues.push({code: 'STEP_NAME_REQUIRED', message: `Step ${index + 1}: name is required.`, stepId: step.id});
            if (!step.sql.trim()) issues.push({code: 'STEP_SQL_REQUIRED', message: `Step ${index + 1}: SQL is required.`, stepId: step.id});
            if (step.timeout) {
                const timeout = Number(step.timeout);
                if (!Number.isFinite(timeout) || timeout < 1 || timeout > 3600) {
                    issues.push({code: 'TIMEOUT_RANGE', message: `Step ${index + 1}: timeout must be between 1 and 3600 seconds.`, stepId: step.id});
                }
            }

            switch (step.expectation.kind) {
                case 'fieldEqualsFirst':
                case 'fieldEqualsLast':
                case 'fieldEquals': {
                    if (!((step.expectation as any).column || '').trim()) {
                        issues.push({code: 'EXPECT_COLUMN_REQUIRED', message: `Step ${index + 1}: column is required.`, stepId: step.id});
                    }
                    const expected = (step.expectation as any).expected;
                    if (expected === undefined || expected === null || `${expected}`.trim() === '') {
                        issues.push({code: 'EXPECT_VALUE_REQUIRED', message: `Step ${index + 1}: expected value is required.`, stepId: step.id});
                    }
                    break;
                }
                case 'rowsAffected':
                    if (!step.expectation.op) {
                        issues.push({code: 'EXPECT_OP_REQUIRED', message: `Step ${index + 1}: expectation "Rows affected" requires an operator.`, stepId: step.id});
                    }
                    if (!step.expectation.value || !/^[0-9]+$/.test(step.expectation.value)) {
                        issues.push({code: 'EXPECT_VALUE_NUM', message: `Step ${index + 1}: expectation "Rows affected" requires a numeric value.`, stepId: step.id});
                    }
                    break;
                default:
                    break;
            }
        });

        return issues;
    };

    const onCheckStep = async (step: StepDraft) => {
        const vars = extractActionTemplateVars(step.sql);
        if (vars.length > 0) {
            const values: Record<string, string> = {};
            for (const variable of vars) values[variable] = '';
            setVarDialog({open: true, step, vars, values});
            return;
        }

        setStepChecks((prev) => ({...prev, [step.id]: {checking: true, issues: prev[step.id]?.issues || [], touched: true}}));
        try {
            const data: any = await apiPost('/actions/validate-step' as any, {dialect, sqlText: step.sql} as any);
            const issues = Array.isArray(data?.issues) ? data.issues as ValidationIssue[] : [];
            setStepChecks((prev) => ({...prev, [step.id]: {checking: false, issues, touched: true}}));
            if (issues.length === 0) {
                setSnack({open: true, message: `Step "${step.name}": no issues found.`, severity: 'success'});
            }
        } catch (error) {
            console.error(error);
            const issues = validateActionStepSql(step.sql);
            setStepChecks((prev) => ({...prev, [step.id]: {checking: false, issues, touched: true}}));
            if (issues.length === 0) {
                setSnack({open: true, message: `Step "${step.name}": no issues found.`, severity: 'success'});
            }
        }
    };

    const runCheckWithVars = async () => {
        if (!varDialog.step) return;
        const step = varDialog.step;

        for (const variable of varDialog.vars) {
            if ((varDialog.values[variable] || '').trim() === '') {
                setSnack({open: true, message: `Provide a test value for "${variable}".`, severity: 'error'});
                return;
            }
        }

        setStepChecks((prev) => ({...prev, [step.id]: {checking: true, issues: prev[step.id]?.issues || [], touched: true}}));
        try {
            const payload: any = {dialect, sqlText: step.sql, variables: varDialog.values} as any;
            const data: any = await apiPost('/actions/validate-step' as any, payload);
            const issues = Array.isArray(data?.issues) ? data.issues as ValidationIssue[] : [];
            setStepChecks((prev) => ({...prev, [step.id]: {checking: false, issues, touched: true}}));
            setVarDialog({open: false, step: null, vars: [], values: {}});
            if (issues.length === 0) setSnack({open: true, message: `Step "${step.name}": no issues found.`, severity: 'success'});
        } catch (error) {
            console.error(error);
            const issues = validateActionStepSql(step.sql);
            setStepChecks((prev) => ({...prev, [step.id]: {checking: false, issues, touched: true}}));
            setVarDialog({open: false, step: null, vars: [], values: {}});
            if (issues.length === 0) setSnack({open: true, message: `Step "${step.name}": no issues found.`, severity: 'success'});
        }
    };

    const onSave = async () => {
        const issues = validateAll();
        if (issues.length > 0) {
            applyValidationIssues(issues);
            setSnack({open: true, message: 'Please fix validation issues.', severity: 'error'});
            return;
        }

        try {
            setActionErrors({});
            setStepFieldErrors({});
            const payload = {
                name,
                dialect,
                description: description || undefined,
                notes: notes || undefined,
                steps: steps.map((step, order) => ({
                    order,
                    name: step.name.trim(),
                    sqlText: step.sql,
                    timeoutSeconds: step.timeout ? Number(step.timeout) : null,
                    expectation: step.expectation,
                    outputCapture: step.outputCapture,
                    onFailure: step.onFailure,
                })),
            };
            const response: any = await apiPost('/actions' as any, payload as any);
            if (response?.ok === false) throw new Error('Save failed');
            setSnack({open: true, message: 'Action saved.', severity: 'success'});
            void reloadFeatureAvailability();
            window.setTimeout(() => navigate('/actions/list'), 500);
        } catch (error) {
            console.error(error);
            setSnack({open: true, message: 'Failed to save Action.', severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" spacing={1} sx={{justifyContent: 'space-between', flexWrap: 'wrap'}}>
                    <Typography variant="h5">Create DB Action</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Button variant="outlined" onClick={() => setTestOpen(true)}>Test Action</Button>
                        <Button variant="contained" onClick={onSave}>Save Action</Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={3}>
                            <Stack direction={{xs: 'column', md: 'row'}} spacing={2}>
                                <TextField
                                    label="Name"
                                    value={name}
                                    onChange={(event) => {
                                        setName(event.target.value);
                                        if (actionErrors.name) setActionErrors({});
                                    }}
                                    required
                                    fullWidth
                                    error={!!actionErrors.name}
                                    helperText={actionErrors.name}
                                />
                            </Stack>

                            <TextField
                                label="Description (optional)"
                                value={description}
                                onChange={(event) => setDescription(event.target.value)}
                                fullWidth
                                multiline
                                minRows={3}
                                placeholder="What this Action does, dependencies, risks, etc."
                            />

                            <TextField
                                label="Notes (optional)"
                                value={notes}
                                onChange={(event) => setNotes(event.target.value)}
                                fullWidth
                                multiline
                                minRows={3}
                            />
                        </VStack>
                    </CardContent>
                </Card>

                <VStack spacing={2}>
                    <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                        <Typography variant="h6">Steps</Typography>
                        <Button variant="outlined" onClick={onAddStep}>Add step</Button>
                    </HStack>
                    {steps.map((step, index) => (
                        <ActionStepCard
                            key={step.id}
                            step={step}
                            index={index}
                            stepsLength={steps.length}
                            softWrap={softWrap}
                            onToggleSoftWrap={setSoftWrap}
                            onMoveStep={moveStep}
                            onRemoveStep={onRemoveStep}
                            updateStep={updateStep}
                            onCheckStep={onCheckStep}
                            dialect={dialect}
                            onDialectChange={setDialect}
                            stepCheck={stepChecks[step.id]}
                            fieldErrors={stepFieldErrors[step.id]}
                        />
                    ))}
                </VStack>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardActions sx={{justifyContent: 'space-between', flexWrap: 'wrap'}}>
                        <Typography variant="caption" sx={{color: 'text.secondary', ml: 1}}>
                            Review each step and save when ready.
                        </Typography>
                        <HStack spacing={1}>
                            <Button variant="outlined" onClick={() => setTestOpen(true)}>Test Action</Button>
                            <Button variant="contained" onClick={onSave}>Save Action</Button>
                        </HStack>
                    </CardActions>
                </Card>

                <TestActionDialog open={testOpen} onClose={() => setTestOpen(false)} type="database" steps={steps} dialect={dialect}/>

                <ActionSqlVariableDialog
                    open={varDialog.open}
                    vars={varDialog.vars}
                    values={varDialog.values}
                    onClose={() => setVarDialog({open: false, step: null, vars: [], values: {}})}
                    onConfirm={runCheckWithVars}
                    onValueChange={(name, value) => setVarDialog((prev) => ({...prev, values: {...prev.values, [name]: value}}))}
                />

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack((current) => ({...current, open: false}))} anchorOrigin={{vertical: 'top', horizontal: 'center'}}>
                    <Alert onClose={() => setSnack((current) => ({...current, open: false}))} severity={snack.severity} variant="filled" sx={{width: '100%'}}>
                        {snack.message}
                    </Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
