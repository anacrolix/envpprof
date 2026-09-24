package envpprof

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultPprofDir(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("home directory isn't taken from $HOME")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	if got, want := defaultPprofDir(), filepath.Join(home, "pprof"); got != want {
		t.Errorf("with HOME set: got %q, want %q", got, want)
	}
	tmp := t.TempDir()
	t.Setenv("HOME", "")
	t.Setenv("TMPDIR", tmp)
	if got, want := defaultPprofDir(), filepath.Join(tmp, "pprof"); got != want {
		t.Errorf("with HOME unset: got %q, want %q", got, want)
	}
}

func TestProfileFilePattern(t *testing.T) {
	for _, tc := range []struct{ profile, want string }{
		{"cpu", "cpu-42-*.pprof"},
		{"heap", "heap-42-*.pprof"},
		{"trace", "trace-42-*.out"},
	} {
		if got := profileFilePattern(tc.profile, 42); got != tc.want {
			t.Errorf("profileFilePattern(%q) = %q, want %q", tc.profile, got, tc.want)
		}
	}
}
