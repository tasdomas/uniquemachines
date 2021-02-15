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
	mux.HandleFunc("/status", s.serveCountRequest)
	mux.HandleFunc("/update", s.serveUpdateRequest)

	log.Fatal(http.ListenAndServe(":8080", mux))
}

type srv struct {
	d *directory.Directory
}

type UpdateRequest struct {
	MachineId string `json:"id"`
	OldToken  string `json:"old-token"`
	NewToken  string `json:"new-token"`
}

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

type countResponse struct {
	Count int `json:"unique-machines"`
}

func (s *srv) serveCountRequest(w http.ResponseWriter, req *http.Request) {
	cnt := s.d.Count()
	err := json.NewEncoder(w).Encode(countResponse{Count: cnt})
	if err != nil {
		log.Printf("failed to encode response: %w", err)
		return
	}
}
