package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

var monolithUrl string
var moviesUrl string

func main() {
	http.HandleFunc("/api/users", handleUsers)
	http.HandleFunc("/api/movies", handleMovies)
	http.HandleFunc("/health", handleHealth)

	monolithUrl = os.Getenv("MONOLITH_URL")
	if monolithUrl == "" {
		monolithUrl = "http://localhost:8080"
	}
	moviesUrl = os.Getenv("MOVIES_SERVICE_URL")
	if moviesUrl == "" {
		moviesUrl = "http://localhost:8081"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Starting proxy on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func handleMovies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost:
		proxyRequest(w, r, moviesUrl, "/api/movies")
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost:
		proxyRequest(w, r, monolithUrl, "/api/users")
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func proxyRequest(w http.ResponseWriter, r *http.Request, targetBase string, path string) {
	targetURL := targetBase + path
	log.Printf("Proxying %s request to %s", r.Method, targetURL)

	var body io.Reader
	if r.Method == http.MethodPost {
		body = r.Body
	}

	req, err := http.NewRequest(r.Method, targetURL, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header = r.Header.Clone()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Println("error copying response body:", err)
	}
}
