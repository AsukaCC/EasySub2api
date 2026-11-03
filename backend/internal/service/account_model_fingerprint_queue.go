package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const modelFingerprintQueueInterval = 5 * time.Second

type ModelFingerprintQueuedJob struct {
	AccountID string
	Snapshot  *ModelFingerprintSnapshot
}

type modelFingerprintQueueStore interface {
	modelFingerprintStore
	ListQueuedModelFingerprints(context.Context) ([]ModelFingerprintQueuedJob, error)
	ActivateModelFingerprint(context.Context, string, *ModelFingerprintSnapshot) (bool, error)
	CancelModelFingerprint(context.Context, string, string) (bool, error)
}

func IsModelFingerprintPending(status string) bool {
	return status == "queued" || status == "running" || status == "waiting"
}

type modelFingerprintQueueRuntime struct {
	ctx    context.Context
	cancel context.CancelFunc
	wake   chan struct{}
	wg     sync.WaitGroup
	jobs   map[string]context.CancelFunc // Protected by modelFingerprintMu.
}

func (s *AccountTestService) wakeModelFingerprintQueue() {
	if s == nil {
		return
	}
	if _, ok := s.accountRepo.(modelFingerprintQueueStore); !ok {
		return
	}
	s.modelFingerprintMu.Lock()
	defer s.modelFingerprintMu.Unlock()
	if s.modelFingerprintStopped {
		return
	}
	if s.modelFingerprintQueue == nil {
		ctx, cancel := context.WithCancel(context.Background())
		q := &modelFingerprintQueueRuntime{ctx: ctx, cancel: cancel, wake: make(chan struct{}, 1), jobs: make(map[string]context.CancelFunc)}
		s.modelFingerprintQueue = q
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
			ticker := time.NewTicker(modelFingerprintQueueInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-q.wake:
				case <-ticker.C:
				}
				s.dispatchModelFingerprints(q)
			}
		}()
	}
	select {
	case s.modelFingerprintQueue.wake <- struct{}{}:
	default:
	}
}

func (s *AccountTestService) StopModelFingerprintQueue() {
	if s == nil {
		return
	}
	s.modelFingerprintMu.Lock()
	s.modelFingerprintStopped = true
	q := s.modelFingerprintQueue
	if q != nil {
		q.cancel()
	}
	s.modelFingerprintMu.Unlock()
	if q != nil {
		q.wg.Wait()
	}
}

type modelFingerprintSlot struct {
	s    *AccountTestService
	held bool // Accessed only by its worker.
}

func (slot *modelFingerprintSlot) acquire() bool {
	if slot.held {
		return true
	}
	slot.s.modelFingerprintMu.Lock()
	defer slot.s.modelFingerprintMu.Unlock()
	if slot.s.modelFingerprintActive >= 4 || slot.s.modelFingerprintStopped {
		return false
	}
	slot.s.modelFingerprintActive++
	slot.held = true
	return true
}

func (slot *modelFingerprintSlot) release() {
	if !slot.held {
		return
	}
	slot.s.modelFingerprintMu.Lock()
	slot.s.modelFingerprintActive--
	slot.held = false
	slot.s.modelFingerprintMu.Unlock()
}

func (s *AccountTestService) dispatchModelFingerprints(q *modelFingerprintQueueRuntime) {
	store := s.accountRepo.(modelFingerprintQueueStore)
	ctx, cancel := context.WithTimeout(q.ctx, 30*time.Second)
	defer cancel()
	jobs, err := store.ListQueuedModelFingerprints(ctx)
	if err != nil {
		slog.Warn("fingerprint queue lookup failed")
		return
	}
	for _, job := range jobs {
		if ctx.Err() != nil {
			return
		}
		snapshot := job.Snapshot
		a, err := s.accountRepo.GetByID(ctx, job.AccountID)
		if err != nil {
			continue
		}
		reason := s.fingerprintSkipReason(ctx, a)
		if !time.Now().Before(snapshot.ExpiresAt) {
			reason = "queue_expired"
		}
		if reason == "account_busy" {
			continue
		}
		slot := &modelFingerprintSlot{s: s}
		if reason == "" && !slot.acquire() {
			return
		}
		// Claim before launching: another replica may have taken or cancelled it.
		snapshot.Status, snapshot.QueueReason = "running", ""
		snapshot.ExpiresAt = time.Now().UTC().Add(10 * time.Minute)
		claimed, err := store.ActivateModelFingerprint(ctx, job.AccountID, snapshot)
		if err != nil || !claimed {
			slot.release()
			continue
		}
		if reason != "" {
			finished := time.Now().UTC()
			snapshot.Status, snapshot.Error, snapshot.FinishedAt = "skipped", reason, &finished
			if err := store.SaveModelFingerprint(ctx, job.AccountID, snapshot); err != nil {
				slog.Warn("fingerprint skip could not be saved")
			}
			slot.release()
			continue
		}
		workerCtx, stop := context.WithCancel(q.ctx)
		s.modelFingerprintMu.Lock()
		q.jobs[snapshot.ID] = stop
		s.modelFingerprintMu.Unlock()
		q.wg.Add(1)
		go func(job ModelFingerprintQueuedJob, slot *modelFingerprintSlot) {
			defer q.wg.Done()
			defer func() { slot.release(); s.wakeModelFingerprintQueue() }()
			defer stop()
			defer func() { s.modelFingerprintMu.Lock(); delete(q.jobs, job.Snapshot.ID); s.modelFingerprintMu.Unlock() }()
			monitorDone := make(chan struct{})
			go func() {
				defer close(monitorDone)
				s.watchFingerprintCancellation(workerCtx, stop, job.AccountID, job.Snapshot.ID)
			}()
			defer func() { stop(); <-monitorDone }()
			s.runModelFingerprintContext(workerCtx, job.AccountID, job.Snapshot.UserID, job.Snapshot, store, slot)
		}(job, slot)
	}
}

