//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

type balanceUserRepoStub struct {
	*userRepoStub
	adjustErr error
	// changes 记录每次原子余额变更，顺序与调用顺序一致。
	changes      []BalanceChange
	credits      []WalletCreditInput
	pending      map[*dbent.Tx]*User
	failCreditAt int
}

func (s *balanceUserRepoStub) walletUser(ctx context.Context) *User {
	tx := dbent.TxFromContext(ctx)
	if tx == nil {
		return s.user
	}
	if s.pending == nil {
		s.pending = make(map[*dbent.Tx]*User)
	}
	if s.pending[tx] == nil {
		copy := *s.user
		s.pending[tx] = &copy
		tx.OnCommit(func(next dbent.Committer) dbent.Committer {
			return dbent.CommitFunc(func(ctx context.Context, tx *dbent.Tx) error {
				if err := next.Commit(ctx, tx); err != nil {
					return err
				}
				*s.user = *s.pending[tx]
				return nil
			})
		})
	}
	return s.pending[tx]
}

func (s *balanceUserRepoStub) GetByID(ctx context.Context, _ string) (*User, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	copy := *s.walletUser(ctx)
	return &copy, nil
}

func (s *balanceUserRepoStub) GetWalletSummary(ctx context.Context, _ string) (WalletSummary, error) {
	u := s.walletUser(ctx)
	return NewWalletSummary(u.Balance, u.BonusBalance, 0, 0), nil
}

func (s *balanceUserRepoStub) CreditWallet(ctx context.Context, input WalletCreditInput) (WalletMutationResult, error) {
	s.credits = append(s.credits, input)
	if s.failCreditAt == len(s.credits) {
		return WalletMutationResult{}, errors.New("credit failed")
	}
	_, err := s.apply(ctx, func(current float64) float64 { return current + input.Amount })
	if err == nil && input.Kind == WalletKindBonus {
		s.walletUser(ctx).BonusBalance += input.Amount
	}
	return WalletMutationResult{Applied: err == nil}, err
}

func (s *balanceUserRepoStub) DebitWallet(ctx context.Context, input WalletDebitInput) (WalletMutationResult, error) {
	_, err := s.AdjustBalance(ctx, input.UserID, -input.Amount)
	return WalletMutationResult{Applied: err == nil}, err
}

func (s *balanceUserRepoStub) SetWalletBalance(ctx context.Context, input WalletSetInput) (WalletMutationResult, error) {
	_, err := s.SetBalance(ctx, input.UserID, input.RechargeAmount+input.BonusAmount)
	return WalletMutationResult{Applied: err == nil}, err
}

func (s *balanceUserRepoStub) AdjustBalance(ctx context.Context, id string, delta float64) (BalanceChange, error) {
	return s.apply(ctx, func(current float64) float64 { return current + delta })
}

func (s *balanceUserRepoStub) SetBalance(ctx context.Context, id string, value float64) (BalanceChange, error) {
	return s.apply(ctx, func(float64) float64 { return value })
}

func (s *balanceUserRepoStub) apply(ctx context.Context, next func(current float64) float64) (BalanceChange, error) {
	if s.adjustErr != nil {
		return BalanceChange{}, s.adjustErr
	}
	if s.userRepoStub == nil || s.userRepoStub.user == nil {
		return BalanceChange{}, ErrUserNotFound
	}
	u := s.walletUser(ctx)
	change := BalanceChange{Old: u.Balance}
	change.New = next(change.Old)
	if change.New < 0 {
		return change, ErrBalanceNegative
	}
	u.Balance = change.New
	s.changes = append(s.changes, change)
	return change, nil
}

type balanceRedeemRepoStub struct {
	*redeemRepoStub
	created []*RedeemCode
}

func (s *balanceRedeemRepoStub) Create(ctx context.Context, code *RedeemCode) error {
	if code == nil {
		return nil
	}
	clone := *code
	s.created = append(s.created, &clone)
	return nil
}

