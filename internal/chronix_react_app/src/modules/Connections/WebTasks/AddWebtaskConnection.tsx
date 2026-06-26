import {useEffect, useRef, useState} from "react";
import {apiPost} from "@dsherwin/react-api-interface";
import {useNavigate} from 'react-router';
import {useMuiPrompts} from '@dsherwin/mui-kit';
import {confirmOnNavigate, useUnsavedChanges} from "../../../lib/useUnsavedChanges.ts";
import {useFeatureAvailability} from "../../../data/FeatureAvailabilityContext.tsx";
import {useConnections} from "../../../data/ConnectionsContext.tsx";
import {fetchAgentOptions, type AgentOption} from "../../Agents/api.ts";
import {getConnectionApiCollectionPath, getConnectionListPath} from '../api.ts';
import type {WebtaskConnection} from '../types.ts';
import {WebtaskConnectionEditor} from './WebtaskConnectionEditor.tsx';
import {
    createDefaultWebtaskConnectionDraft,
    snapshotWebtaskConnectionDraft,
    validateWebtaskConnectionDraft,
} from './webtaskConnectionEditorUtils';

export const AddWebtaskConnection = () => {
    const navigate = useNavigate();
    const {confirmPrompt} = useMuiPrompts();
    const {checkLimit, reload: reloadFeatureAvailability} = useFeatureAvailability();
    const {reload: reloadConnections} = useConnections();
    const wtLimit = checkLimit('webtask_connections');
    const [agents, setAgents] = useState<AgentOption[]>([]);
    const [agentsLoading, setAgentsLoading] = useState<boolean>(false);
    const [draft, setDraft] = useState<Partial<WebtaskConnection>>(() => createDefaultWebtaskConnectionDraft());
    const [showPassword, setShowPassword] = useState(false);
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState<null | { ok: boolean; message: string }>(null);
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [dirty, setDirty] = useState<boolean>(false);
    const baselineRef = useRef<string>('');
    useUnsavedChanges(dirty);

    useEffect(() => {
        if (!wtLimit.allowed) {
            navigate(getConnectionListPath('webtask'));
        }
    }, [navigate, wtLimit.allowed]);

    useEffect(() => {
        const snapshot = snapshotWebtaskConnectionDraft(draft);
        if (!baselineRef.current) {
            baselineRef.current = snapshot;
            return;
        }
        setDirty(snapshot !== baselineRef.current);
    }, [draft]);

    useEffect(() => {
        let alive = true;
        (async () => {
            try {
                setAgentsLoading(true);
                const options = await fetchAgentOptions();
                if (!alive) return;
                setAgents(options);
            } catch (e) {
                console.log(e);
                setAgents([]);
            } finally {
                setAgentsLoading(false);
            }
        })();
        return () => {
            alive = false
        };
    }, []);

    const validate = (): boolean => {
        const nextErrors = validateWebtaskConnectionDraft(draft);
        setErrors(nextErrors);
        return Object.keys(nextErrors).length === 0;
    };

    const handleTest = async () => {
        if (!validate()) return;
        setTesting(true);
        setTestResult(null);
        try {
            const data = await apiPost(`${getConnectionApiCollectionPath('webtask')}/test`, draft) as any;
            setTestResult({ok: !!data?.ok, message: data?.message || (data?.ok ? 'Connection successful!' : 'Connection failed.')});
        } catch (e: any) {
            setTestResult({ok: false, message: e.message || 'Network error testing connection.'});
        } finally {
            setTesting(false);
        }
    };

    const handleSave = async () => {
        if (!validate()) return;
        try {
            await apiPost(getConnectionApiCollectionPath('webtask'), draft);
            setDirty(false);
            void reloadFeatureAvailability();
            void reloadConnections();
            navigate(getConnectionListPath('webtask'), {state: {refresh: true}});
        } catch (e: any) {
            setErrors({api: e.message || 'Network error saving connection.'});
        }
    };

    return (
        <WebtaskConnectionEditor
            title="Add Web Task Connection"
            infoTooltip="Configure a new API endpoint for automation."
            draft={draft}
            setDraft={setDraft}
            errors={errors}
            agents={agents}
            agentsLoading={agentsLoading}
            showPassword={showPassword}
            setShowPassword={setShowPassword}
            testing={testing}
            testResult={testResult}
            onDismissTestResult={() => setTestResult(null)}
            onTest={handleTest}
            onCancel={() => confirmOnNavigate(dirty, navigate, confirmPrompt)(getConnectionListPath('webtask'))}
            onSave={handleSave}
            saveLabel="Save Connection"
            apiError={errors.api}
        />
    );
};
