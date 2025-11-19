package handlers

import (
	"bytes"
	"log"
	"net/http"
	"text/template"
)

func renderTemplate(w http.ResponseWriter, templatePath string, data interface{}) {
	// Read and parse the HTML template
	tmpl, err := template.ParseFiles(templatePath)
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
	tmpl, err := template.ParseFiles("./templates/email_notification.html")
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
