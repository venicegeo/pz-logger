package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var startTime = time.Now()

var logData []string

// all fields required
type Message struct {
	Service  string `json:"service"`
	Address  string `json:"address"`
	Time     string `json:"time"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func (m *Message) ToString() string {
	s := fmt.Sprintf("[%s, %s, %s, %s, %s]",
		m.Service, m.Address, m.Time, m.Severity, m.Message)
	return s
}

type AdminLoggerMessage struct {
	NumMessages int `json:"num_messages"`
}

type AdminMessage struct {
	StartTime time.Time `json:"starttime"`
	Logger AdminLoggerMessage `json:"logger"`
}

func handleLoggerPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	var mssg Message
	err = json.Unmarshal(data, &mssg)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	if mssg.Service == "" || mssg.Address == "" || mssg.Time == "" ||
		mssg.Severity == "" || mssg.Message == "" {
		http.Error(w, fmt.Sprintf("required field missing"), http.StatusBadRequest)
		return
	}

	log.Print(mssg)

	logData = append(logData, mssg.ToString())
}

func handleAdminGet(w http.ResponseWriter, r *http.Request) {
	m := AdminMessage{StartTime: startTime, Logger: AdminLoggerMessage{NumMessages: len(logData)}}

	data, err := json.Marshal(m)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func handleLoggerGet(w http.ResponseWriter, r *http.Request) {

	s := ""
	for _, m := range(logData) {
		s += m + "\n"
	}

	data := []byte(s)

	w.Write(data)
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func runLoggerServer(host string, port string) error {

	r := mux.NewRouter()
	r.HandleFunc("/log/admin", handleAdminGet).
		Methods("GET")
	r.HandleFunc("/log", handleLoggerPost).
		Methods("POST")
	r.HandleFunc("/log", handleLoggerGet).
		Methods("GET")

	server := &http.Server{Addr: host + ":" + port, Handler: Log(r)}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// not reached
	return nil
}

func app() int {
	var host = flag.String("host", "localhost", "host name")
	var port = flag.String("port", "12341", "port number")

	flag.Parse()

	log.Printf("starting logger: host=%s, port=%s", *host, *port)

	err := runLoggerServer(*host, *port)
	if err != nil {
		fmt.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string) int {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)
	return app()
}

func main() {
	os.Exit(app())
}
