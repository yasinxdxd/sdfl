package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Request to actual server
type ProgramRequest struct {
	Code                   string   `json:"code"`
	Description            string   `json:"description"`
	SequenceRepresentation []string `json:"sequence_representation"`
	Name                   string   `json:"name"`
	PreviewImage           string   `json:"preview_image"` // base64 string
	Tags                   []string `json:"tags"`
}

func cppHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	code := r.FormValue("code")
	sequence := r.FormValue("sequence_representation")
	sequenceRepresentation := strings.Split(sequence, "\n")

	// Collect all tags (multiple fields named "tags")
	tags := r.Form["tags"]

	// Read preview_image (required now)
	file, _, err := r.FormFile("preview_image")
	if err != nil {
		http.Error(w, "preview_image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	previewBytes, _ := io.ReadAll(file)

	// Base64 encode preview image before sending to Flask
	previewB64 := base64.StdEncoding.EncodeToString(previewBytes)

	// Validate required fields
	if code == "" || len(sequenceRepresentation) == 0 || name == "" || description == "" || len(tags) < 3 {
		http.Error(w, "missing required fields or tags < 3", http.StatusBadRequest)
		return
	}

	// Build JSON for actual server
	reqBody := ProgramRequest{
		Code:                   code,
		SequenceRepresentation: sequenceRepresentation,
		Description:            description,
		Name:                   name,
		PreviewImage:           previewB64, // base64 string
		Tags:                   tags,       // []string
	}

	jsonBytes, _ := json.Marshal(reqBody)

	// Forward to actual Flask server
	resp, err := http.Post("http://localhost:5000/program", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		http.Error(w, "Remote Server unreachable!", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bridge server shutting down"))
	go func() {
		fmt.Println("Shutdown requested. Exiting...")
		os.Exit(0)
	}()
}

func main() {
	http.HandleFunc("/program", cppHandler)       // C++ posts here
	http.HandleFunc("/shutdown", shutdownHandler) // Shutdown endpoint

	fmt.Println("Bridge server listening on :9999")
	if err := http.ListenAndServe(":9999", nil); err != nil {
		fmt.Println("Server stopped:", err)
	}
}
