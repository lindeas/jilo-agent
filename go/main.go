/*
Jilo Agent

Description: Remote agent for Jilo Web with http API
Author: Yasen Pramatarov
License: GPLv2
Project URL: https://lindeas.com/jilo
Year: 2024
Version: 0.1
*/

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "strconv"
    "strings"
)

// NginxData holds the nginx data structure for the API response to /nginx
type NginxData struct {
    NginxState		string	`json:"nginx_status"`
    NginxConnections	int	`json:"nginx_connections"`
}

// getNginxState checks the status of the nginx service
func getNginxState() string {
    output, err := exec.Command("systemctl", "is-active", "nginx").Output()
    if err != nil {
        log.Printf("Error checking the nginx state: %v", err)
        return "error"
    }
    state := strings.TrimSpace(string(output))
    if state == "active" {
        return "running"
    }
    return "not running"
}

// getNginxConnections gets the number of active connections to the specified web port
func getNginxConnections() int {
    output, err := exec.Command("bash", "-c", "netstat -an | grep ':80' | wc -l").Output()
    if err != nil {
        log.Printf("Error counting the Nginx connections: %v", err)
        return -1
    }
    connections := strings.TrimSpace(string(output))
    connectionsInt, err := strconv.Atoi(connections)
    if err != nil {
        log.Printf("Error converting connections to integer number: %v", err)
        return -1
    }
    return connectionsInt
}

// nginxHandler handles the /nginx endpoint
func nginxHandler(w http.ResponseWriter, r *http.Request) {
    data := NginxData {
        NginxState:		getNginxState(),
        NginxConnections:	getNginxConnections(),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

// main sets up the http server and the routes
func main() {
    http.HandleFunc("/nginx", nginxHandler)
    fmt.Println("Starting agent server on port 8080.")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Could not start the server: %v\n", err)
    }
}
