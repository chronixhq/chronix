import React, {useCallback, useEffect, useState} from 'react';
import {
    Accordion,
    AccordionDetails,
    AccordionSummary,
    Alert,
    Box,
    Button,
    Chip,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    IconButton,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    TextField,
    Typography
} from '@mui/material';
import {Add as AddIcon, Close as CloseIcon, Delete as DeleteIcon, ExpandMore as ExpandMoreIcon} from '@mui/icons-material';
import {apiGet, apiPost} from '@dsherwin/react-api-interface';
import type {DatabaseStepTestResult, ShellStepTestResult, WebtaskStepTestResult} from './types';
import {HStack, VStack} from '@dsherwin/mui-kit';

interface Connection {
    id: number;
    name: string;
    agent_os?: string;
}

interface TestActionDialogProps {
    open: boolean;
    onClose: () => void;
    type: 'database' | 'shell' | 'webtask';
    steps: any[];
    dialect?: string;
}

function VariableEditor({vars, onChange, disabled}: { vars: Record<string, string>, onChange: (vars: Record<string, string>) => void, disabled?: boolean }) {
    const entries = Object.entries(vars);

    const onAdd = () => onChange({...vars, '': ''});
    const onRemove = (k: string) => {
        const next = {...vars};
        delete next[k];
        onChange(next);
    };
    const onUpdateKey = (oldK: string, newK: string) => {
        if (oldK === newK) return;
        const next = {...vars};
        const val = next[oldK];
        delete next[oldK];
        next[newK] = val;
        onChange(next);
    };
    const onUpdateVal = (k: string, v: string) => onChange({...vars, [k]: v});

    return (
        <Box sx={{ p: 2, border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
            <HStack alignItems="center" justifyContent="space-between" sx={{ mb: 1 }}>
                <Typography variant="subtitle2">Test Variables</Typography>
                <Button size="small" startIcon={<AddIcon/>} onClick={onAdd} disabled={disabled}>Add Variable</Button>
            </HStack>
            {entries.length === 0 && <Typography variant="body2" sx={{
                color: "text.secondary"
            }}>No variables defined.</Typography>}
            <VStack spacing={1}>
                {entries.map(([k, v], i) => (
                    <HStack key={i} spacing={1} alignItems="flex-start">
                        <TextField 
                            label="Key" 
                            size="small" 
                            value={k} 
                            onChange={(e) => onUpdateKey(k, e.target.value)} 
                            disabled={disabled}
                            sx={{ width: 200 }}
                        />
                        <TextField 
                            label="Value" 
                            size="small" 
                            value={v} 
                            onChange={(e) => onUpdateVal(k, e.target.value)} 
                            disabled={disabled}
                            fullWidth
                        />
                        <IconButton size="small" onClick={() => onRemove(k)} disabled={disabled} color="error" sx={{ mt: 0.5 }}>
                            <DeleteIcon fontSize="small"/>
                        </IconButton>
                    </HStack>
                ))}
            </VStack>
        </Box>
    );
}

export const TestActionDialog: React.FC<TestActionDialogProps> = ({open, onClose, type, steps, dialect}) => {
    const [connections, setConnections] = useState<Connection[]>([]);
    const [selectedConnId, setSelectedConnId] = useState<number | ''>('');
    const [testing, setTesting] = useState(false);
    const [results, setResults] = useState<(DatabaseStepTestResult | ShellStepTestResult | WebtaskStepTestResult)[]>([]);
    const [error, setError] = useState<string | null>(null);
    const [testVars, setTestVars] = useState<Record<string, string>>({});

    const loadConnections = useCallback(async () => {
        try {
            let endpoint = '/connections';
            if (type === 'shell') endpoint = '/shell/connections';
            if (type === 'webtask') endpoint = '/connections/webtask';
            
            const data = await apiGet(endpoint) as any;
            setConnections(Array.isArray(data) ? data : []);
        } catch (e) {
            console.error(e);
            setError('Failed to load connections');
        }
    }, [type]);

    useEffect(() => {
        if (open) {
            loadConnections();
            setResults([]);
            setError(null);

            if (type === 'database' || type === 'shell' || type === 'webtask') {
                const allVars = new Set<string>();
                const capturedVars = new Set<string>();
                steps.forEach(s => {
                    const texts = type === 'database' ? [s.sql || ''] : (type === 'shell' ? [s.command || '', s.scriptText || '', s.shellPath || '', s.workingDir || ''] : [s.url || '', s.body || '']);
                    
                    // Headers (for webtasks)
                    if (type === 'webtask' && s.headers && typeof s.headers === 'object') {
                        Object.values(s.headers).forEach(v => {
                            if (typeof v === 'string') texts.push(v);
                        });
                    }

                    // Env (for shell)
                    if (type === 'shell' && s.env && typeof s.env === 'object') {
                        Object.values(s.env).forEach(v => {
                            if (typeof v === 'string') texts.push(v);
                        });
                    }
                    
                    // Capture variables (for webtasks)
                    if (type === 'webtask' && s.responseCapture && typeof s.responseCapture === 'object') {
                        Object.keys(s.responseCapture).forEach(k => {
                            if (k) capturedVars.add(k);
                        });
                    }

                    // Capture variables (for database and shell)
                    if ((type === 'database' || type === 'shell') && s.outputCapture && typeof s.outputCapture === 'object') {
                        Object.keys(s.outputCapture).forEach(k => {
                            if (k) capturedVars.add(k);
                        });
                    }
                    
                    // Also check expectation values for variables
                    if (s.expectation && typeof s.expectation === 'object') {
                        Object.values(s.expectation).forEach(v => {
                            if (typeof v === 'string') texts.push(v);
                        });
                    }

                    const re = /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}/g;
                    const re2 = /\$\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}/g;
                    
                    texts.forEach(t => {
                        let m;
                        while ((m = re.exec(t))) {
                            allVars.add(m[1]);
                        }
                        while ((m = re2.exec(t))) {
                            allVars.add(m[1]);
                        }
                    });
                });

                // Remove captured vars from allVars because they will be supplied by the steps themselves
                capturedVars.forEach(v => allVars.delete(v));

                setTestVars(prev => {
                    const next: Record<string, string> = {};
                    
                    allVars.forEach(v => {
                        next[v] = prev[v] !== undefined ? prev[v] : '';
                    });
                    
                    return next;
                });
            }
        }
    }, [open, type, steps, loadConnections]);

    const selectedConn = connections.find(c => c.id === selectedConnId);
    const osMismatchSteps = type === 'shell' && selectedConn?.agent_os ? steps.filter(s => {
        const sp = (s.shellPath || '').toLowerCase();
        if (selectedConn.agent_os !== 'windows') {
            return sp.startsWith('c:\\') || sp.includes('\\windows\\');
        }
        if (selectedConn.agent_os === 'windows') {
            return sp.startsWith('/') && (sp.includes('/bin/') || sp.includes('/usr/'));
        }
        return false;
    }) : [];

    const onTest = async () => {
        if (selectedConnId === '') return;
        setTesting(true);
        setError(null);
        setResults([]);

        const filteredVars: Record<string, string> = {};
        for (const [k, v] of Object.entries(testVars)) {
            if (k.trim()) {
                filteredVars[k.trim()] = v;
            }
        }

        try {
            let endpoint = '/actions/test';
            if (type === 'shell') endpoint = '/shell/actions/test';
            if (type === 'webtask') endpoint = '/actions/webtask/test';

            let payload: any;
            if (type === 'database') {
                payload = {
                    connectionId: selectedConnId,
                    dialect,
                    steps: steps.map((s, order) => ({
                        order,
                        name: s.name,
                        sqlText: s.sql,
                        timeoutSeconds: Number(s.timeout || '60'),
                        expectation: s.expectation,
                        outputCapture: s.outputCapture,
                        onFailure: s.onFailure || 'exit'
                    })),
                    variables: filteredVars
                };
            } else if (type === 'shell') {
                payload = {
                    shellConnectionId: selectedConnId,
                    steps: steps.map((s, order) => ({
                        order,
                        name: s.name,
                        runMode: s.runMode,
                        command: s.command,
                        scriptText: s.scriptText,
                        shellPath: s.shellPath,
                        workingDir: s.workingDir,
                        timeoutSeconds: Number(s.timeout || '60'),
                        outputCaptureMaxBytes: Number(s.outputCaptureMaxBytes || '65536'),
                        outputTruncation: s.outputTruncation,
                        expectation: s.expectation,
                        outputCapture: s.outputCapture,
                        onFailure: s.onFailure || 'exit',
                        env: s.env
                    })),
                    variables: filteredVars
                };
            } else {
                payload = {
                    connectionId: selectedConnId,
                    steps: steps.map((s, i) => ({
                        stepOrder: i + 1,
                        name: s.name,
                        method: s.method,
                        url: s.url,
                        headers: s.headers,
                        body: s.body,
                        timeoutSeconds: Number(s.timeout || '60'),
                        expectation: s.expectation,
                        responseCapture: s.responseCapture,
                        onFailure: s.onFailure || 'exit'
                    })),
                    variables: filteredVars
                };
            }

            const data = await apiPost(endpoint, payload) as any;
            setResults(Array.isArray(data) ? data : []);
        } catch (e: any) {
            console.error(e);
            setError(e.message || 'Test execution failed');
        } finally {
            setTesting(false);
        }
    };

    return (
        <Dialog 
            open={open} 
            onClose={(_e, reason) => {
                if (reason !== 'backdropClick') {
                    onClose();
                }
            }} 
            maxWidth="lg" 
            fullWidth
        >
            <DialogTitle sx={{ m: 0, p: 2 }}>
                Test Action
                <IconButton
                    aria-label="close"
                    onClick={onClose}
                    sx={(theme) => ({
                        position: 'absolute',
                        right: 8,
                        top: 8,
                        color: theme.palette.error.main,
                    })}
                >
                    <CloseIcon />
                </IconButton>
            </DialogTitle>
            <DialogContent dividers>
                <VStack spacing={2}>
                    <HStack spacing={2} alignItems="center">
                        <FormControl sx={{minWidth: 300}}>
                            <InputLabel id="conn-select-label">Select Connection</InputLabel>
                            <Select
                                labelId="conn-select-label"
                                label="Select Connection"
                                value={selectedConnId}
                                onChange={(e) => setSelectedConnId(e.target.value as number)}
                                disabled={testing}
                            >
                                {connections.map(c => (
                                    <MenuItem key={c.id} value={c.id}>{c.name}</MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    </HStack>

                    <VariableEditor 
                        vars={testVars} 
                        onChange={setTestVars} 
                        disabled={testing} 
                    />

                    {error && <Alert severity="error">{error}</Alert>}

                    {osMismatchSteps.length > 0 && (
                        <Alert severity="warning">
                            The selected connection ({selectedConn?.name}) is running {selectedConn?.agent_os}, but {osMismatchSteps.length} step(s) appear to use a shell path for a different OS.
                        </Alert>
                    )}

                    {results.length > 0 && (
                        <VStack spacing={2} sx={{mt: 2}}>
                            <Typography variant="h6">Results</Typography>
                            {results.map((res, i) => (
                                <Accordion key={i} defaultExpanded={res.status !== 'success'}>
                                    <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                                        <HStack spacing={2} alignItems="center" sx={{width: '100%'}}>
                                            <Typography sx={{fontWeight: 600}}>Step {res.order}: {res.name}</Typography>
                                            <Chip 
                                                label={res.status.toUpperCase()} 
                                                color={res.status === 'success' ? 'success' : 'error'} 
                                                size="small" 
                                            />
                                            <Box sx={{flexGrow: 1}} />
                                            {res.expectationOk !== undefined && (
                                                <Typography variant="caption" color={res.expectationOk ? 'success.main' : 'error.main'}>
                                                    Expectation: {res.expectationMsg}
                                                </Typography>
                                            )}
                                        </HStack>
                                    </AccordionSummary>
                                    <AccordionDetails>
                                        <VStack spacing={1}>
                                            {res.executionError && (
                                                <Alert severity="error">Execution Error: {res.executionError}</Alert>
                                            )}

                                            {type === 'database' ? (
                                                <DbResultDetail res={res as DatabaseStepTestResult} />
                                            ) : type === 'shell' ? (
                                                <ShellResultDetail res={res as ShellStepTestResult} />
                                            ) : (
                                                <WebtaskResultDetail res={res as WebtaskStepTestResult} />
                                            )}
                                        </VStack>
                                    </AccordionDetails>
                                </Accordion>
                            ))}
                        </VStack>
                    )}
                </VStack>
            </DialogContent>
            <DialogActions sx={{ px: 3, pb: 2 }}>
                <Button
                    variant="contained"
                    color="primary"
                    onClick={onTest}
                    disabled={testing || selectedConnId === ''}
                    startIcon={testing && <CircularProgress size={20} color="inherit" />}
                >
                    {testing ? 'Testing...' : 'Run Test'}
                </Button>
                <Button onClick={onClose} variant="outlined" color="error">Close</Button>
            </DialogActions>
        </Dialog>
    );
};

const DbResultDetail = ({res}: {res: DatabaseStepTestResult}) => (
    <VStack spacing={1}>
        {res.executedCode && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "text.secondary"
                    }}>EXECUTED SQL</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'primary.light',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto',
                    whiteSpace: 'pre-wrap'
                }}>
                    {res.executedCode}
                </Box>
            </VStack>
        )}
        {res.executedArgs && res.executedArgs.length > 0 && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "text.secondary"
                    }}>ARGUMENTS</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'info.light',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto'
                }}>
                    {res.executedArgs.map((arg, idx) => (
                        <div key={idx}>${idx + 1}: {JSON.stringify(arg)}</div>
                    ))}
                </Box>
            </VStack>
        )}
        <HStack spacing={2}>
            <Typography variant="body2">Rows Count: <strong>{res.rowsCount}</strong></Typography>
            <Typography variant="body2">Rows Affected: <strong>{res.rowsAffected}</strong></Typography>
        </HStack>
        {res.capturedVars && Object.keys(res.capturedVars).length > 0 && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "success.main"
                    }}>CAPTURED VARIABLES</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'success.light',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto'
                }}>
                    {Object.entries(res.capturedVars).map(([k, v]) => (
                        <div key={k}>{k}: {JSON.stringify(v)}</div>
                    ))}
                </Box>
            </VStack>
        )}
        {res.resultLines && res.resultLines.length > 0 && (
            <TableContainer component={Paper} variant="outlined" sx={{maxHeight: 300}}>
                <Table size="small" stickyHeader>
                    <TableHead>
                        <TableRow>
                            {Object.keys(res.resultLines[0]).map(k => (
                                <TableCell key={k} sx={{fontWeight: 600}}>{k}</TableCell>
                            ))}
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {res.resultLines.map((row, i) => (
                            <TableRow key={i}>
                                {Object.values(row).map((v, j) => (
                                    <TableCell key={j}>{String(v)}</TableCell>
                                ))}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </TableContainer>
        )}
    </VStack>
);

const ShellResultDetail = ({res}: {res: ShellStepTestResult}) => (
    <VStack spacing={1}>
        {res.executedCode && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "text.secondary"
                    }}>EXECUTED CODE</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'primary.light',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto',
                    whiteSpace: 'pre-wrap'
                }}>
                    {res.executedCode}
                </Box>
            </VStack>
        )}
        <Typography variant="body2">Exit Code: <strong>{res.exitCode}</strong></Typography>
        
        {res.capturedVars && Object.keys(res.capturedVars).length > 0 && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "success.main"
                    }}>CAPTURED VARIABLES</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'success.light',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto'
                }}>
                    {Object.entries(res.capturedVars).map(([k, v]) => (
                        <div key={k}>{k}: {JSON.stringify(v)}</div>
                    ))}
                </Box>
            </VStack>
        )}

        {res.stdout && (
            <VStack spacing={0.5}>
                <Typography variant="caption" sx={{
                    fontWeight: 600
                }}>STDOUT {res.stdoutTruncated && '(Truncated)'}</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'common.white',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto',
                    whiteSpace: 'pre-wrap'
                }}>
                    {res.stdout}
                </Box>
            </VStack>
        )}

        {res.stderr && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "error.main"
                    }}>STDERR {res.stderrTruncated && '(Truncated)'}</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'error.light',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto',
                    whiteSpace: 'pre-wrap'
                }}>
                    {res.stderr}
                </Box>
            </VStack>
        )}
    </VStack>
);

