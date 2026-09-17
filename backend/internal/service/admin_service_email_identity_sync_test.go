//go:build unit

package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type ensureEmailCall struct {
	userID string
	email  string
}

type replaceEmailCall struct {
	userID   string
	oldEmail string
	newEmail string
}

type emailSyncRepoStub struct {
	WalletRepository
	user         *User
	nextID       int64
	updateCalls  int
	created      []*User
	updated      []*User
	ensureCalls  []ensureEmailCall
	replaceCalls []replaceEmailCall
	ensureErr    error
	replaceErr   error
}

func (s *emailSyncRepoStub) CreateWithEmailAliasGuard(ctx context.Context, user *User) error {
	return s.Create(ctx, user)
}

func (s *emailSyncRepoStub) Create(_ context.Context, user *User) error {
	if s.nextID != 0 && user.ID == "" {
		user.ID = strconv.FormatInt(s.nextID, 10)
	}
	s.created = append(s.created, user)
	s.user = user
	return nil
}

func (s *emailSyncRepoStub) GetByID(_ context.Context, _ string) (*User, error) {
	if s.user == nil {
		return nil, ErrUserNotFound
	}
	cloned := *s.user
	return &cloned, nil
}

func (s *emailSyncRepoStub) GetByEmail(_ context.Context, _ string) (*User, error) {
	return nil, ErrUserNotFound
}

func (s *emailSyncRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	return nil, fmt.Errorf("unexpected GetFirstAdmin call")
}

func (s *emailSyncRepoStub) Update(_ context.Context, user *User, _ UserUpdateFields) error {
	s.updateCalls++
	s.updated = append(s.updated, user)
	s.user = user
	return nil
}

func (s *emailSyncRepoStub) Delete(context.Context, string) error { return nil }

func (s *emailSyncRepoStub) GetUserAvatar(context.Context, string) (*UserAvatar, error) {
	return nil, fmt.Errorf("unexpected GetUserAvatar call")
}

func (s *emailSyncRepoStub) UpsertUserAvatar(context.Context, string, UpsertUserAvatarInput) (*UserAvatar, error) {
	return nil, fmt.Errorf("unexpected UpsertUserAvatar call")
}

func (s *emailSyncRepoStub) DeleteUserAvatar(context.Context, string) error {
	return fmt.Errorf("unexpected DeleteUserAvatar call")
}

func (s *emailSyncRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, fmt.Errorf("unexpected List call")
}

func (s *emailSyncRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, fmt.Errorf("unexpected ListWithFilters call")
}

func (s *emailSyncRepoStub) GetLatestUsedAtByUserIDs(context.Context, []string) (map[string]*time.Time, error) {
	return map[string]*time.Time{}, nil
}

func (s *emailSyncRepoStub) GetLatestUsedAtByUserID(context.Context, string) (*time.Time, error) {
	return nil, nil
}

func (s *emailSyncRepoStub) UpdateUserLastActiveAt(context.Context, string, time.Time) error {
	return nil
}

func (s *emailSyncRepoStub) UpdateBalance(context.Context, string, float64) error { return nil }

func (s *emailSyncRepoStub) DeductBalance(context.Context, string, float64) error { return nil }

func (s *emailSyncRepoStub) UpdateConcurrency(context.Context, string, int) error { return nil }

func (s *emailSyncRepoStub) ExistsByEmail(context.Context, string) (bool, error) { return false, nil }

func (s *emailSyncRepoStub) ExistsByEmailAlias(context.Context, string) (bool, error) {
	return false, nil
}

func (s *emailSyncRepoStub) AdjustBalance(ctx context.Context, id string, delta float64) (BalanceChange, error) {
	panic("unexpected AdjustBalance call")
}

func (s *emailSyncRepoStub) SetBalance(ctx context.Context, id string, value float64) (BalanceChange, error) {
	panic("unexpected SetBalance call")
}

func (s *emailSyncRepoStub) RemoveGroupFromAllowedGroups(context.Context, string) (int64, error) {
	return 0, nil
}

func (s *emailSyncRepoStub) BatchSetConcurrency(context.Context, []string, int) (int, error) {
	return 0, nil
}
func (s *emailSyncRepoStub) BatchAddConcurrency(context.Context, []string, int) (int, error) {
	return 0, nil
}
func (s *emailSyncRepoStub) BatchUpdateLimits(context.Context, []string, *int, *int) (int, error) {
	return 0, nil
}

func (s *emailSyncRepoStub) AddGroupToAllowedGroups(context.Context, string, string) error { return nil }

func (s *emailSyncRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, string, string) error {
	return nil
}

func (s *emailSyncRepoStub) ListUserAuthIdentities(context.Context, string) ([]UserAuthIdentityRecord, error) {
	return nil, nil
}

func (s *emailSyncRepoStub) UnbindUserAuthProvider(context.Context, string, string) error { return nil }

func (s *emailSyncRepoStub) UpdateTotpSecret(context.Context, string, *string) error { return nil }

func (s *emailSyncRepoStub) EnableTotp(context.Context, string) error { return nil }

func (s *emailSyncRepoStub) DisableTotp(context.Context, string) error { return nil }
func (s *emailSyncRepoStub) GetByIDIncludeDeleted(ctx context.Context, id string) (*User, error) {
	return s.GetByID(ctx, id)
}

func (s *emailSyncRepoStub) EnsureEmailAuthIdentity(_ context.Context, userID string, email string) error {
	s.ensureCalls = append(s.ensureCalls, ensureEmailCall{userID: userID, email: email})
	return s.ensureErr
}

func (s *emailSyncRepoStub) ReplaceEmailAuthIdentity(_ context.Context, userID string, oldEmail, newEmail string) error {
	s.replaceCalls = append(s.replaceCalls, replaceEmailCall{
		userID:   userID,
		oldEmail: oldEmail,
		newEmail: newEmail,
	})
	return s.replaceErr
}

func TestAdminService_CreateUser_DoesNotReturnPartialSuccessFromEmailIdentityResync(t *testing.T) {
	repo := &emailSyncRepoStub{
		nextID:    55,
		ensureErr: fmt.Errorf("unexpected email resync"),
	}
	svc := &adminServiceImpl{userRepo: repo}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "admin-created@example.com",
		Password: "strong-pass",
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "55", user.ID)
	require.Empty(t, repo.ensureCalls)
	require.Empty(t, repo.replaceCalls)
}

func TestAdminService_UpdateUser_DoesNotReturnPartialSuccessFromEmailIdentityResync(t *testing.T) {
	repo := &emailSyncRepoStub{
		user: &User{
			ID: "91",
			Email:       "before@example.com",
			Role:        RoleUser,
			Status:      StatusActive,
			Concurrency: 3,
		},
		replaceErr: fmt.Errorf("unexpected email resync"),
	}
	svc := &adminServiceImpl{userRepo: repo}

	updated, err := svc.UpdateUser(context.Background(), "91", &UpdateUserInput{
		Email: "after@example.com",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, "after@example.com", updated.Email)
	require.Empty(t, repo.replaceCalls)
	require.Empty(t, repo.ensureCalls)
}
