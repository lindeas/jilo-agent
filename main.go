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
    "flag"
    "fmt"
    "encoding/json"
    "gopkg.in/yaml.v2"
    "github.com/dgrijalva/jwt-go"
    "io/ioutil"
    "log"
    "net/http"
    "net/http/httptest"
    "os"
    "os/exec"
    "strconv"
    "strings"
)

// Config holds the structure of the configuration file
type Config struct {
    AgentPort		int	`yaml:"agent_port"`
    SSLcert		string	`yaml:"ssl_cert"`
    SSLkey		string	`yaml:"ssl_key"`
    SecretKey		string	`yaml:"secret_key"`
    NginxPort		int	`yaml:"nginx_port"`
    ProsodyPort		int	`yaml:"prosody_port"`
    JicofoStatsURL	string	`yaml:"jicofo_stats_url"`
    JVBStatsURL		string	`yaml:"jvb_stats_url"`
    JibriHealthURL	string	`yaml:"jibri_health_url"`
}

// Claims holds JWT access right
type Claims struct {
    Username		string	`json:"sub"`
    Role		string	`json:"role"`
    jwt.StandardClaims
}

// StatusData holds the status of the agent and its endpoints
type StatusData struct {
    AgentStatus string `json:"agent_status"`
    Endpoints   map[string]string `json:"endpoints"`
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

var secretKey []byte

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

// authenticationJWT handles the JWT auth
func authenticationJWT(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")

        // DEBUG print the token for debugging (only for debug, remove in prod)
        //log.Println("Received token:", tokenString)

        // empty auth header
        if tokenString == "" {
            log.Println("No Authorization header received")
            http.Error(w, "Auth header not received", http.StatusUnauthorized)
            return
        }

        // remove "Bearer "
        if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
            tokenString = tokenString[7:]
        } else {
            log.Println("Bearer token missing")
            http.Error(w, "Malformed Authorization header", http.StatusUnauthorized)
            return
        }

        // DEBUG print out the token for debugging (remove in production!)
        //log.Printf("Received JWT: %s", tokenString)

        claims := &Claims{}

        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            // DEBUG log secret key for debugging (remove from production!)
            //log.Printf("Parsing JWT with secret key: %s", secretKey)
            return secretKey, nil
        })

        // JWT errors and error logging
        if err != nil {
            // log the error message for debugging (not in prod!)
            log.Printf("JWT parse error: %v", err)
            if err == jwt.ErrSignatureInvalid {
                http.Error(w, "Invalid JWT signature", http.StatusUnauthorized)
                return
            }
            http.Error(w, "Error parsing JWT: "+err.Error(), http.StatusUnauthorized)
            return
        }

        // JWT invalid
        if !token.Valid {
            log.Println("Invalid JWT token")
            http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// statusHandler handles the /status endpoint
func statusHandler(config Config, w http.ResponseWriter, r *http.Request) {
    // Check if the agent is running
    // FIXME add logic here to check if the agent is running OK, with no errors
    agentStatus := "running"

    // Prepare the endpoint status map
    endpointStatuses := make(map[string]string)

    // Determine protocol based on SSL config
    protocol := "http"
    if config.SSLcert != "" && config.SSLkey != "" {
        protocol = "https"
    }

    // Check if each endpoint is available or not
    endpoints := []string{"nginx", "prosody", "jicofo", "jvb", "jibri"}
    for _, endpoint := range endpoints {
        endpointURL := fmt.Sprintf("%s://localhost:%d/%s", protocol, config.AgentPort, endpoint)

        req, err := http.NewRequest(http.MethodGet, endpointURL, nil)
        if err != nil {
            endpointStatuses[endpoint] = "not available"
            continue
        }

        // Copy the JWT token from the original request
        req.Header.Set("Authorization", r.Header.Get("Authorization"))

        // Create a response recorder to capture the response
        rr := httptest.NewRecorder()

        // Call the respective handler with the new request
        switch endpoint {
        case "nginx":
            nginxHandler(config, rr, req)
        case "prosody":
            prosodyHandler(config, rr, req)
        case "jicofo":
            jicofoHandler(config, rr, req)
        case "jvb":
            jvbHandler(config, rr, req)
        case "jibri":
            jibriHandler(config, rr, req)
        }

        // Check the status code from the response recorder
        if rr.Result().StatusCode == http.StatusOK {
            endpointStatuses[endpoint] = "available"
        } else {
            endpointStatuses[endpoint] = "not available"
        }

    }

    // Prepare the response data and send back the JSON
    statusData := StatusData{
        AgentStatus: agentStatus,
        Endpoints:   endpointStatuses,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(statusData)
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


// CORS Middleware to handle CORS for all endpoints
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins or restrict to specific domain
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight (OPTIONS) request
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent) // Respond with 204 No Content for preflight
            return
        }

        // Pass the request to the next handler
        next.ServeHTTP(w, r)
    })
}


// main sets up the http server and the routes
func main() {

    // Define a flag for the config file
    configFile := flag.String("c", "./jilo-agent.conf", "Specify the agent config file")

    // Parse the flags
    flag.Parse()

    // Check if the file exists, fallback to default config file if not
    if _, err := os.Stat(*configFile); os.IsNotExist(err) {
        fmt.Println("Config file not found, using default values")
    }

    // Load the configuration from the specified file (option -c) or the default config file name
    config := loadConfig(*configFile)

    secretKey = []byte(config.SecretKey)

    mux := http.NewServeMux()

    // endpoints
    mux.Handle("/status", authenticationJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        statusHandler(config, w, r)
    })))
    mux.Handle("/nginx", authenticationJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        nginxHandler(config, w, r)
    })))
    mux.Handle("/prosody", authenticationJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        prosodyHandler(config, w, r)
    })))
    mux.Handle("/jicofo", authenticationJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        jicofoHandler(config, w, r)
    })))
    mux.Handle("/jvb", authenticationJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        jvbHandler(config, w, r)
    })))
    mux.Handle("/jibri", authenticationJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        jibriHandler(config, w, r)
    })))

    // add the CORS headers to the mux
    corsHandler := corsMiddleware(mux)

    // start the http server
    agentPortStr := fmt.Sprintf(":%d", config.AgentPort)
    fmt.Printf("Starting Jilo agent on port %d.\n", config.AgentPort)
//    if err := http.ListenAndServe(agentPortStr, corsHandler); err != nil {
    if err := http.ListenAndServeTLS(agentPortStr, config.SSLcert, config.SSLkey, corsHandler); err != nil {
        log.Fatalf("Could not start the agent: %v\n", err)
    }
}
