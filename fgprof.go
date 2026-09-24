package envpprof

import (
	"fmt"
	"io"

	g "github.com/anacrolix/generics"
	"github.com/felixge/fgprof"
)

// We have to do this to reliably run before init(). Plus as a bonus you can check from other code
// if it ran when you expected.
var registeredFgprof = func() bool {
	g.MapMustAssignNew(
		profilers,
		"fgprof",
		newContinuousWriter(func(w io.Writer) (func() error, error) {
			stop := fgprof.Start(w, fgprof.FormatPprof)
			// continuousWriter logs any error returned by stop.
			return func() (err error) {
				// fgprof divides by the sample rate when exporting, which is zero if it's
				// stopped before taking a sample.
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("fgprof panicked stopping: %v", r)
					}
				}()
				return stop()
			}, nil
		}))
	return true
}()
