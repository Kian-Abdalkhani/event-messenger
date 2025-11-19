package handlers

import "event-messenger.com/models"

type HomePageData struct {
	Events []models.EventPreview
}

type EventFormData struct {
	EventName   string
	SubjectName string
}

type SubmissionSuccessData struct {
	Name     string
	Message  string
	Filename string
}
