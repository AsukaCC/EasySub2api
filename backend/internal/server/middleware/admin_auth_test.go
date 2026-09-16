//go:build unit

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminAuthJWTValidatesTokenVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret", ExpireHour: 1}}
	authService := service.NewAuthService(nil, nil, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil)

	admin := &service.User{
		ID:           "user-1",
		Email:        "admin@example.com",
		Role:         service.RoleAdmin,
		Status:       service.StatusActive,
		TokenVersion: 2,
		Concurrency:  1,
	}

	userRepo := &stubUserRepo{
		getByID: func(ctx context.Context, id string) (*service.User, error) {
			if id != admin.ID {
				return nil, service.ErrUserNotFound
			}
			clone := *admin
			return &clone, nil
		},
	}
	userService := service.NewUserService(userRepo, nil, nil, nil)

	router := gin.New()
	router.Use(gin.HandlerFunc(NewAdminAuthMiddleware(authService, userService, nil, nil)))
	router.GET("/t", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	t.Run("token_version_mismatch_rejected", func(t *testing.T) {
		token, err := authService.GenerateToken(context.Background(), &service.User{
			ID:           admin.ID,
			Email:        admin.Email,
			Role:         admin.Role,
			TokenVersion: admin.TokenVersion - 1,
		})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Body.String(), "TOKEN_REVOKED")
	})

	t.Run("token_version_match_allows", func(t *testing.T) {
		token, err := authService.GenerateToken(context.Background(), &service.User{
			ID:           admin.ID,
			Email:        admin.Email,
			Role:         admin.Role,
			TokenVersion: admin.TokenVersion,
		})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("websocket_token_version_mismatch_rejected", func(t *testing.T) {
		token, err := authService.GenerateToken(context.Background(), &service.User{
			ID:           admin.ID,
			Email:        admin.Email,
			Role:         admin.Role,
			TokenVersion: admin.TokenVersion - 1,
		})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Sec-WebSocket-Protocol", "sub2api-admin, jwt."+token)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Body.String(), "TOKEN_REVOKED")
	})

	t.Run("websocket_token_version_match_allows", func(t *testing.T) {
		token, err := authService.GenerateToken(context.Background(), &service.User{
			ID:           admin.ID,
			Email:        admin.Email,
			Role:         admin.Role,
			TokenVersion: admin.TokenVersion,
		})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Sec-WebSocket-Protocol", "sub2api-admin, jwt."+token)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

type stubUserRepo struct {
	getByID func(ctx context.Context, id string) (*service.User, error)
}

func (s *stubUserRepo) Create(ctx context.Context, user *service.User) error {
	panic("unexpected Create call")
}

func (s *stubUserRepo) CreateWithEmailAliasGuard(ctx context.Context, user *service.User) error {
	panic("unexpected CreateWithEmailAliasGuard call")
}

func (s *stubUserRepo) GetByID(ctx context.Context, id string) (*service.User, error) {
	if s.getByID == nil {
		panic("GetByID not stubbed")
	}
	return s.getByID(ctx, id)
}

func (s *stubUserRepo) GetByEmail(ctx context.Context, email string) (*service.User, error) {
	panic("unexpected GetByEmail call")
}

func (s *stubUserRepo) GetFirstAdmin(ctx context.Context) (*service.User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (s *stubUserRepo) Update(ctx context.Context, user *service.User, fields service.UserUpdateFields) error {
	panic("unexpected Update call")
}

func (s *stubUserRepo) Delete(ctx context.Context, id string) error {
	panic("unexpected Delete call")
}

func (s *stubUserRepo) GetUserAvatar(ctx context.Context, userID string) (*service.UserAvatar, error) {
	return nil, nil
}

func (s *stubUserRepo) UpsertUserAvatar(ctx context.Context, userID string, input service.UpsertUserAvatarInput) (*service.UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}

func (s *stubUserRepo) DeleteUserAvatar(ctx context.Context, userID string) error {
	panic("unexpected DeleteUserAvatar call")
}

func (s *stubUserRepo) List(ctx context.Context, params pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *stubUserRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *stubUserRepo) GetLatestUsedAtByUserIDs(ctx context.Context, userIDs []string) (map[string]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs call")
}

func (s *stubUserRepo) GetLatestUsedAtByUserID(ctx context.Context, userID string) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID call")
}

func (s *stubUserRepo) UpdateUserLastActiveAt(ctx context.Context, userID string, activeAt time.Time) error {
	panic("unexpected UpdateUserLastActiveAt call")
}

func (s *stubUserRepo) UpdateBalance(ctx context.Context, id string, amount float64) error {
	panic("unexpected UpdateBalance call")
}

func (s *stubUserRepo) DeductBalance(ctx context.Context, id string, amount float64) error {
	panic("unexpected DeductBalance call")
}

func (s *stubUserRepo) AdjustBalance(ctx context.Context, id string, delta float64) (service.BalanceChange, error) {
	panic("unexpected AdjustBalance call")
}

func (s *stubUserRepo) SetBalance(ctx context.Context, id string, value float64) (service.BalanceChange, error) {
	panic("unexpected SetBalance call")
}

func (s *stubUserRepo) GetWalletSummary(context.Context, string) (service.WalletSummary, error) {
	panic("unexpected GetWalletSummary call")
}
func (s *stubUserRepo) CreditWallet(context.Context, service.WalletCreditInput) (service.WalletMutationResult, error) {
	panic("unexpected CreditWallet call")
}
func (s *stubUserRepo) DebitWallet(context.Context, service.WalletDebitInput) (service.WalletMutationResult, error) {
	panic("unexpected DebitWallet call")
}
func (s *stubUserRepo) SetWalletBalance(context.Context, service.WalletSetInput) (service.WalletMutationResult, error) {
	panic("unexpected SetWalletBalance call")
}
func (s *stubUserRepo) HoldWallet(context.Context, service.WalletHoldInput) (service.WalletHoldResult, error) {
	panic("unexpected HoldWallet call")
}
func (s *stubUserRepo) CaptureWalletHold(context.Context, string, string) (service.WalletHoldResult, error) {
	panic("unexpected CaptureWalletHold call")
}
func (s *stubUserRepo) ReleaseWalletHold(context.Context, string, string) (service.WalletHoldResult, error) {
	panic("unexpected ReleaseWalletHold call")
}
func (s *stubUserRepo) RefundWalletHold(context.Context, string, float64, string) (service.WalletMutationResult, error) {
	panic("unexpected RefundWalletHold call")
}
func (s *stubUserRepo) ListWalletTransactions(context.Context, string, int, int) (service.WalletTransactionPage, error) {
	panic("unexpected ListWalletTransactions call")
}
func (s *stubUserRepo) ExpireBonusBalances(context.Context, int) ([]string, error) {
	panic("unexpected ExpireBonusBalances call")
}

func (s *stubUserRepo) UpdateConcurrency(ctx context.Context, id string, amount int) error {
	panic("unexpected UpdateConcurrency call")
}

func (s *stubUserRepo) BatchSetConcurrency(context.Context, []string, int) (int, error) {
	return 0, nil
}
func (s *stubUserRepo) BatchAddConcurrency(context.Context, []string, int) (int, error) {
	return 0, nil
}
func (s *stubUserRepo) BatchUpdateLimits(context.Context, []string, *int, *int) (int, error) {
	return 0, nil
}

func (s *stubUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}

func (s *stubUserRepo) ExistsByEmailAlias(ctx context.Context, email string) (bool, error) {
	panic("unexpected ExistsByEmailAlias call")
}

func (s *stubUserRepo) RemoveGroupFromAllowedGroups(ctx context.Context, groupID string) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}

func (s *stubUserRepo) RemoveGroupFromUserAllowedGroups(ctx context.Context, userID string, groupID string) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}

func (s *stubUserRepo) AddGroupToAllowedGroups(ctx context.Context, userID string, groupID string) error {
	panic("unexpected AddGroupToAllowedGroups call")
}

func (s *stubUserRepo) ListUserAuthIdentities(ctx context.Context, userID string) ([]service.UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities call")
}

func (s *stubUserRepo) UnbindUserAuthProvider(context.Context, string, string) error {
	panic("unexpected UnbindUserAuthProvider call")
}

func (s *stubUserRepo) UpdateTotpSecret(ctx context.Context, userID string, encryptedSecret *string) error {
	panic("unexpected UpdateTotpSecret call")
}

func (s *stubUserRepo) EnableTotp(ctx context.Context, userID string) error {
	panic("unexpected EnableTotp call")
}

func (s *stubUserRepo) DisableTotp(ctx context.Context, userID string) error {
	panic("unexpected DisableTotp call")
}

func (s *stubUserRepo) GetByIDIncludeDeleted(ctx context.Context, id string) (*service.User, error) {
	panic("unexpected GetByIDIncludeDeleted call")
}
