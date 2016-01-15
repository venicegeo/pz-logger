package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var startTime = time.Now()

var logData []string

func handleLoggerPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	var mssg piazza.LogMessage
	err = json.Unmarshal(data, &mssg)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	err = mssg.Validate()
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	log.Printf("LOG: %s\n", mssg.ToString())

	logData = append(logData, mssg.ToString())
}

func handleAdminGet(w http.ResponseWriter, r *http.Request) {
	m := piazza.AdminResponse{StartTime: startTime, Logger: &piazza.AdminResponseLogger{NumMessages: len(logData)}}

	data, err := json.Marshal(m)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func handleLoggerGet(w http.ResponseWriter, r *http.Request) {

	s := ""
	for _, m := range logData {
		s += m + "\n"
	}

	data := []byte(s)

	w.Write(data)
}

func runLoggerServer(discoveryURL string, port string) error {

	myAddress := fmt.Sprintf("%s:%s", "localhost", port)
	myURL := fmt.Sprintf("http://%s/log", myAddress)

	piazza.RegistryInit(discoveryURL)
	err := piazza.RegisterService("pz-logger", "core-service", myURL)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/log/admin", handleAdminGet).
		Methods("GET")
	r.HandleFunc("/log", handleLoggerPost).
		Methods("POST")
	r.HandleFunc("/log", handleLoggerGet).
		Methods("GET")

	server := &http.Server{Addr: myAddress, Handler: piazza.ServerLogHandler(r)}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// not reached
	return nil
}

func app() int {
	var discovery = flag.String("discovery", "http://localhost:3000", "URL of pz-discovery")
	var port = flag.String("port", "12341", "port number for pz-logger")

	flag.Parse()

	log.Printf("starting logger: discovery=%s, port=%s", *discovery, *port)

	err := runLoggerServer(*discovery, *port)
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
