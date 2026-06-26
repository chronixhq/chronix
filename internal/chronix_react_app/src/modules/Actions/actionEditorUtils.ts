import {type ActionStep, type StepDraft, type ValidationIssue} from './types';

export function createEditorId() {
    try {
        return globalThis.crypto?.randomUUID?.() || Math.random().toString(36).slice(2);
    } catch {
        return Math.random().toString(36).slice(2);
    }
}

export function createActionDraftStep(stepNumber: number, overrides: Partial<StepDraft> = {}): StepDraft {
    return {
        id: createEditorId(),
        name: `Step ${stepNumber}`,
        sql: '',
        timeout: '',
        expectation: {kind: 'noError'},
        outputCapture: {},
        onFailure: 'exit',
        ...overrides,
    };
}

export function toDraftSteps(steps?: ActionStep[]): StepDraft[] {
    return (steps || [])
        .slice()
        .sort((left, right) => left.order - right.order)
        .map((step) => ({
            id: step.id,
            name: step.name,
            sql: step.sqlText || '',
            timeout: step.timeoutSeconds == null ? '60' : String(step.timeoutSeconds),
            expectation: step.expectation ?? {kind: 'noError'},
            outputCapture: (step as any).outputCapture || {},
            onFailure: step.onFailure,
        }));
}

export function extractActionTemplateVars(sql: string): string[] {
    const re = /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}/g;
    const out = new Set<string>();
    let match: RegExpExecArray | null;
    while ((match = re.exec(sql))) {
        out.add(match[1]);
    }
    return Array.from(out);
}

export function validateActionStepSql(text: string): ValidationIssue[] {
    const errors: ValidationIssue[] = [];
    const source = text.trim();
    if (!source) {
        errors.push({code: 'EMPTY_SQL', message: 'SQL is required.'});
        return errors;
    }

    let depth = 0;
    for (let index = 0; index < source.length; index++) {
        const ch = source[index];
        if (ch === '(') depth++;
        else if (ch === ')') depth--;
        if (depth < 0) {
            errors.push({code: 'UNBALANCED_PARENS', message: 'Too many closing parentheses.'});
            break;
        }
    }
    if (depth !== 0) {
        errors.push({code: 'UNBALANCED_PARENS', message: 'Unbalanced parentheses.'});
    }

    let inSingle = false;
    for (let index = 0; index < source.length; index++) {
        const ch = source[index];
        if (ch === "'") {
            if (inSingle) {
                if (source[index + 1] === "'") {
                    index++;
                } else {
                    inSingle = false;
                }
            } else {
                inSingle = true;
            }
        }
    }
    if (inSingle) {
        errors.push({code: 'UNTERMINATED_STRING', message: 'Unterminated single-quoted string literal.'});
    }

    const openBlockIdx = source.indexOf('/*');
    const closeBlockIdx = source.indexOf('*/');
    if (openBlockIdx !== -1 && (closeBlockIdx === -1 || closeBlockIdx < openBlockIdx)) {
        errors.push({code: 'UNTERMINATED_COMMENT', message: 'Unterminated block comment (/* */).'});
    }

    return errors;
}
