import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, Snackbar, TextField, Typography} from '@mui/material';
import {Add} from '@mui/icons-material';
import {apiGet, apiPut} from '@dsherwin/react-api-interface';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {useNavigate, useParams} from 'react-router';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {TestActionDialog} from './TestActionDialog';
import {buildShellActionPayload, createShellDraftStep, toShellDraftSteps} from './shellActionEditorUtils';
import {ShellActionStepCard} from './ShellActionStepCard';
import {type ShellStepDraft} from './types';

export const EditShellAction = () => {
    const {id} = useParams();
    const navigate = useNavigate();
    const {reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {confirmPrompt} = useMuiPrompts();
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [notes, setNotes] = useState('');
    const [steps, setSteps] = useState<ShellStepDraft[]>([]);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [testOpen, setTestOpen] = useState(false);
    const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({open: false, message: '', severity: 'success'});

    useEffect(() => {
        const load = async () => {
            try {
                const data = await apiGet(`/shell/actions/${id}`) as any;
                if ((data as any).suspended) {
                    navigate('/actions/list');
                    return;
                }
                setName(data.name);
                setDescription(data.description || '');
                setNotes(data.notes || '');
                setSteps(toShellDraftSteps(data.steps));
            } catch (error: any) {
                setError(error?.message || 'Failed to load shell action');
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [id, navigate]);

    const onAddStep = () => {
        setSteps((prev) => [...prev, createShellDraftStep(prev.length + 1)]);
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

    const onMoveStep = (stepId: string, dir: number) => {
        setSteps((prev) => {
            const index = prev.findIndex((step) => step.id === stepId);
            const nextIndex = index + dir;
            if (nextIndex < 0 || nextIndex >= prev.length) return prev;
            const next = [...prev];
            [next[index], next[nextIndex]] = [next[nextIndex], next[index]];
            return next;
        });
    };

    const updateStep = (stepId: string, patch: Partial<ShellStepDraft>) => {
        setSteps((prev) => prev.map((step) => (step.id === stepId ? {...step, ...patch} : step)));
    };

    async function onSave() {
        if (!name.trim()) {
            setError('Name is required');
            return;
        }
        setSaving(true);
        setError(null);
        try {
            await apiPut(`/shell/actions/${id}`, buildShellActionPayload(name, description, notes, steps));
            setSnack({open: true, message: 'Action updated successfully', severity: 'success'});
            void reloadFeatureAvailability();
            navigate('/actions/list', {state: {refresh: true}});
        } catch (error: any) {
            setError(error?.message || 'Failed to update shell action');
        } finally {
            setSaving(false);
        }
    }

    if (loading) return <Box sx={{p: 2}}>Loading...</Box>;

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Typography variant="h5">Edit Shell Action</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Button variant="outlined" onClick={() => setTestOpen(true)}>Test Action</Button>
                        <Button variant="outlined" onClick={() => navigate('/actions/list')}>Cancel</Button>
                        <Button variant="contained" onClick={() => void onSave()} disabled={saving || !name}>Save Changes</Button>
                    </HStack>
                </HStack>
                <Divider/>

                {error && <Alert severity="error">{error}</Alert>}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={3}>
                            <TextField label="ID" value={id} disabled fullWidth/>
                            <TextField label="Name" value={name} onChange={(event) => setName(event.target.value)} required fullWidth/>
                            <TextField label="Description (optional)" value={description} onChange={(event) => setDescription(event.target.value)} fullWidth multiline minRows={2}/>
                            <TextField label="Notes (optional)" value={notes} onChange={(event) => setNotes(event.target.value)} fullWidth multiline minRows={2}/>
                        </VStack>
                    </CardContent>
                </Card>

                <VStack spacing={2}>
                    <HStack alignItems="center" justifyContent="space-between">
                        <Typography variant="h6">Steps</Typography>
                        <Button variant="outlined" startIcon={<Add/>} onClick={onAddStep}>Add step</Button>
                    </HStack>
                    {steps.map((step, index) => (
                        <ShellActionStepCard
                            key={step.id}
                            step={step}
                            index={index}
                            stepsLength={steps.length}
                            onRemove={onRemoveStep}
                            onMove={onMoveStep}
                            updateStep={updateStep}
                        />
                    ))}
                </VStack>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardActions sx={{justifyContent: 'flex-end'}}>
                        <Button variant="outlined" onClick={() => setTestOpen(true)}>Test Action</Button>
                        <Button variant="outlined" onClick={() => navigate('/actions/list')}>Cancel</Button>
                        <Button variant="contained" onClick={() => void onSave()} disabled={saving || !name}>Save Changes</Button>
                    </CardActions>
                </Card>

                <TestActionDialog open={testOpen} onClose={() => setTestOpen(false)} type="shell" steps={steps}/>

                <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack((current) => ({...current, open: false}))}>
                    <Alert severity={snack.severity}>{snack.message}</Alert>
                </Snackbar>
            </VStack>
        </Box>
    );
};
