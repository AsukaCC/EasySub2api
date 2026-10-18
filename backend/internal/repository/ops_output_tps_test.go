package repository

import (
	"context"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestOutputTPSNullAndMeasuredSamples(t *testing.T) {
	for _, measured := range []bool{false, true} {
		t.Run(map[bool]string{false: "empty", true: "measured"}[measured], func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			rows := sqlmock.NewRows([]string{"p5", "p10", "p50", "avg", "count"})
			if measured {
				rows.AddRow(2.0, 3.0, 10.0, 15.0, 20)
			} else {
				rows.AddRow(nil, nil, nil, nil, 0)
			}
			mock.ExpectQuery(`(?s)percentile_cont\(0.05\).*NULLIF\(ul.duration_ms, 0\).*ul.image_count = 0.*ul.image_output_tokens = 0`).WillReturnRows(rows)
			result, err := (&opsRepository{db: db}).queryOutputTPS(context.Background(), &service.OpsDashboardFilter{StartTime: time.Now().Add(-time.Hour), EndTime: time.Now()})
			require.NoError(t, err)
			if measured {
				require.Equal(t, int64(20), result.SampleCount)
				require.Equal(t, 10.0, *result.P50)
			} else {
				require.Zero(t, result.SampleCount)
				require.Nil(t, result.P50)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
