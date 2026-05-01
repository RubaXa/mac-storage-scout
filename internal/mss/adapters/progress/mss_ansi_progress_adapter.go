// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Emit compact live scan progress to terminal.
package progress

import (
	"fmt"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// MssAnsiProgressAdapter renders periodic progress status for interactive terminals.
//
// @purpose Provide lightweight live scan feedback without affecting worker throughput.
// @consumer internal/mss/app/mss_scan_orchestrator.go
// @implements {MssProgressEmitterPort} internal/mss/ports/mss_progress_emitter_port.go
type MssAnsiProgressAdapter struct{}

// @see {MssProgressEmitterPort#Start} internal/mss/ports/mss_progress_emitter_port.go
// @purpose Start periodic tty progress renderer and return stop callback.
// @consumer internal/mss/app/mss_scan_orchestrator.go
// @param counters Shared atomic counters pointer.
// @returns Stop callback.
// @post Returned stop function is safe to call once after Start.
func (a *MssAnsiProgressAdapter) Start(counters *atomic.Pointer[domain.MssCounters]) func() {
	if !mssIsTTY() {
		return func() {}
	}
	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()
		defer close(done)
		for {
			select {
			case <-stop:
				fmt.Fprint(os.Stdout, "\r\x1b[2K")
				return
			case <-t.C:
				c := counters.Load()
				if c == nil {
					continue
				}
				elapsed := time.Since(c.StartedAt).Truncate(time.Second)
				fmt.Fprintf(
					os.Stdout,
					"\r\x1b[2Kscanned: dirs=%d files=%d bytes=%s errors=%d elapsed=%s",
					c.DirsScanned,
					c.FilesScanned,
					domain.MssHumanBytes(c.BytesSeen),
					c.Errors,
					elapsed,
				)
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			close(stop)
			<-done
		})
	}
}

// mssIsTTY checks whether stdout is interactive terminal.
//
// @purpose Disable progress rendering for non-interactive output streams.
// @consumer MssAnsiProgressAdapter.Start.
// @returns True when tty output is available.
func mssIsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	return os.Getenv("TERM") != ""
}
