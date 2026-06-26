import type {Dispatch, ReactNode, SetStateAction} from 'react'
import {Alert, Box, Button, Card, CardActions, CardContent, CardHeader, Divider, FormControl, FormControlLabel, FormHelperText, IconButton, InputAdornment, InputLabel, MenuItem, Select, Stack, Switch, TextField, Tooltip, Typography} from "@mui/material";
import {HStack, VStack} from '@dsherwin/mui-kit';
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import Visibility from "@mui/icons-material/Visibility";
import VisibilityOff from "@mui/icons-material/VisibilityOff";
import CheckCircleOutlinedIcon from "@mui/icons-material/CheckCircleOutlined";
import WarningAmberIcon from "@mui/icons-material/WarningAmber";
import type {DbConnectionDraft, DbDriver} from '../types.ts';
import type {AgentOption} from '../../Agents/api.ts';
import {DEFAULT_DB_CONNECTIONS} from './databaseConnectionForm.ts';

export interface DatabaseConnectionEditorProps {
    title: string
    infoTooltip: string
    draft: DbConnectionDraft
    setDraft: Dispatch<SetStateAction<DbConnectionDraft>>
    errors: Record<string, string>
    agents: AgentOption[]
    agentsLoading?: boolean
    showPassword: boolean
    setShowPassword: (show: boolean) => void
    testing?: boolean
    loading?: boolean
    testResult: null | { ok: boolean; message: string }
    onTest: () => void
    onCancel: () => void
    onSave: () => void
    saveLabel: string
    previewDsn: string
    headerAction?: ReactNode
    dangerZone?: ReactNode
    onNameInput?: () => void
}

