package hssh

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func writeResponse(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Write(body)
}

// sshHandler handles /ssh requests
func sshHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Command string `json:"command_string"`
		Host    string `json:"server_ip"`
	}

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}

	defer r.Body.Close()
	err = json.Unmarshal(d, &req)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}

	client, err := NewSSHClient(req.Host, w)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	err = client.ExecuteCommand(req.Command)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
	}
}

// StartServer will start the
func StartServer(addr string) {
	r := mux.NewRouter()
	r.HandleFunc("/ssh", sshHandler).Methods("POST")
	http.ListenAndServe(addr, r)
}
