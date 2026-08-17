# envpprof

[![pkg.go.dev badge](https://img.shields.io/badge/pkg.go.dev-reference-blue)](https://pkg.go.dev/github.com/anacrolix/envpprof)

Run-time configuration of Go's pprof features, and of the default HTTP mux, via the `GOPPROF` environment variable.

```go
import _ "github.com/anacrolix/envpprof"
```

`envpprof` has an `init` function that runs at process initialization and checks `GOPPROF`. The variable is a comma-separated list of keys, each optionally taking a `=value`, for example `GOPPROF=http,block` or `GOPPROF=http=:6060,cpu`.

Importing the package also publishes a `numGoroutine` [expvar](https://pkg.go.dev/expvar).

Requires Go 1.24 or later.

## Keys

Key | Effect
--- | ------
`http` | Serves the default HTTP muxer [`"net/http".DefaultServeMux`](https://pkg.go.dev/net/http#pkg-variables). With no value, it listens on the first free TCP port from `6061` upwards on `localhost`. With a value, that value is used as the listen address: a bare port (`http=6060`) is taken as `localhost:6060`, otherwise it's used as a full `host:port` (`http=:6060`, `http=0.0.0.0:6060`). The PID and the resolved address are logged. `DefaultServeMux` is frequently the default location to expose status and debugging endpoints, including those provided by [`net/http/pprof`](https://pkg.go.dev/net/http/pprof), which `envpprof` imports for you. Failing to listen on an explicitly given address panics; a failure to find a free port when no value is given is only logged.
`cpu` | Calls [`"runtime/pprof".StartCPUProfile`](https://pkg.go.dev/runtime/pprof#StartCPUProfile), writing to a temporary file in `$HOME/pprof` named with the prefix `cpu`. `Stop` must run for the profile to be flushed and usable.
`trace` | Calls [`"runtime/trace".Start`](https://pkg.go.dev/runtime/trace#Start), writing to a `trace`-prefixed file. Like `cpu`, it needs `Stop` to flush.
`fgprof` | Runs [`github.com/felixge/fgprof`](https://github.com/felixge/fgprof) in pprof format, writing to an `fgprof`-prefixed file. Unlike `cpu`, this samples off-CPU (blocked) time too. Needs `Stop` to flush.
`heap` | Writes the `heap` profile to a `heap`-prefixed file when `Stop` is invoked. No run-time configuration is needed to collect it.
`block` | Calls [`"runtime".SetBlockProfileRate(10000)`](https://pkg.go.dev/runtime#SetBlockProfileRate), enabling profiling of goroutine blocking events, and writes the profile to a `block`-prefixed file on `Stop`. If `http` is enabled, the profile is also exposed at `/debug/pprof/block`.
`mutex` | Calls [`"runtime".SetMutexProfileFraction(100)`](https://pkg.go.dev/runtime#SetMutexProfileFraction), enabling profiling of mutex contention events, and writes the profile to a `mutex`-prefixed file on `Stop`. If `http` is enabled, the profile is also exposed at `/debug/pprof/mutex`.

The `block` and `mutex` rates are the "safe rates" recommended by the [Datadog Go profiler notes](https://github.com/DataDog/go-profiler-notes/blob/main/guide/README.md#go-profilers).

Profile files are created in `$HOME/pprof` (the directory is created if missing) with a random suffix, and are never removed. Their names are logged.

## Stopping

Every key except `http` writes a profile only when profiling is stopped, so `Stop` needs to run before the process exits. Any of the following work:

```go
func main() {
	defer envpprof.Stop()
	// ...
}
```

```go
func main() {
	stop := envpprof.Init()
	defer stop()
	// ...
}
```

[`Init`](https://pkg.go.dev/github.com/anacrolix/envpprof#Init) returns the stop function directly, which is harder to forget than the package-level [`Stop`](https://pkg.go.dev/github.com/anacrolix/envpprof#Stop). If profiling was enabled and the stop function is garbage collected without ever being called, `envpprof` logs a warning that `Stop` was forgotten.

For tests, [`TestMain`](https://pkg.go.dev/github.com/anacrolix/envpprof#TestMain) handles the whole lifecycle:

```go
func TestMain(m *testing.M) {
	envpprof.TestMain(m)
}
```

It takes an `interface{ Run() int }` rather than `*testing.M`, so that `envpprof` doesn't pull `testing` into non-test builds. `*testing.M` satisfies it.
