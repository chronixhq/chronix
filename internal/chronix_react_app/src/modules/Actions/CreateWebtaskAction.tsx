import {useEffect, useRef, useState} from 'react';
import {apiPost} from '@dsherwin/react-api-interface';
import {useNavigate} from 'react-router';
import {useMuiPrompts} from '@dsherwin/mui-kit';
import {useFeatureAvailability} from '../../data/FeatureAvailabilityContext.tsx';
import {confirmOnNavigate, useUnsavedChanges} from '../../lib/useUnsavedChanges.ts';
import {WebtaskActionEditor} from './WebtaskActionEditor';
import {type WebtaskStepDraft} from './types';
import {
    appendWebtaskDraftStep,
    buildWebtaskActionPayload,
    createWebtaskDraftStep,
    moveWebtaskDraftStep,
    removeWebtaskDraftStep,
    snapshotWebtaskActionDraft,
    updateWebtaskDraftStep,
    validateWebtaskAction,
} from './webtaskActionEditorUtils';

export const CreateWebtaskAction = () => {
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
    const [steps, setSteps] = useState<WebtaskStepDraft[]>(() => [createWebtaskDraftStep(1)]);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [testOpen, setTestOpen] = useState(false);
    const [helpOpen, setHelpOpen] = useState(false);
    const [dirty, setDirty] = useState(false);
    const baselineRef = useRef('');

    useUnsavedChanges(dirty);

    useEffect(() => {
        const snapshot = snapshotWebtaskActionDraft(name, description, notes, steps);
        if (!baselineRef.current) {
            baselineRef.current = snapshot;
            return;
        }
        setDirty(snapshot !== baselineRef.current);
    }, [name, description, notes, steps]);

    const onAddStep = () => setSteps((prev) => appendWebtaskDraftStep(prev));
    const onUpdateStep = (id: string, patch: Partial<WebtaskStepDraft>) => setSteps((prev) => updateWebtaskDraftStep(prev, id, patch));
    const onRemoveStep = (id: string) => setSteps((prev) => removeWebtaskDraftStep(prev, id));
    const onMoveStep = (id: string, direction: number) => setSteps((prev) => moveWebtaskDraftStep(prev, id, direction));

    async function onSave() {
        const validationError = validateWebtaskAction(name, steps);
        if (validationError) {
            setError(validationError);
            return;
        }

        setSaving(true);
        setError(null);
        try {
            await apiPost('/actions/webtask', buildWebtaskActionPayload(name, description, notes, steps));
            setDirty(false);
            void reloadFeatureAvailability();
            navigate('/actions/list');
        } catch (error: any) {
            setError(error?.message || 'Failed to create action');
        } finally {
            setSaving(false);
        }
    }

    return (
        <WebtaskActionEditor
            title="Create Web Task Action"
            saveLabel="Create Action"
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
