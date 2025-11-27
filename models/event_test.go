package models

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Helper to create dummy image file
func createDummyImage(filename string) error {
	path := filepath.Join(".data/uploads", filename)
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte("dummy image"))
	return err
}

// Helper to check if file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filepath.Join(".data/uploads", filename))
	return err == nil
}

func TestDeleteEventImages_RemovesFiles(t *testing.T) {
	// Create dummy files
	img1 := "testimg1.jpg"
	img2 := "testimg2.jpg"
	createDummyImage(img1)
	createDummyImage(img2)

	// Remove images from directory if test does not
	defer os.Remove(filepath.Join(".data/uploads", img1))
	defer os.Remove(filepath.Join(".data/uploads", img2))

	// Patch GetSubmissionsByEventSlug
	GetSubmissionsByEventSlug = func(slug string) ([]Submission, error) {
		return []Submission{
			{Name: "A", Filename: img1},
			{Name: "B", Filename: img2},
		}, nil
	}

	e := &Event{Slug: "test-event", Name: "Test Event"}
	err := e.deleteEventImages()
	if err != nil {
		t.Fatalf("deleteEventImages returned error: %v", err)
	}

	if fileExists(img1) || fileExists(img2) {
		t.Errorf("Expected image files to be deleted")
	}
}

func TestDeleteEventImages_NoSubmissions(t *testing.T) {
	GetSubmissionsByEventSlug = func(slug string) ([]Submission, error) {
		return []Submission{}, nil
	}
	e := &Event{Slug: "empty-event", Name: "Empty Event"}
	err := e.deleteEventImages()
	if err != nil {
		t.Fatalf("deleteEventImages returned error: %v", err)
	}
}

func TestDeleteEventImages_SubmissionNoImage(t *testing.T) {
	GetSubmissionsByEventSlug = func(slug string) ([]Submission, error) {
		return []Submission{
			{Name: "A", Filename: ""},
		}, nil
	}
	e := &Event{Slug: "no-img", Name: "No Image"}
	err := e.deleteEventImages()
	if err != nil {
		t.Fatalf("deleteEventImages returned error: %v", err)
	}

}

func TestDeleteEventImages_SubmissionImageNotFound(t *testing.T) {
	GetSubmissionsByEventSlug = func(slug string) ([]Submission, error) {
		return []Submission{
			{Name: "A", Filename: "no_img.jpg"},
		}, nil
	}
	e := &Event{Slug: "not-found", Name: "Not Found"}
	err := e.deleteEventImages()
	if err != nil {
		t.Fatalf("deleteEventImages returned error: %v", err)
	}
}

func TestDeleteEventImages_ErrorFromGetSubmissions(t *testing.T) {
	GetSubmissionsByEventSlug = func(slug string) ([]Submission, error) {
		return []Submission{}, fmt.Errorf("Error querying submissions")
	}
	e := &Event{Slug: "not-found", Name: "Not Found"}
	err := e.deleteEventImages()
	if err == nil {
		t.Fatalf("deleteEventImages did not return an error: %v", err)
	}
}
