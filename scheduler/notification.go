package scheduler

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"event-messenger.com/handlers"
	"event-messenger.com/models"
	"event-messenger.com/utils"
)

// Set a cap so that email doesn't reach SMTP limit (25MB limit)
const maxSubmissionsPerEmail = 150

type SubmissionEmailData struct {
	MessageText  string
	From         string
	ImageDataURI string
}

func sendEventNotification(event *models.Event) error {
	submissions, err := models.GetSubmissionsByEventSlug(event.Slug)
	if err != nil {
		log.Printf("Error retreiving submissions")
		return err
	}

	if len(submissions) == 0 {
		log.Printf("No submissions were made for this event")
		return nil
	}

	// If more submissions than email limit, cap emails in message at that limit
	if len(submissions) > maxSubmissionsPerEmail {
		log.Printf("Event %s has %d submissions, capping at %d for email size", event.Name, len(submissions), maxSubmissionsPerEmail)
		submissions = submissions[:maxSubmissionsPerEmail]
	}

	// Build submission data with base64-encoded images
	submissionData := make([]SubmissionEmailData, 0, len(submissions))
	for _, sub := range submissions {

		imgUri, err := encodeImageAsDataURI(sub.Filename)
		// If error encoding image, input blank string for image URI
		if err != nil {
			log.Printf("could not encode image for submission from %v", sub.Name)
			imgUri = ""
		}
		data := SubmissionEmailData{
			MessageText:  sub.Message,
			From:         sub.Name,
			ImageDataURI: imgUri,
		}

		submissionData = append(submissionData, data)
	}

	// Prepare template data
	templateData := struct {
		EventName       string
		RecipientName   string
		EventDate       time.Time
		Submissions     []SubmissionEmailData
		TotalCount      int
		CoordinatorName string
	}{
		EventName:       event.Name,
		RecipientName:   event.RecipientName,
		EventDate:       event.EventDate,
		Submissions:     submissionData,
		TotalCount:      len(submissions),
		CoordinatorName: event.Coordinator,
	}

	// Render email HTML
	htmlContent, err := handlers.RenderEmailTemplate(templateData)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Log email size for debugging
	emailSizeKB := len(htmlContent) / 1024
	log.Printf("Email size for event %s: %d KB (%d submissions)", event.Name, emailSizeKB, len(submissions))

	if emailSizeKB > 15000 { // Warn if over 15MB
		log.Printf("WARNING: Email size may be too large for SMTP")
	}

	// Send email
	subject := fmt.Sprintf("Your %s Messages", event.Name)
	err = utils.SendEmailNotification(event.RecipientEmail, subject, htmlContent)
	if err != nil {
		log.Printf("Failed to send email for event %s: %v", event.Name, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Mark email as sent
	err = event.MarkEmailSent()
	if err != nil {
		log.Printf("WARNING: Email sent for event %s but failed to mark as sent in DB: %v",
			event.Name, err)
		return fmt.Errorf("failed to mark email as sent: %w", err)
	}

	log.Printf("Successfully sent notification for event: %s to %s", event.Name, event.RecipientEmail)
	return nil
}

func encodeImageAsDataURI(filename string) (string, error) {
	imagePath := filepath.Join("./data/uploads", filename)
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("could not read image file: %w", err)
	}

	// Detect MIME type from file extension
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := "image/jpeg" // Default
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	}

	// Encode to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image), nil
}
