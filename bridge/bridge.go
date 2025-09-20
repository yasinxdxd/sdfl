package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Data sent by other Go app
type GoAppData struct {
	Code                   string   `json:"code"`
	SequenceRepresentation []string `json:"sequence_representation"`
}

// Combined request to Flask
type ProgramRequest struct {
	Code                   string   `json:"code"`
	Description            string   `json:"description"`
	SequenceRepresentation []string `json:"sequence_representation"`
	Name                   string   `json:"name"`
	PreviewImage           []byte   `json:"preview_image,omitempty"`
}

var (
	// buffer: store latest Go app data keyed by some program key
	buffer = struct {
		sync.Mutex
		data GoAppData
	}{}
)

func goAppHandler(w http.ResponseWriter, r *http.Request) {
	var d GoAppData
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}

	buffer.Lock()
	buffer.data = d
	buffer.Unlock()
	w.WriteHeader(200)
}

func cppHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form", 400)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	// optional preview_image
	var previewImage []byte
	if file, _, err := r.FormFile("preview_image"); err == nil {
		defer file.Close()
		previewImage, _ = io.ReadAll(file)
	}

	// get latest Go app data
	buffer.Lock()
	goData := buffer.data
	buffer.Unlock()

	// Check if both parts exist
	if goData.Code == "" || len(goData.SequenceRepresentation) == 0 || name == "" || description == "" {
		// Not ready yet, just acknowledge receipt
		w.WriteHeader(200)
		fmt.Fprintln(w, "Data stored, waiting for both parts to be ready")
		return
	}

	// build JSON for Flask
	reqBody := ProgramRequest{
		Code:                   goData.Code,
		SequenceRepresentation: goData.SequenceRepresentation,
		Description:            description,
		Name:                   name,
		PreviewImage:           previewImage,
	}

	jsonBytes, _ := json.Marshal(reqBody)

	// pretty print to terminal
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonBytes, "", "  "); err == nil {
		fmt.Println("=== Data to be sent to Flask ===")
		fmt.Println(prettyJSON.String())
		fmt.Println("=== End of Data ===")
	}

	// forward to Flask
	resp, err := http.Post("http://localhost:5000/program", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		http.Error(w, "Flask unreachable", 502)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/goapp", goAppHandler) // other Go app posts here
	http.HandleFunc("/program", cppHandler) // C++ posts here to trigger send
	http.ListenAndServe(":9999", nil)
}
