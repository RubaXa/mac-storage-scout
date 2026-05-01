// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Validate progress adapter lifecycle behavior in non-interactive test environments.
package progress

import (
	"mac-storage-scout/internal/mss/domain"
	"sync/atomic"
	"testing"
)

func TestMssAnsiProgressAdapterStartNoTTY(t *testing.T) {
	adapter := &MssAnsiProgressAdapter{}

	var counters atomic.Pointer[domain.MssCounters]
	stop := adapter.Start(&counters)

	// In test environments Start commonly returns a no-op stop callback.
	// Stop should always be safe to call without panicking.
	stop()
	stop()
}
