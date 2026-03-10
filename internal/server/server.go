package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"milvus-kb-demo/internal/auth"
	"milvus-kb-demo/internal/history"
	"milvus-kb-demo/internal/service"
)

type Server struct {
	userManager    *auth.UserManager
	historyManager *history.HistoryManager
	service        *service.Service
	mux            *http.ServeMux
	handler        http.Handler
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
	// Middleware Chain
	handler := s.corsMiddleware(s.recoveryMiddleware(s.loggingMiddleware(s.mux)))

	// Auth
	s.mux.HandleFunc("/api/register", s.handleRegister)
	s.mux.HandleFunc("/api/login", s.handleLogin)

	// User Profile
	s.mux.HandleFunc("/api/user/profile", s.handleUserProfile)
	s.mux.HandleFunc("/api/user/update", s.handleUserUpdate)
	s.mux.HandleFunc("/api/user/avatar", s.handleAvatarUpload)

	// Conversations
	s.mux.HandleFunc("/api/conversations", s.handleConversations)             // GET (List), POST (Create)
	s.mux.HandleFunc("/api/conversations/", s.handleConversationDetail)       // GET (Detail), DELETE
	s.mux.HandleFunc("/api/conversations/rename", s.handleConversationRename) // POST

	// Chat
	s.mux.HandleFunc("/api/chat", s.handleChat)

	// Static Files (UI)
	// Support serving uploaded avatars
	s.mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	fs := http.FileServer(http.Dir(s.staticDir))
	s.mux.Handle("/", fs)

	// Assign the wrapped handler to the server's handler in Run()
	// But since s.mux is used directly in Run(), we need to change Run() or store the wrapped handler
	s.handler = handler
}

func (s *Server) Run(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: s.handler, // Use the wrapped handler
		// Disable timeouts for SSE
		// ReadTimeout:  30 * time.Second,
		// WriteTimeout: 30 * time.Second,
	}
	log.Printf("Starting web server on %s", addr)
	return server.ListenAndServe()
}

// Middlewares

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins for dev
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
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

type RenameConversationRequest struct {
	ConversationID string `json:"conversation_id"`
	NewTitle       string `json:"new_title"`
}

type ChatRequest struct {
	Username       string `json:"username"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

type UserUpdateRequest struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// Handlers

func (s *Server) handleUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	if username == "" {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username is required"})
		return
	}
	user, err := s.userManager.GetUser(username)
	if err != nil {
		s.jsonResponse(w, http.StatusNotFound, Response{Error: err.Error()})
		return
	}
	s.jsonResponse(w, http.StatusOK, user)
}

func (s *Server) handleUserUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid request body"})
		return
	}
	if req.Username == "" {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username is required"})
		return
	}

	// Get existing user to keep avatar
	user, err := s.userManager.GetUser(req.Username)
	if err != nil {
		s.jsonResponse(w, http.StatusNotFound, Response{Error: "User not found"})
		return
	}

	if err := s.userManager.UpdateUser(req.Username, req.Email, req.Phone, user.Avatar, req.Nickname); err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.jsonResponse(w, http.StatusOK, Response{Message: "Profile updated successfully"})
}

func (s *Server) handleAvatarUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 5MB)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "File too large or invalid form"})
		return
	}

	username := r.FormValue("username")
	if username == "" {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Username is required"})
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Avatar file is required"})
		return
	}
	defer file.Close()

	// Ensure upload dir exists
	uploadDir := "./uploads"
	if _, statErr := os.Stat(uploadDir); os.IsNotExist(statErr) {
		os.Mkdir(uploadDir, 0755)
	}

	// Generate filename: username_timestamp_filename
	filename := fmt.Sprintf("%s_%d_%s", username, time.Now().Unix(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, Response{Error: "Failed to create file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, Response{Error: "Failed to save file"})
		return
	}

	// Update DB with relative URL
	avatarURL := "/uploads/" + filename
	if err := s.userManager.UpdateAvatar(username, avatarURL); err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, Response{Error: "Failed to update database"})
		return
	}

	s.jsonResponse(w, http.StatusOK, Response{Message: "Avatar uploaded successfully", Error: avatarURL}) // Using Error field for URL for simplicity or add field
}

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

func (s *Server) handleConversationRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RenameConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid request body"})
		return
	}
	if req.ConversationID == "" || req.NewTitle == "" {
		s.jsonResponse(w, http.StatusBadRequest, Response{Error: "ConversationID and NewTitle are required"})
		return
	}

	if err := s.historyManager.RenameConversation(req.ConversationID, req.NewTitle); err != nil {
		s.jsonResponse(w, http.StatusNotFound, Response{Error: err.Error()})
		return
	}
	s.jsonResponse(w, http.StatusOK, Response{Message: "Conversation renamed successfully"})
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

	// Async Title Generation if it's the first message or still "New Chat"
	go func(convID, content string) {
		conv, err := s.historyManager.GetConversation(convID)
		if err == nil && (conv.Title == "New Chat" || len(conv.Messages) <= 2) {
			// Only generate title if it's basically empty or just started
			newTitle, err := s.service.GenerateTitle(context.Background(), content)
			if err == nil && newTitle != "" {
				s.historyManager.RenameConversation(convID, strings.Trim(newTitle, "\""))
			}
		}
	}(req.ConversationID, req.Content)

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
	conversation, convErr := s.historyManager.GetConversation(req.ConversationID)
	if convErr != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", convErr.Error())
		flusher.Flush()
		return
	}
	streamChan, err := s.service.AnswerStream(ctx, conversation.Messages)
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
