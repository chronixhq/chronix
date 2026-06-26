import {Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField, Typography} from '@mui/material';
import {VStack} from '@dsherwin/mui-kit';

export const ActionSqlVariableDialog = ({
    open,
    vars,
    values,
    onClose,
    onConfirm,
    onValueChange,
}: {
    open: boolean;
    vars: string[];
    values: Record<string, string>;
    onClose: () => void;
    onConfirm: () => void;
    onValueChange: (name: string, value: string) => void;
}) => (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
        <DialogTitle>Provide test values</DialogTitle>
        <DialogContent dividers>
            <Typography variant="body2" component="p">
                Your SQL contains template variables. To check the SQL syntax, we need test values for each variable. These are only used for validation.
            </Typography>
            <VStack spacing={1}>
                {vars.map((variable) => (
                    <TextField
                        key={variable}
                        label={variable}
                        value={values[variable] || ''}
                        onChange={(event) => onValueChange(variable, event.target.value)}
                        fullWidth
                    />
                ))}
            </VStack>
        </DialogContent>
        <DialogActions>
            <Button onClick={onClose}>Cancel</Button>
            <Button variant="contained" onClick={onConfirm}>Run Check</Button>
        </DialogActions>
    </Dialog>
);
