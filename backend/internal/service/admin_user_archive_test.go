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

func (s *archivePolicyUserRepoStub) PermanentlyDeleteUser(context.Context, string) (*UserPurgeResult, error) {
	s.permanentlyPurged = true
	return &UserPurgeResult{}, nil
}

func (s *archivePolicyUserRepoStub) RestoreArchivedUser(context.Context, string) error {
	s.restored = true
	return nil
}

func TestDeleteUserPermanentlyDeletesUsersRegardlessOfRechargeHistory(t *testing.T) {
	repo := &archivePolicyUserRepoStub{
		user: &User{ID: "0199-user", Email: "paid@example.com", Role: RoleUser},
	}
	service := &adminServiceImpl{userRepo: repo}

	err := service.DeleteUser(context.Background(), "0199-user")

	require.NoError(t, err)
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
