package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

type ModelFingerprintSchedule struct {
	AccountID   string                      `json:"account_id"`
	UserID      string                      `json:"-"`
	Enabled     bool                        `json:"enabled"`
	Options     ModelFingerprintOptions     `json:"options"`
	NextRunAt   time.Time                   `json:"next_run_at"`
	Window      *ModelFingerprintTimeWindow `json:"time_window,omitempty"`
	TestsPerDay int                         `json:"tests_per_day,omitempty"`
}

// Zero retains the hourly cadence of schedules created before daily counts.
func (s ModelFingerprintSchedule) NextRun(now time.Time) (time.Time, error) {
	w := s.Window
	if w == nil {
		w = &ModelFingerprintTimeWindow{EndHour: 24, Timezone: "UTC"}
	}
	if s.TestsPerDay == 0 {
		return w.NextRun(now)
	}
	loc, err := w.location()
	if err != nil {
		return time.Time{}, err
	}
	if s.TestsPerDay < 1 || s.TestsPerDay > min(144, (w.EndHour-w.StartHour)*60) {
		return time.Time{}, apperrors.BadRequest("INVALID_FINGERPRINT_FREQUENCY", "Tests per day must be between 1 and 144, with at least one minute between tests")
	}
	local := now.In(loc)
	for day := 0; day < 4; day++ {
		date := time.Date(local.Year(), local.Month(), local.Day()+day, 12, 0, 0, 0, loc)
		start := time.Date(date.Year(), date.Month(), date.Day(), w.StartHour, 0, 0, 0, loc)
		end := time.Date(date.Year(), date.Month(), date.Day(), w.EndHour, 0, 0, 0, loc)
		// A nonexistent opening hour on a DST transition has no slots that day.
		if start.Hour() != w.StartHour || !end.After(start) {
			continue
		}
		minutes := int(end.Sub(start) / time.Minute)
		for i := 0; i < s.TestsPerDay; i++ {
			next := start.Add(time.Duration(i*minutes/s.TestsPerDay) * time.Minute)
			if next.After(now) && w.Contains(next) {
				return next.UTC(), nil
			}
		}
	}
	return time.Time{}, errors.New("no fingerprint test slot available")
}

// The end hour is exclusive; 0-24 represents the entire day.
type ModelFingerprintTimeWindow struct {
	StartHour int    `json:"start_hour"`
	EndHour   int    `json:"end_hour"`
	Timezone  string `json:"timezone"`
}

func (w ModelFingerprintTimeWindow) location() (*time.Location, error) {
	if w.StartHour < 0 || w.StartHour > 23 || w.EndHour < 1 || w.EndHour > 24 || w.StartHour >= w.EndHour {
		return nil, apperrors.BadRequest("INVALID_FINGERPRINT_WINDOW", "Test window must have 0 <= start_hour < end_hour <= 24")
	}
	if w.Timezone == "" || w.Timezone == "Local" {
		return nil, apperrors.BadRequest("INVALID_FINGERPRINT_WINDOW", "Select a valid test timezone")
	}
	loc, err := time.LoadLocation(w.Timezone)
	if err != nil {
		return nil, apperrors.BadRequest("INVALID_FINGERPRINT_WINDOW", "Select a valid test timezone")
	}
	return loc, nil
}

func (w ModelFingerprintTimeWindow) Contains(now time.Time) bool {
	loc, err := w.location()
	return err == nil && now.In(loc).Hour() >= w.StartHour && now.In(loc).Hour() < w.EndHour
}

func (w ModelFingerprintTimeWindow) NextRun(now time.Time) (time.Time, error) {
	loc, err := w.location()
	if err != nil {
		return time.Time{}, err
	}
	// Iterate real minutes so half-hour offsets and daylight-saving transitions
	// still land on a valid local whole hour, strictly after now.
	for next, limit := now.Truncate(time.Minute).Add(time.Minute), now.Add(72*time.Hour); next.Before(limit); next = next.Add(time.Minute) {
		local := next.In(loc)
		if local.Minute() == 0 && local.Hour() >= w.StartHour && local.Hour() < w.EndHour {
			return next.UTC(), nil
		}
	}
	return time.Time{}, errors.New("no fingerprint test slot available")
}

func (s *AccountTestService) normalizeFingerprintWindow(schedule *ModelFingerprintSchedule) {
	if schedule.Window == nil {
		schedule.Window = &ModelFingerprintTimeWindow{EndHour: 24}
	}
	if schedule.Window.Timezone == "" {
		schedule.Window.Timezone = "Asia/Shanghai"
		if s.cfg != nil && s.cfg.Timezone != "" {
			schedule.Window.Timezone = s.cfg.Timezone
		}
	}
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
	schedule, err := store.GetModelFingerprintSchedule(ctx, id)
	if err != nil {
		return nil, err
	}
	legacy := schedule.Window == nil
	s.normalizeFingerprintWindow(schedule)
	// Round old half-hour plans to the first eligible hourly slot.
	if schedule.Enabled {
		window := schedule.Window
		if legacy {
			// Existing plans stay UTC-aligned until the displayed local window is saved.
			window = &ModelFingerprintTimeWindow{EndHour: 24, Timezone: "UTC"}
		}
		at := schedule.NextRunAt.Add(-time.Nanosecond)
		if now := time.Now().UTC(); at.Before(now) && !window.Contains(now) {
			at = now
		}
		calculation := *schedule
		calculation.Window = window
		schedule.NextRunAt, err = calculation.NextRun(at)
	}
	if schedule.TestsPerDay == 0 {
		schedule.TestsPerDay = schedule.Window.EndHour - schedule.Window.StartHour
	}
	return schedule, err
}

func (s *AccountTestService) SetFingerprintSchedule(ctx context.Context, schedule *ModelFingerprintSchedule) error {
	s.normalizeFingerprintWindow(schedule)
	if _, err := schedule.Window.location(); err != nil {
		return err
	}
	if schedule.TestsPerDay == 0 {
		schedule.TestsPerDay = schedule.Window.EndHour - schedule.Window.StartHour
	}
	next, err := schedule.NextRun(time.Now().UTC())
	if err != nil {
		return err
	}
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
	schedule.NextRunAt = next
	return store.SetModelFingerprintSchedule(ctx, schedule)
}

// Existing minute scheduler owns the lifecycle. Claims advance the next slot
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
