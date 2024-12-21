package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func main() {
	h := newHandler()
	http.HandleFunc("/health", h.health)
	http.HandleFunc("/reset", h.reset)
	http.HandleFunc("/configure", h.configure)
	http.HandleFunc("/list", h.list)
	http.HandleFunc("/", h.catchAll)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server closed")
}

type handler struct {
	mu        sync.RWMutex
	endpoints map[string]any
}

func newHandler() *handler {
	return &handler{
		endpoints: map[string]any{},
	}
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *handler) reset(w http.ResponseWriter, r *http.Request) {
	h.endpoints = map[string]any{}
	w.WriteHeader(http.StatusOK)
}

// EndpointConfigurationRequest is the request body for the /configure endpoint.
type EndpointConfigurationRequest struct {
	Path     string `json:"path"`
	Response any    `json:"response"`
}

func (h *handler) configure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	request := EndpointConfigurationRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	if request.Response == nil {
		http.Error(w, "Response is required", http.StatusBadRequest)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.endpoints[request.Path] = request.Response

	w.WriteHeader(http.StatusOK)
}

func (h *handler) catchAll(w http.ResponseWriter, r *http.Request) {
	response, exists := h.endpoints[r.URL.Path]
	if !exists {
		http.NotFound(w, r)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Default().Println("Failed to marshal response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Default().Println("Failed to write response:", err)
	}
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	b, err := json.Marshal(h.endpoints)
	if err != nil {
		log.Default().Println("Failed to marshal endpoints:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Default().Println("Failed to write response:", err)
	}
}
