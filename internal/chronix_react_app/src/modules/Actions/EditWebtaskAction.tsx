import {useEffect, useRef, useState} from 'react';
import {Alert, Box, CircularProgress} from '@mui/material';
import {apiGet, apiPut} from '@dsherwin/react-api-interface';
import {useNavigate, useParams} from 'react-router';
import {useMuiPrompts} from '@dsherwin/mui-kit';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {confirmOnNavigate, useUnsavedChanges} from '../../lib/useUnsavedChanges.ts';
import {WebtaskActionEditor} from './WebtaskActionEditor';
import {type WebtaskStepDraft} from './types';
import {
    appendWebtaskDraftStep,
    buildWebtaskActionPayload,
    moveWebtaskDraftStep,
    removeWebtaskDraftStep,
    snapshotWebtaskActionDraft,
    toWebtaskDraftSteps,
    updateWebtaskDraftStep,
    validateWebtaskAction,
} from './webtaskActionEditorUtils';

export const EditWebtaskAction = () => {
    const navigate = useNavigate();
    const {id} = useParams();
    const {reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {confirmPrompt} = useMuiPrompts();
    const [loading, setLoading] = useState(true);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [notes, setNotes] = useState('');
    const [steps, setSteps] = useState<WebtaskStepDraft[]>([]);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [testOpen, setTestOpen] = useState(false);
    const [helpOpen, setHelpOpen] = useState(false);
    const [dirty, setDirty] = useState(false);
    const baselineRef = useRef('');

    useUnsavedChanges(dirty);

    useEffect(() => {
        const load = async () => {
            setLoading(true);
            setLoadError(null);
            try {
                const data = await apiGet(`/actions/webtask/${id}`) as any;
                if (data?.suspended) {
                    navigate('/actions/list');
                    return;
                }

                const loadedSteps = toWebtaskDraftSteps(data.steps);
                setName(data.name);
                setDescription(data.description || '');
                setNotes(data.notes || '');
                setSteps(loadedSteps);
                baselineRef.current = snapshotWebtaskActionDraft(data.name, data.description || '', data.notes || '', loadedSteps);
                setDirty(false);
            } catch (error: any) {
                setLoadError(error?.message || 'Failed to load action');
            } finally {
                setLoading(false);
            }
        };

        void load();
    }, [id, navigate]);

    useEffect(() => {
        if (!baselineRef.current || loading) return;
        setDirty(snapshotWebtaskActionDraft(name, description, notes, steps) !== baselineRef.current);
    }, [name, description, notes, steps, loading]);

    const onAddStep = () => setSteps((prev) => appendWebtaskDraftStep(prev));
    const onUpdateStep = (stepId: string, patch: Partial<WebtaskStepDraft>) => setSteps((prev) => updateWebtaskDraftStep(prev, stepId, patch));
    const onRemoveStep = (stepId: string) => setSteps((prev) => removeWebtaskDraftStep(prev, stepId));
    const onMoveStep = (stepId: string, direction: number) => setSteps((prev) => moveWebtaskDraftStep(prev, stepId, direction));

    async function onSave() {
        const validationError = validateWebtaskAction(name, steps);
        if (validationError) {
            setError(validationError);
            return;
        }

        setSaving(true);
        setError(null);
        try {
            await apiPut(`/actions/webtask/${id}`, buildWebtaskActionPayload(name, description, notes, steps));
            setDirty(false);
            void reloadFeatureAvailability();
            navigate('/actions/list');
        } catch (error: any) {
            setError(error?.message || 'Failed to update action');
        } finally {
            setSaving(false);
        }
    }

    if (loading) {
        return (
            <Box sx={{p: 8, textAlign: 'center'}}>
                <CircularProgress/>
            </Box>
        );
    }

    if (loadError) {
        return (
            <Box sx={{px: {xs: 1, md: 2}, py: 2}}>
                <Alert severity="error">{loadError}</Alert>
            </Box>
        );
    }

    return (
        <WebtaskActionEditor
            title="Edit Web Task Action"
            saveLabel="Save Changes"
            name={name}
            description={description}
            notes={notes}
            steps={steps}
            saving={saving}
            error={error}
            testOpen={testOpen}
            helpOpen={helpOpen}
            onNameChange={setName}
            onDescriptionChange={setDescription}
            onNotesChange={setNotes}
            onAddStep={onAddStep}
            onUpdateStep={onUpdateStep}
            onRemoveStep={onRemoveStep}
            onMoveStep={onMoveStep}
            onOpenHelp={() => setHelpOpen(true)}
            onCloseHelp={() => setHelpOpen(false)}
            onOpenTest={() => setTestOpen(true)}
            onCloseTest={() => setTestOpen(false)}
            onCancel={() => confirmOnNavigate(dirty, navigate, confirmPrompt)('/actions/list')}
            onSave={() => void onSave()}
        />
    );
};
