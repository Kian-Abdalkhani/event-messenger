package handlers

import (
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"event-messenger.com/models"
	"event-messenger.com/utils"
	"golang.org/x/image/draw"
)

const (
	MaxNameLength    = 100
	MaxMessageLength = 500
	MaxFileSize      = 10 << 20 // 10 MB
	MaxImageWidth    = 800
)

// Allowed MIME types for image uploads
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// saveSubmission saves a submission to the database
func saveSubmission(eventId int, name string, message string, filename string) error {
	submission := models.Submission{
		EventID:  eventId,
		Name:     name,
		Message:  message,
		Filename: filename,
	}

	// Save submission to the event database
	err := submission.Save()

	return err
}

// create formHandler that handles the generation of the input form
func SubmissionFormHandler(w http.ResponseWriter, r *http.Request, slug string) {

	event, err := models.GetEventBySlug(slug)
	if err != nil {
		http.Error(w, "Unable to retreive event data", http.StatusInternalServerError)
		slog.Error("Unable to retreive event data", "error", err)
		return
	}

	data := struct {
		EventName     string
		RecipientName string
		EventSlug     string
	}{
		EventName:     event.Name,
		RecipientName: event.RecipientName,
		EventSlug:     slug,
	}

	renderTemplate(w, utils.ProjectPath("templates", "submission_form.html"), data)

}

// create submitHandler for handling the submission of the message and contents (include success/error msg)
func SubmissionHandler(w http.ResponseWriter, r *http.Request, slug string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form (10 MB max)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get form values
	name := r.FormValue("name")
	message := r.FormValue("message")

	if len(name) > MaxNameLength {
		http.Error(w, fmt.Sprintf("Name exceeds maximum length of %d characters", MaxNameLength), http.StatusBadRequest)
		return
	}

	if len(message) > MaxMessageLength {
		http.Error(w, fmt.Sprintf("Message exceeds maximum length of %d characters", MaxMessageLength), http.StatusBadRequest)
		return
	}

	if len(name) == 0 || len(message) == 0 {
		http.Error(w, "Name and message are required", http.StatusBadRequest)
		return
	}

	// Handle file upload
	file, handler, err := r.FormFile("image")

	if err != nil {
		if err == http.ErrMissingFile {
			http.Error(w, "Image upload is required", http.StatusBadRequest)
			slog.Error("No image found in submission")
			return
		}
		http.Error(w, "Error processing image", http.StatusInternalServerError)
		slog.Error("Error processing image", "error", err)
		return
	}

	defer file.Close()

	//Validate file is an image by checking MIME type
	buffer := make([]byte, 512) // checks the first 512 bytes of a program
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Reset file pointer to beginning after reading
	_, err = file.Seek(0, 0)
	if err != nil {
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)

	if !allowedImageTypes[contentType] {
		http.Error(w, "Only image files (JPEG, PNG, GIF, WebP) are allowed", http.StatusBadRequest)
		slog.Error("Invalid file type", "content-type", contentType)
		return
	}

	// Decode and resize image
	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Error processing image", http.StatusInternalServerError)
		slog.Error("Image decode error", "error", err)
		return
	}
	slog.Info("Successfully decoded image format", "format", format)

	// Resize if width exceeds limit
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	slog.Info("Original image dimensions", "width", width, "height", height)

	if width > MaxImageWidth {
		// Calculate new dimensions maintaining aspect ratio
		newWidth := MaxImageWidth
		newHeight := (height * MaxImageWidth) / width

		// Create resized image
		resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)
		img = resized
		slog.Info("Resized image to ", "newWidth", newWidth, "newHeight", newHeight)
	}

	var filename string

	// Create uploads directory if it doesn't exist
	os.MkdirAll("./data/uploads", os.ModePerm)

	// Create unique filename (replace all img extensions with .jpg)
	baseFilename := filepath.Base(handler.Filename)
	ext := filepath.Ext(baseFilename)
	nameWithoutExt := baseFilename[:len(baseFilename)-len(ext)]
	filename = fmt.Sprintf("%d_%s.jpg", time.Now().Unix(), nameWithoutExt)
	filePath := utils.ProjectPath("data", "uploads", filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		slog.Error("File save error", "error", err)
		return
	}

	defer dst.Close()

	// Encode as JPEG with quality 85
	err = jpeg.Encode(dst, img, &jpeg.Options{Quality: 85})
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		slog.Error("JPEG encode error", "error", err)
		return
	}

	event, err := models.GetEventBySlug(slug)
	if err != nil {
		http.Error(w, "Unable to retreive event data", http.StatusInternalServerError)
		slog.Error("Unable to retreive event data", "error", err)
		return
	}

	// Save to database
	err = saveSubmission(event.ID, name, message, filename)
	if err != nil {
		http.Error(w, "Error saving submission to database", http.StatusInternalServerError)
		slog.Error("Database save error", "error", err)
		return
	}

	// Prepare data for template
	data := map[string]string{
		"Name":      name,
		"Message":   message,
		"Filename":  filename,
		"EventSlug": slug,
	}

	// Get saved file size for verification
	fileInfo, _ := dst.Stat()
	slog.Info("Successfully saved processed image: %s (size: %.2f KB)", filename, float64(fileInfo.Size())/1024)

	renderTemplate(w, utils.ProjectPath("templates", "success.html"), data)

	slog.Info("Received submission", "name", name)
}

func ViewSubmissionsByEvent(w http.ResponseWriter, r *http.Request, slug string) {
	submissions, err := models.GetSubmissionsByEventSlug(slug)
	if err != nil {
		http.Error(w, "Error retrieving event submissions", http.StatusInternalServerError)
		slog.Error("Error retreiving submissions", "error", err)
		return
	}

	renderTemplate(w, utils.ProjectPath("templates", "view_messages.html"), submissions)

}
