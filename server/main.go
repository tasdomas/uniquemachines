package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/tasdomas/uniquemachines/server/directory"
)

func main() {
	s := &srv{
		d: directory.New(),
	}

	mux := http.NewServeMux()
	// status returns the number of unique machines being tracked by the server.
	mux.HandleFunc("/status", s.serveCountRequest)
	// update handles requests coming in from clients.
	mux.HandleFunc("/update", s.serveUpdateRequest)

	log.Fatal(http.ListenAndServe(":8080", mux))
}

// srv handles incoming API requests.
type srv struct {
	d *directory.Directory
}

// UpdateRequest contains data sent by the client.
type UpdateRequest struct {
	// MachineId is the machine id of the client.
	MachineId string `json:"id"`
	// OldToken and NewToken represent the token
	// swap proposed by the client.
	OldToken string `json:"old-token"`
	NewToken string `json:"new-token"`
}

// UpdateResponse is the response of the server.
// In cases where a cloned machine is detected, the server will
// issue a new machineId to the client.
type UpdateResponse struct {
	MachineId string `json:"id"`
}

func (s *srv) serveUpdateRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "expecting POST", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	var r UpdateRequest
	err := dec.Decode(&r)
	if err != nil {
		log.Print("failed to decode request: %v", err)
		return
	}
	log.Printf("new machine update request for id %v", r.MachineId)
	id := s.d.UpdateMachine(r.MachineId, r.OldToken, r.NewToken)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(UpdateResponse{MachineId: id})
	if err != nil {
		log.Printf("failed to encode response: %w", err)
		return
	}
}

type CountResponse struct {
	Count int `json:"unique-machines"`
}

func (s *srv) serveCountRequest(w http.ResponseWriter, req *http.Request) {
	cnt := s.d.Count()
	err := json.NewEncoder(w).Encode(CountResponse{Count: cnt})
	if err != nil {
		log.Printf("failed to encode response: %w", err)
		return
	}
}
