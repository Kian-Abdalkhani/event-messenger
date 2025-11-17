package handlers

import (
	"log"
	"net/http"
	"time"

	"congrats-project.com/models"
	"congrats-project.com/utils"
)

func CreateEventForm(w http.ResponseWriter, r *http.Request) {

	events, err := models.GetAllActiveEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error retreiving events %v", err)
		return
	}

	data := struct {
		Events []models.Event
	}{
		Events: events,
	}

	renderTemplate(w, "templates/create_event_form.html", data)

}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// grab values from the form
	name := r.FormValue("name")
	slug := utils.GenerateSlug(name) // Auto-generate from name
	description := r.FormValue("description")
	coordinator := r.FormValue("coordinator")
	coordinatorContact := r.FormValue("coordinator_contact")
	recipientName := r.FormValue("recipientName")
	recipientContact := r.FormValue("recipientContact")
	websiteLink := utils.GetEventURL(slug, r)

	// Validate required fields
	if name == "" || recipientName == "" || recipientContact == "" {
		http.Error(w, "Name, recipient name, and recipient contact are required", http.StatusBadRequest)
		return
	}

	eventDate, err := time.Parse("2006-01-02", r.FormValue("event_date"))
	if err != nil {
		http.Error(w, "Invalid event date format", http.StatusBadRequest)
		return
	}

	// Fix: Event date should be in the future
	if eventDate.Before(time.Now().Truncate(24 * time.Hour)) {
		http.Error(w, "Event date must be in the future", http.StatusBadRequest)
		return
	}

	event := models.NewEvent(
		name,
		slug,
		eventDate,
		models.WithDescription(description),
		models.WithCoordinator(coordinator, coordinatorContact),
		models.WithRecipient(recipientName, recipientContact),
		models.WithWebsiteLink(websiteLink),
	)

	err = event.SaveEvent()
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/events", http.StatusSeeOther)
}
