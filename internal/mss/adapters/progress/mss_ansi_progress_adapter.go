// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Emit compact live scan progress to terminal.
package progress

import (
	"fmt"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"sync/atomic"
	"time"
)

type MssAnsiProgressAdapter struct{}

// @implements {MssProgressEmitterPort} internal/mss/ports/mss_progress_emitter_port.go
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

	return func() {
		close(stop)
		<-done
	}
}

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
