package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/IBM/sarama"
)

var kafkaProducer sarama.AsyncProducer

func main() {
	kafkaBootstrapServers := os.Getenv("KAFKA_BROKERS")
	if kafkaBootstrapServers == "" {
		kafkaBootstrapServers = "localhost:9092"
	}

	brokers := []string{kafkaBootstrapServers}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	var err error
	kafkaProducer, err = sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Обработка ошибок продьюсера в отдельной горутине
	go func() {
		for err := range kafkaProducer.Errors() {
			log.Printf("Failed to write message to Kafka: %v", err)
		}
	}()

	http.HandleFunc("/api/events/health", handleHealth)
	http.HandleFunc("/api/events/movie", handleEventMovie)
	http.HandleFunc("/api/events/user", handleEventUser)
	http.HandleFunc("/api/events/payment", handleEventPayment)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Printf("Starting events service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func handleEventMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		MovieID int    `json:"movie_id"`
		Title   string `json:"title"`
		Action  string `json:"action"`
		UserID  int    `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	msgBytes, _ := json.Marshal(payload)
	kafkaProducer.Input() <- &sarama.ProducerMessage{
		Topic: "movie-events",
		Value: sarama.ByteEncoder(msgBytes),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func handleEventUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		UserID    int    `json:"user_id"`
		Username  string `json:"username"`
		Action    string `json:"action"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	msgBytes, _ := json.Marshal(payload)
	kafkaProducer.Input() <- &sarama.ProducerMessage{
		Topic: "user-events",
		Value: sarama.ByteEncoder(msgBytes),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func handleEventPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		PaymentID  int     `json:"payment_id"`
		UserID     int     `json:"user_id"`
		Amount     float64 `json:"amount"`
		Status     string  `json:"status"`
		Timestamp  string  `json:"timestamp"`
		MethodType string  `json:"method_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	msgBytes, _ := json.Marshal(payload)
	kafkaProducer.Input() <- &sarama.ProducerMessage{
		Topic: "payment-events",
		Value: sarama.ByteEncoder(msgBytes),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
