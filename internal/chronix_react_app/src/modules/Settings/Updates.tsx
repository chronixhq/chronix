import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Backdrop,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Divider,
  FormControl,
  FormControlLabel,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  Snackbar,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { HStack, useMuiPrompts, VStack } from "@dsherwin/mui-kit";
import { SystemUpdate } from "@mui/icons-material";
import { SectionHelp } from "../../main/SectionHelp";
import { HELP_SECTIONS } from "../../main/appShellManifest.ts";
import { useUnsavedChanges } from "../../lib/useUnsavedChanges.ts";
import {
  applyUpdate,
  checkForUpdates,
  fetchUpdaterAgents,
  fetchUpdaterStatus,
  saveAgentUpdaterSettings,
  saveAppUpdaterSettings,
  updateAgentNow,
} from "./api.ts";
import type {
  UpdateAgentInfo,
  UpdaterStatus,
  UpdaterVersionInfo,
} from "./types.ts";
import { waitForServerVersion } from "./updatePolling.ts";

export const UpdatesPage = () => {
  const { confirmPrompt } = useMuiPrompts();
  const [status, setStatus] = useState<UpdaterStatus | null>(null);
  const [agents, setAgents] = useState<UpdateAgentInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [serverUpdateInProgress, setServerUpdateInProgress] = useState(false);
  const [serverUpdateTargetVersion, setServerUpdateTargetVersion] =
    useState("");
  const [agentBusy, setAgentBusy] = useState<Record<string, boolean>>({});
  const allowServerUpdateReloadRef = useRef(false);
  const [snack, setSnack] = useState<{
    open: boolean;
    message: string;
    severity: "success" | "error" | "info";
  }>({
    open: false,
    message: "",
    severity: "info",
  });
  useUnsavedChanges(serverUpdateInProgress, allowServerUpdateReloadRef);

  const fetchStatus = async () => {
    try {
      const [res, agentRes] = await Promise.all([
        fetchUpdaterStatus(),
        fetchUpdaterAgents(),
      ]);
      setStatus(res);
      setAgents(agentRes);
    } catch (e) {
      console.error(e);
      setSnack({
        open: true,
        message: "Failed to load updater status.",
        severity: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
  }, []);

  const targetAgentVersion =
    status?.availableVersion?.["chronix-agent"]?.version || "";
  const agentVersionInfo = status?.availableVersion?.["chronix-agent"];
  const outdatedAgents = targetAgentVersion
    ? agents.filter((agent) => agent.version !== targetAgentVersion)
    : [];
  const onlineOutdatedAgents = outdatedAgents.filter((agent) => agent.online);
  const offlineOutdatedAgents = outdatedAgents.filter((agent) => !agent.online);
  const anyAgentBusy = Object.values(agentBusy).some(Boolean);

  const onCheck = async () => {
    setBusy(true);
    try {
      const res = await checkForUpdates();
      if (res.available) {
        const manifest =
          typeof res.manifest === "object" && res.manifest
            ? (res.manifest as Record<string, unknown>)
            : {};
        const server =
          typeof manifest.server === "object" && manifest.server
            ? (manifest.server as Record<string, unknown>)
            : {};
        setSnack({
          open: true,
          message: `New app version available: ${String(server.version || "unknown")}`,
          severity: "info",
        });
      } else {
        const manifest =
          typeof res.manifest === "object" && res.manifest
            ? (res.manifest as Record<string, unknown>)
            : {};
        const agentInfo =
          typeof manifest["chronix-agent"] === "object" &&
          manifest["chronix-agent"]
            ? (manifest["chronix-agent"] as Record<string, unknown>)
            : null;
        const latestAgentVersion =
          typeof agentInfo?.version === "string" ? agentInfo.version : "";
        if (
          latestAgentVersion &&
          agents.some((a) => a.version !== latestAgentVersion)
        ) {
          setSnack({
            open: true,
            message: `New agent version available: ${latestAgentVersion}`,
            severity: "info",
          });
        } else {
          setSnack({
            open: true,
            message: "Chronix is up to date.",
            severity: "success",
          });
        }
      }
      fetchStatus();
    } catch (e) {
      console.error(e);
      setSnack({
        open: true,
        message: "Failed to check for updates.",
        severity: "error",
      });
    } finally {
      setBusy(false);
    }
  };

  const onApply = async () => {
    const targetVersion = status?.availableVersion?.server?.version;
    if (!targetVersion) {
      setSnack({
        open: true,
        message: "No app update is currently available.",
        severity: "error",
      });
      return;
    }

    const ok = await confirmPrompt({
      title: "Apply Update",
      message:
        "Are you sure you want to apply the update and restart the server?",
      buttonText: "Apply and Restart",
      cancelButtonText: "Cancel",
    });
    if (!ok) return;
    setBusy(true);
    setServerUpdateInProgress(true);
    setServerUpdateTargetVersion(targetVersion);
    allowServerUpdateReloadRef.current = false;
    setSnack((prev) => ({ ...prev, open: false }));
    try {
      await applyUpdate();

      await waitForServerVersion(
        () => fetchUpdaterStatus({ fresh: true }),
        targetVersion,
      );

      allowServerUpdateReloadRef.current = true;
      window.location.reload();
    } catch (e) {
      console.error(e);
      allowServerUpdateReloadRef.current = false;
      setServerUpdateInProgress(false);
      setServerUpdateTargetVersion("");
      const message =
        e instanceof Error && e.message ? e.message : "Failed to apply update.";
      setSnack({ open: true, message, severity: "error" });
      setBusy(false);
    }
  };

  const onSaveAppSettings = async () => {
    if (!status) return;
    setBusy(true);
    try {
      await saveAppUpdaterSettings({
        enabled: status.enabled,
        mode: status.mode,
        windowStart: status.windowStart,
      });
      setSnack({
        open: true,
        message: "App updater settings saved.",
        severity: "success",
      });
    } catch (e) {
      console.error(e);
      setSnack({
        open: true,
        message: "Failed to save app updater settings.",
        severity: "error",
      });
    } finally {
      setBusy(false);
    }
  };

  const onSaveAgentSettings = async () => {
    if (!status) return;
    setBusy(true);
    try {
      await saveAgentUpdaterSettings({
        agentEnabled: status.agentEnabled,
        agentMode: status.agentMode,
        agentWindowStart: status.agentWindowStart,
      });
      setSnack({
        open: true,
        message: "Agent updater settings saved.",
        severity: "success",
      });
    } catch (e) {
      console.error(e);
      setSnack({
        open: true,
        message: "Failed to save agent updater settings.",
        severity: "error",
      });
    } finally {
      setBusy(false);
    }
  };

  const startAgentUpdate = async (
    agent: UpdateAgentInfo,
    options?: {
      showInitiatedSnack?: boolean;
      showSuccessSnack?: boolean;
      showTimeoutSnack?: boolean;
    },
  ) => {
    if (!targetAgentVersion) return false;
    setAgentBusy((s) => ({ ...s, [agent.uuid]: true }));
    try {
      await updateAgentNow(agent.uuid);
      if (options?.showInitiatedSnack !== false) {
        setSnack({
          open: true,
          message: `Update initiated for ${agent.name}`,
          severity: "success",
        });
      }

      // Poll for agent to come back online with new version
      let attempts = 0;
      const maxAttempts = 60; // 2 minutes with 2s interval
      const poll = async () => {
        try {
          const agentRes = await fetchUpdaterAgents();
          const updatedAgent = agentRes.find((a) => a.uuid === agent.uuid);

          if (
            updatedAgent &&
            updatedAgent.online &&
            updatedAgent.version === targetAgentVersion
          ) {
            if (options?.showSuccessSnack !== false) {
              setSnack({
                open: true,
                message: `Agent ${agent.name} updated successfully to ${targetAgentVersion}`,
                severity: "success",
              });
            }
            setAgentBusy((s) => ({ ...s, [agent.uuid]: false }));
            // Refresh everything to update the page state
            void fetchStatus();
          } else {
            attempts++;
            if (attempts < maxAttempts) {
              setTimeout(poll, 2000);
            } else {
              if (options?.showTimeoutSnack !== false) {
                setSnack({
                  open: true,
                  message: `Agent ${agent.name} update verification timed out.`,
                  severity: "error",
                });
              }
              setAgentBusy((s) => ({ ...s, [agent.uuid]: false }));
              void fetchStatus();
            }
          }
        } catch {
          // Ignore errors during polling (e.g. transient network issues)
          attempts++;
          if (attempts < maxAttempts) {
            setTimeout(poll, 2000);
          } else {
            setAgentBusy((s) => ({ ...s, [agent.uuid]: false }));
            void fetchStatus();
          }
        }
      };
      // Start polling after a short delay to give the agent time to disconnect
      setTimeout(poll, 5000);
      return true;
    } catch (e) {
      console.error(e);
      const message = e instanceof Error && e.message ? e.message : agent.name;
      setSnack({
        open: true,
        message: `Failed to update agent: ${message}`,
        severity: "error",
      });
      setAgentBusy((s) => ({ ...s, [agent.uuid]: false }));
      return false;
    }
  };

  const onUpdateAgent = async (agent: UpdateAgentInfo) => {
    if (!agent.online) {
      setSnack({
        open: true,
        message: `Agent ${agent.name} is offline and cannot be updated right now.`,
        severity: "info",
      });
      return;
    }

    const ok = await confirmPrompt({
      title: "Update Agent",
      message: `Are you sure you want to update agent '${agent.name}'? The agent will restart.`,
      buttonText: "Update Agent",
      cancelButtonText: "Cancel",
    });
    if (!ok) return;

    await startAgentUpdate(agent);
  };

  const onUpdateAllAgents = async () => {
    if (!targetAgentVersion || onlineOutdatedAgents.length === 0) {
      setSnack({
        open: true,
        message: "No online agents are ready for a bulk update.",
        severity: "info",
      });
      return;
    }

    const summaryParts = [
      `This will start agent updates for ${onlineOutdatedAgents.length} online ${onlineOutdatedAgents.length === 1 ? "agent" : "agents"} to version ${targetAgentVersion}.`,
    ];
    if (offlineOutdatedAgents.length > 0) {
      summaryParts.push(
        `${offlineOutdatedAgents.length} offline ${offlineOutdatedAgents.length === 1 ? "agent will" : "agents will"} be skipped.`,
      );
    }

    const ok = await confirmPrompt({
      title: "Update All Agents",
      message: summaryParts.join(" "),
      buttonText: "Update All",
      cancelButtonText: "Cancel",
    });
    if (!ok) return;

    const results = await Promise.all(
      onlineOutdatedAgents.map((agent) =>
        startAgentUpdate(agent, {
          showInitiatedSnack: false,
          showSuccessSnack: false,
          showTimeoutSnack: false,
        }),
      ),
    );

    const startedCount = results.filter(Boolean).length;
    const failedCount = results.length - startedCount;
    const messageParts = [];
    messageParts.push(
      `Started updates for ${startedCount} ${startedCount === 1 ? "agent" : "agents"}.`,
    );
    if (offlineOutdatedAgents.length > 0) {
      messageParts.push(
        `Skipped ${offlineOutdatedAgents.length} offline ${offlineOutdatedAgents.length === 1 ? "agent" : "agents"}.`,
      );
    }
    if (failedCount > 0) {
      messageParts.push(
        `${failedCount} ${failedCount === 1 ? "update failed to start" : "updates failed to start"}.`,
      );
    }

    setSnack({
      open: true,
      message: messageParts.join(" "),
      severity: failedCount > 0 ? "error" : "success",
    });
  };

  const isServerOutdated = () => {
    if (!status?.availableVersion?.server) return false;
    const current = status.currentVersion;
    const latest = status.availableVersion.server.version;
    if (current === "" || !latest) return false;
    return current !== latest;
  };

  if (loading)
    return <Typography sx={{ p: 4 }}>Loading updater status...</Typography>;

  const UpdateAlert = ({
    title,
    info,
    onApply,
    busy,
    buttonText,
  }: {
    title: string;
    info: UpdaterVersionInfo;
    onApply: () => void;
    busy: boolean;
    buttonText: string;
  }) => (
    <Alert
      severity="warning"
      variant="outlined"
      sx={{
        mt: 1,
        borderColor: "warning.main",
        backgroundColor: "rgba(237, 108, 2, 0.05)",
        "& .MuiAlert-message": { width: "100%" },
      }}
    >
      <VStack spacing={1} sx={{ width: "100%" }}>
        <HStack justifyContent="space-between" alignItems="center">
          <Typography
            variant="subtitle2"
            sx={{
              fontWeight: "bold",
            }}
          >
            {title}: {info.version}
          </Typography>
          <Typography
            variant="caption"
            sx={{
              color: "text.secondary",
            }}
          >
            Released: {new Date(info.release_date).toLocaleDateString()}
          </Typography>
        </HStack>
        <Typography
          variant="body2"
          sx={{
            whiteSpace: "pre-wrap",
            color: "text.secondary",
            fontSize: "0.85rem",
          }}
        >
          {info.release_notes}
        </Typography>
        <Box sx={{ mt: 1 }}>
          <Button
            variant="contained"
            color="warning"
            size="small"
            onClick={onApply}
            disabled={busy}
            sx={{ fontWeight: "bold" }}
          >
            {buttonText}
          </Button>
        </Box>
      </VStack>
    </Alert>
  );

  return (
    <Box sx={{ px: { xs: 1, md: 2 }, py: 2, width: "100%" }}>
      <Backdrop
        open={serverUpdateInProgress}
        sx={(theme) => ({
          zIndex: theme.zIndex.drawer + 20,
          backgroundColor: "rgba(15, 23, 42, 0.72)",
          px: 2,
        })}
      >
        <Card
          variant="outlined"
          sx={{
            width: "100%",
            maxWidth: 520,
            borderRadius: 3,
            boxShadow: 12,
          }}
        >
          <CardContent sx={{ p: 4 }}>
            <VStack
              spacing={2}
              sx={{ alignItems: "center", textAlign: "center" }}
            >
              <CircularProgress color="primary" size={40} />
              <Typography variant="h5" sx={{ fontWeight: 700 }}>
                Updating Chronix...
              </Typography>
              <Typography variant="body1" sx={{ color: "text.secondary" }}>
                {serverUpdateTargetVersion
                  ? `Chronix is downloading and restarting into version ${serverUpdateTargetVersion}.`
                  : "Chronix is downloading and restarting."}
              </Typography>
              <Typography
                variant="body2"
                sx={{ color: "text.secondary", maxWidth: 420 }}
              >
                Please wait while the server finishes updating. This page will
                refresh automatically when Chronix is ready, and actions are
                temporarily disabled during the update.
              </Typography>
            </VStack>
          </CardContent>
        </Card>
      </Backdrop>
      <VStack spacing={2} sx={{ maxWidth: 1000, width: "100%", mx: "auto" }}>
        <HStack
          alignItems="center"
          justifyContent="space-between"
          sx={{ flexWrap: "wrap" }}
        >
          <Box sx={{ display: "flex", alignItems: "center" }}>
            <Typography variant="h5">Application Updates</Typography>
            <SectionHelp section={HELP_SECTIONS.updates} />
          </Box>
          <HStack spacing={1}>
            <Button
              variant="outlined"
              startIcon={<SystemUpdate />}
              onClick={onCheck}
              disabled={busy}
            >
              Check Now
            </Button>
          </HStack>
        </HStack>
        <Divider
          sx={{
            borderColor: (theme) =>
              theme.palette.mode === "light"
                ? "rgba(25, 118, 210, 0.2)"
                : "rgba(25, 118, 210, 0.4)",
          }}
        />

        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 6 }}>
            <Card variant="outlined" sx={{ borderRadius: 3, height: "100%" }}>
              <CardContent>
                <VStack spacing={2}>
                  <Typography variant="h6">Chronix App Status</Typography>
                  <HStack justifyContent="space-between">
                    <Typography
                      sx={{
                        color: "text.secondary",
                      }}
                    >
                      Current Version:
                    </Typography>
                    <Typography
                      sx={{
                        fontWeight: "bold",
                      }}
                    >
                      {status?.currentVersion}
                    </Typography>
                  </HStack>
                  <HStack justifyContent="space-between">
                    <Typography
                      sx={{
                        color: "text.secondary",
                      }}
                    >
                      Last Checked:
                    </Typography>
                    <Typography>
                      {status?.lastCheckTime
                        ? new Date(status.lastCheckTime).toLocaleString()
                        : "Never"}
                    </Typography>
                  </HStack>

                  {isServerOutdated() && status?.availableVersion?.server ? (
                    <UpdateAlert
                      title="Update Available"
                      info={status.availableVersion.server}
                      onApply={onApply}
                      busy={busy}
                      buttonText={`UPDATE TO ${status.availableVersion.server.version} NOW`}
                    />
                  ) : (
                    <Alert severity="success" variant="outlined" sx={{ mt: 1 }}>
                      Chronix App is up to date.
                    </Alert>
                  )}
                </VStack>
              </CardContent>
            </Card>
          </Grid>

          <Grid size={{ xs: 12, md: 6 }}>
            <Card variant="outlined" sx={{ borderRadius: 3, height: "100%" }}>
              <CardContent>
                <VStack spacing={3}>
                  <HStack justifyContent="space-between" alignItems="center">
                    <Typography variant="h6">App Update Settings</Typography>
                    <Button
                      variant="contained"
                      size="small"
                      onClick={onSaveAppSettings}
                      disabled={busy}
                    >
                      Save Settings
                    </Button>
                  </HStack>

                  <VStack spacing={2}>
                    <FormControlLabel
                      control={
                        <Switch
                          checked={status?.enabled || false}
                          onChange={(e) =>
                            setStatus((s) =>
                              s ? { ...s, enabled: e.target.checked } : null,
                            )
                          }
                        />
                      }
                      label="Enable Automatic Checks"
                    />

                    <Grid container spacing={2}>
                      <Grid size={{ xs: 12, sm: 6 }}>
                        <FormControl fullWidth size="small">
                          <InputLabel>Mode</InputLabel>
                          <Select
                            value={status?.mode || "notify"}
                            label="Mode"
                            onChange={(e) =>
                              setStatus((s) =>
                                s
                                  ? { ...s, mode: e.target.value as string }
                                  : null,
                              )
                            }
                            disabled={!status?.enabled}
                          >
                            <MenuItem value="notify">Notify Only</MenuItem>
                            <MenuItem value="automatic">
                              Fully Automatic
                            </MenuItem>
                          </Select>
                        </FormControl>
                      </Grid>
                      <Grid size={{ xs: 12, sm: 6 }}>
                        <TextField
                          label="Window (HH:MM)"
                          size="small"
                          placeholder="HH:MM (Empty for anytime)"
                          value={status?.windowStart || ""}
                          onChange={(e) =>
                            setStatus((s) =>
                              s ? { ...s, windowStart: e.target.value } : null,
                            )
                          }
                          disabled={
                            !status?.enabled || status?.mode !== "automatic"
                          }
                          fullWidth
                          helperText="Start time for 1-hour update window"
                        />
                      </Grid>
                    </Grid>
                  </VStack>
                </VStack>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        <Divider
          sx={{
            my: 1,
            borderColor: (theme) =>
              theme.palette.mode === "light"
                ? "rgba(25, 118, 210, 0.2)"
                : "rgba(25, 118, 210, 0.4)",
          }}
        />

        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 6 }}>
            <VStack spacing={2}>
              <HStack
                justifyContent="space-between"
                alignItems="center"
                sx={{ width: "100%", flexWrap: "wrap" }}
              >
                <Typography variant="h6">Agent Updates</Typography>
                {outdatedAgents.length > 1 &&
                  onlineOutdatedAgents.length > 0 && (
                    <Button
                      variant="outlined"
                      color="warning"
                      size="small"
                      startIcon={<SystemUpdate />}
                      onClick={onUpdateAllAgents}
                      disabled={busy || anyAgentBusy}
                    >
                      {anyAgentBusy
                        ? "UPDATING..."
                        : `UPDATE ALL (${onlineOutdatedAgents.length})`}
                    </Button>
                  )}
              </HStack>
              {agents.length > 0 ? (
                <Grid container spacing={2}>
                  {agents.map((agent) => (
                    <Grid key={agent.uuid} size={{ xs: 12 }}>
                      <Card variant="outlined" sx={{ borderRadius: 3 }}>
                        <CardContent sx={{ pb: "16px !important" }}>
                          <VStack spacing={1}>
                            <HStack
                              justifyContent="space-between"
                              alignItems="center"
                            >
                              <Typography
                                variant="subtitle1"
                                sx={{
                                  fontWeight: "bold",
                                }}
                              >
                                {agent.name}
                              </Typography>
                              <HStack spacing={1} alignItems="center">
                                <Typography
                                  variant="caption"
                                  sx={{
                                    color: "text.secondary",
                                  }}
                                >
                                  {agent.online ? "Online" : "Offline"}
                                </Typography>
                                <Box
                                  sx={{
                                    width: 10,
                                    height: 10,
                                    borderRadius: "50%",
                                    bgcolor: agent.online
                                      ? "success.main"
                                      : "error.main",
                                    boxShadow: agent.online
                                      ? "0 0 8px rgba(76, 175, 80, 0.5)"
                                      : "none",
                                  }}
                                />
                              </HStack>
                            </HStack>
                            <HStack justifyContent="space-between">
                              <Typography
                                variant="body2"
                                sx={{
                                  color: "text.secondary",
                                }}
                              >
                                Current Version:
                              </Typography>
                              <Typography
                                variant="body2"
                                sx={{
                                  fontWeight: "medium",
                                }}
                              >
                                {agent.version || "Unknown"}
                              </Typography>
                            </HStack>

                            {agentVersionInfo &&
                            agent.version !== targetAgentVersion ? (
                              <UpdateAlert
                                title="Update Available"
                                info={agentVersionInfo}
                                onApply={() => onUpdateAgent(agent)}
                                busy={!!agentBusy[agent.uuid] || !agent.online}
                                buttonText={
                                  agentBusy[agent.uuid]
                                    ? "UPDATING..."
                                    : !agent.online
                                      ? "AGENT OFFLINE"
                                      : `UPDATE TO ${targetAgentVersion} NOW`
                                }
                              />
                            ) : (
                              <Alert
                                severity="success"
                                variant="outlined"
                                sx={{ mt: 1 }}
                              >
                                Agent is up to date.
                              </Alert>
                            )}
                          </VStack>
                        </CardContent>
                      </Card>
                    </Grid>
                  ))}
                </Grid>
              ) : (
                <Alert severity="info" variant="outlined">
                  No agents connected.
                </Alert>
              )}
            </VStack>
          </Grid>

          <Grid size={{ xs: 12, md: 6 }}>
            <Card variant="outlined" sx={{ borderRadius: 3, height: "100%" }}>
              <CardContent>
                <VStack spacing={3}>
                  <HStack justifyContent="space-between" alignItems="center">
                    <Typography variant="h6">Agent Update Settings</Typography>
                    <Button
                      variant="contained"
                      size="small"
                      onClick={onSaveAgentSettings}
                      disabled={busy}
                    >
                      Save Settings
                    </Button>
                  </HStack>

                  <VStack spacing={2}>
                    <FormControlLabel
                      control={
                        <Switch
                          checked={status?.agentEnabled || false}
                          onChange={(e) =>
                            setStatus((s) =>
                              s
                                ? { ...s, agentEnabled: e.target.checked }
                                : null,
                            )
                          }
                        />
                      }
                      label="Enable Automatic Checks"
                    />

                    <Grid container spacing={2}>
                      <Grid size={{ xs: 12, sm: 6 }}>
                        <FormControl fullWidth size="small">
                          <InputLabel>Mode</InputLabel>
                          <Select
                            value={status?.agentMode || "notify"}
                            label="Mode"
                            onChange={(e) =>
                              setStatus((s) =>
                                s
                                  ? {
                                      ...s,
                                      agentMode: e.target.value as string,
                                    }
                                  : null,
                              )
                            }
                            disabled={!status?.agentEnabled}
                          >
                            <MenuItem value="notify">Notify Only</MenuItem>
                            <MenuItem value="automatic">
                              Fully Automatic
                            </MenuItem>
                          </Select>
                        </FormControl>
                      </Grid>
                      <Grid size={{ xs: 12, sm: 6 }}>
                        <TextField
                          label="Window (HH:MM)"
                          size="small"
                          placeholder="HH:MM (Empty for anytime)"
                          value={status?.agentWindowStart || ""}
                          onChange={(e) =>
                            setStatus((s) =>
                              s
                                ? { ...s, agentWindowStart: e.target.value }
                                : null,
                            )
                          }
                          disabled={
                            !status?.agentEnabled ||
                            status?.agentMode !== "automatic"
                          }
                          fullWidth
                          helperText="Start time for 1-hour update window"
                        />
                      </Grid>
                    </Grid>
                  </VStack>
                </VStack>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        <Snackbar
          open={snack.open}
          autoHideDuration={4000}
          onClose={() => setSnack((s) => ({ ...s, open: false }))}
          anchorOrigin={{ vertical: "top", horizontal: "center" }}
        >
          <Alert
            onClose={() => setSnack((s) => ({ ...s, open: false }))}
            severity={snack.severity}
            variant="filled"
            sx={{ width: "100%" }}
          >
            {snack.message}
          </Alert>
        </Snackbar>
      </VStack>
    </Box>
  );
};
