package webtaskactionsapi

import "gorm.io/datatypes"

type webtaskActionStepPayload struct {
	ID              any                `json:"id"`
	StepOrder       int64              `json:"stepOrder"`
	Name            string             `json:"name"`
	Method          string             `json:"method"`
	URL             string             `json:"url"`
	Headers         *datatypes.JSONMap `json:"headers"`
	Body            *string            `json:"body"`
	TimeoutSeconds  *int64             `json:"timeoutSeconds"`
	Expectation     *datatypes.JSONMap `json:"expectation"`
	ResponseCapture *datatypes.JSONMap `json:"responseCapture"`
	OnFailure       *string            `json:"onFailure"`
	Variables       *datatypes.JSONMap `json:"variables"`
	ActionID        any                `json:"actionId"`
}

type webtaskActionPayload struct {
	ID          any                        `json:"id"`
	Name        string                     `json:"name" binding:"required"`
	Description string                     `json:"description"`
	Notes       string                     `json:"notes"`
	Steps       []webtaskActionStepPayload `json:"steps"`
	ActionType  any                        `json:"action_type"`
	Dialect     any                        `json:"dialect"`
	CreatedAt   any                        `json:"createdAt"`
	UpdatedAt   any                        `json:"updatedAt"`
}

type webtaskActionTestPayload struct {
	ConnectionID int64                           `json:"connectionId"`
	Steps        []webtaskActionStepModelPayload `json:"steps"`
	Variables    map[string]any                  `json:"variables"`
	ActionID     any                             `json:"id"`
	Dialect      any                             `json:"dialect"`
}

type webtaskActionStepModelPayload = struct {
	ID              *int64             `json:"id"`
	ActionID        int64              `json:"actionId"`
	StepOrder       *int64             `json:"stepOrder"`
	Name            string             `json:"name"`
	Method          string             `json:"method"`
	URL             string             `json:"url"`
	Headers         *datatypes.JSONMap `json:"headers"`
	Body            *string            `json:"body"`
	TimeoutSeconds  *int64             `json:"timeoutSeconds"`
	Expectation     *datatypes.JSONMap `json:"expectation"`
	ResponseCapture *datatypes.JSONMap `json:"responseCapture"`
	OnFailure       *string            `json:"onFailure"`
}

func pickID(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
