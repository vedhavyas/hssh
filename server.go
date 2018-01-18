package hssh

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func writeResponse(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Write(body)
}

// streamData reads 1024 bytes of data and flushes it to response writer
// breaks on EOF or read error
func streamData(w http.ResponseWriter, pr *io.PipeReader) {
	buf := make([]byte, 1024)
	for {
		n, err := pr.Read(buf)
		if err != nil {
			pr.Close()
			break
		}

		d := buf[0:n]
		w.Write(d)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		buf = buf[0:0]
	}
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

	pr, pw := io.Pipe()
	defer pw.Close()

	client, err := NewSSHClient(req.Host, pw)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	go streamData(w, pr)
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
