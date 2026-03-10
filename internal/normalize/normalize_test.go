package normalize

import (
	"strings"
	"testing"
)

func TestBranch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-feature", "my-feature"},
		{"feature/MY-TICKET-123_some.fix", "feature-my-ticket-123-some-fix"},
		{"UPPER_CASE", "upper-case"},
		{"---leading-trailing---", "leading-trailing"},
		{"multiple///slashes___underscores", "multiple-slashes-underscores"},
		{"a.b.c.d", "a-b-c-d"},
		{strings.Repeat("a", 60), strings.Repeat("a", 50)},
		// Verify that different prefixes produce different normalized names
		{"mbriggs/feature/foo", "mbriggs-feature-foo"},
		{"mbriggs/bugfix/foo", "mbriggs-bugfix-foo"},
	}
	for _, tt := range tests {
		got := Branch(tt.input)
		if got != tt.want {
			t.Errorf("Branch(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
