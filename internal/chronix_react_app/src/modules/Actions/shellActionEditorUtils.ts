import {type ShellStepDraft} from './types';

export const COMMON_SHELLS = [
    {label: '/bin/bash', os: 'Linux / macOS'},
    {label: '/bin/sh', os: 'Linux / macOS'},
    {label: '/bin/zsh', os: 'macOS / Linux'},
    {label: 'C:\\Windows\\System32\\cmd.exe', os: 'Windows'},
    {label: 'C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe', os: 'Windows'},
] as const;

function createShellEditorId() {
    return globalThis.crypto?.randomUUID?.() || Math.random().toString(36).slice(2);
}

export function createShellDraftStep(stepNumber: number): ShellStepDraft {
    return {
        id: createShellEditorId(),
        name: `Step ${stepNumber}`,
        runMode: 'command',
        command: '',
        scriptText: '',
        shellPath: '/bin/sh',
        workingDir: '',
        timeout: '60',
        env: {},
        outputCaptureMaxBytes: '65536',
        outputTruncation: 'tail',
        expectation: {kind: 'noError'},
        outputCapture: {},
        onFailure: 'exit',
    };
}

export function toShellDraftSteps(steps?: any[]): ShellStepDraft[] {
    return (steps || [])
        .slice()
        .sort((left, right) => left.order - right.order)
        .map((step) => ({
            id: step.id ? String(step.id) : createShellEditorId(),
            name: step.name,
            runMode: step.runMode,
            command: step.command || '',
            scriptText: step.scriptText || '',
            shellPath: step.shellPath || '/bin/sh',
            workingDir: step.workingDir || '',
            timeout: String(step.timeoutSeconds ?? '60'),
            env: step.env || {},
            outputCaptureMaxBytes: String(step.outputCaptureMaxBytes ?? '65536'),
            outputTruncation: step.outputTruncation || 'tail',
            expectation: step.expectation ?? {kind: 'noError'},
            outputCapture: step.outputCapture || {},
            onFailure: step.onFailure || 'exit',
        }));
}

export function buildShellActionPayload(name: string, description: string, notes: string, steps: ShellStepDraft[]) {
    return {
        name,
        description: description || undefined,
        notes: notes || undefined,
        steps: steps.map((step, index) => ({
            order: index,
            name: step.name,
            runMode: step.runMode,
            command: step.runMode === 'command' ? step.command : undefined,
            scriptText: step.runMode === 'script' ? step.scriptText : undefined,
            shellPath: step.shellPath,
            workingDir: step.workingDir || undefined,
            timeoutSeconds: parseInt(step.timeout) || 60,
            env: Object.fromEntries(Object.entries(step.env).filter(([key]) => key.trim() !== '')),
            outputCaptureMaxBytes: parseInt(step.outputCaptureMaxBytes) || 65536,
            outputTruncation: step.outputTruncation,
            expectation: step.expectation,
            outputCapture: step.outputCapture,
            onFailure: step.onFailure,
        })),
    };
}
