package envpprof

import (
	"fmt"
	"os"

	"github.com/anacrolix/log"
)

func logWroteProfile(f *os.File, profile string) {
	log.Printf("wrote %v profile to %q", profile, f.Name())
}

func newPprofFileOrLog(profile string) (f *os.File) {
	err := os.MkdirAll(pprofDir, 0750)
	if err != nil {
		log.Printf("error creating pprof dir for %v profile: %v", profile, err)
		return nil
	}
	f, err = os.CreateTemp(pprofDir, profileFilePattern(profile, os.Getpid()))
	if err != nil {
		log.Printf("error creating %v pprof file: %v", profile, err)
	}
	return
}

// Returns an os.CreateTemp pattern for the profile file, including the process ID so files can be
// matched to processes, and an extension for the tools that read them.
func profileFilePattern(profile string, pid int) string {
	ext := "pprof"
	if profile == "trace" {
		// The extension used in the runtime/trace and "go tool trace" docs.
		ext = "out"
	}
	return fmt.Sprintf("%s-%d-*.%s", profile, pid, ext)
}
