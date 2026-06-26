import {apiGet} from '@dsherwin/react-api-interface'
import type {Action} from './types'
import type {ConnectionKind} from '../Connections/types'

export type ActionKind = Extract<ConnectionKind, 'database' | 'shell' | 'webtask'>

export function normalizeAction(raw: any, actionType: ActionKind): Action {
    return {
        ...raw,
        id: String(raw?.id ?? ''),
        actionType,
        name: raw?.name ?? '',
        description: raw?.description ?? undefined,
        notes: raw?.notes ?? undefined,
        enabled: raw?.enabled ?? undefined,
        suspended: raw?.suspended ?? undefined,
        steps: Array.isArray(raw?.steps) ? raw.steps : [],
        createdAt: raw?.createdAt ?? raw?.created_at ?? undefined,
        updatedAt: raw?.updatedAt ?? raw?.updated_at ?? undefined,
    }
}

export async function fetchActionsByKind(kind: ActionKind): Promise<Action[]> {
    const endpoint = kind === 'database' ? '/actions' : kind === 'shell' ? '/shell/actions' : '/actions/webtask'
    const data = await apiGet(endpoint)
    return (Array.isArray(data) ? data : []).map((action) => normalizeAction(action, kind))
}

export async function fetchAllActions(): Promise<Action[]> {
    const [database, shell, webtask] = await Promise.all([
        fetchActionsByKind('database'),
        fetchActionsByKind('shell'),
        fetchActionsByKind('webtask'),
    ])
    return [...database, ...shell, ...webtask]
}
