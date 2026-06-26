import {useEffect, useMemo, useRef, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, Snackbar, TextField, Typography} from '@mui/material';
import {useLocation, useNavigate, useParams} from 'react-router';
import {apiGet, apiPost, apiPut} from '@dsherwin/react-api-interface';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {confirmOnNavigate, useUnsavedChanges} from '../../lib/useUnsavedChanges.ts';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {TestActionDialog} from './TestActionDialog';
import {ActionSqlVariableDialog} from './ActionSqlVariableDialog';
import {ActionStepCard} from './ActionStepCard';
import {createActionDraftStep, extractActionTemplateVars, toDraftSteps, validateActionStepSql} from './actionEditorUtils';
import {type Action, type Dialect, type StepDraft, type ValidationIssue} from './types';

function useQuery() {
    const {search} = useLocation();
    return useMemo(() => new URLSearchParams(search), [search]);
}

export const EditAction = () => {
    const navigate = useNavigate();
    const {reload: reloadFeatureAvailability} = useFeatureAvailability();
    const query = useQuery();

    const [loading, setLoading] = useState<boolean>(true);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({open: false, message: '', severity: 'info'});

    const [id, setId] = useState<string>('');
    const [name, setName] = useState<string>('');
    const [dialect, setDialect] = useState<Dialect>('generic');
    const [description, setDescription] = useState<string>('');
    const [notes, setNotes] = useState<string>('');
    const [steps, setSteps] = useState<StepDraft[]>([]);
    const [softWrap, setSoftWrap] = useState<boolean>(true);
    const [testOpen, setTestOpen] = useState(false);
    const [dirty, setDirty] = useState<boolean>(false);
    const baselineRef = useRef<string>('');
    const [stepChecks, setStepChecks] = useState<Record<string, { checking: boolean; issues: ValidationIssue[]; touched: boolean }>>({});
    const [varDialog, setVarDialog] = useState<{ open: boolean; step: StepDraft | null; vars: string[]; values: Record<string, string> }>({open: false, step: null, vars: [], values: {}});

    const params = useParams();
    useEffect(() => {
        const selectedId = (params as any)?.id || query.get('id') || '';

        const load = async () => {
            setLoading(true);
            setLoadError(null);
            try {
                let action: Action | undefined;
                if (selectedId) {
                    action = await apiGet(`/actions/${encodeURIComponent(selectedId)}`) as any;
                } else {
                    const data = await apiGet('/actions') as any;
                    action = Array.isArray(data) ? data[0] : data;
                }
                if (!action) throw new Error('Action not found');
                if ((action as any).suspended) {
                    navigate('/actions/list');
                    return;
                }

                const nextSteps = toDraftSteps(action.steps);
                setId(action.id);
                setName(action.name);
                setDialect(action.dialect as Dialect);
                setDescription(action.description || '');
                setNotes(action.notes || '');
                setSteps(nextSteps);
                baselineRef.current = JSON.stringify({
                    name: action.name,
                    dialect: action.dialect as Dialect,
                    description: action.description || '',
                    notes: action.notes || '',
                    steps: nextSteps,
                });
                setDirty(false);
            } catch (error) {
                console.error(error);
                setLoadError('Failed to load action');
            } finally {
                setLoading(false);
            }
        };

        void load();
    }, [navigate, params, query]);

    useEffect(() => {
        if (baselineRef.current === '') return;
        const current = JSON.stringify({name, dialect, description, notes, steps});
        setDirty(current !== baselineRef.current);
    }, [name, dialect, description, notes, steps]);

    useUnsavedChanges(dirty);

    const {confirmPrompt} = useMuiPrompts();

    const onAddStep = () => {
        setSteps((prev) => [...prev, createActionDraftStep(prev.length + 1, {timeout: '60'})]);
    };

    const onRemoveStep = async (stepId: string) => {
        const ok = await confirmPrompt({
            title: 'Delete Step',
            message: 'Are you sure you want to delete this step? This cannot be undone.',
            buttonText: 'Delete',
        });
        if (!ok) return;
        setSteps((prev) => (prev.length <= 1 ? prev : prev.filter((step) => step.id !== stepId)));
    };

    const moveStep = (stepId: string, dir: -1 | 1) => {
        setSteps((prev) => {
            const index = prev.findIndex((step) => step.id === stepId);
            if (index === -1) return prev;
            const nextIndex = index + dir;
            if (nextIndex < 0 || nextIndex >= prev.length) return prev;
            const next = prev.slice();
            const [item] = next.splice(index, 1);
            next.splice(nextIndex, 0, item);
            return next;
        });
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
        } catch (error) {
            console.error(error);
            const issues = validateActionStepSql(step.sql);
            setStepChecks((prev) => ({...prev, [step.id]: {checking: false, issues, touched: true}}));
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
        } catch (error) {
            console.error(error);
            const issues = validateActionStepSql(step.sql);
            setStepChecks((prev) => ({...prev, [step.id]: {checking: false, issues, touched: true}}));
            setVarDialog({open: false, step: null, vars: [], values: {}});
        }
    };

    const onSave = async () => {
        if (!name.trim()) {
            setSnack({open: true, message: 'Name is required.', severity: 'error'});
            return;
        }
        try {
            const payload = {
                name,
                dialect,
                description: description || undefined,
                notes: notes || undefined,
                steps: steps.map((step, order) => ({
                    order,
                    name: step.name.trim(),
                    sqlText: step.sql,
                    timeoutSeconds: Number(step.timeout || '60'),
                    expectation: step.expectation,
                    outputCapture: step.outputCapture,
                    onFailure: step.onFailure || 'exit',
                })),
            };
            const response: any = await apiPut(`/actions/${encodeURIComponent(id)}` as any, payload as any);
            if (response?.ok === false) throw new Error('Update failed');
            setSnack({open: true, message: 'Action updated.', severity: 'success'});
            void reloadFeatureAvailability();
            baselineRef.current = JSON.stringify({name, dialect, description, notes, steps});
            setDirty(false);
            navigate('/actions/list', {state: {refresh: true}});
        } catch (error) {
            console.error(error);
            setSnack({open: true, message: 'Failed to update action.', severity: 'error'});
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Typography variant="h5">Edit Action</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Button variant="outlined" onClick={() => setTestOpen(true)}>Test Action</Button>
                        <Button variant="outlined" onClick={() => confirmOnNavigate(dirty, navigate, confirmPrompt)('/actions/list')}>Cancel</Button>
                        <Button variant="contained" onClick={onSave}>Save Changes</Button>
                    </HStack>
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                {loadError && <Alert severity="error">{loadError}</Alert>}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        {loading ? (
                            <Typography variant="body2" sx={{color: 'text.secondary'}}>Loading…</Typography>
                        ) : (
                            <VStack spacing={3}>
                                <TextField label="ID" value={id} disabled fullWidth/>
                                <TextField label="Name" value={name} onChange={(event) => setName(event.target.value)} required fullWidth/>
                                <TextField label="Description (optional)" value={description} onChange={(event) => setDescription(event.target.value)} fullWidth multiline minRows={3}/>
                                <TextField label="Notes (optional)" value={notes} onChange={(event) => setNotes(event.target.value)} fullWidth multiline minRows={3}/>
                            </VStack>
                        )}
                    </CardContent>
                </Card>

                {!loading && (
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
                                updateStep={(stepId, patch) => setSteps((prev) => prev.map((entry) => (entry.id === stepId ? {...entry, ...patch} : entry)))}
                                onCheckStep={onCheckStep}
                                dialect={dialect}
                                onDialectChange={setDialect}
                                stepCheck={stepChecks[step.id]}
                            />
                        ))}
                    </VStack>
                )}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="outlined" onClick={() => setTestOpen(true)}>Test Action</Button>
                        <Button variant="outlined" onClick={() => confirmOnNavigate(dirty, navigate, confirmPrompt)('/actions/list')}>Cancel</Button>
                        <Button variant="contained" onClick={onSave}>Save Changes</Button>
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
