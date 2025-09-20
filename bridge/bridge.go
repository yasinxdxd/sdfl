package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Request to Flask
type ProgramRequest struct {
	Code                   string   `json:"code"`
	Description            string   `json:"description"`
	SequenceRepresentation []string `json:"sequence_representation"`
	Name                   string   `json:"name"`
	PreviewImage           []byte   `json:"preview_image,omitempty"`
}

func cppHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	code := r.FormValue("code")
	sequence := r.Form["sequence_representation"] // çoklu gönderim için

	// optional preview_image
	var previewImage []byte
	if file, _, err := r.FormFile("preview_image"); err == nil {
		defer file.Close()
		previewImage, _ = io.ReadAll(file)
	}

	// Check if required fields exist
	if code == "" || len(sequence) == 0 || name == "" || description == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Build JSON for Flask
	reqBody := ProgramRequest{
		Code:                   code,
		SequenceRepresentation: sequence,
		Description:            description,
		Name:                   name,
		PreviewImage:           previewImage,
	}

	jsonBytes, _ := json.Marshal(reqBody)

	// Debug log
	fmt.Println("=== Data to be sent to Flask ===")
	fmt.Println(string(jsonBytes))
	fmt.Println("=== End of Data ===")

	// Forward to Flask
	resp, err := http.Post("http://localhost:5000/program", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		http.Error(w, "Flask unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/program", cppHandler) // C++ posts here
	fmt.Println("Bridge server listening on :9999")
	http.ListenAndServe(":9999", nil)
}
