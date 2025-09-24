package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
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

func insertProgramHandler(w http.ResponseWriter, r *http.Request) {
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

type ProgramMetadata struct {
	ProgramID    uint64   `json:"program_id"`
	Name         string   `json:"name"`
	PreviewImage string   `json:"preview_image"` // base64
	Tags         []string `json:"tags"`
	CreatedAt    string   `json:"created_at"`
}

var cacheLock sync.RWMutex

const cacheFile = "programs_cache.bin"

func updateCache(limit int) error {
	resp, err := http.Get(fmt.Sprintf("http://localhost:5000/programs?limit=%d", limit))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Flask error: %s", string(body))
	}

	// Decode JSON from Flask
	var programs []ProgramMetadata
	if err := json.NewDecoder(resp.Body).Decode(&programs); err != nil {
		return err
	}

	// Open file for binary writing
	cacheLock.Lock()
	defer cacheLock.Unlock()
	f, err := os.Create(cacheFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write number of programs
	if err := binary.Write(f, binary.LittleEndian, uint64(len(programs))); err != nil {
		return err
	}

	for _, p := range programs {
		// program_id
		if err := binary.Write(f, binary.LittleEndian, p.ProgramID); err != nil {
			return err
		}

		// name
		if err := writeString(f, p.Name); err != nil {
			return err
		}

		// created_at
		if err := writeString(f, p.CreatedAt); err != nil {
			return err
		}

		// preview_image (decode from base64 into raw bytes)
		imgBytes, err := base64.StdEncoding.DecodeString(p.PreviewImage)
		if err != nil {
			return err
		}
		if err := writeBytes(f, imgBytes); err != nil {
			return err
		}

		// tags
		if err := binary.Write(f, binary.LittleEndian, uint64(len(p.Tags))); err != nil {
			return err
		}
		for _, tag := range p.Tags {
			if err := writeString(f, tag); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, uint64(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func writeBytes(w io.Writer, b []byte) error {
	if err := binary.Write(w, binary.LittleEndian, uint64(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func getProgramsHandler(w http.ResponseWriter, r *http.Request) {
	if err := updateCache(50); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Cache updated"))
}

type GenerateProgramRequest struct {
	StateSize   int      `json:"state_size"`
	MaxLength   int      `json:"max_length"`
	StartTokens []string `json:"start_tokens",omitempty`
}

func generateProgramHandler(w http.ResponseWriter, r *http.Request) {
	// handle c++

	// send actual server
	reqBody := GenerateProgramRequest{
		StateSize: 8,
		MaxLength: 500,
	}

	jsonBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:5000/generate_global", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	// decode JSON response
	var data struct {
		Sequence []string `json:"sequence"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write sequence to file
	f, err := os.Create("generated_random_sequence.seq")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	for _, line := range data.Sequence {
		_, _ = f.WriteString(line + "\n")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sequence written to generated_random_sequence.seq\n"))
}

type GenerateMixedProgramRequest struct {
	StateSizes  int      `json:"state_sizes"`
	MaxLength   int      `json:"max_length"`
	StartTokens []string `json:"start_tokens",omitempty`
}

func generateMixedProgramHandler(w http.ResponseWriter, r *http.Request) {
	// handle c++

	// send actual server
	reqBody := GenerateMixedProgramRequest{
		// StateSize: 8,
		MaxLength: 500,
	}

	jsonBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:5000/generate_mixed_global", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	// decode JSON response
	var data struct {
		Sequence []string `json:"sequence"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write sequence to file
	f, err := os.Create("generated_mixed_random_sequence.seq")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	for _, line := range data.Sequence {
		_, _ = f.WriteString(line + "\n")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sequence written to generated_mixed_random_sequence.seq\n"))
}

func main() {
	http.HandleFunc("/program", insertProgramHandler) // C++ posts here
	http.HandleFunc("/shutdown", shutdownHandler)     // Shutdown endpoint
	http.HandleFunc("/programs", getProgramsHandler)
	http.HandleFunc("/generate_random", generateProgramHandler)
	http.HandleFunc("/generate_mixed_random", generateMixedProgramHandler)

	fmt.Println("Bridge server listening on :9999")
	if err := http.ListenAndServe(":9999", nil); err != nil {
		fmt.Println("Server stopped:", err)
	}
}
