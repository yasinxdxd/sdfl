package main

import (
	"bytes"
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
	sequence := r.FormValue("sequence_representation")
	sequence_representation := strings.Split(sequence, "\n")
	fmt.Printf("%v", sequence_representation)

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

	// build JSON for actual server
	reqBody := ProgramRequest{
		Code:                   code,
		SequenceRepresentation: sequence_representation,
		Description:            description,
		Name:                   name,
		PreviewImage:           previewImage,
	}

	jsonBytes, _ := json.Marshal(reqBody)

	// forward to actual server
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
