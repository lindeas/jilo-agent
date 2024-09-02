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
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strconv"
    "strings"
)

// Config holds the structure of the configuration file
type Config struct {
    AgentPort		int	`json:"agent_port"`
    NginxPort		int	`json:"nginx_port"`
    JicofoStatsURL	string	`json:"jicofo_stats_url"`
}

// NginxData holds the nginx data structure for the API response to /nginx
type NginxData struct {
    NginxState		string	`json:"nginx_state"`
    NginxConnections	int	`json:"nginx_connections"`
}

// JicofoData holds the Jicofo data structure for the API response to /jicofo
type JicofoData struct {
    JicofoState		string			`json:"jicofo_state"`
    JicofoAPIData	map[string]interface{}	`json:"jicofo_api_data"`
}

// getServiceState checks the status of the speciied service
func getServiceState(service string) string {
    output, err := exec.Command("systemctl", "is-active", service).Output()
    if err != nil {
        log.Printf("Error checking the service \"%v\" state: %v", service, err)
        return "error"
    }
    state := strings.TrimSpace(string(output))
    if state == "active" {
        return "running"
    }
    return "not running"
}

// getNginxConnections gets the number of active connections to the specified web port
func getNginxConnections(port int) int {
    cmd := fmt.Sprintf("netstat -an | grep ':%d' | wc -l", port)
    output, err := exec.Command("bash", "-c", cmd).Output()
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

// getJitsiAPIData gets the response from the specified Jitsi stats API
func getJitsiAPIData(service string, url string) map[string]interface{} {
    cmd := fmt.Sprintf("curl -s %v", url)
    output, err := exec.Command("bash", "-c", cmd).Output()
    if err != nil {
        log.Printf("Error getting the \"%v\" API stats: %v", service, err)
        return map[string]interface{}{"error": "failed to get the Jitsi API stats"}
    }
    var result map[string]interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        log.Printf("Error in parsing the JSON: %v", err)
        return map[string]interface{}{"error": "invalid JSON format"}
    }
    return result
}

// loadConfig loads the configuration from a JSON config file
func loadConfig(filename string) (Config, error) {
    var config Config

    file, err := os.Open(filename)
    if err != nil {
        return config, err
    }
    defer file.Close()

    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        return config, err
    }

    if err := json.Unmarshal(bytes, &config); err != nil {
        return config, err
    }

    return config, nil
}

// nginxHandler handles the /nginx endpoint
func nginxHandler(config Config, w http.ResponseWriter, r *http.Request) {
    data := NginxData {
        NginxState:		getServiceState("nginx"),
        NginxConnections:	getNginxConnections(config.NginxPort),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

// jicofoHandler handles the /jicofo endpoint
func jicofoHandler(config Config, w http.ResponseWriter, r *http.Request) {
    data := JicofoData {
        JicofoState:		getServiceState("jicofo"),
        JicofoAPIData:		getJitsiAPIData("jicofo", config.JicofoStatsURL),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}


// main sets up the http server and the routes
func main() {

    // load the configuration
    config, err := loadConfig("jilo-agent.json")
    if err != nil {
        log.Fatalf("Error loading the config file: %v\n", err)
    }

    // endpoints
    http.HandleFunc("/nginx", func(w http.ResponseWriter, r *http.Request) {
        nginxHandler(config, w, r)
    })
    http.HandleFunc("/jicofo", func(w http.ResponseWriter, r *http.Request) {
        jicofoHandler(config, w, r)
    })

    // start the http server
    agentPortStr := fmt.Sprintf(":%d", config.AgentPort)
    fmt.Printf("Starting Jilo agent server on port %d.\n", config.AgentPort)
    if err := http.ListenAndServe(agentPortStr, nil); err != nil {
        log.Fatalf("Could not start the server: %v\n", err)
    }
}
