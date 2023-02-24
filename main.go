package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Sachingeek125/GolangAuth/routers"
	mux "github.com/gorilla/mux"
)

const port = 8080

func cleanup() {
	log.Printf("Server is being shut downed!......")
}

func main() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGALRM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	// define mux router
	r := mux.NewRouter()

	r.HandleFunc("/register", routers.Register).Methods("POST")
	r.HandleFunc("/login", routers.Login).Methods("POST")
	r.HandleFunc("/logout", routers.Logout).Methods("GET")

	http.Handle("/", r)

	server := newServer(":"+strconv.Itoa(port), r)
	log.Printf("Starting server on %d", port)
	defer cleanup()
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func newServer(s string, r *mux.Router) *http.Server {
	return &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      r,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	}

}
