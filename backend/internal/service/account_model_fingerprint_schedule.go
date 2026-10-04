package service

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type ModelFingerprintSchedule struct {
	AccountID string                  `json:"account_id"`
	UserID    string                  `json:"-"`
	Enabled   bool                    `json:"enabled"`
	Options   ModelFingerprintOptions `json:"options"`
	NextRunAt time.Time               `json:"next_run_at"`
}

type ModelFingerprintHistory struct {
	Items    []*ModelFingerprintSnapshot `json:"items"`
	Total    int                         `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"page_size"`
}

type modelFingerprintHistoryStore interface {
	ListModelFingerprintHistory(context.Context, string, int, int) (*ModelFingerprintHistory, error)
	GetModelFingerprintSchedule(context.Context, string) (*ModelFingerprintSchedule, error)
	SetModelFingerprintSchedule(context.Context, *ModelFingerprintSchedule) error
	ListDueModelFingerprints(context.Context, time.Time) ([]ModelFingerprintSchedule, error)
	CleanupModelFingerprintHistory(context.Context, time.Time) error
}

func (s *AccountTestService) GetFingerprintHistory(ctx context.Context, id string, page, size int) (*ModelFingerprintHistory, error) {
	store, ok := s.accountRepo.(modelFingerprintHistoryStore)
	if !ok {
		return nil, errors.New("fingerprint history unavailable")
	}
	return store.ListModelFingerprintHistory(ctx, id, page, size)
}

func (s *AccountTestService) GetFingerprintSchedule(ctx context.Context, id string) (*ModelFingerprintSchedule, error) {
	store, ok := s.accountRepo.(modelFingerprintHistoryStore)
	if !ok {
		return nil, errors.New("fingerprint schedules unavailable")
	}
	return store.GetModelFingerprintSchedule(ctx, id)
}

func (s *AccountTestService) SetFingerprintSchedule(ctx context.Context, schedule *ModelFingerprintSchedule) error {
	a, err := s.accountRepo.GetByID(ctx, schedule.AccountID)
	if err != nil {
		return err
	}
	if schedule.Enabled {
		if _, err := s.fingerprintKey(ctx, a, schedule.UserID, schedule.Options.APIKeyID); err != nil {
			return err
		}
		if err := ValidateAccountProtectionConfiguration(a); err != nil {
			return err
		}
		if err := s.validateFingerprintOptions(ctx, a, &schedule.Options); err != nil {
			return err
		}
	}
	if schedule.UserID == "" {
		return errors.New("missing audit actor")
	}
	store, ok := s.accountRepo.(modelFingerprintHistoryStore)
	if !ok {
		return errors.New("fingerprint schedules unavailable")
	}
	schedule.NextRunAt = time.Now().UTC().Truncate(30 * time.Minute).Add(30 * time.Minute)
	return store.SetModelFingerprintSchedule(ctx, schedule)
}

// Existing minute scheduler owns the lifecycle. Claims advance the half-hour slot
// atomically with the account lease, so multiple replicas cannot duplicate a run.
func (s *AccountTestService) RunDueModelFingerprints(ctx context.Context) {
	if s == nil {
		return
	}
	store, ok := s.accountRepo.(modelFingerprintHistoryStore)
	if !ok {
		return
	}
	now := time.Now().UTC()
	if err := store.CleanupModelFingerprintHistory(ctx, now); err != nil {
		slog.Warn("fingerprint history cleanup failed")
	}
	plans, err := store.ListDueModelFingerprints(ctx, now)
	if err != nil {
		slog.Warn("fingerprint schedule lookup failed")
		return
	}
	for _, plan := range plans {
		s.modelFingerprintMu.Lock()
		full := s.modelFingerprintActive >= 4
		s.modelFingerprintMu.Unlock()
		if full || ctx.Err() != nil {
			return
		}
		plan.Options.Source = "scheduled"
		if _, err := s.StartModelFingerprint(ctx, plan.AccountID, plan.Options.Model, plan.UserID, plan.Options); err != nil {
			slog.Warn("scheduled fingerprint could not start", "account_id", plan.AccountID)
		}
	}
}
