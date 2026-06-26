import {Chip, TextField, Tooltip, Typography} from '@mui/material';
import {HStack, VStack} from '@dsherwin/mui-kit';

export const JobVariablesEditor = ({
    detectedVars,
    varValues,
    onChange,
}: {
    detectedVars: string[];
    varValues: Record<string, string>;
    onChange: (name: string, value: string) => void;
}) => (
    <VStack spacing={1}>
        <HStack alignItems="center" spacing={1}>
            <Typography variant="h6">Variables</Typography>
            <Tooltip title="Detected from the selected Action's SQL ({{name}})">
                <Chip size="small" label={`${detectedVars.length}`}/>
            </Tooltip>
        </HStack>
        {detectedVars.length === 0 ? (
            <Typography variant="body2" sx={{color: 'text.secondary'}}>
                No variables detected.
            </Typography>
        ) : (
            <VStack spacing={1}>
                {detectedVars.map((variable) => (
                    <TextField
                        key={variable}
                        label={variable}
                        value={varValues[variable] || ''}
                        onChange={(event) => onChange(variable, event.target.value)}
                        fullWidth
                    />
                ))}
            </VStack>
        )}
    </VStack>
);
