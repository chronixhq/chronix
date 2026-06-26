package sqlutil

import "testing"

func TestResolveSQLTemplateVariables(t *testing.T) {
	cases := []struct {
		name string
		in   string
		vars map[string]any
		want string
	}{
		{"empty", "", nil, ""},
		{"whitespace", "  \n\t  ", nil, ""},
		{"no vars", "SELECT 1", nil, "SELECT 1"},
		{"single var", "SELECT {{n}};", map[string]any{"n": 42}, "SELECT 42;"},
		{"repeated var", "{{x}} + {{x}}", map[string]any{"x": 3}, "3 + 3"},
		{"two vars", "{{a}}/{{b}}", map[string]any{"a": 6, "b": 2}, "6/2"},
		{"underscore name", "{{foo_bar}}", map[string]any{"foo_bar": "ok"}, "ok"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ResolveSQLTemplateVariables(c.in, c.vars)
			if got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}
