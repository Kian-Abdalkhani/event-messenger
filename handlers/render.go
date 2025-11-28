package handlers

import (
	"bytes"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"text/template"

	"event-messenger.com/utils"
)

func renderTemplate(w http.ResponseWriter, templatePath string, data any) {
	// Check if it is an absolute path
	var fullPath string
	if filepath.IsAbs(templatePath) {
		fullPath = templatePath
	} else {
		slog.Error("Ensure template path is absolute", "templatePath", templatePath)
		// fullPath = utils.ProjectPath(templatePath)
	}

	// Read and parse the HTML template
	tmpl, err := template.ParseFiles(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Render error: %v", err)
	}
}

func RenderEmailTemplate(data any) (string, error) {
	// Make path absolute relative to executable
	fullPath := utils.ProjectPath("templates", "email_notification.html")

	tmpl, err := template.ParseFiles(fullPath)
	if err != nil {
		log.Printf("Email template parse error: %v", err)
		return "", err
	}

	// Use a bytes.Buffer instead of http.ResponseWriter to capture the output
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Printf("Email template render error: %v", err)
		return "", err
	}

	// Return rendered HTML as a string
	return buf.String(), nil

}