const WebtaskResultDetail = ({res}: {res: WebtaskStepTestResult}) => (
    <VStack spacing={1}>
        <HStack spacing={2}>
            <Typography variant="body2">Status: <strong>{res.responseStatus}</strong></Typography>
            <Typography variant="body2">Latency: <strong>{res.latencyMs}ms</strong></Typography>
        </HStack>
        <VStack spacing={0.5}>
            <Typography
                variant="caption"
                sx={{
                    fontWeight: 600,
                    color: "text.secondary"
                }}>REQUEST URL</Typography>
            <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{res.requestMethod} {res.requestUrl}</Typography>
        </VStack>
        {res.capturedVars && Object.keys(res.capturedVars).length > 0 && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "success.main"
                    }}>CAPTURED VARIABLES</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'success.light',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto'
                }}>
                    {Object.entries(res.capturedVars).map(([k, v]) => (
                        <div key={k}>{k}: {JSON.stringify(v)}</div>
                    ))}
                </Box>
            </VStack>
        )}
        {res.responseBody && (
            <VStack spacing={0.5}>
                <Typography
                    variant="caption"
                    sx={{
                        fontWeight: 600,
                        color: "text.secondary"
                    }}>RESPONSE BODY</Typography>
                <Box sx={{
                    p: 1,
                    backgroundColor: 'grey.900',
                    color: 'common.white',
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    borderRadius: 1,
                    overflowX: 'auto',
                    whiteSpace: 'pre-wrap',
                    maxHeight: 300
                }}>
                    {res.responseBody}
                </Box>
            </VStack>
        )}
    </VStack>
);
