import {describe, expect, it} from 'vitest';
import {createActionDraftStep, extractActionTemplateVars, toDraftSteps, validateActionStepSql} from './actionEditorUtils';

describe('actionEditorUtils', () => {
    it('extracts distinct template variables', () => {
        expect(extractActionTemplateVars('select * from foo where id = {{id}} and org = {{ org_id }} and id = {{id}}')).toEqual(['id', 'org_id']);
    });

    it('builds a default draft step with sane defaults', () => {
        const draft = createActionDraftStep(2);
        expect(draft.name).toBe('Step 2');
        expect(draft.expectation.kind).toBe('noError');
        expect(draft.onFailure).toBe('exit');
    });

    it('converts action steps into sorted drafts', () => {
        expect(toDraftSteps([
            {id: 'b', order: 2, name: 'Second', sqlText: 'select 2'},
            {id: 'a', order: 1, name: 'First', sqlText: 'select 1', timeoutSeconds: 30},
        ] as any)).toEqual([
            expect.objectContaining({id: 'a', timeout: '30'}),
            expect.objectContaining({id: 'b', timeout: '60'}),
        ]);
    });

    it('flags basic SQL issues', () => {
        expect(validateActionStepSql("select ('oops")).toEqual([
            {code: 'UNBALANCED_PARENS', message: 'Unbalanced parentheses.'},
            {code: 'UNTERMINATED_STRING', message: 'Unterminated single-quoted string literal.'},
        ]);
    });
});
