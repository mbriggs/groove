package git

import "testing"

func TestProjectNameFromRemote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"git@github.com:user/myapp.git", "myapp"},
		{"git@github.com:user/myapp", "myapp"},
		{"https://github.com/user/myapp.git", "myapp"},
		{"https://github.com/user/myapp", "myapp"},
		{"git@gitlab.com:org/subgroup/myapp.git", "myapp"},
	}

	for _, tt := range tests {
		got := ProjectNameFromRemote(tt.input)
		if got != tt.want {
			t.Errorf("ProjectNameFromRemote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