type authCacheInvalidatorStub struct {
	userIDs  []string
	groupIDs []string
	keys     []string
}

type adminRechargeAffiliateAccruerStub struct {
	calls  []adminRechargeAffiliateAccrual
	rebate float64
	err    error
}

type adminRechargeAffiliateAccrual struct {
	userID string
	amount float64
}

func (s *adminRechargeAffiliateAccruerStub) AccrueInviteRebate(_ context.Context, userID string, amount float64) (float64, error) {
	s.calls = append(s.calls, adminRechargeAffiliateAccrual{userID: userID, amount: amount})
	return s.rebate, s.err
}

func adminRechargeSettingService(enabled bool) *SettingService {
	values := map[string]string{}
	if enabled {
		values[SettingKeyAffiliateAdminRechargeEnabled] = "true"
	}
	return NewSettingService(&settingRepoStub{values: values}, nil)
}

func (s *authCacheInvalidatorStub) InvalidateAuthCacheByKey(ctx context.Context, key string) {
	s.keys = append(s.keys, key)
}

func (s *authCacheInvalidatorStub) InvalidateAuthCacheByUserID(ctx context.Context, userID string) {
	s.userIDs = append(s.userIDs, userID)
}

func (s *authCacheInvalidatorStub) InvalidateAuthCacheByGroupID(ctx context.Context, groupID string) {
	s.groupIDs = append(s.groupIDs, groupID)
}

// 管理员调账必须走原子的 AdjustBalance/SetBalance，而不是"读余额→算新值→整行写回"，
// 后者会把并发的计费扣款覆盖掉。userRepoStub.Update 对未预期的调用会 panic，
// 因此这里同时证明它没被走到。
func TestAdminService_UpdateUserBalance_UsesAtomicPrimitives(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		amount    float64
		want      BalanceChange
	}{
		{name: "add", operation: "add", amount: 5, want: BalanceChange{Old: 10, New: 15}},
		{name: "subtract", operation: "subtract", amount: 4, want: BalanceChange{Old: 10, New: 6}},
		{name: "set", operation: "set", amount: 2, want: BalanceChange{Old: 10, New: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &balanceUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: "7", Balance: 10}}}
			svc := &adminServiceImpl{
				userRepo:       repo,
				redeemCodeRepo: &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}},
			}

			user, err := svc.UpdateUserBalance(context.Background(), "7", tt.amount, tt.operation, "")
			require.NoError(t, err)
			require.Equal(t, []BalanceChange{tt.want}, repo.changes)
			require.Equal(t, tt.want.New, user.Balance)
		})
	}
}

func TestAdminService_UpdateUserBalance_RejectsNegativeResult(t *testing.T) {
	repo := &balanceUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: "7", Balance: 3}}}
	svc := &adminServiceImpl{
		userRepo:       repo,
		redeemCodeRepo: &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}},
	}

	_, err := svc.UpdateUserBalance(context.Background(), "7", 4, "subtract", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "balance cannot be negative")
	require.Empty(t, repo.changes, "refused adjustment must not be applied")
	require.Equal(t, 3.0, repo.userRepoStub.user.Balance)
}

func TestAdminService_UpdateUserBalance_RejectsUnknownOperation(t *testing.T) {
	repo := &balanceUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: "7", Balance: 10}}}
	svc := &adminServiceImpl{
		userRepo:       repo,
		redeemCodeRepo: &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}},
	}

	_, err := svc.UpdateUserBalance(context.Background(), "7", 1, "multiply", "")
	require.Error(t, err)
	require.Empty(t, repo.changes)
}

func TestAdminService_UpdateUserBalance_InvalidatesAuthCache(t *testing.T) {
	baseRepo := &userRepoStub{user: &User{ID: "7", Balance: 10}}
	repo := &balanceUserRepoStub{userRepoStub: baseRepo}
	redeemRepo := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       redeemRepo,
		authCacheInvalidator: invalidator,
	}

	_, err := svc.UpdateUserBalance(context.Background(), "7", 5, "add", "")
	require.NoError(t, err)
	require.Equal(t, []string{"7"}, invalidator.userIDs)
	require.Len(t, redeemRepo.created, 1)
}

