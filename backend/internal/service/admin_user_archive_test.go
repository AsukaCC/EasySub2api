package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type archivePolicyUserRepoStub struct {
	UserRepository
	user              *User
	hasRecharge       bool
	archived          bool
	permanentlyPurged bool
	restored          bool
}

func (s *archivePolicyUserRepoStub) GetByID(context.Context, string) (*User, error) {
	if s.user == nil {
		return nil, ErrUserNotFound
	}
	copy := *s.user
	if s.restored {
		copy.DeletedAt = nil
	}
	return &copy, nil
}

func (s *archivePolicyUserRepoStub) GetByIDIncludeDeleted(context.Context, string) (*User, error) {
	if s.user == nil {
		return nil, ErrUserNotFound
	}
	copy := *s.user
	return &copy, nil
}

func (s *archivePolicyUserRepoStub) Delete(context.Context, string) error {
	s.archived = true
	return nil
}

func (s *archivePolicyUserRepoStub) HasRechargeRecords(context.Context, string) (bool, error) {
	return s.hasRecharge, nil
}

func (s *archivePolicyUserRepoStub) PermanentlyDeleteUser(context.Context, string) (*InactiveUserPurgeResult, error) {
	s.permanentlyPurged = true
	return &InactiveUserPurgeResult{}, nil
}

func (s *archivePolicyUserRepoStub) RestoreArchivedUser(context.Context, string) error {
	s.restored = true
	return nil
}

func TestDeleteUserWithPolicyArchivesUsersWithRechargeRecords(t *testing.T) {
	repo := &archivePolicyUserRepoStub{
		user:        &User{ID: "0199-user", Email: "paid@example.com", Role: RoleUser},
		hasRecharge: true,
	}
	service := &adminServiceImpl{userRepo: repo}

	result, err := service.DeleteUserWithPolicy(context.Background(), "0199-user")

	require.NoError(t, err)
	require.Equal(t, UserDeletionModeArchived, result.Mode)
	require.True(t, repo.archived)
	require.False(t, repo.permanentlyPurged)
}

func TestDeleteUserWithPolicyPermanentlyDeletesUsersWithoutRechargeRecords(t *testing.T) {
	repo := &archivePolicyUserRepoStub{
		user: &User{ID: "0199-user", Email: "free@example.com", Role: RoleUser},
	}
	service := &adminServiceImpl{userRepo: repo}

	result, err := service.DeleteUserWithPolicy(context.Background(), "0199-user")

	require.NoError(t, err)
	require.Equal(t, UserDeletionModePermanentlyDeleted, result.Mode)
	require.False(t, repo.archived)
	require.True(t, repo.permanentlyPurged)
}

func TestRestoreArchivedUserUsesRepositoryCapability(t *testing.T) {
	deletedAt := time.Date(2026, time.September, 3, 0, 0, 0, 0, time.UTC)
	repo := &archivePolicyUserRepoStub{
		user: &User{ID: "0199-user", Email: "restore@example.com", Role: RoleUser, DeletedAt: &deletedAt},
	}
	service := &adminServiceImpl{userRepo: repo}

	user, err := service.RestoreArchivedUser(context.Background(), "0199-user")

	require.NoError(t, err)
	require.True(t, repo.restored)
	require.Nil(t, user.DeletedAt)
}
