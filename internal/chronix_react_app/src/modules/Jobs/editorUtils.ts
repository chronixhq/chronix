import type {Action} from '../Actions/types'
import type {ActionKind} from '../Actions/api'
import {fetchActionsByKind} from '../Actions/api'
import {fetchConnectionsByKind} from '../Connections/api'

export function extractVarsFromAction(action?: Action): string[] {
    if (!action?.steps) return [];
    const set = new Set<string>();
    const re1 = /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}/g;
    const re2 = /\$\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}/g;

    for (const s of action.steps) {
        const texts = [
            s.sqlText || '',
            s.command || '',
            s.scriptText || '',
            s.shellPath || '',
            s.workingDir || '',
            s.url || '',
            s.body || '',
        ];

        if (s.headers && typeof s.headers === 'object') {
            Object.keys(s.headers).forEach(k => texts.push(k));
            Object.values(s.headers).forEach(v => {
                if (typeof v === 'string') texts.push(v);
            });
        }

        if (s.expectation && typeof s.expectation === 'object') {
            Object.values(s.expectation).forEach(v => {
                if (typeof v === 'string') texts.push(v);
            });
        }

        if (s.env && typeof s.env === 'object') {
            Object.values(s.env).forEach(v => {
                if (typeof v === 'string') texts.push(v);
            });
        }

        texts.forEach(t => {
            let m;
            while ((m = re1.exec(t)) !== null) set.add(m[1]);
            while ((m = re2.exec(t)) !== null) set.add(m[1]);
        });
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b));
}

export function cronLooksValid(cron: string): boolean {
    const parts = cron.trim().split(/\s+/);
    if (parts.length !== 5) return false;
    const fieldRe = /^\*|(\*\/[0-9]+)|([0-9]{1,2})(-[0-9]{1,2})?(,[0-9]{1,2}(-[0-9]{1,2})?)*$/;
    return parts.every(p => fieldRe.test(p));
}

export function ordinal(n: number): string {
    const suffixes = ["th", "st", "nd", "rd"];
    const value = n % 100;
    return n + (suffixes[(value - 20) % 10] || suffixes[value] || suffixes[0]);
}

export async function fetchJobEditorResources(targetKind: ActionKind) {
    const [connections, actions] = await Promise.all([
        fetchConnectionsByKind(targetKind),
        fetchActionsByKind(targetKind),
    ]);

    return {connections, actions};
}
