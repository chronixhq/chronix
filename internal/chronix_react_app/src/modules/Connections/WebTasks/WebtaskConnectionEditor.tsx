import type {Dispatch, ReactNode, SetStateAction} from 'react'
import {Alert, Box, Button, Card, CardActions, CardContent, CardHeader, Divider, FormControl, FormControlLabel, FormHelperText, IconButton, InputAdornment, InputLabel, MenuItem, Select, Stack, Switch, TextField, Tooltip, Typography} from "@mui/material";
import {HStack, VStack} from '@dsherwin/mui-kit';
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import Visibility from "@mui/icons-material/Visibility";
import VisibilityOff from "@mui/icons-material/VisibilityOff";
import type {AgentOption} from '../../Agents/api.ts';
import type {WebTaskAuthType, WebtaskConnection} from '../types.ts';

const REDACTED_VALUE = '<redacted>'

type EditorState = Partial<WebtaskConnection>

function secretInputProps(value?: string, placeholder = '') {
    const isRedacted = value === REDACTED_VALUE
    return {
        isRedacted,
        placeholder: isRedacted ? '••••••••' : placeholder,
        value: isRedacted ? '' : (value || ''),
        required: !isRedacted,
    }
}

export interface WebtaskConnectionEditorProps {
    title: string
    infoTooltip: string
    draft: EditorState
    setDraft: Dispatch<SetStateAction<EditorState>>
    errors: Record<string, string>
    agents: AgentOption[]
    agentsLoading?: boolean
    showPassword: boolean
    setShowPassword: (show: boolean) => void
    loading?: boolean
    testing?: boolean
    testResult: null | { ok: boolean; message: string }
    onDismissTestResult: () => void
    onTest: () => void
    onCancel: () => void
    onSave: () => void
    saveLabel: string
    headerAction?: ReactNode
    dangerZone?: ReactNode
    apiError?: string
}

