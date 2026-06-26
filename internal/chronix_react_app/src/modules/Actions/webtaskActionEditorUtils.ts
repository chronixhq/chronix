import {type WebtaskStepDraft} from './types';

function createWebtaskEditorId() {
    return globalThis.crypto?.randomUUID?.() || Math.random().toString(36).slice(2);
}

function compactRecord<T>(record: Record<string, T>) {
    return Object.fromEntries(Object.entries(record || {}).filter(([key]) => key.trim() !== ''));
}

export function createWebtaskDraftStep(stepNumber: number): WebtaskStepDraft {
    return {
        id: createWebtaskEditorId(),
        name: `Step ${stepNumber}`,
        method: 'GET',
        url: '',
        headers: {},
        body: '',
        timeout: '30',
        expectation: {kind: 'statusCode', op: '==', value: '200'},
        responseCapture: {},
        onFailure: 'exit',
    };
}

export function appendWebtaskDraftStep(steps: WebtaskStepDraft[]): WebtaskStepDraft[] {
    return [...steps, createWebtaskDraftStep(steps.length + 1)];
}

export function updateWebtaskDraftStep(
    steps: WebtaskStepDraft[],
    stepId: string,
    patch: Partial<WebtaskStepDraft>,
): WebtaskStepDraft[] {
    return steps.map((step) => (step.id === stepId ? {...step, ...patch} : step));
}

export function removeWebtaskDraftStep(steps: WebtaskStepDraft[], stepId: string): WebtaskStepDraft[] {
    return steps.filter((step) => step.id !== stepId);
}

export function moveWebtaskDraftStep(
    steps: WebtaskStepDraft[],
    stepId: string,
    direction: number,
): WebtaskStepDraft[] {
    const index = steps.findIndex((step) => step.id === stepId);
    const nextIndex = index + direction;
    if (index < 0 || nextIndex < 0 || nextIndex >= steps.length) {
        return steps;
    }

    const next = [...steps];
    [next[index], next[nextIndex]] = [next[nextIndex], next[index]];
    return next;
}

export function toWebtaskDraftSteps(steps?: any[]): WebtaskStepDraft[] {
    return (steps || [])
        .slice()
        .sort((left, right) => (left.stepOrder ?? left.order ?? 0) - (right.stepOrder ?? right.order ?? 0))
        .map((step, index) => ({
            id: step.id ? String(step.id) : createWebtaskEditorId(),
            name: step.name || `Step ${index + 1}`,
            method: step.method || 'GET',
            url: step.url || '',
            headers: step.headers || {},
            body: step.body || '',
            timeout: String(step.timeoutSeconds ?? step.timeout ?? '30'),
            expectation: step.expectation || {kind: 'statusCode', op: '==', value: '200'},
            responseCapture: step.responseCapture || {},
            onFailure: step.onFailure || 'exit',
        }));
}

export function validateWebtaskAction(name: string, steps: WebtaskStepDraft[]): string | null {
    if (!name.trim()) {
        return 'Action name is required';
    }
    if (steps.some((step) => !step.url.trim())) {
        return 'All steps must have a URL';
    }
    return null;
}

export function snapshotWebtaskActionDraft(
    name: string,
    description: string,
    notes: string,
    steps: WebtaskStepDraft[],
): string {
    return JSON.stringify({name, description, notes, steps});
}

export function buildWebtaskActionPayload(
    name: string,
    description: string,
    notes: string,
    steps: WebtaskStepDraft[],
) {
    return {
        name,
        description,
        notes,
        action_type: 'webtask',
        dialect: 'generic',
        steps: steps.map((step, index) => ({
            stepOrder: index + 1,
            name: step.name,
            method: step.method,
            url: step.url,
            headers: compactRecord(step.headers || {}),
            body: step.body,
            timeoutSeconds: parseInt(step.timeout, 10) || 30,
            expectation: step.expectation,
            responseCapture: compactRecord(step.responseCapture || {}),
            onFailure: step.onFailure,
        })),
    };
}

export function toWebtaskTestSteps(steps: WebtaskStepDraft[]) {
    return steps.map((step, index) => ({
        order: index + 1,
        name: step.name,
        method: step.method,
        url: step.url,
        headers: compactRecord(step.headers || {}),
        body: step.body,
        timeout: step.timeout,
        expectation: step.expectation,
        responseCapture: compactRecord(step.responseCapture || {}),
        onFailure: step.onFailure,
    }));
}
