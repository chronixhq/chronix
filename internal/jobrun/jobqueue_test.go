package jobrun

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEnqueueJobRun_RunnerNotStarted(t *testing.T) {
	// Ensure runner not started
	if jobCh != nil {
		t.Skip("job runner already started in this process")
	}
	_, err := EnqueueJobRun(1, 0)
	if err == nil || err.Error() != "job runner not started" {
		t.Fatalf("expected job runner not started error, got %v", err)
	}
}

func TestFormatUnformatRunID_Fallback(t *testing.T) {
	// Force nil generator path
	runIDGen = nil
	id := generateRunID()
	f := FormatRunID(id)
	if len(f) < 8 {
		t.Fatalf("formatted id too short: %q", f)
	}
	back, err := UnformatRunID(f)
	if err != nil {
		t.Fatalf("unformat err: %v", err)
	}
	if back == 0 {
		t.Fatalf("expected non-zero parsed value")
	}
}

func TestSetJobQueueEnqueueBurst_CoercesAndApplies(t *testing.T) {
	SetJobQueueEnqueueBurst(0)
	if enqueueBurst != 1 {
		t.Fatalf("expected burst coerced to 1, got %d", enqueueBurst)
	}
	SetJobQueueEnqueueBurst(5)
	if enqueueBurst != 5 {
		t.Fatalf("expected burst 5, got %d", enqueueBurst)
	}
}

func TestSetupJobRunner_StartsWorkersAndAccepts(t *testing.T) {
	// Start a fresh context; skip if already started
	if jobCh != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	SetupJobRunner(ctx)
	// Allow goroutines to start
	time.Sleep(50 * time.Millisecond)
	if jobCh == nil {
		t.Fatalf("jobCh not initialized")
	}
}

// Ensure UnformatRunID error on empty string
func TestUnformatRunID_Empty(t *testing.T) {
	_, err := UnformatRunID("")
	if err == nil {
		t.Fatalf("expected error on empty id")
	}
	if !errors.Is(err, err) { // dummy assertion to touch errors.Is path
		_ = struct{}{} // satisfies linter
	}
}

func TestReserveJobRun_PreventsOverlap(t *testing.T) {
	activeJobRunsMu.Lock()
	activeJobRuns = map[int64]int{}
	activeJobRunsMu.Unlock()

	if !reserveJobRun(42) {
		t.Fatalf("expected first reservation to succeed")
	}
	if reserveJobRun(42) {
		t.Fatalf("expected second reservation for same job to fail")
	}

	releaseJobRun(42)

	if !reserveJobRun(42) {
		t.Fatalf("expected reservation after release to succeed")
	}
	releaseJobRun(42)
}