export const DatabaseConnectionEditor = ({
    title,
    infoTooltip,
    draft,
    setDraft,
    errors,
    agents,
    agentsLoading = false,
    showPassword,
    setShowPassword,
    testing = false,
    loading = false,
    testResult,
    onTest,
    onCancel,
    onSave,
    saveLabel,
    previewDsn,
    headerAction,
    dangerZone,
    onNameInput,
}: DatabaseConnectionEditorProps) => {
    const updateDraft = (patch: Partial<DbConnectionDraft>) => {
        setDraft((current) => ({...current, ...patch}))
    }

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
                        <Alert
                            icon={testResult.ok ? <CheckCircleOutlinedIcon/> : <WarningAmberIcon/>}
                            severity={testResult.ok ? 'success' : 'error'}
                            sx={{borderRadius: 3}}
                        >
                            {testResult.message}
                        </Alert>
                    )}
                    <Card variant="outlined" sx={{borderRadius: 3, bgcolor: 'action.hover'}}>
                        <CardActions sx={{p: 2, justifyContent: 'space-between', flexWrap: 'wrap', gap: 2}}>
                            <Typography variant="body2" sx={{color: "text.secondary"}}>
                                Tip: Test changes before saving.
                            </Typography>
                            <HStack spacing={1}>
                                <Button variant="outlined" onClick={onTest} disabled={testing || loading}>
                                    {testing ? 'Testing…' : 'Test connection'}
                                </Button>
                                <Button variant="outlined" onClick={onCancel}>Cancel</Button>
                                <Button variant="contained" onClick={onSave} disabled={loading}>
                                    {saveLabel}
                                </Button>
                            </HStack>
                        </CardActions>
                    </Card>
                </VStack>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="General Information"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Stack spacing={3}>
                            <Stack direction={{xs: "column", md: "row"}} spacing={2}>
                                <TextField
                                    label="Connection name"
                                    value={draft.name}
                                    onChange={(event) => {
                                        onNameInput?.()
                                        updateDraft({name: event.target.value})
                                    }}
                                    fullWidth
                                    error={!!errors.name}
                                    helperText={errors.name || "Example: Chronix Primary Postgres"}
                                />

                                <FormControl fullWidth>
                                    <InputLabel id="driver-label">Database type</InputLabel>
                                    <Select
                                        labelId="driver-label"
                                        label="Database type"
                                        value={draft.driver}
                                        onChange={(event) => updateDraft({driver: event.target.value as DbDriver})}
                                    >
                                        <MenuItem value="postgres">PostgreSQL</MenuItem>
                                        <MenuItem value="mysql">MySQL / MariaDB</MenuItem>
                                        <MenuItem value="sqlite">SQLite</MenuItem>
                                        <MenuItem value="mssql">SQL Server</MenuItem>
                                        <MenuItem value="oracle">Oracle</MenuItem>
                                        <MenuItem value="snowflake">Snowflake</MenuItem>
                                    </Select>
                                </FormControl>
                            </Stack>

                            <TextField
                                label="Description (optional)"
                                placeholder="What is this connection used for?"
                                value={draft.description || ""}
                                onChange={(event) => updateDraft({description: event.target.value})}
                                fullWidth
                                multiline
                                minRows={3}
                                maxRows={6}
                            />

                            <FormControl fullWidth disabled={agentsLoading}>
                                <InputLabel id="agent-label">Agent (optional)</InputLabel>
                                <Select
                                    labelId="agent-label"
                                    label="Agent (optional)"
                                    value={draft.agentUuid || ""}
                                    onChange={(event) => {
                                        const nextValue = event.target.value as string
                                        updateDraft({agentUuid: nextValue ? nextValue : undefined})
                                    }}
                                >
                                    <MenuItem value=""><em>Server (no agent)</em></MenuItem>
                                    {agents.map((agent) => (
                                        <MenuItem key={agent.uuid} value={agent.uuid}>{agent.name}</MenuItem>
                                    ))}
                                </Select>
                                <FormHelperText>
                                    Choose an Agent to run queries for this connection. Leave empty to run on the server.
                                </FormHelperText>
                            </FormControl>
                        </Stack>
                    </CardContent>
                </Card>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="Database Settings"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Stack spacing={3}>
                            {draft.driver === "sqlite" ? (
                                <TextField
                                    label="SQLite file path"
                                    placeholder="/var/chronix/chronix.db"
                                    value={draft.filePath || ""}
                                    onChange={(event) => updateDraft({filePath: event.target.value})}
                                    fullWidth
                                    error={!!errors.filePath}
                                    helperText={errors.filePath || "Absolute path recommended. Will be created if absent."}
                                />
                            ) : (
                                <>
                                    <Stack direction={{xs: "column", md: "row"}} spacing={2}>
                                        <TextField
                                            label={draft.driver === "snowflake" ? "Account" : "Host"}
                                            placeholder={draft.driver === "snowflake" ? "xy12345.us-east-1" : "db.example.com"}
                                            value={draft.host || ""}
                                            onChange={(event) => updateDraft({host: event.target.value})}
                                            fullWidth
                                            error={!!errors.host}
                                            helperText={errors.host}
                                        />
                                        {draft.driver !== "snowflake" && (
                                            <TextField
                                                label="Port"
                                                placeholder={DEFAULT_DB_CONNECTIONS[draft.driver].port}
                                                value={draft.port || ""}
                                                onChange={(event) => updateDraft({port: event.target.value})}
                                                sx={{maxWidth: 180}}
                                                error={!!errors.port}
                                                helperText={errors.port}
                                                slotProps={{
                                                    htmlInput: {
                                                        inputMode: "numeric",
                                                        pattern: "[0-9]*",
                                                    }
                                                }}
                                            />
                                        )}
                                    </Stack>

                                    <Stack direction={{xs: "column", md: "row"}} spacing={2}>
                                        <TextField
                                            label={draft.driver === "mysql" ? "Schema (optional)" : draft.driver === "snowflake" ? "Database (optional)" : "Database"}
                                            placeholder={draft.driver === "postgres" ? "postgres" : draft.driver === "mssql" ? "master" : ""}
                                            value={draft.database || ""}
                                            onChange={(event) => updateDraft({database: event.target.value})}
                                            fullWidth
                                            error={!!errors.database}
                                            helperText={errors.database}
                                        />
                                        <TextField
                                            label="Username"
                                            value={draft.username || ""}
                                            onChange={(event) => updateDraft({username: event.target.value})}
                                            fullWidth
                                            error={!!errors.username}
                                            helperText={errors.username}
                                        />
                                        <TextField
                                            label={draft.hasPassword ? 'Password (set — leave blank to keep)' : 'Password'}
                                            type={showPassword ? "text" : "password"}
                                            value={draft.password || ""}
                                            onChange={(event) => updateDraft({password: event.target.value})}
                                            fullWidth
                                            slotProps={{
                                                input: {
                                                    endAdornment: (
                                                        <InputAdornment position="end">
                                                            <IconButton onClick={() => setShowPassword(!showPassword)} edge="end" aria-label="toggle password visibility">
                                                                {showPassword ? <VisibilityOff/> : <Visibility/>}
                                                            </IconButton>
                                                        </InputAdornment>
                                                    )
                                                }
                                            }}
                                        />
                                    </Stack>

                                    {(draft.driver === "postgres" || draft.driver === "mysql") && (
                                        <Stack direction={{xs: "column", md: "row"}} spacing={2} sx={{alignItems: "center"}}>
                                            <FormControlLabel
                                                control={<Switch checked={!!draft.sslEnabled} onChange={(event) => updateDraft({sslEnabled: event.target.checked})}/>}
                                                label="Use SSL/TLS"
                                            />
                                            {draft.driver === "postgres" && (
                                                <FormControl sx={{minWidth: 220}}>
                                                    <InputLabel id="sslmode-label">SSL mode</InputLabel>
                                                    <Select
                                                        labelId="sslmode-label"
                                                        label="SSL mode"
                                                        value={draft.sslMode || "prefer"}
                                                        onChange={(event) => updateDraft({sslMode: event.target.value as DbConnectionDraft["sslMode"]})}
                                                        disabled={!draft.sslEnabled}
                                                    >
                                                        <MenuItem value="prefer">prefer</MenuItem>
                                                        <MenuItem value="require">require</MenuItem>
                                                        <MenuItem value="verify-ca">verify-ca</MenuItem>
                                                        <MenuItem value="verify-full">verify-full</MenuItem>
                                                    </Select>
                                                    <FormHelperText>
                                                        Postgres-specific. Turn off "Use SSL/TLS" to disable SSL entirely.
                                                    </FormHelperText>
                                                </FormControl>
                                            )}
                                        </Stack>
                                    )}

                                    {draft.driver === "mssql" && (
                                        <FormControlLabel
                                            control={
                                                <Switch
                                                    checked={!!draft.trustServerCertificate}
                                                    onChange={(event) => updateDraft({trustServerCertificate: event.target.checked})}
                                                />
                                            }
                                            label="Trust server certificate (dev only)"
                                        />
                                    )}
                                </>
                            )}
                        </Stack>
                    </CardContent>
                </Card>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="Advanced Options"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Stack spacing={2}>
                            <TextField
                                label="Extra parameters (k=v&k2=v2)"
                                placeholder={
                                    draft.driver === "postgres"
                                        ? "application_name=chronix&search_path=public"
                                        : draft.driver === "mysql"
                                            ? "tls=true&timeout=2s"
                                            : draft.driver === "mssql"
                                                ? "encrypt=true&connection+timeout=5"
                                                : ""
                                }
                                value={draft.params || ""}
                                onChange={(event) => updateDraft({params: event.target.value})}
                                fullWidth
                            />
                            <Divider sx={{borderColor: (theme) => theme.palette.mode === 'light' ? 'rgba(25, 118, 210, 0.2)' : 'rgba(25, 118, 210, 0.4)'}}/>
                            <Stack direction={{xs: "column", md: "row"}} spacing={2} sx={{alignItems: "center"}}>
                                <FormControlLabel
                                    control={<Switch checked={!!draft.autoCheckEnabled} onChange={(event) => updateDraft({autoCheckEnabled: event.target.checked, autoCheckSeconds: event.target.checked ? (draft.autoCheckSeconds || 3600) : draft.autoCheckSeconds})}/>}
                                    label="Automatic health check"
                                />
                                <FormControl sx={{minWidth: 220}} disabled={!draft.autoCheckEnabled}>
                                    <InputLabel id="autocheck-every-label">Check every</InputLabel>
                                    <Select
                                        labelId="autocheck-every-label"
                                        label="Check every"
                                        value={String(draft.autoCheckSeconds || 3600)}
                                        onChange={(event) => updateDraft({autoCheckSeconds: Number(event.target.value)})}
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
                        </Stack>
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
                            <Typography variant="body2" sx={{color: "text.secondary"}}>
                                Configure specific destinations for alerts related to this connection's health.
                            </Typography>
                            <TextField
                                label="Alert Emails"
                                placeholder="email1@example.com, email2@example.com"
                                value={draft.alertEmails || ""}
                                onChange={(event) => updateDraft({alertEmails: event.target.value})}
                                fullWidth
                                helperText="Comma-separated list of email addresses. If empty, system defaults are used."
                            />
                            <TextField
                                label="Alert Phones (SMS)"
                                placeholder="+15550001111"
                                value={draft.alertPhones || ""}
                                onChange={(event) => updateDraft({alertPhones: event.target.value})}
                                fullWidth
                                helperText="Comma-separated list of E.164 phone numbers. If empty, system defaults are used."
                            />
                            <FormControlLabel
                                control={
                                    <Switch
                                        checked={!!draft.notifyOnFailure}
                                        onChange={(event) => updateDraft({notifyOnFailure: event.target.checked})}
                                    />
                                }
                                label="Notify on health check failure"
                            />
                        </Stack>
                    </CardContent>
                </Card>

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardHeader
                        title="Connection String Preview"
                        sx={{bgcolor: 'action.hover', px: 1.75, py: 1.1}}
                        slotProps={{title: {sx: {fontSize: '1.2rem'}}}}
                    />
                    <Divider/>
                    <CardContent>
                        <Box sx={{p: 1, bgcolor: "action.hover", borderRadius: 1, fontFamily: "monospace", overflowX: "auto"}}>
                            {previewDsn}
                        </Box>
                        <FormHelperText>
                            Passwords are never stored in the preview. Actual driver wiring happens server-side.
                        </FormHelperText>
                    </CardContent>
                </Card>

                {dangerZone}
            </VStack>
        </Box>
    )
}
