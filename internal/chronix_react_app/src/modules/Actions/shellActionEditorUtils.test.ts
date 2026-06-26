import {describe, expect, it} from 'vitest';
import {buildShellActionPayload, createShellDraftStep, toShellDraftSteps} from './shellActionEditorUtils';

describe('shellActionEditorUtils', () => {
    it('creates a default shell draft step', () => {
        const step = createShellDraftStep(3);
        expect(step.name).toBe('Step 3');
        expect(step.runMode).toBe('command');
        expect(step.shellPath).toBe('/bin/sh');
        expect(step.outputCaptureMaxBytes).toBe('65536');
    });

    it('normalizes loaded shell action steps', () => {
        expect(toShellDraftSteps([
            {id: 2, order: 2, name: 'Two', runMode: 'script', scriptText: 'echo two'},
            {id: 1, order: 1, name: 'One', runMode: 'command', command: 'echo one', timeoutSeconds: 10},
        ])).toEqual([
            expect.objectContaining({id: '1', timeout: '10', command: 'echo one'}),
            expect.objectContaining({id: '2', timeout: '60', scriptText: 'echo two'}),
        ]);
    });

    it('builds a shell action payload with filtered env vars', () => {
        const payload = buildShellActionPayload('Action', '', '', [{
            ...createShellDraftStep(1),
            env: {'': 'drop', GOOD: 'yes'},
            timeout: '90',
        }]);
        expect(payload).toEqual({
            name: 'Action',
            description: undefined,
            notes: undefined,
            steps: [
                expect.objectContaining({
                    order: 0,
                    timeoutSeconds: 90,
                    env: {GOOD: 'yes'},
                }),
            ],
        });
    });
});