func TestAdminService_UpdateUserBalance_NoChangeNoInvalidate(t *testing.T) {
	baseRepo := &userRepoStub{user: &User{ID: "7", Balance: 10}}
	repo := &balanceUserRepoStub{userRepoStub: baseRepo}
	redeemRepo := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       redeemRepo,
		authCacheInvalidator: invalidator,
	}

	_, err := svc.UpdateUserBalance(context.Background(), "7", 10, "set", "")
	require.NoError(t, err)
	require.Empty(t, invalidator.userIDs)
	require.Empty(t, redeemRepo.created)
}

func TestAdminService_UpdateUserBalance_NeverAccruesAffiliateRebate(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		operation string
		amount    float64
		wantCalls []adminRechargeAffiliateAccrual
	}{
		{
			name:      "disabled by default",
			operation: "add",
			amount:    5,
		},
		{
			name:      "enabled add",
			enabled:   true,
			operation: "add",
			amount:    0.1,
		},
		{
			name:      "enabled set increase",
			enabled:   true,
			operation: "set",
			amount:    15,
		},
		{
			name:      "enabled subtract",
			enabled:   true,
			operation: "subtract",
			amount:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseRepo := &userRepoStub{user: &User{ID: "7", Balance: 10}}
			repo := &balanceUserRepoStub{userRepoStub: baseRepo}
			redeemRepo := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
			affiliate := &adminRechargeAffiliateAccruerStub{}
			svc := &adminServiceImpl{
				userRepo:         repo,
				redeemCodeRepo:   redeemRepo,
				settingService:   adminRechargeSettingService(tt.enabled),
				affiliateService: affiliate,
			}

			_, err := svc.UpdateUserBalance(context.Background(), "7", tt.amount, tt.operation, "")
			require.NoError(t, err)
			require.Equal(t, tt.wantCalls, affiliate.calls)
		})
	}
}

func TestAdminService_UpdateUserBalance_LegacyAffiliateSettingDoesNotAccrue(t *testing.T) {
	baseRepo := &userRepoStub{user: &User{ID: "7", Balance: 10}}
	repo := &balanceUserRepoStub{userRepoStub: baseRepo}
	redeemRepo := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	affiliate := &adminRechargeAffiliateAccruerStub{err: errors.New("affiliate unavailable")}
	svc := &adminServiceImpl{
		userRepo:         repo,
		redeemCodeRepo:   redeemRepo,
		settingService:   adminRechargeSettingService(true),
		affiliateService: affiliate,
	}

	user, err := svc.UpdateUserBalance(context.Background(), "7", 5, "add", "")
	require.NoError(t, err)
	require.Equal(t, 15.0, user.Balance)
	require.Empty(t, affiliate.calls)
	require.Len(t, redeemRepo.created, 1)
}