export const WebtaskConnectionEditor = ({
    title,
    infoTooltip,
    draft,
    setDraft,
    errors,
    agents,
    agentsLoading = false,
    showPassword,
    setShowPassword,
    loading = false,
    testing = false,
    testResult,
    onDismissTestResult,
    onTest,
    onCancel,
    onSave,
    saveLabel,
    headerAction,
    dangerZone,
    apiError,
}: WebtaskConnectionEditorProps) => {
    const authType = draft.authType || 'none'
    const authConfig = draft.authConfig || {}

    const updateDraft = (patch: Partial<EditorState>) => {
        setDraft((current) => ({...current, ...patch}))
    }

    const updateAuthConfig = (field: string, value: any) => {
        setDraft((current) => ({
            ...current,
            authConfig: {...(current.authConfig || {}), [field]: value}
        }))
    }

    const bearerProps = secretInputProps(authConfig.token, 'eyJhbG...')
    const basicPasswordProps = secretInputProps(authConfig.password)
    const headerValueProps = secretInputProps(authConfig.header_value)

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={2} sx={{maxWidth: 800, width: '100%', mx: 'auto'}}>
                <HStack alignItems="center" spacing={1} sx={{justifyContent: 'space-between', flexWrap: 'wrap'}}>
                    <HStack alignItems="center" spacing={1}>
                        <Typography variant="h5">{title}</Typography>
                        <Tooltip title={infoTooltip}>
                            <InfoOutlinedIcon fontSize="small"/>
                        </Tooltip>
                    </HStack>
                    {headerAction}
                </HStack>
                <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>

                <VStack spacing={2} sx={{width: '100%'}}>
                    {testResult && (
                        <Alert severity={testResult.ok ? "success" : "error"} onClose={onDismissTestResult}>
                            {testResult.message}
                        </Alert>
                    )}
                    <Card variant="outlined" sx={{borderRadius: 3, bgcolor: 'action.hover'}}>
                        <CardActions sx={{p: 2, justifyContent: 'space-between', flexWrap: 'wrap', gap: 2}}>
                            <Typography variant="body2" sx={{
                                color: "text.secondary"
                            }}>Tip: Test changes before saving.</Typography>
                            <HStack spacing={1}>
                                <Button variant="outlined" onClick={onTest} disabled={testing || loading}>{testing ? 'Testing…' : 'Test connection'}</Button>
                                <Button variant="outlined" onClick={onCancel}>Cancel</Button>
                                <Button variant="contained" onClick={onSave} disabled={testing || loading}>{saveLabel}</Button>
                            </HStack>
                        </CardActions>
                    </Card>
                </VStack>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="General Settings"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Stack spacing={3}>
                            <TextField
                                label="Connection Name"
                                placeholder="e.g. Production API"
                                value={draft.name || ""}
                                onChange={e => updateDraft({name: e.target.value})}
                                error={!!errors.name}
                                helperText={errors.name}
                                fullWidth
                                required
                            />
                            <TextField
                                label="Description (optional)"
                                placeholder="Optional notes about this connection"
                                value={draft.description || ""}
                                onChange={e => updateDraft({description: e.target.value})}
                                multiline
                                rows={2}
                                fullWidth
                            />
                            <TextField
                                label="Base URL"
                                placeholder="https://api.example.com"
                                value={draft.baseUrl || ""}
                                onChange={e => updateDraft({baseUrl: e.target.value})}
                                fullWidth
                                helperText="Optional. If set, Web Task steps can use relative URLs."
                            />
                            <FormControl fullWidth>
                                <InputLabel>Route via Agent (optional)</InputLabel>
                                <Select
                                    value={draft.agentUuid || ""}
                                    onChange={e => updateDraft({agentUuid: e.target.value || null})}
                                    label="Route via Agent (optional)"
                                >
                                    <MenuItem value=""><em>Local (Direct from Server)</em></MenuItem>
                                    {agents.map((agent) => <MenuItem key={agent.uuid} value={agent.uuid}>{agent.name}</MenuItem>)}
                                    {agentsLoading && <MenuItem disabled>Loading agents...</MenuItem>}
                                </Select>
                                <FormHelperText>Execute requests from a remote network position.</FormHelperText>
                            </FormControl>
                        </Stack>
                    </CardContent>
                </Card>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="Authentication"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Stack spacing={3}>
                            <FormControl fullWidth>
                                <InputLabel>Auth Type</InputLabel>
                                <Select
                                    value={authType}
                                    onChange={e => updateDraft({authType: e.target.value as WebTaskAuthType, authConfig: {}})}
                                    label="Auth Type"
                                >
                                    <MenuItem value="none">None</MenuItem>
                                    <MenuItem value="basic">Basic Auth (Username/Password)</MenuItem>
                                    <MenuItem value="bearer">Bearer Token</MenuItem>
                                    <MenuItem value="header">Custom API Key Header</MenuItem>
                                </Select>
                            </FormControl>

                            {authType === "basic" && (
                                <HStack spacing={2}>
                                    <TextField
                                        label="Username"
                                        value={authConfig.username || ""}
                                        onChange={e => updateAuthConfig('username', e.target.value)}
                                        error={!!errors.username}
                                        helperText={errors.username}
                                        fullWidth
                                        required
                                    />
                                    <TextField
                                        label="Password"
                                        type={showPassword ? "text" : "password"}
                                        placeholder={basicPasswordProps.placeholder}
                                        value={basicPasswordProps.value}
                                        onChange={e => updateAuthConfig('password', e.target.value)}
                                        error={!!errors.password}
                                        helperText={errors.password}
                                        fullWidth
                                        required={basicPasswordProps.required}
                                        slotProps={{
                                            input: {
                                                endAdornment: (
                                                    <InputAdornment position="end">
                                                        <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                                                            {showPassword ? <VisibilityOff/> : <Visibility/>}
                                                        </IconButton>
                                                    </InputAdornment>
                                                )
                                            }
                                        }}
                                    />
                                </HStack>
                            )}

                            {authType === "bearer" && (
                                <TextField
                                    label="Bearer Token"
                                    type={showPassword ? "text" : "password"}
                                    placeholder={bearerProps.placeholder}
                                    value={bearerProps.value}
                                    onChange={e => updateAuthConfig('token', e.target.value)}
                                    error={!!errors.token}
                                    helperText={errors.token}
                                    fullWidth
                                    required={bearerProps.required}
                                    slotProps={{
                                        input: {
                                            endAdornment: (
                                                <InputAdornment position="end">
                                                    <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                                                        {showPassword ? <VisibilityOff/> : <Visibility/>}
                                                    </IconButton>
                                                </InputAdornment>
                                            )
                                        }
                                    }}
                                />
                            )}

                            {authType === "header" && (
                                <HStack spacing={2}>
                                    <TextField
                                        label="Header Name"
                                        placeholder="X-API-Key"
                                        value={authConfig.header_name || ""}
                                        onChange={e => updateAuthConfig('header_name', e.target.value)}
                                        error={!!errors.header_name}
                                        helperText={errors.header_name}
                                        fullWidth
                                        required
                                    />
                                    <TextField
                                        label="Header Value"
                                        type={showPassword ? "text" : "password"}
                                        placeholder={headerValueProps.placeholder}
                                        value={headerValueProps.value}
                                        onChange={e => updateAuthConfig('header_value', e.target.value)}
                                        error={!!errors.header_value}
                                        helperText={errors.header_value}
                                        fullWidth
                                        required={headerValueProps.required}
                                        slotProps={{
                                            input: {
                                                endAdornment: (
                                                    <InputAdornment position="end">
                                                        <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                                                            {showPassword ? <VisibilityOff/> : <Visibility/>}
                                                        </IconButton>
                                                    </InputAdornment>
                                                )
                                            }
                                        }}
                                    />
                                </HStack>
                            )}
                        </Stack>
                    </CardContent>
                </Card>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="Health Check"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Stack direction={{xs: 'column', md: 'row'}} spacing={2} sx={{
                            alignItems: "center"
                        }}>
                            <FormControlLabel
                                control={<Switch checked={!!draft.autoCheckEnabled} onChange={e => updateDraft({autoCheckEnabled: e.target.checked})}/>}
                                label="Automatic health check"
                            />
                            <FormControl sx={{minWidth: 220}} disabled={!draft.autoCheckEnabled}>
                                <InputLabel id="autocheck-every-label">Check every</InputLabel>
                                <Select
                                    labelId="autocheck-every-label"
                                    label="Check every"
                                    value={String(draft.autoCheckSeconds || 300)}
                                    onChange={(e) => updateDraft({autoCheckSeconds: Number(e.target.value)})}
                                >
                                    <MenuItem value={60}>1 minute</MenuItem>
                                    <MenuItem value={300}>5 minutes</MenuItem>
                                    <MenuItem value={900}>15 minutes</MenuItem>
                                    <MenuItem value={1800}>30 minutes</MenuItem>
                                    <MenuItem value={3600}>1 hour</MenuItem>
                                    <MenuItem value={21600}>6 hours</MenuItem>
                                    <MenuItem value={43200}>12 hours</MenuItem>
                                    <MenuItem value={86400}>1 day</MenuItem>
                                </Select>
                                <FormHelperText>Minimum 1 minute; maximum 1 day.</FormHelperText>
                            </FormControl>
                        </Stack>
                        <Typography
                            variant="caption"
                            sx={{
                                color: "text.secondary",
                                mt: 1,
                                display: 'block'
                            }}>
                            When enabled, the server will periodically check the connectivity by sending a HEAD or GET request to the Base URL (if provided).
                        </Typography>
                    </CardContent>
                </Card>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="Alerts"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Stack spacing={3}>
                            <Typography variant="body2" sx={{
                                color: "text.secondary"
                            }}>
                                Configure specific destinations for alerts related to this connection&apos;s health.
                            </Typography>
                            <TextField
                                label="Alert Emails"
                                placeholder="email1@example.com, email2@example.com"
                                value={draft.alertEmails || ""}
                                onChange={(e) => updateDraft({alertEmails: e.target.value})}
                                fullWidth
                                helperText="Comma-separated list of email addresses. If empty, system defaults are used."
                            />
                            <TextField
                                label="Alert Phones (SMS)"
                                placeholder="+15550001111"
                                value={draft.alertPhones || ""}
                                onChange={(e) => updateDraft({alertPhones: e.target.value})}
                                fullWidth
                                helperText="Comma-separated list of E.164 phone numbers. If empty, system defaults are used."
                            />
                            <FormControlLabel
                                control={
                                    <Switch
                                        checked={!!draft.notifyOnFailure}
                                        onChange={(e) => updateDraft({notifyOnFailure: e.target.checked})}
                                    />
                                }
                                label="Notify on health check failure"
                            />
                        </Stack>
                    </CardContent>
                </Card>

                {dangerZone}

                {apiError && <Alert severity="error">{apiError}</Alert>}
            </VStack>
        </Box>
    );
}
