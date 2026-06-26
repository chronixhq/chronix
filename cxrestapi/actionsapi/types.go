package actionsapi

import "gorm.io/datatypes"

type actionTestPayload struct {
	ConnectionID int64               `json:"connectionId"`
	Dialect      string              `json:"dialect"`
	Steps        []actionStepPayload `json:"steps"`
	Variables    map[string]any      `json:"variables"`
}

type actionStepPayload struct {
	Order          int64          `json:"order"`
	Name           string         `json:"name"`
	SQLText        string         `json:"sqlText"`
	TimeoutSeconds *int64         `json:"timeoutSeconds"`
	Expectation    map[string]any `json:"expectation"`
	OutputCapture  map[string]any `json:"outputCapture"`
	OnFailure      *string        `json:"onFailure"`
}

type actionPayload struct {
	Name        string              `json:"name"`
	Dialect     string              `json:"dialect"`
	Description *string             `json:"description"`
	Notes       *string             `json:"notes"`
	Steps       []actionStepPayload `json:"steps"`
	Enabled     *bool               `json:"enabled"`
	Suspended   *bool               `json:"suspended"`
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
