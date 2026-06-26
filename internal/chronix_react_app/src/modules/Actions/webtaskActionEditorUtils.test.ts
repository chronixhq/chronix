import {describe, expect, it} from 'vitest';
import {
    appendWebtaskDraftStep,
    buildWebtaskActionPayload,
    createWebtaskDraftStep,
    toWebtaskDraftSteps,
    toWebtaskTestSteps,
    validateWebtaskAction,
} from './webtaskActionEditorUtils';

describe('webtaskActionEditorUtils', () => {
    it('creates a default webtask draft step', () => {
        const step = createWebtaskDraftStep(2);

        expect(step.name).toBe('Step 2');
        expect(step.method).toBe('GET');
        expect(step.timeout).toBe('30');
        expect(step.expectation).toEqual({kind: 'statusCode', op: '==', value: '200'});
    });

    it('appends steps with incremented labels', () => {
        const steps = appendWebtaskDraftStep([createWebtaskDraftStep(1)]);

        expect(steps).toHaveLength(2);
        expect(steps[1].name).toBe('Step 2');
    });

    it('normalizes loaded steps', () => {
        expect(toWebtaskDraftSteps([
            {id: 2, stepOrder: 2, name: 'Two', method: 'POST', url: '/two', timeoutSeconds: 15},
            {id: 1, stepOrder: 1, name: 'One', method: 'GET', url: '/one'},
        ])).toEqual([
            expect.objectContaining({id: '1', name: 'One', timeout: '30', url: '/one'}),
            expect.objectContaining({id: '2', name: 'Two', timeout: '15', method: 'POST'}),
        ]);
    });

    it('builds payloads with compacted headers and captures', () => {
        const payload = buildWebtaskActionPayload('Action', '', '', [{
            ...createWebtaskDraftStep(1),
            url: 'https://example.com',
            headers: {'': 'drop', Authorization: 'Bearer token'},
            responseCapture: {'': {source: 'header'}, token: {source: 'header', name: 'X-Token'}},
            timeout: '45',
        }]);

        expect(payload).toEqual({
            name: 'Action',
            description: '',
            notes: '',
            action_type: 'webtask',
            dialect: 'generic',
            steps: [
                expect.objectContaining({
                    stepOrder: 1,
                    url: 'https://example.com',
                    headers: {Authorization: 'Bearer token'},
                    responseCapture: {token: {source: 'header', name: 'X-Token'}},
                    timeoutSeconds: 45,
                }),
            ],
        });
    });

    it('builds normalized test steps', () => {
        const steps = toWebtaskTestSteps([{
            ...createWebtaskDraftStep(1),
            url: 'https://example.com',
            headers: {'': 'drop', Accept: 'application/json'},
        }]);

        expect(steps).toEqual([
            expect.objectContaining({
                order: 1,
                url: 'https://example.com',
                headers: {Accept: 'application/json'},
            }),
        ]);
    });

    it('validates the required action fields', () => {
        expect(validateWebtaskAction('', [createWebtaskDraftStep(1)])).toBe('Action name is required');
        expect(validateWebtaskAction('Action', [{...createWebtaskDraftStep(1), url: ''}])).toBe('All steps must have a URL');
        expect(validateWebtaskAction('Action', [{...createWebtaskDraftStep(1), url: 'https://example.com'}])).toBeNull();
    });
});
