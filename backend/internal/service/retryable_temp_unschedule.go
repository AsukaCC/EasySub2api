package service

import (
	"context"
	"log"
	"time"
)

const retryableErrorCooldown = time.Minute

func tempUnscheduleRetryableBadRequest(ctx context.Context, repo AccountRepository, accountID, logPrefix string) {
	tempUnscheduleAccount(ctx, repo, accountID, logPrefix, "retryable upstream bad request")
}

func tempUnscheduleEmptyResponse(ctx context.Context, repo AccountRepository, accountID, logPrefix string) {
	tempUnscheduleAccount(ctx, repo, accountID, logPrefix, "empty stream response")
}

func tempUnscheduleAccount(ctx context.Context, repo AccountRepository, accountID, logPrefix, reason string) {
	until := time.Now().Add(retryableErrorCooldown)
	if err := repo.SetTempUnschedulable(ctx, accountID, until, reason); err != nil {
		log.Printf("%s temp_unschedule_failed account=%v error=%v", logPrefix, accountID, err)
		return
	}
	log.Printf("%s temp_unscheduled account=%v until=%v reason=%q", logPrefix, accountID, until.Format("15:04:05"), reason)
}
