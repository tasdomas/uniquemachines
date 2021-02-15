package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/errgo.v1"
)

const defaultServer = "http://localhost:8080"

const pingInterval = time.Second * 5

func main() {
	var st *status
	var err error

	serverURL := os.Getenv("SERVER")
	if serverURL == "" {
		serverURL = defaultServer
	}

	st, err = loadStatus()
	if os.IsNotExist(errgo.Cause(err)) {
		machineId, err := getMachineId()
		if os.IsNotExist(errgo.Cause(err)) || machineId == "" {
			// /etc/machine-id is empty on docker containers
			machineId = uuid.New().String()
		} else if err != nil {
			log.Fatal(err)
		}
		st = &status{
			MachineId: machineId,
			Token:     uuid.New().String(),
		}
		err = st.store()
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}
	run(st, serverURL, pingInterval)
}

func getMachineId() (string, error) {
	f, err := os.Open("/etc/machine-id")
	if err != nil {
		return "", errgo.Mask(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return strings.TrimSpace(string(b)), nil
}

const statusFile = "/tmp/token-status"

type status struct {
	MachineId string `json:"machine-id"`
	Token     string `json:"token"`
}

func loadStatus() (*status, error) {
	f, err := os.Open(statusFile)
	if err != nil {
		return nil, errgo.Mask(err, os.IsNotExist)
	}
	var s status
	dec := json.NewDecoder(f)
	err = dec.Decode(&s)
	if err != nil {
		return nil, errgo.Mask(err)
	}
	return &s, nil
}

func (st *status) store() error {
	f, err := os.OpenFile(statusFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return errgo.Mask(err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(st)
	return errgo.Mask(err)
}

type UpdateRequest struct {
	MachineId string `json:"id"`
	OldToken  string `json:"old-token"`
	NewToken  string `json:"new-token"`
}

type UpdateResponse struct {
	MachineId string `json:"id"`
}

func (st *status) ping(upstream string, newToken string) error {
	url := upstream + "/update"
	req := UpdateRequest{
		MachineId: st.MachineId,
		OldToken:  st.Token,
		NewToken:  newToken,
	}
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return errgo.Mask(err)
	}
	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return errgo.Mask(err)
	}
	defer resp.Body.Close()
	var update UpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&update)
	if update.MachineId != st.MachineId {
		log.Printf("new machine id issued: %s", update.MachineId)
	}
	st.MachineId = update.MachineId
	st.Token = newToken
	return errgo.Mask(st.store())
}

// run periodically pings upstream server with new token.
func run(st *status, server string, d time.Duration) error {
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			newToken := uuid.New().String()
			err := st.ping(server, newToken)
			if err != nil {
				log.Printf("error pinging upstream %v", err)
				continue
			}
			log.Printf("update succesful")
		}
	}

}
