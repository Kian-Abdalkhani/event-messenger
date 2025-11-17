package scheduler

import (
	"log"
	"time"

	"congrats-project.com/models"
)

// StartCleanupScheduler runs weekly to clean up old events
func StartCleanupScheduler(graceDays int) {
	log.Printf("Cleanup scheduler started - will run weekly with %d day grace period", graceDays)

	go func() {
		for {
			// Run cleanup weekly at 2 AM
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())

			// If we've passed 2 AM today, schedule for next week
			if now.After(nextRun) {
				nextRun = nextRun.AddDate(0, 0, 7)
			}

			duration := time.Until(nextRun)
			log.Printf("Next cleanup scheduled for: %s (in %v)", nextRun.Format("2006-01-02 15:04:05"), duration)

			time.Sleep(duration)

			log.Println("Running scheduled cleanup...")
			cleanupOldEvents(graceDays)
		}
	}()
}

func cleanupOldEvents(graceDays int) {
	events, err := models.GetEventsReadyForDeletion(graceDays)
	if err != nil {
		log.Printf("Error retrieving events for cleanup: %v", err)
		return
	}

	if len(events) == 0 {
		log.Println("No events ready for cleanup")
		return
	}

	log.Printf("Found %d events ready for cleanup", len(events))

	for _, event := range events {
		log.Printf("Deleting event: %s (sent %d days ago)",
			event.Name,
			int(time.Since(event.EmailSentAt.Time).Hours()/24))

		err := event.DeleteEvent()
		if err != nil {
			log.Printf("Failed to delete event %s: %v", event.Name, err)
			continue
		}

		log.Printf("Successfully deleted event: %s", event.Name)
	}
}
