import {type ReactNode} from 'react';
import {Alert, Button, Card, CardContent, CardHeader, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControl, FormControlLabel, FormHelperText, IconButton, InputAdornment, InputLabel, MenuItem, Radio, RadioGroup, Select, Stack, Switch, TextField, Typography} from '@mui/material';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import CheckIcon from '@mui/icons-material/Check';
import {HStack} from '@dsherwin/mui-kit';
import {type ShellConnectionDraft, type ShellConnectionSecretFlags, type ShellConnectionUiState} from './shellConnectionEditorUtils';

const sectionTitleSlotProps = {title: {sx: {fontSize: '1.2rem'}}} as const;

function sectionCard(title: string, children: ReactNode) {
    return (
        <Card variant="outlined" sx={{borderRadius: 3}}>
            <CardHeader title={title} sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}} slotProps={sectionTitleSlotProps}/>
            <Divider/>
            <CardContent>{children}</CardContent>
        </Card>
    );
}

export const ShellConnectionForm = ({
    draft,
    onDraftChange,
    secretFlags,
    uiState,
    onUiChange,
    agents,
    agentsLoading,
    generatingKey,
    onGenerateKeyPair,
    onCopyGeneratedPublicKey,
    onDismissGeneratedPublicKey,
    onClearSecret,
}: {
    draft: ShellConnectionDraft;
    onDraftChange: (patch: Partial<ShellConnectionDraft>) => void;
    secretFlags: ShellConnectionSecretFlags;
    uiState: ShellConnectionUiState;
    onUiChange: (patch: Partial<ShellConnectionUiState>) => void;
    agents: Array<{ uuid: string; name: string }>;
    agentsLoading: boolean;
    generatingKey: boolean;
    onGenerateKeyPair: () => void;
    onCopyGeneratedPublicKey: () => void;
    onDismissGeneratedPublicKey: () => void;
    onClearSecret?: (field: 'ssh_password' | 'ssh_private_key' | 'ssh_key_pass' | 'sudo_password', label: string) => void;
}) => (
    <>
        {sectionCard('General Information', (
            <Stack spacing={2.5}>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    Give this connection a name and description so it stays easy to identify later.
                </Typography>
                <TextField label="Connection name" value={draft.name} onChange={(event) => onDraftChange({name: event.target.value})} fullWidth required helperText="Example: Production Web Server"/>
                <TextField label="Description (optional)" placeholder="What is this connection used for?" value={draft.description} onChange={(event) => onDraftChange({description: event.target.value})} fullWidth multiline minRows={2} maxRows={6}/>
            </Stack>
        ))}

        {sectionCard('Execution Target', (
            <Stack spacing={2.5}>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    Specify where commands will execute. You can run them on the server or route them through a Chronix Agent.
                </Typography>
                <FormControl fullWidth disabled={agentsLoading}>
                    <InputLabel id="agent-label">Agent (optional)</InputLabel>
                    <Select labelId="agent-label" label="Agent (optional)" value={draft.agentUUID || ''} onChange={(event) => onDraftChange({agentUUID: (event.target.value as string) || ''})}>
                        <MenuItem value=""><em>Server (no agent)</em></MenuItem>
                        {agents.map((agent) => (
                            <MenuItem key={agent.uuid} value={agent.uuid}>{agent.name}</MenuItem>
                        ))}
                    </Select>
                    <FormHelperText>
                        Route execution through a specific Chronix Agent, or leave empty to run on the server.
                    </FormHelperText>
                </FormControl>
                <FormControl>
                    <Typography variant="subtitle2" sx={{mb: 0.5}}>Connection Mode</Typography>
                    <RadioGroup row value={draft.mode} onChange={(_, value) => onDraftChange({mode: (value as any) || 'localhost'})}>
                        <FormControlLabel value="localhost" control={<Radio/>} label="Localhost"/>
                        <FormControlLabel value="ssh" control={<Radio/>} label="SSH"/>
                    </RadioGroup>
                    <FormHelperText>Localhost runs on the server or selected Agent. SSH connects to a remote host.</FormHelperText>
                </FormControl>
            </Stack>
        ))}

        {draft.mode === 'ssh' && sectionCard('SSH Connection Details', (
            <Stack spacing={2.5}>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    Specify how to connect to the remote host over SSH.
                </Typography>
                <HStack spacing={2}>
                    <TextField label="Host" placeholder="server.example.com" value={draft.host} onChange={(event) => onDraftChange({host: event.target.value})} required sx={{flex: 1}}/>
                    <TextField
                        label="Port"
                        value={draft.port}
                        onChange={(event) => onDraftChange({port: event.target.value.replace(/[^0-9]/g, '')})}
                        sx={{width: 120}}
                        helperText="Default 22"
                        slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}} as any}
                    />
                </HStack>
                <TextField label="SSH Username" placeholder="e.g., ubuntu" value={draft.sshUsername} onChange={(event) => onDraftChange({sshUsername: event.target.value})} required fullWidth/>
                <FormControl>
                    <Typography variant="subtitle2" sx={{mb: 0.5}}>Authentication Method</Typography>
                    <RadioGroup row value={draft.authMethod} onChange={(_, value) => onDraftChange({authMethod: (value as any) || 'key'})}>
                        <FormControlLabel value="password" control={<Radio/>} label="Password"/>
                        <FormControlLabel value="key" control={<Radio/>} label="Private Key"/>
                    </RadioGroup>
                </FormControl>

                {draft.authMethod === 'password' ? (
                    <TextField
                        label={secretFlags.hasPassword ? 'SSH Password (set — leave blank to keep)' : 'SSH Password'}
                        type={uiState.showSshPassword ? 'text' : 'password'}
                        value={draft.sshPassword}
                        onChange={(event) => onDraftChange({sshPassword: event.target.value})}
                        fullWidth
                        helperText="Password for the SSH user. Stored securely."
                        slotProps={{
                            input: {
                                endAdornment: (
                                    <InputAdornment position="end">
                                        {secretFlags.hasPassword && onClearSecret && (
                                            <Button size="small" color="error" variant="text" onClick={() => onClearSecret('ssh_password', 'SSH password')} sx={{mr: 1}}>
                                                Clear
                                            </Button>
                                        )}
                                        <IconButton onClick={() => onUiChange({showSshPassword: !uiState.showSshPassword})} edge="end" aria-label="toggle password visibility">
                                            {uiState.showSshPassword ? <VisibilityOff/> : <Visibility/>}
                                        </IconButton>
                                    </InputAdornment>
                                ),
                            },
                        }}
                    />
                ) : (
                    <>
                        {uiState.generatedPublicKey && (
                            <Alert severity="info" sx={{mb: 2}}>
                                <Typography variant="subtitle2">New Public Key Generated:</Typography>
                                <Typography variant="body2" sx={{fontFamily: 'monospace', mt: 1, wordBreak: 'break-all', userSelect: 'all'}}>
                                    {uiState.generatedPublicKey}
                                </Typography>
                                <Typography variant="caption" sx={{mt: 1, display: 'block'}}>
                                    Copy this public key and add it to your target&apos;s <code>~/.ssh/authorized_keys</code> file.
                                </Typography>
                                <HStack spacing={1} sx={{mt: 1}}>
                                    <Button size="small" startIcon={uiState.copied ? <CheckIcon/> : <ContentCopyIcon/>} onClick={onCopyGeneratedPublicKey}>
                                        {uiState.copied ? 'Copied!' : 'Copy'}
                                    </Button>
                                    <Button size="small" onClick={onDismissGeneratedPublicKey}>Dismiss</Button>
                                </HStack>
                            </Alert>
                        )}

                        <HStack justifyContent="space-between" alignItems="center">
                            <Typography variant="subtitle2">
                                {secretFlags.hasPrivateKey ? 'Private Key (set — paste to replace)' : `Private Key (${uiState.keyFormat === 'openssh' ? 'OpenSSH' : 'PEM'})`}
                            </Typography>
                            <Button size="small" variant="outlined" onClick={() => onUiChange({showFormatDialog: true})} disabled={generatingKey}>
                                {generatingKey ? 'Generating...' : 'Generate Key Pair'}
                            </Button>
                        </HStack>
                        <TextField
                            placeholder="-----BEGIN PRIVATE KEY-----"
                            value={draft.sshPrivateKey}
                            onChange={(event) => onDraftChange({sshPrivateKey: event.target.value})}
                            fullWidth
                            multiline
                            minRows={6}
                            maxRows={10}
                            helperText="Paste your OpenSSH or PEM formatted private key."
                            slotProps={{
                                input: {
                                    endAdornment: secretFlags.hasPrivateKey && onClearSecret ? (
                                        <InputAdornment position="end" sx={{alignSelf: 'flex-start', mt: 1}}>
                                            <Button size="small" color="error" variant="text" onClick={() => onClearSecret('ssh_private_key', 'private key')}>
                                                Clear
                                            </Button>
                                        </InputAdornment>
                                    ) : undefined,
                                },
                            }}
                        />
                        <TextField
                            label={secretFlags.hasKeyPass ? 'Key Passphrase (set — type to replace)' : 'Key Passphrase (optional)'}
                            type={uiState.showSshKeyPass ? 'text' : 'password'}
                            value={draft.sshKeyPass}
                            onChange={(event) => onDraftChange({sshKeyPass: event.target.value})}
                            fullWidth
                            helperText="If your private key is protected by a passphrase."
                            slotProps={{
                                input: {
                                    endAdornment: (
                                        <InputAdornment position="end">
                                            {secretFlags.hasKeyPass && onClearSecret && (
                                                <Button size="small" color="error" variant="text" onClick={() => onClearSecret('ssh_key_pass', 'key passphrase')} sx={{mr: 1}}>
                                                    Clear
                                                </Button>
                                            )}
                                            <IconButton onClick={() => onUiChange({showSshKeyPass: !uiState.showSshKeyPass})} edge="end" aria-label="toggle password visibility">
                                                {uiState.showSshKeyPass ? <VisibilityOff/> : <Visibility/>}
                                            </IconButton>
                                        </InputAdornment>
                                    ),
                                },
                            }}
                        />
                    </>
                )}
            </Stack>
        ))}

        {sectionCard('Privileges & User Context', (
            <Stack spacing={2.5}>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    Configure how the shell session is initialized on the target machine.
                </Typography>
                <TextField label="Run as user (optional)" placeholder="e.g., chronix" value={draft.runAsUser} onChange={(event) => onDraftChange({runAsUser: event.target.value})} fullWidth/>
                <FormControl>
                    <FormControlLabel control={<Switch checked={draft.sudo} onChange={(event) => onDraftChange({sudo: event.target.checked})}/>} label="Use sudo"/>
                    <FormHelperText>Execute commands with elevated privileges using sudo.</FormHelperText>
                </FormControl>

                {draft.sudo && (
                    <TextField
                        label={secretFlags.hasSudoPassword ? 'Sudo Password (set — type to replace)' : 'Sudo Password (optional)'}
                        type={uiState.showSudoPassword ? 'text' : 'password'}
                        value={draft.sudoPassword}
                        onChange={(event) => onDraftChange({sudoPassword: event.target.value})}
                        fullWidth
                        helperText="Used for sudo authentication via stdin (-S) if required."
                        slotProps={{
                            input: {
                                endAdornment: (
                                    <InputAdornment position="end">
                                        {secretFlags.hasSudoPassword && onClearSecret && (
                                            <Button size="small" color="error" variant="text" onClick={() => onClearSecret('sudo_password', 'sudo password')} sx={{mr: 1}}>
                                                Clear
                                            </Button>
                                        )}
                                        <IconButton onClick={() => onUiChange({showSudoPassword: !uiState.showSudoPassword})} edge="end" aria-label="toggle password visibility">
                                            {uiState.showSudoPassword ? <VisibilityOff/> : <Visibility/>}
                                        </IconButton>
                                    </InputAdornment>
                                ),
                            },
                        }}
                    />
                )}
            </Stack>
        ))}

        {sectionCard('Health Check', (
            <Stack spacing={2}>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    Enable automatic connectivity checks to make sure this connection stays healthy.
                </Typography>
                <FormControlLabel control={<Switch checked={draft.autoCheckEnabled} onChange={(event) => onDraftChange({autoCheckEnabled: event.target.checked})}/>} label="Enable Automatic Health Checks"/>
                {draft.autoCheckEnabled && (
                    <TextField
                        label="Check Interval (seconds)"
                        value={draft.autoCheckInterval}
                        onChange={(event) => onDraftChange({autoCheckInterval: event.target.value.replace(/[^0-9]/g, '')})}
                        sx={{maxWidth: 300}}
                        helperText="How often to probe the connection. Recommended: 300 (5 minutes)."
                        slotProps={{htmlInput: {inputMode: 'numeric', pattern: '[0-9]*'}}}
                    />
                )}
            </Stack>
        ))}

        {sectionCard('Alerts', (
            <Stack spacing={3}>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                    Configure specific destinations for alerts related to this connection&apos;s health.
                </Typography>
                <TextField label="Alert Emails" placeholder="email1@example.com, email2@example.com" value={draft.alertEmails} onChange={(event) => onDraftChange({alertEmails: event.target.value})} fullWidth helperText="Comma-separated list of email addresses. If empty, system defaults are used."/>
                <TextField label="Alert Phones (SMS)" placeholder="+15550001111" value={draft.alertPhones} onChange={(event) => onDraftChange({alertPhones: event.target.value})} fullWidth helperText="Comma-separated list of E.164 phone numbers. If empty, system defaults are used."/>
                <FormControlLabel control={<Switch checked={draft.notifyOnFailure} onChange={(event) => onDraftChange({notifyOnFailure: event.target.checked})}/>} label="Notify on health check failure"/>
            </Stack>
        ))}

        <Dialog open={uiState.showFormatDialog} onClose={() => onUiChange({showFormatDialog: false})}>
            <DialogTitle>Select Key Format</DialogTitle>
            <DialogContent>
                <Typography variant="body2" sx={{mb: 2, mt: 1}}>
                    Choose the format for the generated ED25519 private key.
                </Typography>
                <RadioGroup value={uiState.keyFormat} onChange={(_, value) => onUiChange({keyFormat: value as any})}>
                    <FormControlLabel value="openssh" control={<Radio/>} label="OpenSSH (Recommended)"/>
                    <FormControlLabel value="pkcs8" control={<Radio/>} label="PEM (PKCS#8)"/>
                </RadioGroup>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => onUiChange({showFormatDialog: false})}>Cancel</Button>
                <Button onClick={onGenerateKeyPair} variant="contained">Generate</Button>
            </DialogActions>
        </Dialog>
    </>
);
