import {Add} from '@mui/icons-material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import {Alert, Box, Button, Card, CardContent, Divider, Stack, TextField, Tooltip, Typography} from '@mui/material';
import {HStack, VStack} from '@dsherwin/mui-kit';
import {TestActionDialog} from './TestActionDialog';
import {type WebtaskStepDraft} from './types';
import {WebtaskActionHelpDialog} from './WebtaskActionHelpDialog';
import {WebtaskActionStepCard} from './WebtaskActionStepCard';
import {toWebtaskTestSteps} from './webtaskActionEditorUtils';

export const WebtaskActionEditor = ({
    title,
    saveLabel,
    name,
    description,
    notes,
    steps,
    saving,
    error,
    testOpen,
    helpOpen,
    onNameChange,
    onDescriptionChange,
    onNotesChange,
    onAddStep,
    onUpdateStep,
    onRemoveStep,
    onMoveStep,
    onOpenHelp,
    onCloseHelp,
    onOpenTest,
    onCloseTest,
    onCancel,
    onSave,
}: {
    title: string;
    saveLabel: string;
    name: string;
    description: string;
    notes: string;
    steps: WebtaskStepDraft[];
    saving: boolean;
    error: string | null;
    testOpen: boolean;
    helpOpen: boolean;
    onNameChange: (value: string) => void;
    onDescriptionChange: (value: string) => void;
    onNotesChange: (value: string) => void;
    onAddStep: () => void;
    onUpdateStep: (id: string, patch: Partial<WebtaskStepDraft>) => void;
    onRemoveStep: (id: string) => void;
    onMoveStep: (id: string, direction: number) => void;
    onOpenHelp: () => void;
    onCloseHelp: () => void;
    onOpenTest: () => void;
    onCloseTest: () => void;
    onCancel: () => void;
    onSave: () => void;
}) => (
    <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
        <VStack spacing={2} sx={{maxWidth: 1000, width: '100%', mx: 'auto', pb: 10}}>
            <HStack alignItems="center" spacing={1} sx={{justifyContent: 'space-between', flexWrap: 'wrap'}}>
                <HStack alignItems="center" spacing={1}>
                    <Typography variant="h5">{title}</Typography>
                    <Tooltip title="Chain multiple API requests together.">
                        <InfoOutlinedIcon fontSize="small"/>
                    </Tooltip>
                </HStack>
                <HStack spacing={1} sx={{mt: {xs: 1, sm: 0}}}>
                    <Button variant="outlined" onClick={onOpenTest}>Test Action</Button>
                    <Button variant="outlined" onClick={onCancel}>Cancel</Button>
                    <Button variant="contained" color="primary" onClick={onSave} disabled={saving}>{saveLabel}</Button>
                </HStack>
            </HStack>
            <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

            {error && <Alert severity="error">{error}</Alert>}

            <Card variant="outlined" sx={{borderRadius: 3}}>
                <CardContent>
                    <Stack spacing={2}>
                        <TextField label="Action Name" value={name} onChange={(event) => onNameChange(event.target.value)} fullWidth required/>
                        <TextField label="Description (optional)" value={description} onChange={(event) => onDescriptionChange(event.target.value)} fullWidth multiline rows={2}/>
                        <TextField label="Notes (optional)" value={notes} onChange={(event) => onNotesChange(event.target.value)} fullWidth multiline rows={2}/>
                    </Stack>
                </CardContent>
            </Card>

            <HStack justifyContent="space-between" alignItems="center" sx={{mt: 2}}>
                <Typography variant="h6">Steps</Typography>
                <Button startIcon={<Add/>} onClick={onAddStep} variant="outlined" size="small">
                    Add Step
                </Button>
            </HStack>

            <VStack spacing={2}>
                {steps.map((step, index) => (
                    <WebtaskActionStepCard
                        key={step.id}
                        step={step}
                        index={index}
                        stepsLength={steps.length}
                        updateStep={onUpdateStep}
                        onRemove={onRemoveStep}
                        onMove={onMoveStep}
                        onOpenHelp={onOpenHelp}
                    />
                ))}
            </VStack>

            <Button startIcon={<Add/>} onClick={onAddStep} variant="outlined" fullWidth sx={{py: 2, borderStyle: 'dashed', borderRadius: 3}}>
                Add Another Step
            </Button>

            <TestActionDialog open={testOpen} onClose={onCloseTest} type="webtask" steps={toWebtaskTestSteps(steps)}/>
            <WebtaskActionHelpDialog open={helpOpen} onClose={onCloseHelp}/>
        </VStack>
    </Box>
);
