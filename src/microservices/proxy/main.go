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
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	targetUrl := moviesUrl + "/api/movies"
	log.Printf("TargetURL %s", targetUrl)
	// Создаем новый GET-запрос к целевому API
	req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Можно пробросить заголовки (например, Authorization)
	req.Header = r.Header.Clone()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	// Устанавливаем статус-код
	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("error copying response body:", err)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	targetUrl := monolithUrl + "/api/users"
	log.Printf("TargetURL %s", targetUrl)
	// Создаем новый GET-запрос к целевому API
	req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Можно пробросить заголовки (например, Authorization)
	req.Header = r.Header.Clone()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	// Устанавливаем статус-код
	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("error copying response body:", err)
	}
}
