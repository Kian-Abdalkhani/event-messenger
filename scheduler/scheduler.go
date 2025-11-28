package scheduler

import (
	"fmt"
	"log/slog"
	"time"

	"event-messenger.com/models"
	"event-messenger.com/utils"
)

// runs at set intervals for sending notifications on event dates
func StartDailyNotifications() {
	// Gather events
	events, err := models.GetActiveEventsForToday()
	if err != nil {
		slog.Error("scheduler could not retreive events", "error", err)
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
			err = event.MarkEventInactive()
			if err != nil {
				slog.Error("Error marking event as inactive", "event_name", event.Name, "error", err)
			}
			continue
		}
		err := sendEventNotification(&event)
		if err != nil {
			slog.Error("Could not sent notification for event", "event_name", event.Name, "recipient", event.RecipientEmail, "error", err)
			err = event.MarkEventInactive()
			if err != nil {
				slog.Error("Error marking event as inactive", "event_name", event.Name, "error", err)
			}
			continue
		}
		err = event.MarkEventInactive()
		if err != nil {
			slog.Error("Error marking event as inactive", "event_name", event.Name, "error", err)
		}
	}

}

func CheckFunnelActive() {
	tc := utils.NewTailscaleClient()
	isActive, err := tc.CheckFunnelActive()
	if err != nil || !isActive {
		slog.Error("Funnel is not currently active, check tailscale configuration", "error", err)
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
			slog.Debug(fmt.Sprintf("Next notification check scheduled for: %s (in %v)", nextRun.Format("2006-01-02 15:04:05"), duration.Round(time.Second)))

			time.Sleep(duration)

			slog.Debug("Running scheduled tasks...")
			StartDailyNotifications()
			CheckFunnelActive()
		}
	}()
}
