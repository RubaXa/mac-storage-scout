// @task spec/tasks/mac-storage-scout.task-09.md
// @purpose Benchmark walker strategies to establish performance baseline and identify optimization candidates.
package fs

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// benchRoot returns the tree root for benchmarks.
// Override with MSS_BENCH_ROOT env var to point at a large real tree.
// Falls back to os.TempDir() for CI (small tree, measures overhead not throughput).
func benchRoot() string {
	if r := os.Getenv("MSS_BENCH_ROOT"); r != "" {
		return r
	}
	return os.TempDir()
}

func baseCfg(workers int) domain.MssScanConfig {
	return domain.MssScanConfig{
		Paths:          []string{benchRoot()},
		ThresholdBytes: 1,
		TopN:           1,
		SizeMode:       domain.MssSizeModeLogical,
		Workers:        workers,
	}
}

// BenchmarkWalkerCurrent — baseline: current MssGoFsWalkerAdapter (workers = NumCPU*2)
func BenchmarkWalkerCurrent(b *testing.B) {
	adapter := &MssGoFsWalkerAdapter{}
	cfg := baseCfg(0) // 0 = auto (NumCPU*2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Walk(context.Background(), cfg, func(domain.MssWalkEvent) {})
	}
}

// BenchmarkWalkerWorkersCPU — tuning: workers = NumCPU (half of default)
func BenchmarkWalkerWorkersCPU(b *testing.B) {
	adapter := &MssGoFsWalkerAdapter{}
	cfg := baseCfg(runtime.NumCPU())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Walk(context.Background(), cfg, func(domain.MssWalkEvent) {})
	}
}

// BenchmarkWalkerWorkersCPUx4 — tuning: workers = NumCPU*4
func BenchmarkWalkerWorkersCPUx4(b *testing.B) {
	adapter := &MssGoFsWalkerAdapter{}
	cfg := baseCfg(runtime.NumCPU() * 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Walk(context.Background(), cfg, func(domain.MssWalkEvent) {})
	}
}

// BenchmarkWalkerWorkers1 — worst-case: single worker (sequential baseline)
func BenchmarkWalkerWorkers1(b *testing.B) {
	adapter := &MssGoFsWalkerAdapter{}
	cfg := baseCfg(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Walk(context.Background(), cfg, func(domain.MssWalkEvent) {})
	}
}

// BenchmarkWalkerFilepathWalkDir — stdlib comparison: filepath.WalkDir (sequential, uses DirEntry)
func BenchmarkWalkerFilepathWalkDir(b *testing.B) {
	root := benchRoot()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filepath.WalkDir(root, func(_ string, _ os.DirEntry, err error) error {
			return nil
		})
	}
}

// BenchmarkWalkerFilepathWalk — stdlib comparison: filepath.Walk (sequential, calls Stat per entry)
func BenchmarkWalkerFilepathWalk(b *testing.B) {
	root := benchRoot()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filepath.Walk(root, func(_ string, _ os.FileInfo, err error) error {
			return nil
		})
	}
}
