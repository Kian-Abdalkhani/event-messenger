package scheduler

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"event-messenger.com/models"
)

// runs at set intervals for sending notifications on event dates
func StartDailyNotifications() {
	// Gather events
	events, err := models.GetEventsForToday()
	if err != nil {
		log.Fatalf("scheduler could not retreive events: %v", err)
		return
	}

	// Ends process if no events are dated for today
	if len(events) == 0 {
		return
	}

	// iterate through each event that is dated for today
	for _, event := range events {
		if event.EmailSent {
			slog.Info(fmt.Sprintf("Skipping event %s - email already sent", event.Name))
			continue
		}
		err := sendEventNotification(&event)
		if err != nil {
			log.Fatalf("Could not sent notification for event: %s to %s: %v", event.Name, event.RecipientEmail, err)
			continue
		}
	}

}

func StartScheduler(hourToRun int) {
	slog.Info(fmt.Sprintf("Scheduler started - will run daily at %d:00", hourToRun))

	// Run immediately at startup
	go func() {
		slog.Debug("Running initial notification check on startup...")
		StartDailyNotifications()
	}()

	// Start the daily scheduler
	go func() {
		for {
			now := time.Now()

			// Calculate next run time
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), hourToRun, 0, 0, 0, now.Location())

			// If we've passed the run time today, schedule for tomorrow
			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			duration := time.Until(nextRun)
			slog.Debug(fmt.Sprintf("Next notification check scheduled for: %s (in %v)", nextRun.Format("2006-01-02 15:04:05"), duration))

			time.Sleep(duration)

			slog.Debug("Running scheduled notification check...")
			StartDailyNotifications()
		}
	}()
}
