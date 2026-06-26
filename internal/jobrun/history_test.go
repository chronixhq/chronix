package jobrun

import (
	"testing"
)

func TestJobRunsMem_QueueStartFinishSuccess(t *testing.T) {
	// Reset package-level state for test isolation
	jobRunsMu.Lock()
	jobRunsByJob = map[int64][]JobRunRecord{}
	consecutiveFailures = map[int64]int{}
	maxRunsPerJob = 3
	jobRunsMu.Unlock()

	jobID := int64(1001)
	runID := "run-abc"

	RecordJobRunQueued(jobID, runID, 0)
	jobRunsMu.Lock()
	if len(jobRunsByJob[jobID]) != 1 {
		jobRunsMu.Unlock()
		t.Fatalf("expected 1 run after queued, got %d", len(jobRunsByJob[jobID]))
	}
	if jobRunsByJob[jobID][0].Status != "queued" {
		jobRunsMu.Unlock()
		t.Fatalf("status after queued = %s", jobRunsByJob[jobID][0].Status)
	}
	jobRunsMu.Unlock()

	RecordJobRunStarted(jobID, runID)
	jobRunsMu.Lock()
	if jobRunsByJob[jobID][0].Status != "running" || jobRunsByJob[jobID][0].StartedAt == nil {
		jobRunsMu.Unlock()
		t.Fatalf("expected running with StartedAt set")
	}
	jobRunsMu.Unlock()

	RecordJobRunFinished(jobID, runID, "success", "ok")
	jobRunsMu.Lock()
	defer jobRunsMu.Unlock()
	if jobRunsByJob[jobID][0].Status != "success" || jobRunsByJob[jobID][0].FinishedAt == nil {
		t.Fatalf("expected success with FinishedAt set")
	}
	if jobRunsByJob[jobID][0].Message == nil || *jobRunsByJob[jobID][0].Message != "ok" {
		t.Fatalf("expected message 'ok'")
	}
	if consecutiveFailures[jobID] != 0 {
		t.Fatalf("expected consecutiveFailures reset to 0, got %d", consecutiveFailures[jobID])
	}
}

func TestJobRunsMem_RingBuffer_Truncates(t *testing.T) {
	jobRunsMu.Lock()
	jobRunsByJob = map[int64][]JobRunRecord{}
	consecutiveFailures = map[int64]int{}
	maxRunsPerJob = 2
	jobRunsMu.Unlock()

	jobID := int64(2002)
	RecordJobRunQueued(jobID, "r1", 0)
	RecordJobRunQueued(jobID, "r2", 0)
	RecordJobRunQueued(jobID, "r3", 0)

	jobRunsMu.Lock()
	defer jobRunsMu.Unlock()
	arr := jobRunsByJob[jobID]
	if len(arr) != 2 {
		t.Fatalf("expected 2 records, got %d", len(arr))
	}
	// New items are prepended (most recent first). With cap=2, should hold r3, r2
	if arr[0].RunID != "r3" || arr[1].RunID != "r2" {
		t.Fatalf("unexpected ring contents: %#v", arr)
	}
}