// Cancellation is stored first, so it works even when a different replica owns the worker.
func (s *AccountTestService) CancelModelFingerprint(ctx context.Context, accountID, jobID string) (*ModelFingerprintSnapshot, error) {
	store, ok := s.accountRepo.(modelFingerprintQueueStore)
	if !ok {
		return nil, errors.New("fingerprint queue unavailable")
	}
	changed, err := store.CancelModelFingerprint(ctx, accountID, jobID)
	if err != nil {
		return nil, err
	}
	if changed {
		s.modelFingerprintMu.Lock()
		if q := s.modelFingerprintQueue; q != nil {
			if cancel := q.jobs[jobID]; cancel != nil {
				cancel()
			}
		}
		s.modelFingerprintMu.Unlock()
	}
	snapshot, err := s.GetModelFingerprint(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if snapshot == nil || snapshot.ID != jobID {
		return nil, apperrors.New(409, "MODEL_FINGERPRINT_CHANGED", "The account test has changed; refresh before cancelling")
	}
	return snapshot, nil
}

func (s *AccountTestService) watchFingerprintCancellation(ctx context.Context, cancel context.CancelFunc, accountID, jobID string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		// Check immediately too: cancellation can race with worker registration.
		readCtx, stop := context.WithTimeout(ctx, 3*time.Second)
		current, err := s.GetModelFingerprint(readCtx, accountID)
		stop()
		if err == nil && (current == nil || current.ID != jobID || !IsModelFingerprintPending(current.Status)) {
			cancel()
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *AccountTestService) pauseFingerprint(ctx context.Context, id string, snapshot *ModelFingerprintSnapshot, store modelFingerprintStore, slot *modelFingerprintSlot, reason string) bool {
	slot.release()
	s.wakeModelFingerprintQueue()
	if snapshot.Status != "waiting" || snapshot.QueueReason != reason {
		snapshot.Status, snapshot.QueueReason = "waiting", reason
		if err := store.SaveModelFingerprint(ctx, id, snapshot); err != nil {
			snapshot.Status, snapshot.Error = "failed", "save_failed"
			return false
		}
	}
	timer := time.NewTimer(modelFingerprintQueueInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (s *AccountTestService) waitForFingerprintSlot(ctx context.Context, id string, snapshot *ModelFingerprintSnapshot, store modelFingerprintStore, slot *modelFingerprintSlot) bool {
	for ctx.Err() == nil {
		a, err := s.accountRepo.GetByID(ctx, id)
		if err != nil {
			snapshot.Status, snapshot.Error = "skipped", "account_unavailable"
			return false
		}
		reason := s.fingerprintSkipReason(ctx, a)
		if reason != "" && reason != "account_busy" {
			snapshot.Status, snapshot.Error = "skipped", reason
			return false
		}
		if reason == "" && slot.acquire() {
			if snapshot.Status == "waiting" {
				snapshot.Status, snapshot.QueueReason = "running", ""
				if err := store.SaveModelFingerprint(ctx, id, snapshot); err != nil {
					snapshot.Status, snapshot.Error = "failed", "save_failed"
					return false
				}
			}
			return true
		}
		if reason == "" {
			reason = "workers_busy"
		}
		if !s.pauseFingerprint(ctx, id, snapshot, store, slot, reason) {
			return false
		}
	}
	return false
}
