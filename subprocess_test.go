package envpprof

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const helperEnv = "ENVPPROF_TEST_HELPER"

// Runs this test binary again as a subprocess with GOPPROF set, executing the helper named by mode.
// Returns the output and the HOME directory given to the subprocess.
func runHelper(t *testing.T, mode, gopprof string) (out, home string) {
	t.Helper()
	home = t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), helperEnv+"="+mode, "GOPPROF="+gopprof, "HOME="+home)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper %q failed: %v\n%s", mode, err, b)
	}
	return string(b), home
}

func gcABit() {
	for range 5 {
		runtime.GC()
		time.Sleep(20 * time.Millisecond)
	}
}

func TestHelperProcess(t *testing.T) {
	switch os.Getenv(helperEnv) {
	case "":
		t.Skip("only run as a subprocess")
	case "drop-init":
		_ = Init()
		gcABit()
	case "call-init":
		stop := Init()
		gcABit()
		stop()
	case "call-stop":
		gcABit()
		Stop()
	case "stop-twice":
		Stop()
		Stop()
	}
}

func TestForgotStop(t *testing.T) {
	const warning = "forgot to Stop()"
	for _, tc := range []struct {
		mode string
		warn bool
	}{
		{"drop-init", true},
		{"call-init", false},
		{"call-stop", false},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			out, _ := runHelper(t, tc.mode, "heap")
			if got := strings.Contains(out, warning); got != tc.warn {
				t.Errorf("warning logged = %v, want %v; output:\n%s", got, tc.warn, out)
			}
		})
	}
}

func TestStopTwiceWritesProfilesOnce(t *testing.T) {
	out, home := runHelper(t, "stop-twice", "heap,block,mutex,cpu")
	entries, err := os.ReadDir(filepath.Join(home, "pprof"))
	if err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for _, e := range entries {
		counts[strings.TrimRight(e.Name(), "0123456789")]++
	}
	for _, name := range []string{"heap", "block", "mutex", "cpu"} {
		if counts[name] != 1 {
			t.Errorf("got %v %v profiles, want 1; output:\n%s", counts[name], name, out)
		}
	}
}
