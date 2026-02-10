package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"milvus-kb-demo/internal/auth"
	"milvus-kb-demo/internal/history"
	"milvus-kb-demo/internal/service"
)

type Server struct {
	userManager    *auth.UserManager
	historyManager *history.HistoryManager
	service        *service.Service
	mux            *http.ServeMux
	staticDir      string
}

func NewServer(userManager *auth.UserManager, historyManager *history.HistoryManager, svc *service.Service, staticDir string) *Server {
	s := &Server{
		userManager:    userManager,
		historyManager: historyManager,
		service:        svc,
		mux:            http.NewServeMux(),
		staticDir:      staticDir,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Auth
	s.mux.HandleFunc("/api/register", s.handleRegister)
	s.mux.HandleFunc("/api/login", s.handleLogin)

	// Conversations
	s.mux.HandleFunc("/api/conversations", s.handleConversations)       // GET (List), POST (Create)
	s.mux.HandleFunc("/api/conversations/", s.handleConversationDetail) // GET (Detail), DELETE

	// Chat
	s.mux.HandleFunc("/api/chat", s.handleChat)

	// Static Files (UI) - though user said no need for frontend, we keep it serving just in case
	fs := http.FileServer(http.Dir(s.staticDir))
	s.mux.Handle("/", fs)
}

func (s *Server) Run(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: s.mux,
		// Disable timeouts for SSE
		// ReadTimeout:  30 * time.Second,
		// WriteTimeout: 30 * time.Second,
	}
	log.Printf("Starting web server on %s", addr)
	return server.ListenAndServe()
}

// Data Structures

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type CreateConversationRequest struct {
	Username string `json:"username"`
	Title    string `json:"title"`
}

type ChatRequest struct {
	Username       string `json:"username"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

// Handlers

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid request body"})
		return
	}
	if creds.Username == "" || creds.Password == "" {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username and password are required"})
		return
	}
	if err := s.userManager.Register(creds.Username, creds.Password); err != nil {
		s.jsonResponse(w, http.StatusConflict, Response{Error: err.Error()})
		return
	}
	s.jsonResponse(w, http.StatusCreated, Response{Message: "User registered successfully"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid request body"})
		return
	}
	if err := s.userManager.Login(creds.Username, creds.Password); err != nil {
		s.jsonResponse(w, http.StatusUnauthorized, Response{Error: "Invalid credentials"})
		return
	}
	s.jsonResponse(w, http.StatusOK, Response{Message: "Login successful"})
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List conversations
		username := r.URL.Query().Get("username")
		if username == "" {
			s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username is required"})
			return
		}
		convs, err := s.historyManager.GetConversations(username)
		if err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, Response{Error: err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusOK, convs)

	case http.MethodPost:
		// Create conversation
		var req CreateConversationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid request body"})
			return
		}
		if req.Username == "" {
			s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username is required"})
			return
		}
		conv, err := s.historyManager.CreateConversation(req.Username, req.Title)
		if err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, Response{Error: err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusCreated, conv)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleConversationDetail(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL /api/conversations/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/conversations/")
	if path == "" {
		http.Error(w, "Conversation ID required", http.StatusBadRequest)
		return
	}
	// Handle trailing slash if present, though TrimPrefix handles it if ID is simple
	parts := strings.Split(path, "/")
	conversationID := parts[0]

	username := r.URL.Query().Get("username")
	if username == "" {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username is required"})
		return
	}

	// Verify ownership
	conv, err := s.historyManager.GetConversation(conversationID)
	if err != nil {
		s.jsonResponse(w, http.StatusNotFound, Response{Error: "Conversation not found"})
		return
	}
	if conv.UserID != username {
		s.jsonResponse(w, http.StatusForbidden, Response{Error: "Access denied"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.jsonResponse(w, http.StatusOK, conv)

	case http.MethodDelete:
		if err := s.historyManager.DeleteConversation(conversationID); err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, Response{Error: err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusOK, Response{Message: "Conversation deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid request body"})
		return
	}

	if req.Username == "" || req.ConversationID == "" || req.Content == "" {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username, conversation_id and content are required"})
		return
	}

	// Save User Message
	if err := s.historyManager.AddMessage(req.ConversationID, "user", req.Content); err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, Response{Error: "Failed to save message: " + err.Error()})
		return
	}

	// Setup SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Stream Answer
	ctx := r.Context()
	streamChan, err := s.service.AnswerStream(ctx, req.Content)
	if err != nil {
		// If error happens before stream starts, try to send JSON error if headers not flushed (unlikely if we set SSE headers)
		// Better send an error event
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	var fullAnswer strings.Builder

	for event := range streamChan {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		if event.Type == "token" {
			fullAnswer.WriteString(event.Content)
		}
	}

	// Save Assistant Message
	if err := s.historyManager.AddMessage(req.ConversationID, "assistant", fullAnswer.String()); err != nil {
		log.Printf("Failed to save assistant message: %v", err)
	}
}

func (s *Server) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
