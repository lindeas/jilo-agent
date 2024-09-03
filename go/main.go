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
    "gopkg.in/yaml.v2"
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
    AgentPort		int	`yaml:"agent_port"`
    NginxPort		int	`yaml:"nginx_port"`
    ProsodyPort		int	`yaml:"prosody_port"`
    JicofoStatsURL	string	`yaml:"jicofo_stats_url"`
    JVBStatsURL		string	`yaml:"jvb_stats_url"`
    JibriHealthURL	string	`yaml:"jibri_health_url"`
}

// NginxData holds the nginx data structure for the API response to /nginx
type NginxData struct {
    NginxState		string	`json:"nginx_state"`
    NginxConnections	int	`json:"nginx_connections"`
}

// ProsodyData holds the prosody data structure for the API response to /prosody
type ProsodyData struct {
    ProsodyState	string	`json:"prosody_state"`
    ProsodyConnections	int	`json:"prosody_connections"`
}

// JicofoData holds the Jicofo data structure for the API response to /jicofo
type JicofoData struct {
    JicofoState		string			`json:"jicofo_state"`
    JicofoAPIData	map[string]interface{}	`json:"jicofo_api_data"`
}

// JVBData holds the JVB data structure for the API response to /jvb
type JVBData struct {
    JVBState		string			`json:"jvb_state"`
    JVBAPIData		map[string]interface{}	`json:"jvb_api_data"`
}

// JibriData holds the Jibri data structure for the API response to /jibri
type JibriData struct {
    JibriState		string			`json:"jibri_state"`
    JibriHealthData	map[string]interface{}	`json:"jibri_health_data"`
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

// getServiceConnections gets the number of active connections to the specified port
func getServiceConnections(service string, port int) int {
    cmd := fmt.Sprintf("netstat -an | grep ':%d' | wc -l", port)
    output, err := exec.Command("bash", "-c", cmd).Output()
    if err != nil {
        log.Printf("Error counting the \"%v\" connections: %v", service, err)
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

// loadConfig loads the configuration from a YAML config file
func loadConfig(filename string) (Config) {

    // default config values
    config := Config {
        AgentPort: 8081, // default Agent port (we avoid 80, 443, 8080 and 8888)
        NginxPort: 80, // default nginx port
        ProsodyPort: 5222, // default prosody port
        JicofoStatsURL: "http://localhost:8888/stats", // default Jicofo stats URL
        JVBStatsURL: "http://localhost:8080/colibri/stats", // default JVB stats URL
        JibriHealthURL: "http://localhost:2222/jibri/api/v1.0/health", // default Jibri health URL
    }

    // we try to load the config file; use default values otherwise
    file, err := os.Open(filename)
    if err != nil {
        log.Printf("Can't open the config file \"%v\". Using default values.", filename)
        return config
    }
    defer file.Close()

    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        log.Printf("There was an error reading the config file. Using default values")
        return config
    }

    if err := yaml.Unmarshal(bytes, &config); err != nil {
        log.Printf("Error parsing the config file. Using default values.")
    }

    return config
}

// nginxHandler handles the /nginx endpoint
func nginxHandler(config Config, w http.ResponseWriter, r *http.Request) {
    data := NginxData {
        NginxState:		getServiceState("nginx"),
        NginxConnections:	getServiceConnections("nginx", config.NginxPort),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

// prosodyHandler handles the /prosody endpoint
func prosodyHandler(config Config, w http.ResponseWriter, r *http.Request) {
    data := ProsodyData {
        ProsodyState:		getServiceState("prosody"),
        ProsodyConnections:	getServiceConnections("prosody", config.ProsodyPort),
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

// jvbHandler handles the /jvb endpoint
func jvbHandler(config Config, w http.ResponseWriter, r *http.Request) {
    data := JVBData {
        JVBState:		getServiceState("jitsi-videobridge2"),
        JVBAPIData:		getJitsiAPIData("jitsi-videobridge2", config.JVBStatsURL),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

// jibriHandler handles the /jibri endpoint
func jibriHandler(config Config, w http.ResponseWriter, r *http.Request) {
    data := JibriData {
        JibriState:		getServiceState("jibri"),
        JibriHealthData:	getJitsiAPIData("jibri", config.JibriHealthURL),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}


// main sets up the http server and the routes
func main() {

    // load the configuration
    config := loadConfig("jilo-agent.conf")

    // endpoints
    http.HandleFunc("/nginx", func(w http.ResponseWriter, r *http.Request) {
        nginxHandler(config, w, r)
    })
    http.HandleFunc("/prosody", func(w http.ResponseWriter, r *http.Request) {
        prosodyHandler(config, w, r)
    })
    http.HandleFunc("/jicofo", func(w http.ResponseWriter, r *http.Request) {
        jicofoHandler(config, w, r)
    })
    http.HandleFunc("/jvb", func(w http.ResponseWriter, r *http.Request) {
        jvbHandler(config, w, r)
    })
    http.HandleFunc("/jibri", func(w http.ResponseWriter, r *http.Request) {
        jibriHandler(config, w, r)
    })

    // start the http server
    agentPortStr := fmt.Sprintf(":%d", config.AgentPort)
    fmt.Printf("Starting Jilo agent on port %d.\n", config.AgentPort)
    if err := http.ListenAndServe(agentPortStr, nil); err != nil {
        log.Fatalf("Could not start the agent: %v\n", err)
    }
}
