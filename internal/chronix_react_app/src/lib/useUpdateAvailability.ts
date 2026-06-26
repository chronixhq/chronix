import {useEffect, useMemo, useState} from 'react';
import {fetchUpdaterAgents, fetchUpdaterStatus} from '../modules/Settings/api.ts';
import type {UpdateAgentInfo, UpdaterStatus} from '../modules/Settings/types.ts';

type UpdateAvailabilityState = {
    status: UpdaterStatus | null;
    agents: UpdateAgentInfo[];
    loading: boolean;
    error: string | null;
    hasServerUpdateNotice: boolean;
    hasAgentUpdateNotice: boolean;
    shouldShowUpdateNotice: boolean;
    dismissUpdateNotice: () => void;
};

const UPDATE_NOTICE_STORAGE_PREFIX = 'chronix:update-notice-dismissed:';

export function useUpdateAvailability(): UpdateAvailabilityState {
    const [status, setStatus] = useState<UpdaterStatus | null>(null);
    const [agents, setAgents] = useState<UpdateAgentInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [dismissedNoticeKey, setDismissedNoticeKey] = useState<string | null>(null);

    useEffect(() => {
        let active = true;
        (async () => {
            try {
                const [statusRes, agentRes] = await Promise.all([
                    fetchUpdaterStatus(),
                    fetchUpdaterAgents(),
                ]);
                if (!active) return;
                setStatus(statusRes);
                setAgents(agentRes);
                setError(null);
            } catch (e) {
                console.error(e);
                if (!active) return;
                setError('Failed to load updater status.');
            } finally {
                if (active) setLoading(false);
            }
        })();
        return () => {
            active = false;
        };
    }, []);

    const hasServerUpdateNotice = useMemo(() => {
        const current = status?.currentVersion;
        const latest = status?.availableVersion?.server?.version;
        if (!current || !latest || current === latest) return false;
        return status?.mode !== 'automatic';
    }, [status]);

    const hasAgentUpdateNotice = useMemo(() => {
        const latest = status?.availableVersion?.['chronix-agent']?.version;
        if (!latest) return false;
        const anyOutdated = agents.some(agent => !!agent.version && agent.version !== latest);
        if (!anyOutdated) return false;
        return status?.agentMode !== 'automatic';
    }, [agents, status]);

    const updateNoticeKey = useMemo(() => {
        const parts: string[] = [];
        if (hasServerUpdateNotice && status?.availableVersion?.server?.version) {
            parts.push(`server:${status.availableVersion.server.version}`);
        }
        if (hasAgentUpdateNotice && status?.availableVersion?.['chronix-agent']?.version) {
            parts.push(`agent:${status.availableVersion['chronix-agent'].version}`);
        }
        return parts.length > 0 ? parts.join('|') : null;
    }, [hasAgentUpdateNotice, hasServerUpdateNotice, status]);

    useEffect(() => {
        if (!updateNoticeKey) {
            setDismissedNoticeKey(null);
            return;
        }
        if (typeof window === 'undefined') {
            setDismissedNoticeKey(null);
            return;
        }
        try {
            const stored = window.sessionStorage.getItem(`${UPDATE_NOTICE_STORAGE_PREFIX}${updateNoticeKey}`);
            setDismissedNoticeKey(stored === 'true' ? updateNoticeKey : null);
        } catch (storageError) {
            console.error('Failed to read dismissed update notice from sessionStorage.', storageError);
            setDismissedNoticeKey(null);
        }
    }, [updateNoticeKey]);

    const shouldShowUpdateNotice = !!updateNoticeKey && dismissedNoticeKey !== updateNoticeKey;

    const dismissUpdateNotice = () => {
        if (!updateNoticeKey || typeof window === 'undefined') return;
        try {
            window.sessionStorage.setItem(`${UPDATE_NOTICE_STORAGE_PREFIX}${updateNoticeKey}`, 'true');
        } catch (storageError) {
            console.error('Failed to persist dismissed update notice in sessionStorage.', storageError);
        }
        setDismissedNoticeKey(updateNoticeKey);
    };

    return {
        status,
        agents,
        loading,
        error,
        hasServerUpdateNotice,
        hasAgentUpdateNotice,
        shouldShowUpdateNotice,
        dismissUpdateNotice,
    };
}
