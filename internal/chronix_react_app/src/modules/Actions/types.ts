// Shared Action types for Create, List, and Edit components
// Triggering reload to fix export visibility issue

export const TYPES_VERSION = '1.0.0';

export type Dialect = 'postgres' | 'mysql' | 'generic' | 'sqlite' | 'tsql';

export type FailurePolicy = 'exit' | 'continue';

export type ExpectationKind =
  | 'none'
  | 'noError'
  | 'rowExists'
  | 'noRowsReturned'
  | 'fieldEqualsFirst'
  | 'fieldEqualsLast'
  | 'fieldEquals' // deprecated, treated as first row for backward compatibility
  | 'rowsAffected'
  | 'contains'
  | 'notContains'
  | 'firstLineEquals'
  | 'lastLineEquals'
  | 'regexMatch'
  | 'bodyRegex'
  | 'exitCodeEquals'
  | 'statusCode'
  | 'bodyContains'
  | 'jsonPath'
  | 'latency';

export interface StepExpectationBase { kind: ExpectationKind }
export interface ExpectNone extends StepExpectationBase { kind: 'none' }
export interface ExpectNoError extends StepExpectationBase { kind: 'noError' }
export interface ExpectRowExists extends StepExpectationBase { kind: 'rowExists' }
export interface ExpectNoRowsReturned extends StepExpectationBase { kind: 'noRowsReturned' }
export interface ExpectFieldEqualsFirst extends StepExpectationBase { kind: 'fieldEqualsFirst'; column?: string; expected?: string }
export interface ExpectFieldEqualsLast extends StepExpectationBase { kind: 'fieldEqualsLast'; column?: string; expected?: string }
export interface ExpectFieldEqualsDeprecated extends StepExpectationBase { kind: 'fieldEquals'; column?: string; expected?: string }
export interface ExpectRowsAffected extends StepExpectationBase { kind: 'rowsAffected'; op?: '>=' | '==' | '<='; value?: string }
export interface ExpectContains extends StepExpectationBase { kind: 'contains'; value?: string }
export interface ExpectNotContains extends StepExpectationBase { kind: 'notContains'; value?: string }
export interface ExpectFirstLineEquals extends StepExpectationBase { kind: 'firstLineEquals'; value?: string }
export interface ExpectLastLineEquals extends StepExpectationBase { kind: 'lastLineEquals'; value?: string }
export interface ExpectRegexMatch extends StepExpectationBase { kind: 'regexMatch'; value?: string; group?: string; expected?: string }
export interface ExpectBodyRegex extends StepExpectationBase { kind: 'bodyRegex'; value?: string; group?: string; expected?: string }
export interface ExpectExitCodeEquals extends StepExpectationBase { kind: 'exitCodeEquals'; value?: string }
export interface ExpectHttpStatus extends StepExpectationBase { kind: 'statusCode'; op?: '==' | '!=' | '>' | '<' | '>=' | '<='; value?: string }
export interface ExpectBodyContains extends StepExpectationBase { kind: 'bodyContains'; value?: string }
export interface ExpectJsonPath extends StepExpectationBase { kind: 'jsonPath'; path?: string; value?: string }
export interface ExpectLatency extends StepExpectationBase { kind: 'latency'; value?: string }

export type StepExpectation =
  | ExpectNone
  | ExpectNoError
  | ExpectRowExists
  | ExpectNoRowsReturned
  | ExpectFieldEqualsFirst
  | ExpectFieldEqualsLast
  | ExpectFieldEqualsDeprecated
  | ExpectRowsAffected
  | ExpectContains
  | ExpectNotContains
  | ExpectFirstLineEquals
  | ExpectLastLineEquals
  | ExpectRegexMatch
  | ExpectBodyRegex
  | ExpectExitCodeEquals
  | ExpectHttpStatus
  | ExpectBodyContains
  | ExpectJsonPath
  | ExpectLatency;

// UI/editor draft types for steps (used in CreateAction/EditAction UIs)
export interface StepDraft {
  id: string;
  name: string;
  sql: string;
  timeout: string; // seconds as string; empty means inherit/default
  expectation: StepExpectation; // required in editors; callers should provide a default (e.g., {kind:'no_error'})
  outputCapture?: Record<string, any>;
  onFailure?: FailurePolicy; // what to do if expectation not met
}

export interface ShellStepDraft {
  id: string;
  name: string;
  runMode: 'command' | 'script';
  command: string;
  scriptText: string;
  shellPath: string;
  workingDir: string;
  timeout: string;
  env: Record<string, string>;
  outputCaptureMaxBytes: string;
  outputTruncation: 'head' | 'tail';
  expectation: StepExpectation;
  outputCapture?: Record<string, any>;
  onFailure?: FailurePolicy;
}

export type WebTaskMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

export interface WebtaskStepDraft {
  id: string;
  name: string;
  method: WebTaskMethod;
  url: string;
  headers: Record<string, string>;
  body: string;
  timeout: string;
  expectation: StepExpectation;
  responseCapture: Record<string, any>;
  onFailure?: FailurePolicy;
}

// Generic validation issue shape used by local validators
export interface ValidationIssue {
  code: string;
  message: string;
  stepId?: string;
}

export interface ActionStep {
  id: string;
  order: number;
  name: string;
  sqlText: string;
  timeoutSeconds?: number | null;
  expectation?: StepExpectation;
  outputCapture?: Record<string, any>;
  onFailure?: FailurePolicy;
}

export interface Action {
  id: string;
  name: string;
  dialect?: Dialect;
  actionType?: 'database' | 'shell' | 'webtask';
  description?: string;
  notes?: string;
  enabled?: boolean;
  suspended?: boolean;
  steps?: any[];
  createdAt?: string;
  updatedAt?: string;
}

export interface DatabaseStepTestResult {
  order: number;
  name: string;
  status: 'success' | 'error' | 'warning' | 'canceled';
  executedCode: string;
  executedArgs?: any[];
  rowsCount: number;
  rowsAffected: number;
  resultLines?: Record<string, any>[];
  expectationOk: boolean;
  expectationMsg: string;
  executionError?: string;
  capturedVars?: Record<string, any>;
}

export interface ShellStepTestResult {
  order: number;
  name: string;
  status: 'success' | 'error' | 'warning' | 'canceled';
  executedCode: string;
  exitCode: number;
  stdout: string;
  stderr: string;
  stdoutTruncated: boolean;
  stderrTruncated: boolean;
  expectationOk: boolean;
  expectationMsg: string;
  executionError?: string;
  capturedVars?: Record<string, any>;
}

export interface WebtaskStepTestResult {
  order: number;
  name: string;
  status: 'success' | 'error' | 'warning' | 'canceled';
  executedCode: string;
  requestUrl: string;
  requestMethod: string;
  requestHeaders?: Record<string, any>;
  requestBody?: string;
  responseStatus: number;
  responseHeaders?: Record<string, any>;
  responseBody?: string;
  latencyMs: number;
  expectationOk: boolean;
  expectationMsg: string;
  executionError?: string;
  capturedVars?: Record<string, any>;
}
