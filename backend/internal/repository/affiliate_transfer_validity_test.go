//go:build unit

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestAffiliateTransferValidityReadsCurrentSettings(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	for _, days := range []string{"30", "60"} {
		mock.ExpectQuery(`SELECT key, value FROM settings WHERE key IN`).
			WithArgs(service.SettingKeyAffiliateTransferValidityDays, service.SettingKeyBonusBalanceDefaultValidityDays).
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow(service.SettingKeyAffiliateTransferValidityDays, days).
				AddRow(service.SettingKeyBonusBalanceDefaultValidityDays, "180")).RowsWillBeClosed()
		got, err := affiliateTransferValidityDays(context.Background(), db)
		require.NoError(t, err)
		require.Equal(t, service.ResolveAffiliateTransferValidityDays(days, ""), got)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAffiliateTransferValidityPropagatesReadFailures(t *testing.T) {
	failure := errors.New("settings unavailable")
	for _, kind := range []string{"query", "rows", "scan"} {
		t.Run(kind, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			query := mock.ExpectQuery(`SELECT key, value FROM settings WHERE key IN`)
			switch kind {
			case "query":
				query.WillReturnError(failure)
			case "rows":
				query.WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("key", "30").RowError(0, failure)).RowsWillBeClosed()
			case "scan":
				query.WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("key", nil)).RowsWillBeClosed()
			}
			_, err = affiliateTransferValidityDays(context.Background(), db)
			require.Error(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
