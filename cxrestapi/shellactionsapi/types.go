package shellactionsapi

import "gorm.io/datatypes"

type shellActionTestPayload struct {
	ShellConnectionID int64                    `json:"shellConnectionId"`
	Steps             []shellActionStepPayload `json:"steps"`
	Variables         map[string]any           `json:"variables"`
}

type shellActionStepPayload struct {
	Order                 int64             `json:"order"`
	Name                  string            `json:"name"`
	RunMode               string            `json:"runMode"`
	Command               *string           `json:"command"`
	ScriptText            *string           `json:"scriptText"`
	ShellPath             string            `json:"shellPath"`
	WorkingDir            *string           `json:"workingDir"`
	TimeoutSeconds        *int64            `json:"timeoutSeconds"`
	Env                   map[string]string `json:"env"`
	OutputCaptureMaxBytes int64             `json:"outputCaptureMaxBytes"`
	OutputTruncation      string            `json:"outputTruncation"`
	Expectation           map[string]any    `json:"expectation"`
	OutputCapture         map[string]any    `json:"outputCapture"`
	OnFailure             *string           `json:"onFailure"`
}

type shellActionPayload struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description"`
	Notes       *string                  `json:"notes"`
	Steps       []shellActionStepPayload `json:"steps"`
	Enabled     *bool                    `json:"enabled"`
	Suspended   *bool                    `json:"suspended"`
}

func pickID(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func capOutputCaptureBytes(b int64) int64 {
	if b < 1 {
		return 1
	}
	if b > 1048576 {
		return 1048576
	}
	return b
}

func toMap(m any) *datatypes.JSONMap {
	if m == nil {
		return nil
	}
	jm := datatypes.JSONMap{}
	if mm, ok := m.(map[string]any); ok {
		if len(mm) == 0 {
			return nil
		}
		for k, v := range mm {
			jm[k] = v
		}
		return &jm
	}
	if mm, ok := m.(map[string]string); ok {
		if len(mm) == 0 {
			return nil
		}
		for k, v := range mm {
			jm[k] = v
		}
		return &jm
	}
	return nil
}
