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

var logData []piazza.LogMessage

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi. I'm pz-logger.")
}

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

	logData = append(logData, mssg)
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

	data, err := json.Marshal(logData)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func runLoggerServer(serviceAddress string, discoverAddress string, debug bool) error {

	//myAddress := fmt.Sprintf(":%s", port)
	//myURL := fmt.Sprintf("http://%s/log", myAddress)

	//piazza.RegistryInit(discoveryURL)
	//err := piazza.RegisterService("pz-logger", "core-service", myURL)
	//if err != nil {
//		return err
//	}

	r := mux.NewRouter()
	r.HandleFunc("/log/admin", handleAdminGet).
		Methods("GET")
	r.HandleFunc("/log", handleLoggerPost).
		Methods("POST")
	r.HandleFunc("/log", handleLoggerGet).
		Methods("GET")
	r.HandleFunc("/", handleHealthCheck).
		Methods("GET")

	server := &http.Server{Addr: serviceAddress, Handler: piazza.ServerLogHandler(r)}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// not reached
	return nil
}

func app() int {

	var err error

	// handles the command line flags, finds the discover service, registers us,
	// and figures out our own server address
	svc, err := piazza.NewDiscoverService(os.Args[0], "localhost:12341", "localhost:3000")
	if err != nil {
		log.Print(err)
		return 1
	}

	err = runLoggerServer(svc.BindTo, svc.DiscoverAddress, *svc.DebugFlag)
	if err != nil {
		log.Print(err)
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
