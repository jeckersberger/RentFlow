package config

import "testing"

func TestFeatureEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "missing", value: "", want: false},
		{name: "true", value: "true", want: true},
		{name: "upper true", value: "TRUE", want: true},
		{name: "one", value: "1", want: true},
		{name: "false", value: "false", want: false},
		{name: "invalid", value: "enabled", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const key = "CRATEDESK_TEST_FEATURE_GATE"
			t.Setenv(key, tt.value)
			if got := FeatureEnabled(key); got != tt.want {
				t.Fatalf("FeatureEnabled(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