func TestAdminService_RechargeBonusTiers(t *testing.T) {
	for _, tc := range []struct {
		name, operation, kind string
		amount, bonus         float64
	}{
		{"below threshold", "add", WalletKindRecharge, 49.99, 0},
		{"at threshold", "add", WalletKindRecharge, 50, 3},
		{"highest threshold only", "add", WalletKindRecharge, 200, 12.12345678},
		{"above threshold", "add", WalletKindRecharge, 500, 12.12345678},
		{"manual bonus", "add", WalletKindBonus, 200, 0},
		{"set balance", "set", WalletKindRecharge, 200, 0},
		{"subtract", "subtract", WalletKindRecharge, 50, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
			t.Cleanup(func() { _ = client.Close() })
			repo := &balanceUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: "7", Balance: 100}}}
			audit := &balanceRedeemRepoStub{}
			invalidator := &authCacheInvalidatorStub{}
			svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: audit, entClient: client, authCacheInvalidator: invalidator,
				settingService: NewSettingService(&settingRepoStub{values: map[string]string{SettingRechargeBonusTiers: `[{"threshold_cny":200,"bonus_points":12.12345678},{"threshold_cny":50,"bonus_points":3}]`}}, nil)}
			if tc.operation == "add" && tc.kind == WalletKindRecharge {
				mock.ExpectBegin()
				mock.ExpectCommit()
			}
			started := time.Now().UTC()
			user, err := svc.UpdateUserWalletBalance(context.Background(), "7", tc.amount, tc.operation, tc.kind, 30, "admin test")
			require.NoError(t, err)
			if tc.operation == "add" {
				require.InDelta(t, 100+tc.amount+tc.bonus, user.Balance, 1e-8)
				if tc.bonus > 0 {
					require.Len(t, repo.credits, 2)
					base, bonus := repo.credits[0], repo.credits[1]
					require.Equal(t, tc.bonus, bonus.Amount)
					require.Equal(t, WalletKindBonus, bonus.Kind)
					require.Equal(t, base.SourceID, bonus.SourceID)
					require.NotEqual(t, base.IdempotencyKey, bonus.IdempotencyKey)
					require.WithinDuration(t, started.Add(rechargeBonusValidity), *bonus.ExpiresAt, time.Second)
					require.InDelta(t, tc.amount+tc.bonus, audit.created[0].Value, 1e-8)
				} else {
					require.Len(t, repo.credits, 1)
				}
			}
			require.Equal(t, []string{"7"}, invalidator.userIDs)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAdminService_RechargeBonusRollsBackBeforeRetry(t *testing.T) {
	for _, failure := range []string{"principal", "bonus", "response read"} {
		t.Run(failure, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
			t.Cleanup(func() { _ = client.Close() })
			repo := &balanceUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: "7", Balance: 10}}}
			switch failure {
			case "principal":
				repo.failCreditAt = 1
			case "bonus":
				repo.failCreditAt = 2
			case "response read":
				repo.getErr = errors.New("read failed")
			}
			audit := &balanceRedeemRepoStub{}
			invalidator := &authCacheInvalidatorStub{}
			svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: audit, entClient: client, authCacheInvalidator: invalidator,
				settingService: NewSettingService(&settingRepoStub{values: map[string]string{SettingRechargeBonusTiers: `[{"threshold_cny":50,"bonus_points":3}]`}}, nil)}
			mock.ExpectBegin()
			mock.ExpectRollback()
			_, err = svc.UpdateUserBalance(context.Background(), "7", 50, "add", "")
			require.Error(t, err)
			require.Equal(t, 10.0, repo.user.Balance)
			require.Zero(t, repo.user.BonusBalance)
			require.Empty(t, audit.created)
			require.Empty(t, invalidator.userIDs)
			repo.failCreditAt, repo.getErr = 0, nil
			mock.ExpectBegin()
			mock.ExpectCommit()
			_, err = svc.UpdateUserBalance(context.Background(), "7", 50, "add", "")
			require.NoError(t, err)
			require.Equal(t, 63.0, repo.user.Balance)
			require.Equal(t, 3.0, repo.user.BonusBalance)
			require.Len(t, audit.created, 1)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAdminService_RechargeBonusSettingsFailureDoesNotCredit(t *testing.T) {
	repo := &balanceUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: "7", Balance: 10}}}
	svc := &adminServiceImpl{userRepo: repo, settingService: NewSettingService(&settingRepoStub{err: errors.New("settings unavailable")}, nil)}
	_, err := svc.UpdateUserBalance(context.Background(), "7", 50, "add", "")
	require.ErrorContains(t, err, "load recharge bonus tiers")
	require.Empty(t, repo.credits)
	require.Equal(t, 10.0, repo.user.Balance)
}
