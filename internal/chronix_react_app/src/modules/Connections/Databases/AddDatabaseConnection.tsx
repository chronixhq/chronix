import {useEffect, useMemo, useState} from "react";
import {apiPost} from "@dsherwin/react-api-interface";
import {useNavigate} from 'react-router';
import {useFeatureAvailability} from "../../../data/FeatureAvailabilityContext.tsx";
import {useConnections} from "../../../data/ConnectionsContext.tsx";
import {fetchAgentOptions, type AgentOption} from "../../Agents/api.ts";
import {type DbConnectionDraft, type DbDriver} from '../types.ts';
import {DatabaseConnectionEditor} from './DatabaseConnectionEditor.tsx';
import {buildDatabaseConnectionDsn, buildDatabaseConnectionPreviewDsn, DEFAULT_DB_CONNECTIONS} from './databaseConnectionForm.ts';

const isEmpty = (value?: string | number | null) => value === undefined || value === null || String(value).trim() === "";

type ConnectionTestResponse = {
    ok?: boolean
    message?: string
}

export const AddDatabaseConnection = () => {
    const navigate = useNavigate();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {reload: reloadConnections} = useConnections();
    const dbLimit = checkLimit('db_connections');

    const [agents, setAgents] = useState<AgentOption[]>([]);
    const [agentsLoading, setAgentsLoading] = useState(false);
    const [draft, setDraft] = useState<DbConnectionDraft>({
        name: "",
        driver: "postgres",
        description: "",
        ...DEFAULT_DB_CONNECTIONS.postgres,
    });
    const [showPassword, setShowPassword] = useState(false);
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState<null | { ok: boolean; message: string }>(null);
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [nameTouched, setNameTouched] = useState(false);

    useEffect(() => {
        if (!dbLimit.allowed) {
            navigate('/databases/list');
        }
    }, [dbLimit.allowed, navigate]);

    useEffect(() => {
        setDraft((current) => ({
            ...current,
            ...DEFAULT_DB_CONNECTIONS[current.driver],
            ...(current.driver === "sqlite"
                ? {host: undefined, port: undefined, database: undefined, username: undefined, password: undefined}
                : {filePath: undefined}),
        }));
        setTestResult(null);
    }, [draft.driver]);

    useEffect(() => {
        let alive = true;
        (async () => {
            try {
                setAgentsLoading(true);
                const list = await fetchAgentOptions();
                if (!alive) return;
                setAgents(list);
            } catch (e) {
                console.error(e);
                setAgents([]);
            } finally {
                setAgentsLoading(false);
            }
        })();
        return () => {
            alive = false;
        };
    }, []);

    useEffect(() => {
        if (nameTouched || draft.driver === "sqlite") return;
        const database = (draft.database || "").trim();
        const host = (draft.host || "").trim();
        const autoName = database && host ? `${database}@${host}` : (database || host);
        if (autoName && draft.name !== autoName) {
            setDraft((current) => ({...current, name: autoName}));
        }
    }, [draft.database, draft.driver, draft.host, draft.name, nameTouched]);

    const validate = (): boolean => {
        const nextErrors: Record<string, string> = {};

        if (isEmpty(draft.name)) nextErrors.name = "Give this connection a label.";

        if (draft.driver === "sqlite") {
            if (isEmpty(draft.filePath)) nextErrors.filePath = "Path to the database file is required.";
        } else {
            if (draft.driver === "snowflake") {
                if (isEmpty(draft.host)) nextErrors.host = "Account is required.";
            } else {
                if (isEmpty(draft.host)) nextErrors.host = "Host is required.";
                if (isEmpty(draft.port)) nextErrors.port = "Port is required.";
            }
            if (isEmpty(draft.database) && draft.driver !== "mysql" && draft.driver !== "snowflake") nextErrors.database = "Database is required.";
            if (isEmpty(draft.username)) nextErrors.username = "Username is required.";
        }

        setErrors(nextErrors);
        return Object.keys(nextErrors).length === 0;
    };

    const rawDsn = useMemo(() => buildDatabaseConnectionDsn(draft), [draft]);
    const previewDsn = useMemo(() => buildDatabaseConnectionPreviewDsn(draft), [draft]);

    const handleTest = async () => {
        if (!validate()) return;
        setTesting(true);
        setTestResult(null);
        try {
            const body: { driver: DbDriver; dsn: string; agentUuid?: string } = {
                driver: draft.driver,
                dsn: rawDsn,
            };
            if (draft.agentUuid) body.agentUuid = draft.agentUuid;
            const response = await apiPost("/connections/test", body) as ConnectionTestResponse;
            setTestResult({ok: !!response?.ok, message: response?.message || (response?.ok ? "Connection succeeded" : "Connection failed")});
        } catch (e) {
            console.error(e);
            setTestResult({ok: false, message: "Network error"});
        } finally {
            setTesting(false);
        }
    };

    const handleSave = async () => {
        if (!validate()) return;
        try {
            const payload: {
                name: string
                driver: DbDriver
                description?: string
                dsn: string
                autoCheckEnabled: boolean
                autoCheckSeconds: number
                alertEmails: string
                alertPhones: string
                notifyOnFailure: boolean
                agentUuid?: string
            } = {
                name: draft.name,
                driver: draft.driver,
                description: draft.description,
                dsn: rawDsn,
                autoCheckEnabled: !!draft.autoCheckEnabled,
                autoCheckSeconds: draft.autoCheckSeconds || 3600,
                alertEmails: draft.alertEmails || "",
                alertPhones: draft.alertPhones || "",
                notifyOnFailure: !!draft.notifyOnFailure,
            };
            if (draft.agentUuid) payload.agentUuid = draft.agentUuid;
            const response = await apiPost("/connections", payload) as ConnectionTestResponse;
            if (response?.ok === false) throw new Error('Save failed');
            void reloadFeatureAvailability();
            void reloadConnections();
            setTestResult({ok: true, message: "Saved. Redirecting…"});
            window.setTimeout(() => navigate('/databases/list'), 700);
        } catch (e) {
            console.error(e);
            setTestResult({ok: false, message: "Save failed"});
        }
    };

    return (
        <DatabaseConnectionEditor
            title="Add database connection"
            infoTooltip="Used by jobs and explorers. You can edit later."
            draft={draft}
            setDraft={setDraft}
            errors={errors}
            agents={agents}
            agentsLoading={agentsLoading}
            showPassword={showPassword}
            setShowPassword={setShowPassword}
            testing={testing}
            testResult={testResult}
            onTest={handleTest}
            onCancel={() => navigate('/databases/list')}
            onSave={handleSave}
            saveLabel="Save & Continue"
            previewDsn={previewDsn}
            onNameInput={() => setNameTouched(true)}
        />
    );
}
