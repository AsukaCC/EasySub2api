//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountLoginCredentialsValidation(t *testing.T) {
	for _, value := range []string{
		"JBSWY3DPEHPK3PXP", "jbsw y3dp-ehpk 3pxp", "MY======",
		"otpauth://totp/Test?secret=JBSWY3DPEHPK3PXP&algorithm=SHA256&digits=8&period=60",
	} {
		credentials := map[string]any{AccountLoginUsernameKey: "person@example.com", AccountLoginPasswordKey: " password with spaces ", AccountLoginTOTPKey: value}
		require.NoError(t, NormalizeAccountLoginCredentials(credentials))
		require.Equal(t, " password with spaces ", credentials[AccountLoginPasswordKey])
		require.Equal(t, credentials, SanitizeStoredCredentials(PlatformGrok, credentials))
	}
	for _, value := range []any{
		true, 123, "invalid!", "A", "MZ", strings.Repeat("A", 4097),
		"otpauth://hotp/Test?secret=JBSWY3DPEHPK3PXP",
		"https://example.com/?secret=JBSWY3DPEHPK3PXP",
		"otpauth://totp/Test?secret=JBSWY3DPEHPK3PXP&period=0",
		"otpauth://totp/Test?secret=JBSWY3DPEHPK3PXP&period=301",
		"otpauth://totp/Test?secret=JBSWY3DPEHPK3PXP&algorithm=MD5",
		"otpauth://totp/Test?secret=JBSWY3DPEHPK3PXP&digits=7",
		"otpauth://totp/Test?secret=JBSWY3DPEHPK3PXP&secret=MY",
	} {
		err := NormalizeAccountLoginCredentials(map[string]any{AccountLoginTOTPKey: value})
		require.Error(t, err)
		require.NotContains(t, err.Error(), "JBSWY3DPEHPK3PXP")
	}
}

func TestAccountLoginCredentialsPreservedAndExplicitlyCleared(t *testing.T) {
	repo := &updateAccountCredsRepoStub{account: &Account{ID: "one", Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive,
		Credentials: map[string]any{"access_token": "token", AccountLoginUsernameKey: "login", AccountLoginPasswordKey: "password", AccountLoginTOTPKey: "JBSWY3DPEHPK3PXP"}}}
	svc := &adminServiceImpl{accountRepo: repo}
	_, err := svc.UpdateAccount(context.Background(), "one", &UpdateAccountInput{Credentials: map[string]any{"base_url": "https://example.com"}})
	require.NoError(t, err)
	require.Equal(t, "password", repo.account.GetCredential(AccountLoginPasswordKey))
	require.Equal(t, "JBSWY3DPEHPK3PXP", repo.account.GetCredential(AccountLoginTOTPKey))
	_, err = svc.UpdateAccount(context.Background(), "one", &UpdateAccountInput{Credentials: map[string]any{AccountLoginPasswordKey: "", AccountLoginTOTPKey: ""}})
	require.NoError(t, err)
	require.Empty(t, repo.account.GetCredential(AccountLoginPasswordKey))
	require.Empty(t, repo.account.GetCredential(AccountLoginTOTPKey))
	require.Equal(t, "login", repo.account.GetCredential(AccountLoginUsernameKey))
	require.Equal(t, "token", repo.account.GetCredential("access_token"))
}
