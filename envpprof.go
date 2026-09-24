package envpprof

import (
	"expvar"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/anacrolix/log"
)

var pprofDir = defaultPprofDir()

// Profiles go in the user's home directory, or the temporary directory if there isn't one (for
// example when $HOME isn't set in containers and services).
func defaultPprofDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "pprof")
	}
	return filepath.Join(home, "pprof")
}

// Stop ends CPU profiling, waiting for writes to complete. If heap profiling is enabled, it also
// writes the heap profile to a file. Stop should be deferred from main if cpu or heap profiling
// are to be used through envpprof.
func Stop() {
	cleanupForgotStop.Stop()
	stopProfilers()
}

func stopProfilers() {
	for _, profiler := range profilers {
		profiler.stop()
	}
}

func startHTTP(value string) {
	var l net.Listener
	if value == "" {
		// Try a few ports near the conventional pprof port 6060, then let the OS pick one.
		for port := 6061; port <= 6070 && l == nil; port++ {
			l, _ = net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		}
		if l == nil {
			l, _ = net.Listen("tcp", "localhost:0")
		}
		if l == nil {
			log.Print("unable to create envpprof listener for http")
			return
		}
	} else {
		var addr string
		_, _, err := net.SplitHostPort(value)
		if err == nil {
			addr = value
		} else {
			addr = "localhost:" + value
		}
		l, err = net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
	}
	log.Printf("(pid=%d) envpprof serving http://%s", os.Getpid(), l.Addr())
	go func() {
		defer l.Close()
		log.Printf("error serving http on envpprof listener: %s", http.Serve(l, nil))
	}()
}

var (
	// Strong reference to the value watched for forgetting to Stop. The package holds it until
	// Init hands it off to the stop function it returns: there's no way to tell if the
	// package-level Stop will be called, but a dropped func can be detected by the GC.
	forgotStopToken   *forgotStopValueType
	cleanupForgotStop runtime.Cleanup
)

func init() {
	expvar.Publish("numGoroutine", expvar.Func(func() interface{} { return runtime.NumGoroutine() }))

	envValue := os.Getenv("GOPPROF")
	if envValue == "" {
		return
	}
	needStop := false
	for _, item := range strings.Split(envValue, ",") {
		equalsPos := strings.IndexByte(item, '=')
		var key, value string
		if equalsPos < 0 {
			key = item
		} else {
			key = item[:equalsPos]
			value = item[equalsPos+1:]
		}
		switch key {
		case "http":
			startHTTP(value)
		default:
			profiler, ok := profilers[key]
			if ok {
				profiler.start(key)
				needStop = true
			} else {
				log.Printf("unexpected GOPPROF key %q", key)
			}
		}
	}
	// This only installs the warning if profiling is enabled. But it could be for any consumer of
	// envpprof...
	if needStop {
		forgotStopToken, cleanupForgotStop = makeForget()
	}
}

// Contains a pointer so it isn't a tiny pointer-free allocation: the runtime may batch those
// together, and then the cleanup may never run. See runtime.AddCleanup.
type forgotStopValueType struct{ _ *byte }

func makeForget() (*forgotStopValueType, runtime.Cleanup) {
	forgot := new(forgotStopValueType)
	return forgot, runtime.AddCleanup(
		forgot,
		func(struct{}) {
			log.Printf("envpprof: forgot to Stop()")
		},
		struct{}{},
	)
}

// Synchronous init that returns the cleanup function directly with no risk. Future proofing for a
// safer way to do it. If profiling is enabled and the returned func is garbage collected without
// being called, a warning is logged.
func Init() (stop func()) {
	// Take ownership of the token from the package so that only the returned func keeps it alive.
	token := forgotStopToken
	forgotStopToken = nil
	return func() {
		Stop()
		// Keep the token reachable until the cleanup has been stopped.
		runtime.KeepAlive(token)
	}
}

// Runs main test suite with clean handled for you. Takes an interface rather
// than *testing.M so this package doesn't pull "testing" into non-test builds;
// *testing.M satisfies it.
func TestMain(m interface{ Run() int }) {
	stop := Init()
	code := m.Run()
	stop()
	os.Exit(code)
}
