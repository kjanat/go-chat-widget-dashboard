package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"

	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
	"github.com/kjanat/go-chat-widget-dashboard/internal/services"
)

// Handler holds all the service dependencies
type Handler struct {
	customerService *services.CustomerService
	chatService     *services.ChatService
	openaiService   *services.OpenAIService
	authService     *services.AuthService
	store           *sessions.CookieStore
	upgrader        websocket.Upgrader
	templates       *template.Template
}

// New creates a new handler with all dependencies
func New(
	customerService *services.CustomerService,
	chatService *services.ChatService,
	openaiService *services.OpenAIService,
	authService *services.AuthService,
	store *sessions.CookieStore,
	templates *template.Template,
) *Handler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			customerID := mux.Vars(r)["customerID"]
			return customerService.ValidateOrigin(origin, customerID)
		},
	}

	return &Handler{
		customerService: customerService,
		chatService:     chatService,
		openaiService:   openaiService,
		authService:     authService,
		store:           store,
		upgrader:        upgrader,
		templates:       templates,
	}
}

// WidgetJS serves the JavaScript widget
func (h *Handler) WidgetJS(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer")
	if customerID == "" {
		http.Error(w, "Customer ID required", http.StatusBadRequest)
		return
	}

	// Validate API key if provided
	apiKey := r.URL.Query().Get("key")

	customer, err := h.customerService.GetActiveByID(customerID)
	if err != nil {
		http.Error(w, "Invalid customer", http.StatusNotFound)
		return
	}

	if customer.APIKey != "" && customer.APIKey != apiKey {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	config := models.WidgetConfig{
		CustomerID:  customer.ID,
		BrandColors: customer.BrandColorsToJSON(),
		LogoURL:     customer.LogoURL,
		ModelPath:   customer.ModelPath,
		Animations:  customer.Animations,
		WSEndpoint:  fmt.Sprintf("wss://%s/ws/%s", r.Host, customer.ID),
	}

	h.templates.ExecuteTemplate(w, "widget.js", config)
}

// WebSocket handles chat WebSocket connections
func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	customerID := mux.Vars(r)["customerID"]

	customer, err := h.customerService.GetActiveByID(customerID)
	if err != nil {
		http.Error(w, "Invalid customer", http.StatusNotFound)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Create session
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = fmt.Sprintf("user-%d", time.Now().UnixNano())
	}

	sessionID, err := h.chatService.CreateSession(customerID, userID)
	if err != nil {
		log.Println("Error creating session:", err)
		return
	}

	// Handle messages
	for {
		var msgReq models.ChatMessageRequest
		err := conn.ReadJSON(&msgReq)
		if err != nil {
			break
		}

		// Log user message
		_, err = h.chatService.CreateMessage(sessionID, "user", msgReq.Content, "")
		if err != nil {
			log.Println("Error logging message:", err)
			continue
		}

		// Generate AI response
		response, emotion := h.openaiService.GenerateResponse(customer.OpenAIPrompt, msgReq.Content)

		// Log assistant message
		messageID, err := h.chatService.CreateMessage(sessionID, "assistant", response, emotion)
		if err != nil {
			log.Println("Error logging assistant message:", err)
			continue
		}

		// Send response
		msgResp := models.ChatMessageResponse{
			Content:   response,
			Emotion:   emotion,
			MessageID: messageID,
		}
		conn.WriteJSON(msgResp)
	}

	// End session
	h.chatService.EndSession(sessionID)
}

// DashboardLogin handles admin login
func (h *Handler) DashboardLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.authService.GetAdminUser(username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}

	session, _ := h.store.Get(r, "admin-session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Save(r, w)

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// Dashboard handles the main admin dashboard
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	customers, err := h.customerService.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Username":  session.Values["username"],
		"Customers": customers,
	}

	h.templates.ExecuteTemplate(w, "dashboard.html", data)
}

// DashboardLogout handles admin logout
func (h *Handler) DashboardLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")

	// Clear the session
	session.Values["user_id"] = nil
	session.Values["username"] = nil
	session.Options.MaxAge = -1 // This deletes the session
	session.Save(r, w)

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// CustomerEdit handles customer editing
func (h *Handler) CustomerEdit(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	customerID := mux.Vars(r)["id"]

	if r.Method == "GET" {
		customer, err := h.customerService.GetByID(customerID)
		if err != nil {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}

		h.templates.ExecuteTemplate(w, "customer-edit.html", customer)
		return
	}

	// Handle POST - update customer
	customer := &models.Customer{
		Name:           r.FormValue("name"),
		Email:          r.FormValue("email"),
		BrandColors:    r.FormValue("brand_colors"),
		LogoURL:        r.FormValue("logo_url"),
		OpenAIPrompt:   r.FormValue("openai_prompt"),
		AllowedDomains: r.FormValue("allowed_domains"),
		Active:         r.FormValue("active") == "on",
	}

	err := h.customerService.Update(customerID, customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// CustomerCreate handles customer creation
func (h *Handler) CustomerCreate(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create new customer
	customer := &models.Customer{
		Name:           r.FormValue("name"),
		Email:          r.FormValue("email"),
		BrandColors:    r.FormValue("brand_colors"),
		LogoURL:        r.FormValue("logo_url"),
		OpenAIPrompt:   r.FormValue("openai_prompt"),
		AllowedDomains: r.FormValue("allowed_domains"),
		Active:         true, // New customers are active by default
	}

	err := h.customerService.Create(customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// ModelUpload handles 3D model uploads
func (h *Handler) ModelUpload(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	customerID := mux.Vars(r)["id"]

	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("model")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create models directory if it doesn't exist
	os.MkdirAll("./uploads/models", 0755)

	// Save file
	filename := fmt.Sprintf("%s_%s", customerID, handler.Filename)
	filePath := filepath.Join("./uploads/models", filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update customer record
	modelPath := "/models/" + filename
	err = h.customerService.UpdateModelPath(customerID, modelPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"path":   modelPath,
	})
}

// ModelDownload handles 3D model downloads
func (h *Handler) ModelDownload(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	customerID := mux.Vars(r)["id"]

	customer, err := h.customerService.GetByID(customerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	if customer.ModelPath == "" {
		http.Error(w, "No model uploaded", http.StatusNotFound)
		return
	}

	// Get the actual file path
	filePath := filepath.Join("./uploads/models", filepath.Base(customer.ModelPath))

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Model file not found", http.StatusNotFound)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(customer.ModelPath)))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// ChatLogs handles chat log viewing
func (h *Handler) ChatLogs(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	customerID := r.URL.Query().Get("customer")

	sessions, err := h.chatService.GetSessions(customerID, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Sessions":   sessions,
		"CustomerID": customerID,
	}

	h.templates.ExecuteTemplate(w, "chat-logs.html", data)
}
