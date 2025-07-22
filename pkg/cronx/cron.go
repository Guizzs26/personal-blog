package cronx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/service"
	"github.com/robfig/cron/v3"
)

func StartCleanupCronJob(authService *service.AuthService) error {
	c := cron.New()

	// Schedule cleanup every 1 minute (for testing)
	_, err := c.AddFunc("* * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := authService.CleanupExpiredOrRevokedTokens(ctx); err != nil {
			fmt.Printf("failed to clean up expired/revoked tokens: %v\n", err)
		} else {
			fmt.Println("Expired/revoked tokens cleaned up successfully")
		}
	})

	if err != nil {
		return errors.New("failed to schedule cleanup cron job: " + err.Error())
	}

	c.Start()
	return nil
}
