import {apiGet} from '@dsherwin/react-api-interface'

export interface AgentOption {
    uuid: string
    name: string
}

export async function fetchAgentOptions(): Promise<AgentOption[]> {
    const list = await apiGet('/agents')
    return (Array.isArray(list) ? list : []).map((agent: any) => ({
        uuid: String(agent?.uuid ?? ''),
        name: String(agent?.name || agent?.uuid || ''),
    })).filter((agent) => agent.uuid)
}

export function mergeSelectedAgent(options: AgentOption[], agentUuid?: string | null, agentName?: string | null): AgentOption[] {
    if (!agentUuid) return options
    if (options.some((option) => option.uuid === agentUuid)) return options
    return [...options, {uuid: agentUuid, name: agentName || agentUuid}]
}
