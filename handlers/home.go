package handlers

import (
	"net/http"

	"congrats-project.com/models"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Retrieve data on all events
	events, err := models.GetActiveEventPreviews()
	if err != nil {
		http.Error(w, "Error loading events", http.StatusInternalServerError)
		return
	}

	data := struct {
		Events []models.EventPreview
	}{
		Events: events,
	}

	renderTemplate(w, "templates/home.html", data)

}
