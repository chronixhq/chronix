import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardActions, CardContent, Divider, TextField, Typography} from '@mui/material';
import {Add} from '@mui/icons-material';
import {apiPost} from '@dsherwin/react-api-interface';
import {useNavigate} from 'react-router';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {TestActionDialog} from './TestActionDialog';
import {HStack, useMuiPrompts, VStack} from '@dsherwin/mui-kit';
import {buildShellActionPayload, createShellDraftStep} from './shellActionEditorUtils';
import {ShellActionStepCard} from './ShellActionStepCard';
import {type ShellStepDraft} from './types';

export const CreateShellAction = () => {
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
    const [description, setDescription] = useState('');
    const [notes, setNotes] = useState('');
    const [steps, setSteps] = useState<ShellStepDraft[]>([createShellDraftStep(1)]);
    const [saving, setSaving] = useState(false);
    const [testOpen, setTestOpen] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const onAddStep = () => {
        setSteps((prev) => [...prev, createShellDraftStep(prev.length + 1)]);
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

    const onMoveStep = (id: string, dir: number) => {
        setSteps((prev) => {
            const index = prev.findIndex((step) => step.id === id);
            const nextIndex = index + dir;
            if (nextIndex < 0 || nextIndex >= prev.length) return prev;
            const next = [...prev];
            [next[index], next[nextIndex]] = [next[nextIndex], next[index]];
            return next;
        });
    };

    const updateStep = (id: string, patch: Partial<ShellStepDraft>) => {
        setSteps((prev) => prev.map((step) => (step.id === id ? {...step, ...patch} : step)));
    };

    async function onSave() {
        if (!name.trim()) {
            setError('Name is required');
            return;
        }
        setSaving(true);
        setError(null);
        try {
            await apiPost('/shell/actions', buildShellActionPayload(name, description, notes, steps));
            void reloadFeatureAvailability();
            navigate('/actions/list');
        } catch (error: any) {
            setError(error?.message || 'Failed to create shell action');
        } finally {
            setSaving(false);
        }
    }

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" justifyContent="space-between" sx={{flexWrap: 'wrap'}}>
                    <Typography variant="h5">Create Shell Action</Typography>
                    <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                        <Button variant="outlined" onClick={() => setTestOpen(true)}>Test Action</Button>
                        <Button variant="outlined" onClick={() => navigate('/actions/list')}>Cancel</Button>
                        <Button variant="contained" onClick={() => void onSave()} disabled={saving || !name}>Save Action</Button>
                    </HStack>
                </HStack>
                <Divider/>

                {error && <Alert severity="error">{error}</Alert>}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={3}>
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
                        <Button variant="contained" onClick={() => void onSave()} disabled={saving || !name}>Save Action</Button>
                    </CardActions>
                </Card>

                <TestActionDialog open={testOpen} onClose={() => setTestOpen(false)} type="shell" steps={steps}/>
            </VStack>
        </Box>
    );
};
