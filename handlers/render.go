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
	// Get the executable's directory
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir = filepath.Dir(ex)
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
