package gelf_server

import (
	jsonPkg "encoding/json"
	journalPkg "github.com/cbuschka/gold/internal/journal"
	"github.com/gorilla/mux"
	"github.com/kataras/golog"
	gelf "gopkg.in/Graylog2/go-gelf.v2/gelf"
	"net"
	"net/http"
)

func ServeHttp(addr string, journal journalPkg.Journal) error {

	httpListener, err := net.Listen("tcp", addr)
	golog.Infof("GELF http listener listening on %s...", addr)
	if err != nil {
		return err
	}
	defer httpListener.Close()
	httpHandler := newHttpHandler(journal)
	err = http.Serve(httpListener, httpHandler)
	return err
}

func newHttpHandler(journal journalPkg.Journal) http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/gelf", func(w http.ResponseWriter, r *http.Request) {

		var gelfMessage gelf.Message
		err := jsonPkg.NewDecoder(r.Body).Decode(&gelfMessage)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		senderHost := r.Header.Get("X-Forwarded-For")
		if senderHost == "" {
			senderHost = r.RemoteAddr
		}
		message := journalPkg.FromGelfMessage(&gelfMessage, senderHost, "http")

		err = journal.WriteMessage(message)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Error(w, "", http.StatusCreated)
	}).Methods("POST")

	return http.Handler(router)
}
