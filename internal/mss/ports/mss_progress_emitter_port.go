// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define live progress contract.
package ports

import (
	"mac-storage-scout/internal/mss/domain"
	"sync/atomic"
)

type MssProgressEmitterPort interface {
	Start(counters *atomic.Pointer[domain.MssCounters]) func()
}
