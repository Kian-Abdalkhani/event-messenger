package handlers

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

var baseDir string

func init() {
	// Use current working directory instead of executable path
	// This works for both 'go run' and compiled binaries
	var err error
	baseDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
}

func renderTemplate(w http.ResponseWriter, templatePath string, data interface{}) {
	// Make path absolute relative to executable
	fullPath := filepath.Join(baseDir, templatePath)

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
	fullPath := filepath.Join(baseDir, "templates/email_notification.html")

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
