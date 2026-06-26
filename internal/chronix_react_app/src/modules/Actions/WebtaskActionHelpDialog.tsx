import {Button, Dialog, DialogActions, DialogContent, DialogTitle, Divider, Typography} from '@mui/material';

export const WebtaskActionHelpDialog = ({
    open,
    onClose,
}: {
    open: boolean;
    onClose: () => void;
}) => (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
        <DialogTitle>Web Task Help & Variables</DialogTitle>
        <DialogContent dividers>
            <Typography variant="subtitle1" gutterBottom sx={{fontWeight: 'bold'}}>
                Variable Substitution
            </Typography>
            <Typography variant="body2" component="div" gutterBottom>
                Chronix supports dynamic variable injection using the <code>{`\${variableName}`}</code> syntax.
                You can use variables in the URL, request headers, request body, and result expectations.
            </Typography>
            <Typography variant="body2" component="div" gutterBottom>
                <strong>Sources:</strong>
                <ul>
                    <li><strong>Job Variables:</strong> Variables defined at the job level when this action is scheduled.</li>
                    <li><strong>Captured Variables:</strong> Data extracted from previous step responses in this action using response capture.</li>
                </ul>
            </Typography>

            <Divider sx={{my: 2}}/>

            <Typography variant="subtitle1" gutterBottom sx={{fontWeight: 'bold'}}>
                Request Headers
            </Typography>
            <Typography variant="body2" component="div" gutterBottom>
                Custom HTTP headers to send with the request. Common uses include:
                <ul>
                    <li><code>Authorization: Bearer {'${apiToken}'}</code></li>
                    <li><code>Content-Type: application/json</code></li>
                    <li><code>X-API-Key: {'${secretKey}'}</code></li>
                </ul>
            </Typography>

            <Divider sx={{my: 2}}/>

            <Typography variant="subtitle1" gutterBottom sx={{fontWeight: 'bold'}}>
                Request Body
            </Typography>
            <Typography variant="body2" component="div" gutterBottom>
                The payload sent for `POST`, `PUT`, and `PATCH` requests. Variable substitution works here too,
                so you can build dynamic JSON or form-encoded bodies.
            </Typography>

            <Divider sx={{my: 2}}/>

            <Typography variant="subtitle1" gutterBottom sx={{fontWeight: 'bold'}}>
                Result Expectation
            </Typography>
            <Typography variant="body2" component="div" gutterBottom>
                Determines if a step is considered passed or failed. You can use variables in expectation values too:
                <ul>
                    <li><strong>Status Code:</strong> Compare the response code, for example <code>== 200</code>.</li>
                    <li><strong>Body Contains:</strong> Check if the raw response body includes a string.</li>
                    <li><strong>JSONPath Match:</strong> Evaluate a JSONPath expression and compare the result.</li>
                    <li><strong>Latency:</strong> Fail the step if the response takes too long.</li>
                </ul>
            </Typography>

            <Divider sx={{my: 2}}/>

            <Typography variant="subtitle1" gutterBottom sx={{fontWeight: 'bold'}}>
                Response Capture
            </Typography>
            <Typography variant="body2" component="div" gutterBottom>
                Extract data from the response to use in later steps or reporting:
                <ul>
                    <li><strong>JSONPath:</strong> Extract values from JSON, like <code>$.user.id</code>.</li>
                    <li><strong>Header:</strong> Capture a specific response header like <code>Set-Cookie</code>.</li>
                    <li><strong>Regex:</strong> Use a capture group to pull text out of any response body.</li>
                </ul>
            </Typography>
        </DialogContent>
        <DialogActions>
            <Button onClick={onClose} autoFocus>Close</Button>
        </DialogActions>
    </Dialog>
);
