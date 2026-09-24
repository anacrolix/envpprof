package envpprof

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

const helperEnv = "ENVPPROF_TEST_HELPER"

// Runs this test binary again as a subprocess with GOPPROF set, executing the helper named by mode.
func runHelper(t *testing.T, mode, gopprof string) string {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), helperEnv+"="+mode, "GOPPROF="+gopprof, "HOME="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper %q failed: %v\n%s", mode, err, out)
	}
	return string(out)
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
			out := runHelper(t, tc.mode, "heap")
			if got := strings.Contains(out, warning); got != tc.warn {
				t.Errorf("warning logged = %v, want %v; output:\n%s", got, tc.warn, out)
			}
		})
	}
}
