package envpprof

import (
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
	f, err = os.CreateTemp(pprofDir, profile)
	if err != nil {
		log.Printf("error creating %v pprof file: %v", profile, err)
	}
	return
}
