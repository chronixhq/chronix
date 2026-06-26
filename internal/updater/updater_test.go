package updater

import "testing"

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "patch upgrade", current: "0.0.28", latest: "0.0.29", want: true},
		{name: "same version", current: "0.0.28", latest: "0.0.28", want: false},
		{name: "downgrade is not newer", current: "0.0.28", latest: "0.0.27", want: false},
		{name: "numeric comparison not lexical", current: "0.0.9", latest: "0.0.10", want: true},
		{name: "leading v supported", current: "v0.1.0", latest: "v0.1.1", want: true},
		{name: "dev build does not claim updates", current: "dev", latest: "0.1.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNewerVersion(tt.current, tt.latest); got != tt.want {
				t.Fatalf("IsNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{a: "0.0.28", b: "0.0.28", want: 0},
		{a: "0.0.28", b: "0.0.29", want: -1},
		{a: "0.1.0", b: "0.0.29", want: 1},
		{a: "1.2", b: "1.2.0", want: 0},
	}

	for _, tt := range tests {
		if got := compareVersions(tt.a, tt.b); got != tt.want {
			t.Fatalf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
