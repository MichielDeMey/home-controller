package main

import (
    "encoding/json"
    "fmt"
    "net/http"

    "home-controller/lib"
    "home-controller/models"

    "github.com/gorilla/mux"
)

const serial string = "123abc"
const iface_name string = "eth0"

func main() {
    go monitor.MonitorIP()

    // Setup the server routing
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/", Index)
    router.HandleFunc("/status", Status)

    // Listen for incoming HTTP connections
    fmt.Println("Listening on port 9000")
    http.ListenAndServe(":9000", router)
}

// GET /
func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to the home controller!")
}

// GET /status
func Status(w http.ResponseWriter, r *http.Request) {
    statusMessage := models.Status{
        Serial:  serial,
        Healthy: true,
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)

    json.NewEncoder(w).Encode(statusMessage)
}
